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

	externalClient := &apis.OExternal{}
	handlerApi := func(c echo.Context) error {
		return apis.Fetchcall(c, externalClient)
	}
	e.GET("/api", handlerApi)

	handlerData := func(c echo.Context) error {
		return apis.LocResults(c, externalClient)
	}
	e.GET("/data/:lat/:long", handlerData)

	port := getEnv("PORT", "1323")

	e.Logger.Fatal(e.Start("localhost:" + port))
}
