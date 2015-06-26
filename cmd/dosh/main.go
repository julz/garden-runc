// dosh connects to a running garden-runc container daemon, spawn a process, stream output
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/container_daemon"
	"github.com/cloudfoundry-incubator/garden-linux/container_daemon/unix_socket"

	_ "github.com/cloudfoundry-incubator/garden-linux/iodaemon"
)

func main() {
	socketPath := flag.String("socket", "./run/initd.sock", "Path to socket")
	user := flag.String("user", "vcap", "User to change to")
	dir := flag.String("dir", "", "Working directory for the running process")

	var envVars container_daemon.StringList
	flag.Var(&envVars, "env", "Environment variables to set for the command.")

	pidfile := flag.String("pidfile", "", "File to save container-namespaced pid of spawned process to")
	flag.Bool("rsh", false, "RSH compatibility mode")

	flag.Parse()

	extraArgs := flag.Args()
	if len(extraArgs) == 0 {
		// Default is to run a shell.
		extraArgs = []string{"/bin/sh"}
	}

	var tty *garden.TTYSpec
	resize := make(chan os.Signal)
	if terminal.IsTerminal(syscall.Stdin) {
		tty = &garden.TTYSpec{}
		signal.Notify(resize, syscall.SIGWINCH)
	}

	var pidfileWriter container_daemon.PidfileWriter = container_daemon.NoPidfile{}
	if *pidfile != "" {
		pidfileWriter = container_daemon.Pidfile{
			Path: *pidfile,
		}
	}

	rlimitFromEnv := func(envVar string) *uint64 {
		strVal := os.Getenv(envVar)
		if strVal == "" {
			return nil
		}

		var val uint64
		fmt.Sscanf(strVal, "%d", &val)
		return &val
	}

	process := &container_daemon.Process{
		Connector: &unix_socket.Connector{
			SocketPath: *socketPath,
		},

		Term: container_daemon.TermPkg{},

		Pidfile: pidfileWriter,

		SigwinchCh: resize,

		Spec: &garden.ProcessSpec{
			Path: extraArgs[0],
			Args: extraArgs[1:],
			Env:  envVars.List,
			Dir:  *dir,
			User: *user,
			TTY:  tty, // used as a boolean -- non-nil = attach pty
			Limits: garden.ResourceLimits{
				As:         rlimitFromEnv("RLIMIT_AS"),
				Core:       rlimitFromEnv("RLIMIT_CORE"),
				Cpu:        rlimitFromEnv("RLIMIT_CPU"),
				Data:       rlimitFromEnv("RLIMIT_DATA"),
				Fsize:      rlimitFromEnv("RLIMIT_FSIZE"),
				Locks:      rlimitFromEnv("RLIMIT_LOCKS"),
				Memlock:    rlimitFromEnv("RLIMIT_MEMLOCK"),
				Msgqueue:   rlimitFromEnv("RLIMIT_MSGQUEUE"),
				Nice:       rlimitFromEnv("RLIMIT_NICE"),
				Nofile:     rlimitFromEnv("RLIMIT_NOFILE"),
				Nproc:      rlimitFromEnv("RLIMIT_NPROC"),
				Rss:        rlimitFromEnv("RLIMIT_RSS"),
				Rtprio:     rlimitFromEnv("RLIMIT_RTPRIO"),
				Sigpending: rlimitFromEnv("RLIMIT_SIGPENDING"),
				Stack:      rlimitFromEnv("RLIMIT_STACK"),
			},
		},

		IO: &garden.ProcessIO{
			Stdin:  os.Stdin,
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		},
	}

	err := process.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start process: %s", err)
		os.Exit(container_daemon.UnknownExitStatus)
	}

	defer process.Cleanup()

	exitCode, err := process.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wait for process: %s", err)
		os.Exit(container_daemon.UnknownExitStatus)
	}

	os.Exit(exitCode)
}
