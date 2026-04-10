package gcloud

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/models"
	"github.com/volvlabs/nebularcore/tools/eventclient"
	"google.golang.org/api/option"
)

type Client struct {
	ctx context.Context
	cfg models.GcloudConfig
}

func NewEventClient(cfg models.GcloudConfig) (eventclient.Client, error) {
	ctx := context.Background()

	return &Client{
		ctx: ctx,
		cfg: cfg,
	}, nil
}

func (c *Client) Send(events ...eventclient.Event) error {
	client, err := pubsub.NewClient(
		c.ctx, c.cfg.ProjectId, option.WithCredentialsFile(c.cfg.CredfileLocation))
	if err != nil {
		return err
	}
	defer client.Close()

	topic := client.Topic(c.cfg.PubSub.Topic)
	defer topic.Stop()

	var results []*pubsub.PublishResult
	for _, event := range events {
		eventDetail, err := json.Marshal(event)
		if err != nil {
			return err
		}
		result := topic.Publish(c.ctx, &pubsub.Message{
			Data: eventDetail,
		})
		results = append(results, result)
	}

	for _, result := range results {
		id, err := result.Get(c.ctx)
		if err != nil {
			log.Err(err).Msgf("Gcloud PubSub: publishing of event")
			return err
		}

		log.Info().Msgf("Gcloud PubSub: published a event with ID: %v", id)
	}

	return nil
}
