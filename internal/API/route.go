package api

import (
	"fmt"
	"net/http"
	"os"

	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	grpc "github.com/anormalprgrmr/DKV_DB/internal/grpc"
	"github.com/gin-gonic/gin"
	// For access to api.DB
)

type PutRequestPayload struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func InitRoutes(db *dal.DB, r *gin.Engine) *gin.Engine {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/objects", func(c *gin.Context) {
		if !is_master {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Can't Set in Replica",
			})
			return
		}
		var payload PutRequestPayload
		var err error
		if err := c.ShouldBindJSON(&payload); err != nil {
			// If binding fails (missing key or value, wrong format, etc.)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON body",
			})
			return
		}

		tx := db.WriteTx()
		col, err := tx.GetCollection([]byte(dal.DEFAULT_COLLECTION))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "default collection does not exist",
			})
			return
		}

		//1. set
		err = col.Put([]byte(payload.Key), []byte(payload.Value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "can't put key/value in master",
			})
			return
		}
		// 2. replicate
		err = grpc.ReplicaSet(payload.Key, payload.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "can't set in replicas",
			})
			// revert change in master
			tx.Rollback()
			return
		}
		// 3. Commit
		err = tx.Commit()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "can't commit in master",
			})
			os.Exit(-1)

		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	r.GET("/objects/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
			return
		}
		fmt.Printf("key: %v \n", key)
		tx := db.ReadTx()
		col, err := tx.GetCollection([]byte(dal.DEFAULT_COLLECTION))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "default collection does not exist",
			})
			return
		}
		item, err := col.Find([]byte(key))
		if err != nil || item == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			fmt.Println(item)
			c.JSON(http.StatusOK, gin.H{"value": string(item.Value())})
		}
		tx.Rollback()
	})

	return r
}
