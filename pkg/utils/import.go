package utils

import "time"

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

type Import interface {
	Import(params ImportParams) (*Result, error)
}
