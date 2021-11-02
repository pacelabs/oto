package otohttp

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Server handles oto requests.
type Server struct {
	// Basepath is the path prefix to match.
	// Default: /oto/
	Basepath string

	routes map[string]http.Handler
	// NotFound is the http.Handler to use when a resource is
	// not found.
	NotFound http.Handler
	// OnErr is called when there is an error.
	OnErr func(w http.ResponseWriter, r *http.Request, err error)
}

// NewServer makes a new Server.
func NewServer() *Server {
	return &Server{
		Basepath: "/oto/",
		routes:   make(map[string]http.Handler),
		OnErr: func(w http.ResponseWriter, r *http.Request, err error) {
			errObj := struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			}
			if err := Encode(w, r, http.StatusInternalServerError, errObj); err != nil {
				log.Printf("failed to encode error: %s\n", err)
			}
		},
		NotFound: http.NotFoundHandler(),
	}
}

// Register adds a handler for the specified service method.
func (s *Server) Register(service, method string, h http.HandlerFunc) {
	s.routes[fmt.Sprintf("%s%s.%s", s.Basepath, service, method)] = h
}

// ServeHTTP serves the request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.NotFound.ServeHTTP(w, r)
		return
	}
	h, ok := s.routes[r.URL.Path]
	if !ok {
		s.NotFound.ServeHTTP(w, r)
		return
	}
	h.ServeHTTP(w, r)
}

// Encode writes the response.
func Encode(w http.ResponseWriter, r *http.Request, status int, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "encode json")
	}
	var out io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzip.NewWriter(w)
		out = gzw
		defer gzw.Close()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := out.Write(b); err != nil {
		return err
	}
	return nil
}

// Decode unmarshals the object in the request into v.
func Decode(r *http.Request, v interface{}) error {
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		return fmt.Errorf("Decode: read body: %w", err)
	}
	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return fmt.Errorf("Decode: json.Unmarshal: %w", err)
	}
	return nil
}
