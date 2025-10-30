package application

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/phenirain/sso/docs"
	adminClient "github.com/phenirain/sso/internal/application/admin/client"
	adminOrder "github.com/phenirain/sso/internal/application/admin/order"
	adminProduct "github.com/phenirain/sso/internal/application/admin/product"
	adminReport "github.com/phenirain/sso/internal/application/admin/report"
	"github.com/phenirain/sso/internal/application/auth"
	clientClient "github.com/phenirain/sso/internal/application/client/client"
	clientOrder "github.com/phenirain/sso/internal/application/client/order"
	clientProduct "github.com/phenirain/sso/internal/application/client/product"
	manager "github.com/phenirain/sso/internal/application/manager"
	"github.com/phenirain/sso/internal/config"
	"github.com/phenirain/sso/pkg/echomiddleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	pbAdmin "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	pbClient "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
	pbManager "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupHTTPServer(cfg *config.Config, authService auth.AuthService, jwt echomiddleware.Jwt, log *slog.Logger) (*echo.Echo, error) {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echomiddleware.JwtValidation(jwt))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// Swagger UI endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.GET("/v", func(c echo.Context) error {
		return c.String(http.StatusOK, "JWT IS VALID")
	})

	// Создание gRPC клиентов

	// Admin gRPC клиенты
	adminConn, err := grpc.NewClient(cfg.GRPC.Admin, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to admin gRPC server", slog.String("address", cfg.GRPC.Admin), slog.String("error", err.Error()))
		return nil, err
	}

	adminClientService := pbAdmin.NewClientServiceClient(adminConn)
	adminProductService := pbAdmin.NewProductServiceClient(adminConn)
	adminOrderService := pbAdmin.NewOrderServiceClient(adminConn)
	adminReportService := pbAdmin.NewReportServiceClient(adminConn)

	// Client gRPC клиенты
	clientConn, err := grpc.NewClient(cfg.GRPC.Client, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to client gRPC server", slog.String("address", cfg.GRPC.Client), slog.String("error", err.Error()))
		return nil, err
	}

	clientClientService := pbClient.NewClientServiceClient(clientConn)
	clientProductService := pbClient.NewProductServiceClient(clientConn)
	clientOrderService := pbClient.NewOrderServiceClient(clientConn)

	// Manager gRPC клиенты
	managerConn, err := grpc.NewClient(cfg.GRPC.Manager, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to manager gRPC server", slog.String("address", cfg.GRPC.Manager), slog.String("error", err.Error()))
		return nil, err
	}

	managerManagerService := pbManager.NewManagerServiceClient(managerConn)

	log.Info("gRPC clients initialized successfully")

	registerAuthRoutes(e, authService)
	registerAdminRoutes(e, adminClientService, adminProductService, adminOrderService, adminReportService)
	registerClientRoutes(e, clientClientService, clientProductService, clientOrderService)
	registerManagerRoutes(e, managerManagerService)

	return e, nil
}

func registerAuthRoutes(e *echo.Echo, authService auth.AuthService) {
	authHandler := auth.NewHandler(authService)
	auth := e.Group("/auth")
	auth.POST("/logIn", authHandler.LogIn)
	auth.POST("/signUp", authHandler.SignUp)
	auth.POST("/refresh", authHandler.Refresh)
}

