package grpc

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/anormalprgrmr/DKV_DB/proto"
	"google.golang.org/grpc"
)
type replica struct{
	client pb.DKVDBServiceClient
}

var replicas []replica


func AddReplica(address string) error {
	//var r replica
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return errors.New("Couldn't connect to replica")
	}

	client := pb.NewDKVDBServiceClient(conn)
	replicas = append(replicas, replica{client})
	return nil

}

var id uint64 = 0
func ReplicaSet(key string, value string) error {
	id += 1
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	setRequest := pb.SetRequest{Key: key,Value: value, Id:id }
	txRequest := pb.TxRequest{Id: id}
	var ar []replica 
	var abort = false
	for i , r := range replicas {
		resp, _ := r.client.Set(ctx, &setRequest)
		if !resp.Success {
			log.Printf("set in replica %v failed\n", i)
			abort = true
			break;
		}else{
			ar = append(ar, r)
		}
	}
	if abort {
		for i, r := range ar {
			resp, _ := r.client.Abort(ctx, &txRequest)
			if !resp.Success {
				log.Printf("abort in replica %v failed\n", i)
			}
		}
		return errors.New("Set in Replica Failed")
	}
	for i , r := range replicas {
		resp, _ := r.client.Commit(ctx, &txRequest)
		if !resp.Success {
			log.Printf("commit in replica %v failed\n", i)
			abort = true
			break;
		}else{
			ar = append(ar, r)
		}
	}
	return nil

}
