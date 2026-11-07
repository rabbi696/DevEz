package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newQueueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue [site]",
		Short: "Manage background Laravel queue workers (Herd Pro feature)",
		Long: "Runs and supervises `php artisan queue:work` in the background for your Laravel sites.\n" +
			"Automatically restarts workers on failure and logs output to the log directory.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runQueueStatus(cmd)
			}
			return runQueueToggle(cmd, args[0])
		},
	}

	startCmd := &cobra.Command{
		Use:   "start [site]",
		Short: "Start the background queue worker for a site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setQueueState(cmd, args, true)
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop [site]",
		Short: "Stop the background queue worker for a site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setQueueState(cmd, args, false)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "List all sites and queue worker states",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runQueueStatus(cmd)
		},
	}

	logsCmd := &cobra.Command{
		Use:   "logs [site]",
		Short: "View queue worker logs for a site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}
			site, err := siteArg(s, args)
			if err != nil {
				return err
			}
			logFile := filepath.Join(s.root, "log", "queue-"+site.Name+".log")
			data, err := os.ReadFile(logFile)
			if err != nil {
				cmd.Printf("No logs yet for queue-%s (at %s)\n", site.Name, logFile)
				return nil
			}
			cmd.Print(string(data))
			return nil
		},
	}

	cmd.AddCommand(startCmd, stopCmd, statusCmd, logsCmd)
	return cmd
}

func setQueueState(cmd *cobra.Command, args []string, on bool) error {
	s, err := load()
	if err != nil {
		return err
	}

	site, err := siteArg(s, args)
	if err != nil {
		return err
	}

	found := false
	for i := range s.cfg.Sites {
		if s.cfg.Sites[i].Name == site.Name {
			s.cfg.Sites[i].Queue = on
			found = true
			break
		}
	}
	if !found {
		// Parked site being configured: add explicit site entry
		site.Queue = on
		s.cfg.Sites = append(s.cfg.Sites, site)
	}

	if err := s.commit(cmd.Context(), cmd.OutOrStdout()); err != nil {
		return err
	}

	stateStr := "started"
	if !on {
		stateStr = "stopped"
	}
	cmd.Printf("✓ Queue worker for %s %s\n", site.Name, stateStr)
	return nil
}

func runQueueToggle(cmd *cobra.Command, siteName string) error {
	s, err := load()
	if err != nil {
		return err
	}
	site, err := siteArg(s, []string{siteName})
	if err != nil {
		return err
	}
	return setQueueState(cmd, []string{site.Name}, !site.Queue)
}

func runQueueStatus(cmd *cobra.Command) error {
	s, err := load()
	if err != nil {
		return err
	}

	all := s.cfg.AllSites()
	if len(all) == 0 {
		cmd.Println("No sites linked or parked.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SITE\tQUEUE WORKER\tLOG FILE")
	for _, site := range all {
		status := "stopped"
		if site.Queue {
			status = "running (background)"
		}
		log := filepath.Join("log", "queue-"+site.Name+".log")
		fmt.Fprintf(w, "%s\t%s\t%s\n", site.Name, status, log)
	}
	return w.Flush()
}
