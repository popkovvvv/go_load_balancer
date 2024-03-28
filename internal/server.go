package internal

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func StartServer(loadBalanceStrategy string, serverUrls []string, interval string, serverPort string) {
	serverConfig := ServerConfig{HealthCheckInterval: interval, Servers: serverUrls, ListenPort: serverPort}
	healthCheckInterval, err := time.ParseDuration(serverConfig.HealthCheckInterval)

	if err != nil {
		log.Fatalf("Invalid health check interval: %s", err.Error())
	}

	var servers = serverConfig.GetServers()
	for _, server := range servers {
		go func(s *Server) {
			for range time.Tick(healthCheckInterval) {
				res, err := http.Get(s.URL.String())
				s.Mutex.Lock()
				if err != nil || res.StatusCode >= 500 {
					s.Healthy = false
				} else {
					s.Healthy = true
				}
				s.Mutex.Unlock()
				fmt.Printf("Server url '%s' is '%s'\n", s.URL, strconv.FormatBool(s.Healthy))
			}
		}(server)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server, err := serverConfig.GetServer(loadBalanceStrategy, servers)
		if err != nil || server == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Chosen server url '%s'\n", server.URL)

		server.Mutex.Lock()
		server.ActiveConnections++
		server.Mutex.Unlock()
		server.Proxy().ServeHTTP(w, r)
		server.Mutex.Lock()
		server.ActiveConnections--
		server.Mutex.Unlock()
	})

	log.Println("Starting server on port", serverConfig.ListenPort)
	err = http.ListenAndServe(serverConfig.ListenPort, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
