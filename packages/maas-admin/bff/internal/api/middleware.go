package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/opendatahub-io/gen-ai/internal/config"

	"github.com/julienschmidt/httprouter"
	"github.com/opendatahub-io/gen-ai/internal/integrations"

	"github.com/google/uuid"
	"github.com/opendatahub-io/gen-ai/internal/constants"
	helper "github.com/opendatahub-io/gen-ai/internal/helpers"
	"github.com/opendatahub-io/gen-ai/internal/integrations/maas"
	"github.com/rs/cors"
)

func (app *App) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
				logger := helper.GetContextLoggerFromReq(r)
				logger.Error("Recovered from panic", slog.String("stack_trace", string(debug.Stack())))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *App) EnableCORS(next http.Handler) http.Handler {
	if len(app.config.AllowedOrigins) == 0 {
		// CORS is disabled, this middleware becomes a noop.
		return next
	}

	c := cors.New(cors.Options{
		AllowedOrigins:     app.config.AllowedOrigins,
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "PUT", "POST", "PATCH", "DELETE"},
		AllowedHeaders:     []string{},
		Debug:              app.config.LogLevel == slog.LevelDebug,
		OptionsPassthrough: false,
	})

	return c.Handler(next)
}

func (app *App) EnableTelemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adds a unique id to the context to allow tracing of requests
		traceId := uuid.NewString()
		ctx := context.WithValue(r.Context(), constants.TraceIdKey, traceId)

		// logger will only be nil in tests.
		if app.logger != nil {
			traceLogger := app.logger.With(slog.String("trace_id", traceId))
			ctx = context.WithValue(ctx, constants.TraceLoggerKey, traceLogger)

			traceLogger.Debug("Incoming HTTP request", slog.Any("request", helper.RequestLogValuer{Request: r}))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *App) InjectRequestIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//skip use headers check if we are not on the configured API path prefix (i.e. we are on /healthcheck and / (static fe files) )
		// Check for both direct API path and prefixed API path
		isAPIPath := strings.HasPrefix(r.URL.Path, app.config.APIPathPrefix) ||
			strings.HasPrefix(r.URL.Path, constants.PathPrefix+app.config.APIPathPrefix)

		if !isAPIPath {
			next.ServeHTTP(w, r)
			return
		}

		// If authentication is disabled, skip identity extraction
		if app.config.AuthMethod == config.AuthMethodDisabled {
			next.ServeHTTP(w, r)
			return
		}

		identity, err := app.kubernetesClientFactory.ExtractRequestIdentity(r.Header)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), constants.RequestIdentityKey, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AttachMaaSClient middleware creates a MaaS client and attaches it to context.
// This middleware can be used independently and doesn't require namespace.
//
// In mock mode, creates a mock client. In real mode, uses autodiscovery or configured MaaS URL.
// Uses RequestIdentity from context for authentication, consistent with other clients.
func (app *App) AttachMaaSClient(next func(http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		// Use request-scoped logger to avoid nil-panic in tests/environments where app.logger is not set
		logger := helper.GetContextLoggerFromReq(r)

		var maasClient maas.MaaSClient

		// Check if running in mock mode
		if app.config.MockMaaSClient {
			logger.Debug("MOCK MODE: creating mock MaaS client")
			// In mock mode, use empty URL since mock factory ignores it
			maasClient = app.maasClientFactory.CreateClient("", "", app.config.InsecureSkipVerify, app.rootCAs)
		} else {
			var serviceURL string

			// Configuration Priority:
			// 1. MAAS_URL env var (if set for local dev) - works even without cluster domain
			// 2. Autodiscovered endpoint (production default) - requires cluster domain
			// 3. MaaS unavailable - attach nil and let handler decide if it's needed

			if app.config.MaaSURL != "" {
				// Priority 1: Use environment variable if explicitly set
				serviceURL = app.config.MaaSURL
				logger.Debug("Using MAAS_URL environment variable (developer override)",
					"serviceURL", serviceURL)
			} else if app.clusterDomain != "" {
				// Priority 2: Autodiscovery using cached cluster domain (from service account at startup)
				serviceURL = fmt.Sprintf("https://maas.%s/maas-api", app.clusterDomain)
				logger.Debug("Using autodiscovered MaaS endpoint from cached cluster domain",
					"clusterDomain", app.clusterDomain,
					"serviceURL", serviceURL)
			} else {
				// Priority 3: MaaS unavailable - neither env var nor cluster domain available
				logger.Debug("MaaS unavailable: no MAAS_URL configured and cluster domain not available")
				ctx = context.WithValue(ctx, constants.MaaSClientKey, nil)
				next(w, r.WithContext(ctx), ps)
				return
			}

			// Get RequestIdentity from context (set by InjectRequestIdentity middleware)
			identity, ok := ctx.Value(constants.RequestIdentityKey).(*integrations.RequestIdentity)
			if !ok || identity == nil {
				app.serverErrorResponse(w, r, fmt.Errorf("missing RequestIdentity in context"))
				return
			}

			logger.Debug("Creating MaaS client",
				"serviceURL", serviceURL,
				"hasAuthToken", identity.Token != "")

			// Create MaaS client per-request using app factory with auth token from RequestIdentity
			maasClient = app.maasClientFactory.CreateClient(serviceURL, identity.Token, app.config.InsecureSkipVerify, app.rootCAs)
		}

		// Attach ready-to-use client to context
		ctx = context.WithValue(ctx, constants.MaaSClientKey, maasClient)
		r = r.WithContext(ctx)

		next(w, r, ps)
	}
}
