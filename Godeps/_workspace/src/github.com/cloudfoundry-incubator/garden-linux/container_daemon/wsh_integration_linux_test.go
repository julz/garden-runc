package container_daemon_test

import (
	"os/exec"
	"syscall"

	"github.com/cloudfoundry-incubator/garden-linux/container_daemon"

	"io/ioutil"

	"path"

	"github.com/cloudfoundry-incubator/garden-linux/container_daemon/unix_socket"
	"github.com/cloudfoundry-incubator/garden-linux/linux_backend"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type FakeCommandRunner struct {
}

func (r *FakeCommandRunner) Start(cmd *exec.Cmd) error {
	return cmd.Start()
}

func (r *FakeCommandRunner) Wait(cmd *exec.Cmd) byte {
	return exitStatusFromErr(cmd.Wait())
}

var _ = Describe("wsh and daemon integration", func() {
	var daemon *container_daemon.ContainerDaemon
	var tempDir string
	var socketPath string

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		socketPath = path.Join(tempDir, "test.sock")
		listener, err := unix_socket.NewListenerFromPath(socketPath)
		Expect(err).ToNot(HaveOccurred())

		daemon = &container_daemon.ContainerDaemon{
			CmdPreparer: &container_daemon.ProcessSpecPreparer{
				Users:           container_daemon.LibContainerUser{},
				ProcStarterPath: procStarterBin,
				Rlimits:         &container_daemon.RlimitsManager{},
			},
			Spawner: &container_daemon.Spawn{
				Runner: &FakeCommandRunner{},
			},
		}

		go func(listener container_daemon.Listener) {
			defer GinkgoRecover()
			Expect(daemon.Run(listener)).To(Succeed())
		}(listener)
	})

	It("should run a program when no pidfile is specified", func() {
		wshCmd := exec.Command(wshBin,
			"--socket", socketPath,
			"--user", "root",
			"echo", "hello")

		op, err := wshCmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(op)).To(Equal("hello\n"))
	})

	It("should avoid a race condition when sending a kill signal", func(done Done) {
		for i := 0; i < 20; i++ {
			pidfilePath := path.Join(tempDir, "cmd.pid")
			wshCmd := exec.Command(wshBin,
				"--socket", socketPath,
				"--pidfile", pidfilePath,
				"--user", "root",
				"sh", "-c",
				`while true; do echo -n "x"; sleep 1; done`)

			err := wshCmd.Start()
			Expect(err).ToNot(HaveOccurred())

			Expect(kill(pidfilePath, syscall.SIGKILL)).To(Succeed())
			Expect(err).ToNot(HaveOccurred())
			Expect(exitStatusFromErr(wshCmd.Wait())).To(Equal(byte(255)))
		}
		close(done)
	}, 40.0)

	It("receives the correct exit status and output from a process which is sent SIGTERM", func(done Done) {
		stdout := gbytes.NewBuffer()

		pidfilePath := path.Join(tempDir, "cmd.pid")
		wshCmd := exec.Command(wshBin,
			"--socket", socketPath,
			"--pidfile", pidfilePath,
			"--user", "root",
			"sh", "-c", `
				  trap 'echo termed; exit 42' TERM

					while true; do
					  echo waiting
					  sleep 1
					done
				`)
		wshCmd.Stdout = stdout
		wshCmd.Stderr = GinkgoWriter

		err := wshCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("waiting"))

		Expect(kill(pidfilePath, syscall.SIGTERM)).To(Succeed())

		Expect(exitStatusFromErr(wshCmd.Wait())).To(Equal(byte(42)))
		Eventually(stdout, "2s").Should(gbytes.Say("termed"))

		close(done)
	}, 320.0)

	It("receives the correct exit status and output from a process exits 255", func(done Done) {
		for i := 0; i < 200; i++ {
			stdout := gbytes.NewBuffer()

			wshCmd := exec.Command(wshBin,
				"--socket", socketPath,
				"--user", "root",
				"sh", "-c", `
					for i in $(seq 0 512); do
					  echo 0123456789
					done

					echo ended
					exit 255
				`)
			wshCmd.Stdout = stdout
			wshCmd.Stderr = GinkgoWriter

			err := wshCmd.Start()
			Expect(err).ToNot(HaveOccurred())

			Expect(exitStatusFromErr(wshCmd.Wait())).To(Equal(byte(255)))
			Eventually(stdout, "2s").Should(gbytes.Say("ended"))
		}
		close(done)
	}, 320.0)

	It("applies the provided rlimits", func() {
		wshCmd := exec.Command(wshBin,
			"--socket", socketPath,
			"--user", "root",
			"sh", "-c",
			"ulimit -n")

		wshCmd.Env = append(wshCmd.Env, "RLIMIT_NOFILE=16")

		op, err := wshCmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(op)).To(Equal("16\n"))
	})
})

func exitStatusFromErr(err error) byte {
	if exitError, ok := err.(*exec.ExitError); ok {
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		return byte(waitStatus.ExitStatus())
	} else if err != nil {
		println("exitStatusFromErr found error", err)
		return container_daemon.UnknownExitStatus
	} else {
		return 0
	}
}

func kill(pidFilePath string, signal syscall.Signal) error {
	pid, err := linux_backend.PidFromFile(pidFilePath)
	if err != nil {
		return err
	}

	return syscall.Kill(pid, signal)
}
