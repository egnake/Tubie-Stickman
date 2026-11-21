import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {FFMPEG, run, save, loadProject, probe} from '../scripts/autopilot-core.mjs';
import {render, painter} from '../scripts/autopilot-render.mjs';
import {
  CYBER_ACTOR_ROLES,
  CYBER_ACTOR_POSES,
  CYBER_WIDGET_KINDS,
  CYBER_MOODS,
  CYBER_CAMERAS,
  CYBER_EXTRA_SFX,
  validateCyberShot,
  validateCyberLayer,
} from '../scripts/cyber-visuals.mjs';

function makeBaseDir(t, {voiceDuration = 2.0, withMusic = false} = {}) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'tubie-cyber-test-'));
  t.after(() => fs.rmSync(dir, {recursive: true, force: true}));
  run(FFMPEG, ['-v', 'error', '-f', 'lavfi', '-i', `sine=frequency=340:duration=${voiceDuration}`, '-ar', '48000', '-ac', '1', '-y', path.join(dir, 'voice.wav')]);
  if (withMusic) {
    run(FFMPEG, ['-v', 'error', '-f', 'lavfi', '-i', `sine=frequency=110:duration=${voiceDuration}`, '-ar', '48000', '-ac', '2', '-y', path.join(dir, 'music.wav')]);
  }
  return dir;
}

function nonBackgroundFraction(rgba, w, h, bg = [10, 14, 23]) {
  let diff = 0;
  const total = w * h;
  for (let i = 0; i < total; i++) {
    const o = i * 4;
    if (Math.abs(rgba[o] - bg[0]) > 14 || Math.abs(rgba[o + 1] - bg[1]) > 14 || Math.abs(rgba[o + 2] - bg[2]) > 14) {
      diff++;
    }
  }
  return diff / total;
}

test('cyber_actor renders all roles and all 10 poses cleanly onto canvas', async t => {
  const dir = makeBaseDir(t, {voiceDuration: 2.0});
  const layers = [];
  for (const [idx, pose] of CYBER_ACTOR_POSES.entries()) {
    const role = CYBER_ACTOR_ROLES[idx % CYBER_ACTOR_ROLES.length];
    layers.push({
      type: 'cyber_actor',
      role,
      pose,
      x: 120 + (idx % 3) * 320,
      y: 260 + Math.floor(idx / 3) * 350,
      w: 220,
      h: 320,
      flip: idx % 2 === 1,
    });
  }
  const file = path.join(dir, 'project.json');
  save(file, {
    version: 1,
    style: 'B1',
    theme: 'cyber',
    backgroundColor: '#0A0E17',
    width: 360,
    height: 640,
    fps: 12,
    captionStyle: 'viral',
    captionY: 1780,
    voice: {path: 'voice.wav'},
    captions: [{start: 0, end: 2.0, text: 'SİBER KALKAN'}],
    shots: [{start: 0, end: 2.0, mood: 'danger', camera: 'punch_in', layers}],
    sfx: [],
  });
  const draw = await painter(loadProject(file));
  const canvas = draw(0.45);
  const rgba = canvas.getContext('2d').getImageData(0, 0, canvas.width, canvas.height).data;
  assert.ok(nonBackgroundFraction(rgba, canvas.width, canvas.height) > 0.08, 'Expected visible vector actors on canvas');
});

