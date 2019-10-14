/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */


package utils

import (
    "crypto/tls"
    "crypto/x509"
    "errors"
    "github.com/nalej/derrors"
    grpc_connectivity_manager_go "github.com/nalej/grpc-connectivity-manager-go"
    "github.com/nalej/grpc-infrastructure-go"
    "github.com/nalej/grpc-organization-go"
    "github.com/nalej/grpc-utils/pkg/tools"
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "io/ioutil"
    "sync"
    "fmt"
    "context"
)

const(
    // App cluster api port
    APP_CLUSTER_API_PORT uint32 = 443
)

// Internal struct to store information about a cluster connection. This struct can be used to query the latest
// known status of a cluster.
type ClusterEntry struct {
    // Hostname for this cluster
    Hostname string
    // Cordon true if this cluster is in a cordon status
    Cordon bool
}



// The connection helpers assists in the maintenance and connection of several app clusters.
type ConnectionsHelper struct {
    // Application cluster clients
    AppClusterClients *tools.ConnectionsMap
    // Singleton instance of connections
    onceAppClusterClients sync.Once
    // Translation map between cluster ids and their ip addresses
    ClusterReference map[string]ClusterEntry
    // useTLS connections
    useTLS bool
    // path for the CA
    caCertPath string
    // path for the Client cert
    clientCertPath string
    // skip CA validation
    SkipServerCertValidation bool
}

func NewConnectionsHelper(useTLS bool, clientCertPath string, caCertPath string, skipServerCertValidation bool) *ConnectionsHelper {

    return &ConnectionsHelper{
        ClusterReference: make(map[string]ClusterEntry, 0),
        useTLS: useTLS,
        clientCertPath: clientCertPath,
        caCertPath: caCertPath,
        SkipServerCertValidation: skipServerCertValidation,
    }
}

func (h *ConnectionsHelper) GetAppClusterClients() *tools.ConnectionsMap {
    h.onceAppClusterClients.Do(func(){
        h.AppClusterClients = tools.NewConnectionsMap(clusterClientFactory)
        if h.ClusterReference == nil {
            h.ClusterReference = make(map[string]ClusterEntry, 0)
        }
    })
    return h.AppClusterClients
}


// Factory in charge of generating new connections for Conductor->cluster communication.
//  params:
//   hostname of the target server
//   port of the target server
//   useTLS flag indicating whether to use the TLS security
//   caCert path of the CA certificate
//   SkipServerCertValidation skip the validation of the CA
//  return:
//   client and error if any
func clusterClientFactory(hostname string, port int, params...interface{}) (*grpc.ClientConn, error) {
    log.Debug().Str("hostname", hostname).Int("port", port).Int("len", len(params)).Interface("params", params).Msg("calling cluster client factory")
    if len(params) != 4 {
        log.Fatal().Interface("params",params).Msg("cluster client factory called with not enough parameters")
    }
    useTLS := params[0].(bool)
    clientCertPath := params[1].(string)
    caCertPath := params[2].(string)
    skipServerCertValidation := params[3].(bool)
    return secureClientFactory(hostname, port, useTLS, clientCertPath, caCertPath, skipServerCertValidation)
}

