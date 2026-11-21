// Procedural vector cybersecurity stickmen, hooded hacker, 8 live animated widgets,
// audio-reactive HUD, camera rig, and atmospheric transition engine for Tubie Studio (theme === 'cyber').

const clamp = x => Math.max(0, Math.min(1, x));
const easeOutCubic = x => 1 - Math.pow(1 - clamp(x), 3);
const easeOutBack = x => {
  const t = clamp(x);
  const c1 = 1.70158;
  const c3 = c1 + 1;
  return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
};

export const CYBER_LAYER_TYPES = ['cyber_actor', 'cyber_widget', 'cyber_callout'];
export const CYBER_ACTOR_ROLES = ['victim', 'hero', 'hacker'];
export const CYBER_ACTOR_POSES = [
  'idle_browse',
  'typing',
  'shocked',
  'panic_run',
  'thinking',
  'pointing',
  'shield_hero',
  'hacker_lurk',
  'hacker_strike',
  'celebrate',
];
export const CYBER_WIDGET_KINDS = [
  'notification',
  'terminal',
  'url_inspect',
  'network_mitm',
  'password_crack',
  'threat_meter',
  'checklist',
  'versus',
];
export const CYBER_MOODS = ['hook', 'danger', 'breach', 'twist', 'defense', 'cta'];
export const CYBER_CAMERAS = ['wide', 'punch_in', 'zoom_slow', 'shake', 'glitch_cut'];
export const CYBER_EXTRA_SFX = [
  'whoosh',
  'riser',
  'boom',
  'notification',
  'heartbeat',
  'radar',
  'shield',
  'buzz',
];

const MOOD_PALETTES = {
  hook: {primary: '#38BDF8', secondary: '#94A3B8', glow: 'rgba(56, 189, 248, 0.09)', border: 'rgba(56, 189, 248, 0.42)'},
  danger: {primary: '#F43F5E', secondary: '#F59E0B', glow: 'rgba(244, 63, 94, 0.11)', border: 'rgba(244, 63, 94, 0.48)'},
  breach: {primary: '#EF4444', secondary: '#FBBF24', glow: 'rgba(239, 68, 68, 0.13)', border: 'rgba(239, 68, 68, 0.52)'},
  twist: {primary: '#A78BFA', secondary: '#FBBF24', glow: 'rgba(167, 139, 250, 0.10)', border: 'rgba(167, 139, 250, 0.44)'},
  defense: {primary: '#10B981', secondary: '#34D399', glow: 'rgba(16, 185, 129, 0.10)', border: 'rgba(16, 185, 129, 0.48)'},
  cta: {primary: '#38BDF8', secondary: '#34D399', glow: 'rgba(56, 189, 248, 0.09)', border: 'rgba(56, 189, 248, 0.45)'},
};

export function validateCyberShot(shot) {
  if (shot.mood !== undefined && !CYBER_MOODS.includes(shot.mood)) {
    throw new Error(`Invalid cyber shot mood: ${shot.mood}`);
  }
  if (shot.camera !== undefined && !CYBER_CAMERAS.includes(shot.camera)) {
    throw new Error(`Invalid cyber shot camera: ${shot.camera}`);
  }
  for (const k of ['sceneIndex', 'sceneCount', 'beat']) {
    if (shot[k] !== undefined && !Number.isFinite(shot[k])) {
      throw new Error(`Invalid cyber shot ${k}`);
    }
  }
}

export function validateCyberLayer(l) {
  if (l.type === 'cyber_actor') {
    if (l.role !== undefined && !CYBER_ACTOR_ROLES.includes(l.role)) {
      throw new Error(`Invalid cyber_actor role: ${l.role}`);
    }
    if (l.pose !== undefined && !CYBER_ACTOR_POSES.includes(l.pose)) {
      throw new Error(`Invalid cyber_actor pose: ${l.pose}`);
    }
    if (l.flip !== undefined && typeof l.flip !== 'boolean') {
      throw new Error('cyber_actor flip must be boolean');
    }
    return;
  }
  if (l.type === 'cyber_widget') {
    if (!CYBER_WIDGET_KINDS.includes(l.kind)) {
      throw new Error(`Invalid cyber_widget kind: ${l.kind}`);
    }
    if (l.h !== undefined && (!Number.isFinite(l.h) || l.h <= 0 || l.h > 10000)) {
      throw new Error('cyber_widget height must be positive');
    }
    if (l.data !== undefined && (!l.data || typeof l.data !== 'object' || Array.isArray(l.data))) {
      throw new Error('cyber_widget data must be an object');
    }
    return;
  }
  if (l.type === 'cyber_callout') {
    if (typeof l.text !== 'string' || !l.text.trim() || l.text.length > 100) {
      throw new Error('cyber_callout text must be 1-100 characters');
    }
  }
}

export function sampleSpeechLevel(project, t) {
  const arr = project?.speechLevels;
  if (!Array.isArray(arr) || !arr.length) return 0;
  const fps = project.fps || 30;
  const idx = Math.max(0, t * fps);
  const i0 = Math.min(arr.length - 1, Math.floor(idx));
  const i1 = Math.min(arr.length - 1, i0 + 1);
  const frac = idx - Math.floor(idx);
  const v = arr[i0] * (1 - frac) + arr[i1] * frac;
  return Number.isFinite(v) ? clamp(v) : 0;
}

function rrect(ctx, x, y, w, h, r) {
  const rad = Math.max(0, Math.min(r, w / 2, h / 2));
  ctx.beginPath();
  if (typeof ctx.roundRect === 'function') ctx.roundRect(x, y, w, h, rad);
  else ctx.rect(x, y, w, h);
}

function wrapTextSafe(ctx, text, maxWidth, maxLines = 4) {
  const words = String(text || '').trim().split(/\s+/).filter(Boolean);
  if (!words.length) return [];
  const lines = [];
  let line = '';
  for (const word of words) {
    const candidate = line ? `${line} ${word}` : word;
    if (ctx.measureText(candidate).width <= maxWidth) {
      line = candidate;
    } else if (line) {
      lines.push(line);
      line = word;
    } else {
      let cut = word;
      while (cut.length > 1 && ctx.measureText(`${cut}…`).width > maxWidth) {
        cut = cut.slice(0, -1);
      }
      lines.push(`${cut}…`);
      line = '';
    }
    if (lines.length >= maxLines) break;
  }
  if (line && lines.length < maxLines) lines.push(line);
  return lines.slice(0, maxLines);
}

function formatClock(sec) {
  const total = Math.max(0, Math.floor(sec || 0));
  const m = String(Math.floor(total / 60)).padStart(2, '0');
  const s = String(total % 60).padStart(2, '0');
  return `${m}:${s}`;
}

function drawWarningTriangle(ctx, cx, cy, r, color) {
  ctx.save();
  ctx.strokeStyle = color;
  ctx.fillStyle = color;
  ctx.lineWidth = Math.max(2, r * 0.18);
  ctx.lineJoin = 'round';
  ctx.beginPath();
  ctx.moveTo(cx, cy - r);
  ctx.lineTo(cx - r * 0.92, cy + r * 0.78);
  ctx.lineTo(cx + r * 0.92, cy + r * 0.78);
  ctx.closePath();
  ctx.stroke();
  ctx.fillRect(cx - r * 0.08, cy - r * 0.35, r * 0.16, r * 0.56);
  ctx.fillRect(cx - r * 0.08, cy + r * 0.34, r * 0.16, r * 0.18);
  ctx.restore();
}

function drawShieldBadge(ctx, cx, cy, r, strokeColor, fillColor, drawProg = 1) {
  ctx.save();
  ctx.beginPath();
  ctx.moveTo(cx, cy - r);
  ctx.lineTo(cx + r * 0.85, cy - r * 0.65);
  ctx.lineTo(cx + r * 0.75, cy + r * 0.25);
  ctx.quadraticCurveTo(cx, cy + r * 0.95, cx, cy + r * 1.05);
  ctx.quadraticCurveTo(cx, cy + r * 0.95, cx - r * 0.75, cy + r * 0.25);
  ctx.lineTo(cx - r * 0.85, cy - r * 0.65);
  ctx.closePath();
  ctx.fillStyle = fillColor || 'rgba(16, 185, 129, 0.14)';
  ctx.fill();
  ctx.strokeStyle = strokeColor || '#10B981';
  ctx.lineWidth = Math.max(2.5, r * 0.14);
  ctx.stroke();

  const p = clamp(drawProg);
  if (p > 0.05) {
    const p0 = [cx - r * 0.34, cy + r * 0.02];
    const p1 = [cx - r * 0.06, cy + r * 0.30];
    const p2 = [cx + r * 0.38, cy - r * 0.28];
    ctx.beginPath();
    ctx.moveTo(p0[0], p0[1]);
    if (p <= 0.45) {
      const f = p / 0.45;
      ctx.lineTo(p0[0] + (p1[0] - p0[0]) * f, p0[1] + (p1[1] - p0[1]) * f);
    } else {
      const f = (p - 0.45) / 0.55;
      ctx.lineTo(p1[0], p1[1]);
      ctx.lineTo(p1[0] + (p2[0] - p1[0]) * f, p1[1] + (p2[1] - p1[1]) * f);
    }
    ctx.strokeStyle = strokeColor || '#10B981';
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';
    ctx.stroke();
  }
  ctx.restore();
}

function drawPadlockVector(ctx, cx, cy, r, color, cracked = false) {
  ctx.save();
  ctx.strokeStyle = color;
  ctx.fillStyle = color;
  ctx.lineWidth = Math.max(2.6, r * 0.24);
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';

  // Shackle
  const openShift = cracked ? -r * 0.24 : 0;
  ctx.beginPath();
  ctx.arc(cx, cy - r * 0.35 + openShift, r * 0.52, Math.PI, 0, false);
  if (!cracked) {
    ctx.lineTo(cx + r * 0.52, cy);
  }
  ctx.moveTo(cx - r * 0.52, cy - r * 0.35 + openShift);
  ctx.lineTo(cx - r * 0.52, cy);
  ctx.stroke();

  // Body
  rrect(ctx, cx - r * 0.78, cy - r * 0.05, r * 1.56, r * 1.15, r * 0.25);
  ctx.stroke();

  if (cracked) {
    ctx.beginPath();
    ctx.moveTo(cx - r * 0.2, cy + r * 0.15);
    ctx.lineTo(cx + r * 0.1, cy + r * 0.52);
    ctx.lineTo(cx - r * 0.12, cy + r * 0.85);
    ctx.stroke();
  } else {
    ctx.beginPath();
    ctx.arc(cx, cy + r * 0.48, r * 0.18, 0, Math.PI * 2);
    ctx.fill();
  }
  ctx.restore();
}

