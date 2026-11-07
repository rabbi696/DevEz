package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func newOpenCommand() *cobra.Command {
	var openHTTP bool
	var openMail bool

	cmd := &cobra.Command{
		Use:   "open [site]",
		Short: "Open a site or service in your default browser",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}

			if openMail {
				url := fmt.Sprintf("http://127.0.0.1:%d/mail", s.cfg.Ports.HTTP)
				cmd.Printf("Opening mail inbox: %s\n", url)
				return openURL(url)
			}

			site, err := siteArg(s, args)
			if err != nil {
				return err
			}

			scheme := "https"
			port := s.cfg.Ports.HTTPS
			if openHTTP || !site.Secured {
				scheme = "http"
				port = s.cfg.Ports.HTTP
			}

			domain := s.cfg.Domain(site)
			targetURL := fmt.Sprintf("%s://%s", scheme, domain)
			if (scheme == "https" && port != 443) || (scheme == "http" && port != 80) {
				targetURL = fmt.Sprintf("%s://%s:%d", scheme, domain, port)
			}

			cmd.Printf("Opening %s\n", targetURL)
			return openURL(targetURL)
		},
	}

	cmd.Flags().BoolVar(&openHTTP, "http", false, "open over plain HTTP instead of HTTPS")
	cmd.Flags().BoolVar(&openMail, "mail", false, "open DevEz mail viewer")
	return cmd
}

func newCodeCommand() *cobra.Command {
	var editor string

	cmd := &cobra.Command{
		Use:   "code [site]",
		Short: "Open a site's project folder in your code editor",
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

			targetEditor := editor
			if targetEditor == "" {
				targetEditor = detectEditor()
			}

			cmd.Printf("Opening %s in %s…\n", site.Path, targetEditor)
			c := exec.Command(targetEditor, site.Path)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			return c.Start()
		},
	}

	cmd.Flags().StringVar(&editor, "editor", "", "code editor command (e.g. code, cursor, pstorm, subl)")
	return cmd
}

func detectEditor() string {
	if env := os.Getenv("EDITOR"); env != "" {
		return env
	}
	for _, candidate := range []string{"cursor", "code", "pstorm", "subl", "zed"} {
		if path, err := exec.LookPath(candidate); err == nil && path != "" {
			return candidate
		}
	}
	if runtime.GOOS == "darwin" {
		return "open"
	}
	return "xdg-open"
}

func openURL(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("cmd", "/c", "start", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	return c.Start()
}
