package daos

import (
	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/models/config"
	"github.com/volvlabs/nebularcore/tools/types"

	"gorm.io/gorm"
)

type Dao struct {
	dbConn         *gorm.DB
	tenantConfig   *config.TenantConfig
	databaseConfig *config.DatabaseConfig
}

func New(
	dbConn *gorm.DB,
	tenantConfig *config.TenantConfig,
	databaseConfig *config.DatabaseConfig,
) *Dao {
	return &Dao{
		dbConn: dbConn, 
		tenantConfig: tenantConfig,
		databaseConfig: databaseConfig,
	}
}

func (d *Dao) DB() *gorm.DB {
	return d.dbConn
}

func (d *Dao) Save(model entities.Model) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) (err error) {
		if !model.HasId() {
			err = tx.Create(model).Error
		} else {
			err = tx.Save(model).Error
		}
		return
	})
}

func (d *Dao) FindBy(model entities.Model, where any) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Where(where).First(model).Error
	})
}

func (d *Dao) Update(model entities.Model, where any, column string, value any) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Model(model).Where(where).Update(column, value).Error
	})
}

func (d *Dao) Updates(model entities.Model, updates entities.Model) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Model(model).Updates(updates).Error
	})
}

func (d *Dao) Delete(model entities.Model) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Model(model).Updates(map[string]any{"is_deleted": true, "deleted_at": types.NowDateTime()}).Error
	})
}