test('all 8 cyber_widget kinds, cyber_callout, and moods paint cleanly in portrait and landscape', async t => {
  const widgetData = {
    notification: {sender: 'Güvenlik', body: 'Hesabın riske girdi!', badge: 'ACİL'},
    terminal: {header: 'kali@soc', lines: ['$ scan --port 443', '[!] LEAK DETECTED']},
    url_inspect: {fakeUrl: 'https://banka-onay1a.com', realUrl: 'banka.com.tr', stamp: 'SAHTE!'},
    network_mitm: {left: 'TELEFON', middle: 'SAHTE AP', right: 'MODEM', packets: ['DEAUTH', 'HASH']},
    password_crack: {password: 'ahmet1998', crackTime: '0.4 SN'},
    threat_meter: {level: 95, label: 'YÜKSEK RİSK', safe: false},
    checklist: {items: ['Kural 1', 'Kural 2', 'Kural 3']},
    versus: {leftTitle: 'TEHLİKE', leftText: 'Ortak ağ', rightTitle: 'GÜVENLİ', rightText: 'VPN + 2FA'},
  };

  for (const layout of [{w: 360, h: 640, cy: 1780}, {w: 640, h: 360, cy: 630}]) {
    const dir = makeBaseDir(t, {voiceDuration: 2.0});
    const shots = CYBER_WIDGET_KINDS.map((kind, i) => ({
      start: i * 0.25,
      end: (i + 1) * 0.25,
      mood: CYBER_MOODS[i % CYBER_MOODS.length],
      camera: CYBER_CAMERAS[i % CYBER_CAMERAS.length],
      layers: [
        {type: 'cyber_widget', kind, x: layout.w > layout.h ? 560 : 80, y: 120, w: 620, h: 360, data: widgetData[kind]},
        {type: 'cyber_callout', text: 'DİKKAT!', x: 100, y: 60, w: 300, color: '#FF3B4E'},
      ],
    }));
    const file = path.join(dir, 'project.json');
    save(file, {
      version: 1,
      style: 'B1',
      theme: 'cyber',
      backgroundColor: '#0A0E17',
      width: layout.w,
      height: layout.h,
      fps: 12,
      captionStyle: 'viral',
      captionSize: layout.w > layout.h ? 44 : 56,
      captionY: layout.cy,
      voice: {path: 'voice.wav'},
      captions: [{start: 0, end: 2.0, text: 'İSTANBUL SİBER KORUMA', words: [{word: 'İSTANBUL', start: 0, end: 0.7}, {word: 'SİBER', start: 0.7, end: 1.3}, {word: 'KORUMA', start: 1.3, end: 2.0}]}],
      shots,
      sfx: CYBER_EXTRA_SFX.map((kind, idx) => ({time: 0.1 + idx * 0.2, kind, gain: 0.12})),
    });
    const proj = loadProject(file);
    const draw = await painter(proj);
    for (let i = 0; i < shots.length; i++) {
      const canvas = draw(i * 0.25 + 0.1);
      const rgba = canvas.getContext('2d').getImageData(0, 0, canvas.width, canvas.height).data;
      assert.ok(nonBackgroundFraction(rgba, canvas.width, canvas.height) > 0.05, `Widget ${CYBER_WIDGET_KINDS[i]} should render pixels`);
    }
  }
});

test('validation rejects invalid cyber specs and prevents cyber layer leaks into classic B1', t => {
  assert.throws(() => validateCyberLayer({type: 'cyber_actor', role: 'alien'}), /role/);
  assert.throws(() => validateCyberLayer({type: 'cyber_actor', pose: 'backflip'}), /pose/);
  assert.throws(() => validateCyberLayer({type: 'cyber_widget', kind: 'hologram'}), /kind/);
  assert.throws(() => validateCyberShot({mood: 'chill'}), /mood/);
  assert.throws(() => validateCyberShot({camera: 'dutch_roll'}), /camera/);

  const dir = makeBaseDir(t, {voiceDuration: 1.0});
  const file = path.join(dir, 'project.json');
  save(file, {
    version: 1,
    style: 'B1',
    width: 640,
    height: 360,
    fps: 12,
    voice: {path: 'voice.wav'},
    captions: [{start: 0, end: 1.0, text: 'Classic story.'}],
    shots: [{start: 0, end: 1.0, layers: [{type: 'cyber_actor', role: 'victim', pose: 'shocked', x: 100, y: 100, w: 200}]}],
    sfx: [],
  });
  assert.throws(() => loadProject(file), /Unknown layer cyber_actor/);

  save(file, {
    version: 1,
    style: 'B1',
    width: 640,
    height: 360,
    fps: 12,
    voice: {path: 'voice.wav'},
    captions: [{start: 0, end: 1.0, text: 'Classic story.'}],
    shots: [{start: 0, end: 1.0, layers: [{type: 'text', text: 'SAFE', x: 80, y: 80}]}],
    sfx: [{time: 0.2, kind: 'whoosh', gain: 0.1}],
  });
  assert.throws(() => loadProject(file), /Unknown SFX/);
});

