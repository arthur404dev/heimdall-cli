package shell

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// IPCClient handles IPC communication with the shell daemon
type IPCClient struct {
	conn net.Conn
	port int
}

// NewIPCClient creates a new IPC client
func NewIPCClient(port int) (*IPCClient, error) {
	if port == 0 {
		port = 9999 // Default port
	}

	// Connect to daemon
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}

	return &IPCClient{
		conn: conn,
		port: port,
	}, nil
}

// SendMessage sends a message to the daemon and returns the response
func (c *IPCClient) SendMessage(message string) (string, error) {
	// Send message
	if _, err := fmt.Fprintf(c.conn, "%s\n", message); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Read response
	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// Close closes the IPC connection
func (c *IPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IPCServer handles incoming IPC connections
type IPCServer struct {
	port     int
	listener net.Listener
	handler  func(string) string
}

// NewIPCServer creates a new IPC server
func NewIPCServer(port int, handler func(string) string) (*IPCServer, error) {
	if port == 0 {
		port = 9999 // Default port
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start IPC server: %w", err)
	}

	return &IPCServer{
		port:     port,
		listener: listener,
		handler:  handler,
	}, nil
}

// Start starts the IPC server
func (s *IPCServer) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if listener was closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			continue
		}

		// Handle connection in goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single IPC connection
func (s *IPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// Read message
		message, err := reader.ReadString('\n')
		if err != nil {
			return // Connection closed
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		// Process message
		response := s.handler(message)

		// Send response
		fmt.Fprintf(conn, "%s\n", response)
	}
}

// Stop stops the IPC server
func (s *IPCServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
