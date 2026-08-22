#!/usr/bin/env node
/**
 * Korrekturen am Quelltext von clusterzx/ts6-manager, bevor das Image gebaut wird.
 *
 * Der Bau holt das Original bei jedem Lauf frisch. Die Korrekturen duerfen deshalb
 * nicht im Repo als geaenderte Quelldateien liegen - sie werden hier auf den frisch
 * geklonten Baum angewendet. So ueberlebt jede Korrektur die Synchronisierung.
 *
 * Aufruf:  node patches/apply-fixes.mjs <pfad-zum-upstream-baum>
 *
 * Bricht mit Rueckgabewert 1 ab, sobald eine Fundstelle nicht mehr passt. Das ist
 * Absicht: lieber ein roter Bau als ein Image, in dem still eine Korrektur fehlt.
 */

import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';

const wurzel = process.argv[2];
if (!wurzel) {
  console.error('Aufruf: node patches/apply-fixes.mjs <pfad-zum-upstream-baum>');
  process.exit(1);
}

const ENGINE = 'packages/backend/src/bot-engine/engine.ts';
const RUNNER = 'packages/backend/src/bot-engine/flow-runner.ts';
const CLIENT = 'packages/backend/src/ts-client/webquery-client.ts';
const EDITOR = 'packages/frontend/src/pages/BotEditor.tsx';

/**
 * Eine Korrektur.
 *  titel    – erscheint im Bauprotokoll
 *  fertig   – trifft zu, wenn die Korrektur bereits drin ist (Original hat sie behoben)
 *  suchen   – Fundstelle im Original
 *  ersetzen – Ersatz
 */
