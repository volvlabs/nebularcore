package daos

import (
	"gitlab.com/volvlabs/nebularcore/models"

	"gorm.io/gorm"
)

type Dao struct {
	dbConn *gorm.DB
}

func New(dbConn *gorm.DB) *Dao {
	return &Dao{dbConn}
}

func (d *Dao) DB() *gorm.DB {
	return d.dbConn
}

func (d *Dao) Save(model models.Model) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) (err error) {
		if !model.HasId() {
			err = tx.Create(model).Error
		} else {
			err = tx.Save(model).Error
		}
		return
	})
}

func (d *Dao) FindBy(model models.Model, where any) error {
	return d.dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Where(where).First(model).Error
	})
}
