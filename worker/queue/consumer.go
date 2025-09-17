// Package queue provides functions for interacting with RabbitMQ
package queue

import (
	log "worker-service/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (rmq *RabbitMQClient) ConsumeMessages(queueName string) <-chan amqp.Delivery {
	msgs, err := rmq.channel.Consume(queueName, "", false, false, false, false, nil)
	log.FailOnError(err, "Failed to consume messages")

	return msgs
}
