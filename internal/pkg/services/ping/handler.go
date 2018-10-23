//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the implementation of the ping handler.

package ping

import (
	"context"
	"errors"
	"fmt"
	"github.com/nalej/example-grpc/ping"
	"github.com/rs/zerolog/log"
)

// Handler structure that will implement the ping operations.
type Handler struct{}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(ctx context.Context, request *ping.PingRequest) (*ping.PingResponse, error) {
	if request == nil {
		return nil, errors.New("invalid request")
	}
	log.Info().Interface("request", *request).Msg("Ping")
	response := ping.PingResponse{Msg: fmt.Sprintf("Pong: %d", request.RequestNumber)}
	return &response, nil
}