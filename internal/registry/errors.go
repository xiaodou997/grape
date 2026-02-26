package registry

import "errors"

var (
	ErrPackageNotFound  = errors.New("package not found")
	ErrTarballNotFound  = errors.New("tarball not found")
	ErrInvalidPackage   = errors.New("invalid package name")
	ErrStorageFailed    = errors.New("storage operation failed")
)
