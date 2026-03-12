# dennis

A self-hosted DNS filtering server with a web management UI. Assign filtering profiles to devices on your network by MAC address — block ads, social media, adult content, or lock IoT devices to an allowlist.

Built with Go and React, designed to run on a Raspberry Pi.

Named after Dennis the Menace. Built to block content for my in-house menace.

## Features

- **Per-device filtering** — map MAC addresses to profiles via ARP resolution
- **Block & allowlist rules** — wildcard domain patterns (`**.tiktok.com`)
- **DNS query logs** — full history of every query with client, profile, and action
- **Web UI** — manage profiles, devices, and settings from a browser
- **Single binary** — SQLite database, no external dependencies

## Architecture

```
┌─────────────┐     DNS queries      ┌───────────────────┐
│  LAN device │ ──────────────────▶  │   DNS server :53  │
└─────────────┘                      │                   │
                                     │   Go binary       │
┌─────────────┐     /api/*           │                   │
│  Browser    │ ──────────────────▶  │   HTTP API :9090  │
└─────────────┘         ▲            └────────┬──────────┘
                        │                     │
               ┌────────┴───────┐    ┌────────▼──────────┐
               │  nginx :80     │    │   SQLite (volume) │
               │  React SPA     │    └──────────────────-┘
               └────────────────┘
```

## Quick start (Docker)

```bash
# Clone the repo onto your Raspberry Pi (or cross-build on your machine)
git clone https://github.com/jmercereau/dennis-go
cd dennis-go

# Tell Docker which LAN IP to bind the DNS port to (see note below)
echo "PI_IP=$(hostname -I | awk '{print $1}')" > .env

# (Optional) seed the database from the example config
docker compose run --rm backend ./dns-filter -db /data/dns.db -seed /app/config.yaml

# Start everything
docker compose up -d
```

The web UI will be available at `http://<pi-ip>`.

### Port 53 and local DNS

Port 53 is bound to the Pi's LAN IP only (not `0.0.0.0`), set via `PI_IP` in `.env`. This avoids conflicting with any local DNS process already running on the loopback interface (e.g. `dnsmasq` spawned by NetworkManager, or `systemd-resolved`).

To use Dennis as the DNS server for your whole network, point your **router's DHCP DNS setting** to the Pi's LAN IP. Devices will then receive it automatically on lease renewal.

## Manual build

**Prerequisites:** Go 1.25+, Node 20+

```bash
# Build frontend + backend
make build

# Run (creates dns.db on first start)
./dns-filter -db dns.db -api :9090
```

### CLI flags

| Flag | Default | Description |
|---|---|---|
| `-db` | `dns.db` | Path to the SQLite database |
| `-api` | `:9090` | Address for the HTTP management API |
| `-dns-only` | `false` | Run only the DNS server, skip the HTTP API |
| `-seed <file>` | — | Seed the database from a YAML config file, then exit |
| `-export <file>` | — | Export the database to a YAML config file, then exit |

## Configuration

On first run the database is empty. You can either configure everything through the UI or seed from a YAML file:

```bash
./dns-filter -db dns.db -seed config.yaml
```

See [config.yaml](config.yaml) for a full example covering profiles (parental controls, ad blocking, IoT allowlists) and device assignments.

### Profile rules

```yaml
profiles:
  - name: kids
    block:               # deny these domains, allow everything else
      - "**.tiktok.com"
      - "**.instagram.com"

  - name: iot
    allow_only:          # deny everything except these domains
      - "**.googleapis.com"
      - "**.ntp.org"
```

Patterns use `**` as a wildcard that matches any number of subdomain segments.

### Exporting current config

```bash
./dns-filter -db dns.db -export backup.yaml
```

## Deploying on Raspberry Pi

### Pi 4 / Pi 5 (arm64)

The `docker-compose.yml` builds natively on the Pi. Copy the repo, create a `.env` file, and run:

```bash
echo "PI_IP=$(hostname -I | awk '{print $1}')" > .env
docker compose up -d
```

### Cross-building from a Mac / x86 machine

```bash
docker buildx build --platform linux/arm64 -t youruser/dennis-backend --push .
docker buildx build --platform linux/arm64 -t youruser/dennis-frontend --push ./frontend
```

Then on the Pi, update `docker-compose.yml` to use the image names instead of `build:` blocks.

### Pi 3 / 32-bit OS (arm/v7)

Replace `linux/arm64` with `linux/arm/v7` in the build commands above.

## API

The HTTP API is served at `:9090`. All routes are under `/api/`.

| Method | Path | Description |
|---|---|---|
| GET/PUT | `/api/settings` | Server settings (listen address, default profile) |
| GET/PUT | `/api/upstreams` | Upstream DNS servers |
| GET/POST | `/api/profiles` | List / create profiles |
| GET/PUT/DELETE | `/api/profiles/{name}` | Manage a specific profile |
| GET/POST | `/api/devices` | List / add devices |
| GET/PUT/DELETE | `/api/devices/{mac}` | Manage a specific device |
| GET | `/api/logs` | Query DNS logs |
| GET | `/api/seen-devices` | Devices seen in logs but not yet assigned |
