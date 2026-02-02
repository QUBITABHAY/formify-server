package main

import (
	"log"
	"net/http"
	"os"

	"formify/server/internal/database"
	"formify/server/internal/db"
	"formify/server/internal/form"
	"formify/server/internal/response"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	queries := db.New(database.DBPool)

	formRepo := form.NewRepository(queries)
	responseRepo := response.NewRepository(queries)
	formService := form.NewService(formRepo)
	responseService := response.NewService(responseRepo)
	formHandler := form.NewHandler(formService)
	responseHandler := response.NewHandler(responseService)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Server is running")
	})
	api := e.Group("/api")

	api.POST("/responses/:form_id", responseHandler.CreateResponse)

	protected := api.Group("")
	// Uncomment the following line to enable authentication middleware
	// protected.Use(customMw.Auth)

	protected.GET("/users/:id/forms", formHandler.GetUserForms)

	protected.POST("/forms", formHandler.CreateForm)
	protected.GET("/forms/:id", formHandler.GetForm)
	protected.PUT("/forms/:id", formHandler.UpdateForm)
	protected.DELETE("/forms/:id", formHandler.DeleteForm)
	protected.POST("/forms/:id/publish", formHandler.PublishForm)
	protected.POST("/forms/:id/unpublish", formHandler.UnpublishForm)
	protected.GET("/forms/:id/responses", responseHandler.GetFormResponses)

	protected.GET("/responses/:id", responseHandler.GetResponse)
	protected.DELETE("/responses/:id", responseHandler.DeleteResponse)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
