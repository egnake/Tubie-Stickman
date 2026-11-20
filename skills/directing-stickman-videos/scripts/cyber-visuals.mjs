// Procedural vector cybersecurity stickmen, hooded hacker, 8 live animated widgets,
// camera rig, and atmospheric transition engine for Tubie Studio (theme === 'cyber').

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
  hook: {primary: '#00F0FF', secondary: '#00FF88', glow: 'rgba(0, 240, 255, 0.14)', border: 'rgba(0, 240, 255, 0.55)'},
  danger: {primary: '#FF3B4E', secondary: '#FFE500', glow: 'rgba(255, 42, 68, 0.22)', border: 'rgba(255, 59, 78, 0.65)'},
  breach: {primary: '#FF3B30', secondary: '#FF9F0A', glow: 'rgba(255, 59, 48, 0.26)', border: 'rgba(255, 59, 48, 0.72)'},
  twist: {primary: '#E056FD', secondary: '#FFCC00', glow: 'rgba(224, 86, 253, 0.16)', border: 'rgba(224, 86, 253, 0.55)'},
  defense: {primary: '#00FF66', secondary: '#00F0FF', glow: 'rgba(0, 255, 102, 0.18)', border: 'rgba(0, 255, 102, 0.65)'},
  cta: {primary: '#00F0FF', secondary: '#00FF66', glow: 'rgba(0, 240, 255, 0.16)', border: 'rgba(0, 240, 255, 0.60)'},
};

