// Package queue provides functions for interacting with RabbitMQ
package queue

import (
	"fmt"
	log "worker-service/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	URL        string
	channel    *amqp.Channel
	connection *amqp.Connection
}

func (rmq *RabbitMQClient) Connect() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(rmq.URL)
	log.FailOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	log.FailOnError(err, "Failed to connect to RabbitMQ")

	rmq.channel = ch
	rmq.connection = conn

	fmt.Println("✅ Connected to RabbitMQ")

	return conn, ch
}

func (rmq *RabbitMQClient) DeclareQueue(queueName string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (*amqp.Queue, error) {
	queue, err := rmq.channel.QueueDeclare(
		queueName,  // name of the queue
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		noWait,     // no-wait
		args,       // arguments
	)
	log.FailOnError(err, "Failed to declare a queue")

	return &queue, err
}

func (rmq *RabbitMQClient) CloseConnection() {
	rmq.channel.Close()
	rmq.connection.Close()
}
