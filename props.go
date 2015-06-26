package gardenrunc

import (
	"sync"

	"github.com/cloudfoundry-incubator/garden"
)

type PropsHandler struct {
	mu    sync.RWMutex
	props map[string]string
}

func (c *PropsHandler) GetProperties() (garden.Properties, error) {
	return c.properties(), nil
}

func (c *PropsHandler) properties() garden.Properties {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.props
}

func (c *PropsHandler) GetProperty(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.props[name], nil
}

func (c *PropsHandler) SetProperty(name string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.props[name] = value
	return nil
}

func (c *PropsHandler) RemoveProperty(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.props, name)
	return nil
}

func (c *PropsHandler) HasProperties(props garden.Properties) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, v := range props {
		if prop, ok := c.props[k]; !ok || prop != v {
			return false
		}
	}

	return true
}
