package integration_test

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/cloudfoundry-incubator/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"

	"github.com/julz/garden-runc/integration/runner"
)

var gardenBin string

var gardenRunner *runner.Runner
var gardenProcess ifrit.Process

var client garden.Client

func startGarden(argv ...string) garden.Client {
	gardenAddr := fmt.Sprintf("/tmp/garden_%d.sock", GinkgoParallelNode())
	gardenRunner = runner.New("unix", gardenAddr, gardenBin, argv...)
	gardenProcess = ifrit.Invoke(gardenRunner)

	return gardenRunner.NewClient()
}

func restartGarden(argv ...string) {
	Expect(client.Ping()).To(Succeed(), "tried to restart garden while it was not running")
	gardenProcess.Signal(syscall.SIGTERM)
	Eventually(gardenProcess.Wait(), 10).Should(Receive())

	startGarden(argv...)
}

func ensureGardenRunning() {
	if err := client.Ping(); err != nil {
		client = startGarden()
	}

	Expect(client.Ping()).ToNot(HaveOccurred())
}

func TestLifecycle(t *testing.T) {
	SynchronizedBeforeSuite(func() []byte {
		gardenPath, err := gexec.Build("github.com/julz/garden-runc/cmd/garden-runc")
		Expect(err).ToNot(HaveOccurred())
		return []byte(gardenPath)
	}, func(gardenPath []byte) {
		gardenBin = string(gardenPath)
	})

	AfterEach(func() {
		ensureGardenRunning()
		gardenProcess.Signal(syscall.SIGQUIT)
		Eventually(gardenProcess.Wait(), 10).Should(Receive())
	})

	SynchronizedAfterSuite(func() {
		//noop
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Lifecycle Suite")
}
