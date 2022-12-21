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

type VMan struct {
	trans *transcoder.Transcoder
}

// FFmpeg params
const (
	Copy   = "copy"
	MpegTS = "mpegts"
)

func New() *VMan {
	v := new(VMan)
	v.trans = new(transcoder.Transcoder)
	conf, _ := ffmpeg.Configure()
	conf.FfprobeBin = strings.Trim(conf.FfprobeBin, "\r")
	conf.FfmpegBin = strings.Trim(conf.FfmpegBin, "\r")
	v.trans.SetConfiguration(conf)
	return v
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

func (v *VMan) NewDefaultConfig() FFConfig {
	return FFConfig{
		UseHWAccel: true,
		AudioCodec: Copy,
		VideoCodec: Copy,
		OutArgs:    []string{"-bsf:v", "h264_mp4toannexb"},
		OutFormat:  MpegTS,
	}
}

type FFConfig struct {
	UseHWAccel             bool
	AudioCodec, VideoCodec string
	InArgs, OutArgs        []string
	OutFormat              string
}

func getIntermediateFilename(video string) string {
	return strings.Replace(video, ".MP4", ".ts", -1)
}

func getMergedOutputFilename(video string) string {
	return filepath.Join(filepath.Dir(video),
		fmt.Sprintf("%s-merged%s", strings.Replace(filepath.Base(video), filepath.Ext(video), "", -1), filepath.Ext(video)),
	)
}

//nolint:golint,unused,errcheck
func (v *VMan) merge(output string, bar *mpb.Bar, ffConfig FFConfig, videos ...string) error {
	err := v.trans.InitializeEmptyTranscoder()
	if err != nil {
		return err
	}

	if ffConfig.UseHWAccel {
		v.trans.MediaFile().SetRawInputArgs([]string{"-hwaccel", "cuda"})
	}

	err = v.trans.SetInputPath(fmt.Sprintf("concat:%s",
		strings.Join(videos, "|"),
	),
	)
	if err != nil {
		return err
	}

	err = v.trans.SetOutputPath(
		output)
	if err != nil {
		return err
	}
	v.trans.MediaFile().SetVideoCodec(ffConfig.VideoCodec)
	v.trans.MediaFile().SetAudioCodec(ffConfig.AudioCodec)

	done := v.trans.Run(true)

	progress := v.trans.Output()

	for msg := range progress {
		s, _ := strconv.Atoi(msg.FramesProcessed)
		bar.SetCurrent(int64(s))
	}

	err = <-done
	for _, file := range videos {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return err
}

//nolint:golint,unused,errcheck
func (v *VMan) convert(video, output string, bar *mpb.Bar, ffConfig FFConfig) error {
	err := v.trans.InitializeEmptyTranscoder()
	if err != nil {
		return err
	}

	if ffConfig.UseHWAccel {
		v.trans.MediaFile().SetRawInputArgs([]string{"-hwaccel", "cuda"})
	}
	err = v.trans.SetInputPath(video)
	if err != nil {
		return err
	}

	err = v.trans.SetOutputPath(
		output)
	if err != nil {
		return err
	}
	v.trans.MediaFile().SetVideoCodec(ffConfig.VideoCodec)
	v.trans.MediaFile().SetAudioCodec(ffConfig.AudioCodec)
	v.trans.MediaFile().SetRawOutputArgs(ffConfig.OutArgs)
	v.trans.MediaFile().SetOutputFormat(ffConfig.OutFormat)

	done := v.trans.Run(true)

	progress := v.trans.Output()

	for msg := range progress {
		s, _ := strconv.Atoi(msg.FramesProcessed)
		bar.SetCurrent(int64(s))
	}

	err = <-done
	return err
}

func (v *VMan) Merge(videos ...string) error {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))
	nfiles := len(videos)

	totalFrames := 0
	intermediates := []string{}
	for i := 0; i < nfiles; i++ {
		ffprobe := utils.NewFFprobe(nil)

		head, err := ffprobe.Frames(videos[i])
		if err != nil {
			return err
		}
		totalFrames += head.Streams[0].Frames

		bar := getBar(p, head.Streams[0].Frames, fmt.Sprintf("%s%s", "✂️", filepath.Base(videos[i])))

		wg.Add(1)
		go func(current int) {
			defer wg.Done()
			intermediate := getIntermediateFilename(videos[current])
			intermediates = append(intermediates, intermediate)
			err := v.convert(videos[current], intermediate, bar, v.NewDefaultConfig())
			if err != nil {
				log.Fatal(err.Error())
			}
		}(i)
	}
	p.Wait()

	nonAsync := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))
	newBar := getBar(nonAsync, totalFrames, fmt.Sprintf("%s%s", "🐈", filepath.Base(videos[0])))

	err := v.merge(getMergedOutputFilename(videos[0]), newBar, v.NewDefaultConfig(), intermediates...)
	if err != nil {
		log.Fatal(err.Error())
	}
	nonAsync.Wait()

	return nil
}

//nolint:golint,unused,errcheck
func (v *VMan) extractGPMF(input string) (*[]byte, error) {
	err := v.trans.InitializeEmptyTranscoder()
	if err != nil {
		return nil, err
	}

	err = v.trans.SetInputPath(input)
	if err != nil {
		return nil, err
	}

	r, err := v.trans.CreateOutputPipe("rawvideo")
	if err != nil {
		return nil, err
	}

	v.trans.MediaFile().SetRawOutputArgs([]string{"-map", "0:3"})
	v.trans.MediaFile().SetOutputFormat("rawvideo")
	v.trans.MediaFile().SetVideoCodec("copy")

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

	done := v.trans.Run(false)

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
