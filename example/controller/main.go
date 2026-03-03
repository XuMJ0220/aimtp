package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type createResponse struct {
	Message string `json:"message"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("health check from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/dags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()
		log.Printf("create dag request received, size=%d", len(body))

		resp := createResponse{
			Message: fmt.Sprintf("received %d bytes", len(body)),
		}
		data, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	server := &http.Server{
		Addr:              ":8088",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("example controller listening on %s", server.Addr)
	_ = server.ListenAndServe()
}
