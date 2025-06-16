package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	domainServices "github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// HealthHandler maneja los endpoints de health check
type HealthHandler struct {
	config       *config.Config
	logger       logger.Logger
	cacheService domainServices.CacheService
}

// NewHealthHandler crea una nueva instancia del handler de health
func NewHealthHandler(cfg *config.Config, appLogger logger.Logger, cache domainServices.CacheService) *HealthHandler {
	return &HealthHandler{
		config:       cfg,
		logger:       appLogger,
		cacheService: cache,
	}
}

// HealthStatus representa el estado de salud de un componente
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth representa el estado de salud de un componente específico
type ComponentHealth struct {
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// OverallHealth representa el estado general de salud del sistema
type OverallHealth struct {
	Status      HealthStatus                `json:"status"`
	Version     string                      `json:"version"`
	Timestamp   time.Time                   `json:"timestamp"`
	Uptime      time.Duration               `json:"uptime"`
	Environment string                      `json:"environment"`
	Components  map[string]*ComponentHealth `json:"components"`
}

var (
	startTime = time.Now()
)

// Liveness godoc
// @Summary Liveness probe
// @Description Check if the application is alive and running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	h.logger.Info(ctx, "Liveness check requested",
		logger.String("request_id", requestID),
	)

	// Simple liveness check - just return that the service is running
	data := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime),
		"version":   h.config.RESTAPI.Version,
	}

	apiResponse := response.Success(data)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// Readiness godoc
// @Summary Readiness probe
// @Description Check if the application is ready to serve traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse[OverallHealth]
// @Success 503 {object} response.APIResponse[OverallHealth]
// @Router /health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	h.logger.Info(ctx, "Readiness check requested",
		logger.String("request_id", requestID),
	)

	health := h.performHealthChecks(ctx)

	apiResponse := response.Success(health)
	apiResponse.RequestID = requestID

	// Determine HTTP status based on overall health
	statusCode := http.StatusOK
	if health.Status == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded but log warnings
	}

	c.JSON(statusCode, apiResponse)
}

// Health godoc
// @Summary Comprehensive health check
// @Description Get detailed health information about all system components
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse[OverallHealth]
// @Success 503 {object} response.APIResponse[OverallHealth]
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	h.logger.Info(ctx, "Health check requested",
		logger.String("request_id", requestID),
	)

	health := h.performHealthChecks(ctx)

	apiResponse := response.Success(health)
	apiResponse.RequestID = requestID

	// Log health status
	if health.Status == HealthStatusUnhealthy {
		h.logger.Error(ctx, "System health check failed",
			nil,
			logger.String("status", string(health.Status)),
			logger.String("request_id", requestID),
		)
	} else if health.Status == HealthStatusDegraded {
		h.logger.Warn(ctx, "System health degraded",
			logger.String("status", string(health.Status)),
			logger.String("request_id", requestID),
		)
	}

	// Determine HTTP status
	statusCode := http.StatusOK
	if health.Status == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, apiResponse)
}

// performHealthChecks ejecuta todas las verificaciones de salud
func (h *HealthHandler) performHealthChecks(ctx context.Context) *OverallHealth {
	components := make(map[string]*ComponentHealth)

	// Check database
	components["database"] = h.checkDatabase(ctx)

	// Check cache (if configured)
	if h.cacheService != nil {
		components["cache"] = h.checkCache(ctx)
	}

	// Check external APIs
	components["external_api_primary"] = h.checkExternalAPI(ctx, h.config.External.Primary)
	components["external_api_secondary"] = h.checkExternalAPI(ctx, h.config.External.Secondary)

	// Check third-party stock API
	components["stock_api"] = h.checkStockAPI(ctx)

	// Determine overall status
	overallStatus := h.determineOverallStatus(components)

	return &OverallHealth{
		Status:      overallStatus,
		Version:     h.config.RESTAPI.Version,
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime),
		Environment: h.config.App.Env,
		Components:  components,
	}
}

