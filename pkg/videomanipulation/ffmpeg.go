package videomanipulation

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
)

var trans *transcoder.Transcoder

func New() {
	trans = new(transcoder.Transcoder)
	conf, _ := ffmpeg.Configure()
	conf.FfprobeBin = strings.Trim(conf.FfprobeBin, "\r")
	conf.FfmpegBin = strings.Trim(conf.FfmpegBin, "\r")
	trans.SetConfiguration(conf)
}

//nolint:golint,unused,errcheck
func extractGPMF(input string) (*[]byte, error) {
	err := trans.InitializeEmptyTranscoder()
	if err != nil {
		return nil, err
	}

	err = trans.SetInputPath(input)
	if err != nil {
		return nil, err
	}

	r, err := trans.CreateOutputPipe("rawvideo")
	if err != nil {
		return nil, err
	}

	trans.MediaFile().SetRawOutputArgs([]string{"-map", "0:3"})
	trans.MediaFile().SetOutputFormat("rawvideo")
	trans.MediaFile().SetVideoCodec("copy")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	extractError := make(chan error)
	extractData := make(chan []byte)

	go func() {
		defer r.Close()
		defer wg.Done()

		data, err := ioutil.ReadAll(r)

		extractData <- data
		extractError <- err
	}()

	done := trans.Run(false)

	err = <-done
	if err != nil {
		return nil, err
	}

	if extractErr := <-extractError; extractErr != nil {
		return nil, extractErr
	}

	wg.Wait()

	dataExtracted := <-extractData
	return &dataExtracted, nil
}
