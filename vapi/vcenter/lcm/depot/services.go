// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package depot

import (
	"context"
	"net/http"

	"github.com/vmware/govmomi/vapi/rest"
)

const (
	basePath = "/api/vcenter/lcm/depot/services"
)

// Manager extends rest.Client, adding vCenter LCM Depot Services related methods.
type Manager struct {
	*rest.Client
}

// NewManager creates a new Manager instance with the given client.
func NewManager(client *rest.Client) *Manager {
	return &Manager{
		Client: client,
	}
}

// AddressType defines the various types of address used for service endpoint.
type AddressType string

const (
	// AddressTypeFqdn represents Fully Qualified Domain Name.
	AddressTypeFqdn AddressType = "Fqdn"
	// AddressTypeIPv4 represents IPv4 address format.
	AddressTypeIPv4 AddressType = "IPv4"
	// AddressTypeIPv6 represents IPv6 address format.
	AddressTypeIPv6 AddressType = "IPv6"
)

// Address defines external service node address configurations.
type Address struct {
	// Type specifies the type of address used for the service.
	Type AddressType `json:"type"`
	// Value is the actual address corresponding to the specified type.
	Value string `json:"value"`
}

// Node defines the external service node configurations.
type Node struct {
	// Name is the common name of the service node.
	Name string `json:"name"`
	// Addresses is the list of addresses to connect to the service node.
	Addresses []Address `json:"addresses"`
	// Port is the service node port on which the service is hosted.
	Port string `json:"port"`
	// BaseUrl is the URL prefix pointing to the root of the service.
	BaseUrl string `json:"base_url"`
	// Certificates to be used to securely connect with the service node.
	Certificates []string `json:"certificates"`
}

// Service defines the connection configuration for specific service.
type Service struct {
	// Name is the external service name.
	Name string `json:"name"`
	// Type is the type of external service.
	// Supported types are fds-files_server and fds-image_registry.
	Type string `json:"type"`
	// Key is the external service key, identifier for the service unique across the fleet.
	Key string `json:"key"`
	// Version is the external service version.
	Version string `json:"version"`
	// Nodes is the list of nodes/instances belonging to the same external service.
	Nodes []Node `json:"nodes"`
}

// SetSpec defines the connection configuration for all external services.
type SetSpec struct {
	// Services is a list of configurations which contains the connection details of the external services
	// to be registered with vCenter appliance.
	Services []Service `json:"services"`
}

// Info defines the connection configuration for all external services.
type Info struct {
	// Services lists the configurations containing the connection details of the external services
	// which are registered with vCenter appliance.
	Services []Service `json:"services"`
}

// Set configures the external services connection details with vCenter appliance.
// Ref: https://developer.broadcom.com/xapis/vsphere-automation-api/latest/api/vcenter/lcm/depot/services/put
//
// Return errors:
//   - Unauthorized: if the caller is not authorized.
//   - Unauthenticated: if the caller is not authenticated.
//   - InvalidArgument: if passed arguments are invalid.
//   - Error: if there is some unknown internal error.
func (m *Manager) Set(ctx context.Context, config SetSpec) error {
	path := m.Resource(basePath)
	req := path.Request(http.MethodPut, config)
	return m.Do(ctx, req, nil)
}

// Get gets the configuration of the external services registered with vCenter appliance.
// Returns the configuration containing connection and certificate details of external services.
// Ref: https://developer.broadcom.com/xapis/vsphere-automation-api/latest/api/vcenter/lcm/depot/services/get
//
// Return errors:
//   - Unauthorized: if the caller is not authorized.
//   - Unauthenticated: if the caller is not authenticated.
//   - Error: if there is other error raised.
func (m *Manager) Get(ctx context.Context) (*Info, error) {
	path := m.Resource(basePath)
	req := path.Request(http.MethodGet)
	var info Info
	err := m.Do(ctx, req, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetServiceSpec gets the configuration of the specific external service registered with vCenter appliance.
// Returns the Service if the specific FDS service is configured, otherwise returns NotFound error.
// Ref: https://developer.broadcom.com/xapis/vsphere-automation-api/latest/api/vcenter/lcm/depot/services/<service-Type>/get
//
// Parameters:
//   - serviceType: specific external service type.
//
// Return errors:
//   - NotFound: if the source is not found
//   - Unauthenticated: if the session is not authenticated
//   - Unauthorized: if the session is not authorized to perform this operation.
//   - Error: if there is another generic error.
func (m *Manager) GetServiceSpec(ctx context.Context, serviceType string) (*Service, error) {
	path := m.Resource(basePath).WithSubpath(serviceType)
	req := path.Request(http.MethodGet)
	var service Service
	err := m.Do(ctx, req, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}
