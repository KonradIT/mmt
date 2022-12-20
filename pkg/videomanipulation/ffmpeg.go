package videomanipulation

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/konradit/mmt/pkg/utils"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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

func getBar(progressBar *mpb.Progress, total int, name string) *mpb.Bar {
	return progressBar.AddBar(int64(total),
		mpb.PrependDecorators(
			decor.Name(name),
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
			),
		),
	)
}

//nolint:golint,unused,errcheck
func merge(bar *mpb.Bar, videos ...string) error {
	err := trans.InitializeEmptyTranscoder()
	if err != nil {
		return err
	}

	trans.MediaFile().SetRawInputArgs([]string{"-hwaccel", "cuda"})
	intermediates := []string{}

	for _, x := range videos {
		intermediates = append(intermediates, strings.Replace(x, ".MP4", ".ts", -1))
	}
	err = trans.SetInputPath(fmt.Sprintf("concat:%s",
		strings.Join(intermediates, "|"),
	),
	)
	if err != nil {
		return err
	}

	err = trans.SetOutputPath(
		strings.Replace(videos[0], "-01.MP4", "-merged.MP4", -1))
	if err != nil {
		return err
	}
	trans.MediaFile().SetVideoCodec("copy")
	trans.MediaFile().SetAudioCodec("copy")

	done := trans.Run(true)

	progress := trans.Output()

	for msg := range progress {
		s, _ := strconv.Atoi(msg.FramesProcessed)
		bar.SetCurrent(int64(s))
	}

	err = <-done
	for _, file := range intermediates {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return err
}

//nolint:golint,unused,errcheck
func convert(video string, bar *mpb.Bar) error {
	err := trans.InitializeEmptyTranscoder()
	if err != nil {
		return err
	}

	trans.MediaFile().SetRawInputArgs([]string{"-hwaccel", "cuda"})
	err = trans.SetInputPath(video)
	if err != nil {
		return err
	}

	err = trans.SetOutputPath(
		strings.Replace(video, ".MP4", ".ts", -1))
	if err != nil {
		return err
	}
	trans.MediaFile().SetVideoCodec("copy")
	trans.MediaFile().SetAudioCodec("copy")
	trans.MediaFile().SetRawOutputArgs([]string{"-bsf:v", "h264_mp4toannexb"})
	trans.MediaFile().SetOutputFormat("mpegts")

	done := trans.Run(true)

	progress := trans.Output()

	for msg := range progress {
		s, _ := strconv.Atoi(msg.FramesProcessed)
		bar.SetCurrent(int64(s))
	}

	err = <-done
	return err
}

func Merge(videos ...string) error {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))
	nfiles := len(videos)

	totalFrames := 0
	for i := 0; i < nfiles; i++ {
		ffprobe := utils.NewFFprobe(nil)

		head, err := ffprobe.Frames(videos[i])
		if err != nil {
			return err
		}
		totalFrames += head.Streams[0].Frames
		bar := getBar(p, head.Streams[0].Frames, filepath.Base(videos[i]))

		wg.Add(1)
		go func(current int) {
			defer wg.Done()
			err := convert(videos[current], bar)
			if err != nil {
				log.Fatal(err.Error())
			}
		}(i)
	}
	p.Wait()

	nonAsync := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))
	newBar := getBar(nonAsync, totalFrames, "Merging")
	err := merge(newBar, videos...)
	if err != nil {
		log.Fatal(err.Error())
	}
	nonAsync.Wait()

	return nil
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
