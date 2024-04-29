package ipaas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/imageBuilder"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ imageBuilder.ImageBuilder = new(IpaasImageBuilder)

const (
	ResponseStatusSuccess = "success"
	ResponseStatusFailed  = "failed"

	ResponseErrorFaultService = "service"
	ResponseErrorFaultUser    = "user"
)

type IpaasImageBuilder struct {
	uri              string
	requestQueueName string
}

func NewIpaasImageBuilder(uri, requestQueue string) *IpaasImageBuilder {
	return &IpaasImageBuilder{
		uri:              uri,
		requestQueueName: requestQueue,
	}
}

func (i *IpaasImageBuilder) BuildImage(ctx context.Context, info model.Request) error {
	body, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	return i.sendToRabbitmq(ctx, body)
}

func (i *IpaasImageBuilder) sendToRabbitmq(ctx context.Context, body []byte) error {
	connection, err := amqp.Dial(i.uri)
	if err != nil {
		return fmt.Errorf("ampq.Dial: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}
	defer channel.Close()

	return channel.PublishWithContext(
		ctx,
		"",                 // exchange
		i.requestQueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (i *IpaasImageBuilder) ValidateImageResponse(response model.BuildResponse) (string, error) {
	if response.Status != ResponseStatusSuccess {
		return "", fmt.Errorf("response status: %s", response.Status)
	}

	return response.ImageID, nil
}
