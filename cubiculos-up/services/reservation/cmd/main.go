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
	cubicleID := req.CubicleId
	log.Printf("Checking availability for cubicle ID: %s", cubicleID)

	// 1. Definir la hora de referencia (ahora)
	// Es crucial usar UTC para coincidir con el formato de tu base de datos y evitar problemas de zona horaria.
	now := time.Now().In(time.UTC)

	// 2. Buscar la reserva activa AHORA, si existe

	// Consulta SQL: Busca una reserva CONFIRMED que haya comenzado (start_time <= now)
	// y que aún no haya terminado (end_time > now).
	activeRow := s.db.QueryRowContext(ctx, `
        SELECT end_time
        FROM reservations
        WHERE record_type = $1 AND status = 'CONFIRMED' 
          AND end_time > $2
        ORDER BY start_time ASC
        LIMIT 1
    `, cubicleID, now)

	var currentEndTime time.Time
	err := activeRow.Scan(&currentEndTime)

	if err != nil && err != sql.ErrNoRows {
		// Error de SQL (no es solo que no haya filas)
		log.Printf("SQL Error checking active reservation: %v", err)
		return nil, err
	}

	// 3. Determinar AvailableNow
	availableNow := (err == sql.ErrNoRows) // Es true si no se encontró una fila (no hay reservas activas)

	var nextAvailableTime time.Time

	if !availableNow {
		// 4. Si hay una reserva activa, la próxima disponibilidad es cuando termine la actual.
		// Ya tenemos currentEndTime de la consulta anterior.
		nextAvailableTime = currentEndTime

	} else {
		// 5. Si está disponible ahora, busca la PRÓXIMA reserva

		// Consulta SQL: Busca la reserva CONFIRMED más próxima en el futuro (start_time > now).
		nextRow := s.db.QueryRowContext(ctx, `
            SELECT start_time
            FROM reservations
            WHERE record_type = $1 AND status = 'CONFIRMED' AND start_time > $2
            ORDER BY start_time ASC
            LIMIT 1
        `, cubicleID, now)

		err = nextRow.Scan(&nextAvailableTime)

		if err != nil && err != sql.ErrNoRows {
			// Error de SQL
			log.Printf("SQL Error checking next reservation: %v", err)
			return nil, err
		}

		if err == sql.ErrNoRows {
			// No hay reservas futuras, la disponibilidad es indefinida (usamos una marca de tiempo lejana o la hora actual)
			// Usamos la hora actual, ya que técnicamente está disponible AHORA.
			nextAvailableTime = now
		}
	}

	// 6. Construir la respuesta
	return &pb.CheckAvailabilityResponse{
		Availability: &pb.Availability{
			AvailableNow:  availableNow,
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
