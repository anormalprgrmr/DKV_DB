package api

import (
	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	"github.com/gin-gonic/gin"
)

var DB *dal.DAL // Global DB instance used by handlers

func RunServer(PORT int) {
	DB, _ = dal.GetDal("db.db")
	r := gin.Default()

	InitRoutes(r)

	r.Run(":8180") // Bind explicitly, this ignores PORT param for now
}
