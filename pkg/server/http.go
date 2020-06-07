package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	Now         time.Time
	TraceID     string
	Method      string
	RequestPath string
	StatusCode  int
}

type server struct {
	router *httprouter.Router
	logger *zap.SugaredLogger
}

// NewServer returns a HTTP server for accessing the account service.
// The server implements http.Handler.
func NewServer(logger *zap.SugaredLogger, e Endpoints) *server {
	router := httprouter.New()
	router.HandleOPTIONS = true

	s := server{
		router: router,
		logger: logger,
	}

	// initialise server router
	for _, e := range e.Endpoints() {
		s.handle(e.Method, e.Path, e.Handler)
	}

	return &s
}

// func (s *server) routes() {
// 	s.handle("GET", "/accounts/:id", s.handleGetAccount())
// }

func (s *server) handle(method, path string, handler http.Handler) {

	// Create the function to execute for each request
	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Set the context with the required values to process the request
		v := Values{
			Now:         time.Now(),
			TraceID:     uuid.New().String(),
			Method:      method,
			RequestPath: path,
		}

		ctx = context.WithValue(ctx, KeyValues, &v)

		s.log(handler).ServeHTTP(w, r.WithContext(ctx))
	}

	s.router.HandlerFunc(method, path, h)
}

// ServeHTTP implements http.Handler
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

//
// helpers
//

func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	if v, ok := r.Context().Value(KeyValues).(*Values); ok {
		v.StatusCode = status
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			// TODO!
			panic(err)
		}
	}
}

//
// handlers
//