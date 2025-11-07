package main

import (
	"os"

	api "github.com/anormalprgrmr/DKV_DB/internal/API"
	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
)

func main() {

	options := &dal.Options{
		PageSize:       os.Getpagesize(),
		MinFillPercent: 0.0125,
		MaxFillPercent: 0.025,
	}
	db_path, exists := os.LookupEnv("DB_PATH")
	if !exists {
    		db_path = "./mainTest"
	}

	db, _ := dal.GetDal(db_path, options)
	defer db.Close()

	c := dal.NewCollection([]byte("collection1"), db.Root)
	c.DAL = db
	api.RunServer(c, 8080)

	// dal, _ := dal.GetDal("mainTest")

	// node, _ := dal.GetNode(dal.Root)
	// node.DAL = dal
	// index, containingNode, _ := node.FindKey([]byte("Key1"))
	// res := containingNode.Items[index]

	// fmt.Printf("\n key is: %s, value is: %s", res.Key, res.Value)
	// // Close the db
	// _ = dal.Close()
}
