package delivery

import (
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/delivery/handlers"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(productUseCase usecase.ProductUseCaseInterface, logger *zap.Logger) *Router {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	v1 := engine.Group("/api/v1")
	productHandler := handlers.NewProductHandler(productUseCase, logger)

	v1.POST("/products", productHandler.UpsertProductWithSizes)
	v1.GET("/products", productHandler.ListProductsForMonitoring)
	v1.GET("/products/nm/:nm_id", productHandler.GetProductByNmID)
	v1.GET("/product-sizes/option/:option_id", productHandler.GetProductSizeByOptionID)

	return &Router{engine: engine}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(port)
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