test('offline E2E render encodes portrait 9:16 and landscape 16:9 cyber MP4s with sidechain BGM ducking', async t => {
  const cases = [
    {name: 'portrait', width: 360, height: 640, captionY: 1780, captionSize: 56, widgetPos: {x: 80, y: 980, w: 1120, h: 640}, actorPos: {x: 160, y: 420, w: 360, h: 480}},
    {name: 'landscape', width: 640, height: 360, captionY: 630, captionSize: 44, widgetPos: {x: 560, y: 135, w: 660, h: 420}, actorPos: {x: 80, y: 190, w: 260, h: 350}},
  ];

  for (const tc of cases) {
    const dir = makeBaseDir(t, {voiceDuration: 2.0, withMusic: true});
    const file = path.join(dir, 'project.json');
    save(file, {
      version: 1,
      style: 'B1',
      theme: 'cyber',
      backgroundColor: '#0A0E17',
      width: tc.width,
      height: tc.height,
      fps: 12,
      captionStyle: 'viral',
      captionSize: tc.captionSize,
      captionY: tc.captionY,
      voice: {path: 'voice.wav'},
      music: {path: 'music.wav', gain: 0.15},
      captions: [
        {start: 0, end: 1.0, text: 'SAHTE LİNKE TIKLAMA', words: [{word: 'SAHTE', start: 0, end: 0.35}, {word: 'LİNKE', start: 0.35, end: 0.7}, {word: 'TIKLAMA', start: 0.7, end: 1.0}]},
        {start: 1.0, end: 2.0, text: 'HEMEN 2FA AÇ', words: [{word: 'HEMEN', start: 1.0, end: 1.35}, {word: '2FA', start: 1.35, end: 1.7}, {word: 'AÇ', start: 1.7, end: 2.0}]},
      ],
      shots: [
        {
          start: 0,
          end: 1.0,
          mood: 'danger',
          camera: 'punch_in',
          layers: [
            {type: 'cyber_actor', role: 'victim', pose: 'shocked', ...tc.actorPos},
            {type: 'cyber_widget', kind: 'url_inspect', ...tc.widgetPos, data: {fakeUrl: 'https://giris-onay1.com', realUrl: 'resmi.com', stamp: 'TUZAK'}},
            {type: 'cyber_callout', text: 'DİKKAT!', x: 90, y: 70, w: 280, color: '#FF3B4E'},
          ],
        },
        {
          start: 1.0,
          end: 2.0,
          mood: 'defense',
          camera: 'zoom_slow',
          layers: [
            {type: 'cyber_actor', role: 'hero', pose: 'shield_hero', ...tc.actorPos},
            {type: 'cyber_widget', kind: 'checklist', ...tc.widgetPos, data: {items: ['Linki kontrol et', '2FA aç']}},
          ],
        },
      ],
      sfx: [
        {time: 0.05, kind: 'whoosh', gain: 0.16},
        {time: 0.4, kind: 'radar', gain: 0.16},
        {time: 1.05, kind: 'shield', gain: 0.18},
      ],
    });

    const rep = await render(loadProject(file), {full: true, replace: true});
    assert.equal(rep.full_decode, 'pass', `${tc.name} E2E decode failed`);

    const film = path.join(dir, 'output', 'film.mp4');
    assert.ok(fs.existsSync(film), `${tc.name} film.mp4 missing`);
    const info = probe(film);
    const vStream = info.streams.find(s => s.codec_type === 'video');
    const aStream = info.streams.find(s => s.codec_type === 'audio');
    assert.ok(vStream && aStream, `${tc.name} missing A/V stream`);
    assert.equal(Number(vStream.width), tc.width);
    assert.equal(Number(vStream.height), tc.height);
  }
});

