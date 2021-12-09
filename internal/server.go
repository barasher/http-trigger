package internal

import (
	"fmt"
	"net/http"
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
	port     uint
	commands map[string]string
}

func LoadConfiguration(f string) (ServerConf, error) {
	sc := ServerConf{
		commands: make(map[string]string),
	}
	return sc, nil
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
	w.WriteHeader(http.StatusOK)
	l.Info().Int64("duration", duration).Str("state", state).Msg("Invocation complete")
}

func NewServer(conf ServerConf) (*Server, error) {
	s := Server{
		port: conf.port,
	}
	s.router = mux.NewRouter()

	for cmdKey, cmd := range conf.commands {
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
		h := promhttp.InstrumentHandlerDuration(histoDur, execHandler{cmdKey: cK, cmd: c})
		s.router.HandleFunc(fmt.Sprintf("/exec/{%v}", cK), h).Methods(http.MethodPost)
	}

	s.router.Handle("/metrics", promhttp.Handler())

	return &s, nil
}
