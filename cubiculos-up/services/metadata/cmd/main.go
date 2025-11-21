package main

import (
	"context"
	"log"
	"net"

	pb "cubiculosup.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type metadataServer struct {
	pb.UnimplementedMetadataServiceServer
}

func (s *metadataServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	// Aquí normalmente harías consulta a PostgreSQL,
	// pero para MVP respondemos datos quemados.
	m := &pb.Metadata{
		Id:       req.CubicleId,
		Name:     "Cubículo " + req.CubicleId,
		Location: "Biblioteca planta baja",
		Capacity: 4,
	}

	return &pb.GetMetadataResponse{
		Metadata: m,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMetadataServiceServer(s, &metadataServer{})

	//local
	reflection.Register(s)

	log.Println("Metadata service running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
