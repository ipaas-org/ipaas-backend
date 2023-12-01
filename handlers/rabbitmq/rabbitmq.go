package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Connection        *amqp.Connection
	Channel           *amqp.Channel
	ResponseQueue     amqp.Queue
	Delivery          <-chan amqp.Delivery
	responseQueueName string
	uri               string
	// requestQueueName  string

	l          *logrus.Logger
	Controller *controller.Controller

	Done     chan struct{}
	Error    <-chan error
	restarts int
}

func NewRabbitMQ(uri, requestQueue, responseQueue string, controller *controller.Controller, logger *logrus.Logger) *RabbitMQ {
	logger.Infof("listening on %s for responses", responseQueue)
	return &RabbitMQ{
		uri:               uri,
		l:                 logger,
		responseQueueName: responseQueue,
		Controller:        controller,
		Done:              make(chan struct{}),
		restarts:          0,
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

	responseQueue, err := r.Channel.QueueDeclare(
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
		responseQueue.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
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

func (r *RabbitMQ) Start(ctx context.Context, ID int, routineMonitor chan int) {
	defer func(restarts int) {
		r.l.Info("rabbitmq connection closed")
		rec := recover()
		r.l.Debug("recover:", rec)
		if rec != nil {
			r.l.Error("rabbitmq routine panic:", rec)
			r.l.Error(string(debug.Stack()))
		}
		if err := r.Close(); err != nil {
			r.l.Errorf("error closing connection with rmq: %v:", err)
		}
		if ctx.Err() == nil {
			//  && restarts <= 5 {
			// 	if restarts > 0 {
			// 		time.Sleep(3 * time.Second)
			// 	}
			routineMonitor <- ID
		} else {
			r.l.Infof("rabbitmq routine [ID=%d] not restarting", ID)
			r.Done <- struct{}{}
		}
	}(r.restarts)

	r.l.Infof("starting rabbitmq routine [ID=%d]", ID)
	if err := r.Connect(); err != nil {
		r.l.Error("r.Connect():", err)
		r.restarts++
		return
	}

	r.l.Infof("rabbitmq routine [ID=%d] connected", ID)

	r.consume(ctx)
	r.l.Info("rabbitmq done consuming")
}

func (r *RabbitMQ) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			r.l.Info("stopping rabbitmq consumer")
			return
		case d := <-r.Delivery:
			r.l.Info("received message from rabbitmq")
			r.l.Debugf("received: %q", string(d.Body))
			if d.Body == nil {
				// if err := d.Ack(false); err != nil {
				// 	r.l.Errorf("r.Consume.Ack(): %v:", err)
				// 	return
				// }
				continue
			}
			response := new(model.BuildResponse)
			if err := json.Unmarshal(d.Body, response); err != nil {
				r.l.Errorf("r.Consume.json.Unmarshal(): %v:", err)
				r.l.Debug(string(d.Body))
				// if err := d.Nack(false, false); err != nil {
				// 	r.l.Errorf("r.Consume.Nack(): %v:", err)
				// 	return
				// }
				continue
				//TODO: should send to a unprocessable queue
			}

			r.l.Debug(response)
			if response.IsError {
				r.l.Info("r.Controller: error building image:", response.Message)
				r.l.Info("r.Controller: error building image fault:", response.Fault)
				// if response.Fault == model.ResponseErrorFaultService {
				// 	//TODO: resend the message to the queue to process again, at least 3 times
				// 	//if it fails notify the user that the build failed and to try again later
				// } else {
				// 	//TODO: notify the user that the build failed and the reason
				// 	//build error is in response.Message
				// }
				if response.Fault == model.ResponseErrorFaultUser {
					r.Controller.FailedBuild(ctx, response)
				}
				continue
				// if err := d.Nack(false, false); err != nil {
				// 	r.l.Errorf("r.Consume.Nack(): %v:", err)
				// 	return
				// }
			}

			if err := r.Controller.CreateApplicationFromApplicationIDandImageID(ctx, response.ApplicationID, response.ImageID, response.BuiltCommit); err != nil {
				r.l.Errorf("error creating application after image builder response %v:", err)
				// if err := d.Nack(false, false); err != nil {
				// 	r.l.Errorf("r.Consume.Nack(): %v:", err)
				// 	return
				// }
				continue
			}

			r.l.Info("application created successfully")
		}
	}
}

/*
{
	"applicationID":"654e3ad0368ca328670b7d55",
	"repo":"https://github.com/Vano2903/testing",
	"status":"success",
	"imageID":"b5c16592-27c6-4df9-89ca-7260c8a37e09",
	"imageName":"",
	"buildCommit":"8c76e330c5e4119f814f4e66dce4514082157503",
	"isError":false,
	"fault":"",
	"message":""
}
*/
