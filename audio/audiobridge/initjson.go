package audiobridge

import (
	"encoding/json"
	"fmt"
	"ledfx/config"
)

type Wrapper interface {
	// AsJSON returns the JSON representation of a configuration struct
	AsJSON() ([]byte, error)
}

// AirPlayInputJSON configures an AirPlay input (server)
type AirPlayInputJSON struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Verbose bool   `json:"verbose,omitempty"`
}

func (a AirPlayInputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&a)
}

// AirPlayOutputJSON configures an AirPlay output
type AirPlayOutputJSON struct {
	SearchKey  string            `json:"search_key"`
	SearchType AirPlaySearchType `json:"search_type"`
	Verbose    bool              `json:"verbose,omitempty"`
}

func (a AirPlayOutputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&a)
}

// LocalInputJSON configures a local input (capture)
type LocalInputJSON struct {
	AudioDevice *config.AudioDevice `json:"audio_device,omitempty"`
	Verbose     bool                `json:"verbose,omitempty"`
}

func (l LocalInputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&l)
}

// LocalOutputJSON configures a local output (playback)
type LocalOutputJSON struct {
	Verbose bool `json:"verbose,omitempty"`
}

func (l LocalOutputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&l)
}

type YouTubeInputJSON struct {
	Verbose bool `json:"verbose,omitempty"`
}

func (y YouTubeInputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&y)
}

// JSONWrapper returns an interpreter for JSON-based configuration
// parameters.
func (br *Bridge) JSONWrapper() *BridgeJSONWrapper {
	if br.jsonWrapper == nil {
		br.jsonWrapper = &BridgeJSONWrapper{
			br: br,
		}
	}
	return br.jsonWrapper
}

// StartAirPlayInput takes a marshalled AirPlayInputJSON.
func (w *BridgeJSONWrapper) StartAirPlayInput(jsonData []byte) (err error) {
	conf := AirPlayInputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	if err := w.br.StartAirPlayInput(conf.Name, conf.Port, conf.Verbose); err != nil {
		return fmt.Errorf("error starting AirPlay Server: %w", err)
	}

	return nil
}

// AddAirPlayOutput takes a marshalled AirPlayOutputJSON
func (w *BridgeJSONWrapper) AddAirPlayOutput(jsonData []byte) (err error) {
	conf := AirPlayOutputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.AddAirPlayOutput(conf.SearchKey, conf.SearchType, conf.Verbose); err != nil {
		return fmt.Errorf("error adding AirPlay Client: %w", err)
	}
	return nil
}

// StartLocalInput takes a marshalled LocalInputJSON
func (w *BridgeJSONWrapper) StartLocalInput(jsonData []byte) (err error) {
	conf := LocalInputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.StartLocalInput(*conf.AudioDevice, conf.Verbose); err != nil {
		return fmt.Errorf("error starting local capture: %w", err)
	}
	return nil
}

// AddLocalOutput takes a marshalled LocalOutputJSON
func (w *BridgeJSONWrapper) AddLocalOutput(jsonData []byte) (err error) {
	conf := LocalOutputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.AddLocalOutput(conf.Verbose); err != nil {
		return fmt.Errorf("error starting local playback: %w", err)
	}
	return nil
}

// StartYouTubeInput takes a marshalled YouTubeInputJSON
func (w *BridgeJSONWrapper) StartYouTubeInput(jsonData []byte) (err error) {
	conf := YouTubeInputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.StartYoutubeInput(conf.Verbose); err != nil {
		return fmt.Errorf("error starting YouTubeSet input: %w", err)
	}
	return nil
}
