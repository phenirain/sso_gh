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
	"github.com/phenirain/sso/internal/application/client"
	"github.com/phenirain/sso/internal/application/manager"
	"github.com/phenirain/sso/internal/config"
	grpcService "github.com/phenirain/sso/internal/services/grpc"
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
	productGroup.POST("/by-article", productHandler.GetProductByArticle)
	productGroup.GET("", productHandler.GetProducts)
	productGroup.DELETE("", productHandler.DeleteProduct)

	// Order routes
	orderHandler := adminOrder.NewOrderHandler(orderService)
	orderGroup := adminGroup.Group("/order")
	orderGroup.GET("/statuses", orderHandler.GetOrderStatuses)
	orderGroup.GET("/clients", orderHandler.GetOrderClients)
	orderGroup.GET("/products", orderHandler.GetOrderProducts)
	orderGroup.POST("", orderHandler.CreateOrUpdateOrder)
	orderGroup.POST("/by-id", orderHandler.GetOrderById)
	orderGroup.DELETE("", orderHandler.DeleteOrder)

	// Client routes
	clientHandler := adminClient.NewClientHandler(clientService)
	clientGroup := adminGroup.Group("/client")
	clientGroup.GET("/users", clientHandler.GetUsers)
	clientGroup.POST("", clientHandler.CreateClient)
	clientGroup.GET("", clientHandler.GetClients)
	clientGroup.DELETE("", clientHandler.DeleteClient)

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
	// Создаем wrapper для сервисов
	clientServiceWrapper := grpcService.NewClientServiceWrapper(
		clientServiceClient,
		productServiceClient,
		orderServiceClient,
	)

	clientHandler := client.NewHandler(clientServiceWrapper)
	clientGroup := e.Group("/client")

	// Client profile routes
	clientGroup.POST("/register", clientHandler.RegisterClient)
	clientGroup.POST("/profile", clientHandler.FillClientProfile)
	clientGroup.GET("/profile", clientHandler.GetClientProfile)
	clientGroup.DELETE("", clientHandler.DeleteClient)

	// Product routes
	productGroup := clientGroup.Group("/product")
	productGroup.POST("/base-models", clientHandler.GetAllBaseModels)
	productGroup.POST("", clientHandler.GetProducts)
	productGroup.POST("/by-article", clientHandler.GetProduct)
	productGroup.POST("/favorites", clientHandler.ActionProductToFavorites)
	productGroup.GET("/favorites", clientHandler.GetFavoriteProducts)

	// Order routes
	orderGroup := clientGroup.Group("/order")
	orderGroup.POST("", clientHandler.CreateOrder)
	orderGroup.POST("/complete", clientHandler.CompleteOrder)
	orderGroup.POST("/add-product", clientHandler.AddProductToOrder)
	orderGroup.GET("", clientHandler.GetClientOrders)
	orderGroup.POST("/by-id", clientHandler.GetOrderById)
	orderGroup.POST("/cancel", clientHandler.CancelOrder)

	// Orders list route
	clientGroup.GET("/orders", clientHandler.GetClientOrders)
}

func registerManagerRoutes(e *echo.Echo, managerServiceClient pbManager.ManagerServiceClient) {
	// Создаем wrapper для сервиса
	managerServiceWrapper := grpcService.NewManagerServiceWrapper(managerServiceClient)

	managerHandler := manager.NewHandler(managerServiceWrapper)
	managerGroup := e.Group("/manager")

	// Order routes
	orderGroup := managerGroup.Group("/order")
	orderGroup.GET("", managerHandler.GetAllOrders)
	orderGroup.POST("/by-id", managerHandler.GetOrderById)
	orderGroup.POST("/give", managerHandler.GiveOrder)
	orderGroup.POST("/cancel", managerHandler.CancelOrder)

	// Orders list route
	managerGroup.GET("/orders", managerHandler.GetAllOrders)
}
