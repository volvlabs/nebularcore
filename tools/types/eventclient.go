package types

type EventClient string

const (
	GcloudPubSubClient   EventClient = "GcloudPubSub"
	AWSEventBridgeClient EventClient = "AWSEventBridgeClient"
	AWSSQSClient         EventClient = "AWSSQSClient"
)
