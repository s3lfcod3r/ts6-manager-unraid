# TS6 Manager – Unraid Fork

Dieses Repository ist ein automatisch synchronisierter Fork von [clusterzx/ts6-manager](https://github.com/clusterzx/ts6-manager) mit zusätzlichen Dateien für **Unraid** und **automatischen Docker-Image-Builds**.

## 🔄 Automatische Updates

Ein GitHub Actions Workflow prüft alle **6 Stunden**, ob es neue Commits im Original-Repo gibt. Wenn ja:
1. Werden die Änderungen automatisch in dieses Repo gemergt
2. Wird ein neues Docker Image gebaut und auf `ghcr.io` gepusht

Du musst also **nichts manuell tun** – dein Unraid-Container wird beim nächsten Pull automatisch aktuell sein.

## 🐳 Unraid Setup

### Schritt 1: Docker Images verfügbar machen

Nach dem ersten automatischen Build findest du die Images unter:
- `ghcr.io/kabelsalatundklartext/ts6-manager-backend:latest`
- `ghcr.io/kabelsalatundklartext/ts6-manager-frontend:latest`

### Schritt 2: Unraid Community Applications Templates

Kopiere die URLs der XML-Templates in Unraid → Apps → Template Repositories:
```
https://raw.githubusercontent.com/kabelsalatundklartext/ts6-manager/main/unraid/
```

Oder importiere die XMLs direkt über **Add Container → Template** in Unraid.

### Schritt 3: Container konfigurieren

**Backend zuerst starten**, dann Frontend.

Benötigte Werte für das Backend:
- `JWT_SECRET` – Generieren mit: `openssl rand -base64 32`
- `ENCRYPTION_KEY` – Generieren mit: `openssl rand -base64 32`

### Schritt 4: Erstkonfiguration

Öffne `http://UNRAID-IP:3000/setup` und lege deinen Admin-Account an.

Danach unter **Settings → Connections** deinen TeamSpeak Server eintragen (Host, WebQuery Port, API Key).

## 📁 Zusätzliche Dateien in diesem Fork

| Datei | Beschreibung |
|-------|-------------|
| `.github/workflows/sync-upstream.yml` | Auto-Sync + Docker Build |
| `docker-compose.unraid.yml` | Docker Compose für Unraid |
| `unraid/ts6-manager-backend.xml` | Unraid CA Template (Backend) |
| `unraid/ts6-manager-frontend.xml` | Unraid CA Template (Frontend) |

## ⚙️ Upstream

Original-Projekt: https://github.com/clusterzx/ts6-manager  
Lizenz: MIT