export function validateCyberShot(shot) {
  if (shot.mood !== undefined && !CYBER_MOODS.includes(shot.mood)) {
    throw new Error(`Invalid cyber shot mood: ${shot.mood}`);
  }
  if (shot.camera !== undefined && !CYBER_CAMERAS.includes(shot.camera)) {
    throw new Error(`Invalid cyber shot camera: ${shot.camera}`);
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

function drawShieldBadge(ctx, cx, cy, r, strokeColor, fillColor) {
  ctx.save();
  ctx.beginPath();
  ctx.moveTo(cx, cy - r);
  ctx.lineTo(cx + r * 0.85, cy - r * 0.65);
  ctx.lineTo(cx + r * 0.75, cy + r * 0.25);
  ctx.quadraticCurveTo(cx, cy + r * 0.95, cx, cy + r * 1.05);
  ctx.quadraticCurveTo(cx, cy + r * 0.95, cx - r * 0.75, cy + r * 0.25);
  ctx.lineTo(cx - r * 0.85, cy - r * 0.65);
  ctx.closePath();
  ctx.fillStyle = fillColor || 'rgba(0, 255, 102, 0.16)';
  ctx.fill();
  ctx.strokeStyle = strokeColor || '#00FF66';
  ctx.lineWidth = Math.max(2.5, r * 0.14);
  ctx.stroke();

  ctx.beginPath();
  ctx.moveTo(cx - r * 0.34, cy + r * 0.02);
  ctx.lineTo(cx - r * 0.06, cy + r * 0.30);
  ctx.lineTo(cx + r * 0.38, cy - r * 0.28);
  ctx.strokeStyle = strokeColor || '#00FF66';
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.stroke();
  ctx.restore();
}

export function drawCyberAtmosphere(ctx, plane, t, shot, dt) {
  const mood = shot?.mood || 'hook';
  const pal = MOOD_PALETTES[mood] || MOOD_PALETTES.hook;
  const isPortrait = plane.height > 1500;
  const captionBand = isPortrait ? plane.height - 520 : plane.height - 110;

  // 1. Subtle technical grid
  ctx.save();
  ctx.beginPath();
  ctx.strokeStyle = mood === 'danger' || mood === 'breach'
    ? 'rgba(255, 65, 85, 0.065)'
    : mood === 'defense'
      ? 'rgba(0, 255, 128, 0.060)'
      : 'rgba(0, 240, 255, 0.055)';
  ctx.lineWidth = 0.9;
  for (let x = 32.5; x < plane.width; x += 64) {
    ctx.moveTo(x, 0);
    ctx.lineTo(x, plane.height);
  }
  for (let y = 16.5; y < plane.height; y += 64) {
    ctx.moveTo(0, y);
    ctx.lineTo(plane.width, y);
  }
  ctx.stroke();
  ctx.restore();

  // 2. Deterministic Binary Rain (low visual clutter, steady motion)
  ctx.save();
  ctx.font = '700 17px B1Caption';
  ctx.textBaseline = 'top';
  const cols = 16;
  const colW = plane.width / cols;
  for (let c = 0; c < cols; c++) {
    const speed = 130 + ((c * 43) % 105);
    const offset = (c * 317) % plane.height;
    const baseY = (t * speed + offset) % (plane.height + 280) - 140;
    for (let row = 0; row < 5; row++) {
      const ch = ((c * 3 + row + Math.floor(t * 5)) % 2 === 0) ? '1' : '0';
      const alpha = (0.075 - row * 0.012).toFixed(3);
      ctx.fillStyle = (mood === 'danger' || mood === 'breach')
        ? `rgba(255, 70, 90, ${alpha})`
        : `rgba(0, 255, 136, ${alpha})`;
      ctx.fillText(ch, c * colW + 26, baseY - row * 22);
    }
  }
  ctx.restore();

  // 3. Mood Ambient Radial Glow
  ctx.save();
  const aura = ctx.createRadialGradient(
    plane.width * 0.5,
    plane.height * 0.42,
    plane.width * 0.06,
    plane.width * 0.5,
    plane.height * 0.42,
    plane.width * 0.72,
  );
  aura.addColorStop(0, pal.glow);
  aura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = aura;
  ctx.fillRect(0, 0, plane.width, plane.height);
  ctx.restore();

  // 4. Radar Scanline Sweep
  ctx.save();
  const sweepSpan = Math.max(220, captionBand - 200);
  const scanY = (t * 195) % sweepSpan + 130;
  const scanGrad = ctx.createLinearGradient(0, scanY - 30, 0, scanY + 30);
  scanGrad.addColorStop(0, 'rgba(0, 240, 255, 0)');
  scanGrad.addColorStop(0.5, pal.glow);
  scanGrad.addColorStop(1, 'rgba(0, 240, 255, 0)');
  ctx.fillStyle = scanGrad;
  ctx.fillRect(36, scanY - 30, plane.width - 72, 60);
  ctx.restore();

  // 5. Holographic Cyber Deck Floor Horizon (grounds actors & workstations)
  ctx.save();
  const deckY = isPortrait ? 1485 : 558;
  const deckGlow = ctx.createLinearGradient(60, deckY, plane.width - 60, deckY);
  deckGlow.addColorStop(0, 'rgba(0, 0, 0, 0)');
  deckGlow.addColorStop(0.25, pal.border);
  deckGlow.addColorStop(0.5, pal.primary);
  deckGlow.addColorStop(0.75, pal.border);
  deckGlow.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.strokeStyle = deckGlow;
  ctx.lineWidth = 2.5;
  ctx.beginPath();
  ctx.moveTo(60, deckY);
  ctx.lineTo(plane.width - 60, deckY);
  ctx.stroke();

  const perspH = isPortrait ? 185 : 95;
  const perspGrad = ctx.createLinearGradient(0, deckY, 0, deckY + perspH);
  perspGrad.addColorStop(0, pal.glow);
  perspGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.strokeStyle = perspGrad;
  ctx.lineWidth = 1.2;
  for (let gx = 160; gx <= plane.width - 160; gx += 160) {
    ctx.beginPath();
    ctx.moveTo(gx, deckY);
    ctx.lineTo(gx + (gx - plane.width / 2) * 0.35, deckY + perspH);
    ctx.stroke();
  }
  ctx.beginPath();
  ctx.moveTo(105, deckY + perspH * 0.32);
  ctx.lineTo(plane.width - 105, deckY + perspH * 0.32);
  ctx.moveTo(85, deckY + perspH * 0.68);
  ctx.lineTo(plane.width - 85, deckY + perspH * 0.68);
  ctx.stroke();
  ctx.restore();

  // 6. Cinematic Vignette + Danger Pulse
  ctx.save();
  const vig = ctx.createRadialGradient(
    plane.width / 2,
    plane.height * 0.45,
    plane.width * 0.24,
    plane.width / 2,
    plane.height * 0.45,
    plane.width * 0.88,
  );
  vig.addColorStop(0, 'rgba(0, 0, 0, 0)');
  vig.addColorStop(1, 'rgba(4, 7, 14, 0.72)');
  ctx.fillStyle = vig;
  ctx.fillRect(0, 0, plane.width, plane.height);

  if (mood === 'danger' || mood === 'breach') {
    const pulse = 0.08 + 0.06 * Math.abs(Math.sin(t * 6.5 + dt));
    ctx.strokeStyle = `rgba(255, 42, 68, ${pulse.toFixed(3)})`;
    ctx.lineWidth = 22;
    ctx.strokeRect(11, 11, plane.width - 22, plane.height - 22);
  }
  ctx.restore();
}

export function drawCyberTopHud(ctx, plane, t, duration) {
  const isPortrait = plane.height > 1500;
  const bLen = isPortrait ? 42 : 28;
  const bPad = isPortrait ? 36 : 24;
  const hudFontSize = isPortrait ? 24 : 18;

  ctx.save();
  ctx.strokeStyle = 'rgba(0, 240, 255, 0.50)';
  ctx.lineWidth = isPortrait ? 3.5 : 2.5;
  ctx.beginPath(); ctx.moveTo(bPad, bPad + bLen); ctx.lineTo(bPad, bPad); ctx.lineTo(bPad + bLen, bPad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(plane.width - bPad - bLen, bPad); ctx.lineTo(plane.width - bPad, bPad); ctx.lineTo(plane.width - bPad, bPad + bLen); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(bPad, plane.height - bPad - bLen); ctx.lineTo(bPad, plane.height - bPad); ctx.lineTo(bPad + bLen, plane.height - bPad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(plane.width - bPad - bLen, plane.height - bPad); ctx.lineTo(plane.width - bPad, plane.height - bPad); ctx.lineTo(plane.width - bPad, plane.height - bPad - bLen); ctx.stroke();

  ctx.font = `700 ${hudFontSize}px B1Caption`;
  ctx.fillStyle = 'rgba(0, 240, 255, 0.88)';
  ctx.textBaseline = 'top';
  ctx.fillText('[ SİBER İSTİHBARAT ]', bPad + 12, bPad + 8);

  const recBlink = Math.floor(t * 2.2) % 2 === 0 ? 0.95 : 0.25;
  ctx.fillStyle = `rgba(255, 59, 48, ${recBlink})`;
  ctx.beginPath();
  ctx.arc(plane.width - bPad - (isPortrait ? 165 : 135), bPad + hudFontSize * 0.72, isPortrait ? 6 : 5, 0, Math.PI * 2);
  ctx.fill();

  ctx.fillStyle = 'rgba(255, 59, 48, 0.95)';
  ctx.fillText('REC', plane.width - bPad - (isPortrait ? 148 : 120), bPad + 8);
  ctx.fillStyle = 'rgba(255, 255, 255, 0.82)';
  ctx.fillText('4K UHD', plane.width - bPad - (isPortrait ? 95 : 74), bPad + 8);

  const barY = bPad + hudFontSize + 14;
  const barW = plane.width - (bPad + 12) * 2;
  const pct = clamp(t / Math.max(0.1, duration || 30));
  ctx.fillStyle = 'rgba(255, 255, 255, 0.08)';
  ctx.fillRect(bPad + 12, barY, barW, 3);
  const barGrad = ctx.createLinearGradient(bPad + 12, 0, bPad + 12 + barW, 0);
  barGrad.addColorStop(0, '#00F0FF');
  barGrad.addColorStop(1, '#00FF66');
  ctx.fillStyle = barGrad;
  ctx.fillRect(bPad + 12, barY, barW * pct, 3);

  if (isPortrait) {
    ctx.font = '600 13px B1Caption';
    ctx.fillStyle = 'rgba(0, 240, 255, 0.45)';
    ctx.fillText('NODE 0x7F // SSL_TLS · AES-256 MONITORING', bPad + 12, barY + 10);
    for (let eq = 0; eq < 5; eq++) {
      const eqH = 4 + Math.abs(Math.sin(t * (7 + eq * 2.3) + eq)) * 11;
      ctx.fillStyle = 'rgba(0, 255, 102, 0.65)';
      ctx.fillRect(plane.width - bPad - 44 + eq * 6, barY + 22 - eqH, 3.5, eqH);
    }
  }
  ctx.restore();
}

export function applyCyberCamera(ctx, plane, shot, dt) {
  const cam = shot?.camera || 'wide';
  const span = Math.max(0.5, (shot?.end ?? 1) - (shot?.start ?? 0));
  const prog = clamp(dt / span);
  let scale = 1.0;
  let dx = 0;
  let dy = 0;
  let rot = 0;

  if (cam === 'punch_in') {
    const p = easeOutCubic(dt / 0.24);
    scale = 1.0 + 0.055 * p + 0.012 * Math.sin(dt * 1.8);
  } else if (cam === 'zoom_slow') {
    scale = 1.0 + 0.058 * prog;
  } else if (cam === 'shake') {
    const decay = Math.exp(-dt * 4.5);
    dx = Math.sin(dt * 52) * 11 * decay;
    dy = Math.cos(dt * 63) * 8 * decay;
    scale = 1.02 + 0.04 * decay;
    rot = Math.sin(dt * 40) * 0.014 * decay;
  } else if (cam === 'glitch_cut') {
    const settle = 1 - easeOutCubic(dt / 0.18);
    scale = 1.0 + 0.075 * settle;
    dx = settle > 0.05 ? Math.sin(dt * 95) * 9 * settle : 0;
  } else {
    scale = 1.0 + 0.012 * Math.sin(dt * 2.1);
  }

  const cx = plane.width / 2;
  const cy = plane.height * 0.46;
  ctx.translate(cx + dx, cy + dy);
  ctx.rotate(rot);
  ctx.scale(scale, scale);
  ctx.translate(-cx, -cy);
}

export function drawCyberTransitionOverlay(ctx, plane, shot, dt) {
  if (dt < 0 || dt >= 0.16) return;
  const frac = 1 - dt / 0.16;
  const cam = shot?.camera || 'wide';
  const mood = shot?.mood || 'hook';
  ctx.save();
  if (cam === 'glitch_cut' || mood === 'breach' || mood === 'danger') {
    const bars = 6;
    for (let i = 0; i < bars; i++) {
      const y = ((i * 311 + Math.floor(dt * 200)) % Math.max(100, plane.height - 40));
      const h = 4 + ((i * 7) % 10);
      ctx.fillStyle = i % 2 === 0
        ? `rgba(0, 240, 255, ${(0.22 * frac).toFixed(3)})`
        : `rgba(255, 42, 68, ${(0.24 * frac).toFixed(3)})`;
      ctx.fillRect(0, y, plane.width, h);
    }
  } else {
    ctx.fillStyle = `rgba(0, 240, 255, ${(0.09 * frac).toFixed(3)})`;
    ctx.fillRect(0, 0, plane.width, plane.height);
  }
  ctx.restore();
}

export function drawCyberActor(ctx, l, dt, t) {
  const w = l.w || 360;
  const h = l.h || w * 1.18;
  const role = l.role || 'victim';
  const pose = l.pose || (role === 'hacker' ? 'hacker_lurk' : 'idle_browse');
  const isHacker = role === 'hacker';
  const isHero = role === 'hero';
  const cx = w * 0.5;
  const breath = Math.sin(t * 4.2 + dt * 1.5) * (h * 0.012);
  const strokeW = Math.max(4, w * 0.026);
  const headR = w * 0.175;
  const headY = h * 0.22 + breath;
  const neckY = headY + headR;
  const shoulderY = neckY + h * 0.05;
  const hipY = shoulderY + h * 0.27;
  const footY = h * 0.92;

  ctx.save();
  if (l.flip) {
    ctx.translate(w, 0);
    ctx.scale(-1, 1);
  }

  // Aura backdrop
  const auraColor = isHacker
    ? 'rgba(255, 42, 68, 0.28)'
    : isHero
      ? 'rgba(0, 255, 102, 0.24)'
      : 'rgba(0, 240, 255, 0.24)';
  const auraGrad = ctx.createRadialGradient(cx, h * 0.44, w * 0.06, cx, h * 0.44, w * 0.52);
  auraGrad.addColorStop(0, auraColor);
  auraGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = auraGrad;
  ctx.beginPath();
  ctx.arc(cx, h * 0.44, w * 0.52, 0, Math.PI * 2);
  ctx.fill();

  const bodyColor = isHacker ? '#FF3B4E' : isHero ? '#ECFFF4' : '#F0FAFF';
  const glowColor = isHacker ? '#FF2A44' : isHero ? '#00FF66' : '#00F0FF';

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
    const tap = pose === 'typing' ? Math.sin(t * 22) * (h * 0.022) : Math.sin(t * 5) * (h * 0.01);
    lElbow = [cx - w * 0.16, shoulderY + h * 0.12];
    rElbow = [cx + w * 0.19, shoulderY + h * 0.11];
    lHand = [cx + w * 0.04, shoulderY + h * 0.14 + tap];
    rHand = [cx + w * 0.16, shoulderY + h * 0.12 - tap];
  } else if (pose === 'shocked') {
    const tremor = Math.sin(t * 28) * (w * 0.016);
    bodyLean = tremor;
    lElbow = [cx - w * 0.28, headY + headR * 0.25];
    rElbow = [cx + w * 0.28, headY + headR * 0.25];
    lHand = [cx - headR * 0.96, headY - headR * 0.05];
    rHand = [cx + headR * 0.96, headY - headR * 0.05];
    lFoot = [cx - w * 0.19, footY];
    rFoot = [cx + w * 0.19, footY];
  } else if (pose === 'panic_run') {
    const stride = Math.sin(t * 15);
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
    const jab = pose === 'hacker_strike' ? Math.sin(t * 24) * (h * 0.03) : Math.sin(t * 5) * (h * 0.01);
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

  // Legs + Torso + Arms
  ctx.save();
  ctx.shadowColor = glowColor;
  ctx.shadowBlur = 14;
  ctx.strokeStyle = bodyColor;

  if (isHacker) {
    // Hooded dark hacker cloak behind torso
    ctx.save();
    ctx.fillStyle = '#0B0F19';
    ctx.beginPath();
    ctx.moveTo(headCx, headY - headR * 1.48);
    ctx.quadraticCurveTo(headCx + headR * 1.42, headY - headR * 0.4, shoulderCx + w * 0.30, hipY + h * 0.12);
    ctx.lineTo(shoulderCx - w * 0.30, hipY + h * 0.12);
    ctx.quadraticCurveTo(headCx - headR * 1.42, headY - headR * 0.4, headCx, headY - headR * 1.48);
    ctx.closePath();
    ctx.fill();
    ctx.strokeStyle = '#FF2A44';
    ctx.lineWidth = strokeW * 0.85;
    ctx.stroke();
    ctx.restore();
  }

  // Legs
  ctx.beginPath();
  ctx.moveTo(cx, hipY); ctx.lineTo(lKnee[0], lKnee[1]); ctx.lineTo(lFoot[0], lFoot[1]);
  ctx.moveTo(cx, hipY); ctx.lineTo(rKnee[0], rKnee[1]); ctx.lineTo(rFoot[0], rFoot[1]);
  ctx.stroke();

  // Spine & Neck
  ctx.beginPath();
  ctx.moveTo(headCx, headY + headR * 0.9);
  ctx.lineTo(shoulderCx, shoulderY);
  ctx.lineTo(cx, hipY);
  ctx.stroke();

  // Arms
  ctx.beginPath();
  ctx.moveTo(shoulderCx, shoulderY); ctx.lineTo(lElbow[0], lElbow[1]); ctx.lineTo(lHand[0], lHand[1]);
  ctx.moveTo(shoulderCx, shoulderY); ctx.lineTo(rElbow[0], rElbow[1]); ctx.lineTo(rHand[0], rHand[1]);
  ctx.stroke();

  // Head
  ctx.fillStyle = isHacker ? '#07090E' : '#0B1220';
  ctx.beginPath();
  ctx.arc(headCx, headY, headR, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Signature two hair tufts for victim/hero stickman
  if (!isHacker) {
    ctx.beginPath();
    ctx.moveTo(headCx - headR * 0.08, headY - headR);
    ctx.quadraticCurveTo(headCx + headR * 0.12, headY - headR * 1.38, headCx + headR * 0.36, headY - headR * 1.32);
    ctx.moveTo(headCx + headR * 0.04, headY - headR);
    ctx.quadraticCurveTo(headCx + headR * 0.28, headY - headR * 1.30, headCx + headR * 0.48, headY - headR * 1.15);
    ctx.stroke();
  }
  ctx.restore();

  // Facial Expressions
  const blink = (t % 3.4) > 3.24;
  ctx.save();
  if (isHacker) {
    // Glowing crimson narrow hacker eyes
    ctx.fillStyle = '#FF2A44';
    ctx.shadowColor = '#FF2A44';
    ctx.shadowBlur = 14;
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

    // Glitch sparks around hacker
    ctx.strokeStyle = 'rgba(255, 42, 68, 0.78)';
    ctx.lineWidth = 2.5;
    const sp = Math.sin(t * 18) * 6;
    ctx.beginPath();
    ctx.moveTo(headCx - headR * 1.25, headY - headR * 0.5 + sp);
    ctx.lineTo(headCx - headR * 1.05, headY - headR * 0.5 + sp);
    ctx.moveTo(headCx + headR * 1.05, headY + headR * 0.2 - sp);
    ctx.lineTo(headCx + headR * 1.30, headY + headR * 0.2 - sp);
    ctx.stroke();
  } else {
    const eyeY = headY - headR * 0.12;
    const eyeOff = headR * 0.34;
    ctx.fillStyle = '#FFFFFF';
    if (blink) {
      ctx.strokeStyle = '#FFFFFF';
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

    // Eyebrows & mouth by pose
    ctx.strokeStyle = '#FFFFFF';
    ctx.lineWidth = Math.max(2.5, strokeW * 0.36);
    if (pose === 'shocked' || pose === 'panic_run') {
      ctx.beginPath();
      ctx.moveTo(headCx - eyeOff - 10, eyeY - headR * 0.28); ctx.lineTo(headCx - eyeOff + 8, eyeY - headR * 0.38);
      ctx.moveTo(headCx + eyeOff + 10, eyeY - headR * 0.28); ctx.lineTo(headCx + eyeOff - 8, eyeY - headR * 0.38);
      ctx.stroke();

      ctx.fillStyle = '#0A0E17';
      ctx.beginPath();
      ctx.ellipse(headCx, headY + headR * 0.38, headR * 0.26, headR * 0.30, 0, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();

      // Panic sweat drops
      ctx.strokeStyle = '#00F0FF';
      ctx.lineWidth = 2.4;
      const dropY = Math.sin(t * 14) * 4;
      ctx.beginPath();
      ctx.ellipse(headCx - headR * 1.28, headY - headR * 0.25 + dropY, 6, 4, -0.4, 0, Math.PI * 2);
      ctx.ellipse(headCx + headR * 1.28, headY - headR * 0.25 - dropY, 6, 4, 0.4, 0, Math.PI * 2);
      ctx.stroke();
    } else if (pose === 'thinking') {
      ctx.beginPath();
      ctx.moveTo(headCx - eyeOff - 8, eyeY - headR * 0.26); ctx.lineTo(headCx - eyeOff + 8, eyeY - headR * 0.30);
      ctx.moveTo(headCx + eyeOff - 8, eyeY - headR * 0.34); ctx.quadraticCurveTo(headCx + eyeOff, eyeY - headR * 0.44, headCx + eyeOff + 8, eyeY - headR * 0.28);
      ctx.moveTo(headCx - headR * 0.22, headY + headR * 0.42);
      ctx.quadraticCurveTo(headCx, headY + headR * 0.35, headCx + headR * 0.24, headY + headR * 0.42);
      ctx.stroke();
    } else if (pose === 'celebrate' || pose === 'shield_hero') {
      ctx.beginPath();
      ctx.arc(headCx, headY + headR * 0.24, headR * 0.28, 0.15 * Math.PI, 0.85 * Math.PI);
      ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.moveTo(headCx - headR * 0.18, headY + headR * 0.38);
      ctx.quadraticCurveTo(headCx, headY + headR * 0.46, headCx + headR * 0.20, headY + headR * 0.36);
      ctx.stroke();
    }
  }
  ctx.restore();

  // Held accessories
  if (pose === 'idle_browse') {
    ctx.save();
    ctx.fillStyle = '#0F172A';
    ctx.strokeStyle = '#00F0FF';
    ctx.lineWidth = 2.5;
    ctx.shadowColor = '#00F0FF';
    ctx.shadowBlur = 12;
    rrect(ctx, rHand[0] - 8, rHand[1] - 26, 28, 44, 6);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = 'rgba(0, 240, 255, 0.45)';
    ctx.fillRect(rHand[0] - 4, rHand[1] - 20, 20, 28);
    ctx.restore();
  } else if (pose === 'shield_hero') {
    drawShieldBadge(ctx, rHand[0] + w * 0.04, rHand[1] - h * 0.02, w * 0.15, '#00FF66', 'rgba(0, 255, 102, 0.20)');
  }

  ctx.restore();
}

export function drawCyberWidget(ctx, l, dt, t) {
  const w = l.w || 1000;
  const h = l.h || 480;
  const kind = l.kind || 'terminal';
  const data = l.data || {};
  const sc = Math.min(w / 960, h / 440);

  ctx.save();
  // Glass Cyber Container Card
  const borderCol = (kind === 'checklist' || (kind === 'threat_meter' && data.safe))
    ? 'rgba(0, 255, 102, 0.65)'
    : (kind === 'notification' || kind === 'url_inspect' || kind === 'password_crack' || kind === 'threat_meter')
      ? 'rgba(255, 59, 78, 0.65)'
      : 'rgba(0, 240, 255, 0.60)';

  ctx.shadowColor = borderCol;
  ctx.shadowBlur = 22;
  ctx.fillStyle = 'rgba(8, 13, 24, 0.92)';
  rrect(ctx, 0, 0, w, h, 24);
  ctx.fill();
  ctx.shadowBlur = 0;
  ctx.strokeStyle = borderCol;
  ctx.lineWidth = 2.6;
  ctx.stroke();

  // Top neon accent hairline
  ctx.fillStyle = borderCol;
  rrect(ctx, 28, 0, w - 56, 4, 2);
  ctx.fill();

  const pad = Math.round(28 * Math.max(0.75, sc));

  if (kind === 'notification') {
    const shakeX = dt < 0.7 ? Math.sin(dt * 42) * 7 * (1 - dt / 0.7) : 0;
    ctx.translate(shakeX, 0);
    const badgeText = String(data.badge || 'ACİL GÜVENLİK BİLDİRİMİ').toLocaleUpperCase('tr-TR');
    ctx.font = `800 ${Math.round(20 * sc)}px B1Caption`;
    const bw = Math.min(w - pad * 2, ctx.measureText(badgeText).width + 48);
    ctx.fillStyle = 'rgba(255, 42, 68, 0.22)';
    ctx.strokeStyle = '#FF3B4E';
    ctx.lineWidth = 2;
    rrect(ctx, pad, pad, bw, 42 * sc, 10);
    ctx.fill();
    ctx.stroke();

    drawWarningTriangle(ctx, pad + 22 * sc, pad + 21 * sc, 11 * sc, '#FFE500');
    ctx.fillStyle = '#FFE500';
    ctx.textBaseline = 'middle';
    ctx.fillText(badgeText, pad + 42 * sc, pad + 21 * sc);

    // Live countdown pill top-right
    const remSec = Math.max(0, 899 - Math.floor(t * 2));
    const timerTxt = `${String(Math.floor(remSec / 60)).padStart(2, '0')}:${String(remSec % 60).padStart(2, '0')}`;
    ctx.textAlign = 'right';
    ctx.fillStyle = '#FF3B4E';
    ctx.font = `800 ${Math.round(24 * sc)}px B1Caption`;
    ctx.fillText(timerTxt, w - pad, pad + 21 * sc);
    ctx.textAlign = 'left';

    ctx.fillStyle = '#FFFFFF';
    ctx.textBaseline = 'top';
    ctx.font = `800 ${Math.round(32 * sc)}px B1Caption`;
    ctx.fillText(String(data.sender || 'Bilinmeyen Gönderici'), pad, pad + 64 * sc);

    ctx.font = `600 ${Math.round(27 * sc)}px B1Caption`;
    ctx.fillStyle = '#CBD5E1';
    const lines = wrapTextSafe(ctx, data.body || 'Hesabınızı hemen doğrulayın!', w - pad * 2, 3);
    lines.forEach((line, i) => ctx.fillText(line, pad, pad + 116 * sc + i * 38 * sc));

    // Action trap CTA bar inside notification
    const btnY = h - pad - 58 * sc;
    ctx.fillStyle = '#FF3B4E';
    rrect(ctx, pad, btnY, w - pad * 2, 54 * sc, 12);
    ctx.fill();
    ctx.fillStyle = '#FFFFFF';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.font = `800 ${Math.round(24 * sc)}px B1Caption`;
    ctx.fillText('HEMEN DOĞRULA  >>  TUZAK BAĞLANTI', w / 2, btnY + 27 * sc);
    ctx.textAlign = 'left';
  } else if (kind === 'terminal') {
    const headH = Math.round(48 * sc);
    ctx.fillStyle = 'rgba(20, 28, 46, 0.95)';
    rrect(ctx, 2, 2, w - 4, headH, 20);
    ctx.fill();

    for (const [idx, col] of [['#FF5F56', 0], ['#FFBD2E', 1], ['#27C93F', 2]]) {
      ctx.fillStyle = idx;
      ctx.beginPath();
      ctx.arc(pad + col * 24 * sc, headH / 2 + 2, 7.5 * sc, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.font = `700 ${Math.round(20 * sc)}px B1Caption`;
    ctx.fillStyle = '#94A3B8';
    ctx.textBaseline = 'middle';
    ctx.fillText(String(data.header || 'root@kali:~/exploit'), pad + 82 * sc, headH / 2 + 2);

    const lines = Array.isArray(data.lines) && data.lines.length
      ? data.lines.map(String)
      : ['$ scan --target', '[+] Session token captured', '[!] ACCESS GRANTED'];
    const totalChars = lines.reduce((acc, s) => acc + s.length, 0);
    let budget = Math.max(8, Math.floor(easeOutCubic(dt / 1.35) * totalChars));

    ctx.textBaseline = 'top';
    ctx.font = `700 ${Math.round(25 * sc)}px B1Caption`;
    let curY = headH + 26 * sc;
    for (let i = 0; i < lines.length; i++) {
      if (budget <= 0) break;
      const line = lines[i];
      const vis = line.slice(0, budget);
      budget -= vis.length;
      ctx.fillStyle = line.includes('[!]') || line.includes('>>>')
        ? '#FF3B4E'
        : line.includes('[+]')
          ? '#FFE500'
          : '#00FF66';
      const cursor = (budget <= 0 || i === lines.length - 1) && Math.floor(t * 4) % 2 === 0 ? ' _' : '';
      ctx.fillText(vis + cursor, pad, curY);
      curY += Math.round(44 * sc);
    }

    // Animated exploit progress bar at bottom
    const barW = w - pad * 2;
    const barY = h - pad - 30 * sc;
    const prog = easeOutCubic(dt / 1.45);
    ctx.fillStyle = 'rgba(255, 255, 255, 0.08)';
    rrect(ctx, pad, barY, barW, 20 * sc, 10);
    ctx.fill();
    ctx.fillStyle = prog > 0.7 ? '#FF3B4E' : '#00F0FF';
    rrect(ctx, pad, barY, Math.max(16, barW * prog), 20 * sc, 10);
    ctx.fill();
    ctx.font = `800 ${Math.round(18 * sc)}px B1Caption`;
    ctx.fillStyle = '#FFFFFF';
    ctx.textAlign = 'right';
    ctx.fillText(`${Math.round(prog * 100)}%`, w - pad, barY - 26 * sc);
    ctx.textAlign = 'left';
  } else if (kind === 'url_inspect') {
    const fakeUrl = String(data.fakeUrl || 'https://hesap-dogru1ama.com');
    const realUrl = String(data.realUrl || 'resmi-site.com');
    const stamp = String(data.stamp || 'HTTPS ≠ GÜVENLİ!').toLocaleUpperCase('tr-TR');

    ctx.font = `700 ${Math.round(20 * sc)}px B1Caption`;
    ctx.fillStyle = '#94A3B8';
    ctx.textBaseline = 'top';
    ctx.fillText('TARAYICI ADRES ÇUBUĞU ANALİZİ', pad, pad);

    // Fake URL bar
    const barY = pad + 44 * sc;
    const barH = 74 * sc;
    ctx.fillStyle = '#050810';
    ctx.strokeStyle = '#FF3B4E';
    ctx.lineWidth = 2.5;
    rrect(ctx, pad, barY, w - pad * 2, barH, 14);
    ctx.fill();
    ctx.stroke();

    // Small padlock icon + misleading green text + spoofed domain
    ctx.fillStyle = '#00FF66';
    ctx.font = `800 ${Math.round(20 * sc)}px B1Caption`;
    ctx.textBaseline = 'middle';
    ctx.fillText('LOCK', pad + 18 * sc, barY + barH / 2);

    ctx.fillStyle = '#FFFFFF';
    ctx.font = `800 ${Math.round(28 * sc)}px B1Caption`;
    ctx.fillText(fakeUrl, pad + 92 * sc, barY + barH / 2);

    // Moving magnifier ring scanning the spoofed domain
    const scanFrac = clamp(dt / 1.1);
    const magX = pad + 180 * sc + scanFrac * (w - pad * 2 - 260 * sc);
    ctx.save();
    ctx.strokeStyle = '#FFE500';
    ctx.lineWidth = 4;
    ctx.shadowColor = '#FFE500';
    ctx.shadowBlur = 14;
    ctx.beginPath();
    ctx.arc(magX, barY + barH / 2, 34 * sc, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();

    // Real domain comparison row
    ctx.textBaseline = 'top';
    ctx.font = `700 ${Math.round(24 * sc)}px B1Caption`;
    ctx.fillStyle = '#00FF66';
    ctx.fillText(`GERÇEK RESMİ ADRES:  ${realUrl}`, pad, barY + barH + 28 * sc);

    // Slam stamp after 0.35s
    if (dt >= 0.32) {
      const sProg = easeOutBack((dt - 0.32) / 0.22);
      ctx.save();
      ctx.translate(w * 0.5, h - pad - 54 * sc);
      ctx.rotate(-0.04);
      ctx.scale(1.3 - 0.3 * sProg, 1.3 - 0.3 * sProg);
      ctx.fillStyle = 'rgba(255, 42, 68, 0.22)';
      ctx.strokeStyle = '#FF3B4E';
      ctx.lineWidth = 3;
      const stW = Math.min(w - pad * 2, 540 * sc);
      rrect(ctx, -stW / 2, -32 * sc, stW, 64 * sc, 12);
      ctx.fill();
      ctx.stroke();
      ctx.fillStyle = '#FF3B4E';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.font = `900 ${Math.round(28 * sc)}px B1Caption`;
      ctx.fillText(stamp, 0, 0);
      ctx.restore();
    }
  } else if (kind === 'network_mitm') {
    const leftLbl = String(data.left || 'CİHAZIN');
    const midLbl = String(data.middle || 'SAHTE AĞ');
    const rightLbl = String(data.right || 'SUNUCU');
    const packets = Array.isArray(data.packets) && data.packets.length ? data.packets.map(String) : ['ŞİFRE', 'OTP', 'ÇEREZ'];

    const cy = h * 0.46;
    const x1 = w * 0.16;
    const x2 = w * 0.50;
    const x3 = w * 0.84;

    // Connection arcs
    ctx.strokeStyle = 'rgba(0, 240, 255, 0.42)';
    ctx.lineWidth = 4;
    ctx.beginPath();
    ctx.moveTo(x1, cy); ctx.lineTo(x2, cy);
    ctx.stroke();
    ctx.strokeStyle = 'rgba(255, 59, 78, 0.55)';
    ctx.beginPath();
    ctx.moveTo(x2, cy); ctx.lineTo(x3, cy);
    ctx.stroke();

    // Animated intercepted packets
    for (let i = 0; i < packets.length; i++) {
      const pFrac = (t * 0.55 + i * 0.33) % 1;
      const px = x1 + (x3 - x1) * pFrac;
      const intercepted = px >= x2 - 20;
      const py = cy + Math.sin(pFrac * Math.PI * 2) * (18 * sc);
      ctx.save();
      ctx.fillStyle = intercepted ? '#FF3B4E' : '#00F0FF';
      ctx.shadowColor = ctx.fillStyle;
      ctx.shadowBlur = 12;
      rrect(ctx, px - 42 * sc, py - 18 * sc, 84 * sc, 36 * sc, 8);
      ctx.fill();
      ctx.shadowBlur = 0;
      ctx.fillStyle = '#050810';
      ctx.font = `900 ${Math.round(16 * sc)}px B1Caption`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(packets[i], px, py);
      ctx.restore();
    }

    // Nodes
    const nodes = [
      [x1, leftLbl, '#00F0FF'],
      [x2, midLbl, '#FF3B4E'],
      [x3, rightLbl, '#00FF66'],
    ];
    for (const [nx, label, col] of nodes) {
      ctx.save();
      ctx.fillStyle = '#0B1220';
      ctx.strokeStyle = col;
      ctx.lineWidth = 4;
      ctx.shadowColor = col;
      ctx.shadowBlur = 16;
      ctx.beginPath();
      ctx.arc(nx, cy, 46 * sc, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
      ctx.shadowBlur = 0;
      ctx.fillStyle = '#FFFFFF';
      ctx.font = `800 ${Math.round(20 * sc)}px B1Caption`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillText(label, nx, cy + 62 * sc);
      ctx.restore();
    }
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.fillStyle = '#FFE500';
    ctx.font = `800 ${Math.round(22 * sc)}px B1Caption`;
    ctx.fillText('VERİ TRAFİĞİ ORTADAN YAKALANIYOR', w / 2, pad);
    ctx.textAlign = 'left';
  } else if (kind === 'password_crack') {
    const weak = String(data.weak || 'ahmet1998');
    const weakTime = String(data.weakTime || '0.4 SANİYE!');
    const strong = String(data.strong || 'K4p1!Mavi#92Z');
    const rate = String(data.rate || '4.82M DENEME/SN');
    const prog = clamp(dt / 1.1);

    ctx.textBaseline = 'top';
    ctx.font = `800 ${Math.round(22 * sc)}px B1Caption`;
    ctx.fillStyle = '#FF3B4E';
    ctx.fillText(`GPU KABA KUVVET HIZI: ${rate}`, pad, pad);

    // Weak Password Box
    const row1Y = pad + 48 * sc;
    ctx.fillStyle = 'rgba(255, 42, 68, 0.14)';
    ctx.strokeStyle = '#FF3B4E';
    ctx.lineWidth = 2.2;
    rrect(ctx, pad, row1Y, w - pad * 2, 118 * sc, 14);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = '#FFFFFF';
    ctx.font = `800 ${Math.round(32 * sc)}px B1Caption`;
    ctx.fillText(`ZAYIF: "${weak}"`, pad + 20 * sc, row1Y + 18 * sc);
    ctx.fillStyle = '#FFE500';
    ctx.font = `900 ${Math.round(26 * sc)}px B1Caption`;
    ctx.textAlign = 'right';
    ctx.fillText(prog >= 0.85 ? `KIRILDI: ${weakTime}` : `TARANIYOR %${Math.round(prog * 100)}`, w - pad - 20 * sc, row1Y + 22 * sc);
    ctx.textAlign = 'left';

    ctx.fillStyle = 'rgba(255, 255, 255, 0.1)';
    rrect(ctx, pad + 20 * sc, row1Y + 74 * sc, w - pad * 2 - 40 * sc, 20 * sc, 10);
    ctx.fill();
    ctx.fillStyle = '#FF3B4E';
    rrect(ctx, pad + 20 * sc, row1Y + 74 * sc, Math.max(18, (w - pad * 2 - 40 * sc) * prog), 20 * sc, 10);
    ctx.fill();

    // Strong Password Box
    const row2Y = row1Y + 142 * sc;
    ctx.fillStyle = 'rgba(0, 255, 102, 0.12)';
    ctx.strokeStyle = '#00FF66';
    rrect(ctx, pad, row2Y, w - pad * 2, 118 * sc, 14);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = '#00FF66';
    ctx.font = `800 ${Math.round(30 * sc)}px B1Caption`;
    ctx.fillText(`GÜÇLÜ (14+): "${strong}"`, pad + 20 * sc, row2Y + 22 * sc);
    ctx.fillStyle = '#CBD5E1';
    ctx.font = `700 ${Math.round(23 * sc)}px B1Caption`;
    ctx.fillText('KIRILMA SÜRESİ: YÜZYILLAR + 2FA KALKAN', pad + 20 * sc, row2Y + 68 * sc);
  } else if (kind === 'threat_meter') {
    const target = clamp((Number(data.level) || 96) / 100);
    const isSafe = Boolean(data.safe);
    const col = isSafe ? '#00FF66' : '#FF3B4E';
    const cur = target * easeOutCubic(dt / 0.85);
    const gcx = w / 2;
    const gcy = h * 0.68;
    const gr = Math.min(w * 0.34, h * 0.52);

    ctx.lineCap = 'round';
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.12)';
    ctx.lineWidth = 24 * sc;
    ctx.beginPath();
    ctx.arc(gcx, gcy, gr, Math.PI, Math.PI * 2);
    ctx.stroke();

    ctx.strokeStyle = col;
    ctx.shadowColor = col;
    ctx.shadowBlur = 18;
    ctx.beginPath();
    ctx.arc(gcx, gcy, gr, Math.PI, Math.PI + Math.PI * cur);
    ctx.stroke();
    ctx.shadowBlur = 0;

    // Needle
    const ang = Math.PI + Math.PI * cur;
    ctx.strokeStyle = '#FFFFFF';
    ctx.lineWidth = 5 * sc;
    ctx.beginPath();
    ctx.moveTo(gcx, gcy);
    ctx.lineTo(gcx + Math.cos(ang) * (gr - 18 * sc), gcy + Math.sin(ang) * (gr - 18 * sc));
    ctx.stroke();

    ctx.fillStyle = '#FFFFFF';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.font = `900 ${Math.round(54 * sc)}px B1Caption`;
    ctx.fillText(`%${Math.round(cur * 100)}`, gcx, gcy - 34 * sc);

    ctx.fillStyle = col;
    ctx.font = `800 ${Math.round(26 * sc)}px B1Caption`;
    ctx.fillText(String(data.label || (isSafe ? 'TAM SAVUNMA' : 'KRİTİK TEHLİKE')).toLocaleUpperCase('tr-TR'), gcx, gcy + 42 * sc);
    ctx.textAlign = 'left';
  } else if (kind === 'checklist') {
    const items = Array.isArray(data.items) && data.items.length
      ? data.items.map(String).slice(0, 3)
      : ['Linke değil resmi uygulamaya gir', 'Authenticator tabanlı 2FA aç', 'Her hesaba farklı parola kur'];

    ctx.textBaseline = 'top';
    ctx.fillStyle = '#00FF66';
    ctx.font = `800 ${Math.round(22 * sc)}px B1Caption`;
    ctx.fillText('UYGULANACAK 3 ALTIN SAVUNMA KURALI', pad, pad);

    const startY = pad + 44 * sc;
    const rowH = (h - startY - pad) / Math.max(1, items.length);
    items.forEach((item, idx) => {
      const appear = clamp((dt - idx * 0.22) / 0.22);
      if (appear <= 0) return;
      const ry = startY + idx * rowH + (1 - appear) * 18;
      ctx.save();
      ctx.globalAlpha *= appear;
      ctx.fillStyle = 'rgba(0, 255, 102, 0.10)';
      ctx.strokeStyle = 'rgba(0, 255, 102, 0.55)';
      ctx.lineWidth = 2;
      rrect(ctx, pad, ry, w - pad * 2, rowH - 14 * sc, 14);
      ctx.fill();
      ctx.stroke();

      drawShieldBadge(ctx, pad + 34 * sc, ry + (rowH - 14 * sc) / 2, 18 * sc, '#00FF66', 'rgba(0,255,102,0.18)');
      ctx.fillStyle = '#FFFFFF';
      ctx.textBaseline = 'middle';
      ctx.font = `700 ${Math.round(26 * sc)}px B1Caption`;
      const lines = wrapTextSafe(ctx, `0${idx + 1}. ${item}`, w - pad * 2 - 84 * sc, 2);
      const midY = ry + (rowH - 14 * sc) / 2;
      if (lines.length === 1) {
        ctx.fillText(lines[0], pad + 72 * sc, midY);
      } else {
        ctx.fillText(lines[0], pad + 72 * sc, midY - 15 * sc);
        ctx.fillText(lines[1], pad + 72 * sc, midY + 15 * sc);
      }
      ctx.restore();
    });
  } else {
    // versus
    const leftTitle = String(data.leftTitle || 'RİSKLİ HATALAR').toLocaleUpperCase('tr-TR');
    const leftText = String(data.leftText || 'Panikle tıkla ve şifreni paylaş');
    const rightTitle = String(data.rightTitle || 'SİBER KALKAN').toLocaleUpperCase('tr-TR');
    const rightText = String(data.rightText || 'Dur, doğrula ve 2FA aç');

    const colW = (w - pad * 3) / 2;
    const cardH = h - pad * 2;
    // Left Red Card
    ctx.fillStyle = 'rgba(255, 42, 68, 0.14)';
    ctx.strokeStyle = '#FF3B4E';
    ctx.lineWidth = 2.4;
    rrect(ctx, pad, pad, colW, cardH, 16);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = '#FF3B4E';
    ctx.textBaseline = 'top';
    ctx.font = `900 ${Math.round(26 * sc)}px B1Caption`;
    ctx.fillText(`X  ${leftTitle}`, pad + 20 * sc, pad + 26 * sc);
    ctx.fillStyle = '#E2E8F0';
    ctx.font = `600 ${Math.round(24 * sc)}px B1Caption`;
    wrapTextSafe(ctx, leftText, colW - 40 * sc, 4).forEach((ln, i) => {
      ctx.fillText(ln, pad + 20 * sc, pad + 82 * sc + i * 36 * sc);
    });

    // Right Green Card
    const rx = pad * 2 + colW;
    ctx.fillStyle = 'rgba(0, 255, 102, 0.13)';
    ctx.strokeStyle = '#00FF66';
    rrect(ctx, rx, pad, colW, cardH, 16);
    ctx.fill();
    ctx.stroke();

    drawShieldBadge(ctx, rx + 36 * sc, pad + 40 * sc, 18 * sc, '#00FF66', 'rgba(0,255,102,0.2)');
    ctx.fillStyle = '#00FF66';
    ctx.font = `900 ${Math.round(26 * sc)}px B1Caption`;
    ctx.fillText(rightTitle, rx + 66 * sc, pad + 26 * sc);
    ctx.fillStyle = '#E2E8F0';
    ctx.font = `600 ${Math.round(24 * sc)}px B1Caption`;
    wrapTextSafe(ctx, rightText, colW - 40 * sc, 4).forEach((ln, i) => {
      ctx.fillText(ln, rx + 20 * sc, pad + 82 * sc + i * 36 * sc);
    });
  }

  ctx.restore();
}

export function drawCyberCallout(ctx, l, dt) {
  const text = String(l.text || 'DİKKAT!').toLocaleUpperCase('tr-TR');
  const col = l.color || '#FFE500';
  const size = l.size || 38;
  const pop = easeOutBack(dt / 0.16);
  const scale = 1.42 - 0.42 * pop;

  ctx.save();
  ctx.translate((l.w || 400) / 2, 0);
  ctx.rotate(l.angle || -0.04);
  ctx.scale(scale, scale);

  ctx.font = `900 ${size}px B1Caption`;
  const tw = ctx.measureText(text).width;
  const bw = tw + 48;
  const bh = size + 28;

  ctx.shadowColor = col;
  ctx.shadowBlur = 18;
  ctx.fillStyle = 'rgba(7, 10, 18, 0.94)';
  ctx.strokeStyle = col;
  ctx.lineWidth = 3.5;
  rrect(ctx, -bw / 2, -bh / 2, bw, bh, 14);
  ctx.fill();
  ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.fillStyle = col;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(text, 0, 1);
  ctx.restore();
}
