package cli

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type dbConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

func newDBCommand() *cobra.Command {
	var appName string

	cmd := &cobra.Command{
		Use:     "db [site]",
		Aliases: []string{"database"},
		Short:   "Open or manage the project database in TablePlus, TablePro, or CLI",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBOpen(cmd, args, appName)
		},
	}

	openCmd := &cobra.Command{
		Use:   "open [site]",
		Short: "Open the database in TablePlus, TablePro, or another GUI client",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBOpen(cmd, args, appName)
		},
	}
	openCmd.Flags().StringVar(&appName, "app", "", "GUI app to open (e.g. TablePlus, TablePro, DBeaver)")

	dumpCmd := &cobra.Command{
		Use:   "dump [filename.sql] [site]",
		Short: "Export the project database to an SQL dump file",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := load()
			if err != nil {
				return err
			}

			siteArgs := []string{}
			filename := "dump.sql"
			if len(args) > 0 {
				if strings.HasSuffix(args[0], ".sql") {
					filename = args[0]
					if len(args) > 1 {
						siteArgs = []string{args[1]}
					}
				} else {
					siteArgs = []string{args[0]}
					if len(args) > 1 {
						filename = args[1]
					}
				}
			}

			site, err := siteArg(s, siteArgs)
			if err != nil {
				return err
			}

			cfg := parseEnvDB(site.Path)
			cmd.Printf("Dumping database %q (%s) to %s…\n", cfg.Database, cfg.Driver, filename)

			var dumpCmd *exec.Cmd
			switch cfg.Driver {
			case "pgsql", "postgres", "postgresql":
				dumpCmd = exec.Command("pg_dump", "-h", cfg.Host, "-p", cfg.Port, "-U", cfg.Username, "-d", cfg.Database, "-f", filename)
				if cfg.Password != "" {
					dumpCmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
				}
			case "mysql", "mariadb":
				dumpCmd = exec.Command("mysqldump", "-h", cfg.Host, "-P", cfg.Port, "-u", cfg.Username, "-p"+cfg.Password, cfg.Database, "-r", filename)
			default:
				return fmt.Errorf("unsupported database driver for dump: %q", cfg.Driver)
			}

			dumpCmd.Stdout, dumpCmd.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
			if err := dumpCmd.Run(); err != nil {
				return fmt.Errorf("dump failed: %w", err)
			}

			cmd.Printf("✓ Database exported successfully to %s\n", filename)
			return nil
		},
	}

	importCmd := &cobra.Command{
		Use:   "import <filename.sql> [site]",
		Short: "Import an SQL dump file into the project database",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]
			siteArgs := []string{}
			if len(args) > 1 {
				siteArgs = []string{args[1]}
			}

			s, err := load()
			if err != nil {
				return err
			}

			site, err := siteArg(s, siteArgs)
			if err != nil {
				return err
			}

			cfg := parseEnvDB(site.Path)
			cmd.Printf("Importing %s into database %q (%s)…\n", filename, cfg.Database, cfg.Driver)

			f, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("read %s: %w", filename, err)
			}
			defer f.Close()

			var impCmd *exec.Cmd
			switch cfg.Driver {
			case "pgsql", "postgres", "postgresql":
				impCmd = exec.Command("psql", "-h", cfg.Host, "-p", cfg.Port, "-U", cfg.Username, "-d", cfg.Database)
				if cfg.Password != "" {
					impCmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
				}
			case "mysql", "mariadb":
				impCmd = exec.Command("mysql", "-h", cfg.Host, "-P", cfg.Port, "-u", cfg.Username, "-p"+cfg.Password, cfg.Database)
			default:
				return fmt.Errorf("unsupported database driver for import: %q", cfg.Driver)
			}

			impCmd.Stdin = f
			impCmd.Stdout, impCmd.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
			if err := impCmd.Run(); err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			cmd.Printf("✓ Database %s imported successfully\n", cfg.Database)
			return nil
		},
	}

	cmd.Flags().StringVar(&appName, "app", "", "GUI app to open (e.g. TablePlus, TablePro, DBeaver)")
	cmd.AddCommand(openCmd, dumpCmd, importCmd)
	return cmd
}

func runDBOpen(cmd *cobra.Command, args []string, appName string) error {
	s, err := load()
	if err != nil {
		return err
	}

	site, err := siteArg(s, args)
	if err != nil {
		return err
	}

	cfg := parseEnvDB(site.Path)
	dbURI := formatDBURI(cfg)

	cmd.Printf("Database: %s (%s on %s:%s)\n", cfg.Database, cfg.Driver, cfg.Host, cfg.Port)
	cmd.Printf("URI:      %s\n", dbURI)

	targetApp := appName
	if targetApp == "" {
		targetApp = detectDBClient()
	}

	if runtime.GOOS == "darwin" && targetApp != "" {
		cmd.Printf("Opening with %s…\n", targetApp)
		c := exec.Command("open", "-a", targetApp, dbURI)
		return c.Start()
	}

	return openURL(dbURI)
}

func detectDBClient() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	candidates := []string{
		"TablePlus",
		"TablePro",
		"DBeaver",
		"DataGrip",
		"Beekeeper Studio",
		"Postico 2",
		"Sequel Ace",
	}
	for _, name := range candidates {
		if _, err := os.Stat("/Applications/" + name + ".app"); err == nil {
			return name
		}
		home, _ := os.UserHomeDir()
		if _, err := os.Stat(filepath.Join(home, "Applications", name+".app")); err == nil {
			return name
		}
	}
	return ""
}

func parseEnvDB(projectDir string) dbConfig {
	cfg := dbConfig{
		Driver:   "pgsql",
		Host:     "127.0.0.1",
		Port:     "5432",
		Database: "postgres",
		Username: "postgres",
	}

	envFile := filepath.Join(projectDir, ".env")
	f, err := os.Open(envFile)
	if err != nil {
		return cfg
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)

		switch k {
		case "DB_CONNECTION":
			cfg.Driver = v
		case "DB_HOST":
			if v != "" {
				cfg.Host = v
			}
		case "DB_PORT":
			if v != "" {
				cfg.Port = v
			}
		case "DB_DATABASE":
			if v != "" {
				cfg.Database = v
			}
		case "DB_USERNAME":
			cfg.Username = v
		case "DB_PASSWORD":
			cfg.Password = v
		}
	}

	if cfg.Driver == "mysql" || cfg.Driver == "mariadb" {
		if cfg.Port == "5432" || cfg.Port == "" {
			cfg.Port = "3306"
		}
		if cfg.Username == "postgres" || cfg.Username == "" {
			cfg.Username = "root"
		}
	}
	return cfg
}

func formatDBURI(cfg dbConfig) string {
	driver := cfg.Driver
	switch driver {
	case "pgsql", "postgres", "postgresql":
		driver = "postgres"
	case "mysql", "mariadb":
		driver = "mysql"
	case "sqlite":
		return "sqlite://" + cfg.Database
	}

	auth := url.QueryEscape(cfg.Username)
	if cfg.Password != "" {
		auth += ":" + url.QueryEscape(cfg.Password)
	}

	return fmt.Sprintf("%s://%s@%s:%s/%s", driver, auth, cfg.Host, cfg.Port, cfg.Database)
}