// Factory in charge of generation a secure connection with a grpc server.
//  params:
//   hostname of the target server
//   port of the target server
//   useTLS flag indicating whether to use the TLS security
//   clientCertPath to the client cert
//   caCertPath of the CA certificate
//   skipServerCertValidation skip the validation of the CA
//  return:
//   grpc connection and error if any
func secureClientFactory(hostname string, port int, useTLS bool, clientCertPath string, caCertPath string, skipServerCertValidation bool) (*grpc.ClientConn, error) {
    rootCAs := x509.NewCertPool()
    tlsConfig := &tls.Config{
        ServerName:   hostname,
    }

    if caCertPath != "" {
        log.Debug().Str("caCertPath", caCertPath).Msg("loading CA cert")
        caCert, err := ioutil.ReadFile(caCertPath)
        if err != nil {
            return nil, derrors.NewInternalError("Error loading CA certificate")
        }
        added := rootCAs.AppendCertsFromPEM(caCert)
        if !added {
            return nil, derrors.NewInternalError("cannot add CA certificate to the pool")
        }
        tlsConfig.RootCAs = rootCAs
    }

    targetAddress := fmt.Sprintf("%s:%d", hostname, port)
    log.Debug().Str("address", targetAddress).Bool("useTLS", useTLS).Str("caCertPath", caCertPath).Bool("skipServerCertValidation", skipServerCertValidation).Msg("creating secure connection")

    if clientCertPath != "" {
        log.Debug().Str("clientCertPath", clientCertPath).Msg("loading client certificate")
        clientCert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/tls.crt", clientCertPath),fmt.Sprintf("%s/tls.key", clientCertPath))
        if err != nil {
            log.Error().Str("error", err.Error()).Msg("Error loading client certificate")
            return nil, derrors.NewInternalError("Error loading client certificate")
        }

        tlsConfig.Certificates = []tls.Certificate{clientCert}
        tlsConfig.BuildNameToCertificate()
    }

    if skipServerCertValidation {
        tlsConfig.InsecureSkipVerify = true
    }

    creds := credentials.NewTLS(tlsConfig)

    log.Debug().Interface("creds", creds.Info()).Msg("Secure credentials")
    sConn, dErr := grpc.Dial(targetAddress, grpc.WithTransportCredentials(creds))
    if dErr != nil {
        log.Error().Err(dErr).Msg("impossible to create secure client factory connection")
        return nil, derrors.AsError(dErr, "cannot create connection with the signup service")
    }

    return sConn, nil

}

// This is a common sharing function to check the system model and update the available clusters.
// Additionally, the function updates the available connections for musicians and deployment managers.
// The common ClusterReference object is updated with the cluster ids and the corresponding ip.
//  params:
//   organizationId
func(h *ConnectionsHelper) UpdateClusterConnections(organizationId string, client grpc_infrastructure_go.ClustersClient ) error{
    log.Debug().Msg("update cluster connections...")
    // Rebuild the map
    h.ClusterReference = make(map[string]ClusterEntry,0)

    req := grpc_organization_go.OrganizationId{OrganizationId:organizationId}
    clusterList, err := client.ListClusters(context.Background(), &req)
    if err != nil {
        msg := fmt.Sprintf("there was a problem getting the list of " +
            "available cluster for org %s",organizationId)
        log.Error().Err(err).Msg(msg)
        return errors.New(msg)
    }

    toReturn := make([]string,0)
    clusters := h.GetAppClusterClients()

    for _, cluster := range clusterList.Clusters {
        // The cluster is running and is not in cordon status
        if h.isClusterAvailable(cluster){
            targetHostname := fmt.Sprintf("appcluster.%s", cluster.Hostname)
            clusterCordon := cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE_CORDON || cluster.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_OFFLINE_CORDON
            h.ClusterReference[cluster.ClusterId] = ClusterEntry{Hostname: targetHostname, Cordon: clusterCordon}
            targetPort := int(APP_CLUSTER_API_PORT)
            params := make([]interface{}, 0)
            params = append(params, h.useTLS)
            params = append(params, h.clientCertPath)
            params = append(params, h.caCertPath)
            params = append(params, h.SkipServerCertValidation)

            clusters.AddConnection(targetHostname, targetPort, params ... )
            toReturn = append(toReturn, targetHostname)
        }
    }
    return nil
}

// Internal function to check if a cluster meets all the conditions to be added to the list of available clusters.
func (h * ConnectionsHelper) isClusterAvailable(cluster *grpc_infrastructure_go.Cluster) bool {
    // TODO: when state is implemented, check this ->
    //if cluster.State != pbInfrastructure.InfraStatus_RUNNING {
    //    log.Debug().Str("clusterID", cluster.ClusterId).Msg("cluster ignored because it is not running")
    //    return false
    //}

    if cluster.ClusterStatus != grpc_connectivity_manager_go.ClusterStatus_ONLINE {
        log.Debug().Str("clusterID", cluster.ClusterId).Msg("cluster ignored because it is not available")
        return false
    }

    // Others...
    return true
}