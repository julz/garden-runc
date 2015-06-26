// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/julz/garden-runc"
)

type FakeCreator struct {
	CreateStub        func(spec garden.ContainerSpec) (*gardenrunc.Container, error)
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		spec garden.ContainerSpec
	}
	createReturns struct {
		result1 *gardenrunc.Container
		result2 error
	}
}

func (fake *FakeCreator) Create(spec garden.ContainerSpec) (*gardenrunc.Container, error) {
	fake.createMutex.Lock()
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		spec garden.ContainerSpec
	}{spec})
	fake.createMutex.Unlock()
	if fake.CreateStub != nil {
		return fake.CreateStub(spec)
	} else {
		return fake.createReturns.result1, fake.createReturns.result2
	}
}

func (fake *FakeCreator) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeCreator) CreateArgsForCall(i int) garden.ContainerSpec {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return fake.createArgsForCall[i].spec
}

func (fake *FakeCreator) CreateReturns(result1 *gardenrunc.Container, result2 error) {
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 *gardenrunc.Container
		result2 error
	}{result1, result2}
}

var _ gardenrunc.Creator = new(FakeCreator)