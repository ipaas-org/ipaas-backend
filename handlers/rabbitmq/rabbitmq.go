package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	l             *logrus.Logger
	Error         <-chan error
	Connection    *amqp.Connection
	Channel       *amqp.Channel
	ResponseQueue amqp.Queue
	Delivery      <-chan amqp.Delivery
	uri           string
	// requestQueueName  string
	responseQueueName string
	Controller        *controller.Controller
}

func NewRabbitMQ(uri, requestQueue, responseQueue string, controller *controller.Controller, logger *logrus.Logger) *RabbitMQ {
	logger.Infof("listening on %s for responses", responseQueue)
	return &RabbitMQ{
		uri:               uri,
		l:                 logger,
		responseQueueName: responseQueue,
		Controller:        controller,
	}
}

func (r *RabbitMQ) Connect() error {
	r.l.Info("connecting to rabbitmq")
	r.l.Debug(r.uri)
	var err error
	r.Connection, err = amqp.Dial(r.uri)
	if err != nil {
		return fmt.Errorf("ampq.Dial: %w", err)
	}

	r.Channel, err = r.Connection.Channel()
	if err != nil {
		return fmt.Errorf("r.Connection.Channel: %w", err)
	}

	if err = r.Channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("r.Channel.Qos: %w", err)
	}

	q, err := r.Channel.QueueDeclare(
		r.responseQueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("r.Channel.QueueDeclare: %w", err)
	}

	r.Delivery, err = r.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("r.Channel.Consume: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return fmt.Errorf("r.Channel.Close: %w", err)
	}

	if err := r.Connection.Close(); err != nil {
		return fmt.Errorf("r.Connection.Close: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			r.l.Info("stopping rabbitmq consumer")
			return
		case err := <-r.Error:
			r.l.Error(err)
		case d := <-r.Delivery:
			r.l.Info("received message from rabbitmq")
			r.l.Debug(string(d.Body))

			response := new(model.BuildResponse)
			if err := json.Unmarshal(d.Body, response); err != nil {
				r.l.Error("r.Consume.json.Unmarshal(): %w:", err)
				r.l.Debug(string(d.Body))
				if err := d.Nack(false, false); err != nil {
					r.l.Error("r.Consume.Nack(): %w:", err)
				}
			}

			r.l.Debug(response)
			if response.Error != nil {
				r.l.Error("r.Controller: error building image", response.Error.Message)
				r.l.Debug(response.Error.Message)
				if err := d.Nack(false, false); err != nil {
					r.l.Error("r.Consume.Nack(): %w:", err)
				}
			}


			
			if err := r.Controller.CreateContainerFromIDAndImage(ctx, response.UUID, response.LatestCommit, imageID); err != nil {
				r.l.Errorf("r.Controller.CreateContainerFromIDAndImage(): %w:", err)
				if err := d.Nack(false, false); err != nil {
					r.l.Errorf("r.Consume.Nack(): %w:", err)
				}
			}

			if err := d.Ack(false); err != nil {
				r.l.Error("r.Consume.Ack(): %w:", err)
			}
			r.l.Info("acknowledged message from rabbitmq")
		}
	}
}