const KORREKTUREN = [
  {
    datei: ENGINE,
    titel: 'Create Channel legt einen dauerhaften Channel an',
    // Die Oberflaeche bietet nur "Permanent" und "Temporary" an, schreibt bei
    // "Permanent" aber channel_flag_semi_permanent=1. TeamSpeak loescht solche
    // Channels, sobald der letzte Client geht - Auto-Channels werden dadurch
    // endlos neu erzeugt. Ohne jedes Flag ist der Channel ebenfalls temporaer,
    // entgegen dem Kommentar im Original.
    fertig: /params\.channel_flag_permanent\s*=\s*'1'/,
    suchen: /\}\s*else if \(semi === '1'\) \{\s*params\.channel_flag_semi_permanent = '1';\s*\}/,
    ersetzen: `} else {
        // Korrektur: ohne dieses Flag legt TeamSpeak einen temporaeren Channel an.
        params.channel_flag_permanent = '1';
      }`,
  },
  {
    datei: ENGINE,
    titel: 'Set Variable liest den Namen, den die Oberflaeche schreibt',
    // Oberflaeche speichert varName/varValue, das Backend las name/value.
    // Der Knoten legte dadurch immer eine leere Variable an.
    fertig: /variableName:\s*config\.varName/,
    suchen: /variableName:\s*config\.name\s*\|\|\s*''/,
    ersetzen: "variableName: config.varName || config.name || ''",
  },
  {
    datei: ENGINE,
    titel: 'Set Variable liest den Wert, den die Oberflaeche schreibt',
    fertig: /value:\s*config\.varValue/,
    suchen: /value:\s*config\.value\s*\|\|\s*''/,
    ersetzen: "value: config.varValue || config.value || ''",
  },
  {
    datei: RUNNER,
    titel: 'Store As behaelt auch Einzelergebnisse (z. B. whoami)',
    // Bisher wurde immer result[0] abgelegt. Befehle, die ein einzelnes Objekt
    // liefern, kamen dadurch nie an. Arrays verhalten sich unveraendert.
    fertig: /Array\.isArray\(result\)\s*\?\s*result\[0\]\s*:\s*result/,
    suchen: /if \(data\.storeAs && result\?\.\[0\]\) \{\s*ctx\.setTemp\(data\.storeAs, result\[0\]\);/,
    ersetzen: `if (data.storeAs && result) {
      ctx.setTemp(data.storeAs, Array.isArray(result) ? result[0] : result);`,
  },
  {
    datei: CLIENT,
    titel: 'WebQuery wiederholt einmal bei abgerissener Verbindung',
    // Der Client haelt eine einzige dauerhafte Verbindung (keepAlive, maxSockets 1).
    // Wirft der TeamSpeak-Server sie weg - etwa nach einer Flood-Sperre oder weil
    // der Socket veraltet ist - schlaegt jede Abfrage mit "socket hang up" fehl,
    // ohne dass es je einen zweiten Versuch gaebe.
    fertig: /__ts6Wiederholung/,
    suchen: /(httpsAgent: useHttps \? this\.agent : undefined,\s*\}\);)/,
    ersetzen: `$1

    // Korrektur: eine abgerissene Verbindung einmal neu aufbauen, statt die
    // Abfrage sofort als "socket hang up" durchfallen zu lassen.
    this.http.interceptors.response.use(undefined, async (error: any) => {
      const anfrage = error?.config;
      const abgerissen = !error?.response && (
        ['ECONNRESET', 'EPIPE', 'ECONNABORTED', 'ETIMEDOUT'].includes(error?.code) ||
        /socket hang up/i.test(String(error?.message || ''))
      );
      if (anfrage && abgerissen && !anfrage.__ts6Wiederholung) {
        anfrage.__ts6Wiederholung = true;
        await new Promise((fertig) => setTimeout(fertig, 400));
        return this.http.request(anfrage);
      }
      return Promise.reject(error);
    });`,
  },
  {
    datei: EDITOR,
    titel: 'Set Variable bekommt einen Waehler fuer die Rechenart',
    // Das Backend kennt set, increment und append. Die Oberflaeche bot nichts davon
    // an, also stand operation immer auf 'set' - ein Zaehler war damit unerreichbar.
    fertig: /SelectItem value="increment"/,
    suchen: /(\{selectedNodeData\.type === 'variable' && \(\s*<div className="space-y-2">\s*)(<div>\s*<Label className="text-\[10px\] text-muted-foreground">Variable Name<\/Label>)/,
    ersetzen: `$1<div>
                        <Label className="text-[10px] text-muted-foreground">Operation</Label>
                        <Select value={selectedNodeData.config.operation || 'set'} onValueChange={(v) => setNodes((prev) => prev.map((n) => n.id === selectedNode ? { ...n, config: { ...n.config, operation: v } } : n))}>
                          <SelectTrigger className="h-7 text-xs mt-1"><SelectValue /></SelectTrigger>
                          <SelectContent>
                            <SelectItem value="set">Set</SelectItem>
                            <SelectItem value="increment">Increment</SelectItem>
                            <SelectItem value="append">Append</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      $2`,
  },
  {
    datei: EDITOR,
    titel: 'Auswahl "Permanent" schreibt kein semi-permanent mehr',
    fertig: /cfg\.channel_flag_temporary = '0'; delete cfg\.channel_flag_semi_permanent;/,
    suchen: /cfg\.channel_flag_temporary = '0'; cfg\.channel_flag_semi_permanent = '1';/,
    ersetzen: "cfg.channel_flag_temporary = '0'; delete cfg.channel_flag_semi_permanent;",
  },
  {
    datei: RUNNER,
    titel: 'AFK Mover achtet auf Stummschaltung und holt zurueck',
    // Ein ServerQuery bekommt kein Mute-Ereignis - SinusBot kann das nur, weil es sich
    // als echter Sprachclient anmeldet. Der Zustand steht aber in der Clientliste
    // (-voice). Wer laenger als die Schwelle stumm ist, wird in den AFK-Channel
    // geschoben; wer sich wieder entmutet, kommt in seinen alten Channel zurueck.
    fertig: /afkReturnedCount/,
    suchen: /const clients = await client\.executePost\(ctx\.sid, 'clientlist', \{ '-times': '', '-groups': '' \}\);[\s\S]*?ctx\.setTemp\('afkMovedCount', movedCount\);/,
    ersetzen: `const clients = await client.executePost(ctx.sid, 'clientlist', { '-times': '', '-groups': '', '-voice': '' });
    if (!Array.isArray(clients)) return;

    const jetzt = Date.now();
    let movedCount = 0;
    let zurueckCount = 0;

    for (const cl of clients) {
      if (String(cl.client_type) === '1') continue;

      const wer = String(cl.client_database_id || cl.clid);
      const stummSeit = \`mutedsince_\${wer}\`;
      const herkunft = \`afkfrom_\${wer}\`;

      const istStumm = String(cl.client_input_muted) === '1' || String(cl.client_output_muted) === '1';
      const leerlaufSek = (parseInt(cl.client_idle_time) || 0) / 1000;

      // Ausgenommene Gruppen bleiben unangetastet
      if (exemptIds.length > 0 && cl.client_servergroups) {
        const gruppen = String(cl.client_servergroups).split(',');
        if (exemptIds.some(g => gruppen.includes(g))) continue;
      }

      // Sitzt bereits im AFK-Channel: zurueckholen, sobald wieder aktiv
      if (String(cl.cid) === String(afkCid)) {
        if (!istStumm && leerlaufSek < thresholdSec) {
          const zurueck = await ctx.getVariable(herkunft);
          if (zurueck) {
            try {
              await client.executePost(ctx.sid, 'clientmove', { clid: cl.clid, cid: zurueck });
              zurueckCount++;
            } catch { /* Channel gibt es nicht mehr */ }
          }
          await ctx.setVariable(herkunft, '');
          await ctx.setVariable(stummSeit, '');
        }
        continue;
      }

      // Ausserhalb des AFK-Channels
      if (!istStumm) {
        if (await ctx.getVariable(stummSeit)) await ctx.setVariable(stummSeit, '');
        if (leerlaufSek < thresholdSec) continue;
      } else {
        const seit = parseFloat(await ctx.getVariable(stummSeit)) || 0;
        if (!seit) { await ctx.setVariable(stummSeit, String(jetzt)); continue; }
        if ((jetzt - seit) / 1000 < thresholdSec) continue;
      }

      try {
        await ctx.setVariable(herkunft, String(cl.cid));
        await client.executePost(ctx.sid, 'clientmove', { clid: cl.clid, cid: afkCid });
        movedCount++;
      } catch { /* skip clients that can't be moved */ }
    }

    ctx.setTemp('afkMovedCount', movedCount);
    ctx.setTemp('afkReturnedCount', zurueckCount);`,
  },
  {
    datei: RUNNER,
    titel: 'Rank Check sammelt die Online-Zeit wirklich',
    // Das Original liest onlinetime_<cldbid>, schreibt es aber nirgends. Der Kommentar
    // "accumulate via cron" beschreibt einen Sammler, den es nicht gibt - gezaehlt wurde
    // deshalb immer nur die laufende Sitzung. Jetzt wird der Zuwachs seit dem letzten
    // Lauf dazuaddiert und gespeichert; nach einem Neuverbinden faengt die
    // Verbindungszeit wieder bei 0 an, das faengt der Vergleich ab.
    fertig: /lastseen_\$\{cldbid\}/,
    suchen: /\/\/ Get total online time from BotVariable[\s\S]*?const totalHours = totalSeconds \/ 3600;/,
    ersetzen: `const varName = \`onlinetime_\${cldbid}\`;
      const merkName = \`lastseen_\${cldbid}\`;
      const gespeichert = parseFloat(await ctx.getVariable(varName)) || 0;
      const jetztVerbunden = (parseInt(cl.connection_connected_time) || 0) / 1000;
      const zuletzt = parseFloat(await ctx.getVariable(merkName)) || 0;
      const zuwachs = jetztVerbunden >= zuletzt ? jetztVerbunden - zuletzt : jetztVerbunden;
      const totalSeconds = gespeichert + zuwachs;
      await ctx.setVariable(varName, String(totalSeconds));
      await ctx.setVariable(merkName, String(jetztVerbunden));
      const totalHours = totalSeconds / 3600;`,
  },
];

let fehlgeschlagen = 0;
let angewendet = 0;
let uebersprungen = 0;

for (const k of KORREKTUREN) {
  const pfad = join(wurzel, k.datei);

  if (!existsSync(pfad)) {
    console.log(`FEHLT     ${k.datei}`);
    fehlgeschlagen++;
    continue;
  }

  const vorher = readFileSync(pfad, 'utf8');

  if (k.fertig.test(vorher)) {
    console.log(`schon ok  ${k.titel}`);
    uebersprungen++;
    continue;
  }

  const nachher = vorher.replace(k.suchen, k.ersetzen);

  if (nachher === vorher) {
    console.log(`FEHLER    ${k.titel}`);
    console.log(`          Fundstelle in ${k.datei} nicht erkannt.`);
    console.log('          Das Original hat sich vermutlich geaendert - Muster pruefen.');
    fehlgeschlagen++;
    continue;
  }

  writeFileSync(pfad, nachher);
  console.log(`gesetzt   ${k.titel}`);
  angewendet++;
}

console.log('');
console.log(`${angewendet} angewendet, ${uebersprungen} bereits vorhanden, ${fehlgeschlagen} fehlgeschlagen.`);

if (fehlgeschlagen > 0) {
  console.error('');
  console.error('Bau abgebrochen: mindestens eine Korrektur konnte nicht gesetzt werden.');
  process.exit(1);
}
