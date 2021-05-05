package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Here we handle parsing legacy commands, as well as generating telepresence 2
// commands from them.  This will make it easier for users to move to
// telepresence 2.  Note: This isn't exhaustive, but should capture the major
// flags that were used and have a correlated command in telepresence 2.

type legacyCommand struct {
	swapDeployment string
	expose         string
	run            bool
	dockerRun      bool
	processCmd     string
	mount          string
	dockerMount    string
	envFile        string
	envJSON        string

	// kubectl-related flags
	context   string
	namespace string

	unknownFlags []string
}

// Unfortunately we have to do our own flag parsing if we see legacy telepresence
// flags because the run command might include something that cobra might detect
// as a flag e.g. --run python3 -m http.server. In python this was handled by
// using argparse.REMAINDER and there is not similar functionality with cobra.
// There is an open ticket to pass unknown flags to the command:
// https://github.com/spf13/cobra/issues/739
// but until that is addressed, we'll do the flag parsing ourself (which isn't
// the worst because it's a legacy command so the flags won't be growing).
func parseLegacyCommand(args []string) *legacyCommand {
	lc := &legacyCommand{}

	// We don't want to over-index in case somebody has a command that has a
	// flag but doesn't put the value after it.  So we have this helper function
	// to ensure we don't do that.  It may mean the telepresence command at the
	// end fails, but then they'll see the telepresence 2 error messge and can
	// fix it from there.
	getArg := func(i int) string {
		if len(args) > i {
			return args[i]
		}
		return ""
	}
Parsing:
	for i, v := range args {
		switch {
		case v == "--swap-deployment" || v == "-s":
			lc.swapDeployment = getArg(i + 1)
		case v == "--expose":
			lc.expose = getArg(i + 1)
		case v == "--mount":
			lc.mount = getArg(i + 1)
		case v == "--docker-mount":
			lc.dockerMount = getArg(i + 1)
		case v == "--env-json":
			lc.envJSON = getArg(i + 1)
		case v == "--env-file":
			lc.envFile = getArg(i + 1)
		case v == "--namespace":
			lc.namespace = getArg(i + 1)
		case v == "--run":
			lc.run = true
			if nxtArg := getArg(i + 1); nxtArg != "" {
				lc.processCmd = strings.Join(args[i+1:], " ")
			}
			break Parsing
		case v == "--docker-run":
			lc.dockerRun = true
			if nxtArg := getArg(i + 1); nxtArg != "" {
				lc.processCmd = strings.Join(args[i+1:], " ")
			}
			break Parsing
		case strings.Contains(v, "--"):
			lc.unknownFlags = append(lc.unknownFlags, v)
		}
	}
	return lc
}

// genTP2Command constructs a telepresence 2 command based on
// the values that are set in the legacyCommand struct
func (lc *legacyCommand) genTP2Command() (string, error) {
	cmdSlice := []string{"intercept"}
	if lc.swapDeployment != "" {
		cmdSlice = append(cmdSlice, lc.swapDeployment)
	}
	if lc.expose != "" {
		cmdSlice = append(cmdSlice, "--port", lc.expose)
	}

	if lc.envFile != "" {
		cmdSlice = append(cmdSlice, "--env-file", lc.envFile)
	}

	if lc.envJSON != "" {
		cmdSlice = append(cmdSlice, "--env-json", lc.envJSON)
	}

	if lc.context != "" {
		cmdSlice = append(cmdSlice, "--context", lc.context)
	}

	if lc.namespace != "" {
		cmdSlice = append(cmdSlice, "--namespace", lc.namespace)
	}

	if lc.run && lc.dockerRun {
		return "", errors.New("--run and --docker-run are mutually exclusive")
	}

	if lc.run {
		if lc.mount != "" {
			cmdSlice = append(cmdSlice, "--mount", lc.mount)
		}
	}

	if lc.dockerRun {
		if lc.dockerMount != "" {
			cmdSlice = append(cmdSlice, "--docker-mount", lc.dockerMount)
		}
		cmdSlice = append(cmdSlice, "--docker-run")
	}

	if lc.processCmd != "" {
		cmdSlice = append(cmdSlice, "--", lc.processCmd)
	}

	return strings.Join(cmdSlice, " "), nil
}

// translateLegacyCmd tries to detect if a telepresence 1 command was used
// and constructs a telepresence 2 command from that.
func translateLegacyCmd(args []string) (string, []string, error) {
	lc := parseLegacyCommand(args)
	if lc.swapDeployment == "" {
		return "", nil, nil
	}
	tp2Cmd, err := lc.genTP2Command()
	if err != nil {
		return "", nil, nil
	}
	return tp2Cmd, lc.unknownFlags, nil
}

// checkLegacyCmd is mostly a wrapper around translateLegacyCmd. The latter
// is separate to make for easier testing.
func checkLegacyCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	tp2Cmd, unknownFlags, err := translateLegacyCmd(args)
	if err != nil {
		return err
	}
	if len(unknownFlags) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "The following flags used don't have a direct translation to tp2 and haven't been translated: %#v", unknownFlags)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nYou used a telepresence 1 command that roughly translates to the following:\ntelepresence %s\n", tp2Cmd)
	return nil
}
