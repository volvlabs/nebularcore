package tools

import (
	"context"

	"gitlab.com/jideobs/nebularcore/apis"
	"gorm.io/gorm"
)

func GetDBSessionFromContext(ctx context.Context) *gorm.DB {
	dbSession := ctx.Value(apis.ContextDBSessionKey)
	if dbSession == nil {
		return nil
	}

	return dbSession.(*gorm.DB)
}
