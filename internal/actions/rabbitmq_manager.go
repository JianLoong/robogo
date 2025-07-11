package actions

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/JianLoong/robogo/internal/util"
)

// RabbitMQManager manages RabbitMQ connections with proper lifecycle
type RabbitMQManager struct {
	connections map[string]*amqp.Connection
	mutex       sync.RWMutex
}

// NewRabbitMQManager creates a new RabbitMQ connection manager
func NewRabbitMQManager() *RabbitMQManager {
	return &RabbitMQManager{
		connections: make(map[string]*amqp.Connection),
	}
}

// Connect establishes a new RabbitMQ connection
func (rm *RabbitMQManager) Connect(name, connectionString string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Close existing connection if it exists
	if existing, exists := rm.connections[name]; exists {
		if !existing.IsClosed() {
			existing.Close()
		}
	}

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return util.NewMessagingError("failed to connect to RabbitMQ", err, "rabbitmq").
			WithDetails(map[string]interface{}{
				"connection_string": connectionString,
				"connection_name":   name,
			})
	}

	rm.connections[name] = conn
	return nil
}

// GetConnection retrieves a connection by name
func (rm *RabbitMQManager) GetConnection(name string) (*amqp.Connection, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	conn, exists := rm.connections[name]
	if !exists {
		return nil, false
	}

	// Check if connection is still alive
	if conn.IsClosed() {
		// Clean up dead connection
		go func() {
			rm.mutex.Lock()
			delete(rm.connections, name)
			rm.mutex.Unlock()
		}()
		return nil, false
	}

	return conn, true
}

// CloseConnection closes a specific connection
func (rm *RabbitMQManager) CloseConnection(name string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	conn, exists := rm.connections[name]
	if !exists {
		return util.NewValidationError("rabbitmq connection not found", map[string]interface{}{
			"connection_name":        name,
			"available_connections": rm.getConnectionNames(),
		})
	}

	var err error
	if !conn.IsClosed() {
		err = conn.Close()
		if err != nil {
			err = util.NewMessagingError("failed to close RabbitMQ connection", err, "rabbitmq").
				WithDetails(map[string]interface{}{
					"connection_name": name,
				})
		}
	}

	delete(rm.connections, name)
	return err
}

// CloseAll closes all connections
func (rm *RabbitMQManager) CloseAll() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	var lastErr error
	for name, conn := range rm.connections {
		if conn != nil && !conn.IsClosed() {
			if err := conn.Close(); err != nil {
				lastErr = util.NewMessagingError("failed to close RabbitMQ connection during cleanup", err, "rabbitmq").
					WithDetails(map[string]interface{}{
						"connection_name": name,
					})
			}
		}
	}

	// Clear all connections
	rm.connections = make(map[string]*amqp.Connection)
	return lastErr
}

// ListConnections returns the names of all active connections
func (rm *RabbitMQManager) ListConnections() []string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.getConnectionNames()
}

// IsConnected checks if a named connection exists and is active
func (rm *RabbitMQManager) IsConnected(name string) bool {
	_, exists := rm.GetConnection(name)
	return exists
}

// getConnectionNames returns connection names (must be called with lock held)
func (rm *RabbitMQManager) getConnectionNames() []string {
	names := make([]string, 0, len(rm.connections))
	for name := range rm.connections {
		names = append(names, name)
	}
	return names
}