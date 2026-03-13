package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"sync/atomic"

	"github.com/itzmaniss/sysqueue/job"
	"github.com/itzmaniss/sysqueue/metrics"
	"github.com/itzmaniss/sysqueue/queue"
	"golang.org/x/time/rate"
)

type Server struct {
	queue   *queue.Queue
	metrics *metrics.Metrics
	server  *http.Server
	limiter *rate.Limiter
}

func NewServer(q *queue.Queue, m *metrics.Metrics) *Server {
	return &Server{
		queue:   q,
		metrics: m,
		limiter: rate.NewLimiter(rate.Limit(10), 20),
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
	mux.HandleFunc("GET /health", s.HealthCheck)
	mux.HandleFunc("GET /metrics", s.GetMetrics)

	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	s.server = &http.Server{
		Addr:    ":8080",
		Handler: s.rateLimitMiddleware(mux),
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

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"jobs_processed": atomic.LoadInt64(&s.metrics.JobsProcessed),
		"jobs_failed":    atomic.LoadInt64(&s.metrics.JobsFailed),
		"jobs_pending":   int64(len(s.queue.Jobs)),
	})
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
