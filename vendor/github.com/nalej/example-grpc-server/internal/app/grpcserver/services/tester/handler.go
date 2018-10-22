//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the implementation of the Tester handler.

package tester

import (
	"context"
	"errors"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/example-grpc/tester"
	"github.com/rs/zerolog/log"
)

// Handler structure that will implement the ping operations.
type Handler struct{}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// Process a complex request.
func (h *Handler) ProcessComplexRequest(ctx context.Context, request *tester.ComplexRequest) (*tester.Response, error) {
	if request == nil {
		return nil, errors.New("invalid request")
	}
	log.Info().Interface("request", *request).Msg("ProcessComplexRequest")

	if request.InduceFailure {
		iErr := derrors.NewOperationError("user induced error")
		log.Error().Interface("error", iErr).Msg("user requested internal failure")
		return nil, iErr
	}

	log.Debug().Int32("requestNumber", request.RequestNumber).
		Str("name", request.Name).Msg("")
	if request.Metadata != nil {
		m := request.Metadata

		if m.Type == tester.MetadataType_TYPE_A {
			log.Debug().Msg("Metadata TYPE_A")
		} else if m.Type == tester.MetadataType_TYPE_B {
			log.Debug().Msg("Metadata TYPE_B")
		} else {
			log.Warn().Msg("Unknown metadata type")
		}

		log.Debug().Strs("tags", m.Tags).Msg("")
	}

	response := tester.Response{Msg: fmt.Sprintf("Received: %d", request.RequestNumber), IsValid: true}
	return &response, nil
}
