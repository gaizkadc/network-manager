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
	// UseTLS to connect to the application cluster API
	UseTLS bool
	// CaCertPath path for the CA
	CaCertPath string
	// SkipServerCertValidation decide whether to skip CA validation or not
	SkipServerCertValidation bool
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
	log.Info().Interface("configuration", conf).Msg("defined network manager configuration")
}


