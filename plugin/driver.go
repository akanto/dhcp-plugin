package dhcp

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/docker/libnetwork/driverapi"

	log "github.com/Sirupsen/logrus"
)


type networkCreate struct {
	NetworkID string
	Options   map[string]interface{}
	IpV4Data  []driverapi.IPAMData
	ipV6Data  []driverapi.IPAMData
}

func (driver *driver) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create networkCreate
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("received create network request: %+v", create)

	emptyResponse(w)
}

type networkDelete struct {
	NetworkID string
}

func (driver *driver) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete networkDelete
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("Delete network request:  %+v", delete)
	emptyResponse(w)
}

type endpointCreate struct {
	NetworkID  string
	EndpointID string
	Interface  *EndpointInterface
	Options    map[string]interface{}
}

// EndpointInterface represents an interface endpoint.
type EndpointInterface struct {
	Address     string
	AddressIPv6 string
	MacAddress  string
}

type InterfaceName struct {
	SrcName   string
	DstName   string
	DstPrefix string
}

type endpointResponse struct {
	Interface EndpointInterface
}

func (driver *driver) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create endpointCreate
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}

	log.Debugf("Create endpoint request: %+v", create)

	// IP addrs comes from libnetwork ipam via user 'docker network' parameters
	respIface := &EndpointInterface{
		MacAddress: create.Interface.MacAddress,
	}
	resp := &endpointResponse{
		Interface: *respIface,
	}
	log.Debugf("Create endpoint response: %+v", resp)
	objectResponse(w, resp)
}

type endpointDelete struct {
	NetworkID  string
	EndpointID string
}

func (driver *driver) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete endpointDelete
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	log.Debugf("Delete endpoint request: %+v", &delete)
	emptyResponse(w)
	// null check cidr in case driver restarted and doesn't know the network to avoid panic
	if driver.cidr == nil {
		return
	}

	log.Debugf("Delete endpoint %s", delete.EndpointID)


}

type endpointInfoReq struct {
	NetworkID  string
	EndpointID string
}

type endpointInfo struct {
	Value map[string]interface{}
}

func (driver *driver) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var info endpointInfoReq
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	log.Debugf("Endpoint info request: %+v", &info)
	objectResponse(w, &endpointInfo{Value: map[string]interface{}{}})
	log.Debugf("Endpoint info %s", info.EndpointID)
}

type joinInfo struct {
	InterfaceName *InterfaceName
	Gateway       string
	GatewayIPv6   string
}

type join struct {
	NetworkID  string
	EndpointID string
	SandboxKey string
	Options    map[string]interface{}
}

type staticRoute struct {
	Destination string
	RouteType   int
	NextHop     string
}

type joinResponse struct {
	Gateway       string
	InterfaceName InterfaceName
	StaticRoutes  []*staticRoute
}

func (driver *driver) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var j join
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}

	log.Debugf("Join endpoint %s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey)
}

type leave struct {
	NetworkID  string
	EndpointID string
	Options    map[string]interface{}
}

func (driver *driver) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l leave
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	log.Debugf("Leave request: %+v", &l)
	emptyResponse(w)
	log.Debugf("Leave %s:%s", l.NetworkID, l.EndpointID)
}


// Generate a mac addr
func makeMac(ip net.IP) string {
	hw := make(net.HardwareAddr, 6)
	hw[0] = 0x7a
	hw[1] = 0x42
	copy(hw[2:], ip.To4())
	return hw.String()
}

