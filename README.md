<p align="center">
  <img src="design/logo/flock-app-icon.svg" alt="DevEz app icon" width="96" height="96">
</p>

<h1 align="center">DevEz</h1>

<p align="center">
  <strong>Native local PHP & Laravel development environment.</strong><br>
  .test domains with real HTTPS, isolated PHP versions per project, one-click background services, and instant mail & dump catching.<br>
  For macOS, Windows and Linux. No Docker, no virtual machines.
</p>

<p align="center">
  <a href="https://github.com/rabbi696/DevEz"><img src="https://img.shields.io/badge/repo-rabbi696%2FDevEz-blue?style=flat-square" alt="GitHub Repo"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-0D7A8C?style=flat-square" alt="macOS | Windows | Linux">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square" alt="License: AGPL-3.0"></a>
</p>

**DevEz** is a fast, lightweight, and open-source local development environment for PHP and Laravel, operating natively on macOS, Windows, and Linux (similar to Laravel Herd). It runs PHP, Node.js, your databases, and reverse-proxy web servers as native background processes — eliminating the CPU, memory, and filesystem overhead of Docker containers or virtual machines.

> **Note:** DevEz is based on and forked from the open-source [Flock](https://github.com/flock-php/flock) project.

---

## ✨ Features

- **`.test` domains with automatic HTTPS:** Park a projects directory or link an individual project to get immediate `<name>.test` local domains with trusted certificates (no browser security warnings).
- **Multiple PHP versions side by side:** Run PHP 8.1 through 8.5 concurrently. Each project can pin its preferred version via `.devezrc` (or `.flockrc`).
- **Node.js version management:** Automatically respects `.nvmrc` or project configurations.
- **One-click background services:** Native databases and cache services including PostgreSQL, MySQL, MariaDB, MongoDB, Redis, Valkey, Meilisearch, and Typesense.
- **Local mail inbox:** Catches all outgoing emails locally on port 2525 so test emails never reach real recipients.
- **`dump()` & debug timeline:** Captures `dump()` and `dd()` calls, database queries, logs, and rendered views without polluting HTML responses.
- **On-demand Xdebug:** Keep Xdebug available when debugging without slowing down regular development requests.
- **Git worktree support:** Each worktree automatically gets its own isolated domain and database.
- **Public URL sharing:** Instantly share a local site over a temporary public URL via Cloudflare Tunnels (`devez share`).
- **AI Agent integration:** Built-in Model Context Protocol (MCP) server so coding agents (Claude Code, Cursor, Windsurf) can interact with your sites and logs directly.

---

## 🚀 Quick Start (CLI)

```sh
# Park the current projects directory (serves every subfolder at http://<folder>.test)
devez park

# Link a specific directory as a site
devez link my-app

# Secure a site with a trusted local HTTPS certificate
devez secure my-app

# Check status of daemon, sites, and services
devez status

# Install a specific PHP version
devez php install 8.4

# Pin a project's PHP version (saves into .devezrc)
cd my-app
devez pin php 8.4

# Install and start database services
devez services install postgres
devez services install mysql
```

*(Note: For backward compatibility, the `flock` command alias is also supported alongside `devez`.)*

---

## 🛠️ Local Development & Building from Source

### Prerequisites
- **Go:** 1.27+
- **Swift:** 6.0+ (required for macOS menu bar app)
- **Make**

### Build Commands

```sh
# 1. Build the DevEz CLI binary (outputs to bin/devez and symlinks bin/flock)
make build

# 2. Run Go tests
make test

# 3. Run Go vet and format checks
make lint

# 4. Build the macOS Menu Bar App (outputs to build/macos/Flock.app)
SKIP_DMG=1 ./macos/build-app.sh
# or build with disk image (DMG):
make app

# 5. Run Swift tests for the macOS app
cd macos && swift test
```

### Direct Go Execution
To test commands during development without a full rebuild:
```sh
go run ./cmd/devez version
go run ./cmd/devez --help
```

---

## 🔄 Syncing with Upstream

To keep DevEz up to date with updates from upstream Flock:

```sh
# Fetch latest changes from upstream
git fetch upstream

# Merge upstream changes into your main branch
git merge upstream/main

# Re-test and build
make test
make build
```

---

## 👥 Credits & Acknowledgments

DevEz is built upon the foundational work of the **[Flock](https://github.com/flock-php/flock)** project created by **[Guava](https://guava.cz/en)** and its open-source contributors:
- [Guava](https://guava.cz/en)
- [Lukas Frey](https://github.com/lukas-frey)
- [Stepan Mocjak](https://github.com/Uolsen)
- [Marc Beinder](https://github.com/onairmarc)
- [Flock Contributors](https://github.com/flock-php/flock/graphs/contributors)

---

## 📄 License

DevEz is free and open-source software licensed under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0-only). See [NOTICE](NOTICE) for additional permissions.
DevEz is not affiliated with or endorsed by Laravel.
