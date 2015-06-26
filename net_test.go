package gardenrunc_test

import (
	"github.com/cloudfoundry-incubator/garden-linux/port_pool"
	. "github.com/julz/garden-runc"
	"github.com/julz/garden-runc/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Net", func() {
	var container *NetHandler
	var fakeChain *fakes.FakeChain

	BeforeEach(func() {
		fakeChain = new(fakes.FakeChain)
		container = &NetHandler{
			Chain:    fakeChain,
			PortPool: port_pool.New(10, 3),
		}
	})

	Describe("NetIn", func() {
		It("forwards ports using iptables", func() {
			container.NetIn(123, 456)
			Expect(fakeChain.ForwardCallCount()).Should(Equal(1))
		})

		Context("when the host port is 0", func() {
			It("selects a unique host port", func() {
				container.NetIn(0, 456)
				container.NetIn(0, 456)
				Expect(fakeChain.ForwardCallCount()).Should(Equal(2))

				_, _, hostPort1, _, _, _ := fakeChain.ForwardArgsForCall(0)
				_, _, hostPort2, _, _, _ := fakeChain.ForwardArgsForCall(1)
				Expect(hostPort1).NotTo(Equal(hostPort2))
			})

			Context("when the container port is 0", func() {
				It("uses the host port", func() {
					container.NetIn(0, 0)
					Expect(fakeChain.ForwardCallCount()).Should(Equal(1))

					_, _, hostPort, _, _, containerPort := fakeChain.ForwardArgsForCall(0)
					Expect(hostPort).To(Equal(containerPort))
				})
			})
		})

		Context("when a host port is specifically requested", func() {
			It("is used", func() {
				container.NetIn(123, 456)
				Expect(fakeChain.ForwardCallCount()).Should(Equal(1))
				_, _, hostPort1, _, _, _ := fakeChain.ForwardArgsForCall(0)
				Expect(hostPort1).To(Equal(123))
			})
		})

		Context("when the port pool is dry", func() {
			It("returns an error", func() {
				for i := 0; i < 3; i++ {
					_, _, err := container.NetIn(0, 456)
					Expect(err).NotTo(HaveOccurred())
				}

				_, _, err := container.NetIn(0, 456)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
