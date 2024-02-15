package test

import "gitlab.com/jideobs/nebularcore/tools/eventclient"

type eventClientMock struct{}

func (e *eventClientMock) Send(event ...eventclient.Event) error {
	return nil
}
