// initd runs as pid 1 inside a container, listens on a socket and spawns processes
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/cloudfoundry-incubator/garden-linux/container_daemon"
	"github.com/cloudfoundry-incubator/garden-linux/container_daemon/unix_socket"
	csystem "github.com/cloudfoundry-incubator/garden-linux/containerizer/system"
	"github.com/cloudfoundry-incubator/garden-linux/system"
	"github.com/docker/docker/pkg/reexec"
	"github.com/pivotal-golang/lager"
)

func init() {
	reexec.Register("proc_starter", ProcStarter)
}

func main() {
	if reexec.Init() {
		return
	}

	logger := lager.NewLogger("initd")
	socketPath := flag.String("socket", "/run/initd.sock", "path to listen for spawn requests on")
	//unmountPath := flag.String("unmountAfterListening", "/run", "directory to unmount after succesfully listening on -socketPath")

	flag.String("unmountAfterListening", "/run", "directory to unmount after succesfully listening on -socketPath")
	flag.Parse()

	reaper := csystem.StartReaper(logger)
	defer reaper.Stop()

	listener, err := unix_socket.NewListenerFromPath(*socketPath)
	if err != nil {
		fmt.Printf("open %s: %s", *socketPath, err)
		os.Exit(1)
	}

	daemon := &container_daemon.ContainerDaemon{
		CmdPreparer: &container_daemon.ProcessSpecPreparer{
			Users: container_daemon.LibContainerUser{},
			AlwaysDropCapabilities: false,
			ProcStarterPath:        "/garden-bin/initd",
			//DropCapabilities:       *dropCaps,
		},
		Spawner: &container_daemon.Spawn{
			Runner: reaper,
			PTY:    csystem.KrPty,
		},
	}

	// unmount the bind-mounted socket volume now we've started listening
	// if err := syscall.Unmount(*unmountPath, syscall.MNT_DETACH); err != nil {
	// 	fmt.Printf("unmount %s: %s", *unmountPath, err)
	// 	os.Exit(2)
	// }

	if err := daemon.Run(listener); err != nil {
		fmt.Sprintf("run daemon: %s", err)
		os.Exit(1)
	}
}

// proc_starter starts a user process with the correct rlimits and after
// closing any open FDs.
func ProcStarter() {
	runtime.LockOSThread()

	rlimits := flag.String("rlimits", "", "encoded rlimits")
	dropCapabilities := flag.Bool("dropCapabilities", true, "drop capabilities before starting process")
	uid := flag.Int("uid", -1, "user id to run the process as")
	gid := flag.Int("gid", -1, "group id to run the process as")
	extendedWhitelist := flag.Bool("extendedWhitelist", false, "whitelist CAP_SYS_ADMIN in addition to the default set. Use only with -dropCapabilities=true")
	flag.Parse()

	closeFds()

	mgr := &container_daemon.RlimitsManager{}
	must(mgr.Apply(mgr.DecodeLimits(*rlimits)))

	args := flag.Args()

	if *dropCapabilities {
		caps := &system.ProcessCapabilities{Pid: os.Getpid()}
		must(caps.Limit(*extendedWhitelist))
	}

	execer := system.UserExecer{}
	if err := execer.ExecAsUser(*uid, *gid, args[0], args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "proc_starter: ExecAsUser: %s\n", err)
		os.Exit(255)
	}
}

func closeFds() {
	fds, err := ioutil.ReadDir("/proc/self/fd")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: read /proc/self/fd: %s", err)
		os.Exit(255)
	}

	for _, fd := range fds {
		if fd.IsDir() {
			continue
		}

		fdI, err := strconv.Atoi(fd.Name())
		if err != nil {
			panic(err) // cant happen
		}

		if fdI <= 2 {
			continue
		}

		syscall.CloseOnExec(fdI)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
