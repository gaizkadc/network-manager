/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package utils

import (
    "fmt"
    "strings"
)

// Return how would be the VSA of a potential entry.
// params:
//  serviceName
//  organizationId
//  appInstanceId
// return:
//  the fqdn
func GetVSAName(serviceName string, organizationId string, appInstanceId string) string {
    value := fmt.Sprintf("%s-%s-%s", FormatName(serviceName), organizationId[0:10],
        appInstanceId[0:10])
    return value
}


// Format a string removing white spaces and going lowercase
func FormatName(name string) string {
    aux := strings.ToLower(name)
    // replace any space
    aux = strings.Replace(aux, " ", "", -1)
    return aux
}