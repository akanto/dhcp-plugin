package dhcp

import (
	"os/exec"

	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)

func SetupBridge(externalPort string) error {
	bridgeName := "floatingbr"
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	bridge, _ := netlink.LinkByName(bridgeName)

	if bridge == nil {
		log.Debugf("Bridge %s does not exist ", bridgeName)
		out, err := exec.Command("ovs-vsctl", "add-br", bridgeName).CombinedOutput()
		if err != nil {
			log.Fatalf("Bridge %s creation failed been created.  Resp: %s, err: %s", bridgeName, out, err)
		}
		log.Infof("Bridge %s has been created.  Resp: %s", bridgeName, out)

		out, err = exec.Command("ovs-vsctl", "add-port", bridgeName, externalPort).CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to add external port %s.  Resp: %s, err: %s", externalPort, out, err)
		}
		log.Infof("External port %s has been added to %s. Resp: %s", externalPort, bridgeName, out)

		out, err = exec.Command("ifconfig", externalPort, "0.0.0.0").CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to ip address of port %s. Resp: %s, err: %s", externalPort, out, err)
		}
		log.Infof("Ip address of port  %s has been cleaned. Resp: %s", externalPort, out)

		return err
	} else {
		log.Debugf("Bridge %s already exsist", bridgeName)
	}

	return nil
}

func AddLinkToBridge(local string) error {
	bridgeName := "floatingbr"
	out, err := exec.Command("ovs-vsctl", "add-port", bridgeName, local).CombinedOutput()
	if err != nil {
		log.Errorf("Failed to add bridge: %s.  Resp: %s, err: %s",  local, out, err)
		return err
	}
	out, err = exec.Command("ip", "link", "set", "dev", local, "up").CombinedOutput()
	if err != nil {
		log.Errorf("Failed to set u port: %s.  Resp: %s, err: %s",  local, out, err)
		return err
	}
	return err
}

func DelLinkFromBridge(local string) error {
	out, err := exec.Command("ovs-vsctl", "del-port", local).CombinedOutput()
	if err != nil {
		log.Errorf("Failed to delete bridge: %s.  Resp: %s, err: %s",  local, out, err)
	}
	return err
}

func CreateVethPair(local string, remote string) error {
	out, err := exec.Command("ip", "link", "add", remote, "type", "veth", "peer", "name", local).CombinedOutput()
	if err != nil {
		log.Errorf("Veth pair creation failed (%s, %s) creation failed been created.  Resp: %s, err: %s",  local, remote, out, err)
	}
	return err
}