package cmd

import (
	"errors"
	"fmt"

	"devsentry/internal/clierr"
)

func usageError(format string, args ...any) error {
	return clierr.Wrap(clierr.ErrUsage, fmt.Errorf(format, args...))
}

func usageArgs(cmdName string, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return usageError("%s: unknown command %q", cmdName, args[0])
}

func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, clierr.ErrUsage):
		return 2
	case errors.Is(err, clierr.ErrDockerUnavailable):
		return 3
	case errors.Is(err, clierr.ErrTimeout):
		return 4
	case errors.Is(err, clierr.ErrMutation):
		return 5
	default:
		return 1
	}
}
