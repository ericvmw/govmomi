// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package depot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter/lcm/depot"
	"github.com/vmware/govmomi/vim25"

	_ "github.com/vmware/govmomi/vapi/simulator"
	_ "github.com/vmware/govmomi/vapi/vcenter/lcm/depot/simulator"
)

func TestDepotServicesAPI(t *testing.T) {
	simulator.Test(func(ctx context.Context, vc *vim25.Client) {
		rc := rest.NewClient(vc)

		err := rc.Login(ctx, simulator.DefaultLogin)
		require.NoError(t, err, "Failed to login")

		manager := depot.NewManager(rc)

		// Test 1: Get empty services initially
		t.Run("GetEmptyServices", func(t *testing.T) {
			info, err := manager.Get(ctx)
			require.NoError(t, err, "Failed to get services")
			assert.NotNil(t, info, "Info should not be nil")
			assert.Empty(t, info.Services, "Services should be empty initially")
		})

		// Test 2: Set services configuration
		t.Run("SetServices", func(t *testing.T) {
			setSpec := getTestServiceSetSpec()

			err := manager.Set(ctx, *setSpec)
			require.NoError(t, err, "Failed to set services")
		})

		// Test 3: Get all services after setting them
		t.Run("GetAllServices", func(t *testing.T) {
			setSpec := getTestServiceSetSpec()
			require.NoError(t, manager.Set(ctx, *setSpec), "Failed to set services")

			info, err := manager.Get(ctx)
			require.NoError(t, err, "Failed to get services")
			require.NotNil(t, info, "Info should not be nil")
			require.Len(t, info.Services, 2, "Should have 2 services")

			// Verify first service
			service1 := findServiceByType(info.Services, "fds-files_server")
			require.NotNil(t, service1, "Files server service should exist")
			assert.Equal(t, "test-files-server", service1.Name)
			assert.Equal(t, "fds-files_server", service1.Type)
			assert.Equal(t, "test-key-1", service1.Key)
			assert.Equal(t, "1.0.0", service1.Version)
			require.Len(t, service1.Nodes, 1, "Should have 1 node")

			node1 := service1.Nodes[0]
			assert.Equal(t, "node1", node1.Name)
			assert.Equal(t, "443", node1.Port)
			assert.Equal(t, "/depot", node1.BaseUrl)
			require.Len(t, node1.Addresses, 1, "Should have 1 address")
			assert.Equal(t, depot.AddressTypeFqdn, node1.Addresses[0].Type)
			assert.Equal(t, "files.example.com", node1.Addresses[0].Value)
			assert.Equal(t, []string{"cert1", "cert2"}, node1.Certificates)

			// Verify second service
			service2 := findServiceByType(info.Services, "fds-image_registry")
			require.NotNil(t, service2, "Image registry service should exist")
			assert.Equal(t, "test-image-registry", service2.Name)
			assert.Equal(t, "fds-image_registry", service2.Type)
			assert.Equal(t, "test-key-2", service2.Key)
			assert.Equal(t, "2.0.0", service2.Version)
			require.Len(t, service2.Nodes, 1, "Should have 1 node")

			node2 := service2.Nodes[0]
			assert.Equal(t, "registry-node", node2.Name)
			assert.Equal(t, "5000", node2.Port)
			assert.Equal(t, "/v2", node2.BaseUrl)
			require.Len(t, node2.Addresses, 2, "Should have 2 addresses")
			assert.Equal(t, depot.AddressTypeIPv4, node2.Addresses[0].Type)
			assert.Equal(t, "192.168.1.100", node2.Addresses[0].Value)
			assert.Equal(t, depot.AddressTypeIPv6, node2.Addresses[1].Type)
			assert.Equal(t, "2001:db8::1", node2.Addresses[1].Value)
			assert.Equal(t, []string{"registry-cert"}, node2.Certificates)
		})

		// Test 4: Get specific service by type
		t.Run("GetServiceSpec", func(t *testing.T) {
			setSpec := getTestServiceSetSpec()
			require.NoError(t, manager.Set(ctx, *setSpec), "Failed to set services")

			// Get files server service
			service, err := manager.GetServiceSpec(ctx, "fds-files_server")
			require.NoError(t, err, "Failed to get files server service")
			require.NotNil(t, service, "Service should not be nil")

			assert.Equal(t, "test-files-server", service.Name)
			assert.Equal(t, "fds-files_server", service.Type)
			assert.Equal(t, "test-key-1", service.Key)
			assert.Equal(t, "1.0.0", service.Version)
			require.Len(t, service.Nodes, 1, "Should have 1 node")

			// Get image registry service
			service2, err := manager.GetServiceSpec(ctx, "fds-image_registry")
			require.NoError(t, err, "Failed to get image registry service")
			require.NotNil(t, service2, "Service should not be nil")

			assert.Equal(t, "test-image-registry", service2.Name)
			assert.Equal(t, "fds-image_registry", service2.Type)
			assert.Equal(t, "test-key-2", service2.Key)
			assert.Equal(t, "2.0.0", service2.Version)
			require.Len(t, service2.Nodes, 1, "Should have 1 node")
		})

		// Test 5: Get non-existent service should return error
		t.Run("GetNonExistentService", func(t *testing.T) {
			service, err := manager.GetServiceSpec(ctx, "non-existent-service")
			assert.Error(t, err, "Should return error for non-existent service")
			assert.Nil(t, service, "Service should be nil for non-existent service")
		})

		// Test 6: Update services configuration (replace existing)
		t.Run("UpdateServices", func(t *testing.T) {
			newSetSpec := depot.SetSpec{
				Services: []depot.Service{
					{
						Name:    "updated-files-server",
						Type:    "fds-files_server",
						Key:     "updated-key-1",
						Version: "2.0.0",
						Nodes: []depot.Node{
							{
								Name: "updated-node",
								Addresses: []depot.Address{
									{
										Type:  depot.AddressTypeFqdn,
										Value: "updated-files.example.com",
									},
								},
								Port:         "8443",
								BaseUrl:      "/depotv2",
								Certificates: []string{"updated-cert"},
							},
						},
					},
				},
			}

			err := manager.Set(ctx, newSetSpec)
			require.NoError(t, err, "Failed to update services")

			// Verify the update
			info, err := manager.Get(ctx)
			require.NoError(t, err, "Failed to get updated services")
			require.Len(t, info.Services, 1, "Should have only 1 service after update")

			service := info.Services[0]
			assert.Equal(t, "updated-files-server", service.Name)
			assert.Equal(t, "updated-key-1", service.Key)
			assert.Equal(t, "2.0.0", service.Version)
			assert.Equal(t, "updated-node", service.Nodes[0].Name)
			assert.Equal(t, "updated-files.example.com", service.Nodes[0].Addresses[0].Value)
		})

		// Test 7: Clear all services
		t.Run("ClearServices", func(t *testing.T) {
			setSpec := getTestServiceSetSpec()
			require.NoError(t, manager.Set(ctx, *setSpec), "Failed to set services")

			emptySetSpec := depot.SetSpec{
				Services: []depot.Service{},
			}

			err := manager.Set(ctx, emptySetSpec)
			require.NoError(t, err, "Failed to clear services")

			// Verify services are cleared
			info, err := manager.Get(ctx)
			require.NoError(t, err, "Failed to get services after clearing")
			assert.Empty(t, info.Services, "Services should be empty after clearing")

			// Verify getting specific service returns error
			service, err := manager.GetServiceSpec(ctx, "fds-files_server")
			assert.Error(t, err, "Should return error after clearing services")
			assert.Nil(t, service, "Service should be nil after clearing")
		})
	})
}