// checkDatabase verifica la conectividad con la base de datos
func (h *HealthHandler) checkDatabase(ctx context.Context) *ComponentHealth {
	start := time.Now()

	// Create a timeout context for the database check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := cockroachdb.NewConnection(h.config)
	if err != nil {
		return &ComponentHealth{
			Status:      HealthStatusUnhealthy,
			Message:     "Failed to connect to database",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	// Test the connection with a simple query
	sqlDB, err := db.DB.DB()
	if err != nil {
		return &ComponentHealth{
			Status:      HealthStatusUnhealthy,
			Message:     "Failed to get underlying SQL DB",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	if err := sqlDB.PingContext(checkCtx); err != nil {
		return &ComponentHealth{
			Status:      HealthStatusUnhealthy,
			Message:     "Database ping failed",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	return &ComponentHealth{
		Status:      HealthStatusHealthy,
		Message:     "Database connection successful",
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details: map[string]interface{}{
			"host": h.config.Database.Host,
			"port": h.config.Database.Port,
			"name": h.config.Database.Name,
		},
	}
}

// checkCache verifica la conectividad con el servicio de cache
func (h *HealthHandler) checkCache(ctx context.Context) *ComponentHealth {
	start := time.Now()

	if h.cacheService == nil {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Cache service not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	// Create a timeout context for the cache check
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Use the Ping method to test cache connectivity
	if err := h.cacheService.Ping(checkCtx); err != nil {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Cache ping failed",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	// Get cache statistics for additional health info
	stats, err := h.cacheService.GetStats(checkCtx)
	if err != nil {
		// Cache is responding to ping but stats failed - still considered healthy
		return &ComponentHealth{
			Status:      HealthStatusHealthy,
			Message:     "Cache responding but stats unavailable",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"host":        h.config.Cache.Host,
				"port":        h.config.Cache.Port,
				"stats_error": err.Error(),
			},
		}
	}

	return &ComponentHealth{
		Status:      HealthStatusHealthy,
		Message:     "Cache operations successful",
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details: map[string]interface{}{
			"host":  h.config.Cache.Host,
			"port":  h.config.Cache.Port,
			"stats": stats,
		},
	}
}

// checkExternalAPI verifica la conectividad con APIs externas
func (h *HealthHandler) checkExternalAPI(ctx context.Context, apiConfig config.APIConfig) *ComponentHealth {
	start := time.Now()

	if apiConfig.BaseURL == "" {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "External API not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	// Create a timeout context for the API check
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create HTTP client for basic connectivity test
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(checkCtx, "GET", apiConfig.BaseURL, nil)
	if err != nil {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Failed to create request for external API",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
				"api":   apiConfig.Name,
			},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "External API unreachable",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error": err.Error(),
				"api":   apiConfig.Name,
				"url":   apiConfig.BaseURL,
			},
		}
	}
	defer resp.Body.Close()

	// Consider 2xx and 4xx as healthy (4xx means API is responding)
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return &ComponentHealth{
			Status:      HealthStatusHealthy,
			Message:     "External API responding",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"api":         apiConfig.Name,
				"url":         apiConfig.BaseURL,
				"status_code": resp.StatusCode,
			},
		}
	}

	return &ComponentHealth{
		Status:      HealthStatusDegraded,
		Message:     "External API returned server error",
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details: map[string]interface{}{
			"api":         apiConfig.Name,
			"url":         apiConfig.BaseURL,
			"status_code": resp.StatusCode,
		},
	}
}

// checkStockAPI verifica la conectividad con la API de stocks
func (h *HealthHandler) checkStockAPI(ctx context.Context) *ComponentHealth {
	start := time.Now()

	if h.config.ThirdStockAPI.BaseURL == "" {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Stock API not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	if h.config.ThirdStockAPI.Auth == "" {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Stock API auth token not configured",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	// Create a timeout context for the API check
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Use the stock API client with proper authentication
	client := stock_api.NewClient(h.config)

	// Test the client's connectivity using the health check method
	healthResult := client.HealthCheck(checkCtx)

	if healthResult.Error != nil {
		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     "Stock API unreachable or network error",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"error":       healthResult.Error.Error(),
				"api":         h.config.ThirdStockAPI.Name,
				"url":         h.config.ThirdStockAPI.BaseURL,
				"status_code": healthResult.StatusCode,
			},
		}
	}

	// Check if the API is healthy based on status code
	if !healthResult.IsHealthy {
		message := "Stock API responding but with error status"
		if healthResult.StatusCode == 401 || healthResult.StatusCode == 403 {
			message = "Stock API authentication failed - check bearer token"
		}

		return &ComponentHealth{
			Status:      HealthStatusDegraded,
			Message:     message,
			LastChecked: time.Now(),
			Duration:    time.Since(start),
			Details: map[string]interface{}{
				"api":         h.config.ThirdStockAPI.Name,
				"url":         h.config.ThirdStockAPI.BaseURL,
				"status_code": healthResult.StatusCode,
			},
		}
	}

	return &ComponentHealth{
		Status:      HealthStatusHealthy,
		Message:     "Stock API responding with authentication",
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details: map[string]interface{}{
			"api":         h.config.ThirdStockAPI.Name,
			"url":         h.config.ThirdStockAPI.BaseURL,
			"status_code": healthResult.StatusCode,
			"client":      "authenticated",
		},
	}
}

// determineOverallStatus determina el estado general basado en los componentes
func (h *HealthHandler) determineOverallStatus(components map[string]*ComponentHealth) HealthStatus {
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, component := range components {
		switch component.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	// If any critical component (database) is unhealthy, system is unhealthy
	if db, exists := components["database"]; exists && db.Status == HealthStatusUnhealthy {
		return HealthStatusUnhealthy
	}

	// If more than half of components are unhealthy, system is unhealthy
	totalComponents := len(components)
	if totalComponents > 0 && unhealthyCount > totalComponents/2 {
		return HealthStatusUnhealthy
	}

	// If any component is degraded or unhealthy, system is degraded
	if degradedCount > 0 || unhealthyCount > 0 {
		return HealthStatusDegraded
	}

	// All components are healthy
	return HealthStatusHealthy
}
