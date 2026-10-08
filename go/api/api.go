// Package api expone la cola por HTTP.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/edgar981/jukebox-queue/go/queue"
)

// NewHandler arma las rutas sobre un Store.
func NewHandler(s *queue.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /venues/{venue}/queue", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SongID  string `json:"songId"`
			Title   string `json:"title"`
			TableID string `json:"tableId"`
		}
		if !decode(w, r, &body) {
			return
		}
		if blank(body.SongID, body.Title, body.TableID) {
			writeError(w, http.StatusBadRequest, "songId, title and tableId are required")
			return
		}
		it, err := s.Add(r.PathValue("venue"), body.SongID, body.Title, body.TableID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, it)
	})

	mux.HandleFunc("GET /venues/{venue}/queue", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List(r.PathValue("venue")))
	})

	mux.HandleFunc("POST /venues/{venue}/queue/{itemId}/votes", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			TableID string `json:"tableId"`
		}
		if !decode(w, r, &body) {
			return
		}
		if blank(body.TableID) {
			writeError(w, http.StatusBadRequest, "tableId is required")
			return
		}
		it, err := s.Vote(r.PathValue("venue"), r.PathValue("itemId"), body.TableID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, it)
	})

	mux.HandleFunc("POST /venues/{venue}/queue/advance", func(w http.ResponseWriter, r *http.Request) {
		it, err := s.Advance(r.PathValue("venue"))
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, it)
	})

	return mux
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func blank(vals ...string) bool {
	for _, v := range vals {
		if strings.TrimSpace(v) == "" {
			return true
		}
	}
	return false
}

// writeDomainError traduce los errores del dominio a códigos HTTP.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, queue.ErrDuplicateSong), errors.Is(err, queue.ErrAlreadyVoted):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, queue.ErrTableLimit):
		writeError(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, queue.ErrNotFound), errors.Is(err, queue.ErrEmpty):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
