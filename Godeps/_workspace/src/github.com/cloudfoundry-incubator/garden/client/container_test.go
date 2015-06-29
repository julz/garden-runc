package client_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/garden"
	. "github.com/cloudfoundry-incubator/garden/client"
	"github.com/cloudfoundry-incubator/garden/client/connection/fakes"
	wfakes "github.com/cloudfoundry-incubator/garden/fakes"
)

var _ = Describe("Container", func() {
	var container garden.Container

	var fakeConnection *fakes.FakeConnection

	BeforeEach(func() {
		fakeConnection = new(fakes.FakeConnection)
	})

	JustBeforeEach(func() {
		var err error

		client := New(fakeConnection)

		fakeConnection.CreateReturns("some-handle", nil)

		container, err = client.Create(garden.ContainerSpec{})
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("Handle", func() {
		It("returns the container's handle", func() {
			Ω(container.Handle()).Should(Equal("some-handle"))
		})
	})

	Describe("Stop", func() {
		It("sends a stop request", func() {
			err := container.Stop(true)
			Ω(err).ShouldNot(HaveOccurred())

			handle, kill := fakeConnection.StopArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(kill).Should(BeTrue())
		})

		Context("when stopping fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.StopReturns(disaster)
			})

			It("returns the error", func() {
				err := container.Stop(true)
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("Info", func() {
		It("sends an info request", func() {
			infoToReturn := garden.ContainerInfo{
				State: "chillin",
			}

			fakeConnection.InfoReturns(infoToReturn, nil)

			info, err := container.Info()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeConnection.InfoArgsForCall(0)).Should(Equal("some-handle"))

			Ω(info).Should(Equal(infoToReturn))
		})

		Context("when getting info fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.InfoReturns(garden.ContainerInfo{}, disaster)
			})

			It("returns the error", func() {
				_, err := container.Info()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("GetProperties", func() {
		Context("when getting properties succeeds", func() {
			BeforeEach(func() {
				fakeConnection.GetPropertiesReturns(garden.Properties{"Foo": "bar"}, nil)
			})

			It("returns the error", func() {
				result, err := container.GetProperties()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(Equal(garden.Properties{"Foo": "bar"}))
			})
		})

		Context("when getting properties fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.GetPropertiesReturns(nil, disaster)
			})

			It("returns the error", func() {
				_, err := container.GetProperties()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("StreamIn", func() {
		It("sends a stream in request", func() {
			fakeConnection.StreamInStub = func(handle string, dst string, reader io.Reader) error {
				Ω(dst).Should(Equal("to"))

				content, err := ioutil.ReadAll(reader)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(content)).Should(Equal("stuff"))

				return nil
			}

			err := container.StreamIn("to", bytes.NewBufferString("stuff"))
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when streaming in fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.StreamInReturns(
					disaster)
			})

			It("returns the error", func() {
				err := container.StreamIn("to", nil)
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("StreamOut", func() {
		It("sends a stream out request", func() {
			fakeConnection.StreamOutReturns(ioutil.NopCloser(strings.NewReader("kewl")), nil)

			reader, err := container.StreamOut("from")
			bytes, err := ioutil.ReadAll(reader)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(bytes)).Should(Equal("kewl"))

			handle, src := fakeConnection.StreamOutArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(src).Should(Equal("from"))
		})

		Context("when streaming out fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.StreamOutReturns(nil, disaster)
			})

			It("returns the error", func() {
				_, err := container.StreamOut("from")
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("LimitBandwidth", func() {
		It("sends a limit bandwidth request", func() {
			err := container.LimitBandwidth(garden.BandwidthLimits{
				RateInBytesPerSecond: 1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			handle, limits := fakeConnection.LimitBandwidthArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(limits).Should(Equal(garden.BandwidthLimits{RateInBytesPerSecond: 1}))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.LimitBandwidthReturns(garden.BandwidthLimits{}, disaster)
			})

			It("returns the error", func() {
				err := container.LimitBandwidth(garden.BandwidthLimits{})
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("LimitCPU", func() {
		It("sends a limit cpu request", func() {
			err := container.LimitCPU(garden.CPULimits{
				LimitInShares: 1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			handle, limits := fakeConnection.LimitCPUArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(limits).Should(Equal(garden.CPULimits{LimitInShares: 1}))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.LimitCPUReturns(garden.CPULimits{}, disaster)
			})

			It("returns the error", func() {
				err := container.LimitCPU(garden.CPULimits{})
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("LimitDisk", func() {
		It("sends a limit bandwidth request", func() {
			err := container.LimitDisk(garden.DiskLimits{
				ByteHard: 1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			handle, limits := fakeConnection.LimitDiskArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(limits).Should(Equal(garden.DiskLimits{ByteHard: 1}))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.LimitDiskReturns(garden.DiskLimits{}, disaster)
			})

			It("returns the error", func() {
				err := container.LimitDisk(garden.DiskLimits{})
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("LimitMemory", func() {
		It("sends a limit bandwidth request", func() {
			err := container.LimitMemory(garden.MemoryLimits{
				LimitInBytes: 1,
			})
			Ω(err).ShouldNot(HaveOccurred())

			handle, limits := fakeConnection.LimitMemoryArgsForCall(0)
			Ω(handle).Should(Equal("some-handle"))
			Ω(limits).Should(Equal(garden.MemoryLimits{LimitInBytes: 1}))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.LimitMemoryReturns(garden.MemoryLimits{}, disaster)
			})

			It("returns the error", func() {
				err := container.LimitMemory(garden.MemoryLimits{})
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("CurrentBandwidthLimits", func() {
		It("sends an empty limit request and returns its response", func() {
			limitsToReturn := garden.BandwidthLimits{
				RateInBytesPerSecond:      1,
				BurstRateInBytesPerSecond: 2,
			}

			fakeConnection.CurrentBandwidthLimitsReturns(limitsToReturn, nil)

			limits, err := container.CurrentBandwidthLimits()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(limits).Should(Equal(limitsToReturn))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.CurrentBandwidthLimitsReturns(garden.BandwidthLimits{}, disaster)
			})

			It("returns the error", func() {
				_, err := container.CurrentBandwidthLimits()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("CurrentCPULimits", func() {
		It("sends an empty limit request and returns its response", func() {
			limitsToReturn := garden.CPULimits{
				LimitInShares: 1,
			}

			fakeConnection.CurrentCPULimitsReturns(limitsToReturn, nil)

			limits, err := container.CurrentCPULimits()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(limits).Should(Equal(limitsToReturn))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.CurrentCPULimitsReturns(garden.CPULimits{}, disaster)
			})

			It("returns the error", func() {
				_, err := container.CurrentCPULimits()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("CurrentDiskLimits", func() {
		It("sends an empty limit request and returns its response", func() {
			limitsToReturn := garden.DiskLimits{
				BlockSoft: 3,
				BlockHard: 4,
				InodeSoft: 7,
				InodeHard: 8,
				ByteSoft:  11,
				ByteHard:  12,
			}

			fakeConnection.CurrentDiskLimitsReturns(limitsToReturn, nil)

			limits, err := container.CurrentDiskLimits()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(limits).Should(Equal(limitsToReturn))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.CurrentDiskLimitsReturns(garden.DiskLimits{}, disaster)
			})

			It("returns the error", func() {
				_, err := container.CurrentDiskLimits()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("CurrentMemoryLimits", func() {
		It("gets the current limits", func() {
			limitsToReturn := garden.MemoryLimits{
				LimitInBytes: 1,
			}

			fakeConnection.CurrentMemoryLimitsReturns(limitsToReturn, nil)

			limits, err := container.CurrentMemoryLimits()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(limits).Should(Equal(limitsToReturn))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.CurrentMemoryLimitsReturns(garden.MemoryLimits{}, disaster)
			})

			It("returns the error", func() {
				_, err := container.CurrentMemoryLimits()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("Run", func() {
		It("sends a run request and returns the process id and a stream", func() {
			fakeConnection.RunStub = func(handle string, spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
				process := new(wfakes.FakeProcess)

				process.IDReturns(42)
				process.WaitReturns(123, nil)

				go func() {
					defer GinkgoRecover()

					_, err := fmt.Fprintf(io.Stdout, "stdout data")
					Ω(err).ShouldNot(HaveOccurred())

					_, err = fmt.Fprintf(io.Stderr, "stderr data")
					Ω(err).ShouldNot(HaveOccurred())
				}()

				return process, nil
			}

			spec := garden.ProcessSpec{
				Path: "some-script",
			}

			stdout := gbytes.NewBuffer()
			stderr := gbytes.NewBuffer()

			processIO := garden.ProcessIO{
				Stdout: stdout,
				Stderr: stderr,
			}

			process, err := container.Run(spec, processIO)
			Ω(err).ShouldNot(HaveOccurred())

			ranHandle, ranSpec, ranIO := fakeConnection.RunArgsForCall(0)
			Ω(ranHandle).Should(Equal("some-handle"))
			Ω(ranSpec).Should(Equal(spec))
			Ω(ranIO).Should(Equal(processIO))

			Ω(process.ID()).Should(Equal(uint32(42)))

			status, err := process.Wait()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(status).Should(Equal(123))

			Eventually(stdout).Should(gbytes.Say("stdout data"))
			Eventually(stderr).Should(gbytes.Say("stderr data"))
		})
	})

	Describe("Attach", func() {
		It("sends an attach request and returns a stream", func() {
			fakeConnection.AttachStub = func(handle string, processID uint32, io garden.ProcessIO) (garden.Process, error) {
				process := new(wfakes.FakeProcess)

				process.IDReturns(42)
				process.WaitReturns(123, nil)

				go func() {
					defer GinkgoRecover()

					_, err := fmt.Fprintf(io.Stdout, "stdout data")
					Ω(err).ShouldNot(HaveOccurred())

					_, err = fmt.Fprintf(io.Stderr, "stderr data")
					Ω(err).ShouldNot(HaveOccurred())
				}()

				return process, nil
			}

			stdout := gbytes.NewBuffer()
			stderr := gbytes.NewBuffer()

			processIO := garden.ProcessIO{
				Stdout: stdout,
				Stderr: stderr,
			}

			process, err := container.Attach(42, processIO)
			Ω(err).ShouldNot(HaveOccurred())

			attachedHandle, attachedID, attachedIO := fakeConnection.AttachArgsForCall(0)
			Ω(attachedHandle).Should(Equal("some-handle"))
			Ω(attachedID).Should(Equal(uint32(42)))
			Ω(attachedIO).Should(Equal(processIO))

			Ω(process.ID()).Should(Equal(uint32(42)))

			status, err := process.Wait()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(status).Should(Equal(123))

			Eventually(stdout).Should(gbytes.Say("stdout data"))
			Eventually(stderr).Should(gbytes.Say("stderr data"))
		})
	})

	Describe("NetIn", func() {
		It("sends a net in request", func() {
			fakeConnection.NetInReturns(111, 222, nil)

			hostPort, containerPort, err := container.NetIn(123, 456)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(hostPort).Should(Equal(uint32(111)))
			Ω(containerPort).Should(Equal(uint32(222)))

			h, hp, cp := fakeConnection.NetInArgsForCall(0)
			Ω(h).Should(Equal("some-handle"))
			Ω(hp).Should(Equal(uint32(123)))
			Ω(cp).Should(Equal(uint32(456)))
		})

		Context("when the request fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeConnection.NetInReturns(0, 0, disaster)
			})

			It("returns the error", func() {
				_, _, err := container.NetIn(123, 456)
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("NetOut", func() {
		It("sends NetOut requests over the connection", func() {
			Ω(container.NetOut(garden.NetOutRule{
				Networks: []garden.IPRange{garden.IPRangeFromIP(net.ParseIP("1.2.3.4"))},
				Ports: []garden.PortRange{
					{Start: 12, End: 24},
				},
				Log: true,
			})).Should(Succeed())

			h, rule := fakeConnection.NetOutArgsForCall(0)
			Ω(h).Should(Equal("some-handle"))

			Ω(rule.Networks).Should(HaveLen(1))
			Ω(rule.Networks[0]).Should(Equal(garden.IPRange{Start: net.ParseIP("1.2.3.4"), End: net.ParseIP("1.2.3.4")}))

			Ω(rule.Ports).Should(HaveLen(1))
			Ω(rule.Ports[0]).Should(Equal(garden.PortRange{Start: 12, End: 24}))

			Ω(rule.Log).Should(Equal(true))
		})
	})

	Context("when the request fails", func() {
		disaster := errors.New("oh no!")

		BeforeEach(func() {
			fakeConnection.NetOutReturns(disaster)
		})

		It("returns the error", func() {
			err := container.NetOut(garden.NetOutRule{})
			Ω(err).Should(Equal(disaster))
		})
	})
})
