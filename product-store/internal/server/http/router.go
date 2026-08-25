package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(readiness *ReadinessChecker, registry *prometheus.Registry) *Router {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	engine.GET("/ready", func(c *gin.Context) {
		if readiness == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"error":  "readiness checker is not configured",
			})
			return
		}

		ready, lastErr := readiness.IsReady()
		if !ready {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"error":  lastErr,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	if registry != nil {
		engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	}

	return &Router{engine: engine}
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
