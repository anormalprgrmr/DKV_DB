package api

import (
	"net/http"

	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	"github.com/gin-gonic/gin"
	// For access to api.DB
)

func InitRoutes(col *dal.Collection, r *gin.Engine) *gin.Engine {

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/put", func(c *gin.Context) {
		key := c.Query("key")
		value := c.Query("value")
		if key == "" || value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key and value required"})
			return
		}
		err := col.Put([]byte(key), []byte(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		}
	})

	r.GET("/get", func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
			return
		}
		item, err := col.Find([]byte(key))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusOK, gin.H{"value": string(item.Value)})
		}
	})

	return r
}