func registerAdminRoutes(
	e *echo.Echo,
	clientService pbAdmin.ClientServiceClient,
	productService pbAdmin.ProductServiceClient,
	orderService pbAdmin.OrderServiceClient,
	reportService pbAdmin.ReportServiceClient,
) {
	adminGroup := e.Group("/admin")

	// Product routes
	productHandler := adminProduct.NewProductHandler(productService)
	productGroup := adminGroup.Group("/product")
	productGroup.POST("/base-model", productHandler.CreateOrUpdateBaseModel)
	productGroup.POST("/base-models", productHandler.GetAllBaseModels)
	productGroup.DELETE("/base-model", productHandler.DeleteBaseModel)
	productGroup.POST("", productHandler.CreateOrUpdateProduct)
	productGroup.GET(":article", productHandler.GetProductByArticle)
	productGroup.GET("", productHandler.GetProducts)
	productGroup.DELETE(":article", productHandler.DeleteProduct)

	// Order routes
	orderHandler := adminOrder.NewOrderHandler(orderService)
	orderGroup := adminGroup.Group("/order")
	orderGroup.GET("/statuses", orderHandler.GetOrderStatuses)
	orderGroup.GET("/clients", orderHandler.GetOrderClients)
	orderGroup.GET("/products", orderHandler.GetOrderProducts)
	orderGroup.POST("", orderHandler.CreateOrUpdateOrder)
	orderGroup.GET(":id", orderHandler.GetOrderById)
	orderGroup.DELETE(":id", orderHandler.DeleteOrder)

	// Client routes
	clientHandler := adminClient.NewClientHandler(clientService)
	clientGroup := adminGroup.Group("/client")
	clientGroup.GET("/users", clientHandler.GetUsers)
	clientGroup.POST("", clientHandler.CreateClient)
	clientGroup.GET("", clientHandler.GetClients)
	clientGroup.DELETE(":id", clientHandler.DeleteClient)

	// Report routes
	reportHandler := adminReport.NewReportHandler(reportService)
	reportGroup := adminGroup.Group("/report")
	reportGroup.POST("/orders-by-time", reportHandler.GetAmountOfOrdersByTimeOfDay)
	reportGroup.POST("/purchases-by-brands", reportHandler.GetPurchasesByBrands)
	reportGroup.POST("/average-processing-time", reportHandler.GetAverageOrderProcessingTime)

	// Orders list route
	adminGroup.POST("/orders", orderHandler.GetOrders)
}

func registerClientRoutes(
	e *echo.Echo,
	clientServiceClient pbClient.ClientServiceClient,
	productServiceClient pbClient.ProductServiceClient,
	orderServiceClient pbClient.OrderServiceClient,
) {
	clientGroup := e.Group("/client")

	// Client profile routes
	cHandler := clientClient.NewClientHandler(clientServiceClient)
	clientGroup.POST("/register", cHandler.RegisterClient)
	clientGroup.POST("/profile", cHandler.FillClientProfile)
	clientGroup.GET("/profile/:id", cHandler.GetClientProfile)
	clientGroup.DELETE(":id", cHandler.DeleteClient)

	// Product routes
	pHandler := clientProduct.NewProductHandler(productServiceClient)
	productGroup := clientGroup.Group("/product")
	productGroup.POST("/base-models", pHandler.GetAllBaseModels)
	productGroup.POST("", pHandler.GetProducts)
	productGroup.GET(":article", pHandler.GetProduct)
	productGroup.POST("/favorites", pHandler.ActionProductToFavorites)
	productGroup.GET(":id/favorites", pHandler.GetFavoriteProducts)

	// Order routes
	oHandler := clientOrder.NewOrderHandler(orderServiceClient)
	orderGroup := clientGroup.Group("/order")
	orderGroup.POST("", oHandler.CreateOrder)
	orderGroup.POST("/complete", oHandler.CompleteOrder)
	orderGroup.POST("/add-product", oHandler.AddProductToOrder)
	orderGroup.GET("", oHandler.GetClientOrders)
	orderGroup.GET(":id", oHandler.GetOrderById)
	orderGroup.POST(":id/cancel", oHandler.CancelOrder)

	// Orders list route
	clientGroup.GET("/orders", oHandler.GetClientOrders)
}

func registerManagerRoutes(e *echo.Echo, managerServiceClient pbManager.ManagerServiceClient) {
	managerGroup := e.Group("/manager")

	// Order routes
	oHandler := manager.NewOrderHandler(managerServiceClient)
	orderGroup := managerGroup.Group("/order")
	orderGroup.GET("", oHandler.GetAllOrders)
	orderGroup.GET(":id", oHandler.GetOrderById)
	orderGroup.POST("/give", oHandler.GiveOrder)
	orderGroup.POST(":id/cancel", oHandler.CancelOrder)

	// Orders list route
	managerGroup.GET("/orders", oHandler.GetAllOrders)
}
