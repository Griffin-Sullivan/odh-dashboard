package constants

const (
	Version = "1.0.0"

	PathPrefix    = "/gen-ai"
	ApiPathPrefix = "/api/v1"

	// API endpoint paths
	HealthCheckPath = "/healthcheck"
	OpenAPIPath     = PathPrefix + "/openapi"
	OpenAPIJSONPath = PathPrefix + "/openapi.json"
	OpenAPIYAMLPath = PathPrefix + "/openapi.yaml"
	SwaggerUIPath   = PathPrefix + "/swagger-ui"

	// General endpoints
	NamespacesPath   = ApiPathPrefix + "/namespaces"
	UserPath         = ApiPathPrefix + "/user"
	ConfigPath       = ApiPathPrefix + "/config"

	// Model as a Service (MaaS) endpoints
	MaaSModelsPath = ApiPathPrefix + "/maas/models"
	MaaSTokensPath = ApiPathPrefix + "/maas/tokens"
)
