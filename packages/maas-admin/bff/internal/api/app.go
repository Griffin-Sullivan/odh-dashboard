package api

import (
	"context"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	k8s "github.com/opendatahub-io/gen-ai/internal/integrations/kubernetes"
	"github.com/opendatahub-io/gen-ai/internal/integrations/kubernetes/k8smocks"
	"github.com/opendatahub-io/gen-ai/internal/integrations/maas"
	"github.com/opendatahub-io/gen-ai/internal/integrations/maas/maasmocks"
	"github.com/opendatahub-io/gen-ai/internal/repositories"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/julienschmidt/httprouter"
	"github.com/opendatahub-io/gen-ai/internal/cache"
	"github.com/opendatahub-io/gen-ai/internal/config"
	"github.com/opendatahub-io/gen-ai/internal/constants"
	helper "github.com/opendatahub-io/gen-ai/internal/helpers"
)

type App struct {
	config                  config.EnvConfig
	logger                  *slog.Logger
	repositories            *repositories.Repositories
	openAPI                 *OpenAPIHandler
	kubernetesClientFactory k8s.KubernetesClientFactory
	maasClientFactory       maas.MaaSClientFactory
	dashboardNamespace      string
	memoryStore             cache.MemoryStore
	rootCAs                 *x509.CertPool
	clusterDomain           string
}

func NewApp(cfg config.EnvConfig, logger *slog.Logger) (*App, error) {
	logger.Debug("Initializing app with config", slog.Any("config", cfg))
	var rootCAs *x509.CertPool

	// Initialize CA pool if bundle paths are provided
	if len(cfg.BundlePaths) > 0 {
		// Start with system certs if available
		if pool, err := x509.SystemCertPool(); err == nil {
			rootCAs = pool
		} else {
			rootCAs = x509.NewCertPool()
		}
		var loadedAny bool
		for _, p := range cfg.BundlePaths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			// Read and append each PEM bundle; ignore errors per file, log at debug
			pemBytes, readErr := os.ReadFile(p)
			if readErr != nil {
				logger.Debug("CA bundle not readable, skipping", slog.String("path", p), slog.Any("error", readErr))
				continue
			}
			if ok := rootCAs.AppendCertsFromPEM(pemBytes); !ok {
				logger.Debug("No certs appended from PEM bundle", slog.String("path", p))
				continue
			}
			loadedAny = true
			logger.Info("Added CA bundle", slog.String("path", p))
		}
		if !loadedAny {
			// If none were loaded successfully, keep rootCAs nil to fall back to default transport behavior
			rootCAs = nil
			logger.Warn("No CA certificates loaded from bundle-paths; falling back to system defaults")
		}
	}

	// Detect dashboard namespace
	dashboardNamespace, err := helper.GetCurrentNamespace()
	if err != nil {
		logger.Warn("Failed to detect dashboard namespace, using default", "error", err, "default", "opendatahub")
		dashboardNamespace = "opendatahub"
	}
	logger.Info("Detected dashboard namespace", "namespace", dashboardNamespace)

	// Initialize MaaS client factory - clients will be created per request
	var maasClientFactory maas.MaaSClientFactory
	if cfg.MockMaaSClient {
		logger.Info("Using mock MaaS client factory")
		maasClientFactory = maasmocks.NewMockClientFactory()
	} else {
		logger.Info("Using real MaaS client factory", "url", cfg.MaaSURL)
		maasClientFactory = maas.NewRealClientFactory()
	}

	// Initialize OpenAPI handler
	openAPIHandler, err := NewOpenAPIHandler(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI handler: %w", err)
	}

	var k8sFactory k8s.KubernetesClientFactory
	// used only on mocked k8s client
	var testEnv *envtest.Environment

	if cfg.MockK8sClient {
		logger.Info("Using mocked Kubernetes client")
		var ctrlClient client.Client
		ctx, cancel := context.WithCancel(context.Background())
		testEnv, ctrlClient, err = k8smocks.SetupEnvTest(k8smocks.TestEnvInput{
			Users:  k8smocks.DefaultTestUsers,
			Logger: logger,
			Ctx:    ctx,
			Cancel: cancel,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to setup envtest: %w", err)
		}
		k8sFactory, err = k8smocks.NewMockedKubernetesClientFactory(ctrlClient, testEnv, cfg, logger)
	} else {
		k8sFactory, err = k8s.NewKubernetesClientFactory(cfg, logger)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client factory: %w", err)
	}

	// Initialize shared memory store for caching (10 minute cleanup interval)
	memStore := cache.NewMemoryStore()
	logger.Info("Initialized shared memory store")

	// Cache cluster domain at startup using service account
	var clusterDomain string
	if !cfg.MockK8sClient {
		if domain, err := k8s.GetClusterDomainUsingServiceAccount(context.Background(), logger); err != nil {
			logger.Error("Failed to get cluster domain at startup, MaaS autodiscovery will be unavailable", "error", err)
		} else {
			clusterDomain = domain
			logger.Info("Cached cluster domain for MaaS autodiscovery", "domain", clusterDomain)
		}
	}

	app := &App{
		config:                  cfg,
		logger:                  logger,
		repositories:            repositories.NewRepositoriesWithMCP(logger),
		openAPI:                 openAPIHandler,
		kubernetesClientFactory: k8sFactory,
		maasClientFactory:       maasClientFactory,
		dashboardNamespace:      dashboardNamespace,
		memoryStore:             memStore,
		rootCAs:                 rootCAs,
		clusterDomain:           clusterDomain,
	}
	return app, nil
}

