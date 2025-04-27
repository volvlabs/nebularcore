package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"gitlab.com/jideobs/nebularcore/tools/eventclient"
)

type SqsClient struct {
	queueUrl string
	client   *sqs.Client
}

// NewSqsClient creates a new SQS client instance
func NewSqsClient(accessKey, secretKey, region, queueUrl string) (eventclient.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(awsConfig)
	return &SqsClient{
		client:   client,
		queueUrl: queueUrl,
	}, nil
}

// SendMessages sends multiple messages to the SQS queue in batch
func (s *SqsClient) Send(events ...eventclient.Event) error {
	var entries []types.SendMessageBatchRequestEntry

	for i, event := range events {
		messageBody, err := json.Marshal(event)
		if err != nil {
			return err
		}

		entry := types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("msg-%d", i)), // Unique ID for each message in the batch
			MessageBody: aws.String(string(messageBody)),
		}
		entries = append(entries, entry)
	}

	input := &sqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: aws.String(s.queueUrl),
	}

	_, err := s.client.SendMessageBatch(context.TODO(), input)
	return err
}
