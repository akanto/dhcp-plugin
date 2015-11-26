package dhcp


import (
	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)

func NewBridgex() error {
	bridgeName := "dhcpbr0"
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	bridge, _ := netlink.LinkByName(bridgeName)

	if bridge == nil {
		log.Debugf("Bridge %s does not exist ", bridgeName)
		bridge = &netlink.Bridge{la}
		err := netlink.LinkAdd(bridge)
		if err == nil {
			err = netlink.LinkSetUp(bridge)
			log.Debugf("Bridge %s has been added and enabled ", bridgeName)
		}
		return err
	} else {
		log.Debugf("Bridge %s already exsist", bridgeName)
	}

	return nil
}