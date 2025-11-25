package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/onichange/pos-system/pkg/logger"
)

// Server wraps gRPC server with health checks
type Server struct {
	server   *grpc.Server
	health   *health.Server
	logger   *logger.Logger
	port     string
	listener net.Listener
}

// NewServer creates a new gRPC server
func NewServer(port string, log *logger.Logger) (*Server, error) {
	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	}

	server := grpc.NewServer(opts...)
	healthServer := health.NewServer()

	// Register health service
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	reflection.Register(server)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	return &Server{
		server:   server,
		health:   healthServer,
		logger:   log,
		port:     port,
		listener: listener,
	}, nil
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// Start starts the gRPC server
func (s *Server) Start() error {
	s.logger.Infof("gRPC server starting on port %s", s.port)
	return s.server.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.server.GracefulStop()
	s.logger.Info("gRPC server stopped")
}

// SetServingStatus sets the health check status
func (s *Server) SetServingStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.health.SetServingStatus(service, status)
}
