package models

import (
	"encoding/xml"
	godashHttp "github.com/massimo-gollo/godash/http"
)

type MpdFull struct {
	Mpd         *godashHttp.MPD
	MpdMetadata *MpdMetadata
}

func NewMpdFull() *MpdFull {
	return &MpdFull{
		Mpd:         &godashHttp.MPD{},
		MpdMetadata: NewMpdMetadata(),
	}
}

func ParseMpd(mpdBody []byte) (mpd *godashHttp.MPD, err error) {
	err = xml.Unmarshal(mpdBody, &mpd)
	if err != nil {
		return nil, err
	}
	return mpd, nil
}
