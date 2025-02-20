package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	pb "controlserver/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
  serverAddr = fmt.Sprint("v2.controlserver.clearbot.dev:443")
)

func testClient() {
	creds := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	conn, err := grpc.NewClient(serverAddr, creds)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewControlServiceClient(conn)
	ctx := context.Background()

	stream, err := client.SendRobotStatus(ctx)
	var data *pb.RobotStatusData
	var temp float32

	temp = 10
	for {
		temp = temp + 0.1
		data = &pb.RobotStatusData{
			RobotId:  1,
			ArmState: true,
			VehicleControlMode: &pb.VehicleControlMode{
				Manual:   false,
				Offboard: false,
				Auto:     false,
			},
			ClearbotControlMode: &pb.ClearbotControlMode{
				JoystickControl:   false,
				ObstacleAvoidance: false,
				HeadingControl:    false,
				WaypointControl:   false,
			},
			Peripherals: &pb.Peripherals{
				Conveyor: false,
			},
			SystemInfo: &pb.SystemInfo{
				Temperature: temp,
			},
		}
		err := stream.Send(data)
		if err != nil {
			log.Print(err)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}
}

func testServerStreaming() {
	creds := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	conn, err := grpc.NewClient(serverAddr, creds)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewControlServiceClient(conn)
	ctx := context.Background()

  req := pb.GetRobotSetModeRequest{
    RobotId: 1,
  }

	stream, err := client.GetRobotSetModeStream(ctx, &req)
  if err != nil {
    log.Fatal(err)
  }

	for {
		data, err := stream.Recv()
		if err != nil {
			log.Print(err)
			break
		}
    log.Println(data)
	}
}


func main() {
  // testClient()
  testServerStreaming()
}
