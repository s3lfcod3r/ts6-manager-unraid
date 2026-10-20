# AGENTS.md — ts6-manager-unraid

Kurze Orientierung fuer KI-Assistenten (Bonsai/Claude). **Diese Datei zuerst lesen, nicht die grosse README.**
Zum Finden: `rg -n "BEGRIFF"`. Fuer Symbole: Serena `get_symbols_overview`/`find_symbol`. Grosse Dateien nur zeilenweise (`sed -n`).

## Typ & Stack

- Node/JS-TS: `package.json`

## Bauen / Testen / Starten

```bash
npm run dev   # pnpm --parallel -r run dev
npm run build   # pnpm -r run build
npm run lint   # pnpm -r run lint
```

## Struktur (oberste Ebenen)

- `.github/` — CI/Workflows
- `bin/`
- `docs/` — Dokumentation
- `packages/`
- `patches/`
- `unraid/`

---
*Automatisch erzeugt als Navigationshilfe. Bei Bedarf von Hand praeziser machen.*
