package main

import (
	"context"
	"database/sql"
	"fmt"
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

func (s *metadataServer) CreateMetadata(ctx context.Context, req *pb.CreateMetadataRequest) (*pb.CreateMetadataResponse, error) {
	meta := req.Metadata
	_, err := s.db.Exec(`
		INSERT INTO metadata (id, name, location, capacity)
		VALUES ($1, $2, $3, $4)
	`, meta.Id, meta.Name, meta.Location, meta.Capacity)
	if err != nil {
		return nil, err
	}
	return &pb.CreateMetadataResponse{CubicleId: meta.Id}, nil
}

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// 2. Usamos la IP de ClusterIP del servicio para forzar IPv4 (la que obtuviste de 'kubectl get svc')
	// Esto evita el fallo de DNS/IPv6 ([::1]).
	const dbHostIP = "10.96.128.214"
	const dbPort = "5432"

	// 3. Construye la URL de conexión de PostgreSQL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPassword,
		dbHostIP,
		dbPort,
		dbName,
	)

	// Si tus secretos son correctos, dbURL se verá como:
	// postgres://postgres:password@10.96.128.214:5432/cubiculos?sslmode=disable

	// 4. Abre la conexión a la base de datos
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		// Este es el error que causaba el CrashLoopBackOff: ¡resuelto con la IP fija!
		log.Fatalf("ping error: %v", err)
	}
	// --- FIN DEL CÓDIGO DE CONEXIÓN A DB ---

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
