package test

import "github.com/volvlabs/nebularcore/tools/eventclient"

type eventClientMock struct{}

func (e *eventClientMock) Send(event ...eventclient.Event) error {
	return nil
}
