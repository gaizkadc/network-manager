/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */


package utils

import (
    "crypto/tls"
    "crypto/x509"
    "errors"
    "github.com/nalej/derrors"
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
    // skip CA validation
    skipCAValidation bool
}

func NewConnectionsHelper(useTLS bool, caCertPath string, skipCAValidation bool) *ConnectionsHelper {

    return &ConnectionsHelper{
        ClusterReference: make(map[string]ClusterEntry, 0),
        useTLS: useTLS,
        caCertPath: caCertPath,
        skipCAValidation: skipCAValidation,
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
//   skipCAValidation skip the validation of the CA
//  return:
//   client and error if any
func clusterClientFactory(hostname string, port int, params...interface{}) (*grpc.ClientConn, error) {
    log.Debug().Str("hostname", hostname).Int("port", port).Int("len", len(params)).Interface("params", params).Msg("calling cluster client factory")
    if len(params) != 3 {
        log.Fatal().Interface("params",params).Msg("cluster client factory called with not enough parameters")
    }
    useTLS := params[0].(bool)
    caCertPath := params[1].(string)
    skipCAValidation := params[2].(bool)
    return secureClientFactory(hostname, port, useTLS, caCertPath, skipCAValidation)
}

// Factory in charge of generation a secure connection with a grpc server.
//  params:
//   hostname of the target server
//   port of the target server
//   useTLS flag indicating whether to use the TLS security
//   caCert path of the CA certificate
//   skipCAValidation skip the validation of the CA
//  return:
//   grpc connection and error if any
func secureClientFactory(hostname string, port int, useTLS bool, caCertPath string, skipCAValidation bool) (*grpc.ClientConn, error) {
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
    log.Debug().Str("address", targetAddress).Bool("useTLS", useTLS).Str("caCertPath", caCertPath).Bool("skipCAValidation", skipCAValidation).Msg("creating secure connection")

    if skipCAValidation {
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
            h.ClusterReference[cluster.ClusterId] = ClusterEntry{Hostname: targetHostname, Cordon: cluster.Cordon}
            targetPort := int(APP_CLUSTER_API_PORT)
            params := make([]interface{}, 0)
            params = append(params, h.useTLS)
            params = append(params, h.caCertPath)
            params = append(params, h.skipCAValidation)

            clusters.AddConnection(targetHostname, targetPort, params ... )
            toReturn = append(toReturn, targetHostname)
        }
    }
    return nil
}

// Internal function to check if a cluster meets all the conditions to be added to the list of available clusters.
func (h * ConnectionsHelper) isClusterAvailable(cluster *grpc_infrastructure_go.Cluster) bool {
    if cluster.Status != grpc_infrastructure_go.InfraStatus_RUNNING {
        log.Debug().Str("clusterID", cluster.ClusterId).Msg("cluster ignored because it is not running")
        return false
    }
    // Others...
    return true
}