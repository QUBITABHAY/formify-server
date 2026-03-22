package main

import (
	"net/http"

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

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	if err := logger.InitFromEnv(); err != nil {
		panic(err)
	}
	defer logger.Close()

	log := logger.GetLogger()

	cfg := config.Load()

	if err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to initialize database", logger.ToField("error", err))
	}
	defer database.CloseDB()

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

	e := echo.New()

	e.Use(logger.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.GetCORSOrigins(),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Server is running")
	})
	e.GET("/health", shared.HealthCheck)
	e.GET("/health/db", shared.HealthCheckDB)

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/google", authHandler.GoogleLogin)
	auth.GET("/google/callback", authHandler.GoogleCallback)

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

	log.Info("Server starting", logger.ToField("port", cfg.Port))
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Error("Failed to start server", logger.ToField("error", err))
	}
}
