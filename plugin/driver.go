package dhcp

import (
	"encoding/json"
	"net/http"

	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)



func (driver *driver) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create api.CreateNetworkRequest
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

type InterfaceName struct {
	SrcName   string
	DstName   string
	DstPrefix string
}

func (driver *driver) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create api.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}

	log.Debugf("Create endpoint request: %+v", create)
	log.Debugf("Create endpoint request interface: %+v", create.Interface)

	// create and attach local name to the bridge
	local := "ovstap" + create.EndpointID[:5]
	remote := "tap" + create.EndpointID[:5]

	la := netlink.NewLinkAttrs()
	la.Name = local
	veth := &netlink.Veth{ la, remote }

	if err := netlink.LinkAdd(veth); err != nil {
		log.Errorf("could not create veth pair: %s", err)
	}

	//CreateVethPair(local, remote)

	remoteTap, _ := netlink.LinkByName(remote)


	log.Debugf("remoteTap: %+v", remoteTap)
	mac := remoteTap.Attrs().HardwareAddr.String()
	log.Debugf("converted MacAddress: %s", mac)

	ifResult := &api.EndpointInterface {
		MacAddress: mac,
	}

	// IP addrs comes from libnetwork ipam via user 'docker network' parameters
	resp := &api.CreateEndpointResponse{
		Interface: ifResult,
	}
	log.Debugf("Create endpoint response: %+v", resp)
	log.Debugf("Create endpoint response interface: %+v", resp.Interface)
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

	// create and attach local name to the bridge
	local := "ovstap" + delete.EndpointID[:5]
	remote := "tap" + delete.EndpointID[:5]

	la := netlink.NewLinkAttrs()
	la.Name = local
	veth := &netlink.Veth{ la, remote }

	if err := netlink.LinkDel(veth); err != nil {
		log.Errorf("could not create veth pair: %s", err)
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
	var j api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}

	log.Debugf("Join request: %+v", j)

	local := "ovstap" + j.EndpointID[:5]
	AddLinkToBridge(local)

	remote := "tap" + j.EndpointID[:5]

	// SrcName gets renamed to DstPrefix on the container iface
	ifname := &InterfaceName{
		SrcName: remote,
		DstPrefix: "eth",
	}
	res := &joinResponse{
		InterfaceName: *ifname,
		Gateway:       "192.168.1.1",
	}
	log.Debugf("Join response: %+v", res)
	objectResponse(w, res)
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
	local := "ovstap" + l.EndpointID[:5]
	DelLinkFromBridge(local)

	log.Debugf("Leave request: %+v", &l)
	emptyResponse(w)
	log.Debugf("Leave %s:%s", l.NetworkID, l.EndpointID)
}

