package system_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"github.com/cloudfoundry-incubator/garden-linux/containerizer/system"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Execer", func() {
	var execer *system.NamespacingExecer
	var commandRunner *fake_command_runner.FakeCommandRunner

	BeforeEach(func() {
		commandRunner = fake_command_runner.New()
		process := &os.Process{
			Pid: 12,
		}

		commandRunner.WhenRunning(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
			cmd.Process = process
			return nil
		})

		execer = &system.NamespacingExecer{
			CommandRunner:    commandRunner,
			UidMappingOffset: 101,
		}
	})

	Describe("Exec", func() {
		It("executes the given command", func() {
			_, err := execer.Exec("something", "smthg")
			Expect(err).To(Succeed())

			Expect(commandRunner).To(HaveStartedExecuting(
				fake_command_runner.CommandSpec{
					Path: "something",
					Args: []string{
						"smthg",
					},
				},
			))
		})

		It("returns the correct PID", func() {
			pid, err := execer.Exec("something", "smthg")
			Expect(pid).To(Equal(12))
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets the correct flags", func() {
			_, err := execer.Exec("something", "smthg")
			Expect(err).ToNot(HaveOccurred())

			cmd := commandRunner.StartedCommands()[0]
			Expect(cmd.SysProcAttr).ToNot(BeNil())
			flags := syscall.CLONE_NEWIPC
			flags = flags | syscall.CLONE_NEWNET
			flags = flags | syscall.CLONE_NEWNS
			flags = flags | syscall.CLONE_NEWUTS
			flags = flags | syscall.CLONE_NEWPID
			Expect(int(cmd.SysProcAttr.Cloneflags) & flags).ToNot(Equal(0))
		})

		Context("when the container is not privileged", func() {
			It("creates a user namespace", func() {
				_, err := execer.Exec("something", "smthg")
				Expect(err).ToNot(HaveOccurred())

				cmd := commandRunner.StartedCommands()[0]
				Expect(cmd.SysProcAttr).ToNot(BeNil())
				Expect(int(cmd.SysProcAttr.Cloneflags) & syscall.CLONE_NEWUSER).ToNot(Equal(0))
			})

			It("spawns as UID 0 (so that the process is run as container-root rather than 'nobody')", func() {
				_, err := execer.Exec("something", "smthg")
				Expect(err).ToNot(HaveOccurred())

				cmd := commandRunner.StartedCommands()[0]
				Expect(cmd.SysProcAttr).ToNot(BeNil())
				Expect(cmd.SysProcAttr.Credential).To(Equal(&syscall.Credential{
					Uid: 0,
					Gid: 0,
				}))
			})

			It("sets uid and gid mappings", func() {
				_, err := execer.Exec("something", "smthg")
				Expect(err).ToNot(HaveOccurred())

				cmd := commandRunner.StartedCommands()[0]
				Expect(cmd.SysProcAttr.UidMappings[0].HostID).To(Equal(101))
				Expect(cmd.SysProcAttr.GidMappings[0].HostID).To(Equal(101))
			})
		})

		Context("when the container is privileged", func() {
			It("does not create a user namespace", func() {
				execer.Privileged = true

				_, err := execer.Exec("something", "smthg")
				Expect(err).ToNot(HaveOccurred())

				cmd := commandRunner.StartedCommands()[0]
				Expect(cmd.SysProcAttr.Cloneflags & syscall.CLONE_NEWUSER).To(Equal(uintptr(0)))
			})
		})

		It("sets extra files", func() {
			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())
			execer.ExtraFiles = []*os.File{tmpFile}

			_, err = execer.Exec("something", "smthg")
			Expect(err).ToNot(HaveOccurred())

			cmd := commandRunner.StartedCommands()[0]
			Expect(cmd.ExtraFiles).To(HaveLen(1))
			Expect(cmd.ExtraFiles[0]).To(Equal(tmpFile))
		})
	})
})
