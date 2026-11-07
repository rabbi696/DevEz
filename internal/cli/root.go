// Package cli implements the `devez` command.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time: go build -ldflags "-X flock/internal/cli.Version=1.2.3".
var Version = "dev"

// NewRootCommand builds the `devez` command tree.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "devez",
		Aliases:       []string{"flock"},
		Short:         "DevEz — a native local Laravel development environment",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	// cmd.Print* defaults to stderr in cobra; regular output belongs on stdout.
	root.SetOut(os.Stdout)

	root.AddCommand(
		newInstallCommand(), newConflictsCommand(), newUninstallCommand(), newPrivilegedSetupCommand(),
		newDaemonCommand(), newRestartCommand(), newStatusCommand(), newLogsCommand(), newVersionCommand(), newUpdateCommand(),
		newSitesCommand(), newLinkCommand(), newUnlinkCommand(), newParkCommand(), newUnparkCommand(), newProxyCommand(), newUnproxyCommand(), newProxiesCommand(),
		newSecureCommand(true), newSecureCommand(false), newDomainCommand(), newWorktreeCommand(), newShareCommand(), newUseCommand(),
		newIsolateCommand(), newUnisolateCommand(), newIsolatedCommand(), newPinCommand(), newUnpinCommand(), newRootCommand(),
		newPHPCommand(), newNodeCommand(), newServicesCommand(), newXdebugCommand(), newDebugCommand(), newFrankenPHPCommand(), newDumpsCommand(), newEventsCommand(), newMailCommand(),
		newExecCommand(), newMCPCommand(), newBoostCommand(), newChannelCommand(), newExtCommand(),
		newPortsCommand(), newOpenCommand(), newCodeCommand(), newDoctorCommand(), newDBCommand(), newQueueCommand(), newScheduleCommand(), newTunnelCommand(),
	)

	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the DevEz version",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("devez " + Version)
		},
	}
}
