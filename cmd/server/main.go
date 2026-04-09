package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/handlers"
	"github.com/wind7vn/fnb_be/internal/repositories"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/cache"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/common/response"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/db"
)

func main() {
	// 1. Load Configurations
	config.LoadConfig()

	// 2. Initialize Zap Logger
	logger.InitLogger(config.AppConfig.Env)
	defer logger.Sync()

	logger.Log.Info("Starting F&B Management Backend...")

	// 3. Connect DB
	db.ConnectDB()
	db.DB.AutoMigrate(&domain.User{}, &domain.Tenant{}, &domain.TenantMember{})
	cache.ConnectRedis()

	// 4. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "F&B Multi-Tenant API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Log.Error("Unhandled HTTP error: " + err.Error())
			return response.Error(c, errors.NewInternalServer(err))
		},
	})

	// Add CORS middleware to allow FE to connect
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: false,
	}))

	// --- Dependency Injection Engine --- //
	userRepo := repositories.NewUserRepository(db.DB)
	tenantRepo := repositories.NewTenantRepository(db.DB)
	memberRepo := repositories.NewTenantMemberRepository(db.DB)
	
	authService := services.NewAuthService(userRepo, tenantRepo, memberRepo)
	authHandler := handlers.NewAuthHandler(authService)

	aiService := services.NewAIService(config.AppConfig.GeminiApiKey)
	tenantService := services.NewTenantService(tenantRepo, userRepo, memberRepo)
	tenantHandler := handlers.NewTenantHandler(tenantService, aiService)

	productRepo := repositories.NewProductRepository(db.DB)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	tableRepo := repositories.NewTableRepository(db.DB)
	tableService := services.NewTableService(tableRepo)
	tableHandler := handlers.NewTableHandler(tableService)

	orderRepo := repositories.NewOrderRepository(db.DB)
	actionLogRepo := repositories.NewActionLogRepository(db.DB)
	notiRepo := repositories.NewNotificationRepository(db.DB)

	pubSubService := services.NewPubSubService()
	pushNotiService := services.NewNotificationService()
	systemService := services.NewSystemService(actionLogRepo, notiRepo, pushNotiService)
	orderService := services.NewOrderService(orderRepo, productRepo, tableRepo, pubSubService, systemService)
	orderHandler := handlers.NewOrderHandler(orderService)
	systemHandler := handlers.NewSystemHandler(systemService)

	wsHandler := handlers.NewWSHandler(pubSubService)

	// API Version Group
	v1 := app.Group("/api/v1")
	authHandler.SetupRoutes(v1)
	tenantHandler.SetupRoutes(v1)
	productHandler.SetupRoutes(v1)
	tableHandler.SetupRoutes(v1)
	orderHandler.SetupRoutes(v1)
	wsHandler.SetupRoutes(v1)
	systemHandler.SetupRoutes(v1)

	// Setup basic ping route
	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.Success(c, map[string]string{
			"status":  "ok",
			"message": "F&B API is running smoothly",
		})
	})

	// 5. Serve Flutter Web Static Files
	// Lấy từ .env nếu có, ngược lại lấy thư mục web ở cạnh file chạy
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		if config.AppConfig.Env == "development" || config.AppConfig.Env == "" {
			// Point to frontend build folder for local development
			webDir = "../fnb_ui/build/web" 
		} else {
			webDir = "./web" // Fallback khi chạy trên production server
		}
	}

	if _, err := os.Stat(webDir); err == nil {
		logger.Log.Info("Serving Flutter App from: " + webDir)
		app.Static("/", webDir)
		// Fallback for Flutter Router (must be AFTER API routes)
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile(webDir + "/index.html")
		})
	} else {
		logger.Log.Warn("Static web directory not found: " + webDir)
	}

	// Start server on defined port or default to 8080
	port := config.AppConfig.Port
	if port == "" {
		port = "8080"
	}

	logger.Log.Info("Server listening on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
