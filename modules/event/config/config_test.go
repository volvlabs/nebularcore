package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/modules/event/config"
	"github.com/volvlabs/nebularcore/modules/event/types"
)

func TestValidateConfig(t *testing.T) {
	t.Run("should not return any error as configuration set properly", func(t *testing.T) {
		// Arrange:
		cfg := config.Config{
			Backend: types.BackendGoChannels,
			GoChannel: config.GoChannel{
				OutputChannelBuffer: 1024,
				Persistent:          false,
			},
		}

		// Act:
		err := cfg.Validate()

		// Assert:
		assert.Equal(t, nil, err)
	})

	t.Run("should return error because Kafka config was not set even with kafka specified as backend",
		func(t *testing.T) {
			// Arrange:
			cfg := config.Config{
				Backend: types.BackendKafka,
			}

			// Act:
			err := cfg.Validate()

			// Assert:
			assert.Error(t, err)
		})

	t.Run("should return error because brokers list is empty", func(t *testing.T) {
		// Arrange:
		cfg := config.Config{
			Backend: types.BackendKafka,
			Kafka: config.Kafka{
				Brockers: []string{},
			},
		}

		// Act:
		err := cfg.Validate()

		// Assert:
		assert.Error(t, err)
	})

	t.Run("should return no error for properly configured kafka backend", func(t *testing.T) {
		// Arrange:
		cfg := config.Config{
			Backend: types.BackendKafka,
			Kafka: config.Kafka{
				Brockers: []string{
					"localhost:8098",
				},
			},
		}

		// Act:
		err := cfg.Validate()

		// Assert:
		assert.Equal(t, nil, err)
	})
}
