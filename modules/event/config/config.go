package config

import (
	"fmt"

	"github.com/volvlabs/nebularcore/modules/event/types"
	"github.com/volvlabs/nebularcore/tools/validation"
)

var ErrBackendConfiguredNotValid = fmt.Errorf(
	"backend configured not valid, opts are: %s, %s", types.BackendGoChannels, types.BackendKafka)

type Config struct {
	Backend   types.Backend `json:"backend" yaml:"backend"`
	Kafka     Kafka         `json:"kafka" yaml:"kafka" validate:"required_if=Backend kafka"`
	GoChannel GoChannel     `json:"goChannel" yaml:"goChannel" validate:"required_if=Backend goChannel"`
}

type Kafka struct {
	Brockers []string `json:"brockers" yaml:"brockers" validate:"omitempty,min=1,dive,required"`
}

type GoChannel struct {
	OutputChannelBuffer int64 `json:"outputChannelBuffer" yaml:"outputChannelBuffer" validate:"omitempty,required,min=1024"`
	Persistent          bool  `json:"persistent" yaml:"persistent"`
}

func Default() *Config {
	return &Config{
		Backend: types.BackendGoChannels,
		GoChannel: GoChannel{
			OutputChannelBuffer: 1024,
			Persistent:          false,
		},
	}
}

func (c *Config) Key() string {
	return "eventCfg"
}

func (c *Config) Validate() error {
	if !c.Backend.IsValid() {
		return fmt.Errorf(
			"invalid backend specified in config, got: %s, opts are: %s, %s",
			c.Backend, types.BackendGoChannels, types.BackendKafka,
		)
	}

	validator := validation.New()
	if err := validator.Validate(c); err != nil {
		return err
	}
	if c.Backend == types.BackendKafka {
		if err := validator.Validate(c.Kafka); err != nil {
			return err
		}
	}
	if c.Backend == types.BackendGoChannels {
		if err := validator.Validate(c.GoChannel); err != nil {
			return err
		}
	}
	return nil
}
