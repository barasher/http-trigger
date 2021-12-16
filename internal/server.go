package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	stateFailure = "ko"
	stateSuccess = "ok"

	headerExitCode = "http-trigger-exit-code"
	headerDuration = "http-trigger-duration"
)

type ServerConf struct {
	Port     uint              `json:"port"`
	Commands map[string]string `json:"commands"`
}

func LoadConfiguration(f string) (ServerConf, error) {
	conf := ServerConf{
		Commands: make(map[string]string),
	}
	confReader, err := os.Open(f)
	if err != nil {
		return conf, fmt.Errorf("error while opening configuration file %v: %w", f, err)
	}
	defer confReader.Close()
	err = json.NewDecoder(confReader).Decode(&conf)
	if err != nil {
		return conf, fmt.Errorf("error while unmarshaling configuration file %v: %w", f, err)
	}
	return conf, nil
}

type Server struct {
	port   uint
	router *mux.Router
}

type execHandler struct {
	cmdKey string
	cmd    string
}

func (h execHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	l := log.With().Str("cmdId", h.cmdKey).Logger()
	l.Debug().Msg("Invocation...")
	start := time.Now()
	state := stateSuccess
	exitCode := "0"

	cmd := exec.Command(h.cmd)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		state = stateFailure
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = strconv.Itoa(exitErr.ExitCode())
		} else {
			exitCode = "-1"
		}
	}

	duration := time.Now().Sub(start).Milliseconds()
	w.Header().Add(headerDuration, strconv.Itoa(int(duration)))
	w.Header().Add(headerExitCode, exitCode)
	l.Info().Int64("duration", duration).Str("state", state).Msg("Invocation complete")
}

func NewServer(conf ServerConf) (*Server, error) {
	s := Server{
		port: conf.Port,
	}
	s.router = mux.NewRouter()

	for cmdKey, cmd := range conf.Commands {
		cK := cmdKey
		c := cmd
		histoDur := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("http_trigger_exec_%v_request_duration_seconds", cK),
				Help:    fmt.Sprintf("Histogram concerning exec request durations (seconds) for command %v", cK),
				Buckets: []float64{.0025, .005, .01, .025, .05, .1},
			},
			[]string{},
		)
		prometheus.Unregister(histoDur)
		prometheus.MustRegister(histoDur)
		h := promhttp.InstrumentHandlerDuration(histoDur, execHandler{cmdKey: cK, cmd: c})
		s.router.HandleFunc(fmt.Sprintf("/exec/%v", cK), h).Methods(http.MethodPost)
	}

	s.router.Handle("/metrics", promhttp.Handler())

	return &s, nil
}

func (s *Server) Run() error {
	p := s.port
	if p == 0 {
		p = 8080
		log.Info().Msgf("Default port will be used (%v)", p)
	}
	addr := fmt.Sprintf("0.0.0.0:%v", p)
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}
	log.Info().Msg("Server running...")
	return srv.ListenAndServe()
}
