package main

import (
	"github.com/gin-gonic/gin"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
)

func main() {
	app, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	r := gin.New()

	r.Run(":8080")
}
