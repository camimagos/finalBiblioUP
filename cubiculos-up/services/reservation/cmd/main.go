package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "cubiculosup.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type reservationServer struct {
	pb.UnimplementedReservationServiceServer
}

func (s *reservationServer) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {

	// Simulación sencilla: siempre disponible excepto cubículo "C2".
	if req.CubicleId == "C2" {
		return &pb.CheckAvailabilityResponse{
			Availability: &pb.Availability{
				AvailableNow:  false,
				NextAvailable: timestamppb.New(time.Now().Add(30 * time.Minute)),
			},
		}, nil
	}

	return &pb.CheckAvailabilityResponse{
		Availability: &pb.Availability{
			AvailableNow:  true,
			NextAvailable: timestamppb.New(time.Now()),
		},
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterReservationServiceServer(s, &reservationServer{})
	//local
	reflection.Register(s)

	log.Println("Reservation service running on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
