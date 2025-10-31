package api

import (
	"os"

	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	"github.com/gin-gonic/gin"
)

var DB *dal.DAL // Global DB instance used by handlers

func RunServer(c *dal.Collection, PORT int) {
	options := &dal.Options{
		PageSize:       os.Getpagesize(),
		MinFillPercent: 0.0125,
		MaxFillPercent: 0.025,
	}
	DB, _ = dal.GetDal("db.db", options)
	r := gin.Default()

	InitRoutes(c, r)

	r.Run(":8180") // Bind explicitly, this ignores PORT param for now
}
