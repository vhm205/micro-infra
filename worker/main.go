package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"worker-service/db"
	"worker-service/jobs"
	minio "worker-service/pkg"
	"worker-service/queue"
	log "worker-service/utils"
	"worker-service/worker"
	// "github.com/joho/godotenv"
)

var rabbitClient *queue.RabbitMQClient

func main() {
	// err := godotenv.Load()
	// log.FailOnError(err, "Failed to load .env file")
	fmt.Println(os.Getenv("RABBIT_MQ_URL"))

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	isStop := make(chan bool, 1)

	// Init Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init DB
	dbClient := &db.MySQLClient{}
	dbClient.Connect()
	defer dbClient.CloseConnection()

	// Init Minio
	minio := &minio.MinioClient{
		Endpoint:   os.Getenv("MINIO_ENDPOINT"),
		AccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		BucketName: os.Getenv("MINIO_BUCKET_NAME"),
	}
	minio.Connect()

	// Init Worker Pool
	pool := worker.NewWorkerPool(ctx, 3, 10)
	pool.Start()
	defer pool.Stop()

	// Init RabbitMQ
	rabbitClient = &queue.RabbitMQClient{
		URL: os.Getenv("RABBIT_MQ_URL"),
	}
	rabbitClient.Connect()
	defer rabbitClient.CloseConnection()

	msgs := rabbitClient.ConsumeMessages("product-service")

	go func() {
		var count int = 0
		for msg := range msgs {
			// fmt.Printf(" [%d] %s with id %s\n", count, msg.Body, msg.RoutingKey)

			payload := jobs.MessagePayload{}
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.FailOnError(err, "Failed to unmarshal payload")
				msg.Nack(false, false)
				continue
			}

			job := jobs.Job{
				ID:          count,
				Payload:     payload,
				DB:          dbClient.DB,
				MinioClient: minio,
			}
			pool.Submit(job)

			count++
			msg.Ack(true)
		}
	}()

	fmt.Println(" [*] Waiting for logs. To exit press CTRL+C")
	<-isStop
}
