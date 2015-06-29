package process_tracker_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/process_tracker"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
)

var processTracker process_tracker.ProcessTracker
var tmpdir string

var _ = BeforeEach(func() {
	var err error

	tmpdir, err = ioutil.TempDir("", "process-tracker-tests")
	Expect(err).ToNot(HaveOccurred())

	err = os.MkdirAll(filepath.Join(tmpdir, "bin"), 0755)
	Expect(err).ToNot(HaveOccurred())

	err = copyFile(iodaemonBin, filepath.Join(tmpdir, "bin", "iodaemon"))
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	os.RemoveAll(tmpdir)
})

var _ = Describe("Running processes", func() {
	BeforeEach(func() {
		processTracker = process_tracker.New(tmpdir, linux_command_runner.New())
	})

	It("runs the process and returns its exit code", func() {
		cmd := exec.Command("bash", "-c", "exit 42")

		process, err := processTracker.Run(55, cmd, garden.ProcessIO{}, nil, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(process.Wait()).To(Equal(42))
	})

	Describe("signalling a running process", func() {
		var (
			process   garden.Process
			signaller *FakeSignaller
		)

		BeforeEach(func() {
			signaller = &FakeSignaller{}
			cmd := exec.Command("bash", "-c", "echo hi")

			var err error
			process, err = processTracker.Run(2, cmd, garden.ProcessIO{}, nil, signaller)
			Expect(err).NotTo(HaveOccurred())
		})

		It("kills the process with a kill signal", func() {
			Expect(process.Signal(garden.SignalKill)).To(Succeed())
			Expect(signaller.sent).To(Equal([]os.Signal{os.Kill}))
		})

		It("kills the process with a terminate signal", func() {
			Expect(process.Signal(garden.SignalTerminate)).To(Succeed())
			Expect(signaller.sent).To(Equal([]os.Signal{syscall.SIGTERM}))
		})

		It("errors when an unsupported signal is sent", func() {
			Expect(process.Signal(garden.Signal(999))).To(MatchError(HaveSuffix("failed to send signal: unknown signal: 999")))
			Expect(signaller.sent).To(BeNil())
		})
	})

	It("streams the process's stdout and stderr", func() {
		cmd := exec.Command(
			"/bin/bash",
			"-c",
			"echo 'hi out' && echo 'hi err' >&2",
		)

		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()

		_, err := processTracker.Run(55, cmd, garden.ProcessIO{
			Stdout: stdout,
			Stderr: stderr,
		}, nil, nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("hi out\n"))
		Eventually(stderr).Should(gbytes.Say("hi err\n"))
	})

	It("streams input to the process", func() {
		stdout := gbytes.NewBuffer()

		_, err := processTracker.Run(55, exec.Command("cat"), garden.ProcessIO{
			Stdin:  bytes.NewBufferString("stdin-line1\nstdin-line2\n"),
			Stdout: stdout,
		}, nil, nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("stdin-line1\nstdin-line2\n"))
	})

	Context("when there is an error reading the stdin stream", func() {
		It("does not close the process's stdin", func() {
			pipeR, pipeW := io.Pipe()
			stdout := gbytes.NewBuffer()

			process, err := processTracker.Run(55, exec.Command("cat"), garden.ProcessIO{
				Stdin:  pipeR,
				Stdout: stdout,
			}, nil, nil)
			Expect(err).NotTo(HaveOccurred())

			pipeW.Write([]byte("Hello stdin!"))
			Eventually(stdout).Should(gbytes.Say("Hello stdin!"))

			pipeW.CloseWithError(errors.New("Failed"))
			Consistently(stdout, 0.1).ShouldNot(gbytes.Say("."))

			pipeR, pipeW = io.Pipe()
			processTracker.Attach(process.ID(), garden.ProcessIO{
				Stdin: pipeR,
			})

			pipeW.Write([]byte("Hello again, stdin!"))
			Eventually(stdout).Should(gbytes.Say("Hello again, stdin!"))

			pipeW.Close()
			Expect(process.Wait()).To(Equal(0))
		})

		It("supports attaching more than once", func() {
			pipeR, pipeW := io.Pipe()
			stdout := gbytes.NewBuffer()

			process, err := processTracker.Run(55, exec.Command("cat"), garden.ProcessIO{
				Stdin:  pipeR,
				Stdout: stdout,
			}, nil, nil)
			Expect(err).NotTo(HaveOccurred())

			pipeW.Write([]byte("Hello stdin!"))
			Eventually(stdout).Should(gbytes.Say("Hello stdin!"))

			pipeW.CloseWithError(errors.New("Failed"))
			Consistently(stdout, 0.1).ShouldNot(gbytes.Say("."))

			pipeR, pipeW = io.Pipe()
			_, err = processTracker.Attach(process.ID(), garden.ProcessIO{
				Stdin: pipeR,
			})
			Expect(err).ToNot(HaveOccurred())

			pipeW.Write([]byte("Hello again, stdin!"))
			Eventually(stdout).Should(gbytes.Say("Hello again, stdin!"))

			pipeR, pipeW = io.Pipe()

			_, err = processTracker.Attach(process.ID(), garden.ProcessIO{
				Stdin: pipeR,
			})
			Expect(err).ToNot(HaveOccurred())

			pipeW.Write([]byte("Hello again again, stdin!"))
			Eventually(stdout, "1s").Should(gbytes.Say("Hello again again, stdin!"))

			pipeW.Close()
			Expect(process.Wait()).To(Equal(0))
		})
	})

	Context("with a tty", func() {
		It("forwards TTY signals to the process", func() {
			cmd := exec.Command("/bin/bash", "-c", `
				trap "stty size; exit 123" SIGWINCH
				stty size
				read
			`)

			stdout := gbytes.NewBuffer()

			process, err := processTracker.Run(55, cmd, garden.ProcessIO{
				Stdout: stdout,
			}, &garden.TTYSpec{
				WindowSize: &garden.WindowSize{
					Columns: 95,
					Rows:    13,
				},
			}, nil)
			Expect(err).NotTo(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("13 95"))

			process.SetTTY(garden.TTYSpec{
				WindowSize: &garden.WindowSize{
					Columns: 101,
					Rows:    27,
				},
			})

			Eventually(stdout).Should(gbytes.Say("27 101"))
			Expect(process.Wait()).To(Equal(123))
		})

		Describe("when a window size is not specified", func() {
			It("picks a default window size", func() {
				cmd := exec.Command("/bin/bash", "-c", `
					stty size
				`)

				stdout := gbytes.NewBuffer()

				_, err := processTracker.Run(55, cmd, garden.ProcessIO{
					Stdout: stdout,
				}, &garden.TTYSpec{}, nil)
				Expect(err).NotTo(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("24 80"))
			})
		})
	})

	Context("when spawning fails", func() {
		It("returns the error", func() {
			_, err := processTracker.Run(55, exec.Command("/bin/does-not-exist"), garden.ProcessIO{}, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Restoring processes", func() {
	BeforeEach(func() {
		processTracker = process_tracker.New(tmpdir, linux_command_runner.New())
	})

	It("tracks the restored process", func() {
		processTracker.Restore(2, nil)

		activeProcesses := processTracker.ActiveProcesses()
		Expect(activeProcesses).To(HaveLen(1))
		Expect(activeProcesses[0].ID()).To(Equal(uint32(2)))
	})

	It("assigns the signaller to the process", func() {
		signaller := &FakeSignaller{}
		processTracker.Restore(2, signaller)

		activeProcesses := processTracker.ActiveProcesses()
		Expect(activeProcesses).To(HaveLen(1))

		Expect(activeProcesses[0].Signal(garden.SignalKill)).To(Succeed())
		Expect(signaller.sent).To(Equal([]os.Signal{os.Kill}))
	})
})

var _ = Describe("Attaching to running processes", func() {
	BeforeEach(func() {
		processTracker = process_tracker.New(tmpdir, linux_command_runner.New())
	})

	It("streams stdout, stdin, and stderr", func() {
		cmd := exec.Command("bash", "-c", `
			stuff=$(cat)
			echo "hi stdout" $stuff
			echo "hi stderr" $stuff >&2
		`)

		process, err := processTracker.Run(55, cmd, garden.ProcessIO{}, nil, nil)
		Expect(err).NotTo(HaveOccurred())

		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()

		process, err = processTracker.Attach(process.ID(), garden.ProcessIO{
			Stdin:  bytes.NewBufferString("this-is-stdin"),
			Stdout: stdout,
			Stderr: stderr,
		})
		Expect(err).NotTo(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("hi stdout this-is-stdin"))
		Eventually(stderr).Should(gbytes.Say("hi stderr this-is-stdin"))
	})
})

var _ = Describe("Listing active process IDs", func() {
	BeforeEach(func() {
		processTracker = process_tracker.New(tmpdir, linux_command_runner.New())
	})

	It("includes running process IDs", func() {
		stdin1, stdinWriter1 := io.Pipe()
		stdin2, stdinWriter2 := io.Pipe()

		Expect(processTracker.ActiveProcesses()).To(BeEmpty())

		process1, err := processTracker.Run(55, exec.Command("cat"), garden.ProcessIO{
			Stdin: stdin1,
		}, nil, nil)
		Expect(err).ToNot(HaveOccurred())

		Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process1))

		process2, err := processTracker.Run(56, exec.Command("cat"), garden.ProcessIO{
			Stdin: stdin2,
		}, nil, nil)
		Expect(err).ToNot(HaveOccurred())

		Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process1, process2))

		stdinWriter1.Close()
		Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process2))

		stdinWriter2.Close()
		Eventually(processTracker.ActiveProcesses).Should(BeEmpty())
	})
})

type FakeSignaller struct {
	sent []os.Signal
}

func (f *FakeSignaller) Signal(s os.Signal) error {
	f.sent = append(f.sent, s)
	return nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}

	defer s.Close()

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(d, s)
	if err != nil {
		d.Close()
		return err
	}

	return d.Close()
}
