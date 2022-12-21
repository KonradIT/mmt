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

	"github.com/vbauerster/mpb/v8"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
)

type VMan struct {
	trans *transcoder.Transcoder
}

// FFmpeg params
const (
	Copy = "copy"
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

func (v *VMan) NewDefaultConfig() FFConfig {
	return FFConfig{
		UseHWAccel: true,
		AudioCodec: Copy,
		VideoCodec: Copy,
		InArgs:     []string{},
		OutArgs:    []string{},
	}
}

type FFConfig struct {
	UseHWAccel             bool
	AudioCodec, VideoCodec string
	InArgs, OutArgs        []string
	OutFormat              string
}

func getMergedOutputFilename(video string) string {
	return filepath.Join(filepath.Dir(video),
		fmt.Sprintf("%s-merged%s", strings.Replace(filepath.Base(video), filepath.Ext(video), "", -1), filepath.Ext(video)),
	)
}

func (v *VMan) merge(output string, bar *mpb.Bar, ffConfig FFConfig, videos ...string) error {
	err := v.trans.InitializeEmptyTranscoder()
	if err != nil {
		return err
	}

	if ffConfig.UseHWAccel {
		ffConfig.InArgs = append(ffConfig.InArgs, []string{"-hwaccel", "cuda"}...)
	}

	file, err := ioutil.TempFile(filepath.Dir(videos[0]), "filelist.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	for _, video := range videos {
		a := fmt.Sprintf("file '%s'\n", video)
		_, err = file.WriteString(a)
		if err != nil {
			return err
		}
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	err = v.trans.SetInputPath(file.Name())
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
	v.trans.MediaFile().SetRawInputArgs(ffConfig.InArgs)
	v.trans.MediaFile().SetRawOutputArgs(ffConfig.OutArgs)

	done := v.trans.Run(true)

	progress := v.trans.Output()

	for msg := range progress {
		s, _ := strconv.Atoi(msg.FramesProcessed)
		bar.SetCurrent(int64(s))
	}

	err = <-done
	return err
}

func (v *VMan) Merge(bar *mpb.Bar, videos ...string) error {

	mergeConfig := v.NewDefaultConfig()
	mergeConfig.InArgs = append(mergeConfig.InArgs, []string{"-f", "concat", "-safe", "0"}...)
	mergeConfig.OutArgs = append(mergeConfig.OutArgs, []string{"-map", "0:0", "-map", "0:1", "-map", "0:3"}...)

	err := v.merge(getMergedOutputFilename(videos[0]), bar, mergeConfig, videos...)
	if err != nil {
		log.Fatal(err.Error())
	}

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
