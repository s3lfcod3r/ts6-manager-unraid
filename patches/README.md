# Korrekturen am Original

Der Bau klont `clusterzx/ts6-manager` bei jedem Lauf frisch. Geänderte Quelldateien
im Repo würden dabei verlorengehen. Die Korrekturen werden deshalb von
`apply-fixes.mjs` auf den frisch geklonten Baum angewendet, direkt bevor das Image
gebaut wird — so überleben sie jede Synchronisierung.

Passt eine Fundstelle nicht mehr, weil das Original sich geändert hat, **bricht der
Bau ab**. Das ist Absicht: lieber ein roter Bau als ein Image, in dem still eine
Korrektur fehlt.

## Was korrigiert wird

### 1. „Create Channel → Permanent" legt einen temporären Channel an

`packages/backend/src/bot-engine/engine.ts`

Die Oberfläche bietet nur *Permanent* und *Temporary* an, schreibt bei *Permanent*
aber `channel_flag_semi_permanent = '1'`. TeamSpeak löscht semi-permanente Channels,
sobald der letzte Client sie verlässt. Automatisch angelegte Channels verschwinden
dadurch sofort wieder und werden endlos neu erzeugt.

Ohne jedes Flag ist der Channel ebenfalls temporär — entgegen dem Kommentar im
Original („channel will be permanent (TS3 default)"). Gegenprobe am Server:

```
channelcreate ohne Flag  ->  channel_flag_permanent = 0
```

Nach der Korrektur wird alles außer *Temporary* mit `channel_flag_permanent = '1'`
angelegt.

### 2. „Set Variable" ist wirkungslos

`packages/backend/src/bot-engine/engine.ts`

Die Oberfläche speichert `varName` / `varValue`, das Backend liest `name` / `value`.
Der Knoten legt dadurch immer eine leere Variable an. Damit ist auch `increment`
unbrauchbar — also jede Automatik, die zählen muss.

Nach der Korrektur werden beide Schreibweisen gelesen, alte Flows bleiben gültig.

### 3. „Store As" verliert Einzelergebnisse

`packages/backend/src/bot-engine/flow-runner.ts`

Es wurde immer `result[0]` abgelegt. Befehle, die ein einzelnes Objekt liefern
(z. B. `whoami`), kamen deshalb nie an. Arrays verhalten sich unverändert.

### 4. WebQuery gibt bei abgerissener Verbindung sofort auf

`packages/backend/src/ts-client/webquery-client.ts`

Der Client hält eine einzige dauerhafte Verbindung (`keepAlive`, `maxSockets: 1`),
wiederholt aber nie. Wirft der TeamSpeak-Server sie weg — etwa nach einer
ServerQuery-Flood-Sperre oder bei einem veralteten Socket — schlägt jede weitere
Abfrage mit `TSApiError: socket hang up` fehl, bis der Container neu startet.

Nach der Korrektur wird eine abgerissene Verbindung genau einmal neu aufgebaut.

### 5. Auswahl „Permanent" schreibt kein semi-permanent mehr

`packages/frontend/src/pages/BotEditor.tsx`

Gegenstück zu Korrektur 1 auf der Oberflächenseite.

## Nebenbei: ServerQuery-Flood-Grenzen

Kein Programmfehler, aber der häufigste Grund für `socket hang up`: Der TeamSpeak-Server
erlaubt ab Werk **10 Befehle pro 3 Sekunden** und sperrt danach **600 Sekunden**.
Mehrere Flows im kurzen Takt reißen das sofort und halten sich selbst in der Sperre.

```
serverinstance_serverquery_flood_commands          10  ->  200
serverinstance_serverquery_ban_time               600  ->   60
serverinstance_serverquery_max_connections_per_ip   5  ->   20
```

Setzen mit `instanceedit`, oder die Adresse des Managers in die
`query_ip_allowlist` des Servers aufnehmen.

## Selbst prüfen

```bash
git clone --depth 1 https://github.com/clusterzx/ts6-manager.git /tmp/upstream
node patches/apply-fixes.mjs /tmp/upstream
```

Das Skript ist wiederholbar: Ein zweiter Lauf meldet alles als „schon ok" und
ändert nichts.
