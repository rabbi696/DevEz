package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"flock/internal/share"
)

func newTunnelCommand() *cobra.Command {
	var allow []string
	var debug bool

	cmd := &cobra.Command{
		Use:     "tunnel [site]",
		Aliases: []string{"tun"},
		Short:   "Toggle Cloudflare tunnel on/off in background for individual sites",
		Long: "Controls background Cloudflare quick tunnels (trycloudflare.com) per site.\n" +
			"Runs rootless in the background without locking your terminal.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runTunnelStatus(cmd)
			}
			return runTunnelToggle(cmd, args[0], allow, debug)
		},
	}

	onCmd := &cobra.Command{
		Use:     "on [site]",
		Aliases: []string{"start", "enable"},
		Short:   "Start a background Cloudflare tunnel for a site",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}
			site, err := siteArg(s, args)
			if err != nil {
				return err
			}
			return startBackgroundTunnel(cmd, s, site.Name, allow, debug)
		},
	}
	onCmd.Flags().StringArrayVar(&allow, "allow", nil, "only let in this email address, or everyone at a domain with *@domain (repeatable)")
	onCmd.Flags().BoolVar(&debug, "debug", false, "turn the app's debug mode on for visitors: they see stack traces")

	offCmd := &cobra.Command{
		Use:     "off [site]",
		Aliases: []string{"stop", "disable"},
		Short:   "Stop the background Cloudflare tunnel for a site",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}
			site, err := siteArg(s, args)
			if err != nil {
				return err
			}
			return stopBackgroundTunnel(cmd, s, site.Name)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "List all active background tunnels",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTunnelStatus(cmd)
		},
	}

	cmd.Flags().StringArrayVar(&allow, "allow", nil, "only let in this email address, or everyone at a domain with *@domain (repeatable)")
	cmd.Flags().BoolVar(&debug, "debug", false, "turn the app's debug mode on for visitors")
	cmd.AddCommand(onCmd, offCmd, statusCmd)
	return cmd
}

func startBackgroundTunnel(cmd *cobra.Command, s *state, siteName string, allow []string, debug bool) error {
	ctx := cmd.Context()
	client := s.client()

	cmd.Printf("Starting Cloudflare tunnel for %s in background…\n", siteName)
	if _, err := client.StartShare(ctx, siteName, allow, debug); err != nil {
		return fmt.Errorf("start tunnel: %w", err)
	}

	// Wait up to 15 seconds for tunnel to become active and report public URL
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		st, err := client.Share(ctx, siteName)
		if err == nil {
			if st.State == share.Running && st.URL != "" {
				_ = copyToClipboard(st.URL)
				cmd.Printf("\n✓ Cloudflare tunnel active!\n")
				cmd.Printf("  Public URL: %s  (copied to clipboard)\n", st.URL)
				if len(st.Allow) > 0 {
					cmd.Printf("  Allowed:    %s\n", strings.Join(st.Allow, ", "))
				}
				cmd.Printf("  Site:       %s\n", siteName)
				cmd.Printf("  To stop:    devez tunnel off %s\n\n", siteName)
				return nil
			}
			if st.State == share.Failed {
				return fmt.Errorf("tunnel failed: %s", st.Error)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	cmd.Printf("Tunnel is initializing in background. Check URL with `devez tunnel status`.\n")
	return nil
}

func stopBackgroundTunnel(cmd *cobra.Command, s *state, siteName string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	if err := s.client().StopShare(ctx, siteName); err != nil {
		return fmt.Errorf("stop tunnel: %w", err)
	}
	cmd.Printf("✓ Cloudflare tunnel for %s stopped\n", siteName)
	return nil
}

func runTunnelToggle(cmd *cobra.Command, siteName string, allow []string, debug bool) error {
	s, err := load()
	if err != nil {
		return err
	}
	site, err := siteArg(s, []string{siteName})
	if err != nil {
		return err
	}

	st, err := s.client().Share(cmd.Context(), site.Name)
	if err == nil && st.State == share.Running {
		return stopBackgroundTunnel(cmd, s, site.Name)
	}
	return startBackgroundTunnel(cmd, s, site.Name, allow, debug)
}

func runTunnelStatus(cmd *cobra.Command) error {
	s, err := load()
	if err != nil {
		return err
	}

	shares, err := s.client().Shares(cmd.Context())
	if err != nil {
		cmd.Println("No active tunnels (daemon not running or no shares).")
		return nil
	}

	if len(shares) == 0 {
		cmd.Println("No active Cloudflare tunnels. Start one with `devez tunnel on <site>`.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SITE\tSTATUS\tPUBLIC URL")
	for _, sh := range shares {
		urlStr := sh.URL
		if urlStr == "" {
			urlStr = "(" + string(sh.State) + ")"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", sh.Site, sh.State, urlStr)
	}
	return w.Flush()
}
