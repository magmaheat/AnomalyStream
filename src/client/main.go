package main

import (
	"Go_Team00/src/config/proto"
	"context"
	"flag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FrequencyRecord struct {
	SessionID string  `gorm:"column:session_id"`
	Frequency float64 `gorm:"column:frequency"`
	Timestamp int64   `gorm:"column:timestamp"`
}

type Values struct {
	count    int
	sum      float64
	sumSq    float64
	stdValue float64
}

const serverAddress = "localhost:50051"

func main() {
	stdValue := flag.Float64("k", 2.0, "Anomaly detection coefficient")
	flag.Parse()

	dsn := "host=localhost user=user password=password dbname=db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	log.Println("connection to bd completed")

	// Автоматически создаем таблицу на основе структуры
	if err = db.AutoMigrate(&FrequencyRecord{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Подключение к gRPC серверу
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()
	log.Println("connection to server grpc completed")

	client := frequency.NewFrequencyServiceClient(conn)

	// Запрос частот
	stream, err := client.GetFrequencies(context.Background(), &frequency.FrequencyRequest{})
	if err != nil {
		log.Fatalf("Failed to get frequencies: %v", err)
		return
	}

	log.Println("start value processing...")
	// Основной цикл обработки входящих значений
	values := &Values{}
	values.stdValue = *stdValue

	for {
		err = valueProcessing(stream, values, db)
		if err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func valueProcessing(stream frequency.FrequencyService_GetFrequenciesClient, values *Values, db *gorm.DB) error {
	res, err := stream.Recv()
	if err != nil {
		log.Fatalf("Failed to receive data: %v", err)
		return err
	}

	values.count++
	value := res.GetFrequency()
	values.sum += value
	values.sumSq += value * value

	// Вычисление текущего среднего и стандартного отклонения
	mean := values.sum / float64(values.count)
	variance := (values.sumSq / float64(values.count)) - (mean * mean)
	stddev := math.Sqrt(variance)

	// Логирование промежуточных результатов
	if values.count%10 == 0 {
		log.Printf("Processed %d values, mean: %f, stddev: %f", values.count, mean, stddev)
	}

	// Обнаружение аномалий
	if math.Abs(value-mean) > values.stdValue*stddev {
		log.Printf("Anomaly detected: value %f is out of bounds (mean: %f, stddev: %f, k: %f)", value, mean, stddev, values.stdValue)

		record := FrequencyRecord{
			SessionID: res.GetSessionId(),
			Frequency: res.GetFrequency(),
			Timestamp: res.GetTimestamp(),
		}

		if err = db.Create(&record).Error; err != nil {
			log.Fatalf("Failed to create record: %v", err)
			return err
		}
		log.Println("Record inserted successfully")
	}

	return nil
}
