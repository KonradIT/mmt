package errors

import "errors"

var (
	ErrNoCameraDetected        = errors.New("No camera detected")
	ErrUnsupportedCamera       = errors.New("Unsupported camera")
	ErrUnrecognizedMediaFormat = errors.New("Media format unrecognized")
	ErrUnsupportedConnection   = errors.New("Unsupported connection")
)
