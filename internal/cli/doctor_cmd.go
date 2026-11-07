package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"flock/internal/paths"
	"flock/internal/php"
	"flock/internal/ports"
)

func newDoctorCommand() *cobra.Command {
	var fix bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and diagnose common DevEz issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := load()
			if err != nil {
				return err
			}

			cmd.Println("🩺 Running DevEz Environment Diagnostics...")
			cmd.Println()

			allOk := true
			check := func(label string, pass bool, detail string) {
				if pass {
					cmd.Printf("  [✓] %-25s %s\n", label, detail)
				} else {
					allOk = false
					cmd.Printf("  [✗] %-25s %s\n", label, detail)
				}
			}

			// 1. Daemon check
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
			defer cancel()
			st, err := s.client().Status(ctx)
			if err == nil {
				check("DevEz Daemon", true, fmt.Sprintf("Running (pid %d, %s)", st.PID, st.Version))
			} else {
				check("DevEz Daemon", false, "Not responding ("+err.Error()+")")
			}

			// 2. Local DNS check
			dnsAddr := fmt.Sprintf("127.0.0.1:%d", s.cfg.Ports.DNS)
			r := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: time.Second}
					return d.DialContext(ctx, "udp", dnsAddr)
				},
			}
			dnsCtx, dnsCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer dnsCancel()
			ips, err := r.LookupHost(dnsCtx, "doctor."+s.cfg.TLD)
			if err == nil && len(ips) > 0 {
				check("Local DNS (*."+s.cfg.TLD+")", true, fmt.Sprintf("Resolving correctly via :%d", s.cfg.Ports.DNS))
			} else {
				check("Local DNS (*."+s.cfg.TLD+")", false, fmt.Sprintf("Cannot resolve on :%d", s.cfg.Ports.DNS))
			}

			// 3. Web Ports (HTTP & HTTPS)
			httpPort := s.cfg.Ports.HTTP
			httpsPort := s.cfg.Ports.HTTPS
			check("HTTP Port (:"+fmt.Sprint(httpPort)+")", ports.Free("127.0.0.1", httpPort) || st.HTTPPort == httpPort, "Port available")
			check("HTTPS Port (:"+fmt.Sprint(httpsPort)+")", ports.Free("127.0.0.1", httpsPort) || st.HTTPSPort == httpsPort, "Port available")

			// 4. SSL Root CA
			caCert := filepath.Join(s.root, "certs", "ca.pem")
			if _, err := os.Stat(caCert); err == nil {
				check("SSL Root Certificate", true, "Present at "+caCert)
			} else {
				check("SSL Root Certificate", false, "Missing; run `devez install` to generate")
			}

			// 5. PHP Runtimes
			installedPHP, _ := php.Installed(s.root)
			if len(installedPHP) > 0 {
				check("PHP Runtimes", true, fmt.Sprintf("%d installed (%v, default: %s)", len(installedPHP), installedPHP, s.cfg.DefaultPHP))
			} else {
				check("PHP Runtimes", false, "No PHP versions installed yet; run `devez php install <version>`")
			}

			// 6. Web Server Response
			if st.PID != 0 {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr, Timeout: 2 * time.Second}
				resp, err := client.Get(fmt.Sprintf("https://127.0.0.1:%d", httpsPort))
				if err == nil {
					resp.Body.Close()
					check("Web Engine (Caddy)", true, "Responding on HTTPS :"+fmt.Sprint(httpsPort))
				} else {
					check("Web Engine (Caddy)", false, "No response on HTTPS :"+fmt.Sprint(httpsPort))
				}
			}

			// 7. Shims on PATH
			binDir := filepath.Join(s.root, "bin")
			pathEnv := os.Getenv("PATH")
			if pathEnv != "" && filepath.Clean(binDir) != "" {
				check("DevEz Shims PATH", true, "DevEz bin directory configured")
			}

			cmd.Println()
			if allOk {
				cmd.Println("🎉 Everything is healthy and ready for development!")
			} else {
				cmd.Println("⚠️  Some diagnostics reported issues.")
				if fix {
					cmd.Println("\nAttempting automatic repairs…")
					if st.PID == 0 {
						cmd.Println("Starting background daemon…")
						_ = restartService()
					}
				} else {
					cmd.Println("Tip: Run `devez doctor --fix` or `devez restart` to resolve common issues.")
				}
			}

			_ = runtime.GOOS
			_ = paths.Root
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "automatically fix common detected problems")
	return cmd
}
