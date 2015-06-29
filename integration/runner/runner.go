package runner

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cloudfoundry-incubator/garden/client"
	"github.com/cloudfoundry-incubator/garden/client/connection"
	"github.com/onsi/ginkgo"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type Runner struct {
	Command *exec.Cmd

	network string
	addr    string

	bin  string
	argv []string

	tmpdir string
}

func New(network, addr, bin string, argv ...string) *Runner {
	return &Runner{
		network: network,
		addr:    addr,

		bin:  bin,
		argv: argv,

		tmpdir: filepath.Join(
			os.TempDir(),
			fmt.Sprintf("test-garden-%d", ginkgo.GinkgoParallelNode()),
		),
	}
}

func (r *Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := lagertest.NewTestLogger("garden-runner")

	err := os.MkdirAll(r.tmpdir, 0755)
	if err != nil {
		return err
	}

	var appendDefaultFlag = func(ar []string, key, value string) []string {
		for _, a := range r.argv {
			if a == key {
				return ar
			}
		}

		if value != "" {
			return append(ar, key, value)
		} else {
			return append(ar, key)
		}
	}

	gardenArgs := make([]string, len(r.argv))
	copy(gardenArgs, r.argv)

	gardenArgs = appendDefaultFlag(gardenArgs, "--listenNetwork", r.network)
	gardenArgs = appendDefaultFlag(gardenArgs, "--listenAddr", r.addr)
	gardenArgs = appendDefaultFlag(gardenArgs, "--logLevel", "debug")
	gardenArgs = appendDefaultFlag(gardenArgs, "--depotDir", path.Join(r.tmpdir, "containers"))

	//gardenArgs = appendDefaultFlag(gardenArgs, "--networkPool", fmt.Sprintf("10.250.%d.0/24", ginkgo.GinkgoParallelNode()))
	//gardenArgs = appendDefaultFlag(gardenArgs, "--portPoolStart", strconv.Itoa(51000+(1000*ginkgo.GinkgoParallelNode())))
	//gardenArgs = appendDefaultFlag(gardenArgs, "--portPoolSize", "1000")

	var signal os.Signal

	r.Command = exec.Command(r.bin, gardenArgs...)

	process := ifrit.Invoke(&ginkgomon.Runner{
		Name:              "garden-runc",
		Command:           r.Command,
		AnsiColorCode:     "31m",
		StartCheck:        "garden-runc.started",
		StartCheckTimeout: 10 * time.Second,
		Cleanup: func() {
			if signal == syscall.SIGQUIT {
				logger.Info("cleanup-tempdirs")
				if err := os.RemoveAll(r.tmpdir); err != nil {
					logger.Error("cleanup-tempdirs-failed", err, lager.Data{"tmpdir": r.tmpdir})
				} else {
					logger.Info("tempdirs-removed")
				}
			}
		},
	})

	close(ready)

	for {
		select {
		case signal = <-signals:
			// SIGQUIT means clean up the containers, the garden process (SIGTERM) and the temporary directories
			// SIGKILL, SIGTERM and SIGINT are passed through to the garden process
			if signal == syscall.SIGQUIT {
				logger.Info("received-signal SIGQUIT")
				if err := r.destroyContainers(); err != nil {
					logger.Error("destroy-containers-failed", err)
					return err
				}
				logger.Info("destroyed-containers")
				process.Signal(syscall.SIGTERM)
			} else {
				logger.Info("received-signal", lager.Data{"signal": signal})
				process.Signal(signal)
			}

		case waitErr := <-process.Wait():
			logger.Info("process-exited")
			return waitErr
		}
	}
}

func (r *Runner) TryDial() error {
	conn, dialErr := net.DialTimeout(r.network, r.addr, 100*time.Millisecond)

	if dialErr == nil {
		conn.Close()
		return nil
	}

	return dialErr
}

func (r *Runner) NewClient() client.Client {
	return client.New(connection.New(r.network, r.addr))
}

func (r *Runner) destroyContainers() error {
	client := r.NewClient()

	containers, err := client.Containers(nil)
	if err != nil {
		return err
	}

	for _, container := range containers {
		err := client.Destroy(container.Handle())
		if err != nil {
			return err
		}
	}

	return nil
}
