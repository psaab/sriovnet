package filesystem

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// fakeFs is implemented in terms of afero
type fakeFs struct {
	a afero.Afero
}

// NewFakeFs returns a fake Filesystem that exists in-memory, useful for unit tests
func NewFakeFs() Filesystem {
	return &fakeFs{a: afero.Afero{Fs: afero.NewMemMapFs()}}
}

// Stat via afero.Fs.Stat
func (fs *fakeFs) Stat(name string) (os.FileInfo, error) {
	return fs.a.Fs.Stat(name)
}

// Create via afero.Fs.Create
func (fs *fakeFs) Create(name string) (File, error) {
	file, err := fs.a.Fs.Create(name)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// Rename via afero.Fs.Rename
func (fs *fakeFs) Rename(oldpath, newpath string) error {
	return fs.a.Fs.Rename(oldpath, newpath)
}

// MkdirAll via afero.Fs.MkdirAll
func (fs *fakeFs) MkdirAll(path string, perm os.FileMode) error {
	return fs.a.Fs.MkdirAll(path, perm)
}

// Chtimes via afero.Fs.Chtimes
func (fs *fakeFs) Chtimes(name string, atime, mtime time.Time) error {
	return fs.a.Fs.Chtimes(name, atime, mtime)
}

// ReadFile via afero.ReadFile
func (fs *fakeFs) ReadFile(filename string) ([]byte, error) {
	return fs.a.ReadFile(filename)
}

// TempDir via afero.TempDir
func (fs *fakeFs) TempDir(dir, prefix string) (string, error) {
	return fs.a.TempDir(dir, prefix)
}

// TempFile via afero.TempFile
func (fs *fakeFs) TempFile(dir, prefix string) (File, error) {
	file, err := fs.a.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// ReadDir via afero.ReadDir
func (fs *fakeFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	return fs.a.ReadDir(dirname)
}

// Walk via afero.Walk
func (fs *fakeFs) Walk(root string, walkFn filepath.WalkFunc) error {
	return fs.a.Walk(root, walkFn)
}

// RemoveAll via afero.RemoveAll
func (fs *fakeFs) RemoveAll(path string) error {
	return fs.a.RemoveAll(path)
}

// Remove via afero.Remove
func (fs *fakeFs) Remove(name string) error {
	return fs.a.Remove(name)
}

// fakeFile implements File; for use with fakeFs
type fakeFile struct {
	file afero.File
}

// Name via afero.File.Name
func (file *fakeFile) Name() string {
	return file.file.Name()
}

// Write via afero.File.Write
func (file *fakeFile) Write(b []byte) (n int, err error) {
	return file.file.Write(b)
}

// Sync via afero.File.Sync
func (file *fakeFile) Sync() error {
	return file.file.Sync()
}

// Close via afero.File.Close
func (file *fakeFile) Close() error {
	return file.file.Close()
}
