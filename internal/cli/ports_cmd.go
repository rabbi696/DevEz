package cli

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"flock/internal/catalog"
)

func newPortsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ports",
		Aliases: []string{"port"},
		Short:   "View and manage DevEz service ports",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tPORT\tDESCRIPTION")
			fmt.Fprintf(w, "http\t:%d\tWeb HTTP\n", s.cfg.Ports.HTTP)
			fmt.Fprintf(w, "https\t:%d\tWeb HTTPS (SSL)\n", s.cfg.Ports.HTTPS)
			fmt.Fprintf(w, "dns\t:%d\tLocal DNS (*.%s)\n", s.cfg.Ports.DNS, s.cfg.TLD)
			fmt.Fprintf(w, "smtp\t:%d\tMail Catcher (SMTP)\n", s.cfg.Ports.SMTP)
			fmt.Fprintf(w, "postgres\t:%d\tPostgreSQL Database\n", s.cfg.Ports.Postgres)
			fmt.Fprintf(w, "dumps\t:%d\tVar Dump Server\n", s.cfg.Ports.Dumps)
			fmt.Fprintf(w, "agent\t:%d\tPHP Event Agent\n", s.cfg.Ports.Agent)
			fmt.Fprintf(w, "api\t:%d\tDaemon API\n", s.cfg.Ports.API)

			for id, extra := range s.cfg.Services.Extras {
				fmt.Fprintf(w, "%s\t:%d\t%s (catalog)\n", id, extra.Port, extra.Version)
			}
			return w.Flush()
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <service> <port>",
		Short: "Change a service's listening port (e.g. `devez port set https 8443`)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := strings.ToLower(args[0])
			port, err := strconv.Atoi(args[1])
			if err != nil || port < 1 || port > 65535 {
				return fmt.Errorf("invalid port %q: must be between 1 and 65535", args[1])
			}

			s, err := load()
			if err != nil {
				return err
			}

			used := s.cfg.UsedPorts(service)
			if holder, taken := used[port]; taken && holder != service {
				return fmt.Errorf("port %d is already assigned to %q", port, holder)
			}

			matched := false
			switch service {
			case "http":
				s.cfg.Ports.HTTP = port
				matched = true
			case "https", "ssl":
				s.cfg.Ports.HTTPS = port
				matched = true
			case "smtp", "mail":
				s.cfg.Ports.SMTP = port
				matched = true
			case "dns":
				s.cfg.Ports.DNS = port
				matched = true
			case "postgres", "pgsql":
				s.cfg.Ports.Postgres = port
				matched = true
			case "dumps":
				s.cfg.Ports.Dumps = port
				matched = true
			case "agent":
				s.cfg.Ports.Agent = port
				matched = true
			case "api":
				s.cfg.Ports.API = port
				matched = true
			default:
				if extra, ok := s.cfg.Services.Extra(service); ok {
					extra.Port = port
					s.cfg.Services.SetExtra(service, extra)
					matched = true
				} else {
					for _, k := range catalog.Kinds(runtime.GOOS) {
						if k.ID == service {
							return fmt.Errorf("service %q is not installed yet; install it first with `devez services install %s`", service, service)
						}
					}
				}
			}

			if !matched {
				return fmt.Errorf("unknown service %q (choose from http, https, smtp, dns, postgres, dumps, agent, api)", service)
			}

			if err := s.commit(cmd.Context(), cmd.OutOrStdout()); err != nil {
				return err
			}

			cmd.Printf("✓ %s port updated to %d\n", service, port)
			return nil
		},
	}

	cmd.AddCommand(setCmd)
	return cmd
}