func (app *App) Shutdown() error {
	app.logger.Info("shutting down app...")
	// Add any cleanup logic here if needed
	return nil
}

// isAPIRoute checks if the given path is an API route
func (app *App) isAPIRoute(path string) bool {
	return path == constants.HealthCheckPath ||
		path == constants.OpenAPIPath ||
		path == constants.OpenAPIJSONPath ||
		path == constants.OpenAPIYAMLPath ||
		path == constants.SwaggerUIPath ||
		// Match exactly the API path prefix or any sub-path under it
		path == constants.ApiPathPrefix ||
		strings.HasPrefix(path, constants.ApiPathPrefix+"/") ||
		// Match the full gen-ai path prefix
		path == constants.PathPrefix+constants.ApiPathPrefix ||
		strings.HasPrefix(path, constants.PathPrefix+constants.ApiPathPrefix+"/")
}

func (app *App) Routes() http.Handler {
	// Router for /api/v1/*
	apiRouter := httprouter.New()

	apiRouter.NotFound = http.HandlerFunc(app.notFoundResponse)
	apiRouter.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	// MaaS API routes

	// Models (MaaS)
	// apiRouter.GET(constants.MaaSModelsPath, app.AttachNamespace(app.RequireAccessToService(app.AttachMaaSClient(app.MaaSModelsHandler))))

	// // Tokens (MaaS)
	// apiRouter.POST(constants.MaaSTokensPath, app.AttachNamespace(app.RequireAccessToService(app.AttachMaaSClient(app.MaaSIssueTokenHandler))))
	// apiRouter.DELETE(constants.MaaSTokensPath, app.AttachNamespace(app.RequireAccessToService(app.AttachMaaSClient(app.MaaSRevokeAllTokensHandler))))

	// App Router
	appMux := http.NewServeMux()

	// handler for api calls
	appMux.Handle(constants.PathPrefix+constants.ApiPathPrefix+"/", http.StripPrefix(constants.PathPrefix, apiRouter))

	// file server for the frontend file and SPA routes
	staticDir := http.Dir(app.config.StaticAssetsDir)
	fileServer := http.FileServer(staticDir)
	appMux.Handle(constants.ApiPathPrefix+"/", apiRouter)
	appMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctxLogger := helper.GetContextLoggerFromReq(r)

		// Skip API routes
		if (r.URL.Path == "/" || r.URL.Path == "/index.html") ||
			(len(r.URL.Path) > 0 && r.URL.Path[0] == '/' && !app.isAPIRoute(r.URL.Path)) {

			// Check if the requested file exists
			cleanPath := path.Clean(r.URL.Path)
			if _, err := staticDir.Open(cleanPath); err == nil {
				ctxLogger.Debug("Serving static file", slog.String("path", r.URL.Path))
				// Serve the file if it exists
				fileServer.ServeHTTP(w, r)
				return
			}

			// Fallback to index.html for SPA routes
			ctxLogger.Debug("Static asset not found, serving index.html", slog.String("path", r.URL.Path))
			http.ServeFile(w, r, path.Join(app.config.StaticAssetsDir, "index.html"))
			return
		}

		// For API routes, return 404
		http.NotFound(w, r)
	})

	// Create a mux for the healthcheck endpoint
	healthcheckMux := http.NewServeMux()
	healthcheckRouter := httprouter.New()
	healthcheckRouter.GET(constants.HealthCheckPath, app.HealthcheckHandler)
	healthcheckMux.Handle(constants.HealthCheckPath, app.RecoverPanic(app.EnableTelemetry(healthcheckRouter)))

	// Create main mux for all routes
	combinedMux := http.NewServeMux()

	// Health check endpoint (isolated with its own middleware)
	combinedMux.Handle(constants.HealthCheckPath, healthcheckMux)

	// OpenAPI routes (unprotected) - handle these before the main app routes
	combinedMux.HandleFunc(constants.OpenAPIPath, app.openAPI.HandleOpenAPIRedirectWrapper)
	combinedMux.HandleFunc(constants.OpenAPIJSONPath, app.openAPI.HandleOpenAPIJSONWrapper)
	combinedMux.HandleFunc(constants.OpenAPIYAMLPath, app.openAPI.HandleOpenAPIYAMLWrapper)
	combinedMux.HandleFunc(constants.SwaggerUIPath, app.openAPI.HandleSwaggerUIWrapper)

	combinedMux.Handle("/", app.RecoverPanic(app.EnableTelemetry(app.EnableCORS(app.InjectRequestIdentity(appMux)))))

	return combinedMux
}
