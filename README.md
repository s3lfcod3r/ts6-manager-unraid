<div align="center">

**TS6 Manager as a single container for Unraid — rebuilt automatically when the original changes.**<br>
*TS6 Manager als Single-Container für Unraid — wird automatisch neu gebaut, wenn sich das Original ändert.*

[![Build](https://github.com/s3lfcod3r/ts6-manager-unraid/actions/workflows/sync-upstream.yml/badge.svg)](https://github.com/s3lfcod3r/ts6-manager-unraid/actions/workflows/sync-upstream.yml)
![Version](https://img.shields.io/badge/version-1.0.0-33A78C)
[![License](https://img.shields.io/badge/license-MIT%20upstream%20%2B%20all%20rights%20reserved-8A9CAA)](LICENSE)
![Backend](https://img.shields.io/badge/backend-Node.js%20%2B%20Express-1F7763)
![Web](https://img.shields.io/badge/web-React%20%2B%20Vite-43D3AD)
![Deploy](https://img.shields.io/badge/deploy-GHCR%20%E2%86%92%20Unraid-6F95AD)

**[🇬🇧 English](#-english) · [🇩🇪 Deutsch](#-deutsch)**

📰 **What's new / Neuigkeiten** → [CHANGELOG.md](CHANGELOG.md)

</div>

---

<a id="-english"></a>

## 🇬🇧 English

Web-based management interface for TeamSpeak servers — virtual servers, channels, clients, permissions, music bots, automated workflows and embeddable server widgets, all from your browser. Built on the **WebQuery HTTP API** (the ServerQuery replacement in modern TeamSpeak builds); Telnet is not used or supported.

> **ℹ️ Unraid variant:** this is **not** the original project but an Unraid adaptation of it. The original **TS6 Manager** is by **[clusterzx](https://github.com/clusterzx)** at **[clusterzx/ts6-manager](https://github.com/clusterzx/ts6-manager)** (MIT licensed). This repository only adds a single-container `Dockerfile.unraid`, an Unraid template, a set of build-time fixes and a workflow that rebuilds the image whenever upstream changes. **No authorship of TS6 Manager itself is claimed here.** The upstream project is marked as AI-assisted.

Part of the **Self** family (SelfMailer, SelfStream, SelfArchiver, SelfDashboard …) — same deploy style (GHCR → Unraid).

> **TL;DR** — TeamSpeak web admin · music bots · flow engine · widgets · one container · auto-rebuild from upstream

### 📑 Table of contents

[Features](#en-features) · [Unraid additions](#en-additions) · [Architecture](#en-architecture) · [Quick start](#en-quick-start) · [Unraid](#en-unraid) · [Configuration](#en-configuration) · [Bot commands](#en-commands) · [Security](#en-security) · [Tech stack](#en-tech-stack) · [Development](#en-development) · [Roadmap](#en-roadmap)

<a id="en-features"></a>

### ✨ Features

**🖥️ Server management**
- **Dashboard** with live server stats, bandwidth graph and capacity overview
- Virtual server list with start/stop controls
- **Channel tree** with drag-and-drop ordering
- Client list with kick, ban, move, poke actions
- Server & channel group management
- **Permission editor** (server, channel, client, group level)
- Ban list, token / privilege key management, complaint viewer, offline messages
- Server log viewer with filtering
- Channel file browser with upload/download
- Instance-level settings

**🎵 Music bots**
- **Multiple bots per server**, each with its own queue and playback
- **Radio streaming** with ICY metadata and live title updates
- **YouTube playback** via yt-dlp (search, download, queue)
- Music library management (upload, organize, playlists)
- Volume, pause, skip, previous, shuffle, repeat
- Stereo audio with stable 20 ms pacing
- Auto-reconnect with exponential backoff
- In-channel **text commands** for hands-free control
- Music request history

**🎬 Video streaming (upstream only, see Unraid additions)**
- Live video from YouTube, Twitch or direct URLs into TeamSpeak channels
- WebRTC via a Go sidecar relay (Pion), quality presets 480p / 720p / 1080p
- In-browser preview, A/V sync via RTCP Sender Reports

**🔀 Bot flow engine**
- **Visual flow editor** with drag-and-drop node canvas
- Triggers: TS3 events, cron schedules, webhooks (with mandatory secrets), chat commands
- Actions: kick, ban, move, message, poke, channel create/edit/delete, HTTP requests, WebQuery commands
- Conditions, variables, delays, loops, logging
- Animated channel names, placeholder system with filters and expressions
- **Pre-built templates** (temporary channels, AFK mover, idle kicker, online counter, group protection)

**🪧 Server widgets**
- Embeddable **server status banner** for websites and forums
- Token-based public access, as live page, SVG or PNG
- Dark and light themes; channel tree and client list can be shown or hidden

**⚙️ Settings & administration**
- yt-dlp cookie file management (age-restricted or member-only YouTube content), upload as file or paste in the UI
- Admin-only settings panel

<a id="en-additions"></a>

### 🧩 Unraid additions

What this repository adds on top of the original:

- **Single container** (`Dockerfile.unraid`): frontend (nginx) and backend (Node.js) in one image, run by supervisord
- **Unraid template** [`unraid/ts6-manager.xml`](unraid/ts6-manager.xml)
- **Auto-rebuild** ([`sync-upstream.yml`](.github/workflows/sync-upstream.yml)): every 6 hours the original is checked; on a new upstream commit or an own push the image is rebuilt and pushed to `ghcr.io/s3lfcod3r/ts6-manager`
- **Build-time fixes** ([`patches/`](patches/README.md)) applied to the fresh upstream clone before every build — if a fix no longer matches, the build fails on purpose:
  1. *Create Channel → Permanent* really creates a permanent channel
  2. *Set Variable* reads the name/value the UI writes
  3. *Store As* keeps single results (e.g. `whoami`)
  4. WebQuery retries once after a dropped connection (`socket hang up`)
  5. *Set Variable* gets an **Operation** selector (`set`, `increment`, `append`)
  6. UI choice *Permanent* no longer writes semi-permanent
  7. **AFK mover** reacts to mute and moves users back
  8. **Rank check** really accumulates online time
  9. Rank check removes the previous rank

> **⚠️ Not in the Unraid image:** the Go **video sidecar** is not part of the single container — video streaming needs the upstream multi-container setup.

> **💡 Tip:** the most common cause of `socket hang up` is TeamSpeak's ServerQuery flood limit (10 commands / 3 s, then 600 s ban). Raise it with `instanceedit` or add the manager's address to `query_ip_allowlist` — details in [`patches/README.md`](patches/README.md).

<a id="en-architecture"></a>

### 🧭 Architecture

```text
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Frontend   │────▶│   Backend    │────▶│  TS Server      │
│  React SPA   │     │  Express API │     │  WebQuery HTTP  │
│  nginx :80   │     │  Node :3001  │     │  SSH (events)   │
└──────────────┘     └──────┬───────┘     └─────────────────┘
                            │
                     ┌──────┴───────┐
                     │   SQLite     │
                     │   (Prisma)   │
                     └──────┬───────┘
                            │
                     ┌──────┴───────┐
                     │   Sidecar    │   (upstream only)
                     │  Go/Pion     │
                     │  WebRTC :9800│
                     └──────────────┘

Public:  /widget/:token  ──▶  SVG / PNG / JSON (no auth)
```

Four packages in a pnpm monorepo:

| Package | Description |
|---|---|
| `@ts6/common` | Shared types, constants, utilities |
| `@ts6/backend` | Express API, WebQuery client, bot engine, voice bots, widgets |
| `@ts6/frontend` | React SPA with Vite, TailwindCSS, shadcn/ui |
| `sidecar` | Go WebRTC media relay (Pion) for video streaming |

The backend proxies all TeamSpeak API calls; the frontend never sees API keys or server credentials.

<a id="en-quick-start"></a>

### 🚀 Quick start (Docker)

Upstream multi-container setup with [`docker-compose.yml`](docker-compose.yml):

```bash
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)" >> .env
docker compose up -d
```

Open `http://localhost:3000/setup` → create your admin account, then add your TeamSpeak server under **Settings → Connections** (host, WebQuery port, API key).

> **⚠️ Required:** `JWT_SECRET` — the backend refuses to start in production without it. `ENCRYPTION_KEY` is optional but recommended; without it `JWT_SECRET` is used for credential encryption.

<details><summary><b>Setup — build from source / Coolify / reverse proxy</b></summary>

Build from source:

```bash
git clone https://github.com/clusterzx/ts6-manager.git
cd ts6-manager
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)" >> .env
docker compose -f docker-compose.local.yml up -d --build
```

Coolify / reverse proxy: start from [`docker-compose.coolify.yml`](docker-compose.coolify.yml). No `ports` section (the proxy routes), set the domain on the **frontend** service (port 80). If the TS server runs in a separate Docker network, add it as an external network on the backend:

```yaml
services:
  backend:
    networks:
      - ts6-network
      - ts-server-net

networks:
  ts-server-net:
    external: true
    name: your-ts-server-network-id
```

</details>

**Requirements:** TeamSpeak server with **WebQuery HTTP** enabled (not raw/telnet) · WebQuery API key (`apikeyadd` or server admin tools) · SSH access to the TS server (only for flow event triggers) · `yt-dlp` and `ffmpeg` (included in the image).

<a id="en-unraid"></a>

### 📦 Unraid

Add the template from
`https://raw.githubusercontent.com/s3lfcod3r/ts6-manager-unraid/main/unraid/ts6-manager.xml`
or import it under *Docker → Add Container → Template*.

Image `ghcr.io/s3lfcod3r/ts6-manager:latest`, WebUI on port `3000`, data (SQLite + music) in `/mnt/user/appdata/ts6-manager/data` → `/data`. After the start open `http://UNRAID-IP:3000/setup` and create the admin account.

> **⚠️ Secrets:** set **JWT Secret** (required) and **Encryption Key** (recommended), each with `openssl rand -base64 32`.

<a id="en-configuration"></a>

### ⚙️ Configuration

Environment variables (no common prefix).

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `JWT_SECRET` | ✅ | – | Secret for JWT signing, at least 32 characters |
| `ENCRYPTION_KEY` | – | `JWT_SECRET` | Dedicated key for AES-256-GCM credential encryption |
| `TS_ALLOW_SELF_SIGNED` | – | `false` | `true` if your TeamSpeak uses a self-signed certificate |
| `PORT` | – | `3001` | Backend port |
| `DATABASE_URL` | – | `file:./data/ts6webui.db` (Unraid: `file:/data/ts6webui.db`) | SQLite database path |
| `JWT_ACCESS_EXPIRY` | – | `15m` | Access token lifetime |
| `JWT_REFRESH_EXPIRY` | – | `7d` | Refresh token lifetime |
| `FRONTEND_URL` | – | `http://localhost:3000` | CORS origin |
| `MUSIC_DIR` | – | `/data/music` | Directory for downloaded music |
| `SIDECAR_URL` | – | – | URL of the WebRTC sidecar (e.g. `http://ts6-sidecar:9800`) |
| `YT_COOKIE_FILE` | – | – | Netscape-format cookies.txt for yt-dlp (also via **Settings → YouTube**) |

<details><summary><b>Setup — sidecar variables (video streaming)</b></summary>

| Variable | Default | Purpose |
|---|---|---|
| `VIDEO_QUEUE_SIZE` | `2048` | Size of the video RTP queue |
| `AUDIO_QUEUE_SIZE` | `4096` | Size of the audio RTP queue |
| `SYNC_PLAYOUT_BUFFER_MS` | `4` | Small playout buffer for adaptive pacing |
| `SYNC_VIDEO_BIAS_MS` | `4` | Optional extra holdback for video to fine-tune sync |
| `AUDIO_DELAY_MS` | `0` | Legacy / manual audio delay, usually stays at 0 |
| `SIDECAR_DEBUG_LOGS` | `1` | Verbose debug logging |
| `VIDEO_READ_RTP_BUFFER` | `4194304` | UDP OS socket buffer for the video port |
| `AUDIO_READ_RTP_BUFFER` | `1048576` | UDP OS socket buffer for the audio port |
| `VIDEO_BUFSIZE` | `1M` | FFmpeg video buffer |

</details>

<a id="en-commands"></a>

### 💬 Bot commands

When a music bot is in a channel, users there can control it via chat:

| Command | Description |
|---|---|
| `!radio` | List radio stations |
| `!radio <id>` | Play a radio station |
| `!play <url>` | Play from a YouTube URL |
| `!play` | Resume paused playback |
| `!stop` | Stop playback |
| `!pause` | Toggle pause/resume |
| `!skip` / `!next` | Next track |
| `!prev` | Previous track |
| `!vol` / `!vol <0-100>` | Show / set volume |
| `!np` | Show current track |

<a id="en-security"></a>

### 🛡️ Security

- **Setup wizard** for the first admin account (no default credentials)
- **AES-256-GCM** encryption for stored credentials (API keys, SSH passwords)
- **SSRF protection** on all outbound HTTP requests and FFmpeg URLs
- Rate limiting on authentication endpoints
- **JWT** access + refresh token rotation with reuse detection
- Role-based access (admin / viewer) and per-server access control
- WebQuery command whitelist in bot flows (blocks destructive commands)
- Authenticated WebSocket connections, password complexity rules
- **Deliberate trade-offs** (documented): the Unraid single container runs nginx and the backend via supervisord **as root**; the image follows upstream automatically, so new upstream code lands without manual review — only the fixes in `patches/` are checked at build time

Security update history → [CHANGELOG.md](CHANGELOG.md) (entries marked 🔒).

<a id="en-tech-stack"></a>

### 🧱 Tech stack

| Part | Tech |
|---|---|
| Frontend | React 18 · Vite · TailwindCSS · shadcn/ui · TanStack Query + Table · React Flow · Recharts · Zustand |
| Backend | Node.js · Express · Prisma (SQLite) · JWT · WebQuery HTTP client · SSH event listener |
| Voice / video | Custom TS3 voice client (UDP) · Opus · FFmpeg · yt-dlp · Go sidecar with Pion WebRTC v4 |
| Deploy | Docker (single container, supervisord) · GHCR · Unraid |

<a id="en-development"></a>

### 🛠️ Development

Requires Node.js 20+ and pnpm 9+.

```bash
pnpm install
pnpm dev
```

Backend on `:3001`, frontend on `:5173` (Vite). Database (Prisma + SQLite) on first run — the Docker images do this automatically:

```bash
cd packages/backend
npx prisma migrate deploy
```

**Check the build-time fixes against a fresh upstream clone:**

```bash
git clone --depth 1 https://github.com/clusterzx/ts6-manager.git /tmp/upstream
node patches/apply-fixes.mjs /tmp/upstream
```

<a id="en-roadmap"></a>

### 🗺️ Roadmap

- [x] Single container for Unraid + template
- [x] Automatic rebuild on upstream changes and own pushes
- [x] Build-time fixes for bot engine, WebQuery, AFK mover and rank check
- [ ] Video sidecar in the Unraid image

<p align="right"><a href="#-english">⬆ back to top</a></p>

---

<a id="-deutsch"></a>

## 🇩🇪 Deutsch

Web-Oberfläche zum Verwalten von TeamSpeak-Servern — virtuelle Server, Channels, Clients, Rechte, Musikbots, automatische Abläufe und einbettbare Server-Widgets, alles im Browser. Baut auf der **WebQuery-HTTP-API** auf (dem ServerQuery-Ersatz in aktuellen TeamSpeak-Versionen); Telnet wird weder genutzt noch unterstützt.

> **ℹ️ Unraid-Variante:** Das ist **nicht** das Originalprojekt, sondern eine Unraid-Anpassung davon. Das ursprüngliche **TS6 Manager** stammt von **[clusterzx](https://github.com/clusterzx)** und liegt unter **[clusterzx/ts6-manager](https://github.com/clusterzx/ts6-manager)** (MIT-Lizenz). Dieses Repository ergänzt nur ein Single-Container-`Dockerfile.unraid`, eine Unraid-Vorlage, einige Korrekturen beim Bauen und einen Workflow, der das Image bei Änderungen am Original automatisch neu baut. **Es wird keine eigene Urheberschaft am TS6 Manager beansprucht.** Das Originalprojekt ist als KI-unterstützt gekennzeichnet.

Teil der **Self**-Familie (SelfMailer, SelfStream, SelfArchiver, SelfDashboard …) — gleicher Deploy-Stil (GHCR → Unraid).

> **Kurz** — TeamSpeak-Weboberfläche · Musikbots · Flow-Engine · Widgets · ein Container · baut automatisch nach dem Original

### 📑 Inhaltsverzeichnis

[Funktionen](#de-funktionen) · [Unraid-Ergänzungen](#de-ergaenzungen) · [Architektur](#de-architektur) · [Schnellstart](#de-schnellstart) · [Unraid](#de-unraid) · [Konfiguration](#de-konfiguration) · [Bot-Befehle](#de-befehle) · [Sicherheit](#de-sicherheit) · [Tech-Stack](#de-tech-stack) · [Entwicklung](#de-entwicklung) · [Roadmap](#de-roadmap)

<a id="de-funktionen"></a>

### ✨ Funktionen

**🖥️ Server-Verwaltung**
- **Dashboard** mit Live-Werten, Bandbreiten-Kurve und Auslastung
- Liste der virtuellen Server mit Start/Stopp
- **Channel-Baum** mit Sortieren per Drag-and-drop
- Client-Liste mit Kicken, Bannen, Verschieben, Anstupsen
- Server- und Channel-Gruppen verwalten
- **Rechte-Editor** (Server, Channel, Client, Gruppe)
- Bannliste, Token-/Berechtigungsschlüssel, Beschwerden, Offline-Nachrichten
- Server-Log mit Filter
- Channel-Dateibrowser mit Hoch- und Herunterladen
- Einstellungen auf Instanz-Ebene

**🎵 Musikbots**
- **Mehrere Bots pro Server**, jeder mit eigener Warteschlange und Wiedergabe
- **Radio-Streams** mit ICY-Metadaten und Live-Titel
- **YouTube-Wiedergabe** über yt-dlp (Suchen, Laden, Einreihen)
- Musikbibliothek (Hochladen, Ordnen, Playlists)
- Lautstärke, Pause, Weiter, Zurück, Zufall, Wiederholen
- Stereo-Ton mit stabilem 20-ms-Takt
- Automatisches Neuverbinden mit wachsender Wartezeit
- **Text-Befehle** im Channel zum Steuern ohne Oberfläche
- Verlauf der Musikwünsche

**🎬 Video-Streaming (nur im Original, siehe Unraid-Ergänzungen)**
- Live-Video von YouTube, Twitch oder direkten URLs in TeamSpeak-Channels
- WebRTC über ein Go-Sidecar (Pion), Qualitätsstufen 480p / 720p / 1080p
- Vorschau im Browser, Bild/Ton-Abgleich über RTCP Sender Reports

**🔀 Bot-Flow-Engine**
- **Visueller Flow-Editor** mit Knoten per Drag-and-drop
- Auslöser: TS3-Ereignisse, Zeitpläne (Cron), Webhooks (mit Pflicht-Secret), Chat-Befehle
- Aktionen: Kicken, Bannen, Verschieben, Nachricht, Anstupsen, Channel anlegen/ändern/löschen, HTTP-Anfragen, WebQuery-Befehle
- Bedingungen, Variablen, Wartezeiten, Schleifen, Protokoll
- Animierte Channel-Namen, Platzhalter mit Filtern und Ausdrücken
- **Fertige Vorlagen** (temporäre Channels, AFK-Mover, Idle-Kicker, Online-Zähler, Gruppenschutz)

**🪧 Server-Widgets**
- Einbettbares **Server-Status-Banner** für Webseiten und Foren
- Öffentlicher Zugriff per Token, als Live-Seite, SVG oder PNG
- Dunkles und helles Design; Channel-Baum und Client-Liste ein- oder ausblendbar

**⚙️ Einstellungen & Verwaltung**
- yt-dlp-Cookie-Datei verwalten (YouTube-Inhalte mit Altersgrenze oder nur für Mitglieder), als Datei hochladen oder in der Oberfläche einfügen
- Einstellungsbereich nur für Admins

<a id="de-ergaenzungen"></a>

### 🧩 Unraid-Ergänzungen

Was dieses Repository zum Original hinzufügt:

- **Ein Container** (`Dockerfile.unraid`): Frontend (nginx) und Backend (Node.js) in einem Image, gestartet über supervisord
- **Unraid-Vorlage** [`unraid/ts6-manager.xml`](unraid/ts6-manager.xml)
- **Automatischer Neubau** ([`sync-upstream.yml`](.github/workflows/sync-upstream.yml)): alle 6 Stunden wird das Original geprüft; bei neuem Commit dort oder eigenem Push wird das Image neu gebaut und nach `ghcr.io/s3lfcod3r/ts6-manager` geschoben
- **Korrekturen beim Bauen** ([`patches/`](patches/README.md)), die vor jedem Bau auf den frisch geholten Original-Code angewendet werden — passt eine nicht mehr, bricht der Bau absichtlich ab:
  1. *Create Channel → Permanent* legt wirklich einen dauerhaften Channel an
  2. *Set Variable* liest Name/Wert so, wie die Oberfläche sie schreibt
  3. *Store As* behält auch Einzelergebnisse (z. B. `whoami`)
  4. WebQuery versucht es nach einer abgerissenen Verbindung (`socket hang up`) einmal neu
  5. *Set Variable* bekommt eine Auswahl **Operation** (`set`, `increment`, `append`)
  6. Die Auswahl *Permanent* schreibt kein semi-permanent mehr
  7. **AFK-Mover** reagiert auf Stummschaltung und holt Nutzer zurück
  8. **Rank Check** sammelt die Online-Zeit wirklich
  9. Rank Check zieht die vorige Stufe ab

> **⚠️ Nicht im Unraid-Image:** Das Go-**Video-Sidecar** ist nicht Teil des Single-Containers — für Video-Streaming braucht es den Mehr-Container-Aufbau des Originals.

> **💡 Tipp:** Der häufigste Grund für `socket hang up` ist die ServerQuery-Flood-Grenze von TeamSpeak (10 Befehle / 3 s, danach 600 s Sperre). Mit `instanceedit` anheben oder die Adresse des Managers in die `query_ip_allowlist` eintragen — Details in [`patches/README.md`](patches/README.md).

<a id="de-architektur"></a>

### 🧭 Architektur

```text
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Frontend   │────▶│   Backend    │────▶│  TS-Server      │
│  React-SPA   │     │  Express-API │     │  WebQuery HTTP  │
│  nginx :80   │     │  Node :3001  │     │  SSH (Events)   │
└──────────────┘     └──────┬───────┘     └─────────────────┘
                            │
                     ┌──────┴───────┐
                     │   SQLite     │
                     │   (Prisma)   │
                     └──────┬───────┘
                            │
                     ┌──────┴───────┐
                     │   Sidecar    │   (nur im Original)
                     │  Go/Pion     │
                     │  WebRTC :9800│
                     └──────────────┘

Öffentlich:  /widget/:token  ──▶  SVG / PNG / JSON (ohne Anmeldung)
```

Vier Pakete in einem pnpm-Monorepo:

| Paket | Beschreibung |
|---|---|
| `@ts6/common` | Gemeinsame Typen, Konstanten, Hilfsfunktionen |
| `@ts6/backend` | Express-API, WebQuery-Client, Bot-Engine, Sprach-Bots, Widgets |
| `@ts6/frontend` | React-SPA mit Vite, TailwindCSS, shadcn/ui |
| `sidecar` | Go-WebRTC-Medien-Relay (Pion) für Video-Streaming |

Das Backend leitet alle TeamSpeak-Aufrufe weiter; das Frontend sieht nie API-Schlüssel oder Server-Zugangsdaten.

<a id="de-schnellstart"></a>

### 🚀 Schnellstart (Docker)

Mehr-Container-Aufbau des Originals mit [`docker-compose.yml`](docker-compose.yml):

```bash
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)" >> .env
docker compose up -d
```

`http://localhost:3000/setup` öffnen → Admin-Konto anlegen, dann den TeamSpeak-Server unter **Settings → Connections** eintragen (Host, WebQuery-Port, API-Schlüssel).

> **⚠️ Pflicht:** `JWT_SECRET` — ohne startet das Backend im Produktivbetrieb nicht. `ENCRYPTION_KEY` ist optional, aber empfohlen; ohne wird `JWT_SECRET` zum Verschlüsseln der Zugangsdaten benutzt.

<details><summary><b>Einrichtung — aus dem Quellcode bauen / Coolify / Reverse-Proxy</b></summary>

Aus dem Quellcode bauen:

```bash
git clone https://github.com/clusterzx/ts6-manager.git
cd ts6-manager
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)" >> .env
docker compose -f docker-compose.local.yml up -d --build
```

Coolify / Reverse-Proxy: [`docker-compose.coolify.yml`](docker-compose.coolify.yml) als Ausgangspunkt nehmen. Kein `ports`-Abschnitt (der Proxy leitet weiter), Domain am **frontend**-Dienst setzen (Port 80). Läuft der TS-Server in einem eigenen Docker-Netz, dieses beim Backend als externes Netz eintragen:

```yaml
services:
  backend:
    networks:
      - ts6-network
      - ts-server-net

networks:
  ts-server-net:
    external: true
    name: your-ts-server-network-id
```

</details>

**Voraussetzungen:** TeamSpeak-Server mit aktiviertem **WebQuery HTTP** (nicht raw/telnet) · WebQuery-API-Schlüssel (`apikeyadd` oder Server-Admin-Tools) · SSH-Zugang zum TS-Server (nur für Flow-Ereignis-Auslöser) · `yt-dlp` und `ffmpeg` (im Image enthalten).

<a id="de-unraid"></a>

### 📦 Unraid

Vorlage von
`https://raw.githubusercontent.com/s3lfcod3r/ts6-manager-unraid/main/unraid/ts6-manager.xml`
hinzufügen oder unter *Docker → Add Container → Template* importieren.

Image `ghcr.io/s3lfcod3r/ts6-manager:latest`, Weboberfläche auf Port `3000`, Daten (SQLite + Musik) in `/mnt/user/appdata/ts6-manager/data` → `/data`. Nach dem Start `http://UNRAID-IP:3000/setup` öffnen und das Admin-Konto anlegen.

> **⚠️ Secrets:** **JWT Secret** (Pflicht) und **Encryption Key** (empfohlen) setzen, jeweils mit `openssl rand -base64 32`.

<a id="de-konfiguration"></a>

### ⚙️ Konfiguration

Umgebungsvariablen (ohne gemeinsames Präfix).

| Variable | Pflicht | Standard | Zweck |
|---|---|---|---|
| `JWT_SECRET` | ✅ | – | Secret zum Signieren der JWTs, mindestens 32 Zeichen |
| `ENCRYPTION_KEY` | – | `JWT_SECRET` | Eigener Schlüssel für die AES-256-GCM-Verschlüsselung der Zugangsdaten |
| `TS_ALLOW_SELF_SIGNED` | – | `false` | `true`, wenn dein TeamSpeak ein selbst signiertes Zertifikat nutzt |
| `PORT` | – | `3001` | Backend-Port |
| `DATABASE_URL` | – | `file:./data/ts6webui.db` (Unraid: `file:/data/ts6webui.db`) | Pfad der SQLite-Datenbank |
| `JWT_ACCESS_EXPIRY` | – | `15m` | Laufzeit des Access-Tokens |
| `JWT_REFRESH_EXPIRY` | – | `7d` | Laufzeit des Refresh-Tokens |
| `FRONTEND_URL` | – | `http://localhost:3000` | CORS-Ursprung |
| `MUSIC_DIR` | – | `/data/music` | Ordner für geladene Musik |
| `SIDECAR_URL` | – | – | URL des WebRTC-Sidecars (z. B. `http://ts6-sidecar:9800`) |
| `YT_COOKIE_FILE` | – | – | cookies.txt im Netscape-Format für yt-dlp (auch über **Settings → YouTube**) |

<details><summary><b>Einrichtung — Sidecar-Variablen (Video-Streaming)</b></summary>

| Variable | Standard | Zweck |
|---|---|---|
| `VIDEO_QUEUE_SIZE` | `2048` | Größe der Video-RTP-Warteschlange |
| `AUDIO_QUEUE_SIZE` | `4096` | Größe der Audio-RTP-Warteschlange |
| `SYNC_PLAYOUT_BUFFER_MS` | `4` | Kleiner Wiedergabepuffer für die anpassbare Taktung |
| `SYNC_VIDEO_BIAS_MS` | `4` | Optionale zusätzliche Verzögerung fürs Video zum Feinabgleich |
| `AUDIO_DELAY_MS` | `0` | Alte / manuelle Ton-Verzögerung, bleibt meist auf 0 |
| `SIDECAR_DEBUG_LOGS` | `1` | Ausführliches Debug-Protokoll |
| `VIDEO_READ_RTP_BUFFER` | `4194304` | UDP-Socket-Puffer des Betriebssystems für den Video-Port |
| `AUDIO_READ_RTP_BUFFER` | `1048576` | UDP-Socket-Puffer des Betriebssystems für den Audio-Port |
| `VIDEO_BUFSIZE` | `1M` | FFmpeg-Videopuffer |

</details>

<a id="de-befehle"></a>

### 💬 Bot-Befehle

Ist ein Musikbot in einem Channel, können ihn die Nutzer dort per Chat steuern:

| Befehl | Beschreibung |
|---|---|
| `!radio` | Radiosender auflisten |
| `!radio <id>` | Radiosender abspielen |
| `!play <url>` | Von einer YouTube-URL abspielen |
| `!play` | Pausierte Wiedergabe fortsetzen |
| `!stop` | Wiedergabe stoppen |
| `!pause` | Pause an/aus |
| `!skip` / `!next` | Nächster Titel |
| `!prev` | Vorheriger Titel |
| `!vol` / `!vol <0-100>` | Lautstärke anzeigen / setzen |
| `!np` | Aktuellen Titel anzeigen |

<a id="de-sicherheit"></a>

### 🛡️ Sicherheit

- **Einrichtungs-Assistent** für das erste Admin-Konto (keine Standard-Zugangsdaten)
- **AES-256-GCM**-Verschlüsselung gespeicherter Zugangsdaten (API-Schlüssel, SSH-Passwörter)
- **SSRF-Schutz** für alle ausgehenden HTTP-Anfragen und FFmpeg-URLs
- Begrenzung der Anmeldeversuche
- **JWT**-Access- und Refresh-Token mit Rotation und Erkennung von Wiederverwendung
- Rollen (Admin / Betrachter) und Zugriff pro Server
- Erlaubnisliste für WebQuery-Befehle in Flows (sperrt zerstörerische Befehle)
- WebSocket nur mit Anmeldung, Mindestanforderungen an Passwörter
- **Bewusste Kompromisse** (dokumentiert): Der Unraid-Single-Container startet nginx und Backend über supervisord **als Root**; das Image folgt dem Original automatisch, neuer Original-Code kommt also ohne Durchsicht von Hand hinein — beim Bauen geprüft werden nur die Korrekturen in `patches/`

Verlauf der Sicherheitsupdates → [CHANGELOG.md](CHANGELOG.md) (Einträge mit 🔒).

<a id="de-tech-stack"></a>

### 🧱 Tech-Stack

| Teil | Technik |
|---|---|
| Frontend | React 18 · Vite · TailwindCSS · shadcn/ui · TanStack Query + Table · React Flow · Recharts · Zustand |
| Backend | Node.js · Express · Prisma (SQLite) · JWT · WebQuery-HTTP-Client · SSH-Ereignis-Listener |
| Sprache / Video | Eigener TS3-Sprach-Client (UDP) · Opus · FFmpeg · yt-dlp · Go-Sidecar mit Pion WebRTC v4 |
| Deploy | Docker (ein Container, supervisord) · GHCR · Unraid |

<a id="de-entwicklung"></a>

### 🛠️ Entwicklung

Braucht Node.js 20+ und pnpm 9+.

```bash
pnpm install
pnpm dev
```

Backend auf `:3001`, Frontend auf `:5173` (Vite). Datenbank (Prisma + SQLite) beim ersten Start — die Docker-Images erledigen das automatisch:

```bash
cd packages/backend
npx prisma migrate deploy
```

**Korrekturen gegen einen frisch geholten Original-Code prüfen:**

```bash
git clone --depth 1 https://github.com/clusterzx/ts6-manager.git /tmp/upstream
node patches/apply-fixes.mjs /tmp/upstream
```

<a id="de-roadmap"></a>

### 🗺️ Roadmap

- [x] Ein Container für Unraid + Vorlage
- [x] Automatischer Neubau bei Änderungen am Original und bei eigenen Pushes
- [x] Korrekturen für Bot-Engine, WebQuery, AFK-Mover und Rank Check
- [ ] Video-Sidecar im Unraid-Image

<p align="right"><a href="#-deutsch">⬆ nach oben</a></p>

---

## ⚖️ Lizenz / License

**🇩🇪 Alle Rechte vorbehalten.** © 2026 s3lfcod3r / SelfCoder

Du darfst die Software **privat und nicht-kommerziell** installieren und benutzen (z. B. im Homelab, über Docker oder Unraid).

**Ohne schriftliche Erlaubnis verboten:** kommerzielle Nutzung · Kopieren · Verändern · Weitergeben · Einbauen in eigene Projekte (auch Closed Source) · Verkaufen.

Das gilt **nur für die Unraid-Ergänzungen** von s3lfcod3r. Der originale **TS6-Manager**-Code von [clusterzx](https://github.com/clusterzx/ts6-manager) bleibt unter seiner **MIT**-Lizenz.

Details: [LICENSE](LICENSE)

**🇬🇧 All rights reserved.** © 2026 s3lfcod3r / SelfCoder

You may install and use the software **privately and non-commercially** (e.g. in your home lab, via Docker or Unraid).

**Not permitted without written permission:** commercial use · copying · modifying · redistributing · incorporating into other projects (including closed source) · selling.

This covers **only the Unraid additions** by s3lfcod3r. The original **TS6 Manager** code by [clusterzx](https://github.com/clusterzx/ts6-manager) stays under its **MIT** license.

Details: [LICENSE](LICENSE)

---

<div align="center">

**TS6 Manager for Unraid** · part of the Self family · Teil der Self-Familie<br>
made with 🎧 for self-hosting

</div>
