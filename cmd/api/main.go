package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"

	"formify/server/internal/auth"
	"formify/server/internal/config"
	"formify/server/internal/database"
	"formify/server/internal/db"
	fileupload "formify/server/internal/file_upload"
	"formify/server/internal/form"
	"formify/server/internal/integrations/google"
	"formify/server/internal/logger"
	customMw "formify/server/internal/middleware"
	"formify/server/internal/response"
	"formify/server/internal/shared"
	"formify/server/internal/user"
)

func main() {
	if err := logger.InitFromEnv(); err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			logger.GetLogger().Error("Failed to close logger", logger.ToField("error", err))
		}
	}()

	log := logger.GetLogger()

	cfg := config.Load()

	if err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Error("Failed to initialize database", logger.ToField("error", err))
		return
	}
	defer database.CloseDB()

	e := echo.New()
	setupMiddleware(e, cfg)
	setupRoutes(e, cfg, log)

	log.Info("Server starting", logger.ToField("port", cfg.Port))
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Error("Failed to start server", logger.ToField("error", err))
	}
}

func setupMiddleware(e *echo.Echo, cfg *config.Config) {
	e.Use(logger.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.GetCORSOrigins(),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
}

func setupRoutes(e *echo.Echo, cfg *config.Config, log *zap.Logger) {
	queries := db.New(database.DBPool)

	userRepo := user.NewRepository(queries)
	formRepo := form.NewRepository(queries)
	responseRepo := response.NewRepository(queries)

	sheetsService := google.InitSheetsService(cfg.GoogleServiceAccountKeyPath, cfg.GoogleServiceAccountKey)

	userService := user.NewService(userRepo)
	formService := form.NewService(formRepo, responseRepo)
	formGetterAdapter := form.NewFormGetterAdapter(formService)
	responseService := response.NewService(responseRepo, sheetsService, formGetterAdapter, userService)
	authService := auth.NewService(cfg.JWTSecret)

	userHandler := user.NewHandler(userService)
	formHandler := form.NewHandler(formService, sheetsService, userService, responseService)
	responseHandler := response.NewHandler(responseService, formService)
	authHandler := auth.NewHandler(authService, userService, cfg.FrontendURL, cfg.IsProduction())

	uploadService, err := fileupload.NewService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary", logger.ToField("error", err))
	}
	uploadHandler := fileupload.NewHandler(uploadService, formService)

	auth.InitProviders(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleCallbackURL,
		cfg.SessionSecret,
	)

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Server is running")
	})
	e.GET("/health", shared.HealthCheck)
	e.GET("/health/db", shared.HealthCheckDB)

	api := e.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/google", authHandler.GoogleLogin)
	authGroup.GET("/google/callback", authHandler.GoogleCallback)

	api.POST("/users", userHandler.CreateUser)
	api.GET("/forms/share/:share_url", formHandler.GetPublicFormsByShareURL)
	api.POST("/forms/:form_id/responses", responseHandler.CreateResponse)
	api.POST("/forms/:form_id/upload", uploadHandler.UploadFile)

	protected := api.Group("")
	protected.Use(customMw.Auth(cfg.JWTSecret))

	protected.GET("/auth/me", authHandler.Me)

	protected.GET("/users/:id", userHandler.GetUser)

	protected.GET("/users/:id/forms", formHandler.GetUserForms)
	protected.POST("/forms", formHandler.CreateForm)
	protected.GET("/forms/:id", formHandler.GetForm)
	protected.PUT("/forms/:id", formHandler.UpdateForm)
	protected.DELETE("/forms/:id", formHandler.DeleteForm)
	protected.POST("/forms/:id/publish", formHandler.PublishForm)
	protected.POST("/forms/:id/unpublish", formHandler.UnpublishForm)
	protected.POST("/forms/:id/sheets/create", formHandler.CreateAndLinkGoogleSheet)
	protected.DELETE("/forms/:id/sheets/link", formHandler.UnlinkGoogleSheet)

	protected.GET("/forms/:id/responses", responseHandler.GetFormResponses)
	protected.GET("/responses/:id", responseHandler.GetResponse)
	protected.DELETE("/responses/:id", responseHandler.DeleteResponse)
}
