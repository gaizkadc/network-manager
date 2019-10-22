/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package queue

import (
	"context"
	"github.com/nalej/nalej-bus/pkg/queue/application/events"
	"github.com/nalej/network-manager/internal/pkg/server/application"
	"github.com/rs/zerolog/log"
	"time"
)

const ApplicationEventsTimeout = time.Minute * 60


type AppEventsHandler struct {
	// network application manager
	netAppManager *application.Manager
	// operations consumer
	consumer *events.ApplicationEventsConsumer
}

func NewAppEventsHandler (netAppManager *application.Manager, consumer *events.ApplicationEventsConsumer) AppEventsHandler {
	return AppEventsHandler{netAppManager:netAppManager, consumer:consumer}
}

func (a AppEventsHandler) Run () {
	go a.consumeDeploymentServiceStatusUpdateRequest()
	go a.waitRequests()
}

// Endless loop waiting for requests
func (a AppEventsHandler) waitRequests() {
	log.Debug().Msg("wait for requests to be received by the application events queue")
	for {
		somethingReceived := false
		ctx, cancel := context.WithTimeout(context.Background(), ApplicationEventsTimeout)
		currentTime := time.Now()
		err := a.consumer.Consume(ctx)
		somethingReceived = true
		cancel()
		select {
		case <- ctx.Done():
			// the timeout was reached
			if !somethingReceived {
				log.Debug().Str("since", currentTime.Format(time.RFC3339)).Msgf("no message received")
			}
		default:
			if err != nil {
				log.Error().Err(err).Msg("error consuming data from application events")
			}
		}
	}
}

// conductor sends DeploymentServiceStatusUpdateRequest to the bus ant network-manager consumes them
func(a AppEventsHandler) consumeDeploymentServiceStatusUpdateRequest () {
	log.Debug().Msg("waiting for service status update requests...")
	for {
		received := <- a.consumer.Config.ChDeploymentServiceStatusUpdateRequest
		log.Debug().Interface("DeploymentServiceStatusUpdateRequest", received).Msg("<- incoming deployment service status update request")
		err := a.netAppManager.ManageConnections(received)
		if err != nil {
			log.Error().Err(err).Msg("failed processing deployment service status update request")
		}
	}
}