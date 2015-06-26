package gardenrunc

import "fmt"

type DockerCommandError struct {
	Stderr string
	Cause  error
}

func (err DockerCommandError) Error() string {
	return fmt.Sprintf("docker: %s (%s)", err.Stderr, err.Cause)
}