test('offline cyber render supports audio-reactive speechLevels, segmented scenes, loopEnding, and proportional portrait contact sheets', async t => {
  const speechLevels = Array.from({length: 24}, (_, i) => Number((0.05 + (i / 23) * 0.90).toFixed(3)));
  const cases = [
    {name: 'portrait', width: 360, height: 640, captionY: 1780, captionSize: 56, widgetPos: {x: 80, y: 955, w: 1120, h: 575}, actorPos: {x: 160, y: 395, w: 360, h: 500}, expectPortraitSheet: true},
    {name: 'landscape', width: 640, height: 360, captionY: 648, captionSize: 44, widgetPos: {x: 562, y: 98, w: 674, h: 448}, actorPos: {x: 52, y: 242, w: 235, h: 305}, expectPortraitSheet: false},
  ];

  for (const tc of cases) {
    const dir = makeBaseDir(t, {voiceDuration: 2.0, withMusic: true});
    const file = path.join(dir, 'project.json');
    save(file, {
      version: 1,
      style: 'B1',
      theme: 'cyber',
      backgroundColor: '#0A0E17',
      width: tc.width,
      height: tc.height,
      fps: 12,
      loopEnding: true,
      speechLevels,
      captionStyle: 'viral',
      captionSize: tc.captionSize,
      captionY: tc.captionY,
      voice: {path: 'voice.wav'},
      music: {path: 'music.wav', gain: 0.14},
      captions: [
        {start: 0, end: 1.0, text: 'RİSK SEVİYESİ %94', words: [{word: 'RİSK', start: 0, end: 0.35}, {word: 'SEVİYESİ', start: 0.35, end: 0.7}, {word: '%94', start: 0.7, end: 1.0}]},
        {start: 1.0, end: 2.0, text: 'HEMEN VPN AÇ', words: [{word: 'HEMEN', start: 1.0, end: 1.35}, {word: 'VPN', start: 1.35, end: 1.7}, {word: 'AÇ', start: 1.7, end: 2.0}]},
      ],
      shots: [
        {
          start: 0,
          end: 1.0,
          mood: 'hook',
          camera: 'punch_in',
          sceneIndex: 0,
          sceneCount: 2,
          beat: 0.44,
          layers: [
            {type: 'cyber_actor', role: 'victim', pose: 'shocked', ...tc.actorPos},
            {type: 'cyber_widget', kind: 'threat_meter', ...tc.widgetPos, data: {level: 94, label: 'RİSK', safe: false}},
            {type: 'cyber_callout', text: 'KRİTİK RİSK', x: tc.width < tc.height ? 440 : 48, y: tc.width < tc.height ? 326 : 204, w: tc.width < tc.height ? 400 : 320, color: '#FF3B30'},
          ],
        },
        {
          start: 1.0,
          end: 2.0,
          mood: 'defense',
          camera: 'zoom_slow',
          sceneIndex: 1,
          sceneCount: 2,
          beat: 0.44,
          layers: [
            {type: 'cyber_actor', role: 'hero', pose: 'shield_hero', ...tc.actorPos},
            {type: 'cyber_widget', kind: 'url_inspect', ...tc.widgetPos, data: {fakeUrl: 'https://banka-onay1.com', realUrl: 'banka.com.tr', stamp: 'SAHTE'}},
          ],
        },
      ],
      sfx: [
        {time: 0.05, kind: 'whoosh', gain: 0.16},
        {time: 1.05, kind: 'shield', gain: 0.18},
      ],
    });

    const rep = await render(loadProject(file), {full: true, replace: true});
    assert.equal(rep.full_decode, 'pass', `${tc.name} E2E decode failed`);

    const sheetPath = path.join(dir, 'output', 'film-contact-sheet.png');
    assert.ok(fs.existsSync(sheetPath), `${tc.name} contact sheet missing`);
    const ihdr = fs.readFileSync(sheetPath);
    const sheetW = ihdr.readUInt32BE(16);
    const sheetH = ihdr.readUInt32BE(20);
    if (tc.expectPortraitSheet) {
      assert.ok(sheetH > sheetW, `expected portrait contact sheet height > width, got ${sheetW}x${sheetH}`);
    } else {
      assert.equal(sheetW, 960);
      assert.equal(sheetH, 540);
    }
  }
});
