package main

import (
	"log"
	"net/http"
	"time" // Import time package

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/sire-run/sire/internal/mcp/api"
)

func main() {
	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	if err := s.RegisterService(api.NewService(), "mcp"); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/rpc", s)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,   // Example timeout
		WriteTimeout:      10 * time.Second,  // Example timeout
		IdleTimeout:       120 * time.Second, // Example timeout
	}

	log.Println("MCP server listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
