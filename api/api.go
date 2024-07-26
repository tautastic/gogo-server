package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type InitRequest struct {
	BoardSize string   `json:"board-size,omitempty"`
	Komi      string   `json:"komi,omitempty"`
	Handicaps []string `json:"handicaps,omitempty"`
}

type GenmoveRequest struct {
	Color string `json:"color,omitempty"`
}

type GenmoveResponse struct {
	Move string `json:"move,omitempty"`
}

type PlayMoveRequest struct {
	Color     string `json:"color,omitempty"`
	MoveToPos string `json:"move_to_pos,omitempty"`
}

type Server struct {
	authToken  string
	inputChan  chan string
	outputChan chan string
	errorChan  chan error
	readyChan  chan bool
	isReady    bool
}

func NewServer(authToken string, inputChan chan string, outputChan chan string, errorChan chan error, readyChan chan bool) *Server {
	return &Server{
		authToken:  authToken,
		inputChan:  inputChan,
		outputChan: outputChan,
		errorChan:  errorChan,
		readyChan:  readyChan,
		isReady:    false,
	}
}

func (s *Server) checkAuthToken(r *http.Request) bool {
	authorization := r.Header.Get("Authorization")
	return authorization == s.authToken
}

func (s *Server) emptyOutputChannel() {
	for len(s.outputChan) > 0 {
		<-s.outputChan
	}
}

func (s *Server) checkReadyHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuthToken(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if len(s.readyChan) == cap(s.readyChan) || s.isReady {
		s.isReady = true
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
}

func (s *Server) initHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuthToken(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var reqBody InitRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	s.inputChan <- fmt.Sprintf("clear_board")
	s.inputChan <- fmt.Sprintf("boardsize %s", reqBody.BoardSize)
	s.inputChan <- fmt.Sprintf("komi %s", reqBody.Komi)
	switch len(reqBody.Handicaps) {
	case 0:
		break
	case 1:
		s.inputChan <- fmt.Sprint("set_position white ", reqBody.Handicaps[0])
	default:
		s.inputChan <- fmt.Sprint("set_position white ", strings.Join(reqBody.Handicaps, " white "))
	}

	select {
	case err, ok := <-s.errorChan:
		if !ok {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		log.Println("Process error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	default:
		break
	}
}

func (s *Server) playMoveHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuthToken(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var reqBody PlayMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	s.emptyOutputChannel()

	s.inputChan <- fmt.Sprintf("play %s %s", reqBody.Color, reqBody.MoveToPos)

	select {
	case err, ok := <-s.errorChan:
		if !ok {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		log.Println("Process error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	default:
		break
	}
}

func (s *Server) genMoveHandler(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuthToken(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var reqBody GenmoveRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	s.emptyOutputChannel()

	s.inputChan <- fmt.Sprintf("genmove %s", reqBody.Color)

	select {
	case output, ok := <-s.outputChan:
		if !ok {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenmoveResponse{Move: output})
	case err, ok := <-s.errorChan:
		if !ok {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		log.Println("Process error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) Start() {
	http.HandleFunc("/check-ready", s.checkReadyHandler)
	http.HandleFunc("/init", s.initHandler)
	http.HandleFunc("/gen-move", s.genMoveHandler)
	http.HandleFunc("/play-move", s.playMoveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
