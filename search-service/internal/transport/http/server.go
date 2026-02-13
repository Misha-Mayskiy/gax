package http

import (
	"encoding/json"
	"main/internal/domain"
	"main/internal/service"

	"net/http"
	"strconv"

	"github.com/rs/zerolog"
)

type Server struct {
	svc *service.SearchService
	log zerolog.Logger
}

func NewServer(svc *service.SearchService, log zerolog.Logger) *Server {
	return &Server{svc: svc, log: log}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", s.handleSearch)
	return mux
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	t := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(defaultIfEmpty(r.URL.Query().Get("limit"), "20"))
	offset, _ := strconv.Atoi(defaultIfEmpty(r.URL.Query().Get("offset"), "0"))
	highlight := r.URL.Query().Get("highlight") == "true"

	if q == "" && t == "" {
		http.Error(w, "q or type required", http.StatusBadRequest)
		return
	}

	if q == "" && t == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "q or type required",
		})
		return
	}

	results, err := s.svc.Search(r.Context(), domain.SearchQuery{
		Q:         q,
		Type:      t,
		Limit:     limit,
		Offset:    offset,
		Highlight: highlight,
	})
	if err != nil {
		s.log.Error().Err(err).Msg("search error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "search failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"items":  results,
		"limit":  limit,
		"offset": offset,
	})
}

func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func (s *Server) Listen(addr string) error {
	return http.ListenAndServe(addr, s.routes())
}
