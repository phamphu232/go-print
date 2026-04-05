package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	server *http.Server
	mu     sync.Mutex
)

func startServer(host string, port int) {
	mu.Lock()
	defer mu.Unlock()

	// Prevent multiple server instances from running simultaneously
	if server != nil {
		return
	}

	// Determine the primary target address based on input parameters
	targetAddr := fmt.Sprintf("%s:%d", host, port)
	if host == "" || port == 0 {
		targetAddr = fmt.Sprintf("%s:%d", setting.Host, setting.Port)
	}

	// Attempt to open a TCP listener on the target address
	ln, err := net.Listen("tcp", targetAddr)

	if err != nil {
		log.Printf("Failed to bind to %s, falling back to default: %v", targetAddr, err)

		// Fallback: Attempt to bind to the default configuration
		targetAddr = fmt.Sprintf("%s:%d", setting.Host, setting.Port)
		ln, err = net.Listen("tcp", targetAddr)
		if err != nil {
			log.Printf("Critical: Could not start server even with default config: %v", err)
			return
		}
	} else if host != "" && port != 0 {
		// Update and persist configuration only if the new parameters were successful
		setting.Host = host
		setting.Port = port
		saveSetting()
	}

	mux := setupRoutes()

	server = &http.Server{
		Handler: logRequestMiddleware(corsMiddleware(mux)),
	}

	go func() {
		log.Printf("Server started at: http://%s", targetAddr)
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Println("Server error:", err)
		}
	}()
}

func restartServer(host string, port int) {
	stopServer()
	time.Sleep(500 * time.Millisecond)
	startServer(host, port)
}

func stopServer() {
	mu.Lock()
	defer mu.Unlock()

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		server = nil
	}

	log.Println("Server stopped")
}
