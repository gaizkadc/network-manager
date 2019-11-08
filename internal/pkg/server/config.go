/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
	// CACertPath path for the CA
	CACertPath string
	// ClientCertPath path for the CA
	ClientCertPath string
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
	if conf.CACertPath == "" {
		return derrors.NewInvalidArgumentError("CA Cert Path must be defined")
	}
	if conf.ClientCertPath == "" {
		return derrors.NewInvalidArgumentError("Client Cert Path must be defined")
	}

	return nil
}

func (conf *Config) Print() {
	log.Info().Interface("configuration", conf).Msg("defined network manager configuration")
}
