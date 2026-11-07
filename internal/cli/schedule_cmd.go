package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newScheduleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule [site]",
		Short: "Manage background Laravel task schedulers (Herd Pro feature)",
		Long: "Runs and supervises `php artisan schedule:work` in the background for your Laravel sites.\n" +
			"Enables automated testing of scheduled commands without cron configuration.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runScheduleStatus(cmd)
			}
			return runScheduleToggle(cmd, args[0])
		},
	}

	startCmd := &cobra.Command{
		Use:   "start [site]",
		Short: "Start the background task scheduler for a site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setScheduleState(cmd, args, true)
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop [site]",
		Short: "Stop the background task scheduler for a site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setScheduleState(cmd, args, false)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "List all sites and scheduler states",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runScheduleStatus(cmd)
		},
	}

	logsCmd := &cobra.Command{
		Use:   "logs [site]",
		Short: "View scheduler logs for a site",
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
			logFile := filepath.Join(s.root, "log", "schedule-"+site.Name+".log")
			data, err := os.ReadFile(logFile)
			if err != nil {
				cmd.Printf("No logs yet for schedule-%s (at %s)\n", site.Name, logFile)
				return nil
			}
			cmd.Print(string(data))
			return nil
		},
	}

	cmd.AddCommand(startCmd, stopCmd, statusCmd, logsCmd)
	return cmd
}

func setScheduleState(cmd *cobra.Command, args []string, on bool) error {
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
			s.cfg.Sites[i].Schedule = on
			found = true
			break
		}
	}
	if !found {
		site.Schedule = on
		s.cfg.Sites = append(s.cfg.Sites, site)
	}

	if err := s.commit(cmd.Context(), cmd.OutOrStdout()); err != nil {
		return err
	}

	stateStr := "started"
	if !on {
		stateStr = "stopped"
	}
	cmd.Printf("✓ Task scheduler for %s %s\n", site.Name, stateStr)
	return nil
}

func runScheduleToggle(cmd *cobra.Command, siteName string) error {
	s, err := load()
	if err != nil {
		return err
	}
	site, err := siteArg(s, []string{siteName})
	if err != nil {
		return err
	}
	return setScheduleState(cmd, []string{site.Name}, !site.Schedule)
}

func runScheduleStatus(cmd *cobra.Command) error {
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
	fmt.Fprintln(w, "SITE\tSCHEDULER\tLOG FILE")
	for _, site := range all {
		status := "stopped"
		if site.Schedule {
			status = "running (background)"
		}
		log := filepath.Join("log", "schedule-"+site.Name+".log")
		fmt.Fprintf(w, "%s\t%s\t%s\n", site.Name, status, log)
	}
	return w.Flush()
}
