package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending  = "pending"
	StatusFailed   = "failed"
	StatusStarting = "starting"
)

// TODO: IMPORTANTE l'user id quando si crea l'immagine non può essere la mail, meglio usare l'id dell'utente o il suo username
// TODO: non accettare richieste di build image di un applicazione se l'applicazione è già in status pending|updating
// TODO: setta status dell'applicazione prima di inviare la richiesta di build image (pending|updating)
func (c *Controller) BuildImage(ctx context.Context, app *model.Application, providerToken string) error {
	request := model.BuildRequest{
		UUID:      app.ID.String(),
		Token:     providerToken,
		UserID:    app.OwnerUsername,
		Type:      "repo",
		Connector: "github",
		Repo:      app.GithubRepo,
		Branch:    app.GithubBranch,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if err := c.sendToRabbitmq(body); err != nil {
		c.l.Errorf("error sending to rabbitmq: %v", err)
		app.Status = StatusFailed
		if _, err := c.applicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application status: %v", err)
			return fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
		}

		return fmt.Errorf("c.sendToRabbitmq: %w", err)
	}

	return nil
}

func (c *Controller) sendToRabbitmq(body []byte) error {
	connection, err := amqp.Dial(c.rabbitUri)
	if err != nil {
		c.l.Errorf("error dialing rabbitmq: %v", err)
		return fmt.Errorf("ampq.Dial: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		c.l.Errorf("error creating channel: %v", err)
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	if err := channel.Publish(
		"",                 // exchange
		c.requestQueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
		c.l.Error("c.SendResponse.Channel.Publish(): %w:", err)
		return fmt.Errorf("c.Channel.Publish: %w", err)
	}

	c.l.Info("response sent to rabbitmq")
	return nil
}

func (c *Controller) ValidateImageResponse(ctx context.Context, reponse model.BuildResponse) (string, error) {
	if reponse.Status == "failed" {
		c.l.Errorf("error building image: %v", reponse.Error)
		uuid, err := primitive.ObjectIDFromHex(reponse.UUID)
		if err != nil {
			c.l.Errorf("error parsing uuid: %v", err)
			return "", fmt.Errorf("primitive.ObjectIDFromHex: %w", err)
		}
		app, err := c.applicationRepo.FindByID(ctx, uuid)
		if err != nil {
			c.l.Errorf("error finding application: %v", err)
			return "", fmt.Errorf("c.applicationRepo.FindByID: %w", err)
		}

		app.Status = StatusFailed
		if _, err := c.applicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application status: %v", err)
			return "", fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
		}
		return "", fmt.Errorf("error building image: %v", reponse.Error)
	}

	return reponse.ImageID, nil

}
