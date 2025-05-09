package migration_runner

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

type embedSource struct {
	fs       fs.FS
	path     string
	versions []uint
}

func NewEmbedSource(fs fs.FS, path string) source.Driver {
	return &embedSource{
		fs:   fs,
		path: path,
	}
}

func (e *embedSource) Open(url string) (source.Driver, error) {
	return e, nil
}

func (e *embedSource) Close() error {
	return nil
}

func (e *embedSource) First() (version uint, err error) {
	if err := e.loadVersions(); err != nil {
		return 0, err
	}

	if len(e.versions) == 0 {
		return 0, &fs.PathError{Op: "first", Path: e.path, Err: fs.ErrNotExist}
	}

	return e.versions[0], nil
}

func (e *embedSource) Prev(version uint) (prevVersion uint, err error) {
	if err := e.loadVersions(); err != nil {
		return 0, err
	}

	pos := e.findPos(version)
	if pos == -1 || pos == len(e.versions)-1 {
		return 0, &fs.PathError{Op: "prev", Path: e.path, Err: fs.ErrNotExist}
	}

	return e.versions[pos+1], nil
}

func (e *embedSource) Next(version uint) (nextVersion uint, err error) {
	if err := e.loadVersions(); err != nil {
		return 0, err
	}

	pos := e.findPos(version)
	if pos == -1 || pos == 0 {
		return 0, &fs.PathError{Op: "next", Path: e.path, Err: fs.ErrNotExist}
	}

	return e.versions[pos-1], nil
}

func (e *embedSource) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if err := e.loadVersions(); err != nil {
		return nil, "", err
	}

	if !e.versionExists(version) {
		return nil, "", &fs.PathError{Op: "read", Path: e.path, Err: fs.ErrNotExist}
	}

	identifier = fmt.Sprintf("%v.up.sql", version)
	file, err := e.fs.Open(path.Join(e.path, identifier))
	if err != nil {
		return nil, "", err
	}

	return file.(io.ReadCloser), identifier, nil
}

func (e *embedSource) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if err := e.loadVersions(); err != nil {
		return nil, "", err
	}

	if !e.versionExists(version) {
		return nil, "", &fs.PathError{Op: "read", Path: e.path, Err: fs.ErrNotExist}
	}

	identifier = fmt.Sprintf("%v.down.sql", version)
	file, err := e.fs.Open(path.Join(e.path, identifier))
	if err != nil {
		return nil, "", err
	}

	return file.(io.ReadCloser), identifier, nil
}

func (e *embedSource) loadVersions() error {
	if e.versions != nil {
		return nil
	}

	e.versions = make([]uint, 0)

	entries, err := fs.ReadDir(e.fs, e.path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		var version uint
		if _, err := fmt.Sscanf(entry.Name(), "%v.up.sql", &version); err != nil {
			continue
		}

		e.versions = append(e.versions, version)
	}

	// Sort versions descending
	for i := 0; i < len(e.versions)-1; i++ {
		for j := i + 1; j < len(e.versions); j++ {
			if e.versions[i] < e.versions[j] {
				e.versions[i], e.versions[j] = e.versions[j], e.versions[i]
			}
		}
	}

	return nil
}

func (e *embedSource) findPos(version uint) int {
	if err := e.loadVersions(); err != nil {
		return -1
	}

	for i, v := range e.versions {
		if v == version {
			return i
		}
	}

	return -1
}

func (e *embedSource) versionExists(version uint) bool {
	return e.findPos(version) >= 0
}
