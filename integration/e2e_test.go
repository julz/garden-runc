package integration_test

import (
	"io"
	"os/exec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry/gunk/localip"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("garden-runc", func() {
	FIt("runs a process and streams the output", func() {
		client = startGarden()

		container, err := client.Create(garden.ContainerSpec{})
		Expect(err).NotTo(HaveOccurred())

		stdout := gbytes.NewBuffer()
		_, err = container.Run(garden.ProcessSpec{
			User: "root",
			Path: "/bin/echo",
			Args: []string{
				"foo",
			},
		}, garden.ProcessIO{Stdout: io.MultiWriter(GinkgoWriter, stdout), Stderr: GinkgoWriter})

		Expect(err).NotTo(HaveOccurred())
		Eventually(stdout, "2s").Should(gbytes.Say("foo"))
	})

	It("can be pinged", func() {
		client = startGarden()

		container, err := client.Create(garden.ContainerSpec{})
		Expect(err).NotTo(HaveOccurred())

		info, err := container.Info()
		Expect(err).NotTo(HaveOccurred())

		session, err := gexec.Start(exec.Command("ping", "-c", "1", info.ContainerIP), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
	})

	It("can be contacted via a NetIn-ed port", func() {
		client = startGarden()

		container, err := client.Create(garden.ContainerSpec{})
		Expect(err).NotTo(HaveOccurred())

		_, err = container.Run(garden.ProcessSpec{
			Path: "sh",
			Args: []string{"-c", "echo hello | nc -l -p 3333"},
		}, garden.ProcessIO{Stdout: GinkgoWriter, Stderr: GinkgoWriter})
		Expect(err).NotTo(HaveOccurred())

		externalIP, err := localip.LocalIP()
		Expect(err).NotTo(HaveOccurred())

		_, _, err = container.NetIn(9898, 3333)
		Expect(err).NotTo(HaveOccurred())

		nc, err := gexec.Start(exec.Command("nc", externalIP, "9898"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(nc).Should(gexec.Exit(0))
		Eventually(nc).Should(gbytes.Say("hello"))
	})

	PIt("runs a process as a requested user", func() {
	})

	PIt("runs multiple processes in the same container", func() {
	})

	PIt("doesnt copy binaries around(!)", func() {})
})
