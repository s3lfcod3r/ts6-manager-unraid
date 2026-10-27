# 📰 Changelog

Newest version first. Changes to the original TS6 Manager are listed at [clusterzx/ts6-manager](https://github.com/clusterzx/ts6-manager/commits/main).
Neueste Version zuerst. Änderungen am Original-TS6-Manager stehen bei [clusterzx/ts6-manager](https://github.com/clusterzx/ts6-manager/commits/main).

---

## 1.1.0
- 🇬🇧 Security: image builds on **Node 22**, refreshes upstream packages within their ranges on every build, and adds two build-time fixes — **expr-eval → expr-eval-fork** (critical code-execution hole) and **multer 2** for uploads.
- 🇩🇪 Sicherheit: Image baut auf **Node 22**, frischt die Pakete des Originals bei jedem Bau innerhalb ihrer Versionsbereiche auf und bekommt zwei neue Korrekturen — **expr-eval → expr-eval-fork** (kritische Lücke zur Code-Ausführung) und **multer 2** für Uploads.
- 📄 [patches/README.md](patches/README.md)

## 1.0.0
- 🇬🇧 Unraid variant: **single container** (nginx + backend via supervisord), Unraid template, **automatic rebuild** on upstream changes and own pushes, and **9 build-time fixes** (permanent channels, Set Variable incl. operation selector, Store As, WebQuery retry, AFK mover on mute, rank check online time and rank removal); README in Self standard.
- 🇩🇪 Unraid-Variante: **ein Container** (nginx + Backend über supervisord), Unraid-Vorlage, **automatischer Neubau** bei Änderungen am Original und eigenen Pushes sowie **9 Korrekturen beim Bauen** (dauerhafte Channels, Set Variable inkl. Rechenart-Auswahl, Store As, WebQuery-Wiederholung, AFK-Mover bei Stummschaltung, Rank Check Online-Zeit und Stufen-Abzug); README nach Self-Standard.
- 📄 [patches/README.md](patches/README.md)