func TestServicesAddressTypes(t *testing.T) {
	simulator.Test(func(ctx context.Context, vc *vim25.Client) {
		rc := rest.NewClient(vc)

		err := rc.Login(ctx, simulator.DefaultLogin)
		require.NoError(t, err, "Failed to login")

		manager := depot.NewManager(rc)

		// Test all address types
		t.Run("AllAddressTypes", func(t *testing.T) {
			setSpec := depot.SetSpec{
				Services: []depot.Service{
					{
						Name:    "multi-address-service",
						Type:    "fds-files_server",
						Key:     "multi-addr-key",
						Version: "1.0.0",
						Nodes: []depot.Node{
							{
								Name: "multi-addr-node",
								Addresses: []depot.Address{
									{
										Type:  depot.AddressTypeFqdn,
										Value: "service.example.com",
									},
									{
										Type:  depot.AddressTypeIPv4,
										Value: "10.0.0.1",
									},
									{
										Type:  depot.AddressTypeIPv6,
										Value: "fe80::1",
									},
								},
								Port:         "443",
								BaseUrl:      "/",
								Certificates: []string{},
							},
						},
					},
				},
			}

			err := manager.Set(ctx, setSpec)
			require.NoError(t, err, "Failed to set service with multiple address types")

			// Verify all address types are preserved
			service, err := manager.GetServiceSpec(ctx, "fds-files_server")
			require.NoError(t, err, "Failed to get service")
			require.Len(t, service.Nodes[0].Addresses, 3, "Should have 3 addresses")

			addresses := service.Nodes[0].Addresses
			assert.Equal(t, depot.AddressTypeFqdn, addresses[0].Type)
			assert.Equal(t, "service.example.com", addresses[0].Value)
			assert.Equal(t, depot.AddressTypeIPv4, addresses[1].Type)
			assert.Equal(t, "10.0.0.1", addresses[1].Value)
			assert.Equal(t, depot.AddressTypeIPv6, addresses[2].Type)
			assert.Equal(t, "fe80::1", addresses[2].Value)
		})
	})
}

// Helper function to find a service by type
func findServiceByType(services []depot.Service, serviceType string) *depot.Service {
	for i := range services {
		if services[i].Type == serviceType {
			return &services[i]
		}
	}
	return nil
}

func getTestServiceSetSpec() *depot.SetSpec {
	return &depot.SetSpec{
		Services: []depot.Service{
			{
				Name:    "test-files-server",
				Type:    "fds-files_server",
				Key:     "test-key-1",
				Version: "1.0.0",
				Nodes: []depot.Node{
					{
						Name: "node1",
						Addresses: []depot.Address{
							{
								Type:  depot.AddressTypeFqdn,
								Value: "files.example.com",
							},
						},
						Port:         "443",
						BaseUrl:      "/depot",
						Certificates: []string{"cert1", "cert2"},
					},
				},
			},
			{
				Name:    "test-image-registry",
				Type:    "fds-image_registry",
				Key:     "test-key-2",
				Version: "2.0.0",
				Nodes: []depot.Node{
					{
						Name: "registry-node",
						Addresses: []depot.Address{
							{
								Type:  depot.AddressTypeIPv4,
								Value: "192.168.1.100",
							},
							{
								Type:  depot.AddressTypeIPv6,
								Value: "2001:db8::1",
							},
						},
						Port:         "5000",
						BaseUrl:      "/v2",
						Certificates: []string{"registry-cert"},
					},
				},
			},
		},
	}
}
