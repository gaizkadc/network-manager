//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// Ping client

package pinger

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"github.com/rs/zerolog/log"
	"github.com/nalej/example-grpc/ping"
	"time"
)

type Pinger struct {
	Host string
	Port int
}

func NewPinger(host string, port int) *Pinger{
	return &Pinger{host, port}
}

func (p * Pinger) Ping(numPings int) {
	address := fmt.Sprintf("%s:%d", p.Host, p.Port)
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Errs("unable to connect: %v", []error{err})
	}
	defer conn.Close()
	c := ping.NewPingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for it := 0; it < numPings; it++ {
		r, err := c.Ping(ctx, &ping.PingRequest{RequestNumber: int32(it)})
		if err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Interface("response", r).Msg("Received")
	}

}
