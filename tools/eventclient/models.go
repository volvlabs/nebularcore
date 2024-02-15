package eventclient

import (
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

type Event struct {
	Id         uuid.UUID      `json:"id"`
	DetailType string         `json:"detailType"`
	Source     string         `json:"source"`
	Time       types.DateTime `json:"time"`
	Detail     any            `json:"detail"`
}
