package eventclient

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/google/uuid"

	nebularcoretypes "gitlab.com/jideobs/nebularcore/tools/types"
)

type Client interface {
	Send(events ...Event) error
}

type Event struct {
	Id         uuid.UUID
	DetailType string
	Source     string
	Time       nebularcoretypes.DateTime
	Detail     any
}

type AwsEventBridgeClient struct {
	eventBus string
	client   *eventbridge.Client
}

func New(accessKey, secretKey, region, eventBus string) (*AwsEventBridgeClient, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return nil, err
	}
	client := eventbridge.NewFromConfig(awsConfig)
	return &AwsEventBridgeClient{
		client:   client,
		eventBus: eventBus,
	}, nil
}

func (a *AwsEventBridgeClient) Send(events ...Event) error {
	putEventsRequestEntries := []types.PutEventsRequestEntry{}

	for _, event := range events {
		eventDetail, err := json.Marshal(event.Detail)
		if err != nil {
			return err
		}

		putEventRequestEntity := types.PutEventsRequestEntry{
			Detail:       aws.String(string(eventDetail)),
			DetailType:   &event.DetailType,
			EventBusName: &a.eventBus,
			Source:       &event.Source,
			Time:         aws.Time(event.Time.Time()),
		}
		putEventsRequestEntries = append(putEventsRequestEntries, putEventRequestEntity)
	}
	putEventInputs := &eventbridge.PutEventsInput{
		Entries: putEventsRequestEntries,
	}
	_, err := a.client.PutEvents(context.TODO(), putEventInputs)

	return err
}
