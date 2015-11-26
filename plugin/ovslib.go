package dhcp

import (
	"os/exec"

	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)

type Bridge struct  {
	bridgeName string

}

func NewBridge() (*Bridge) {
	b := &Bridge{
		bridgeName: "floatingbr",
	}
	return b
}

func (b *Bridge) setupBridge(externalPort string) error {

	la := netlink.NewLinkAttrs()
	la.Name = b.bridgeName
	bridge, _ := netlink.LinkByName(b.bridgeName)

	if bridge == nil {
		log.Debugf("Bridge %s does not exist ", b.bridgeName)
		out, err := exec.Command("ovs-vsctl", "add-br", b.bridgeName).CombinedOutput()
		if err != nil {
			log.Fatalf("Bridge %s creation failed been created.  Resp: %s, err: %s", b.bridgeName, out, err)
		}
		log.Infof("Bridge %s has been created.  Resp: %s", b.bridgeName, out)

		out, err = exec.Command("ovs-vsctl", "add-port", b.bridgeName, externalPort).CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to add external port %s.  Resp: %s, err: %s", externalPort, out, err)
		}
		log.Infof("External port %s has been added to %s. Resp: %s", externalPort, b.bridgeName, out)

		out, err = exec.Command("ifconfig", externalPort, "0.0.0.0").CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to ip address of port %s. Resp: %s, err: %s", externalPort, out, err)
		}
		log.Infof("Ip address of port  %s has been cleaned. Resp: %s", externalPort, out)

		return err
	} else {
		log.Debugf("Bridge %s already exsist", b.bridgeName)
	}

	return nil
}

func (b *Bridge) addLink(local string) error {
	out, err := exec.Command("ovs-vsctl", "add-port", b.bridgeName, local).CombinedOutput()
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

