// Package shared contains common helpers used across handlers and services.
package shared

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"formify/server/internal/database"
)

func RespondError(c *echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"error": message})
}

func HealthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func HealthCheckDB(c *echo.Context) error {
	ctx := c.Request().Context()
	if database.DBPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_disconnected"})
	}
	if err := database.DBPool.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_disconnected", "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "db": "connected"})
}
