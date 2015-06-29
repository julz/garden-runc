package gardenrunc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/port_pool"
	"github.com/cloudfoundry-incubator/garden-linux/process_tracker"
	"github.com/cloudfoundry/gunk/command_runner"
	"github.com/docker/libnetwork/iptables"
	"github.com/onsi/gomega/gexec"

	runc "github.com/opencontainers/runc"
)

type RuncContainerCreator struct {
	DefaultRootfs string
	Depot         Depot

	DoshPath  string
	InitdPath string

	Chain    *iptables.Chain
	PortPool *port_pool.PortPool

	CommandRunner command_runner.CommandRunner
}

func (c *RuncContainerCreator) Create(spec garden.ContainerSpec) (*Container, error) {
	dir, err := c.Depot.Create()
	if err != nil {
		return nil, fmt.Errorf("create depot dir: %s", err)
	}

	if len(spec.RootFSPath) == 0 {
		spec.RootFSPath = c.DefaultRootfs
	}

	rootfs, err := url.Parse(spec.RootFSPath)
	if err != nil {
		return nil, fmt.Errorf("create: not a valid rootfs path: %s", err)
	}

	if _, err := exec.Command("cp", "-r", rootfs.Path, path.Join(dir, "rootfs")).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("create: copy rootfs: %s", err)
	}

	runcSpec := runc.PortableSpec{
		Version: "0.1",
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Cpus:    1.1,
		Memory:  1024,
		Root: runc.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Capabilities: []string{
			"AUDIT_WRITE",
			"KILL",
			"NET_BIND_SERVICE",
			"SETUID",
			"SETGID",
		},
		Namespaces: []runc.Namespace{
			{
				Type: "process",
			},
			{
				Type: "network",
			},
			{
				Type: "mount",
			},
			{
				Type: "ipc",
			},
			{
				Type: "uts",
			},
		},
		Devices: []string{
			"null",
			"random",
			"full",
			"tty",
			"zero",
			"urandom",
		},
		Mounts: []runc.Mount{
			{
				Type:        "proc",
				Source:      "proc",
				Destination: "/proc",
				Options:     "",
			},
			{
				Type:        "tmpfs",
				Source:      "tmpfs",
				Destination: "/dev",
				Options:     "nosuid,strictatime,mode=755,size=65536k",
			},
			{
				Type:        "devpts",
				Source:      "devpts",
				Destination: "/dev/pts",
				Options:     "nosuid,noexec,newinstance,ptmxmode=0666,mode=0620,gid=5",
			},
			{
				Type:        "tmpfs",
				Source:      "shm",
				Destination: "/dev/shm",
				Options:     "nosuid,noexec,nodev,mode=1777,size=65536k",
			},
			{
				Type:        "mqueue",
				Source:      "mqueue",
				Destination: "/dev/mqueue",
				Options:     "nosuid,noexec,nodev",
			},
			{
				Type:        "sysfs",
				Source:      "sysfs",
				Destination: "/sys",
				Options:     "nosuid,noexec,nodev",
			},
			{
				Type:        "bind",
				Source:      c.InitdPath,
				Destination: "/garden-bin/initd",
				Options:     "bind",
			}, {
				Type:        "bind",
				Source:      path.Join(dir, "run"),
				Destination: "/run/garden",
				Options:     "bind",
			}},

		Processes: []*runc.Process{{
			// User: "root",
			// Args: []string{
			// 	"/bin/ls", "-lR", "/garden-bin",
			// },
			// }},
			User: "root",
			Args: []string{
				"/garden-bin/initd",
				"-socket", "/run/garden/initd.sock",
				"-unmountAfterListening", "/run/garden",
			},
		}},
	}

	data, err := json.MarshalIndent(&runcSpec, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("create: marshal runc spec: %s", err)
	}

	err = ioutil.WriteFile(path.Join(dir, "container.json"), data, 0700)
	if err != nil {
		return nil, fmt.Errorf("create: write runc spec: %s", err)
	}

	os.Setenv("CGO_ENABLED", "1")
	runcBin, err := gexec.Build("github.com/opencontainers/runc")
	if err != nil {
		return nil, fmt.Errorf("create: build runc: %s", err)
	}

	runcCommand := exec.Command(runcBin)
	runcCommand.Dir = dir
	if err := c.CommandRunner.Start(runcCommand); err != nil {
		return nil, fmt.Errorf("create: start runc container: %s", err)
	}

	time.Sleep(2 * time.Second)

	return &Container{
		LimitsHandler: &LimitsHandler{},
		StreamHandler: &StreamHandler{},
		InfoHandler: &InfoHandler{
			Spec:          spec,
			ContainerPath: dir,
			PropsHandler:  &PropsHandler{},
		},
		NetHandler: &NetHandler{
			Chain:    c.Chain,
			PortPool: c.PortPool,
		},
		RunHandler: &RunHandler{
			ProcessTracker: process_tracker.New(dir, c.CommandRunner),
			ContainerCmd: &doshcmd{
				Path:      filepath.Join(dir, "bin", "dosh"),
				InitdSock: filepath.Join(dir, "run", "initd.sock"),
			},
		},
	}, nil
}

type doshcmd struct {
	Path      string
	InitdSock string
}

func (d doshcmd) Cmd(path string, args ...string) *exec.Cmd {
	doshArgs := []string{"-socket", d.InitdSock, "-user", "root"}
	run := []string{path}
	run = append(run, args...)
	return exec.Command(d.Path, append(doshArgs, run...)...)
}
