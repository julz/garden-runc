package gardenrunc

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/nu7hatch/gouuid"
	"github.com/onsi/gomega/gexec"
)

//go:generate counterfeiter . Depot
type Depot interface {
	Create() (string, error)
}

type ContainerDepot struct {
	Dir string
}

func (depot *ContainerDepot) Create() (string, error) {
	containerDir := path.Join(depot.Dir, guid())
	runDir := path.Join(containerDir, "run")
	binDir := path.Join(containerDir, "bin")
	processesDir := path.Join(containerDir, "processes")
	os.MkdirAll(runDir, 0700)
	os.MkdirAll(processesDir, 0700)

	if err := os.MkdirAll(binDir, 0777); err != nil {
		panic(err)
	}

	// FIXME(jz) remove all this hackery and just re-exec
	iodaemonPath, err := gexec.Build("github.com/cloudfoundry-incubator/garden-linux/iodaemon")
	if err != nil {
		return "", fmt.Errorf("build iodaemon: %s", err)
	}

	doshPath, err := gexec.Build("github.com/julz/garden-runc/cmd/dosh")
	if err != nil {
		return "", fmt.Errorf("build dosh: %s", err)
	}

	cp := exec.Command("cp", iodaemonPath, path.Join(binDir, "iodaemon"))
	cp.Stderr = os.Stderr
	if err := cp.Run(); err != nil {
		return "", fmt.Errorf("copy iodaemon: %s", err)
	}

	cp = exec.Command("cp", doshPath, path.Join(binDir, "dosh"))
	cp.Stderr = os.Stderr
	if err := cp.Run(); err != nil {
		return "", fmt.Errorf("copy dosh: %s", err)
	}

	cp = exec.Command("chmod", "u+x", path.Join(binDir, "dosh"))
	cp.Stderr = os.Stderr
	if err := cp.Run(); err != nil {
		return "", fmt.Errorf("chmod dosh: %s", err)
	}

	return containerDir, nil
}

func guid() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return u.String()
}
