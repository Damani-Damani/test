package webserver

import (
	rt "controlserver/realtime"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	cm      *rt.ConnectionManager
	conn    *websocket.Conn
	robotId int
	Send    chan []byte
}

type WsMessageType struct {
	Topic string `json:"topic"`
}

func (c *Client) Sender() {
	pingTicker := time.NewTicker(5 * time.Second)
	defer func() {
		pingTicker.Stop()
		c.cm.RobotConnections[c.robotId].RemoveClient(c) // Remove the client from the connection manager
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.conn.SetWriteDeadline(time.Time{})
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(6 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Reader() {
	var topic WsMessageType
	connection, err := c.cm.GetConnectionForRobot(c.robotId)
	if err != nil {
		return
	}

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			connection.RemoveClient(c)
			break
		}

		err = json.Unmarshal(message, &topic)
		if err != nil {
			continue // Simply do not process the message
		}

		switch topic.Topic {
		case "setMode":
			if connection.GetRobotSetModeStream.Connected() {
				connection.GetRobotSetModeStream.Send(message)
			}
		case "setPoint":
			if connection.GetWaypointsStream.Connected() {
				connection.GetWaypointsStream.Send(message)
			}
		case "joystickValues":
			if connection.GetJoystickControlStream.Connected() {
				connection.GetJoystickControlStream.Send(message)
			}
		case "peripheralControl":
			if connection.GetPeripheralControlStream.Connected() {
				connection.GetPeripheralControlStream.Send(message)
			}
		default:
			log.Printf("Received unimplemented type: %s", topic.Topic)
		}
	}
}

func (c *Client) Disconnect() {
	err := c.conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
	close(c.Send)
}
