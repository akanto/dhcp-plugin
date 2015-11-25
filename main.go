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


// go run main.go --externalPort=enp0s9
// docker network create -d dhcp --subnet=192.168.1.1/28 FLOATING
// docker network rm FLOATING
// docker run -it --net FLOATING akanto/net-perf

func main() {

	var (
		printVersion bool
		externalPort string
		socketAddress string
	)

	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&externalPort, "externalPort", "", "network interface to be connected to the ")
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

	// remove abandoned socket
	if externalPort == "" {
		log.Fatalf("Flag --externalPort is not provided")
	}

	log.Debugf("DHCP plugin: %s,  address: %s, externalPort: %s ", version, socketAddress, externalPort)

	var d dhcp.Driver

	d, err := dhcp.New(version, externalPort)
	if err != nil {
		log.Fatalf("Unable to create driver: %s", err)
	}

	if err := d.Listen(socketAddress); err != nil {
		log.Fatal(err)
	}
	log.Debugf("DHCP plugin initialised")

}

