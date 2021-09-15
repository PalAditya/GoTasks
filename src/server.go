package main

import (
	"InShorts/src/apis"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "InShorts/src/docs" // This line is necessary for go-swagger to find our docs!
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Use(middleware.CORS())
	e.GET("/api", apis.Fetchcall)
	e.GET("/data/:lat/:long", apis.LocResults)

	port := getEnv("PORT", "1323")

	e.Logger.Fatal(e.Start("localhost:" + port))
}
