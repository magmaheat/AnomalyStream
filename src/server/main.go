package main

import (
	"Go_Team00/src/config/proto"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type server struct {
	frequency.UnimplementedFrequencyServiceServer
}

func (s *server) GetFrequencies(req *frequency.FrequencyRequest, stream frequency.FrequencyService_GetFrequenciesServer) error {
	sessionID := uuid.New().String()
	mean := rand.Float64()*20 - 10     // Random mean in range [-10, 10]
	stddev := rand.Float64()*1.2 + 0.3 // Random stddev in range [0.3, 1.5]

	log.Printf("New session: %s, mean: %f, stddev: %f", sessionID, mean, stddev)

	for {
		frequencyValue := rand.NormFloat64()*stddev + mean
		timestamp := time.Now().UTC().Unix()

		res := &frequency.FrequencyResponse{
			SessionId: sessionID,
			Frequency: frequencyValue,
			Timestamp: timestamp,
		}

		if err := stream.Send(res); err != nil {
			return err
		}

		time.Sleep(time.Second) // Send data every second
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	frequency.RegisterFrequencyServiceServer(grpcServer, &server{})

	log.Printf("Server listening on port 50051")
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
