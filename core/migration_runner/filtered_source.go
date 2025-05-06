package migration_runner

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

// filteredSource wraps a source.Driver and filters out excluded files
type filteredSource struct {
	source.Driver
	exclude []string
}

// NewFilteredSource creates a new filteredSource that wraps the given driver and excludes the specified files
func NewFilteredSource(driver source.Driver, exclude []string) source.Driver {
	return &filteredSource{
		Driver:  driver,
		exclude: exclude,
	}
}

// Open implements source.Driver
func (f *filteredSource) Open(url string) (source.Driver, error) {
	return f, nil
}

// First implements source.Driver
func (f *filteredSource) First() (version uint, err error) {
	for {
		version, err = f.Driver.First()
		if err != nil {
			return 0, err
		}
		if !f.isExcluded(version) {
			return version, nil
		}
	}
}

// Next implements source.Driver
func (f *filteredSource) Next(version uint) (nextVersion uint, err error) {
	for {
		nextVersion, err = f.Driver.Next(version)
		if err != nil {
			return 0, err
		}
		if !f.isExcluded(nextVersion) {
			return nextVersion, nil
		}
		version = nextVersion
	}
}

// isExcluded checks if a version should be excluded
func (f *filteredSource) isExcluded(version uint) bool {
	for _, exclude := range f.exclude {
		if strings.HasPrefix(exclude, fmt.Sprintf("%06d", version)) {
			return true
		}
	}
	return false
}

// Close implements source.Driver
func (f *filteredSource) Close() error {
	return f.Driver.Close()
}

// Prev implements source.Driver
func (f *filteredSource) Prev(version uint) (prevVersion uint, err error) {
	for {
		fmt.Println("Version ", version)
		prevVersion, err = f.Driver.Prev(version)
		if err != nil {
			return 0, err
		}
		fmt.Println("PrevVersion ", prevVersion)
		if !f.isExcluded(prevVersion) {
			return prevVersion, nil
		}
		version = prevVersion
	}
}
