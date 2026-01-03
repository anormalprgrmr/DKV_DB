package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	api "github.com/anormalprgrmr/DKV_DB/internal/API"
	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
	grpc "github.com/anormalprgrmr/DKV_DB/internal/grpc"
)
const (
	DEFAULT_HTTP_PORT = 8080
	DEFAULT_GRPC_PORT = 50000
)
type ReplicaConfig struct {
	Address  string `json:"address"`
}
func LoadConfig(filename string) ([]ReplicaConfig, error) {
	// 1. Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// 2. Read file contents
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 3. Unmarshal JSON into slice
	var configs []ReplicaConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return configs, nil
}

var is_master bool
func main() {

	options := &dal.Options{
		//PageSize:       os.Getpagesize(),
		MinFillPercent: 0.0125,
		MaxFillPercent: 0.025,
	}
	db_path, exists := os.LookupEnv("DB_PATH")
	if !exists {
    		db_path = "./mainTest2"
	}

	db, _ := dal.Open(db_path, options)
	defer db.Close()
	
	tx := db.WriteTx()
	_ , _ = tx.CreateCollection([]byte(dal.DEFAULT_COLLECTION))
	_ = tx.Commit()

	replica_configs, err := LoadConfig("replica.json")
	if err != nil {
		fmt.Println("Error loading replica config:", err)
		return
	}
	for _, cfg := range replica_configs {
		log.Println(cfg.Address)
		grpc.AddReplica(cfg.Address)
	}
	http_port := os.Getenv("HTTP_PORT")
	grpc_port := os.Getenv("GRPC_PORT")

	_, is_master = os.LookupEnv("IS_MASTER")
	if is_master {
		log.Println("Running in master mode...")
	}
	go grpc.Server_main(db, grpc_port, is_master)
	api.RunServer(db, http_port, is_master)
}
