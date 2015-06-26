package gardenrunc

import (
	"fmt"
	"net"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/port_pool"
	"github.com/cloudfoundry/gunk/localip"
	"github.com/docker/libnetwork/iptables"
)

//go:generate counterfeiter . Chain
type Chain interface {
	Forward(action iptables.Action, ip net.IP, port int, proto, dest_addr string, dest_port int) error
}

type NetHandler struct {
	ContainerIP string
	Chain       Chain

	PortPool *port_pool.PortPool
}

func (c *NetHandler) NetIn(hostPort, containerPort uint32) (uint32, uint32, error) {
	externalIP, _ := localip.LocalIP()

	if hostPort == 0 {
		var err error
		if hostPort, err = c.PortPool.Acquire(); err != nil {
			return 0, 0, fmt.Errorf("netin: acquire port from pool: %s", err)
		}
	}

	if containerPort == 0 {
		containerPort = hostPort
	}

	if err := c.Chain.Forward(iptables.Append, net.ParseIP(externalIP), int(hostPort), "tcp", c.ContainerIP, int(containerPort)); err != nil {
		return 0, 0, fmt.Errorf("netin %d to %d: %s", hostPort, containerPort, err)
	}

	return 0, 0, nil
}

func (c *NetHandler) NetOut(netOutRule garden.NetOutRule) error {
	return nil
}
