package ping_test

import (
	"context"
	"github.com/nalej/example-grpc-server/internal/app/grpcserver/services/ping"
	pbPing "github.com/nalej/example-grpc/ping"
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
	pbPing.RegisterPingServer(s, ping.NewHandler())
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatal().Errs("failed to listen: %v", []error{err})
		}
	}()
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return listener.Dial()
}

func getClient(t *testing.T) (pbPing.PingClient, *grpc.ClientConn) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	client := pbPing.NewPingClient(conn)
	return client, conn
}

func TestPing(t *testing.T) {
	LaunchServer()
	client, conn := getClient(t)
	defer conn.Close()
	response, err := client.Ping(context.Background(), &pbPing.PingRequest{RequestNumber: 1})
	if err != nil {
		t.Fatal("ping failed", err)
	}
	if response.Msg != "Pong: 1" {
		t.Fatal("wrong answer")
	}
}
