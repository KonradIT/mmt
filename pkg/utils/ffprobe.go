package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	mErrors "github.com/konradit/mmt/pkg/errors"
)

// FFprobe command wrapper

type FFprobe struct {
	ProgramPath string
}

type VideoSizeResponse struct {
	Programs []interface{} `json:"programs"`
	Streams  []struct {
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
	} `json:"streams"`
}

type FramesResponse struct {
	Programs []interface{} `json:"programs"`
	Streams  []struct {
		Frames int `json:"nb_frames,string"`
	} `json:"streams"`
}

type DurationResponse struct {
	Programs []interface{} `json:"programs"`
	Streams  []struct {
		Duration float32 `json:"duration,string"`
	} `json:"streams"`
}

type GPSLocation struct {
	Format struct {
		Tags struct {
			Location string `json:"location"`
		} `json:"tags"`
	} `json:"format"`
}

type StreamsResponse struct {
	Streams []struct {
		Index          int    `json:"index"`
		CodecTagString string `json:"codec_tag_string"`
	} `json:"streams"`
}

func NewFFprobe(path *string) FFprobe {
	ff := FFprobe{}
	if path == nil {
		ff.ProgramPath = "ffprobe"
	} else {
		ff.ProgramPath = *path
	}
	return ff
}

func (f *FFprobe) executeGetFormat(path string) ([]byte, error) {
	args := []string{
		"-select_streams",
		"v:0",
		"-show_format",
		"-of",
		"json",
		path,
	}
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(f.ProgramPath, args...) // #nosec
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (f *FFprobe) executeGetInfo(path string, entries ...string) ([]byte, error) {
	args := []string{
		"-select_streams",
		"v:0",
		"-show_entries",
		fmt.Sprintf("stream=%s", strings.Join(entries, ",")),
		"-of",
		"json",
		path,
	}
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(f.ProgramPath, args...) // #nosec
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (f *FFprobe) Streams(path string) (*StreamsResponse, error) {
	result := StreamsResponse{}

	args := []string{
		"-show_streams",
		"-of",
		"json",
		path,
	}
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(f.ProgramPath, args...) // #nosec
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *FFprobe) VideoSize(path string) (*VideoSizeResponse, error) {
	result := VideoSizeResponse{}
	out, err := f.executeGetInfo(path, "width", "height", "r_frame_rate")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *FFprobe) Frames(path string) (*FramesResponse, error) {
	result := FramesResponse{}
	out, err := f.executeGetInfo(path, "nb_frames")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *FFprobe) Duration(path string) (*DurationResponse, error) {
	result := DurationResponse{}
	out, err := f.executeGetInfo(path, "duration")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *FFprobe) GPSLocation(path string) (*Location, error) {
	result := GPSLocation{}
	out, err := f.executeGetFormat(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}

	baseCoordinates := regexp.MustCompile(`([+-]?(\d+\.?\d*)|(\.\d+))`)

	parts := baseCoordinates.FindAllStringSubmatch(result.Format.Tags.Location, -1)

	if len(parts) != 2 {
		return nil, mErrors.ErrInvalidCoordinatesFormat
	}

	latitude, err := strconv.ParseFloat(parts[0][0], 32)
	if err != nil {
		return nil, err
	}

	longitude, err := strconv.ParseFloat(parts[1][0], 32)
	if err != nil {
		return nil, err
	}

	parsed := Location{
		Latitude:  latitude,
		Longitude: longitude,
	}
	return &parsed, nil
}
