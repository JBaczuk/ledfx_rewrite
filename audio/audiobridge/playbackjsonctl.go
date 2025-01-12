package audiobridge

import (
	"encoding/json"
	"fmt"
)

type PlaybackAction int

const (
	PlaybackActionStop PlaybackAction = iota
	PlaybackActionEnableVerbose
	PlayBackActionDisableVerbose
)

type PlaybackCTLJSON struct {
	Action PlaybackAction `json:"action"`
}

func (capctl PlaybackCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&capctl)
}

func (j *JsonCTL) Playback(jsonData []byte) (err error) {
	conf := PlaybackCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch conf.Action {
	case PlaybackActionStop:
		return j.w.br.Controller().Local().QuitPlayback()
	case PlaybackActionEnableVerbose:
		return j.w.br.Controller().Local().SetVerbose(true)
	case PlayBackActionDisableVerbose:
		return j.w.br.Controller().Local().SetVerbose(false)
	}

	return fmt.Errorf("unknown action '%d'", conf.Action)
}
