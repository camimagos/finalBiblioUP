package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"

	pb "cubiculosup.com/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type metadataServer struct {
	pb.UnimplementedMetadataServiceServer
	db *sql.DB
}

func (s *metadataServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	row := s.db.QueryRow(`
		SELECT id, name, location, capacity
		FROM metadata
		WHERE id = $1
	`, req.CubicleId)

	var m pb.Metadata
	err := row.Scan(&m.Id, &m.Name, &m.Location, &m.Capacity)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return &pb.GetMetadataResponse{Metadata: &m}, nil
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot connect db: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("ping error: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMetadataServiceServer(s, &metadataServer{db: db})
	reflection.Register(s)

	log.Println("Metadata service running with PostgreSQL on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