function drawPointerCursor(ctx, x, y, sc, clicking = false) {
  ctx.save();
  ctx.translate(x, y);
  if (clicking) ctx.scale(0.88, 0.88);
  ctx.shadowColor = 'rgba(0, 0, 0, 0.65)';
  ctx.shadowBlur = 8;
  ctx.fillStyle = '#FFFFFF';
  ctx.strokeStyle = '#070B14';
  ctx.lineWidth = 2.5 * sc;
  ctx.beginPath();
  ctx.moveTo(0, 0);
  ctx.lineTo(0, 28 * sc);
  ctx.lineTo(8 * sc, 21 * sc);
  ctx.lineTo(14 * sc, 31 * sc);
  ctx.lineTo(19 * sc, 28 * sc);
  ctx.lineTo(13 * sc, 18 * sc);
  ctx.lineTo(22 * sc, 18 * sc);
  ctx.closePath();
  ctx.fill();
  ctx.stroke();
  ctx.restore();
}

export function drawCyberAtmosphere(ctx, plane, t, shot, dt, speechLevel = 0) {
  const mood = shot?.mood || 'hook';
  const pal = MOOD_PALETTES[mood] || MOOD_PALETTES.hook;
  const isPortrait = plane.height > 1500;
  const captionBand = isPortrait ? plane.height - 520 : plane.height - 110;
  const speechBoost = clamp(speechLevel);

  // 1. Refined dark-editorial studio dot-grid
  ctx.save();
  ctx.fillStyle = 'rgba(148, 163, 184, 0.08)';
  for (let x = 40; x < plane.width; x += 64) {
    for (let y = 40; y < plane.height; y += 64) {
      ctx.fillRect(x, y, 1.5, 1.5);
    }
  }
  ctx.restore();

  // 2. Whisper-soft editorial network mesh (depth without visual clutter)
  ctx.save();
  const nodes = [];
  const nodeCount = 10;
  for (let i = 0; i < nodeCount; i++) {
    const baseX = ((i * 257 + 110) % (plane.width - 140)) + 70;
    const baseY = ((i * 389 + 150) % Math.max(240, captionBand - 160)) + 90;
    const nx = baseX + Math.sin(t * (0.40 + (i % 3) * 0.14) + i * 1.3) * 30;
    const ny = baseY + Math.cos(t * (0.34 + (i % 4) * 0.12) + i * 0.9) * 24;
    nodes.push([nx, ny]);
  }
  const edgeAlpha = (0.025 + speechBoost * 0.030).toFixed(3);
  ctx.strokeStyle = `rgba(148, 163, 184, ${edgeAlpha})`;
  ctx.lineWidth = 1;
  ctx.beginPath();
  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const dist = Math.hypot(nodes[i][0] - nodes[j][0], nodes[i][1] - nodes[j][1]);
      if (dist < 320) {
        ctx.moveTo(nodes[i][0], nodes[i][1]);
        ctx.lineTo(nodes[j][0], nodes[j][1]);
      }
    }
  }
  ctx.stroke();
  ctx.fillStyle = 'rgba(148, 163, 184, 0.12)';
  for (const [nx, ny] of nodes) {
    ctx.beginPath();
    ctx.arc(nx, ny, 2.2, 0, Math.PI * 2);
    ctx.fill();
  }
  ctx.restore();

  // 3. Soft studio key-light spotlight (reacts subtly to speech energy)
  ctx.save();
  const auraRadius = plane.width * (0.64 + speechBoost * 0.10);
  const aura = ctx.createRadialGradient(
    plane.width * 0.5,
    plane.height * 0.42,
    plane.width * 0.06,
    plane.width * 0.5,
    plane.height * 0.42,
    auraRadius,
  );
  aura.addColorStop(0, pal.glow);
  aura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = aura;
  ctx.fillRect(0, 0, plane.width, plane.height);
  ctx.restore();

  // 4. Clean matte stage floor divider hairline (anchors widget theater & stage)
  ctx.save();
  const deckY = isPortrait ? 1515 : 550;
  const deckGlow = ctx.createLinearGradient(60, deckY, plane.width - 60, deckY);
  deckGlow.addColorStop(0, 'rgba(148, 163, 184, 0)');
  deckGlow.addColorStop(0.25, 'rgba(148, 163, 184, 0.14)');
  deckGlow.addColorStop(0.5, pal.border);
  deckGlow.addColorStop(0.75, 'rgba(148, 163, 184, 0.14)');
  deckGlow.addColorStop(1, 'rgba(148, 163, 184, 0)');
  ctx.strokeStyle = deckGlow;
  ctx.lineWidth = 1.5;
  ctx.beginPath();
  ctx.moveTo(60, deckY);
  ctx.lineTo(plane.width - 60, deckY);
  ctx.stroke();
  ctx.restore();

  // 5. Studio Vignette + Restrained Crimson Edge Glow on Danger/Breach
  ctx.save();
  const vig = ctx.createRadialGradient(
    plane.width / 2,
    plane.height * 0.45,
    plane.width * 0.28,
    plane.width / 2,
    plane.height * 0.45,
    plane.width * 0.90,
  );
  vig.addColorStop(0, 'rgba(0, 0, 0, 0)');
  vig.addColorStop(1, 'rgba(5, 8, 15, 0.65)');
  ctx.fillStyle = vig;
  ctx.fillRect(0, 0, plane.width, plane.height);

  if (mood === 'danger' || mood === 'breach') {
    const edgeAlpha2 = (0.08 + (0.04 + speechBoost * 0.05) * Math.abs(Math.sin(t * 5.0 + dt))).toFixed(3);
    const edgeGrad = ctx.createLinearGradient(0, 0, 0, 140);
    edgeGrad.addColorStop(0, `rgba(244, 63, 94, ${edgeAlpha2})`);
    edgeGrad.addColorStop(1, 'rgba(244, 63, 94, 0)');
    ctx.fillStyle = edgeGrad;
    ctx.fillRect(0, 0, plane.width, 140);
  }
  ctx.restore();
}

export function drawCyberTopHud(ctx, plane, t, duration, shot = null, speechLevel = 0) {
  const isPortrait = plane.height > 1500;
  const bPad = isPortrait ? 44 : 28;
  const hudFontSize = isPortrait ? 20 : 16;
  const speech = clamp(speechLevel);

  ctx.save();
  // Clean segmented Story Arc Progress Bar across the top
  const barX = bPad;
  const barY = isPortrait ? 36 : 24;
  const barW = plane.width - bPad * 2;
  const totalDur = Math.max(0.1, duration || 30);

  if (shot && Number.isFinite(shot.sceneCount) && shot.sceneCount > 1 && Number.isFinite(shot.sceneIndex)) {
    const count = Math.max(2, Math.floor(shot.sceneCount));
    const curIdx = Math.max(0, Math.min(count - 1, Math.floor(shot.sceneIndex)));
    const gap = 8;
    const segW = (barW - gap * (count - 1)) / count;
    const shotSpan = Math.max(0.05, (shot.end ?? totalDur) - (shot.start ?? 0));
    const shotProg = clamp((t - (shot.start ?? 0)) / shotSpan);

    for (let s = 0; s < count; s++) {
      const sx = barX + s * (segW + gap);
      ctx.fillStyle = 'rgba(148, 163, 184, 0.16)';
      rrect(ctx, sx, barY, segW, 4, 2);
      ctx.fill();

      const fillFrac = s < curIdx ? 1 : s === curIdx ? shotProg : 0;
      if (fillFrac > 0) {
        const segGrad = ctx.createLinearGradient(sx, 0, sx + segW, 0);
        segGrad.addColorStop(0, '#38BDF8');
        segGrad.addColorStop(1, '#34D399');
        ctx.fillStyle = segGrad;
        rrect(ctx, sx, barY, Math.max(4, segW * fillFrac), 4, 2);
        ctx.fill();
      }
    }
  } else {
    const pct = clamp(t / totalDur);
    ctx.fillStyle = 'rgba(148, 163, 184, 0.16)';
    rrect(ctx, barX, barY, barW, 4, 2);
    ctx.fill();
    const barGrad = ctx.createLinearGradient(barX, 0, barX + barW, 0);
    barGrad.addColorStop(0, '#38BDF8');
    barGrad.addColorStop(1, '#34D399');
    ctx.fillStyle = barGrad;
    rrect(ctx, barX, barY, Math.max(4, barW * pct), 4, 2);
    ctx.fill();
  }

  // Left: Minimal Chapter Label; Right: Timecode + Sleek 10-Bar Audio Spectrum
  const rowY = barY + 14;
  ctx.font = `700 ${hudFontSize}px B1Caption`;
  ctx.fillStyle = 'rgba(148, 163, 184, 0.88)';
  ctx.textBaseline = 'top';
  ctx.textAlign = 'left';
  if (shot && Number.isFinite(shot.sceneCount) && shot.sceneCount > 1 && Number.isFinite(shot.sceneIndex)) {
    const cur = String(Math.floor(shot.sceneIndex) + 1).padStart(2, '0');
    const tot = String(Math.floor(shot.sceneCount)).padStart(2, '0');
    ctx.fillText(`BÖLÜM ${cur} / ${tot}`, barX, rowY);
  } else {
    ctx.fillText('ANALİZ', barX, rowY);
  }

  const bars = 10;
  const barStep = isPortrait ? 6 : 5;
  const eqW = bars * barStep;
  const eqBaseX = plane.width - bPad - eqW;
  const eqBaseY = rowY + hudFontSize;
  const maxH = isPortrait ? 15 : 12;
  for (let eq = 0; eq < bars; eq++) {
    const harmonic = 0.50 + 0.50 * Math.abs(Math.sin(t * (9.5 + eq * 1.6) + eq * 0.75));
    const eqH = speech > 0.02
      ? 3 + speech * harmonic * maxH
      : 2.5 + 0.5 * Math.abs(Math.sin(t * 2.5 + eq));
    ctx.fillStyle = speech > 0.72 && eq % 3 === 0
      ? 'rgba(52, 211, 153, 0.88)'
      : speech > 0.02
        ? 'rgba(56, 189, 248, 0.82)'
        : 'rgba(148, 163, 184, 0.32)';
    ctx.fillRect(eqBaseX + eq * barStep, eqBaseY - eqH, isPortrait ? 3.5 : 3, eqH);
  }

  const clockStr = `${formatClock(t)} / ${formatClock(duration)}`;
  ctx.textAlign = 'right';
  ctx.fillStyle = 'rgba(203, 213, 225, 0.82)';
  ctx.fillText(clockStr, eqBaseX - 14, rowY);
  ctx.restore();
}

