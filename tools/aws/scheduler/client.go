package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/aws/aws-sdk-go/aws"
)

type Client interface {
	Create(schedule *Schedule) error
	Delete(name string) error
}

type Schedule struct {
	Name          string
	ExecutionTime time.Time
	Metadata      map[string]any
}
type AwsScheduler struct {
	client *scheduler.Client
}

func New(accessKey, secretKey, region string) (*AwsScheduler, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return nil, err
	}
	client := scheduler.NewFromConfig(awsConfig)
	return &AwsScheduler{
		client: client,
	}, nil
}

func (a *AwsScheduler) Create(schedule *Schedule) error {
	executionTime := schedule.ExecutionTime
	cronExpression := fmt.Sprintf("cron(%d %d %d %d * ? *)", executionTime.Minute(), executionTime.Hour(), executionTime.Day(), executionTime.Month())
	target := schedule.Metadata["target"].(map[string]any)
	createScheduleInput := scheduler.CreateScheduleInput{
		Name: aws.String(schedule.Name),
		Target: &types.Target{
			Arn:     aws.String(target["Arn"].(string)),
			RoleArn: aws.String(target["RoleArn"].(string)),
			Input:   aws.String(target["Input"].(string)),
		},
		ScheduleExpression: &cronExpression,
	}
	_, err := a.client.CreateSchedule(context.TODO(), &createScheduleInput)
	if err != nil {
		return err
	}

	return nil
}

func (a *AwsScheduler) Delete(name string) error {
	deleteScheduleInput := scheduler.DeleteScheduleInput{
		Name: aws.String(name),
	}
	_, err := a.client.DeleteSchedule(context.TODO(), &deleteScheduleInput)
	return err
}
