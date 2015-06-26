// initd runs as pid 1 inside a container, listens on a socket and spawns processes
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/garden-linux/container_daemon"
	"github.com/cloudfoundry-incubator/garden-linux/container_daemon/unix_socket"
	"github.com/cloudfoundry-incubator/garden-linux/containerizer/system"
	"github.com/pivotal-golang/lager"
)

func main() {
	logger := lager.NewLogger("initd")
	socketPath := flag.String("socket", "/run/initd.sock", "path to listen for spawn requests on")
	//unmountPath := flag.String("unmountAfterListening", "/run", "directory to unmount after succesfully listening on -socketPath")
	flag.String("unmountAfterListening", "/run", "directory to unmount after succesfully listening on -socketPath")
	flag.Parse()

	reaper := system.StartReaper(logger)
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
		},
		Spawner: &container_daemon.Spawn{
			Runner: reaper,
			PTY:    system.KrPty,
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
