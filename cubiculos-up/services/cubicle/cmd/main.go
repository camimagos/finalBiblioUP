package main

import (
	"context"
	"log"
	"net"

	pb "cubiculosup.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type cubicleServer struct {
	pb.UnimplementedCubicleServiceServer
	metaClient pb.MetadataServiceClient
	resClient  pb.ReservationServiceClient
}

// func NewCubicleServer() *cubicleServer {
// 	// IMPORTANTE:
// 	// Estos hostnames funcionan dentro del cluster:
// 	//    metadata.default.svc.cluster.local:50051
// 	//    reservation.default.svc.cluster.local:50052

// 	metaConn, err := grpc.Dial("metadata.default.svc.cluster.local:50051",
// 		grpc.WithInsecure(),
// 		grpc.WithBlock(),
// 	)
// 	if err != nil {
// 		log.Fatalf("cannot connect metadata service: %v", err)
// 	}

// 	resConn, err := grpc.Dial("reservation.default.svc.cluster.local:50052",
// 		grpc.WithInsecure(),
// 		grpc.WithBlock(),
// 	)
// 	if err != nil {
// 		log.Fatalf("cannot connect reservation service: %v", err)
// 	}

// 	return &cubicleServer{
// 		metaClient: pb.NewMetadataServiceClient(metaConn),
// 		resClient:  pb.NewReservationServiceClient(resConn),
// 	}
// }

// prueba local
func NewCubicleServer() *cubicleServer {
	metaConn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cannot connect to metadata: %v", err)
	}

	resConn, err := grpc.Dial("localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cannot connect to reservation: %v", err)
	}

	return &cubicleServer{
		metaClient: pb.NewMetadataServiceClient(metaConn),
		resClient:  pb.NewReservationServiceClient(resConn),
	}
}

func (s *cubicleServer) GetCubicle(ctx context.Context, req *pb.GetCubicleRequest) (*pb.GetCubicleResponse, error) {
	id := req.CubicleId

	meta, err := s.metaClient.GetMetadata(ctx, &pb.GetMetadataRequest{CubicleId: id})
	if err != nil {
		return nil, err
	}

	avail, err := s.resClient.CheckAvailability(ctx, &pb.CheckAvailabilityRequest{CubicleId: id})
	if err != nil {
		return nil, err
	}

	details := &pb.CubicleDetails{
		Metadata:    meta.Metadata,
		Reservation: avail.Availability,
	}

	return &pb.GetCubicleResponse{
		Details: details,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053") // PUERTO EXTERNO
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCubicleServiceServer(grpcServer, NewCubicleServer())

	//local
	reflection.Register(grpcServer)

	log.Println("Cubicle service running on port 50053")
	log.Println("Ready to be exposed as LoadBalancer in Kubernetes")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
