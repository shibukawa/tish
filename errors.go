package tish

import "errors"

var (
	ErrStackEmpty       = errors.New("directory stack empty")
	ErrRequireParameter = errors.New("require parameter")
)
