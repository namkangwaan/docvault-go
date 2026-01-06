package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	
	"github.com/namkangwaan/docvault-go/internal/config"
	"github.com/namkangwaan/docvault-go/internal/database"
	"github.com/namkangwaan/docvault-go/internal/handlers"
	"github.com/namkangwaan/docvault-go/internal/middleware"
	"github.com/namkangwaan/docvault-go/internal/repository"
	"github.com/namkangwaan/docvault-go/internal/security"
	"github.com/namkangwaan/docvault-go/internal/services"
	"github.com/namkangwaan/docvault-go/internal/storage"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Initialize storage
	nfsStorage, err := storage.NewNFSStorage(&cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	log.Println("Storage initialized successfully")

	// Initialize security validator
	validator, err := security.NewValidator(&cfg.Security)
	if err != nil {
		log.Fatalf("Failed to initialize security validator: %v", err)
	}

	log.Println("Security validator initialized successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	documentRepo := repository.NewDocumentRepository(db.Pool)
	fileRepo := repository.NewFileRepository(db.Pool)

	// Initialize services
	authService := services.NewAuthService(userRepo, &cfg.JWT)
	categoryService := services.NewCategoryService(categoryRepo)
	documentService := services.NewDocumentService(documentRepo, fileRepo)
	fileService := services.NewFileService(fileRepo, nfsStorage, validator)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	fileHandler := handlers.NewFileHandler(fileService, nfsStorage)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		BodyLimit:    int(cfg.Storage.MaxFileSize),
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CORSMiddleware())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Health(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
		}
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// Public API routes
	api := app.Group("/api")

	// Public document routes
	api.Get("/documents", documentHandler.ListDocuments)
	api.Get("/documents/:id", documentHandler.GetDocument)

	// Public category routes
	api.Get("/categories", categoryHandler.ListCategories)

	// Public file routes
	api.Get("/files/:id/download", fileHandler.DownloadFile)
	api.Get("/files/:id/preview", fileHandler.PreviewFile)

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)

	// Admin routes (protected with JWT)
	admin := api.Group("/admin", middleware.AuthMiddleware(authService))

	// Admin document routes
	admin.Post("/documents", documentHandler.CreateDocument)
	admin.Put("/documents/:id", documentHandler.UpdateDocument)
	admin.Delete("/documents/:id", documentHandler.DeleteDocument)

	// Admin category routes
	admin.Post("/categories", categoryHandler.CreateCategory)
	admin.Put("/categories/:id", categoryHandler.UpdateCategory)
	admin.Delete("/categories/:id", categoryHandler.DeleteCategory)

	// Admin file routes
	admin.Post("/files/upload", fileHandler.UploadFile)
	admin.Post("/files/upload/chunked/init", fileHandler.InitChunkedUpload)
	admin.Post("/files/upload/chunked/chunk", fileHandler.UploadChunk)
	admin.Post("/files/upload/chunked/finalize", fileHandler.FinalizeChunkedUpload)
	admin.Get("/files/upload/chunked/status/:session_id", fileHandler.GetUploadStatus)
	admin.Delete("/files/:id", fileHandler.DeleteFile)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