export function applyCyberCamera(ctx, plane, shot, dt, speechLevel = 0) {
  const cam = shot?.camera || 'wide';
  const span = Math.max(0.5, (shot?.end ?? 1) - (shot?.start ?? 0));
  const prog = clamp(dt / span);
  const vocalPulse = clamp(speechLevel) * 0.008;
  let scale = 1.0 + vocalPulse;
  let dx = 0;
  let dy = 0;
  let rot = 0;

  if (cam === 'punch_in') {
    const p = easeOutCubic(dt / 0.24);
    scale += 0.052 * p + 0.010 * Math.sin(dt * 1.8);
  } else if (cam === 'zoom_slow') {
    scale += 0.056 * prog;
  } else if (cam === 'shake') {
    const decay = Math.exp(-dt * 4.8);
    dx = Math.sin(dt * 48) * 10 * decay;
    dy = Math.cos(dt * 58) * 7 * decay;
    scale += 0.02 + 0.036 * decay;
    rot = Math.sin(dt * 36) * 0.012 * decay;
  } else if (cam === 'glitch_cut') {
    const settle = 1 - easeOutCubic(dt / 0.20);
    scale += 0.068 * settle;
    dx = settle > 0.05 ? Math.sin(dt * 85) * 8 * settle : 0;
  } else {
    scale += 0.012 * Math.sin(dt * 2.0);
  }

  const cx = plane.width / 2;
  const cy = plane.height * 0.46;
  ctx.translate(cx + dx, cy + dy);
  ctx.rotate(rot);
  ctx.scale(scale, scale);
  ctx.translate(-cx, -cy);
}

export function drawCyberTransitionOverlay(ctx, plane, shot, dt) {
  if (dt < 0 || dt >= 0.18) return;
  const frac = 1 - dt / 0.18;
  const cam = shot?.camera || 'wide';
  const mood = shot?.mood || 'hook';
  ctx.save();
  if (cam === 'glitch_cut' || mood === 'breach') {
    const bars = 5;
    for (let i = 0; i < bars; i++) {
      const y = ((i * 293 + Math.floor(dt * 160)) % Math.max(100, plane.height - 40));
      const h = 3 + ((i * 5) % 8);
      ctx.fillStyle = i % 2 === 0
        ? `rgba(56, 189, 248, ${(0.12 * frac).toFixed(3)})`
        : `rgba(244, 63, 94, ${(0.14 * frac).toFixed(3)})`;
      ctx.fillRect(0, y, plane.width, h);
    }
  } else {
    const scanY = (1 - frac) * plane.height * 0.85;
    const g = ctx.createLinearGradient(0, scanY - 36, 0, scanY + 36);
    const sweepRgb = mood === 'defense' ? '16, 185, 129' : mood === 'danger' ? '244, 63, 94' : '56, 189, 248';
    g.addColorStop(0, `rgba(${sweepRgb}, 0)`);
    g.addColorStop(0.5, `rgba(${sweepRgb}, ${(0.12 * frac).toFixed(3)})`);
    g.addColorStop(1, `rgba(${sweepRgb}, 0)`);
    ctx.fillStyle = g;
    ctx.fillRect(0, scanY - 36, plane.width, 72);
  }
  ctx.restore();
}

