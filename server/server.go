package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/itzmaniss/sysqueue/job"
	"github.com/itzmaniss/sysqueue/queue"
)

type Server struct {
	queue  *queue.Queue
	server *http.Server
}

func NewServer(q *queue.Queue) *Server {
	return &Server{
		queue: q,
	}
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("\n HTTP server Stopped")
		wg.Done()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /jobs", s.GetJobs)
	mux.HandleFunc("POST /jobs", s.PostJobs)
	mux.HandleFunc("GET /jobs/{id}", s.GetJob)
	mux.HandleFunc("DELETE /jobs/{id}", s.DeleteJob)

	s.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go s.server.ListenAndServe()

	<-ctx.Done()

	s.server.Shutdown(context.Background())

}

func (s *Server) GetJobs(w http.ResponseWriter, r *http.Request) {
	s.queue.Lock.Lock()
	jobs := make([]job.Job, 0, len(s.queue.Jobs))
	for _, j := range s.queue.Jobs {
		jobs = append(jobs, j)
	}
	s.queue.Lock.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) GetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.queue.Lock.Lock()
	j, ok := s.queue.Jobs[id]
	s.queue.Lock.Unlock()
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func (s *Server) PostJobs(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	j := s.queue.Enqueue(body.Command)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(j)
}

func (s *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.queue.Lock.Lock()
	_, ok := s.queue.Jobs[id]
	if !ok {
		s.queue.Lock.Unlock()
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	delete(s.queue.Jobs, id)
	s.queue.Lock.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
