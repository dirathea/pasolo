package frontend

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(e *echo.Echo, fsEmbed embed.FS) {
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "frontend/build/client",
		Filesystem: http.FS(fsEmbed),
		HTML5:      true,
	}))
}
