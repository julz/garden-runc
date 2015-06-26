package gardenrunc_test

import (
	"os/exec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/process_tracker/fake_process_tracker"
	"github.com/julz/garden-runc"
	"github.com/julz/garden-runc/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("RunHandler", func() {
	var fakeContainerCmder *fakes.FakeContainerCmder
	var fakeProcessTracker *fake_process_tracker.FakeProcessTracker
	var container *gardenrunc.RunHandler

	BeforeEach(func() {
		fakeContainerCmder = new(fakes.FakeContainerCmder)
		fakeProcessTracker = new(fake_process_tracker.FakeProcessTracker)

		container = &gardenrunc.RunHandler{
			ContainerCmd:   fakeContainerCmder,
			ProcessTracker: fakeProcessTracker,
		}
	})

	Describe("Run", func() {
		It("spawns the requested program using iodaemon", func() {
			fakeContainerCmder.CmdStub = func(path string, args ...string) *exec.Cmd {
				return exec.Command("dosh", append([]string{path}, args...)...)
			}

			requestedIO := garden.ProcessIO{Stdout: gbytes.NewBuffer()}
			requestedTTY := &garden.TTYSpec{
				WindowSize: &garden.WindowSize{1, 2},
			}

			container.Run(garden.ProcessSpec{
				Path: "some-path",
				Args: []string{"an-arg", "another-arg"},
				TTY:  requestedTTY,
			}, requestedIO)

			Expect(fakeProcessTracker.RunCallCount()).To(Equal(1))
			_, cmd, io, tty, _ := fakeProcessTracker.RunArgsForCall(0)

			Expect(cmd.Path).To(Equal("dosh"))
			Expect(cmd.Args).To(Equal([]string{"dosh", "some-path", "an-arg", "another-arg"}))
			Expect(io).To(Equal(requestedIO))
			Expect(tty).To(Equal(requestedTTY))
		})

		PIt("requests sequential process ids", func() {})
	})

	Describe("Attach", func() {
		It("attaches to the requested process", func() {
			requestedIO := garden.ProcessIO{Stdout: gbytes.NewBuffer()}
			container.Attach(33, requestedIO)

			Expect(fakeProcessTracker.AttachCallCount()).To(Equal(1))
			id, io := fakeProcessTracker.AttachArgsForCall(0)

			Expect(id).To(Equal(uint32(33)))
			Expect(io).To(Equal(requestedIO))
		})
	})

	PIt("adds a signaller to the spawned process", func() {
	})
})
