package migration_runner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/core/migration_runner/mocks"
)

func TestFilteredSource_Open(t *testing.T) {
	mockSrc := mocks.NewDriver(t)
	filtered := migration_runner.NewFilteredSource(mockSrc, []string{"000002"})

	driver, err := filtered.Open("test-url")
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}

func TestFilteredSource_First(t *testing.T) {
	tests := []struct {
		name           string
		exclude        []string
		mockResponses  []uint
		mockErrors     []error
		expectedResult uint
		expectedError  error
	}{
		{
			name:           "no exclusions",
			exclude:        []string{},
			mockResponses:  []uint{1},
			mockErrors:     []error{nil},
			expectedResult: 1,
			expectedError:  nil,
		},
		{
			name:           "skip excluded version",
			exclude:        []string{"000001_"},
			mockResponses:  []uint{1, 2},
			mockErrors:     []error{nil, nil},
			expectedResult: 2,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrc := mocks.NewDriver(t)
			filtered := migration_runner.NewFilteredSource(mockSrc, tt.exclude)

			for i, resp := range tt.mockResponses {
				mockSrc.On("First").Return(resp, tt.mockErrors[i]).Once()
			}

			result, err := filtered.First()
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, result)
			mockSrc.AssertExpectations(t)
		})
	}
}

func TestFilteredSource_Next(t *testing.T) {
	tests := []struct {
		name           string
		exclude        []string
		startVersion   uint
		mockResponses  []uint
		mockErrors     []error
		expectedResult uint
		expectedError  error
	}{
		{
			name:           "no exclusions",
			exclude:        []string{},
			startVersion:   1,
			mockResponses:  []uint{2},
			mockErrors:     []error{nil},
			expectedResult: 2,
			expectedError:  nil,
		},
		{
			name:           "skip excluded version",
			exclude:        []string{"000002_"},
			startVersion:   1,
			mockResponses:  []uint{2, 3},
			mockErrors:     []error{nil, nil},
			expectedResult: 3,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrc := mocks.NewDriver(t)
			filtered := migration_runner.NewFilteredSource(mockSrc, tt.exclude)

			for i, resp := range tt.mockResponses {
				if i == 0 {
					mockSrc.On("Next", tt.startVersion).Return(resp, tt.mockErrors[i]).Once()
				} else {
					mockSrc.On("Next", tt.mockResponses[i-1]).Return(resp, tt.mockErrors[i]).Once()
				}
			}

			result, err := filtered.Next(tt.startVersion)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, result)
			mockSrc.AssertExpectations(t)
		})
	}
}

func TestFilteredSource_Prev(t *testing.T) {
	tests := []struct {
		name           string
		exclude        []string
		startVersion   uint
		mockResponses  []uint
		mockErrors     []error
		expectedResult uint
		expectedError  error
	}{
		{
			name:           "no exclusions",
			exclude:        []string{},
			startVersion:   2,
			mockResponses:  []uint{1},
			mockErrors:     []error{nil},
			expectedResult: 1,
			expectedError:  nil,
		},
		{
			name:           "skip excluded version",
			exclude:        []string{"000002_"},
			startVersion:   3,
			mockResponses:  []uint{2, 1, 0},
			mockErrors:     []error{nil, nil, nil},
			expectedResult: 1,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrc := mocks.NewDriver(t)
			filtered := migration_runner.NewFilteredSource(mockSrc, tt.exclude)

			mockSrc.On("Prev", tt.startVersion).Return(uint(1), nil).Once()

			result, err := filtered.Prev(tt.startVersion)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, result)
			mockSrc.AssertExpectations(t)
		})
	}
}

func TestFilteredSource_Close(t *testing.T) {
	mockSrc := mocks.NewDriver(t)
	filtered := migration_runner.NewFilteredSource(mockSrc, []string{"000002"})

	mockSrc.On("Close").Return(nil).Once()

	err := filtered.Close()
	assert.NoError(t, err)
	mockSrc.AssertExpectations(t)
}