export function drawCyberStageFX(ctx, plane, shot, dt, t, speechLevel = 0) {
  const layers = Array.isArray(shot?.layers) ? shot.layers : [];
  const visibleActors = layers.filter(l =>
    l.type === 'cyber_actor' && dt >= (l.from || 0) && dt < (l.to ?? Infinity),
  );
  if (!visibleActors.length) return;

  const mood = shot?.mood || 'hook';
  const speech = clamp(speechLevel);
  ctx.save();

  if (visibleActors.length >= 2 && (mood === 'danger' || mood === 'breach')) {
    const a = visibleActors[0];
    const b = visibleActors[1];
    const ax = (a.x || 0) + (a.w || 360) * 0.5;
    const ay = (a.y || 0) + (a.h || 420) * 0.36;
    const bx = (b.x || 0) + (b.w || 360) * 0.5;
    const by = (b.y || 0) + (b.h || 420) * 0.36;
    const midX = (ax + bx) * 0.5;
    const midY = Math.min(ay, by) - 60;

    ctx.strokeStyle = `rgba(244, 63, 94, ${(0.28 + speech * 0.25).toFixed(3)})`;
    ctx.lineWidth = 2.5;
    ctx.setLineDash([8, 8]);
    ctx.lineDashOffset = -t * 85;
    ctx.beginPath();
    ctx.moveTo(bx, by);
    ctx.quadraticCurveTo(midX, midY, ax, ay);
    ctx.stroke();
    ctx.setLineDash([]);

    for (let s = 0; s < 3; s++) {
      const p = (t * 0.85 + s * 0.33) % 1;
      const inv = 1 - p;
      const sx = inv * inv * bx + 2 * inv * p * midX + p * p * ax;
      const sy = inv * inv * by + 2 * inv * p * midY + p * p * ay;
      ctx.fillStyle = s % 2 === 0 ? '#F43F5E' : '#FBBF24';
      ctx.shadowColor = ctx.fillStyle;
      ctx.shadowBlur = 4;
      ctx.beginPath();
      ctx.arc(sx, sy, 4.5, 0, Math.PI * 2);
      ctx.fill();
    }
  } else if (mood === 'hook' && visibleActors.length >= 2) {
    const target = visibleActors[0];
    const tx = (target.x || 0) + (target.w || 360) * 0.5;
    const ty = (target.y || 0) + (target.h || 420) * 0.24;
    const rad = (target.w || 360) * (0.28 + 0.025 * Math.sin(t * 6));
    ctx.strokeStyle = 'rgba(244, 63, 94, 0.35)';
    ctx.lineWidth = 2;
    ctx.setLineDash([8, 6]);
    ctx.beginPath();
    ctx.arc(tx, ty, rad, 0, Math.PI * 2);
    ctx.stroke();
    ctx.setLineDash([]);
  } else if (mood === 'defense' || mood === 'cta') {
    const stageTop = plane.height > 1500 ? 410 : 250;
    const stageH = plane.height > 1500 ? 480 : 280;
    for (let p = 0; p < 5; p++) {
      const px = plane.width * (0.20 + ((p * 0.14) % 0.60));
      const frac = (t * 0.42 + p * 0.20) % 1;
      const py = stageTop + stageH * (1 - frac);
      ctx.fillStyle = `rgba(52, 211, 153, ${((1 - frac) * (0.28 + speech * 0.20)).toFixed(3)})`;
      ctx.beginPath();
      ctx.arc(px, py, 3.2, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  ctx.restore();
}

export function drawCyberActor(ctx, l, dt, t, speechLevel = 0) {
  const w = l.w || 360;
  const h = l.h || w * 1.18;
  const role = l.role || 'victim';
  const pose = l.pose || (role === 'hacker' ? 'hacker_lurk' : 'idle_browse');
  const isHacker = role === 'hacker';
  const isHero = role === 'hero';
  const speech = clamp(speechLevel);
  const cx = w * 0.5;
  const breath = Math.sin(t * 4.0 + dt * 1.2) * (h * 0.011);
  const headNod = speech * Math.sin(t * 14) * (h * 0.008);
  const strokeW = Math.max(4, w * 0.026);
  const headR = w * 0.175;
  const headY = h * 0.22 + breath + headNod;
  const neckY = headY + headR;
  const shoulderY = neckY + h * 0.05;
  const hipY = shoulderY + h * 0.27;
  const footY = h * 0.92;

  ctx.save();
  if (l.flip) {
    ctx.translate(w, 0);
    ctx.scale(-1, 1);
  }

  // Elastic entrance squash & stretch anchored on ground pad
  if (dt >= 0 && dt < 0.22) {
    const ent = easeOutBack(dt / 0.22);
    const sy = 0.88 + 0.12 * ent;
    const sx = 2 - sy;
    ctx.translate(cx, footY);
    ctx.scale(sx, sy);
    ctx.translate(-cx, -footY);
  }

  // Grounded Matte Stage Contact Shadow
  const rimCol = isHacker ? 'rgba(244, 63, 94,' : isHero ? 'rgba(16, 185, 129,' : 'rgba(56, 189, 248,';
  ctx.fillStyle = 'rgba(0, 0, 0, 0.42)';
  ctx.beginPath();
  ctx.ellipse(cx, footY + h * 0.015, w * 0.28, h * 0.032, 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.strokeStyle = `${rimCol} 0.26)`;
  ctx.lineWidth = 1.6;
  ctx.stroke();

  // Subtle Impact Ripple on dramatic poses
  if ((pose === 'shocked' || pose === 'shield_hero' || pose === 'hacker_strike') && dt >= 0 && dt < 0.45) {
    const sw = easeOutCubic(dt / 0.45);
    ctx.strokeStyle = `${rimCol} ${(0.24 * (1 - sw)).toFixed(3)})`;
    ctx.lineWidth = 2.2;
    ctx.beginPath();
    ctx.ellipse(cx, footY + h * 0.015, w * (0.28 + sw * 0.22), h * (0.032 + sw * 0.025), 0, 0, Math.PI * 2);
    ctx.stroke();
  }

  // Soft studio key-light backdrop
  const auraColor = isHacker
    ? 'rgba(244, 63, 94, 0.12)'
    : isHero
      ? 'rgba(16, 185, 129, 0.11)'
      : 'rgba(56, 189, 248, 0.11)';
  const auraGrad = ctx.createRadialGradient(cx, h * 0.44, w * 0.06, cx, h * 0.44, w * 0.50);
  auraGrad.addColorStop(0, auraColor);
  auraGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = auraGrad;
  ctx.beginPath();
  ctx.arc(cx, h * 0.44, w * 0.50, 0, Math.PI * 2);
  ctx.fill();

  const bodyColor = isHacker ? '#F43F5E' : isHero ? '#ECFEFF' : '#F1F5F9';

  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.lineWidth = strokeW;

  // Pose dynamics
  let bodyLean = 0;
  let lHand = [cx - w * 0.26, shoulderY + h * 0.19];
  let rHand = [cx + w * 0.26, shoulderY + h * 0.19];
  let lElbow = [cx - w * 0.18, shoulderY + h * 0.10];
  let rElbow = [cx + w * 0.18, shoulderY + h * 0.10];
  let lFoot = [cx - w * 0.15, footY];
  let rFoot = [cx + w * 0.15, footY];
  let lKnee = [cx - w * 0.11, (hipY + footY) * 0.5];
  let rKnee = [cx + w * 0.11, (hipY + footY) * 0.5];

  if (pose === 'idle_browse' || pose === 'typing') {
    bodyLean = w * 0.025;
    const tap = pose === 'typing' ? Math.sin(t * 20) * (h * 0.020) : Math.sin(t * 5) * (h * 0.01);
    lElbow = [cx - w * 0.16, shoulderY + h * 0.12];
    rElbow = [cx + w * 0.19, shoulderY + h * 0.11];
    lHand = [cx + w * 0.04, shoulderY + h * 0.14 + tap];
    rHand = [cx + w * 0.16, shoulderY + h * 0.12 - tap];
  } else if (pose === 'shocked') {
    const tremor = Math.sin(t * 24) * (w * 0.014);
    bodyLean = tremor;
    lElbow = [cx - w * 0.28, headY + headR * 0.25];
    rElbow = [cx + w * 0.28, headY + headR * 0.25];
    lHand = [cx - headR * 0.96, headY - headR * 0.05];
    rHand = [cx + headR * 0.96, headY - headR * 0.05];
    lFoot = [cx - w * 0.19, footY];
    rFoot = [cx + w * 0.19, footY];
  } else if (pose === 'panic_run') {
    const stride = Math.sin(t * 14);
    bodyLean = w * 0.08;
    lElbow = [cx - w * 0.18, shoulderY + h * 0.06];
    rElbow = [cx + w * 0.20, shoulderY + h * 0.02];
    lHand = [cx - w * 0.28 * stride, shoulderY - h * 0.04];
    rHand = [cx + w * 0.28 * stride, shoulderY - h * 0.06];
    lFoot = [cx - w * 0.22 * stride, footY - Math.max(0, stride) * h * 0.07];
    rFoot = [cx + w * 0.22 * stride, footY - Math.max(0, -stride) * h * 0.07];
    lKnee = [cx - w * 0.10 * stride, (hipY + footY) * 0.5 - h * 0.04];
    rKnee = [cx + w * 0.10 * stride, (hipY + footY) * 0.5 - h * 0.04];
  } else if (pose === 'thinking') {
    lElbow = [cx - w * 0.17, shoulderY + h * 0.14];
    lHand = [cx - w * 0.05, shoulderY + h * 0.18];
    rElbow = [cx + w * 0.22, shoulderY + h * 0.11];
    rHand = [cx + w * 0.06, headY + headR * 0.85];
  } else if (pose === 'pointing') {
    const pulse = Math.sin(t * 8) * (w * 0.015);
    lElbow = [cx - w * 0.16, shoulderY + h * 0.14];
    lHand = [cx - w * 0.20, hipY];
    rElbow = [cx + w * 0.18, shoulderY + h * 0.01];
    rHand = [cx + w * 0.36 + pulse, shoulderY - h * 0.04];
  } else if (pose === 'shield_hero') {
    const lift = Math.sin(t * 5) * (h * 0.015);
    lElbow = [cx - w * 0.22, shoulderY + h * 0.04];
    lHand = [cx - w * 0.30, shoulderY - h * 0.06 + lift];
    rElbow = [cx + w * 0.20, shoulderY + h * 0.05];
    rHand = [cx + w * 0.24, shoulderY + h * 0.02 + lift];
  } else if (pose === 'hacker_lurk' || pose === 'hacker_strike') {
    bodyLean = -w * 0.04;
    const jab = pose === 'hacker_strike' ? Math.sin(t * 20) * (h * 0.028) : Math.sin(t * 5) * (h * 0.01);
    lElbow = [cx - w * 0.20, shoulderY + h * 0.10];
    rElbow = [cx + w * 0.20, shoulderY + h * 0.09];
    lHand = [cx - w * 0.26, shoulderY + h * 0.16 + jab];
    rHand = [cx + w * 0.08, shoulderY + h * 0.16 - jab];
  } else if (pose === 'celebrate') {
    const bounce = Math.abs(Math.sin(t * 7)) * (h * 0.045);
    lElbow = [cx - w * 0.22, shoulderY - h * 0.03 - bounce];
    rElbow = [cx + w * 0.22, shoulderY - h * 0.03 - bounce];
    lHand = [cx - w * 0.28, headY - headR * 0.5 - bounce];
    rHand = [cx + w * 0.28, headY - headR * 0.5 - bounce];
  }

  const headCx = cx + bodyLean;
  const shoulderCx = cx + bodyLean * 0.45;

  // Motion speed lines behind panic_run
  if (pose === 'panic_run') {
    ctx.save();
    ctx.strokeStyle = 'rgba(56, 189, 248, 0.36)';
    ctx.lineWidth = 2.2;
    for (let s = 0; s < 3; s++) {
      const sy = headY + s * (h * 0.20);
      const sxOff = ((t * 180 + s * 30) % 45);
      ctx.beginPath();
      ctx.moveTo(cx - w * 0.44 - sxOff * 0.3, sy);
      ctx.lineTo(cx - w * 0.22, sy);
      ctx.stroke();
    }
    ctx.restore();
  }

  // Legs + Torso + Arms (crisp dark-drop shadow, no radioactive neon halo)
  ctx.save();
  ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
  ctx.shadowBlur = 5;
  ctx.strokeStyle = bodyColor;

  if (isHacker) {
    // Hooded matte charcoal hacker cloak with clean coral rim
    ctx.save();
    ctx.fillStyle = '#111827';
    ctx.beginPath();
    ctx.moveTo(headCx, headY - headR * 1.48);
    ctx.quadraticCurveTo(headCx + headR * 1.44, headY - headR * 0.38, shoulderCx + w * 0.31, hipY + h * 0.13);
    ctx.lineTo(shoulderCx - w * 0.31, hipY + h * 0.13);
    ctx.quadraticCurveTo(headCx - headR * 1.44, headY - headR * 0.38, headCx, headY - headR * 1.48);
    ctx.closePath();
    ctx.fill();
    ctx.strokeStyle = '#F43F5E';
    ctx.lineWidth = strokeW * 0.80;
    ctx.stroke();
    ctx.restore();
  }

  // Smooth curved legs
  ctx.beginPath();
  ctx.moveTo(cx, hipY); ctx.quadraticCurveTo(lKnee[0], lKnee[1], lFoot[0], lFoot[1]);
  ctx.moveTo(cx, hipY); ctx.quadraticCurveTo(rKnee[0], rKnee[1], rFoot[0], rFoot[1]);
  ctx.stroke();

  // Curved Spine & Neck
  ctx.beginPath();
  ctx.moveTo(headCx, headY + headR * 0.9);
  ctx.quadraticCurveTo(shoulderCx, (shoulderY + hipY) * 0.5, cx, hipY);
  ctx.stroke();

  // Hero subtle emerald lapel trim for role distinction
  if (isHero) {
    ctx.save();
    ctx.strokeStyle = '#34D399';
    ctx.lineWidth = strokeW * 0.45;
    ctx.beginPath();
    ctx.moveTo(shoulderCx - w * 0.04, shoulderY + h * 0.02);
    ctx.lineTo(shoulderCx + w * 0.04, shoulderY + h * 0.02);
    ctx.stroke();
    ctx.restore();
  }

  // Smooth curved arms
  ctx.beginPath();
  ctx.moveTo(shoulderCx, shoulderY); ctx.quadraticCurveTo(lElbow[0], lElbow[1], lHand[0], lHand[1]);
  ctx.moveTo(shoulderCx, shoulderY); ctx.quadraticCurveTo(rElbow[0], rElbow[1], rHand[0], rHand[1]);
  ctx.stroke();

  // Head
  ctx.fillStyle = isHacker ? '#070A12' : '#0F172A';
  ctx.beginPath();
  ctx.arc(headCx, headY, headR, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Signature swaying two hair tufts for victim/hero stickman
  if (!isHacker) {
    const tuftSway = Math.sin(t * 5.2 + dt) * (headR * 0.08);
    ctx.beginPath();
    ctx.moveTo(headCx - headR * 0.08, headY - headR);
    ctx.quadraticCurveTo(headCx + headR * 0.12, headY - headR * 1.38, headCx + headR * 0.36 + tuftSway, headY - headR * 1.32);
    ctx.moveTo(headCx + headR * 0.04, headY - headR);
    ctx.quadraticCurveTo(headCx + headR * 0.28, headY - headR * 1.30, headCx + headR * 0.48 + tuftSway, headY - headR * 1.15);
    ctx.stroke();
  }
  ctx.restore();

  // Facial Expressions + Real Audio-Reactive Lip-Sync
  const blink = (t % 3.4) > 3.24;
  ctx.save();
  if (isHacker) {
    // Clean sharp coral hacker eyes + subtle visor scan glint
    ctx.fillStyle = '#F43F5E';
    ctx.shadowColor = '#F43F5E';
    ctx.shadowBlur = 4;
    ctx.beginPath();
    ctx.moveTo(headCx - headR * 0.52, headY - headR * 0.14);
    ctx.lineTo(headCx - headR * 0.10, headY - headR * 0.02);
    ctx.lineTo(headCx - headR * 0.45, headY + headR * 0.08);
    ctx.closePath();
    ctx.fill();

    ctx.beginPath();
    ctx.moveTo(headCx + headR * 0.52, headY - headR * 0.14);
    ctx.lineTo(headCx + headR * 0.10, headY - headR * 0.02);
    ctx.lineTo(headCx + headR * 0.45, headY + headR * 0.08);
    ctx.closePath();
    ctx.fill();

    // Visor horizontal scan glint
    const glintX = headCx - headR * 0.5 + ((t * 1.1) % 1) * headR;
    ctx.strokeStyle = 'rgba(255, 228, 230, 0.55)';
    ctx.lineWidth = 1.8;
    ctx.beginPath();
    ctx.moveTo(glintX, headY - headR * 0.14);
    ctx.lineTo(glintX + headR * 0.08, headY + headR * 0.04);
    ctx.stroke();

    // Subtle strike spark at hand on strike
    if (pose === 'hacker_strike') {
      ctx.strokeStyle = '#FBBF24';
      ctx.lineWidth = 2;
      const spark = Math.sin(t * 28) * 5;
      ctx.beginPath();
      ctx.moveTo(rHand[0], rHand[1]); ctx.lineTo(rHand[0] - 10, rHand[1] - 8 + spark);
      ctx.moveTo(rHand[0], rHand[1]); ctx.lineTo(rHand[0] - 12, rHand[1] + 6 - spark);
      ctx.stroke();
    }
  } else {
    const eyeY = headY - headR * 0.12;
    const eyeOff = headR * 0.34;
    ctx.fillStyle = '#F8FAFC';
    if (blink) {
      ctx.strokeStyle = '#F8FAFC';
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.moveTo(headCx - eyeOff - 7, eyeY); ctx.lineTo(headCx - eyeOff + 7, eyeY);
      ctx.moveTo(headCx + eyeOff - 7, eyeY); ctx.lineTo(headCx + eyeOff + 7, eyeY);
      ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.ellipse(headCx - eyeOff, eyeY, headR * 0.11, headR * 0.18, 0, 0, Math.PI * 2);
      ctx.ellipse(headCx + eyeOff, eyeY, headR * 0.11, headR * 0.18, 0, 0, Math.PI * 2);
      ctx.fill();
    }

    ctx.strokeStyle = '#F8FAFC';
    ctx.lineWidth = Math.max(2.5, strokeW * 0.36);
    if (pose === 'shocked' || pose === 'panic_run') {
      ctx.beginPath();
      ctx.moveTo(headCx - eyeOff - 10, eyeY - headR * 0.28); ctx.lineTo(headCx - eyeOff + 8, eyeY - headR * 0.38);
      ctx.moveTo(headCx + eyeOff + 10, eyeY - headR * 0.28); ctx.lineTo(headCx + eyeOff - 8, eyeY - headR * 0.38);
      ctx.stroke();

      ctx.fillStyle = '#0B0F19';
      ctx.beginPath();
      ctx.ellipse(headCx, headY + headR * 0.38, headR * 0.25, headR * (0.26 + speech * 0.08), 0, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();

      // Clean comic shock rays & animated sweat drops
      ctx.strokeStyle = '#FBBF24';
      ctx.lineWidth = 2.4;
      ctx.beginPath();
      ctx.moveTo(headCx - headR * 1.18, headY - headR * 0.45); ctx.lineTo(headCx - headR * 1.44, headY - headR * 0.62);
      ctx.moveTo(headCx + headR * 1.18, headY - headR * 0.45); ctx.lineTo(headCx + headR * 1.44, headY - headR * 0.62);
      ctx.stroke();

      ctx.strokeStyle = '#38BDF8';
      ctx.lineWidth = 2.2;
      const dropY = Math.sin(t * 14) * 5;
      ctx.beginPath();
      ctx.ellipse(headCx - headR * 1.24, headY - headR * 0.15 + dropY, 6, 4, -0.4, 0, Math.PI * 2);
      ctx.ellipse(headCx + headR * 1.24, headY - headR * 0.15 - dropY, 6, 4, 0.4, 0, Math.PI * 2);
      ctx.stroke();
    } else if (pose === 'thinking') {
      ctx.beginPath();
      ctx.moveTo(headCx - eyeOff - 8, eyeY - headR * 0.26); ctx.lineTo(headCx - eyeOff + 8, eyeY - headR * 0.30);
      ctx.moveTo(headCx + eyeOff - 8, eyeY - headR * 0.34); ctx.quadraticCurveTo(headCx + eyeOff, eyeY - headR * 0.44, headCx + eyeOff + 8, eyeY - headR * 0.28);
      ctx.stroke();
      if (speech > 0.08) {
        ctx.fillStyle = '#0B0F19';
        ctx.beginPath();
        ctx.ellipse(headCx, headY + headR * 0.40, headR * 0.18, headR * (0.07 + speech * 0.14), 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.stroke();
      } else {
        ctx.beginPath();
        ctx.moveTo(headCx - headR * 0.22, headY + headR * 0.42);
        ctx.quadraticCurveTo(headCx, headY + headR * 0.35, headCx + headR * 0.24, headY + headR * 0.42);
        ctx.stroke();
      }
    } else if (pose === 'celebrate' || pose === 'shield_hero') {
      if (speech > 0.08) {
        ctx.fillStyle = '#0B0F19';
        ctx.beginPath();
        ctx.ellipse(headCx, headY + headR * 0.35, headR * 0.22, headR * (0.08 + speech * 0.16), 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.stroke();
      } else {
        ctx.beginPath();
        ctx.arc(headCx, headY + headR * 0.24, headR * 0.28, 0.15 * Math.PI, 0.85 * Math.PI);
        ctx.stroke();
      }
    } else {
      if (speech > 0.08) {
        ctx.fillStyle = '#0B0F19';
        ctx.beginPath();
        ctx.ellipse(headCx, headY + headR * 0.38, headR * 0.20, headR * (0.06 + speech * 0.15), 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.stroke();
      } else {
        ctx.beginPath();
        ctx.moveTo(headCx - headR * 0.18, headY + headR * 0.38);
        ctx.quadraticCurveTo(headCx, headY + headR * 0.46, headCx + headR * 0.20, headY + headR * 0.36);
        ctx.stroke();
      }
    }
  }
  ctx.restore();

  // Held accessories & contextual prop accents
  if (pose === 'idle_browse') {
    ctx.save();
    ctx.fillStyle = '#1E293B';
    ctx.strokeStyle = '#38BDF8';
    ctx.lineWidth = 2.2;
    ctx.shadowColor = 'rgba(0, 0, 0, 0.4)';
    ctx.shadowBlur = 5;
    rrect(ctx, rHand[0] - 8, rHand[1] - 26, 28, 44, 6);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = 'rgba(56, 189, 248, 0.35)';
    ctx.fillRect(rHand[0] - 4, rHand[1] - 20, 20, 28);
    // Expanding notification ping rings from phone
    const pingP = (t * 1.4) % 1;
    ctx.strokeStyle = `rgba(56, 189, 248, ${(0.45 * (1 - pingP)).toFixed(3)})`;
    ctx.lineWidth = 1.8;
    ctx.beginPath();
    ctx.arc(rHand[0] + 6, rHand[1] - 18, 12 + pingP * 20, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();
  } else if (pose === 'typing') {
    ctx.save();
    ctx.fillStyle = '#111827';
    ctx.strokeStyle = '#38BDF8';
    ctx.lineWidth = 2;
    rrect(ctx, rHand[0] - 24, rHand[1] - 4, 48, 28, 5);
    ctx.fill();
    ctx.stroke();
    for (let k = 0; k < 3; k++) {
      const kf = (t * 2.2 + k * 0.33) % 1;
      ctx.fillStyle = `rgba(56, 189, 248, ${(0.65 * (1 - kf)).toFixed(3)})`;
      ctx.fillRect(rHand[0] - 14 + k * 12, rHand[1] - 12 - kf * 22, 4, 4);
    }
    ctx.restore();
  } else if (pose === 'thinking') {
    ctx.save();
    ctx.font = `900 ${Math.round(w * 0.11)}px B1Caption`;
    ctx.fillStyle = '#A78BFA';
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 5;
    ctx.fillText('?', headCx + headR * 1.05, headY - headR * 0.8 + Math.sin(t * 4) * 5);
    ctx.restore();
  } else if (pose === 'pointing') {
    ctx.save();
    const beam = ctx.createLinearGradient(rHand[0], rHand[1], rHand[0] + w * 0.25, rHand[1] + h * 0.22);
    beam.addColorStop(0, 'rgba(56, 189, 248, 0.48)');
    beam.addColorStop(1, 'rgba(56, 189, 248, 0)');
    ctx.strokeStyle = beam;
    ctx.lineWidth = 2.5;
    ctx.setLineDash([6, 6]);
    ctx.beginPath();
    ctx.moveTo(rHand[0], rHand[1]);
    ctx.lineTo(rHand[0] + w * 0.24, rHand[1] + h * 0.22);
    ctx.stroke();
    ctx.restore();
  } else if (pose === 'shield_hero') {
    const sx = rHand[0] + w * 0.05;
    const sy = rHand[1] - h * 0.02;
    drawShieldBadge(ctx, sx, sy, w * 0.16, '#10B981', 'rgba(16, 185, 129, 0.16)');
    // Deflected spark particles bouncing off shield rim
    ctx.save();
    for (let k = 0; k < 2; k++) {
      const kp = (t * 1.8 + k * 0.5) % 1;
      const px = sx + w * 0.14 + kp * w * 0.10;
      const py = sy - h * 0.06 + (k === 0 ? -kp : kp) * h * 0.12;
      ctx.fillStyle = `rgba(244, 63, 94, ${(0.75 * (1 - kp)).toFixed(3)})`;
      ctx.beginPath();
      ctx.arc(px, py, 3.2, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.restore();
  } else if (pose === 'celebrate') {
    ctx.save();
    for (let k = 0; k < 4; k++) {
      const kf = (t * 1.5 + k * 0.25) % 1;
      const sx = cx + Math.sin(k * 2.1) * (w * 0.34);
      const sy = headY - kf * (h * 0.18);
      ctx.fillStyle = k % 2 === 0
        ? `rgba(52, 211, 153, ${(0.75 * (1 - kf)).toFixed(3)})`
        : `rgba(251, 191, 36, ${(0.75 * (1 - kf)).toFixed(3)})`;
      ctx.beginPath();
      ctx.arc(sx, sy, 3.8, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.restore();
  }

  ctx.restore();
}

export function drawCyberWidget(ctx, l, dt, t, shotDuration = 4.0, speechLevel = 0) {
  const w = l.w || 1000;
  const h = l.h || 480;
  const kind = l.kind || 'terminal';
  const data = l.data || {};
  const sc = Math.min(w / 960, h / 440);
  const dur = Math.max(1.2, shotDuration || 4.0);
  const speech = clamp(speechLevel);

  ctx.save();
  // Dark-Editorial Glass Card Container
  const accentBarCol = (kind === 'checklist' || (kind === 'threat_meter' && data.safe))
    ? '#10B981'
    : (kind === 'notification' || kind === 'url_inspect' || kind === 'password_crack' || kind === 'threat_meter')
      ? '#F43F5E'
      : '#38BDF8';
  const borderCol = 'rgba(148, 163, 184, 0.24)';

  ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
  ctx.shadowBlur = 20 + speech * 3;
  ctx.shadowOffsetY = 8;
  ctx.fillStyle = 'rgba(17, 24, 39, 0.95)';
  rrect(ctx, 0, 0, w, h, 24);
  ctx.fill();
  ctx.shadowBlur = 0;
  ctx.shadowOffsetY = 0;

  // Clip interior contents strictly inside rounded container so text never overflows
  ctx.save();
  rrect(ctx, 0, 0, w, h, 24);
  ctx.clip();

  // Subtle top specular reflection highlight
  const topSheen = ctx.createLinearGradient(0, 0, 0, h * 0.22);
  topSheen.addColorStop(0, 'rgba(255, 255, 255, 0.045)');
  topSheen.addColorStop(1, 'rgba(255, 255, 255, 0)');
  ctx.fillStyle = topSheen;
  ctx.fillRect(0, 0, w, h * 0.22);

  // Clean semantic top accent bar
  ctx.fillStyle = accentBarCol;
  rrect(ctx, 28, 0, w - 56, 3, 1.5);
  ctx.fill();

  const pad = Math.round(28 * Math.max(0.75, sc));

  if (kind === 'notification') {
    const shakeX = dt < 0.65 ? Math.sin(dt * 38) * 6 * (1 - dt / 0.65) : 0;
    ctx.translate(shakeX, 0);
    const badgeText = String(data.badge || 'ACİL GÜVENLİK BİLDİRİMİ').toLocaleUpperCase('tr-TR');
    ctx.font = `800 ${Math.round(20 * sc)}px B1Caption`;
    const bw = Math.min(w - pad * 2 - 110 * sc, ctx.measureText(badgeText).width + 48);
    ctx.fillStyle = 'rgba(244, 63, 94, 0.14)';
    ctx.strokeStyle = 'rgba(244, 63, 94, 0.65)';
    ctx.lineWidth = 1.8;
    rrect(ctx, pad, pad, bw, 42 * sc, 10);
    ctx.fill();
    ctx.stroke();

    drawWarningTriangle(ctx, pad + 22 * sc, pad + 21 * sc, 11 * sc, '#FBBF24');
    ctx.fillStyle = '#FBBF24';
    ctx.textBaseline = 'middle';
    ctx.fillText(badgeText, pad + 42 * sc, pad + 21 * sc, Math.max(40, bw - 50 * sc));

    const remSec = Math.max(0, 899 - Math.floor(t * 2));
    const timerTxt = `${String(Math.floor(remSec / 60)).padStart(2, '0')}:${String(remSec % 60).padStart(2, '0')}`;
    ctx.textAlign = 'right';
    ctx.fillStyle = '#F43F5E';
    ctx.font = `800 ${Math.round(24 * sc)}px B1Caption`;
    ctx.fillText(timerTxt, w - pad, pad + 21 * sc);
    ctx.textAlign = 'left';

    ctx.fillStyle = '#F8FAFC';
    ctx.textBaseline = 'top';
    ctx.font = `800 ${Math.round(31 * sc)}px B1Caption`;
    ctx.fillText(String(data.sender || 'Bilinmeyen Gönderici'), pad, pad + 60 * sc, w - pad * 2);

    ctx.font = `600 ${Math.round(25 * sc)}px B1Caption`;
    ctx.fillStyle = '#CBD5E1';
    const lines = wrapTextSafe(ctx, data.body || 'Hesabınızı hemen doğrulayın!', w - pad * 2, 3);
    lines.forEach((line, i) => ctx.fillText(line, pad, pad + 108 * sc + i * 35 * sc));

    // Action trap CTA bar + Animated Pointer Cursor glide
    const btnY = h - pad - 58 * sc;
    const btnW = w - pad * 2;
    const cursorP = easeOutCubic(dt / (dur * 0.55));
    const clicked = dt >= dur * 0.52;

    ctx.fillStyle = clicked ? '#E11D48' : '#F43F5E';
    if (clicked) {
      ctx.shadowColor = 'rgba(244, 63, 94, 0.35)';
      ctx.shadowBlur = 6;
    }
    rrect(ctx, pad, btnY, btnW, 54 * sc, 12);
    ctx.fill();
    ctx.shadowBlur = 0;

    ctx.fillStyle = '#FFFFFF';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.font = `800 ${Math.round(23 * sc)}px B1Caption`;
    ctx.fillText('HEMEN DOĞRULA  >>  TUZAK BAĞLANTI', w / 2, btnY + 27 * sc, btnW - 24);
    ctx.textAlign = 'left';

    const curX = w * (0.90 - 0.14 * cursorP);
    const curY = btnY - 60 * sc + 86 * sc * cursorP;
    drawPointerCursor(ctx, curX, curY, sc, clicked && (dt < dur * 0.65));
  } else if (kind === 'terminal') {
    const headH = Math.round(48 * sc);
    ctx.fillStyle = 'rgba(30, 41, 59, 0.88)';
    rrect(ctx, 2, 2, w - 4, headH, 20);
    ctx.fill();

    for (const [colHex, idx] of [['#F43F5E', 0], ['#FBBF24', 1], ['#10B981', 2]]) {
      ctx.fillStyle = colHex;
      ctx.beginPath();
      ctx.arc(pad + idx * 24 * sc, headH / 2 + 2, 7 * sc, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.font = `700 ${Math.round(19 * sc)}px B1Caption`;
    ctx.fillStyle = '#94A3B8';
    ctx.textBaseline = 'middle';
    ctx.fillText(String(data.header || 'root@kali:~/exploit'), pad + 82 * sc, headH / 2 + 2, w - pad * 2 - 90 * sc);

    const rawLines = Array.isArray(data.lines) && data.lines.length
      ? data.lines.map(String)
      : ['$ scan --target', '[+] Session token captured', '[!] ACCESS GRANTED'];
    const wrappedEntries = [];
    ctx.font = `700 ${Math.round(23 * sc)}px B1Caption`;
    for (const rawLine of rawLines) {
      const col = rawLine.includes('[!]') || rawLine.includes('>>>')
        ? '#F43F5E'
        : rawLine.includes('[+]')
          ? '#FBBF24'
          : '#34D399';
      const isAlert = rawLine.includes('[!]') || rawLine.includes('>>>');
      for (const wl of wrapTextSafe(ctx, rawLine, w - pad * 2 - 20 * sc, 2)) {
        wrappedEntries.push({ text: wl, col, isAlert });
      }
    }

    const totalChars = wrappedEntries.reduce((acc, s) => acc + s.text.length, 0);
    const paceSpan = Math.min(dur * 0.72, 2.8);
    let budget = Math.max(10, Math.floor(easeOutCubic(dt / paceSpan) * totalChars));

    ctx.textBaseline = 'top';
    let curY = headH + 22 * sc;
    const maxTermY = h - pad - 56 * sc;
    for (let i = 0; i < wrappedEntries.length && curY < maxTermY; i++) {
      if (budget <= 0) break;
      const entry = wrappedEntries[i];
      const vis = entry.text.slice(0, budget);
      budget -= vis.length;
      if (entry.isAlert && vis.length > 3) {
        const pulse = 0.10 + 0.06 * Math.sin(t * 8);
        ctx.fillStyle = `rgba(244, 63, 94, ${pulse.toFixed(3)})`;
        rrect(ctx, pad - 8, curY - 3, w - pad * 2 + 16, Math.round(32 * sc), 6);
        ctx.fill();
      }
      ctx.fillStyle = entry.col;
      const cursor = (budget <= 0 || i === wrappedEntries.length - 1) && Math.floor(t * 4) % 2 === 0 ? ' _' : '';
      ctx.fillText(vis + cursor, pad, curY);
      curY += Math.round(36 * sc);
    }

    // Animated progress bar + clean telemetry footer
    const barW = w - pad * 2;
    const barY = h - pad - 28 * sc;
    const prog = easeOutCubic(dt / Math.min(dur * 0.78, 3.0));
    ctx.fillStyle = 'rgba(148, 163, 184, 0.16)';
    rrect(ctx, pad, barY, barW, 16 * sc, 8);
    ctx.fill();
    ctx.fillStyle = prog > 0.7 ? '#F43F5E' : '#38BDF8';
    rrect(ctx, pad, barY, Math.max(16, barW * prog), 16 * sc, 8);
    ctx.fill();
    ctx.font = `800 ${Math.round(17 * sc)}px B1Caption`;
    ctx.fillStyle = '#94A3B8';
    const hexTick = ((Math.floor(t * 18) * 4919) & 0xFFFF).toString(16).toUpperCase().padStart(4, '0');
    ctx.fillText(`SYS 0x${hexTick} // PACKETS`, pad, barY - 24 * sc);
    ctx.fillStyle = '#F8FAFC';
    ctx.textAlign = 'right';
    ctx.fillText(`${Math.round(prog * 100)}%`, w - pad, barY - 24 * sc);
    ctx.textAlign = 'left';
  } else if (kind === 'url_inspect') {
    const fakeUrl = String(data.fakeUrl || 'https://hesap-dogru1ama.com');
    const realUrl = String(data.realUrl || 'resmi-site.com');
    const stamp = String(data.stamp || 'HTTPS ≠ GÜVENLİ!').toLocaleUpperCase('tr-TR');

    ctx.font = `700 ${Math.round(20 * sc)}px B1Caption`;
    ctx.fillStyle = '#94A3B8';
    ctx.textBaseline = 'top';
    ctx.fillText('TARAYICI ADRES ÇUBUĞU ANALİZİ', pad, pad);

    // URL Address Card
    const barY = pad + 42 * sc;
    const barH = 74 * sc;
    const lockCracked = dt >= dur * 0.38;
    ctx.fillStyle = '#0F172A';
    ctx.strokeStyle = lockCracked ? '#F43F5E' : '#38BDF8';
    ctx.lineWidth = 2;
    rrect(ctx, pad, barY, w - pad * 2, barH, 14);
    ctx.fill();
    ctx.stroke();

    // Crisp Vector Padlock Icon
    drawPadlockVector(ctx, pad + 34 * sc, barY + barH / 2 - 3 * sc, 16 * sc, lockCracked ? '#F43F5E' : '#10B981', lockCracked);

    // Render URL with coral highlight on suspicious character patterns
    const urlX = pad + 70 * sc;
    const urlMaxW = w - pad * 2 - 86 * sc;
    ctx.font = `800 ${Math.round(26 * sc)}px B1Caption`;
    ctx.textBaseline = 'middle';
    const suspicious = ['1', '.vip', 'OR ', '1=1', 'http://', 'lns'];
    let suspectStart = -1;
    let suspectLen = 0;
    for (const pat of suspicious) {
      const idx = fakeUrl.indexOf(pat);
      if (idx !== -1) {
        suspectStart = idx;
        suspectLen = pat.length;
        break;
      }
    }
    const urlFullW = ctx.measureText(fakeUrl).width;
    const drawScale = urlFullW > urlMaxW ? urlMaxW / urlFullW : 1;
    ctx.save();
    ctx.translate(urlX, barY + barH / 2);
    ctx.scale(drawScale, drawScale);
    if (suspectStart !== -1 && lockCracked) {
      const pre = fakeUrl.slice(0, suspectStart);
      const mid = fakeUrl.slice(suspectStart, suspectStart + suspectLen);
      const post = fakeUrl.slice(suspectStart + suspectLen);
      const preW = ctx.measureText(pre).width;
      const midW = ctx.measureText(mid).width;
      ctx.fillStyle = 'rgba(244, 63, 94, 0.24)';
      rrect(ctx, preW - 3, -18 * sc, midW + 6, 36 * sc, 6);
      ctx.fill();
      ctx.fillStyle = '#F8FAFC';
      ctx.fillText(pre, 0, 0);
      ctx.fillStyle = '#F43F5E';
      ctx.fillText(mid, preW, 0);
      ctx.fillStyle = '#F8FAFC';
      ctx.fillText(post, preW + midW, 0);
    } else {
      ctx.fillStyle = '#F8FAFC';
      ctx.fillText(fakeUrl, 0, 0);
    }
    ctx.restore();

    // Vector Magnifying Glass sweeping across URL
    const scanFrac = easeOutCubic(dt / Math.min(dur * 0.48, 1.3));
    const magX = pad + 140 * sc + scanFrac * Math.min(urlFullW * drawScale, w - pad * 2 - 220 * sc) * 0.65;
    const magY = barY + barH / 2;
    const magR = 30 * sc;
    ctx.save();
    ctx.strokeStyle = '#FBBF24';
    ctx.lineWidth = 4.2 * sc;
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 6;
    ctx.beginPath();
    ctx.arc(magX, magY, magR, 0, Math.PI * 2);
    ctx.stroke();
    ctx.lineCap = 'round';
    ctx.lineWidth = 6.5 * sc;
    ctx.beginPath();
    ctx.moveTo(magX + magR * 0.72, magY + magR * 0.72);
    ctx.lineTo(magX + magR * 1.48, magY + magR * 1.48);
    ctx.stroke();
    ctx.shadowBlur = 0;
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.40)';
    ctx.lineWidth = 2 * sc;
    ctx.beginPath();
    ctx.arc(magX, magY, magR * 0.70, Math.PI * 1.15, Math.PI * 1.55);
    ctx.stroke();
    ctx.restore();

    // Real domain comparison row
    ctx.textBaseline = 'top';
    ctx.font = `700 ${Math.round(23 * sc)}px B1Caption`;
    ctx.fillStyle = '#34D399';
    ctx.fillText(`GERÇEK RESMİ ADRES:  ${realUrl}`, pad, barY + barH + 26 * sc, w - pad * 2);

    // Clean Editorial Stamp Badge
    const stampDelay = Math.min(0.32, dur * 0.35);
    if (dt >= stampDelay) {
      const sProg = easeOutBack((dt - stampDelay) / 0.22);
      const stCx = w * 0.5;
      const stCy = h - pad - 52 * sc;
      if (dt - stampDelay < 0.35) {
        const ringP = easeOutCubic((dt - stampDelay) / 0.35);
        ctx.strokeStyle = `rgba(244, 63, 94, ${(0.35 * (1 - ringP)).toFixed(3)})`;
        ctx.lineWidth = 2.5;
        rrect(ctx, stCx - (270 + ringP * 70) * sc, stCy - (32 + ringP * 16) * sc, (540 + ringP * 140) * sc, (64 + ringP * 32) * sc, 14);
        ctx.stroke();
      }
      ctx.save();
      ctx.translate(stCx, stCy);
      ctx.rotate(-0.025);
      ctx.scale(1.26 - 0.26 * sProg, 1.26 - 0.26 * sProg);
      ctx.fillStyle = 'rgba(244, 63, 94, 0.16)';
      ctx.strokeStyle = '#F43F5E';
      ctx.lineWidth = 2.4;
      const stW = Math.min(w - pad * 2, 540 * sc);
      rrect(ctx, -stW / 2, -32 * sc, stW, 64 * sc, 12);
      ctx.fill();
      ctx.stroke();
      ctx.fillStyle = '#F43F5E';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.font = `900 ${Math.round(27 * sc)}px B1Caption`;
      ctx.fillText(stamp, 0, 0, stW - 24);
      ctx.restore();
    }
  } else if (kind === 'network_mitm') {
    const leftLbl = String(data.left || 'CİHAZIN');
    const midLbl = String(data.middle || 'SAHTE AĞ');
    const rightLbl = String(data.right || 'SUNUCU');
    const packets = Array.isArray(data.packets) && data.packets.length ? data.packets.map(String) : ['ŞİFRE', 'OTP', 'ÇEREZ'];

    const cy = h * 0.48;
    const x1 = w * 0.16;
    const x2 = w * 0.50;
    const x3 = w * 0.84;

    // Curved Bezier traffic lanes
    ctx.strokeStyle = 'rgba(56, 189, 248, 0.42)';
    ctx.lineWidth = 3.5;
    ctx.beginPath();
    ctx.moveTo(x1, cy); ctx.quadraticCurveTo((x1 + x2) / 2, cy - 28 * sc, x2, cy);
    ctx.stroke();
    ctx.strokeStyle = 'rgba(244, 63, 94, 0.52)';
    ctx.beginPath();
    ctx.moveTo(x2, cy); ctx.quadraticCurveTo((x2 + x3) / 2, cy + 28 * sc, x3, cy);
    ctx.stroke();

    // Animated intercepted packets with clean pills
    for (let i = 0; i < packets.length; i++) {
      const pFrac = (t * 0.55 + i * 0.33) % 1;
      const px = x1 + (x3 - x1) * pFrac;
      const intercepted = px >= x2 - 18;
      const py = cy + Math.sin(pFrac * Math.PI * 2) * (-20 * sc);
      ctx.save();
      const pCol = intercepted ? '#F43F5E' : '#38BDF8';
      ctx.fillStyle = pCol;
      ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
      ctx.shadowBlur = 5;
      rrect(ctx, px - 44 * sc, py - 18 * sc, 88 * sc, 36 * sc, 8);
      ctx.fill();
      ctx.shadowBlur = 0;
      ctx.fillStyle = '#070B14';
      ctx.font = `900 ${Math.round(16 * sc)}px B1Caption`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(packets[i], px, py, 78 * sc);
      ctx.restore();
    }

    // Nodes with subtle sonar rings & crisp vector icons
    const nodes = [
      [x1, leftLbl, '#38BDF8', 'phone'],
      [x2, midLbl, '#F43F5E', 'hacker'],
      [x3, rightLbl, '#10B981', 'server'],
    ];
    for (const [nx, label, col, icon] of nodes) {
      ctx.save();
      const sonar = (t * 0.9 + (icon === 'hacker' ? 0.5 : 0)) % 1;
      ctx.strokeStyle = col;
      ctx.globalAlpha *= (1 - sonar) * 0.28;
      ctx.lineWidth = 1.8;
      ctx.beginPath();
      ctx.arc(nx, cy, (46 + sonar * 22) * sc, 0, Math.PI * 2);
      ctx.stroke();
      ctx.restore();

      ctx.save();
      ctx.fillStyle = '#0F172A';
      ctx.strokeStyle = col;
      ctx.lineWidth = 3;
      ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
      ctx.shadowBlur = 6;
      ctx.beginPath();
      ctx.arc(nx, cy, 46 * sc, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
      ctx.shadowBlur = 0;

      // Vector symbol inside node
      ctx.strokeStyle = col;
      ctx.fillStyle = col;
      ctx.lineWidth = 2.5 * sc;
      if (icon === 'phone') {
        rrect(ctx, nx - 11 * sc, cy - 17 * sc, 22 * sc, 34 * sc, 4);
        ctx.stroke();
      } else if (icon === 'hacker') {
        ctx.beginPath();
        ctx.moveTo(nx, cy - 18 * sc); ctx.lineTo(nx + 16 * sc, cy + 15 * sc); ctx.lineTo(nx - 16 * sc, cy + 15 * sc); ctx.closePath();
        ctx.stroke();
      } else {
        rrect(ctx, nx - 16 * sc, cy - 14 * sc, 32 * sc, 11 * sc, 3); ctx.stroke();
        rrect(ctx, nx - 16 * sc, cy + 3 * sc, 32 * sc, 11 * sc, 3); ctx.stroke();
      }

      ctx.fillStyle = '#F8FAFC';
      ctx.font = `800 ${Math.round(19 * sc)}px B1Caption`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillText(label, nx, cy + 60 * sc, w * 0.28);
      ctx.restore();
    }

    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.fillStyle = '#FBBF24';
    ctx.font = `800 ${Math.round(21 * sc)}px B1Caption`;
    const capturedCount = 12 + Math.floor(t * 7);
    ctx.fillText(`VERİ TRAFİĞİ ORTADAN YAKALANIYOR · ${capturedCount} PAKET`, w / 2, pad);
    ctx.textAlign = 'left';
  } else if (kind === 'password_crack') {
    const weak = String(data.weak || data.password || 'ahmet1998');
    const weakTime = String(data.weakTime || data.crackTime || '0.4 SANİYE!');
    const strong = String(data.strong || 'K4p1!Mavi#92Z');
    const rate = String(data.rate || '4.82M DENEME/SN');
    const prog = clamp(dt / Math.min(dur * 0.65, 1.8));

    // Character scramble reveal left-to-right
    const glyphs = '#$%@!974XKVZ';
    const lockedCount = Math.floor(prog * weak.length);
    let shownWeak = '';
    for (let c = 0; c < weak.length; c++) {
      if (c < lockedCount || prog >= 0.99) shownWeak += weak[c];
      else shownWeak += glyphs[(Math.floor(t * 24) + c * 5) % glyphs.length];
    }

    ctx.textBaseline = 'top';
    ctx.font = `800 ${Math.round(21 * sc)}px B1Caption`;
    ctx.fillStyle = '#F43F5E';
    ctx.fillText(`GPU KABA KUVVET HIZI: ${rate}`, pad, pad);

    const row1Y = pad + 46 * sc;
    ctx.fillStyle = 'rgba(244, 63, 94, 0.12)';
    ctx.strokeStyle = 'rgba(244, 63, 94, 0.65)';
    ctx.lineWidth = 2;
    rrect(ctx, pad, row1Y, w - pad * 2, 118 * sc, 14);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = '#F8FAFC';
    ctx.font = `800 ${Math.round(30 * sc)}px B1Caption`;
    ctx.fillText(`ZAYIF: "${shownWeak}"`, pad + 20 * sc, row1Y + 18 * sc, (w - pad * 2) * 0.56);
    ctx.fillStyle = '#FBBF24';
    ctx.font = `900 ${Math.round(24 * sc)}px B1Caption`;
    ctx.textAlign = 'right';
    ctx.fillText(prog >= 0.88 ? `KIRILDI: ${weakTime}` : `TARANIYOR %${Math.round(prog * 100)}`, w - pad - 20 * sc, row1Y + 22 * sc);
    ctx.textAlign = 'left';

    ctx.fillStyle = 'rgba(148, 163, 184, 0.14)';
    rrect(ctx, pad + 20 * sc, row1Y + 74 * sc, w - pad * 2 - 40 * sc, 18 * sc, 9);
    ctx.fill();
    ctx.fillStyle = '#F43F5E';
    rrect(ctx, pad + 20 * sc, row1Y + 74 * sc, Math.max(18, (w - pad * 2 - 40 * sc) * prog), 18 * sc, 9);
    ctx.fill();

    const row2Y = row1Y + 138 * sc;
    ctx.fillStyle = 'rgba(16, 185, 129, 0.11)';
    ctx.strokeStyle = 'rgba(16, 185, 129, 0.65)';
    rrect(ctx, pad, row2Y, w - pad * 2, 118 * sc, 14);
    ctx.fill();
    ctx.stroke();
    drawPadlockVector(ctx, pad + 42 * sc, row2Y + 42 * sc, 16 * sc, '#10B981', false);
    ctx.fillStyle = '#34D399';
    ctx.font = `800 ${Math.round(28 * sc)}px B1Caption`;
    ctx.fillText(`GÜÇLÜ (14+): "${strong}"`, pad + 72 * sc, row2Y + 20 * sc, w - pad * 2 - 90 * sc);
    ctx.fillStyle = '#CBD5E1';
    ctx.font = `700 ${Math.round(22 * sc)}px B1Caption`;
    ctx.fillText('KIRILMA SÜRESİ: YÜZYILLAR + 2FA KALKAN', pad + 72 * sc, row2Y + 66 * sc, w - pad * 2 - 90 * sc);
  } else if (kind === 'threat_meter') {
    const target = clamp((Number(data.level) || 96) / 100);
    const isSafe = Boolean(data.safe);
    const col = isSafe ? '#10B981' : '#F43F5E';
    const settle = easeOutCubic(dt / Math.min(dur * 0.55, 1.3));
    const liveJitter = settle > 0.9 ? Math.sin(t * 7.5) * 0.004 : 0;
    const cur = clamp(target * settle + liveJitter);

    const gcx = w / 2;
    const gcy = h * 0.72;
    const gr = Math.min(w * 0.30, h * 0.48);

    // 24 Segmented Clean Radial Gauge
    const segs = 24;
    const activeSegs = Math.round(cur * segs);
    ctx.lineCap = 'butt';
    ctx.lineWidth = 22 * sc;
    for (let s = 0; s < segs; s++) {
      const a0 = Math.PI + (s / segs) * Math.PI + 0.016;
      const a1 = Math.PI + ((s + 1) / segs) * Math.PI - 0.016;
      const on = s < activeSegs;
      ctx.strokeStyle = on ? col : 'rgba(148, 163, 184, 0.14)';
      if (on) {
        ctx.shadowColor = col;
        ctx.shadowBlur = 4;
      } else {
        ctx.shadowBlur = 0;
      }
      ctx.beginPath();
      ctx.arc(gcx, gcy, gr, a0, a1);
      ctx.stroke();
    }
    ctx.shadowBlur = 0;

    // Outer-Track Tapered Chevron Indicator (never crosses centered %XX readout)
    const ang = Math.PI + Math.PI * cur;
    const tipR = gr + 6 * sc;
    const baseR = gr + 26 * sc;
    const perp = 9 * sc;
    const tipX = gcx + Math.cos(ang) * tipR;
    const tipY = gcy + Math.sin(ang) * tipR;
    const bx = gcx + Math.cos(ang) * baseR;
    const by = gcy + Math.sin(ang) * baseR;
    const nx = -Math.sin(ang) * perp;
    const ny = Math.cos(ang) * perp;
    ctx.fillStyle = '#F8FAFC';
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 5;
    ctx.beginPath();
    ctx.moveTo(tipX, tipY);
    ctx.lineTo(bx + nx, by + ny);
    ctx.lineTo(bx - nx, by - ny);
    ctx.closePath();
    ctx.fill();
    ctx.shadowBlur = 0;

    // Centered %XX Readout strictly in clear interior space
    ctx.fillStyle = '#F8FAFC';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.font = `900 ${Math.round(52 * sc)}px B1Caption`;
    ctx.fillText(`%${Math.round(cur * 100)}`, gcx, gcy - 44 * sc);

    ctx.fillStyle = col;
    ctx.font = `800 ${Math.round(25 * sc)}px B1Caption`;
    ctx.fillText(String(data.label || (isSafe ? 'TAM SAVUNMA' : 'KRİTİK TEHLİKE')).toLocaleUpperCase('tr-TR'), gcx, gcy + 34 * sc, w - pad * 2);

    // Subtle indicator dots flanking label
    const ledAlpha = (!isSafe && Math.floor(t * 5) % 2 === 0) ? 0.85 : 0.25;
    ctx.fillStyle = col;
    ctx.globalAlpha *= ledAlpha;
    ctx.beginPath();
    ctx.arc(gcx - 190 * sc, gcy - 44 * sc, 5.5 * sc, 0, Math.PI * 2);
    ctx.arc(gcx + 190 * sc, gcy - 44 * sc, 5.5 * sc, 0, Math.PI * 2);
    ctx.fill();
    ctx.textAlign = 'left';
  } else if (kind === 'checklist') {
    const items = Array.isArray(data.items) && data.items.length
      ? data.items.map(String).slice(0, 3)
      : ['Linke değil resmi uygulamaya gir', 'Authenticator tabanlı 2FA aç', 'Her hesaba farklı parola kur'];

    ctx.textBaseline = 'top';
    ctx.fillStyle = '#34D399';
    ctx.font = `800 ${Math.round(22 * sc)}px B1Caption`;
    ctx.fillText('UYGULANACAK 3 ALTIN SAVUNMA KURALI', pad, pad);

    const startY = pad + 42 * sc;
    const rowH = (h - startY - pad) / Math.max(1, items.length);
    const stepSpan = Math.min(dur * 0.22, 0.45);
    const activeIdx = Math.min(items.length - 1, Math.floor(dt / Math.max(0.25, dur / Math.max(1, items.length))));

    items.forEach((item, idx) => {
      const appear = clamp((dt - idx * stepSpan) / 0.22);
      if (appear <= 0) return;
      const isSpotlight = idx === activeIdx;
      const ry = startY + idx * rowH + (1 - appear) * 16;
      ctx.save();
      ctx.globalAlpha *= appear;
      ctx.fillStyle = isSpotlight ? 'rgba(16, 185, 129, 0.14)' : 'rgba(16, 185, 129, 0.06)';
      ctx.strokeStyle = isSpotlight ? '#10B981' : 'rgba(16, 185, 129, 0.36)';
      ctx.lineWidth = isSpotlight ? 2.2 : 1.5;
      rrect(ctx, pad, ry, w - pad * 2, rowH - 12 * sc, 14);
      ctx.fill();
      ctx.stroke();

      const checkProg = clamp((dt - idx * stepSpan) / 0.35);
      drawShieldBadge(ctx, pad + 34 * sc, ry + (rowH - 12 * sc) / 2, 18 * sc, '#10B981', 'rgba(16, 185, 129, 0.16)', checkProg);
      ctx.fillStyle = '#F8FAFC';
      ctx.textBaseline = 'middle';
      ctx.font = `700 ${Math.round(25 * sc)}px B1Caption`;
      const lines = wrapTextSafe(ctx, `0${idx + 1}. ${item}`, w - pad * 2 - 84 * sc, 2);
      const midY = ry + (rowH - 12 * sc) / 2;
      if (lines.length === 1) {
        ctx.fillText(lines[0], pad + 72 * sc, midY);
      } else {
        ctx.fillText(lines[0], pad + 72 * sc, midY - 14 * sc);
        ctx.fillText(lines[1], pad + 72 * sc, midY + 14 * sc);
      }
      ctx.restore();
    });
  } else {
    // versus
    const leftTitle = String(data.leftTitle || 'RİSKLİ HATALAR').toLocaleUpperCase('tr-TR');
    const leftText = String(data.leftText || 'Panikle tıkla ve şifreni paylaş');
    const rightTitle = String(data.rightTitle || 'SİBER KALKAN').toLocaleUpperCase('tr-TR');
    const rightText = String(data.rightText || 'Dur, doğrula ve 2FA aç');

    const slideL = easeOutCubic(dt / 0.25);
    const slideR = easeOutCubic((dt - 0.10) / 0.25);
    const colW = (w - pad * 3.4) / 2;
    const cardH = h - pad * 2;

    // Left Coral Card
    ctx.save();
    ctx.translate(-(1 - slideL) * 28, 0);
    ctx.fillStyle = 'rgba(244, 63, 94, 0.11)';
    ctx.strokeStyle = 'rgba(244, 63, 94, 0.65)';
    ctx.lineWidth = 2;
    rrect(ctx, pad, pad, colW, cardH, 16);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = '#F43F5E';
    ctx.textBaseline = 'top';
    ctx.font = `900 ${Math.round(25 * sc)}px B1Caption`;
    ctx.fillText(`X  ${leftTitle}`, pad + 20 * sc, pad + 24 * sc, colW - 40 * sc);
    ctx.fillStyle = '#E2E8F0';
    ctx.font = `600 ${Math.round(23 * sc)}px B1Caption`;
    wrapTextSafe(ctx, leftText, colW - 40 * sc, 4).forEach((ln, i) => {
      ctx.fillText(ln, pad + 20 * sc, pad + 78 * sc + i * 34 * sc);
    });
    ctx.restore();

    // Right Emerald Card
    const rx = w - pad - colW;
    ctx.save();
    ctx.translate((1 - slideR) * 28, 0);
    ctx.fillStyle = 'rgba(16, 185, 129, 0.11)';
    ctx.strokeStyle = '#10B981';
    ctx.shadowColor = 'rgba(0, 0, 0, 0.35)';
    ctx.shadowBlur = 10;
    ctx.lineWidth = 2.2;
    rrect(ctx, rx, pad, colW, cardH, 16);
    ctx.fill();
    ctx.stroke();
    ctx.shadowBlur = 0;

    drawShieldBadge(ctx, rx + 36 * sc, pad + 38 * sc, 18 * sc, '#10B981', 'rgba(16, 185, 129, 0.18)');
    ctx.fillStyle = '#34D399';
    ctx.textBaseline = 'top';
    ctx.font = `900 ${Math.round(25 * sc)}px B1Caption`;
    ctx.fillText(rightTitle, rx + 66 * sc, pad + 24 * sc, colW - 84 * sc);
    ctx.fillStyle = '#E2E8F0';
    ctx.font = `600 ${Math.round(23 * sc)}px B1Caption`;
    wrapTextSafe(ctx, rightText, colW - 40 * sc, 4).forEach((ln, i) => {
      ctx.fillText(ln, rx + 20 * sc, pad + 78 * sc + i * 34 * sc);
    });
    ctx.restore();

    // Center VS Badge
    const vsPop = easeOutBack((dt - 0.14) / 0.22);
    if (vsPop > 0.05) {
      ctx.save();
      ctx.translate(w / 2, h / 2);
      ctx.scale(vsPop, vsPop);
      ctx.fillStyle = '#0F172A';
      ctx.strokeStyle = '#FBBF24';
      ctx.shadowColor = 'rgba(0, 0, 0, 0.5)';
      ctx.shadowBlur = 8;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.arc(0, 0, 26 * sc, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
      ctx.shadowBlur = 0;
      ctx.fillStyle = '#FBBF24';
      ctx.font = `900 ${Math.round(19 * sc)}px B1Caption`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('VS', 0, 1);
      ctx.restore();
    }
  }

  ctx.restore(); // pop interior clip

  // Outer crisp container border on top of clipped surface
  ctx.strokeStyle = borderCol;
  ctx.lineWidth = 1.6;
  rrect(ctx, 0, 0, w, h, 24);
  ctx.stroke();

  ctx.restore();
}

export function drawCyberCallout(ctx, l, dt) {
  const text = String(l.text || 'DİKKAT!').toLocaleUpperCase('tr-TR');
  const col = l.color || '#FBBF24';
  const size = l.size || 34;
  const pop = easeOutBack(dt / 0.16);
  const pulse = 1 + 0.015 * Math.sin(dt * 6);
  const scale = (1.30 - 0.30 * pop) * pulse;

  ctx.save();
  ctx.translate((l.w || 400) / 2, 0);
  ctx.rotate(l.angle || -0.015);
  ctx.scale(scale, scale);

  ctx.font = `900 ${size}px B1Caption`;
  const tw = Math.min(540, ctx.measureText(text).width);
  const bw = tw + 46;
  const bh = size + 26;

  ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
  ctx.shadowBlur = 12;
  ctx.shadowOffsetY = 4;
  ctx.fillStyle = 'rgba(22, 32, 51, 0.96)';
  ctx.strokeStyle = col;
  ctx.lineWidth = 2;
  rrect(ctx, -bw / 2, -bh / 2, bw, bh, 14);
  ctx.fill();
  ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.shadowOffsetY = 0;
  ctx.fillStyle = col;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(text, 0, 1, tw);
  ctx.restore();
}
