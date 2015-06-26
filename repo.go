package gardenrunc

import (
	"sync"

	"github.com/cloudfoundry-incubator/garden"
)

type repo struct {
	store map[string]*Container
	mutex *sync.RWMutex
}

func NewRepo() *repo {
	return &repo{
		store: map[string]*Container{},
		mutex: &sync.RWMutex{},
	}
}

func (cr *repo) All() []*Container {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	return cr.Query(func(c *Container) bool {
		return true
	})
}

func (cr *repo) Add(container *Container) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	cr.store[container.Handle()] = container
}

func (cr *repo) FindByHandle(handle string) (*Container, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	container, ok := cr.store[handle]
	if !ok {
		return nil, garden.ContainerNotFoundError{handle}
	}

	return container, nil
}

func (cr *repo) Delete(container *Container) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	delete(cr.store, container.Handle())
}

func (cr *repo) Query(filter func(*Container) bool) []*Container {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	var matches []*Container
	for _, c := range cr.store {
		if filter(c) {
			matches = append(matches, c)
		}
	}

	return matches
}
