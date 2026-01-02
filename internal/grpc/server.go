package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/anormalprgrmr/DKV_DB/proto"
	dal "github.com/anormalprgrmr/DKV_DB/internal/DAL"
)

type kvServer struct {
	pb.UnimplementedDKVDBServiceServer
}

var db *dal.DB
var tx_map map[uint64]*dal.Tx
func (s *kvServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.Status, error) {
	log.Printf("Set: key=%s value=%s", req.Key, req.Value)
	tx := db.WriteTx()
	tx_map[req.Id] = tx
	col, err := tx.GetCollection([]byte(dal.DEFAULT_COLLECTION))
	if err != nil {
		log.Println("default collection does not exist")
		return &pb.Status{
			Success: false,
			Message: "default collection does not exist",
		}, nil
	}
	//1. set
	err = col.Put([]byte(req.Key), []byte(req.Value))
	if err != nil {
		log.Println("can't put key/value")
		return &pb.Status{
			Success: false,
			Message: "can't put key/value",
		}, nil
	}
	return &pb.Status{
		Success: true,
		Message: "sucess",
	}, nil
}

func (s *kvServer) Commit(ctx context.Context, req *pb.TxRequest) (*pb.Status, error) {
	log.Println("Commit")
	tx := tx_map[req.Id]
	tx.Commit()
	return &pb.Status{
		Success: true,
		Message: "commit successful",
	}, nil
}

func (s *kvServer) Abort(ctx context.Context, req *pb.TxRequest) (*pb.Status, error) {
	log.Println("Abort")
	tx := tx_map[req.Id]
	tx.Rollback()
	return &pb.Status{
		Success: true,
		Message: "abort successful",
	}, nil
}

func Server_main(_db *dal.DB, port string) {
	db = _db
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDKVDBServiceServer(grpcServer, &kvServer{})

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

