// Package camera defines the core interfaces for camera backends and provides
// a self-registration registry so new cameras can be added without modifying
// existing dispatch code.
package camera

import "time"

// Camera is the core interface that every camera backend must implement.
type Camera interface {
	// Name returns the CLI identifier (e.g. "gopro", "dji").
	Name() string

	// Import copies media from the camera to the output directory.
	Import(params ImportParams) (*Result, error)

	// Detect auto-detects whether this camera is connected and returns the
	// input path and connection type.
	Detect() (input string, conn ConnectionType, err error)

	// GuessFromPath reports whether root looks like this camera's storage layout.
	GuessFromPath(root string) bool
}

// Updater is optionally implemented by cameras that support firmware updates.
type Updater interface {
	UpdateFirmware(input string, opts UpdateOptions) error
}

// UpdateOptions holds parameters for firmware updates.
type UpdateOptions struct {
	Model string // sub-model identifier (e.g. Insta360 model variant)
}

// CalendarProvider is optionally implemented by cameras that can report
// the dates on which media was captured.
type CalendarProvider interface {
	CaptureDates(input string, conn ConnectionType) ([]time.Time, error)
}
