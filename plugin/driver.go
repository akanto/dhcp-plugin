package dhcp

import (
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)


type driver struct {
	version    string
	externalPort string
}

func NewDriver(version string, externalPort string) (Driver, error) {
	b := NewBridge()
	err := b.setupBridge(externalPort)

	if err != nil {
		return nil, fmt.Errorf("unable to create the bridge: %s", err)
	}

	d := &driver{
		version: version,
		externalPort: externalPort,
	}
	return d, nil
}


func (driver *driver) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create api.CreateNetworkRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("received create network request: %+v", create)

	resp := &api.CreateNetworkResponse{}

	objectResponse(w, resp)
}


func (driver *driver) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: " + err.Error(), http.StatusBadRequest)
		return
	}
	log.Debugf("Delete network request:  %+v", delete)
	emptyResponse(w)
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
	local, remote := veth(create.EndpointID)

	la := netlink.NewLinkAttrs()
	la.Name = local
	veth := &netlink.Veth{ la, remote }

	if err := netlink.LinkAdd(veth); err != nil {
		log.Errorf("could not create veth pair: %s", err)
	}

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

func (driver *driver) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}

	// create and attach local name to the bridge
	local, remote := veth(delete.EndpointID)

	la := netlink.NewLinkAttrs()
	la.Name = local
	veth := &netlink.Veth{ la, remote }

	if err := netlink.LinkDel(veth); err != nil {
		log.Errorf("could not create veth pair: %s", err)
	}

	log.Debugf("Delete endpoint request: %+v", &delete)
	emptyResponse(w)

	log.Debugf("Delete endpoint %s", delete.EndpointID)


}

func (driver *driver) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var info api.EndpointInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	log.Debugf("Endpoint info request: %+v", &info)
	objectResponse(w, &api.EndpointInfoResponse{Value: map[string]interface{}{}})
	log.Debugf("Endpoint info %s", info.EndpointID)
}

func (driver *driver) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var j api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}

	log.Debugf("Join request: %+v", j)

	local, remote := veth(j.EndpointID)

	b := NewBridge()
	b.addLink(local)

	// SrcName gets renamed to DstPrefix on the container iface
	ifname := &api.InterfaceName{
		SrcName: remote,
		DstPrefix: "eth",
	}

	resp := &api.JoinResponse{
		InterfaceName: ifname,
		Gateway: "192.168.10.1",
	}

	log.Debugf("Join response: %+v", resp)
	objectResponse(w, resp)
	log.Debugf("Join endpoint %s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey)
}

func (driver *driver) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l api.LeaveRequest
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


func veth(endpointId string) (local string, remote string) {
	suffix := endpointId[:5]
	local = "ovstap" + suffix
	remote = "tap" + suffix
	return
}
