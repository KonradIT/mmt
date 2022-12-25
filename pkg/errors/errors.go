package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNoCameraDetected         = errors.New("No camera detected")
	ErrUnrecognizedMediaFormat  = errors.New("Media format unrecognized")
	ErrUnsupportedConnection    = errors.New("Unsupported connection")
	ErrInvalidFile              = errors.New("file invalid (not a video or photo)")
	ErrNoRecognizedSRTFormat    = errors.New("SRT file invalid format (could not read from predefined presets)")
	ErrGeneric                  = errors.New("Generic error")
	ErrNoGPS                    = errors.New("No GPS data found")
	ErrInvalidCoordinatesFormat = errors.New("Invalid coordinates format")
	ErrUnsupportedCamera        = func(camera string) error { return fmt.Errorf("camera %s is not supported", camera) }
	ErrNotFound                 = func(item string) error { return fmt.Errorf("Unable to find %s", item) }
)
