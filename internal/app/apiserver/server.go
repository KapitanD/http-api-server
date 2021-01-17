package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/KapitanD/http-api-server/internal/app/store"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

const (
	sessionName        = "note-session"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
	errIncorrectRequest         = errors.New("incorrect request")
)

type ctxKey int8

type server struct {
	router       *mux.Router
	logger       *logrus.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       mux.NewRouter(),
		logger:       logrus.New(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/sessions", s.handleSessionCreate()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc("/whoami", s.handleWhoami()).Methods("GET")

	notes := s.router.PathPrefix("/notes").Subrouter()
	notes.Use(s.authenticateUser)
	notes.HandleFunc("/", s.handleNotesCreate()).Methods("POST")
	notes.HandleFunc("/{id:[0-9]+}", s.handleNotesUpdate()).Methods("PATCH")
	notes.HandleFunc("/{id:[0-9]+}", s.handleNotesDelete()).Methods("DELETE")
	notes.HandleFunc("/", s.handleNotesGetAll()).Methods("GET")
	notes.HandleFunc("/{id:[0-9]+}", s.handleNotesGet()).Methods("GET")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}

}

func (s *server) handleSessionCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleNotesCreate() http.HandlerFunc {
	type request struct {
		Header string `json:"header"`
		Body   string `json:"body"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(ctxKeyUser).(*model.User)

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		n := &model.Note{
			Header:    req.Header,
			Body:      req.Body,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.store.Notes().Create(n, u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, n)
	}
}

func (s *server) handleNotesUpdate() http.HandlerFunc {

	type request struct {
		Header string `json:"header"`
		Body   string `json:"body"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmp, ok := mux.Vars(r)["id"]
		if !ok {
			s.error(w, r, http.StatusBadRequest, errIncorrectRequest)
			return
		}
		id, err := strconv.Atoi(tmp)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		u := r.Context().Value(ctxKeyUser).(*model.User)

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		n, err := s.store.Notes().FindByID(id)
		if err != nil {
			s.error(w, r, http.StatusNotFound, store.ErrRecordNotFound)
			return
		}
		if n.AuthorID != u.ID {
			s.error(w, r, http.StatusNotFound, store.ErrRecordNotFound)
			return
		}

		un := &model.Note{
			Header:    req.Header,
			Body:      req.Body,
			UpdatedAt: time.Now(),
		}

		if err := s.store.Notes().Update(id, un); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleNotesDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, ok := mux.Vars(r)["id"]
		if !ok {
			s.error(w, r, http.StatusBadRequest, errIncorrectRequest)
			return
		}
		id, err := strconv.Atoi(tmp)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		u := r.Context().Value(ctxKeyUser).(*model.User)

		n, err := s.store.Notes().FindByID(id)
		if err != nil {
			s.respond(w, r, http.StatusOK, nil)
			return
		}

		if n.AuthorID != u.ID {
			s.respond(w, r, http.StatusOK, nil)
			return
		}

		if err := s.store.Notes().Delete(id); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleNotesGetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(ctxKeyUser).(*model.User)

		nl, err := s.store.Notes().FindByUser(u)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
		s.respond(w, r, http.StatusOK, nl)
	}
}

func (s *server) handleNotesGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, ok := mux.Vars(r)["id"]
		if !ok {
			s.error(w, r, http.StatusBadRequest, errIncorrectRequest)
			return
		}
		id, err := strconv.Atoi(tmp)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		u := r.Context().Value(ctxKeyUser).(*model.User)

		n, err := s.store.Notes().FindByID(id)
		if err != nil || n.AuthorID != u.ID {
			s.error(w, r, http.StatusNotFound, store.ErrRecordNotFound)
		}
		s.respond(w, r, http.StatusOK, n)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
