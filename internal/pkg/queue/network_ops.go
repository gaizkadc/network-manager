/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package queue

import (
    "context"
    "github.com/nalej/nalej-bus/pkg/queue/network/ops"
    "github.com/nalej/network-manager/internal/pkg/server/application"
    "github.com/nalej/network-manager/internal/pkg/server/dns"
    "github.com/nalej/network-manager/internal/pkg/server/networks"
    "github.com/rs/zerolog/log"
    "time"
)

// Timeout between incoming messages
const NetworkOpsTimeout = time.Minute * 60

type NetworkOpsHandler struct {
    // reference manager for networks
    netManager *networks.Manager
    // dns manager
    dnsManager *dns.Manager
    // network application manager
    netAppManager *application.Manager
    // operations consumer
    consumer *ops.NetworkOpsConsumer
}

// Instantiate a new network ops handler to manipulate messages from the network ops queue.
// params:
//  netManager
//  cons
func NewNetworkOpsHandler(netManager *networks.Manager, dnsManager *dns.Manager, netAppManager *application.Manager,
    consumer *ops.NetworkOpsConsumer) NetworkOpsHandler {
    return NetworkOpsHandler{netManager: netManager, dnsManager: dnsManager, netAppManager: netAppManager, consumer: consumer}
}

func(n NetworkOpsHandler) Run() {
    go n.consumeAuthorizeMemberRequest()
    go n.consumeDisauthorizeMemberRequest()
    go n.consumeAddDNSEntryRequest()
    go n.consumeDeleteDNSEntryRequest()
    go n.consumeInboundProxy()
    go n.consumeOutboundProxy()
    go n.consumeAddConnectionRequest()
    go n.consumeRemoveConnectionRequest()
    go n.consumeAuthorizeZTConnectionRequest()
    go n.waitRequests()
}

// Endless loop waiting for requests
func (n NetworkOpsHandler) waitRequests() {
    log.Debug().Msg("wait for requests to be received by the network ops queue")
    for {
        somethingReceived := false
        ctx, cancel := context.WithTimeout(context.Background(), NetworkOpsTimeout)
        currentTime := time.Now()
        err := n.consumer.Consume(ctx)
        somethingReceived = true
        cancel()
        select {
        case <- ctx.Done():
            // the timeout was reached
            if !somethingReceived {
                log.Debug().Msgf("no message received since %s",currentTime.Format(time.RFC3339))
            }
        default:
            if err != nil {
                log.Error().Err(err).Msg("error consuming data from network ops")
            }
        }
    }
}

func(n NetworkOpsHandler) consumeAuthorizeMemberRequest () {
    log.Debug().Msg("waiting for authorize member requests...")
    for {
        received := <- n.consumer.Config.ChAuthorizeMembersRequest
        log.Debug().Interface("authorizeMemberRequest", received).Msg("<- incoming authorize member request")
        err := n.netManager.AuthorizeMember(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing authorize member request")
        }
    }
}

func(n NetworkOpsHandler) consumeDisauthorizeMemberRequest() {
    log.Debug().Msg("waiting for disauthorize member requests...")
    for {
        received := <- n.consumer.Config.ChDisauthorizeMembersRequest
        log.Debug().Interface("disauthorizeMemberRequest",received).Msg("<- incoming disauthorize member request")
        err := n.netManager.UnauthorizeMember(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing unauthorize member request")
        }
    }
}

func(n NetworkOpsHandler) consumeAddDNSEntryRequest() {
    log.Debug().Msg("waiting for consume add dns entry request...")
    for {
        received := <- n.consumer.Config.ChAddDNSEntryRequest
        log.Debug().Interface("addDNSEntryRequest", received).Msg("<- incoming add dns entry request")
        err := n.dnsManager.AddDNSEntry(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing add dns entry request")
        }
    }
}

func(n NetworkOpsHandler) consumeDeleteDNSEntryRequest() {
    log.Debug().Msg("waiting for consume delete dns entry request...")
    for {
        received := <- n.consumer.Config.ChDeleteDNSEntryRequest
        log.Debug().Interface("deleteDNSEntryRequest", received).Msg("<- incoming delete dns entry request")
        err := n.dnsManager.DeleteDNSEntry(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing delete dns entry request")
        }
    }
}

func (n NetworkOpsHandler) consumeInboundProxy() {
    log.Debug().Msg("waiting for consume inbound proxy request...")
    for {
        received := <- n.consumer.Config.ChInboundServiceProxy
        log.Debug().Interface("inboundServiceProxy", received).Msg("<- incoming inbound service proxy")
        err := n.netAppManager.RegisterInboundServiceProxy(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing inbound service proxy")
        }
    }
}

func (n NetworkOpsHandler) consumeOutboundProxy() {
    log.Debug().Msg("waiting for consume outbound proxy request...")
    for {
        received := <- n.consumer.Config.ChOutboundService
        log.Debug().Interface("outboundServiceProxy", received).Msg("<- incoming outbound service proxy")
        err := n.netAppManager.RegisterOutboundProxy(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing outbound service proxy")
        }
    }
}

func (n NetworkOpsHandler) consumeAddConnectionRequest() {
    log.Debug().Msg("waiting for consume add connection request...")
    for {
        received := <- n.consumer.Config.ChAddConnectionRequest
        log.Debug().Interface("connection request", received).Msg("<- incoming add connection request")
        err := n.netAppManager.AddConnection(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing add connection request")
        }
    }
}

func (n NetworkOpsHandler) consumeRemoveConnectionRequest() {
    log.Debug().Msg("waiting for consume remove connection request...")
    for {
        received := <- n.consumer.Config.ChRemoveConnectionRequest
        log.Debug().Interface("connection request", received).Msg("<- incoming remove connection request")
        err := n.netAppManager.RemoveConnection(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing remove connection request")
        }
    }
}

func (n NetworkOpsHandler) consumeAuthorizeZTConnectionRequest() {
    log.Debug().Msg("waiting for consume authorize ZT connection request...")
    for {
        received := <- n.consumer.Config.ChAuthorizeZTConnection
        log.Debug().Interface("connection request", received).Msg("<- incoming authorize ZT connection request")
        err := n.netManager.AuthorizeZTConnection(received)
        if err != nil {
            log.Error().Err(err).Msg("failed processing authorize connection request")
        }
    }
}

