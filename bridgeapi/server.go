package bridgeapi

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	"ledfx/bridgeapi/statpoll"
	log "ledfx/logger"
	"net/http"
	"sync"
)

const (
	ArtworkURLPath string = "/api/bridge/artwork"
)

type Server struct {
	mux        *http.ServeMux
	br         *audiobridge.Bridge
	statPoller *statpoll.StatPoller
	upgrader   *websocket.Upgrader
}

func NewServer(callback func(buf audio.Buffer), mux *http.ServeMux) (err error) {
	s := &Server{
		mux: mux,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			WriteBufferPool: &sync.Pool{},
		},
	}

	if s.br, err = audiobridge.NewBridge(callback); err != nil {
		return fmt.Errorf("error initializing new bridge: %w", err)
	}

	s.statPoller = statpoll.New(s.br)

	// Input setter handlers
	s.mux.HandleFunc("/api/bridge/set/input/airplay", s.handleSetInputAirPlay)
	s.mux.HandleFunc("/api/bridge/set/input/youtube", s.handleSetInputYouTube)
	s.mux.HandleFunc("/api/bridge/set/input/capture", s.handleSetInputCapture)

	// Output adder handlers
	s.mux.HandleFunc("/api/bridge/add/output/airplay", s.handleAddOutputAirPlay)
	s.mux.HandleFunc("/api/bridge/add/output/local", s.handleAddOutputLocal)

	// Ctl handlers
	s.mux.HandleFunc("/api/bridge/ctl/youtube/set", s.handleCtlYouTube)
	s.mux.HandleFunc("/api/bridge/ctl/youtube/info", s.handleCtlYouTubeGetInfo)

	s.mux.HandleFunc("/api/bridge/ctl/airplay/set", s.handleCtlAirPlaySet)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/clients", s.handleCtlAirPlayGetClients)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/info", s.handleCtlAirPlayGetInfo)

	// StatPoller handler
	s.mux.HandleFunc("/api/bridge/statpoll/ws", s.handleStatPollInitWs)

	// Artwork handler
	s.mux.HandleFunc(ArtworkURLPath, s.handleArtwork)
	return nil
}

func (s *Server) handleStatPollInitWs(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Errorf("Error upgrading connection to websocket: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
	if err := s.statPoller.AddWebsocket(ws); err != nil {
		log.Logger.Errorf("Error adding websocket to statpoller: %v", err)
		_ = ws.Close()
	}
}

// ############## BEGIN AIRPLAY ##############
func (s *Server) handleSetInputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to AirPlay server....")
	if err := s.br.JSONWrapper().StartAirPlayInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting AirPlay input: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleAddOutputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Adding AirPlay audio output...")
	if err := s.br.JSONWrapper().AddAirPlayOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting AirPlay output: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleCtlAirPlaySet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("method '%s' is not allowed", r.Method)))
		return
	}

	log.Logger.Infoln("Got AirPlay SET CTL request...")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}

	if err := s.br.JSONWrapper().CTL().AirPlaySet(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error getting return JSON from AirPlay CTL: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleCtlAirPlayGetClients(w http.ResponseWriter, r *http.Request) {
	log.Logger.Infoln("Got AirPlay GET CTL request...")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("method '%s' is not allowed", r.Method)))
		return
	}

	clientBytes, err := s.br.JSONWrapper().CTL().AirPlayGetClients()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error getting clients: %v", err)))
		return
	}

	w.Write(clientBytes)

}
func (s *Server) handleCtlAirPlayGetInfo(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.WriteHeader(http.StatusServiceUnavailable)
}

// ############### END AIRPLAY ###############

// ############## BEGIN YOUTUBE ##############
func (s *Server) handleSetInputYouTube(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to YouTubeSet....")
	if err := s.br.JSONWrapper().StartYouTubeInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting YouTubeSet input: %v", err)
		w.Write(errToBytes(err))
		return
	}
}

func (s *Server) handleCtlYouTube(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	if err := s.br.JSONWrapper().CTL().YouTubeSet(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error running YouTubeSet CTL action: %v", err)
		w.Write(errToBytes(err))
		return
	}
}

func (s *Server) handleCtlYouTubeGetInfo(w http.ResponseWriter, r *http.Request) {
	ret, err := s.br.JSONWrapper().CTL().YouTubeGetInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error running YouTubeGet CTL action: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

// ############### END YOUTUBE ###############

// ############## BEGIN LOCAL ##############
func (s *Server) handleSetInputCapture(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to local capture...")
	if err := s.br.JSONWrapper().StartLocalInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting local capture input: %v", err)
		w.Write(errToBytes(err))
		return
	}
}
func (s *Server) handleAddOutputLocal(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Adding local audio output...")
	if err := s.br.JSONWrapper().AddLocalOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting local output: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ############### END LOCAL ###############

// ############## BEGIN MISC ##############
func (s *Server) handleArtwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "image/png")
	io.Copy(w, bytes.NewReader(s.br.Artwork()))
}

// ############### END MISC ###############

func errToBytes(err error) []byte {
	return []byte(err.Error() + "\n")
}
