package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// CORSMiddleware creates a CORS middleware using the provided configuration
func CORSMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	if !corsConfig.Enabled {
		// Return a no-op middleware if CORS is disabled
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	// Configure gin-contrib/cors
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
		AllowWildcard:    corsConfig.AllowWildcard,

		// Custom origin validation if needed
		AllowOriginFunc: func(origin string) bool {
			return corsConfig.IsOriginAllowed(origin)
		},
	})

	return corsMiddleware
}

// SimpleCORSMiddleware creates a simple CORS middleware for development
func SimpleCORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow common development origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:4200",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:4200",
		}

		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Length,Content-Type,Accept,Authorization,X-Requested-With,X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID,X-Response-Time")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", strconv.Itoa(int((12 * time.Hour).Seconds())))

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// ProductionCORSMiddleware creates a strict CORS middleware for production
func ProductionCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		if !originAllowed && origin != "" {
			// Log unauthorized origin attempt
			c.Header("X-CORS-Error", "Origin not allowed")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CORS policy violation",
				"code":  "ORIGIN_NOT_ALLOWED",
			})
			return
		}

		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Length,Content-Type,Accept,Authorization,X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID,X-Response-Time")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", strconv.Itoa(int((24 * time.Hour).Seconds())))

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// validateOrigin checks if an origin matches allowed patterns
func validateOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}
