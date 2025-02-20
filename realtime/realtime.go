package realtime

import (
	"errors"
	"maps"
	"slices"
	"sync"
	"time"
)

type Client interface {
	Sender()
	Reader()
	Disconnect()
}

type Connection struct {
	GetRobotSetModeStream      *StreamChannel
	GetWaypointsStream         *StreamChannel
	GetPeripheralControlStream *StreamChannel
	GetJoystickControlStream   *StreamChannel

	wsClients map[Client](time.Time)
	clientsMu sync.Mutex
}

type ConnectionManager struct {
	RobotConnections map[int]*Connection
}

func NewConnectionManager() ConnectionManager {
	return ConnectionManager{
		RobotConnections: map[int]*Connection{},
	}
}

func createNewConnection() *Connection {
	return &Connection{
		GetRobotSetModeStream:      NewStreamChannel(),
		GetWaypointsStream:         NewStreamChannel(),
		GetPeripheralControlStream: NewStreamChannel(),
		GetJoystickControlStream:   NewStreamChannel(),
		wsClients:                  map[Client]time.Time{},
	}
}

func (cm *ConnectionManager) GetConnectionForRobot(robotId int) (*Connection, error) {
	if connection, exists := cm.RobotConnections[robotId]; exists {
		return connection, nil
	}
	cm.RobotConnections[robotId] = createNewConnection()
	return cm.RobotConnections[robotId], nil
}

func (cm *ConnectionManager) GetClientsForRobot(robotId int) ([]Client, error) {
	if connection, exists := cm.RobotConnections[robotId]; exists {
		connection.clientsMu.Lock()
		clients := slices.Collect(maps.Keys(connection.wsClients))
		connection.clientsMu.Unlock()
		return clients, nil
	}
	return nil, errors.New("No client")
}

func (cm *ConnectionManager) RegisterClient(robotId int, client Client) error {
	if conn, exists := cm.RobotConnections[robotId]; exists {
		conn.wsClients[client] = time.Now()
		return nil
	}
	cm.RobotConnections[robotId] = createNewConnection()
	cm.RobotConnections[robotId].wsClients[client] = time.Now()
	return nil
}

func (c *Connection) RemoveClient(client Client) {
	if _, exists := c.wsClients[client]; exists {
		c.clientsMu.Lock()
		client.Disconnect() // Only closes the client resources
		delete(c.wsClients, client)
		c.clientsMu.Unlock()
	}
}
