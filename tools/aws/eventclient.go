package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"gitlab.com/jideobs/nebularcore/tools/eventclient"
)

type AwsEventBridgeClient struct {
	eventBus string
	client   *eventbridge.Client
}

func NewEventClient(accessKey, secretKey, region, eventBus string) (eventclient.Client, error) {
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

func (a *AwsEventBridgeClient) Send(events ...eventclient.Event) error {
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
