package migration_runner

import (
	"io"

	"github.com/golang-migrate/migrate/v4/source"
)

// chainedSource combines multiple sources into one
type chainedSource struct {
	primary  source.Driver
	fallback source.Driver
}

// Open implements source.Driver
func (c *chainedSource) Open(url string) (source.Driver, error) {
	return c, nil
}

// First implements source.Driver
func (c *chainedSource) First() (uint, error) {
	version, err := c.primary.First()
	if err == nil {
		return version, nil
	}
	return c.fallback.First()
}

// Next implements source.Driver
func (c *chainedSource) Next(version uint) (uint, error) {
	nextVersion, err := c.primary.Next(version)
	if err == nil {
		return nextVersion, nil
	}
	return c.fallback.Next(version)
}

// ReadUp implements source.Driver
func (c *chainedSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := c.primary.ReadUp(version)
	if err == nil {
		return r, identifier, nil
	}
	return c.fallback.ReadUp(version)
}

// ReadDown implements source.Driver
func (c *chainedSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := c.primary.ReadDown(version)
	if err == nil {
		return r, identifier, nil
	}
	return c.fallback.ReadDown(version)
}

// Close implements source.Driver
func (c *chainedSource) Close() error {
	err1 := c.primary.Close()
	err2 := c.fallback.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Prev implements source.Driver
func (c *chainedSource) Prev(version uint) (uint, error) {
	prevVersion, err := c.primary.Prev(version)
	if err == nil {
		return prevVersion, nil
	}
	return c.fallback.Prev(version)
}
