package gardenrunc

import "github.com/cloudfoundry-incubator/garden"

type InfoHandler struct {
	Spec garden.ContainerSpec

	HostIP        string
	ContainerIP   string
	ContainerPath string
	DockerID      string

	*PropsHandler
}

func (i *InfoHandler) Handle() string {
	return i.Spec.Handle
}

func (i *InfoHandler) Info() (garden.ContainerInfo, error) {
	return garden.ContainerInfo{
		State:         "active",
		Events:        []string{},
		HostIP:        i.HostIP,
		ContainerIP:   i.ContainerIP,
		ContainerPath: i.ContainerPath,
		ProcessIDs:    []uint32{},
		Properties:    i.PropsHandler.properties(),
		MappedPorts:   []garden.PortMapping{},
	}, nil
}
