package gardenrunc

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/cloudfoundry-incubator/garden"
)

//go:generate counterfeiter . Creator
type Creator interface {
	Create(spec garden.ContainerSpec) (*Container, error)
}

type Repo interface {
	All() []*Container
	Add(*Container)
	FindByHandle(string) (*Container, error)
	Query(filter func(*Container) bool) []*Container
	Delete(*Container)
}

type Backend struct {
	Creator Creator
	Repo    Repo
}

func (b *Backend) Create(spec garden.ContainerSpec) (garden.Container, error) {
	var err error
	var container *Container

	if container, err = b.Creator.Create(spec); err != nil {
		return nil, err
	}

	b.Repo.Add(container)

	return container, err
}

func (backend *Backend) Start() error {
	script := `
	mkdir -p $1

	if ! mountpoint -q $1; then
	  mount -t tmpfs -o uid=0,gid=0,mode=0755 cgroup $1
	fi

	for subsystem in $(tail -n +2 /proc/cgroups | awk '{print $1}'); do
		mkdir -p ${1}/$subsystem

		if ! mountpoint -q ${1}/$subsystem; then
			mount -n -t cgroup -o $subsystem cgroup ${1}/$subsystem
		fi
	done
	`

	// hack to ensure cgroups are mounted, because they're not in fly
	// need to copy over code from ./bin/setup to ensure cgroups are mounted properly
	// on startup
	if out, err := exec.Command("sh", "-c", script, "/tmp/garden-cgroups").CombinedOutput(); err != nil {
		return fmt.Errorf("create: mount cgroups: %s\n\n%s", err, out)
	}

	return nil
}

func (backend *Backend) Stop() {
}

func (backend *Backend) GraceTime(garden.Container) time.Duration {
	return 5 * time.Minute
}

func (backend *Backend) Ping() error {
	return nil
}

func (backend *Backend) Capacity() (garden.Capacity, error) {
	// lies.
	return garden.Capacity{
		MemoryInBytes: 500000,
		DiskInBytes:   500000,
		MaxContainers: 1000,
	}, nil
}

func (backend *Backend) Destroy(handle string) error {
	return nil
}

func (b *Backend) Containers(props garden.Properties) ([]garden.Container, error) {
	return toGardenContainers(b.Repo.Query(withProperties(props))), nil
}

func (b *Backend) Lookup(handle string) (garden.Container, error) {
	return b.Repo.FindByHandle(handle)
}

func (b *Backend) BulkInfo(handles []string) (map[string]garden.ContainerInfoEntry, error) {
	containers := b.Repo.Query(withHandles(handles))

	infos := make(map[string]garden.ContainerInfoEntry)
	for _, container := range containers {
		info, err := container.Info()
		infos[container.Handle()] = garden.ContainerInfoEntry{
			Info: info,
			Err:  err,
		}
	}

	return infos, nil
}

func (b *Backend) BulkMetrics(handles []string) (map[string]garden.ContainerMetricsEntry, error) {
	containers := b.Repo.Query(withHandles(handles))

	metrics := make(map[string]garden.ContainerMetricsEntry)
	for _, container := range containers {
		metric, err := container.Metrics()
		metrics[container.Handle()] = garden.ContainerMetricsEntry{
			Metrics: metric,
			Err:     err,
		}
	}

	return metrics, nil
}

func withProperties(properties garden.Properties) func(*Container) bool {
	return func(c *Container) bool {
		return c.HasProperties(properties)
	}
}

func withHandles(handles []string) func(*Container) bool {
	return func(c *Container) bool {
		for _, e := range handles {
			if e == c.Handle() {
				return true
			}
		}
		return false
	}
}

func toGardenContainers(cs []*Container) []garden.Container {
	var result []garden.Container
	for _, c := range cs {
		result = append(result, c)
	}

	return result
}
