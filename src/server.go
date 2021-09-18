package main

import (
	"InShorts/src/apis"
	"os"

	"github.com/labstack/echo-contrib/prometheus"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "InShorts/src/docs" // This line is necessary for go-swagger to find our docs!
)

func getEnv(key, fallback string) string {
	value := os.Getenv("PORT")
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	e := echo.New()

	//Middlewares
	e.Use(middleware.CORS())
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	e.GET("/api", apis.Fetchcall)
	e.GET("/data/:lat/:long", apis.LocResults)

	port := getEnv("PORT", "1323")

	e.Logger.Fatal(e.Start("localhost:" + port))
}
