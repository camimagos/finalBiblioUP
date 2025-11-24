package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "cubiculosup.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/resolver"
)

func init() {
	// Fuerza gRPC a usar DNS como scheme por defecto (necesario para Kubernetes)
	resolver.SetDefaultScheme("dns")
}

func newDialer() func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	}
}

type cubicleServer struct {
	pb.UnimplementedCubicleServiceServer
	metaClient pb.MetadataServiceClient
	resClient  pb.ReservationServiceClient
}

// NewCubicleServer inicializa los clientes gRPC para los servicios de Metadata y Reservation.
func NewCubicleServer() *cubicleServer {
	// Definimos un contexto con timeout para la conexión inicial
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Usamos el prefijo "dns:///" para forzar al cliente gRPC a usar el resolvedor DNS de Go.
	// Esto permite el balanceo de carga entre los pods del servicio interno.
	metaAddr := "dns:///metadata:50051"
	resAddr := "dns:///reservation:50052"

	serviceConfig := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)

	// Creamos el dialer personalizado que fuerza IPv4
	dialer := newDialer()

	// --- Conexión a Metadata ---
	metaConn, err := grpc.DialContext(
		ctx,
		metaAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithContextDialer(dialer),
	)

	if err != nil {
		log.Printf("WARNING: Could not connect immediately to metadata service: %v", err)
	}

	// --- Conexión a Reservation (Puerto 50052) ---
	resConn, err := grpc.DialContext(
		ctx,
		resAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		log.Printf("WARNING: Could not connect immediately to reservation service: %v", err)
	}

	return &cubicleServer{
		metaClient: pb.NewMetadataServiceClient(metaConn),
		resClient:  pb.NewReservationServiceClient(resConn),
	}
}

// GetCubicle implementa el método del servicio CubicleService.
func (s *cubicleServer) GetCubicle(ctx context.Context, req *pb.GetCubicleRequest) (*pb.GetCubicleResponse, error) {
	id := req.CubicleId

	// Llama al servicio Metadata (conexión interna)
	meta, err := s.metaClient.GetMetadata(ctx, &pb.GetMetadataRequest{CubicleId: id})
	if err != nil {
		return nil, fmt.Errorf("error calling metadata service: %v", err)
	}

	// Llama al servicio Reservation (conexión interna)
	avail, err := s.resClient.CheckAvailability(ctx, &pb.CheckAvailabilityRequest{CubicleId: id})
	if err != nil {
		return nil, fmt.Errorf("error calling reservation service: %v", err)
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
	// Escucha en todas las interfaces en el puerto 50053
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCubicleServiceServer(grpcServer, NewCubicleServer())

	reflection.Register(grpcServer)

	log.Println("Cubicle service running on port 50053")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
