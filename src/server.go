package main

import (
	"InShorts/src/apis"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/api", apis.Fetchcall)
	e.GET("/users/:lat/:long", apis.UserResults)
	e.Logger.Fatal(e.Start("localhost:1323"))
}
