package main

import (
	"log"
	"net/http"

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

	http.Handle("/rpc", s)

	log.Println("MCP server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
