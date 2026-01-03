package api

import (
	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	"github.com/gin-gonic/gin"
)

//var DB *dal.DAL // Global DB instance used by handlers
var is_master bool
func RunServer(db *dal.DB, port string, _is_master bool) {
	is_master  = _is_master
	r := gin.Default()

	InitRoutes(db, r)

	r.Run(port) // Bind explicitly, this ignores PORT param for now
}
