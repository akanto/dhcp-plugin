package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/akanto/dhcp-plugin/plugin"
)

const (
	version = "0.0.1"
)

func main() {

	var (
		printVersion bool
	    networkInterface string
		socketAddress string
	)

	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&networkInterface, "networkInterface", "eth0", "network interface to be connected to the ")
	flag.StringVar(&socketAddress, "socket", "/run/docker/plugins/dhcp.sock", "socket on which to listen")

	flag.Parse()
	// Only log the debug severity or above.
	log.SetLevel(log.DebugLevel)

	if printVersion {
		fmt.Printf("DHCP plugin %s\n", version)
		os.Exit(0)
	}

	// remove abandoned socket
	if err := os.Remove(socketAddress); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Unable to create driver: %s", err)
	}

	log.Debugf("DHCP plugin: %s,  address: %s, networkInterface: %s ", version, socketAddress, networkInterface)

	var d dhcp.Driver

	d, err := dhcp.New(version, networkInterface)
	if err != nil {
		log.Fatalf("Unable to create driver: %s", err)
	}

	if err := d.Listen(socketAddress); err != nil {
		log.Fatal(err)
	}
	log.Debugf("DHCP plugin initialised")

}

