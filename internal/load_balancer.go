package internal

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Server struct {
	URL               *url.URL   // URL of the backend server.
	ActiveConnections int        // Count of active connections
	Mutex             sync.Mutex // A mutex for safe concurrency
	Healthy           bool
}

type ServerConfig struct {
	HealthCheckInterval string
	Servers             []string
	ListenPort          string
}

func (s *ServerConfig) GetServers() []*Server {
	var servers []*Server
	for _, serverUrl := range s.Servers {
		u, _ := url.Parse(serverUrl)
		servers = append(servers, &Server{URL: u})
	}
	return servers
}

func (s *ServerConfig) GetServer(loadBalancingStrategy string, servers []*Server) (*Server, error) {
	switch loadBalancingStrategy {
	case "ROUND_ROBIN":
		return roundRobin(servers), nil
	case "LEAST_ACTIVE":
		return nextServerLeastActive(servers), nil
	}

	return nil, nil
}

func roundRobin(servers []*Server) *Server {
	for _, s := range servers {
		if s.Healthy && s.ActiveConnections == 0 {
			return s
		}
	}
	return nil
}

func nextServerLeastActive(servers []*Server) *Server {
	leastActiveConnections := -1
	leastActiveServer := servers[0]
	for _, server := range servers {
		server.Mutex.Lock()
		if (server.ActiveConnections < leastActiveConnections || leastActiveConnections == -1) && server.Healthy {
			leastActiveConnections = server.ActiveConnections
			leastActiveServer = server
		}
		server.Mutex.Unlock()
	}

	return leastActiveServer
}

func (s *Server) Proxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}
