// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing Service API
//
// API for the Load Balancing Service
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// BackendHealth The health status of the specified backend server as reported by the primary and standby load balancers.
type BackendHealth struct {

	// A list of the most recent health check results returned for the specified backend server.
	HealthCheckResults []HealthCheckResult `mandatory:"true" json:"healthCheckResults"`

	// The general health status of the specified backend server as reported by the primary and standby load balancers.
	// *   **OK:** Both health checks returned `OK`.
	// *   **WARNING:** One health check returned `OK` and one did not.
	// *   **CRITICAL:** Neither health check returned `OK`.
	// *   **UNKNOWN:** One or both health checks returned `UNKNOWN`, or the system was unable to retrieve metrics at this time.
	Status BackendHealthStatusEnum `mandatory:"true" json:"status"`
}

func (m BackendHealth) String() string {
	return common.PointerString(m)
}

// BackendHealthStatusEnum Enum with underlying type: string
type BackendHealthStatusEnum string

// Set of constants representing the allowable values for BackendHealthStatus
const (
	BackendHealthStatusOk       BackendHealthStatusEnum = "OK"
	BackendHealthStatusWarning  BackendHealthStatusEnum = "WARNING"
	BackendHealthStatusCritical BackendHealthStatusEnum = "CRITICAL"
	BackendHealthStatusUnknown  BackendHealthStatusEnum = "UNKNOWN"
)

var mappingBackendHealthStatus = map[string]BackendHealthStatusEnum{
	"OK":       BackendHealthStatusOk,
	"WARNING":  BackendHealthStatusWarning,
	"CRITICAL": BackendHealthStatusCritical,
	"UNKNOWN":  BackendHealthStatusUnknown,
}

// GetBackendHealthStatusEnumValues Enumerates the set of values for BackendHealthStatus
func GetBackendHealthStatusEnumValues() []BackendHealthStatusEnum {
	values := make([]BackendHealthStatusEnum, 0)
	for _, v := range mappingBackendHealthStatus {
		values = append(values, v)
	}
	return values
}
