package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/sw5005-sus/ceramicraft-commodity-mservice/server/docs"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/api"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/metrics"
	"github.com/sw5005-sus/ceramicraft-user-mservice/common/middleware"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

const (
	servicePrefix = "/product-ms/v1"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	baseRouter := r.Group(servicePrefix)
	{

		baseRouter.Use(metrics.MetricsMiddleware())
		baseRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

		// swagger router
		baseRouter.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

		baseRouter.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		merchantRouter := baseRouter.Group("/merchant")
		{
			merchantRouter.Use(middleware.AuthMiddleware())
			merchantRouter.POST("/products", middleware.RequireRoles("merchant_admin", "product_editor"), api.AddProduct)
			merchantRouter.GET("/product/:id", api.GetProductMerchant)
			merchantRouter.PATCH("/products/:id/status", middleware.RequireRoles("merchant_admin", "product_editor"), api.UpdateProductStatus)
			merchantRouter.PATCH("/products/:id/stock", middleware.RequireRoles("merchant_admin", "product_editor"), api.UpdateProductStock)
			merchantRouter.POST("/images/upload-urls", api.GetImageUploadPresignURL)
			merchantRouter.GET("/products", api.GetMerchantProductList)
			merchantRouter.PUT("/products/:id", middleware.RequireRoles("merchant_admin", "product_editor"), api.EditProductInfo)
			merchantRouter.POST("/products/:id/review/:decision", middleware.RequireRoles("merchant_admin", "product_reviewer"), api.UpdateProductReviewResult)
		}

		customerRouter := baseRouter.Group("/customer")
		{
			customerRouter.GET("/products", api.GetCustomerProductList)
			customerRouter.GET("/product/:id", api.GetProductCustomer)

			authed := customerRouter.Group("")
			{
				authed.Use(middleware.AuthMiddleware())
				authed.GET("/cart", api.GetUserCartInfo)
				authed.POST("/cart/items", api.CreateCartItem)
				authed.PUT("/cart/items/:item_id", api.UpdateCartItem)
				authed.DELETE("/cart/items/:item_id", api.DeleteCartItem)
				authed.GET("/cart/selected-num", api.GetCartSelctedNum)
				authed.GET("/cart/price-estimate", api.GetEstimatePrice)
			}
		}
	}
	return r
}
