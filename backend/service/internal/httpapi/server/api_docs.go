package server

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v5"
)

//go:embed openapi.json
var openAPIDocument []byte

const scalarAPIReference = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Finory API Reference</title>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.70.0"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.json',
        pageTitle: 'Finory API Reference',
        agent: { disabled: true }
      })
    </script>
  </body>
</html>`

func registerAPIDocs(e *echo.Echo) {
	e.GET("/openapi.json", func(c *echo.Context) error {
		return c.Blob(http.StatusOK, "application/json; charset=utf-8", openAPIDocument)
	})
	docs := func(c *echo.Context) error {
		return c.HTML(http.StatusOK, scalarAPIReference)
	}
	e.GET("/docs", docs)
	e.GET("/docs/", docs)
}
