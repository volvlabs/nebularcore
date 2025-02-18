package entities

import (
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gorm.io/gorm"
)

type Auth struct {
	BaseModel

	Identity                     string `gorm:"uniqueIndex"`
	UserTableName                string
	UserId                       uuid.UUID  `gorm:"uniqueIndex"`
	Role                         types.Role `json:"role"`
	PasswordHash                 string
	ResetPasswordToken           string         `json:"resetPasswordToken"`
	ResetPasswordTokenExpiryDate types.DateTime `json:"resetPasswordTokenExpiryDate"`
	OtpSecret                    string         `json:"otpSecret"`
}

func (a *Auth) AfterCreate(tx *gorm.DB) error {
	userOtp, err := security.GenerateUniqueOtpSecret(a.Id)

	if err != nil {
		return err
	}

	return tx.Model(&Auth{}).
		Where("id = ?", a.Id).
		Updates(&Auth{
			OtpSecret: userOtp,
		}).Error
}
