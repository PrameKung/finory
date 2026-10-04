package routes

import "github.com/labstack/echo/v5"

// Router is the subset of Echo used to register feature routes. Both Echo and
// Echo groups implement it, so handlers can be mounted below /api/v1.
type Router interface {
	GET(string, echo.HandlerFunc, ...echo.MiddlewareFunc) echo.RouteInfo
	POST(string, echo.HandlerFunc, ...echo.MiddlewareFunc) echo.RouteInfo
	PATCH(string, echo.HandlerFunc, ...echo.MiddlewareFunc) echo.RouteInfo
	DELETE(string, echo.HandlerFunc, ...echo.MiddlewareFunc) echo.RouteInfo
}
