// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package simulator

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/govmomi/simulator"
	vapi "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vapi/vcenter/lcm/depot"
)

const (
	servicesPath = "/api/vcenter/lcm/depot/services"
)

func init() {
	simulator.RegisterEndpoint(func(s *simulator.Service, r *simulator.Registry) {
		New(s.Listen).Register(s, r)
	})
}

// Handler implements the LCM Depot Services API simulator.
type Handler struct {
	URL          *url.URL
	servicesData map[string]*depot.Service
}

// New creates a Handler instance.
func New(u *url.URL) *Handler {
	return &Handler{
		URL:          u,
		servicesData: make(map[string]*depot.Service),
	}
}

// Register LCM Depot Services API paths with the vapi simulator's http.ServeMux.
func (h *Handler) Register(s *simulator.Service, r *simulator.Registry) {
	if r.IsVPX() {
		s.HandleFunc(servicesPath, h.services)
		s.HandleFunc(servicesPath+"/", h.servicesWithType)
	}
}

// services handles the main services endpoint.
func (h *Handler) services(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return all configured services
		services := make([]depot.Service, 0, len(h.servicesData))
		for _, service := range h.servicesData {
			services = append(services, *service)
		}

		info := depot.Info{
			Services: services,
		}
		vapi.StatusOK(w, info)

	case http.MethodPut:
		// Set/configure services
		var setSpec depot.SetSpec
		if !vapi.Decode(r, w, &setSpec) {
			return
		}

		// Clear existing services and add new ones
		h.servicesData = make(map[string]*depot.Service)
		for i := range setSpec.Services {
			service := &setSpec.Services[i]
			h.servicesData[service.Type] = service
		}

		vapi.StatusOK(w)

	default:
		vapi.ApiErrorNotFound(w)
	}
}

// servicesWithType handles requests to specific service types
func (h *Handler) servicesWithType(w http.ResponseWriter, r *http.Request) {
	// Extract service type from path
	subpath := r.URL.Path[len(servicesPath):]
	serviceType := strings.Trim(subpath, "/")

	if serviceType == "" {
		vapi.ApiErrorNotFound(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get specific service configuration
		if service, exists := h.servicesData[serviceType]; exists {
			vapi.StatusOK(w, *service)
		} else {
			vapi.ApiErrorNotFound(w)
		}

	default:
		vapi.ApiErrorNotFound(w)
	}
}
