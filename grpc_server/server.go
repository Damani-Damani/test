package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "controlserver/proto"
	rt "controlserver/realtime"
	webserver "controlserver/web_server"

	"google.golang.org/grpc"
)

var (
	port = 50051
)

type controlServer struct {
	pb.UnimplementedControlServiceServer
	connectionManager *rt.ConnectionManager
}

func newControlServer(connectionManager *rt.ConnectionManager) *controlServer {
	return &controlServer{
		connectionManager: connectionManager,
	}
}

func SendToClients(data []byte, clients []rt.Client) {
	for _, client := range clients {
		if cc, ok := (client).(*webserver.Client); ok {
			cc.Send <- data
		}
	}
}

func (c *controlServer) SendRobotStatus(stream pb.ControlService_SendRobotStatusServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			log.Printf("Stream ended at: %s", endTime.String())
		}
		if err != nil {
			return err
		}

		clients, err := c.connectionManager.GetClientsForRobot(int(data.RobotId))
		if err != nil {
			continue
		}

		bytesData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		SendToClients(bytesData, clients)
	}
}

func (c *controlServer) SendLocationStatus(stream pb.ControlService_SendLocationStatusServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			log.Printf("Stream ended at: %s", endTime.String())
		}
		if err != nil {
			return err
		}

		clients, err := c.connectionManager.GetClientsForRobot(int(data.RobotId))
		if err != nil {
			continue
		}

		bytesData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		SendToClients(bytesData, clients)
	}
}

func (c *controlServer) SendWaypointStream(stream pb.ControlService_SendWaypointStreamServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			log.Printf("Stream ended at: %s", endTime.String())
		}
		if err != nil {
			return err
		}

		clients, err := c.connectionManager.GetClientsForRobot(int(data.RobotId))
		if err != nil {
			continue
		}

		bytesData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		SendToClients(bytesData, clients)
	}
}

func (c *controlServer) GetRobotSetModeStream(request *pb.GetRobotSetModeRequest, stream pb.ControlService_GetRobotSetModeStreamServer) error {
	connection, err := c.connectionManager.GetConnectionForRobot(int(request.RobotId))
	if err != nil {
		return err
	}
	connection.GetRobotSetModeStream.Add()
	var r pb.GetRobotSetModeData
	for {
		data := <-connection.GetRobotSetModeStream.Channel
		err := json.Unmarshal(data, &r)
		if err != nil {
			continue // Dont send data to stream
		}
		if err := stream.Send(&r); err != nil {
			connection.GetRobotSetModeStream.Remove()
			return err
		}
	}
}

func (c *controlServer) GetWaypointsStream(request *pb.GetWaypointsRequest, stream pb.ControlService_GetWaypointsStreamServer) error {
	connection, err := c.connectionManager.GetConnectionForRobot(int(request.RobotId))
	if err != nil {
		return err
	}
	connection.GetWaypointsStream.Add()
	var r *pb.WaypointsData
	for {
		data := <-connection.GetWaypointsStream.Channel
		err := json.Unmarshal(data, r)
		if err != nil {
			log.Printf("Data from websocket is not valid RobotSetModeStream type.")
			continue // Dont send data to stream
		}
		if err := stream.Send(r); err != nil {
			connection.GetWaypointsStream.Remove()
			return err
		}
	}
}

func (c *controlServer) GetPeripheralControlStream(request *pb.GetPeripheralControlRequest, stream pb.ControlService_GetPeripheralControlStreamServer) error {
	connection, err := c.connectionManager.GetConnectionForRobot(int(request.RobotId))
	if err != nil {
		return err
	}
	connection.GetPeripheralControlStream.Add()
	var r *pb.Peripherals
	for {
		data := <-connection.GetPeripheralControlStream.Channel
		err := json.Unmarshal(data, r)
		if err != nil {
			log.Printf("Data from websocket is not valid RobotSetModeStream type.")
			continue // Dont send data to stream
		}
		if err := stream.Send(r); err != nil {
			connection.GetPeripheralControlStream.Remove()
			return err
		}
	}
}

func (c *controlServer) GetJoystickControlStream(request *pb.GetJoystickControlRequest, stream pb.ControlService_GetJoystickControlStreamServer) error {
	connection, err := c.connectionManager.GetConnectionForRobot(int(request.RobotId))
	if err != nil {
		return err
	}
	connection.GetJoystickControlStream.Add()
	var r *pb.VehicleThrustTorque
	for {
		data := <-connection.GetJoystickControlStream.Channel
		err := json.Unmarshal(data, r)
		if err != nil {
			log.Printf("Data from websocket is not valid RobotSetModeStream type.")
			continue // Dont send data to stream
		}
		if err := stream.Send(r); err != nil {
			connection.GetJoystickControlStream.Remove()
			return err
		}
	}
}

func CreateServer(ctx context.Context, connectionManager *rt.ConnectionManager) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	cs := newControlServer(connectionManager)
	pb.RegisterControlServiceServer(s, cs)

  go func() {
    <- ctx.Done()
    log.Println("Stopping gRPC server")
    s.Stop()
  }()

	s.Serve(l)
}
