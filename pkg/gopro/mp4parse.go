package gopro

import (
	"errors"
	"os"

	"github.com/abema/go-mp4"
)

type HiLights struct {
	Count      int   `json:"count"`
	Timestamps []int `json:"timestamps"`
}

func BoxTypeHMMT() mp4.BoxType { return mp4.StrToBoxType("HMMT") }

type HMMT struct {
	mp4.Box
	Count   int   `mp4:"0,size=32,len=dynamic,int"`
	Entries []int `mp4:"1,size=32,len=dynamic,int"`
}

func (*HMMT) GetType() mp4.BoxType {
	return BoxTypeHMMT()
}

func (h *HMMT) GetFieldLength(name string, ctx mp4.Context) uint {
	return uint(h.Count)
}

func GetHiLights(path string) (*HiLights, error) {
	mp4.AddBoxDef(&HMMT{}, 0)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	hmmtData := &HMMT{}

	_, _ = mp4.ReadBoxStructure(f, func(h *mp4.ReadHandle) (interface{}, error) {
		if h.BoxInfo.IsSupportedType() && h.BoxInfo.Type.String() == "moov" || h.BoxInfo.Type.String() == "udta" || h.BoxInfo.Type.String() == "HMMT" {
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			if h.BoxInfo.Type.String() == "HMMT" {
				hmmtData = box.(*HMMT)
			}
			return h.Expand()
		}
		return nil, nil
	})

	if hmmtData != nil {
		return &HiLights{
			hmmtData.Count,
			hmmtData.Entries}, nil
	}
	return nil, errors.New("No data found")
}
