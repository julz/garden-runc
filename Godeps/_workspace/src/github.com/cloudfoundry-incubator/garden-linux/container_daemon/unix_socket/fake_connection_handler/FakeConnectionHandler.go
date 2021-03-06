// This file was generated by counterfeiter
package fake_connection_handler

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/cloudfoundry-incubator/garden-linux/container_daemon/unix_socket"
)

type FakeConnectionHandler struct {
	HandleStub        func(decoder *json.Decoder) ([]*os.File, int, error)
	handleMutex       sync.RWMutex
	handleArgsForCall []struct {
		decoder *json.Decoder
	}
	handleReturns struct {
		result1 []*os.File
		result2 int
		result3 error
	}
}

func (fake *FakeConnectionHandler) Handle(decoder *json.Decoder) ([]*os.File, int, error) {
	fake.handleMutex.Lock()
	fake.handleArgsForCall = append(fake.handleArgsForCall, struct {
		decoder *json.Decoder
	}{decoder})
	fake.handleMutex.Unlock()
	if fake.HandleStub != nil {
		return fake.HandleStub(decoder)
	} else {
		return fake.handleReturns.result1, fake.handleReturns.result2, fake.handleReturns.result3
	}
}

func (fake *FakeConnectionHandler) HandleCallCount() int {
	fake.handleMutex.RLock()
	defer fake.handleMutex.RUnlock()
	return len(fake.handleArgsForCall)
}

func (fake *FakeConnectionHandler) HandleArgsForCall(i int) *json.Decoder {
	fake.handleMutex.RLock()
	defer fake.handleMutex.RUnlock()
	return fake.handleArgsForCall[i].decoder
}

func (fake *FakeConnectionHandler) HandleReturns(result1 []*os.File, result2 int, result3 error) {
	fake.HandleStub = nil
	fake.handleReturns = struct {
		result1 []*os.File
		result2 int
		result3 error
	}{result1, result2, result3}
}

var _ unix_socket.ConnectionHandler = new(FakeConnectionHandler)
