package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNoCameraDetected         = errors.New("no camera detected")
	ErrUnrecognizedMediaFormat  = errors.New("media format unrecognized")
	ErrUnsupportedConnection    = errors.New("unsupported connection")
	ErrInvalidFile              = errors.New("file invalid (not a video or photo)")
	ErrNoRecognizedSRTFormat    = errors.New("srt file invalid format (could not read from predefined presets)")
	ErrGeneric                  = errors.New("generic error")
	ErrNoGPS                    = errors.New("no GPS data found")
	ErrInvalidCoordinatesFormat = errors.New("invalid coordinates format")
)

func ErrInvalidSuppliedData(data any) error {
	return fmt.Errorf("invalid data: %v", data)
}

func ErrUnsupportedCamera(camera string) error {
	return fmt.Errorf("camera %s is not supported", camera)
}

func ErrNotFound(item string) error {
	return fmt.Errorf("unable to find %s", item)
}
