/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Address where the API service will listen requests.
	Port int
	// System model url
	SystemModelURL string
	// ZT url
	ZTUrl string
	// ZT access token
	ZTAccessToken string
	// Consul DNS URL
	DNSUrl string
	// URL for the message queue
	QueueAddress string
}

func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be defined")
	}
	if conf.SystemModelURL == "" {
		return derrors.NewInvalidArgumentError("System Model URL must be defined")
	}
	if conf.ZTUrl == "" {
		return derrors.NewInvalidArgumentError("ZT URL must be defined")
	}
	if conf.ZTAccessToken == "" {
		return derrors.NewInvalidArgumentError("ZT Access Token must be defined")
	}
	if conf.DNSUrl == "" {
		return derrors.NewInvalidArgumentError("DNS URL must be defined")
	}
	if conf.QueueAddress == "" {
		return derrors.NewInvalidArgumentError("Queue URL must be defined")
	}
	return nil
}

func (conf *Config) Print() {
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("sm", conf.SystemModelURL).Msg("SM URL")
	log.Info().Str("zturl", conf.ZTUrl).Msg("ZT URL")
	log.Info().Str("ztaccesstoken", conf.ZTAccessToken).Msg("ZT Access Token")
	log.Info().Str("dnsurl", conf.DNSUrl).Msg("Consul DNS URL")
	log.Info().Str("queueAddress", conf.QueueAddress).Msg("Queue address")
}
