package camera

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// ConnectionType represents how the camera is connected.
type ConnectionType string

const (
	SDCard  ConnectionType = "sd_card"
	Connect ConnectionType = "connect"
)

// SortOptions controls how imported media is organized on disk.
type SortOptions struct {
	ByLocation bool
	ByCamera   bool
}

// ImportParams holds all parameters for a media import operation.
type ImportParams struct {
	Input, Output, CameraName string
	SkipAuxiliaryFiles        bool
	DateFormat                string
	BufferSize                int
	Prefix                    string
	DateRange                 []time.Time
	TagNames                  []string
	Connection                ConnectionType
	Sort                      SortOptions
}

// Result holds the outcome of an import operation.
type Result struct {
	FilesImported    int
	FilesNotImported []string
	Errors           []error
}

// ResultCounter is a thread-safe counter for tracking import progress
// across concurrent goroutines.
type ResultCounter struct {
	mu               sync.Mutex
	Errors           []error
	FilesNotImported []string
	FilesImported    int
}

// SetFailure records a failed file import.
func (rc *ResultCounter) SetFailure(err error, file string) {
	rc.mu.Lock()
	rc.Errors = append(rc.Errors, err)
	rc.FilesNotImported = append(rc.FilesNotImported, file)
	rc.mu.Unlock()
}

// SetSuccess increments the successful import count.
func (rc *ResultCounter) SetSuccess() {
	rc.mu.Lock()
	rc.FilesImported++
	rc.mu.Unlock()
}

// Get returns a snapshot of the current results.
func (rc *ResultCounter) Get() Result {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return Result{
		FilesImported:    rc.FilesImported,
		FilesNotImported: rc.FilesNotImported,
		Errors:           rc.Errors,
	}
}

// BarType controls the decorator style for progress bars.
type BarType int

const (
	IoTX BarType = iota
	Percentage
)

// GetNewBar creates a new progress bar attached to the given progress container.
func GetNewBar(progressBar *mpb.Progress, total int64, filename string, barType BarType) *mpb.Bar {
	decorator := decor.CountersKiloByte("% .2f / % .2f")
	if barType == Percentage {
		decorator = decor.Percentage(decor.WCSyncSpace)
	}
	return progressBar.AddBar(total,
		mpb.PrependDecorators(
			decor.Name(color.CyanString(fmt.Sprintf("%s: ", filename))),
			decorator,
		),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
			),
		),
	)
}
