package tester_test

import (
	"context"
	"github.com/nalej/example-grpc-server/internal/app/grpcserver/services/tester"
	pbTester "github.com/nalej/example-grpc/tester"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

const bufSize = 1024 * 1024

var listener *bufconn.Listener

func LaunchServer() {
	listener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pbTester.RegisterTesterServer(s, tester.NewHandler())
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatal().Errs("failed to listen: %v", []error{err})
		}
	}()
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return listener.Dial()
}

func getClient(t *testing.T) (pbTester.TesterClient, *grpc.ClientConn) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	client := pbTester.NewTesterClient(conn)
	return client, conn
}

func init() {
	LaunchServer()
}

func TestComplexRequestFail(t *testing.T) {
	client, conn := getClient(t)
	defer conn.Close()
	_, err := client.ProcessComplexRequest(context.Background(),
		&pbTester.ComplexRequest{InduceFailure: true})
	if err == nil {
		t.Fatal("expecting server failure")
	}
}

func TestComplexRequest(t *testing.T) {
	client, conn := getClient(t)
	defer conn.Close()
	response, err := client.ProcessComplexRequest(context.Background(),
		&pbTester.ComplexRequest{RequestNumber: 1, Name: "test request"})
	if err != nil {
		t.Fatal("expecting response", err)
	}
	if response.Msg != "Received: 1" {
		t.Fatalf("wrong response, received: %s", response.Msg)
	}
}
