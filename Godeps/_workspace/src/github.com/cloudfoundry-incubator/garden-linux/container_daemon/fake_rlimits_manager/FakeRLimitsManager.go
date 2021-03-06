// This file was generated by counterfeiter
package fake_rlimits_manager

import (
	"sync"
	"syscall"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/container_daemon"
)

type FakeResourceLimitsManager struct {
	ApplyStub        func(rLimits garden.ResourceLimits) error
	applyMutex       sync.RWMutex
	applyArgsForCall []struct {
		rLimits garden.ResourceLimits
	}
	applyReturns struct {
		result1 error
	}
	RestoreStub        func() error
	restoreMutex       sync.RWMutex
	restoreArgsForCall []struct{}
	restoreReturns     struct {
		result1 error
	}
	PreviousRLimitValueStub        func(rLimitId int) *syscall.Rlimit
	previousRLimitValueMutex       sync.RWMutex
	previousRLimitValueArgsForCall []struct {
		rLimitId int
	}
	previousRLimitValueReturns struct {
		result1 *syscall.Rlimit
	}
}

func (fake *FakeResourceLimitsManager) Apply(rLimits garden.ResourceLimits) error {
	fake.applyMutex.Lock()
	fake.applyArgsForCall = append(fake.applyArgsForCall, struct {
		rLimits garden.ResourceLimits
	}{rLimits})
	fake.applyMutex.Unlock()
	if fake.ApplyStub != nil {
		return fake.ApplyStub(rLimits)
	} else {
		return fake.applyReturns.result1
	}
}

func (fake *FakeResourceLimitsManager) ApplyCallCount() int {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return len(fake.applyArgsForCall)
}

func (fake *FakeResourceLimitsManager) ApplyArgsForCall(i int) garden.ResourceLimits {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return fake.applyArgsForCall[i].rLimits
}

func (fake *FakeResourceLimitsManager) ApplyReturns(result1 error) {
	fake.ApplyStub = nil
	fake.applyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeResourceLimitsManager) Restore() error {
	fake.restoreMutex.Lock()
	fake.restoreArgsForCall = append(fake.restoreArgsForCall, struct{}{})
	fake.restoreMutex.Unlock()
	if fake.RestoreStub != nil {
		return fake.RestoreStub()
	} else {
		return fake.restoreReturns.result1
	}
}

func (fake *FakeResourceLimitsManager) RestoreCallCount() int {
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	return len(fake.restoreArgsForCall)
}

func (fake *FakeResourceLimitsManager) RestoreReturns(result1 error) {
	fake.RestoreStub = nil
	fake.restoreReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeResourceLimitsManager) PreviousRLimitValue(rLimitId int) *syscall.Rlimit {
	fake.previousRLimitValueMutex.Lock()
	fake.previousRLimitValueArgsForCall = append(fake.previousRLimitValueArgsForCall, struct {
		rLimitId int
	}{rLimitId})
	fake.previousRLimitValueMutex.Unlock()
	if fake.PreviousRLimitValueStub != nil {
		return fake.PreviousRLimitValueStub(rLimitId)
	} else {
		return fake.previousRLimitValueReturns.result1
	}
}

func (fake *FakeResourceLimitsManager) PreviousRLimitValueCallCount() int {
	fake.previousRLimitValueMutex.RLock()
	defer fake.previousRLimitValueMutex.RUnlock()
	return len(fake.previousRLimitValueArgsForCall)
}

func (fake *FakeResourceLimitsManager) PreviousRLimitValueArgsForCall(i int) int {
	fake.previousRLimitValueMutex.RLock()
	defer fake.previousRLimitValueMutex.RUnlock()
	return fake.previousRLimitValueArgsForCall[i].rLimitId
}

func (fake *FakeResourceLimitsManager) PreviousRLimitValueReturns(result1 *syscall.Rlimit) {
	fake.PreviousRLimitValueStub = nil
	fake.previousRLimitValueReturns = struct {
		result1 *syscall.Rlimit
	}{result1}
}

var _ container_daemon.ResourceLimitsManager = new(FakeResourceLimitsManager)
