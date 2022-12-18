package utils

import (
	"bytes"
	"encoding/json"
	"os/exec"
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

func NewFFprobe(path *string) FFprobe {
	ff := FFprobe{}
	if path == nil {
		ff.ProgramPath = "ffprobe"
	} else {
		ff.ProgramPath = *path
	}
	return ff
}

func (f *FFprobe) VideoSize(path string) (*VideoSizeResponse, error) {
	args := []string{
		"-select_streams",
		"v:0",
		"-show_entries",
		"stream=width,height,r_frame_rate",
		"-of",
		"json",
		path,
	}

	cmd := exec.Command(f.ProgramPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	result := VideoSizeResponse{}
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
