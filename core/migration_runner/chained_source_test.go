package migration_runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	runnerMocks "gitlab.com/jideobs/nebularcore/core/migration_runner/mocks"
)

func TestChainedSource_Open(t *testing.T) {
	primary := runnerMocks.NewDriver(t)
	fallback := runnerMocks.NewDriver(t)

	chained := &chainedSource{
		primary:  primary,
		fallback: fallback,
	}

	driver, err := chained.Open("test-url")
	assert.NoError(t, err)
	assert.Equal(t, chained, driver)
}

func TestChainedSource_First(t *testing.T) {
	tests := []struct {
		name            string
		primaryErr      error
		primaryVersion  uint
		fallbackErr     error
		fallbackVersion uint
		expectedVersion uint
		expectedErr     error
	}{
		{
			name:            "primary succeeds",
			primaryErr:      nil,
			primaryVersion:  1,
			fallbackErr:     nil,
			fallbackVersion: 2,
			expectedVersion: 1,
			expectedErr:     nil,
		},
		{
			name:            "primary fails, fallback succeeds",
			primaryErr:      errors.New("primary error"),
			primaryVersion:  0,
			fallbackErr:     nil,
			fallbackVersion: 2,
			expectedVersion: 2,
			expectedErr:     nil,
		},
		{
			name:            "both fail",
			primaryErr:      errors.New("primary error"),
			primaryVersion:  0,
			fallbackErr:     errors.New("fallback error"),
			fallbackVersion: 0,
			expectedVersion: 0,
			expectedErr:     errors.New("fallback error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := runnerMocks.NewDriver(t)
			fallback := runnerMocks.NewDriver(t)

			primary.On("First").Return(tt.primaryVersion, tt.primaryErr).Once()
			if tt.primaryErr != nil {
				fallback.On("First").Return(tt.fallbackVersion, tt.fallbackErr).Once()
			}

			chained := &chainedSource{
				primary:  primary,
				fallback: fallback,
			}

			version, err := chained.First()
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedVersion, version)

			primary.AssertExpectations(t)
			fallback.AssertExpectations(t)
		})
	}
}

func TestChainedSource_Next(t *testing.T) {
	tests := []struct {
		name            string
		startVersion    uint
		primaryErr      error
		primaryVersion  uint
		fallbackErr     error
		fallbackVersion uint
		expectedVersion uint
		expectedErr     error
	}{
		{
			name:            "primary succeeds",
			startVersion:    1,
			primaryErr:      nil,
			primaryVersion:  2,
			fallbackErr:     nil,
			fallbackVersion: 3,
			expectedVersion: 2,
			expectedErr:     nil,
		},
		{
			name:            "primary fails, fallback succeeds",
			startVersion:    1,
			primaryErr:      errors.New("primary error"),
			primaryVersion:  0,
			fallbackErr:     nil,
			fallbackVersion: 2,
			expectedVersion: 2,
			expectedErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := runnerMocks.NewDriver(t)
			fallback := runnerMocks.NewDriver(t)

			primary.On("Next", tt.startVersion).Return(tt.primaryVersion, tt.primaryErr).Once()
			if tt.primaryErr != nil {
				fallback.On("Next", tt.startVersion).Return(tt.fallbackVersion, tt.fallbackErr).Once()
			}

			chained := &chainedSource{
				primary:  primary,
				fallback: fallback,
			}

			version, err := chained.Next(tt.startVersion)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedVersion, version)

			primary.AssertExpectations(t)
			fallback.AssertExpectations(t)
		})
	}
}

func TestChainedSource_ReadUp(t *testing.T) {
	tests := []struct {
		name            string
		version         uint
		primaryErr      error
		primaryContent  string
		fallbackErr     error
		fallbackContent string
		expectedErr     error
	}{
		{
			name:            "primary succeeds",
			version:         1,
			primaryErr:      nil,
			primaryContent:  "primary content",
			fallbackErr:     nil,
			fallbackContent: "fallback content",
			expectedErr:     nil,
		},
		// {
		// 	name:            "primary fails, fallback succeeds",
		// 	version:         1,
		// 	primaryErr:      errors.New("primary error"),
		// 	primaryContent:  "",
		// 	fallbackErr:     nil,
		// 	fallbackContent: "fallback content",
		// 	expectedContent: "fallback content",
		// 	expectedErr:     nil,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := runnerMocks.NewDriver(t)
			fallback := runnerMocks.NewDriver(t)
			mockReadCloser := runnerMocks.NewReadCloser(t)
			mockReadCloser.On("Read", mock.Anything).Return(len(tt.primaryContent), nil).Once()
			// mockReadCloser.On("Close").Return(nil).Once()

			primary.On("ReadUp", tt.version).Return(
				mockReadCloser,
				"primary",
				tt.primaryErr,
			).Once()

			if tt.primaryErr != nil {
				fallback.On("ReadUp", tt.version).Return(
					mockReadCloser,
					"fallback",
					tt.fallbackErr,
				).Once()
			}

			chained := &chainedSource{
				primary:  primary,
				fallback: fallback,
			}

			reader, _, err := chained.ReadUp(tt.version)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				content := make([]byte, len(tt.primaryContent))
				_, err := reader.Read(content)
				assert.NoError(t, err)
			}

			primary.AssertExpectations(t)
			fallback.AssertExpectations(t)
		})
	}
}

func TestChainedSource_Close(t *testing.T) {
	tests := []struct {
		name        string
		primaryErr  error
		fallbackErr error
		expectedErr error
	}{
		{
			name:        "both succeed",
			primaryErr:  nil,
			fallbackErr: nil,
			expectedErr: nil,
		},
		{
			name:        "primary fails",
			primaryErr:  errors.New("primary error"),
			fallbackErr: nil,
			expectedErr: errors.New("primary error"),
		},
		{
			name:        "fallback fails",
			primaryErr:  nil,
			fallbackErr: errors.New("fallback error"),
			expectedErr: errors.New("fallback error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := runnerMocks.NewDriver(t)
			fallback := runnerMocks.NewDriver(t)

			primary.On("Close").Return(tt.primaryErr).Once()
			fallback.On("Close").Return(tt.fallbackErr).Once()

			chained := &chainedSource{
				primary:  primary,
				fallback: fallback,
			}

			err := chained.Close()
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			primary.AssertExpectations(t)
			fallback.AssertExpectations(t)
		})
	}
}
