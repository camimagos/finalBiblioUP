package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	pb "cubiculosup.com/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type reservationServer struct {
	pb.UnimplementedReservationServiceServer
	db *sql.DB
}

func (s *reservationServer) CreateReservation(ctx context.Context, req *pb.CreateReservationRequest) (*pb.CreateReservationResponse, error) {

	r := req.Reservation

	_, err := s.db.Exec(`
		INSERT INTO reservations (record_id, record_type, user_id, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		r.RecordId,
		r.RecordType,
		r.UserId,
		r.Start.AsTime(),
		r.End.AsTime(),
		r.Status,
	)

	if err != nil {
		return nil, err
	}

	return &pb.CreateReservationResponse{RecordId: r.RecordId}, nil
}

func (s *reservationServer) CancelReservation(ctx context.Context, req *pb.CancelReservationRequest) (*pb.CancelReservationResponse, error) {

	res, err := s.db.Exec(`DELETE FROM reservations WHERE record_id = $1`, req.RecordId)
	if err != nil {
		return nil, err
	}

	affected, _ := res.RowsAffected()
	return &pb.CancelReservationResponse{Ok: affected > 0}, nil
}

// En services/reservation/cmd/main.go, dentro de type reservationServer struct
func (s *reservationServer) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	// Por ahora, solo devolvemos un estado fijo para probar la conectividad
	// La lógica real de la DB iría aquí
	log.Printf("Checking availability for cubicle ID: %s", req.CubicleId)

	// Aquí iría el código SQL para consultar la base de datos
	// ...

	// Devolvemos una respuesta de ejemplo
	// Nota: Necesitas la importación "google.golang.org/protobuf/types/known/timestamppb"

	// Suponiendo que el cubículo está disponible ahora, y que la próxima reserva
	// comienza en el futuro.
	nextAvailableTime := time.Now().Add(1 * time.Hour)

	return &pb.CheckAvailabilityResponse{
		Availability: &pb.Availability{
			AvailableNow:  true,
			NextAvailable: timestamppb.New(nextAvailableTime),
		},
	}, nil
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot connect db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("ping error: %v", err)
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterReservationServiceServer(s, &reservationServer{db: db})
	reflection.Register(s)

	log.Println("Reservation service running with PostgreSQL on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
