package api

import (
	"github.com/gin-gonic/gin"
)

func RunServer(PORT int) {
	r := gin.Default()

	InitRoutes(r)

	r.Run()
}
