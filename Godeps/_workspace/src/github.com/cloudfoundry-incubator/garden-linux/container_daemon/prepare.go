package container_daemon

import (
	"fmt"
	"os/exec"
	osuser "os/user"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden-linux/process"
)

//go:generate counterfeiter -o fake_rlimits_env_encoder/fake_rlimits_env_encoder.go . RlimitsEnvEncoder
type RlimitsEnvEncoder interface {
	EncodeLimits(garden.ResourceLimits) string
}

//go:generate counterfeiter -o fake_user/fake_user.go . User
type User interface {
	Lookup(name string) (*osuser.User, error)
}

type ProcessSpecPreparer struct {
	Users                  User
	ProcStarterPath        string
	Rlimits                RlimitsEnvEncoder
	AlwaysDropCapabilities bool
}

func (p *ProcessSpecPreparer) PrepareCmd(spec garden.ProcessSpec) (*exec.Cmd, error) {
	usr, err := p.parseUser(spec.User)
	if err != nil {
		return nil, fmt.Errorf("container_daemon: %s", err)
	}

	env, err := createEnvironment(spec.Env, usr)
	if err != nil {
		return nil, fmt.Errorf("container_daemon: %s", err)
	}

	dir := spec.Dir
	if spec.Dir == "" {
		dir = usr.homeDir
	}

	cmd := exec.Command(spec.Path, spec.Args...)
	cmd.Env = env.Array()
	cmd.Dir = dir
	// cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{
	// 	Uid: usr.uid,
	// 	Gid: usr.gid,
	// }}

	return cmd, nil
}

type parsedUser struct {
	username string
	uid      uint32
	gid      uint32
	homeDir  string
}

func (p *ProcessSpecPreparer) parseUser(username string) (parsedUser, error) {
	errs := func(err error) (parsedUser, error) {
		return parsedUser{}, err
	}

	ret := parsedUser{
		username: username,
	}

	if osUser, err := p.Users.Lookup(username); err == nil && osUser != nil {
		if _, err := fmt.Sscanf(osUser.Uid, "%d", &(ret.uid)); err != nil {
			return errs(fmt.Errorf("failed to parse uid %q", osUser.Uid))
		}
		if _, err := fmt.Sscanf(osUser.Gid, "%d", &(ret.gid)); err != nil {
			return errs(fmt.Errorf("failed to parse gid %q", osUser.Gid))
		}

		ret.homeDir = osUser.HomeDir

		return ret, nil
	} else if err == nil {
		return errs(fmt.Errorf("failed to lookup user %s", username))
	} else {
		return errs(fmt.Errorf("lookup user %s: %s", username, err))
	}
}

func createEnvironment(specEnv []string, usr parsedUser) (process.Env, error) {
	env, err := process.NewEnv(specEnv)
	if err != nil {
		return process.Env{}, fmt.Errorf("invalid environment %v: %s", specEnv, err)
	}

	env["USER"] = usr.username
	_, hasHome := env["HOME"]
	if !hasHome {
		env["HOME"] = usr.homeDir
	}

	_, hasPath := env["PATH"]
	if !hasPath {
		if usr.uid == 0 {
			env["PATH"] = DefaultRootPATH
		} else {
			env["PATH"] = DefaultUserPath
		}
	}

	return env, nil
}
