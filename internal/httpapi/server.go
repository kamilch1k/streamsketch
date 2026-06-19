package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kamilch1k/streamsketch/internal/sketch"
)

type Server struct {
	analyzer *sketch.Analyzer
}

type ingestRequest struct {
	Events []sketch.Event `json:"events"`
}

type analyzeRequest struct {
	Config sketch.Config  `json:"config,omitempty"`
	Events []sketch.Event `json:"events"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(config sketch.Config) (http.Handler, error) {
	analyzer, err := sketch.NewAnalyzer(config)
	if err != nil {
		return nil, err
	}

	server := &Server{analyzer: analyzer}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /api/events", server.observe)
	mux.HandleFunc("POST /api/ingest", server.ingest)
	mux.HandleFunc("POST /api/analyze", server.analyze)
	mux.HandleFunc("GET /api/streams", server.streams)
	mux.HandleFunc("GET /api/streams/{stream}/summary", server.summary)
	return mux, nil
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "service": "streamsketch"})
}

func (s *Server) observe(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	request.Body = http.MaxBytesReader(writer, request.Body, 2<<20)

	var event sketch.Event
	if err := decodeJSON(request, &event); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err := s.analyzer.Observe(event); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	summary, _ := s.analyzer.Summary(event.Stream, topK(request))
	writeJSON(writer, http.StatusAccepted, summary)
}

func (s *Server) ingest(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	request.Body = http.MaxBytesReader(writer, request.Body, 4<<20)

	var payload ingestRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if len(payload.Events) == 0 {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "events must not be empty"})
		return
	}
	if err := s.analyzer.ObserveAll(payload.Events); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, s.analyzer.Summaries(topK(request)))
}

func (s *Server) analyze(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	request.Body = http.MaxBytesReader(writer, request.Body, 4<<20)

	var payload analyzeRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if len(payload.Events) == 0 {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "events must not be empty"})
		return
	}

	summaries, err := sketch.Analyze(payload.Events, payload.Config)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, summaries)
}

func (s *Server) streams(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, s.analyzer.Summaries(topK(request)))
}

func (s *Server) summary(writer http.ResponseWriter, request *http.Request) {
	stream := strings.TrimSpace(request.PathValue("stream"))
	summary, ok := s.analyzer.Summary(stream, topK(request))
	if !ok {
		writeJSON(writer, http.StatusNotFound, errorResponse{Error: "stream not found"})
		return
	}
	writeJSON(writer, http.StatusOK, summary)
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			return errors.New("request body too large")
		}
		return err
	}
	return nil
}

func topK(request *http.Request) int {
	value := strings.TrimSpace(request.URL.Query().Get("k"))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
