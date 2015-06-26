package gardenrunc

import "github.com/cloudfoundry-incubator/garden"

type Container struct {
	*InfoHandler
	*NetHandler
	*RunHandler
	*StreamHandler
	*LimitsHandler
}

func (c *Container) Metrics() (garden.Metrics, error) {
	return garden.Metrics{}, nil
}
