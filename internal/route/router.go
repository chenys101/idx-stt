package route

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"stt/internal/controller"
	"stt/internal/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	r.Use(cors.New(config))
	r.Use(middleware.ContentNegotiation())

	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			uc := controller.UserController{}
			users.POST("", uc.CreateUser)
			users.GET("/:id", uc.GetUser)
			users.GET("", uc.ListUsers)
		}
		stocks := api.Group("/std")
		{
			sc := controller.StockController{}
			stocks.GET("", sc.GetStockData)
			stocks.GET("/nc", sc.GetStockDataNotEncrypt)
			stocks.POST("/createStdMonitor", sc.CreateStockMonitor)
			stocks.POST("/updateStdMonitor", sc.UpdateStockMonitor)
			stocks.GET("/getAllStdMonitors", sc.GetAllStockMonitors)
		}
	}

	return r
}
