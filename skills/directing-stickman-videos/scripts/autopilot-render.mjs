import fs from 'node:fs';
import path from 'node:path';
import {spawn} from 'node:child_process';
import {once} from 'node:events';
import {createCanvas, loadImage, GlobalFonts} from '@napi-rs/canvas';
import {ROOT, FFMPEG, sha, hashFile, save, probe, run, mediaPath, fontPaths, outputDir, designPlane, loadProject, concatFileLine} from './autopilot-core.mjs';

import {walkLayers, drawB3Shape, televisionFinish} from './b3-graphics.mjs';
import {appendB4Audio} from './b4-runtime.mjs';
import {checkOpeningFile} from './b4-opening-audio-check.mjs';

const clamp = x => Math.max(0, Math.min(1, x));
const ease = x => { x = clamp(x); return x * x * (3 - 2 * x); };
export function keyState(l, t) {
  const out = {...l};
  for (const prop of ['x', 'y', 'w', 'angle', 'alpha', 'progress', 'scaleX', 'scaleY']) {
    const keys = (l.keys || []).filter(k => k[prop] !== undefined);
    if (!keys.length) continue;
    if (t <= keys[0].t) out[prop] = keys[0][prop];
    else if (t >= keys.at(-1).t) out[prop] = keys.at(-1)[prop];
    else { const i = keys.findIndex(k => k.t > t), a = keys[i - 1], b = keys[i]; out[prop] = a[prop] + (b[prop] - a[prop]) * ease((t - a.t) / (b.t - a.t)); }
  }
  return out;
}
function wrap(ctx, text, maxWidth) {
  const parts = /\s/.test(text) ? text.split(/\s+/) : [...text];
  const separator = /\s/.test(text) ? ' ' : '';
  const lines = []; let line = '';
  for (const part of parts) {
    const next = line ? line + separator + part : part;
    if (ctx.measureText(next).width > maxWidth && line) { lines.push(line); line = part; }
    else line = next;
    if (ctx.measureText(line).width > maxWidth) throw new Error('Text token too wide; shorten it or reduce size');
  }
  if (line) lines.push(line);
  return lines;
}
export async function painter(p) {
  const fonts = fontPaths(p);
  const plane = designPlane(p), captionBand = plane.height - 110;
  if (!GlobalFonts.registerFromPath(fonts.hand, 'B1Hand') || !GlobalFonts.registerFromPath(fonts.caption, 'B1Caption')) throw new Error('Font registration failed');
  const canvas = createCanvas(p.width, p.height), ctx = canvas.getContext('2d');
  const sceneCopy = p.style === 'B3' ? createCanvas(p.width, p.height) : null;
  const images = new Map();
  for (const s of p.shots) for (const l of walkLayers(s.layers)) if (['image', 'pose'].includes(l.type)) {
    const poseSheet = (p.theme === 'cyber' || l.sheet === 'cyber') ? '@b1/character-pose-sheet-cyber.png' : '@b1/character-pose-sheet-v6.png';
    const f = mediaPath(p.base, l.type === 'pose' ? poseSheet : l.path);
    if (!images.has(f)) images.set(f, await loadImage(f));
  }
  function stroke(points, width, color, progress = 1) {
    const lengths = points.slice(1).map((pt, i) => Math.hypot(pt[0] - points[i][0], pt[1] - points[i][1]));
    let left = lengths.reduce((a, b) => a + b, 0) * clamp(progress);
    ctx.strokeStyle = color; ctx.lineWidth = width; ctx.lineCap = 'round'; ctx.lineJoin = 'round'; ctx.beginPath(); ctx.moveTo(...points[0]);
    for (let i = 0; i < lengths.length && left > 0; i++) { const f = Math.min(1, left / Math.max(0.001, lengths[i])); ctx.lineTo(points[i][0] + (points[i + 1][0] - points[i][0]) * f, points[i][1] + (points[i + 1][1] - points[i][1]) * f); left -= lengths[i]; }
    ctx.stroke();
  }
  return function frame(t) {
    ctx.resetTransform(); ctx.clearRect(0, 0, p.width, p.height); ctx.scale(p.width / plane.width, p.width / plane.width);
    const isCyber = p.theme === 'cyber';
    ctx.fillStyle = p.backgroundColor || (isCyber ? '#0A0E17' : (['B3', 'B4'].includes(p.style) ? '#111316' : p.style === 'B2' ? '#f4f0e5' : '#FDFDFC'));
    ctx.fillRect(0, 0, plane.width, plane.height);
    if ((p.style === 'B1' || isCyber) && p.grid !== false) {
      ctx.beginPath();
      ctx.strokeStyle = isCyber ? 'rgba(0, 255, 128, 0.055)' : 'rgba(76,90,105,.036)';
      ctx.lineWidth = 0.8;
      for (let x = 32.5; x < plane.width; x += 64) { ctx.moveTo(x, 0); ctx.lineTo(x, captionBand - 1); }
      for (let y = 16.5; y < captionBand; y += 64) { ctx.moveTo(0, y); ctx.lineTo(plane.width, y); }
      ctx.stroke();
    }
    if (isCyber) {
      // 1. Deterministic Animated Cyber Binary Rain (subtle motion & depth)
      ctx.save();
      ctx.font = '700 18px B1Caption';
      ctx.textBaseline = 'top';
      const cols = 16;
      const colW = plane.width / cols;
      for (let c = 0; c < cols; c++) {
        const speed = 135 + ((c * 43) % 110);
        const offset = (c * 317) % plane.height;
        const baseY = (t * speed + offset) % (plane.height + 320) - 160;
        for (let row = 0; row < 6; row++) {
          const ch = ((c * 3 + row + Math.floor(t * 5)) % 2 === 0) ? '1' : '0';
          ctx.fillStyle = `rgba(0, 255, 128, ${(0.085 - row * 0.012).toFixed(3)})`;
          ctx.fillText(ch, c * colW + 26, baseY - row * 24);
        }
      }
      ctx.restore();

      // 1b. Dynamic Cyber Radar Scanline Sweep
      ctx.save();
      const scanY = (t * 210) % (plane.height - 520) + 220;
      const scanGrad = ctx.createLinearGradient(0, scanY - 36, 0, scanY + 36);
      scanGrad.addColorStop(0, 'rgba(0, 240, 255, 0)');
      scanGrad.addColorStop(0.5, 'rgba(0, 240, 255, 0.055)');
      scanGrad.addColorStop(1, 'rgba(0, 240, 255, 0)');
      ctx.fillStyle = scanGrad;
      ctx.fillRect(40, scanY - 36, plane.width - 80, 72);
      ctx.restore();

      // 1c. Holographic Cyber Deck Floor Horizon (grounds actors & workstations)
      ctx.save();
      const deckY = 1485;
      const deckGlow = ctx.createLinearGradient(60, deckY, plane.width - 60, deckY);
      deckGlow.addColorStop(0, 'rgba(0, 240, 255, 0)');
      deckGlow.addColorStop(0.25, 'rgba(0, 240, 255, 0.28)');
      deckGlow.addColorStop(0.5, 'rgba(0, 255, 128, 0.42)');
      deckGlow.addColorStop(0.75, 'rgba(0, 240, 255, 0.28)');
      deckGlow.addColorStop(1, 'rgba(0, 240, 255, 0)');
      ctx.strokeStyle = deckGlow;
      ctx.lineWidth = 2.5;
      ctx.beginPath();
      ctx.moveTo(60, deckY);
      ctx.lineTo(plane.width - 60, deckY);
      ctx.stroke();

      // Perspective deck grid lines fading below the stage
      const perspGrad = ctx.createLinearGradient(0, deckY, 0, deckY + 185);
      perspGrad.addColorStop(0, 'rgba(0, 240, 255, 0.14)');
      perspGrad.addColorStop(1, 'rgba(0, 240, 255, 0)');
      ctx.strokeStyle = perspGrad;
      ctx.lineWidth = 1.2;
      for (let gx = 160; gx <= plane.width - 160; gx += 160) {
        ctx.beginPath();
        ctx.moveTo(gx, deckY);
        ctx.lineTo(gx + (gx - plane.width / 2) * 0.35, deckY + 185);
        ctx.stroke();
      }
      ctx.beginPath();
      ctx.moveTo(110, deckY + 48);
      ctx.lineTo(plane.width - 110, deckY + 48);
      ctx.moveTo(85, deckY + 108);
      ctx.lineTo(plane.width - 85, deckY + 108);
      ctx.stroke();
      ctx.restore();

      // 2. Cinematic Radial Cyber Vignette
      const vig = ctx.createRadialGradient(plane.width / 2, plane.height * 0.45, plane.width * 0.25, plane.width / 2, plane.height * 0.45, plane.width * 0.85);
      vig.addColorStop(0, 'rgba(0,0,0,0)');
      vig.addColorStop(1, 'rgba(4, 8, 16, 0.72)');
      ctx.fillStyle = vig;
      ctx.fillRect(0, 0, plane.width, plane.height);

      // 3. High-Tech Cyber HUD Corner Brackets
      ctx.save();
      ctx.strokeStyle = 'rgba(0, 240, 255, 0.50)';
      ctx.lineWidth = 3.5;
      const bLen = 42, bPad = 36;
      // Top-Left
      ctx.beginPath(); ctx.moveTo(bPad, bPad + bLen); ctx.lineTo(bPad, bPad); ctx.lineTo(bPad + bLen, bPad); ctx.stroke();
      // Top-Right
      ctx.beginPath(); ctx.moveTo(plane.width - bPad - bLen, bPad); ctx.lineTo(plane.width - bPad, bPad); ctx.lineTo(plane.width - bPad, bPad + bLen); ctx.stroke();
      // Bottom-Left
      ctx.beginPath(); ctx.moveTo(bPad, plane.height - bPad - bLen); ctx.lineTo(bPad, plane.height - bPad); ctx.lineTo(bPad + bLen, plane.height - bPad); ctx.stroke();
      // Bottom-Right
      ctx.beginPath(); ctx.moveTo(plane.width - bPad - bLen, plane.height - bPad); ctx.lineTo(plane.width - bPad, plane.height - bPad); ctx.lineTo(plane.width - bPad, plane.height - bPad - bLen); ctx.stroke();

      // 4. Top Cyber Broadcast HUD (REC 4K UHD, Threat Intel & Progress Bar)
      ctx.font = '700 24px B1Caption';
      ctx.fillStyle = 'rgba(0, 240, 255, 0.85)';
      ctx.textBaseline = 'top';
      ctx.fillText('[ SİBER İSTİHBARAT ]', bPad + 12, bPad + 10);

      const recBlink = Math.floor(t * 2) % 2 === 0 ? 0.95 : 0.30;
      ctx.fillStyle = `rgba(255, 59, 48, ${recBlink})`;
      ctx.beginPath();
      ctx.arc(plane.width - bPad - 165, bPad + 22, 6, 0, Math.PI * 2);
      ctx.fill();

      ctx.fillStyle = 'rgba(255, 59, 48, 0.95)';
      ctx.fillText('REC', plane.width - bPad - 148, bPad + 10);

      ctx.fillStyle = 'rgba(255, 255, 255, 0.80)';
      ctx.fillText('4K UHD', plane.width - bPad - 95, bPad + 10);

      // Subtle retention progress bar under top HUD
      const barW = plane.width - (bPad + 12) * 2;
      const pct = clamp(t / Math.max(0.1, p.duration || 30));
      ctx.fillStyle = 'rgba(255, 255, 255, 0.08)';
      ctx.fillRect(bPad + 12, bPad + 44, barW, 3);
      const barGrad = ctx.createLinearGradient(bPad + 12, 0, bPad + 12 + barW, 0);
      barGrad.addColorStop(0, '#00F0FF');
      barGrad.addColorStop(1, '#00FF66');
      ctx.fillStyle = barGrad;
      ctx.fillRect(bPad + 12, bPad + 44, barW * pct, 3);

      // Micro Telemetry Bar & Equalizer Visualizer
      ctx.font = '600 13px B1Caption';
      ctx.fillStyle = 'rgba(0, 240, 255, 0.42)';
      ctx.fillText('NODE 0x7F // SSL_TLS · AES-256 MONITORING', bPad + 12, bPad + 54);

      for (let eq = 0; eq < 5; eq++) {
        const eqH = 4 + Math.abs(Math.sin(t * (7 + eq * 2.3) + eq)) * 11;
        ctx.fillStyle = 'rgba(0, 255, 102, 0.65)';
        ctx.fillRect(plane.width - bPad - 44 + eq * 6, bPad + 66 - eqH, 3.5, eqH);
      }
      ctx.restore();
    }
    const shot = p.shots.find(s => t >= s.start && t < s.end) || p.shots.at(-1);
    function paintLayers(layers) { for (const layer of layers) {
      const dt = t - shot.start;
      if (dt < (layer.from || 0) || dt >= (layer.to ?? Infinity)) continue;
      const l = keyState(layer, dt), x = l.x || 0, y = l.y || 0, w = l.w || 200;
      ctx.save(); ctx.globalAlpha *= clamp(l.alpha ?? 1); ctx.translate(x, y);
      const px = l.pivotX || 0, py = l.pivotY || 0;
      ctx.translate(px, py); ctx.rotate(l.angle || 0); ctx.scale(l.scaleX ?? 1, l.scaleY ?? 1); ctx.translate(-px, -py);
      if (l.type === 'group') paintLayers(l.children);
      else if (['path', 'rect', 'ellipse'].includes(l.type)) drawB3Shape(ctx, l);
      else if (l.type === 'image') {
        const img = images.get(mediaPath(p.base, l.path)), h = l.h || w * img.height / img.width;
        const scale = Math.min(w / img.width, h / img.height);
        if (isCyber && l.glow) {
          ctx.save();
          ctx.shadowColor = l.glow === 'red' ? 'rgba(255, 59, 48, 0.70)' : (l.glow === 'green' ? 'rgba(0, 255, 102, 0.65)' : (l.glow === 'yellow' ? 'rgba(255, 204, 0, 0.70)' : 'rgba(0, 240, 255, 0.65)'));
          ctx.shadowBlur = 24;
          ctx.drawImage(img, (w - img.width * scale) / 2, (h - img.height * scale) / 2, img.width * scale, img.height * scale);
          ctx.restore();
        } else {
          ctx.drawImage(img, (w - img.width * scale) / 2, (h - img.height * scale) / 2, img.width * scale, img.height * scale);
        }
      } else if (l.type === 'pose') {
        const poseSheet = (isCyber || l.sheet === 'cyber') ? '@b1/character-pose-sheet-cyber.png' : '@b1/character-pose-sheet-v6.png';
        const img = images.get(mediaPath(p.base, poseSheet)), sw = img.width / 3;
        if (isCyber && l.aura) {
          ctx.save();
          const cx = w / 2, cy = (w * img.height / sw) * 0.45, r = w * 0.48;
          const auraGrad = ctx.createRadialGradient(cx, cy, r * 0.15, cx, cy, r);
          auraGrad.addColorStop(0, l.aura === 'red' ? 'rgba(255, 59, 48, 0.24)' : (l.aura === 'green' ? 'rgba(0, 255, 102, 0.22)' : 'rgba(0, 240, 255, 0.22)'));
          auraGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
          ctx.fillStyle = auraGrad;
          ctx.beginPath();
          ctx.arc(cx, cy, r, 0, Math.PI * 2);
          ctx.fill();
          ctx.restore();
        }
        ctx.drawImage(img, l.pose * sw, 0, sw, img.height, 0, 0, w, w * img.height / sw);
      } else if (l.type === 'shadow') {
        const gradient = ctx.createLinearGradient(0, 0, w, 0);
        gradient.addColorStop(0, 'rgba(10,12,14,.96)');
        gradient.addColorStop(.55, 'rgba(10,12,14,.8)');
        gradient.addColorStop(1, 'rgba(10,12,14,0)');
        ctx.fillStyle = gradient; ctx.fillRect(0, 0, w, l.h);
      } else if (l.type === 'text') {
        const size = l.size || 52;
        const fontName = (isCyber || l.font === 'caption' || ['B3', 'B4'].includes(p.style)) ? 'B1Caption' : 'B1Hand';
        ctx.font = `${l.bold ? 700 : 400} ${size}px ${fontName}`;
        ctx.textBaseline = 'top';
        const lines = wrap(ctx, l.text, l.w || 1080 - x);
        if (isCyber && l.badge) {
          const maxLineW = Math.max(...lines.map(s => ctx.measureText(s).width));
          const padX = 28, padY = 12;
          const bw = maxLineW + padX * 2;
          const bh = lines.length * size * 1.15 + padY * 2 - 4;
          const offsetX = l.center ? -bw / 2 : -padX;
          ctx.save();
          ctx.fillStyle = 'rgba(8, 16, 28, 0.90)';
          ctx.strokeStyle = l.badge;
          ctx.lineWidth = 2.5;
          ctx.shadowColor = l.badge;
          ctx.shadowBlur = 16;
          ctx.beginPath();
          if (typeof ctx.roundRect === 'function') ctx.roundRect(offsetX, -padY, bw, bh, 14);
          else ctx.rect(offsetX, -padY, bw, bh);
          ctx.fill();
          ctx.stroke();
          ctx.restore();
        }
        ctx.fillStyle = l.color || (isCyber ? '#E6EDF3' : (['B3', 'B4'].includes(p.style) ? '#e9eaea' : '#191916'));
        if (isCyber && l.center) {
          ctx.textAlign = 'center';
          lines.forEach((s, i) => ctx.fillText(s, 0, i * size * 1.15));
          ctx.textAlign = 'left';
        } else {
          lines.forEach((s, i) => ctx.fillText(s, 0, i * size * 1.15));
        }
      } else {
        let points = l.points;
        if (l.type === 'underline') points = [[0, 1], [w * 0.25, -1], [w * 0.55, 2], [w, -2]];
        if (l.type === 'arrow') points = Array.from({length: 30}, (_, i) => [w * i / 29, -26 * Math.sin(i / 29 * Math.PI)]);
        const defaultColor = isCyber ? '#00FF66' : (p.style === 'B3' ? '#d0d2d4' : '#191916');
        stroke(points, l.lineWidth || (l.type === 'underline' ? 6 : 3), l.color || defaultColor, l.progress ?? 1);
        if (l.type === 'arrow' && (l.progress ?? 1) > 0.95) stroke([[w - 17, -12], [w, 0], [w - 19, 12]], 3, l.color || defaultColor);
      }
      ctx.restore();
    } }
    paintLayers(shot.layers);
    if (sceneCopy && p.television?.enabled !== false) {
      sceneCopy.getContext('2d').drawImage(canvas, 0, 0);
      televisionFinish(ctx, sceneCopy, p.width, p.height, p.television?.intensity ?? 0.45);
    }
    const cue = p.captions.find(c => t >= c.start && t < c.end);
    if (cue) {
      const isModern = p.captionStyle === 'modern' || p.captionStyle === 'viral' || (plane.height > 1500 && p.captionStyle !== 'classic');
      const fontSize = p.captionSize || (isModern ? 56 : (p.style === 'B2' ? 40 : 34));
      ctx.font = `${isModern ? 800 : (p.style === 'B2' ? 500 : 600)} ${fontSize}px B1Caption`;
      ctx.textBaseline = 'middle';
      if (/[\r\n]/.test(cue.text)) throw new Error('B captions allow at most one line; split the timed cue at a semantic boundary');
      const textToDisplay = (isModern && p.captionUppercase !== false) ? cue.text.toUpperCase() : (p.style === 'B3' ? cue.text.toUpperCase() : cue.text);
      const lines = wrap(ctx, textToDisplay, 1120);
      if (lines.length > 1) throw new Error('B captions allow at most one line; split the timed cue at a semantic boundary');
      const cy = p.captionY !== undefined ? p.captionY : (isModern ? Math.round(plane.height * 0.72) : plane.height - 60);
      if (isModern && p.style !== 'B2' && p.style !== 'B3' && p.style !== 'B4') {
        const textWidth = ctx.measureText(textToDisplay).width;
        const boxW = Math.min(1160, textWidth + 64);
        const boxH = Math.max(84, fontSize * 1.6);
        ctx.save();
        ctx.shadowColor = isCyber ? 'rgba(0, 255, 128, 0.40)' : 'rgba(0,0,0,0.50)';
        ctx.shadowBlur = isCyber ? 22 : 16;
        ctx.shadowOffsetY = 6;
        ctx.fillStyle = p.captionBg || (isCyber ? 'rgba(7, 12, 20, 0.96)' : 'rgba(14, 17, 23, 0.94)');
        ctx.beginPath();
        if (typeof ctx.roundRect === 'function') ctx.roundRect(640 - boxW / 2, cy - boxH / 2, boxW, boxH, 20);
        else ctx.rect(640 - boxW / 2, cy - boxH / 2, boxW, boxH);
        ctx.fill();
        ctx.strokeStyle = p.captionBorder || (isCyber ? 'rgba(0, 255, 128, 0.50)' : 'rgba(255, 255, 255, 0.18)');
        ctx.lineWidth = 2.5;
        ctx.stroke();
        ctx.restore();

        const words = (cue.words && cue.words.length) ? cue.words : null;
        if (words && words.length > 0) {
          ctx.textAlign = 'left';
          const spaceW = ctx.measureText(' ').width;
          let curX = 640 - textWidth / 2;
          for (const wObj of words) {
            const wText = (p.captionUppercase !== false) ? wObj.word.toUpperCase() : wObj.word;
            const wWidth = ctx.measureText(wText).width;
            const isActive = t >= wObj.start && t < wObj.end;
            const drawY = isActive && isCyber ? cy - 3 : cy;
            ctx.strokeStyle = '#000000';
            ctx.lineWidth = 6;
            ctx.strokeText(wText, curX, drawY);
            if (isActive) {
              let activeColor = p.captionActiveColor || '#FFE500';
              if (isCyber) {
                if (/(HACKER|ŞİFRE|ÇALIN|SAHTE|TUZA[ĞG]|TEHLİKE|SIZDIR|KAYIP|PANİK|SIFIRLAN|SALDIR|KRİPT|KIRMAK|AĞDAN|HEDİYE|ATLAR|KÖTÜ|BOŞALT|VİRÜS)/.test(wText)) {
                  activeColor = '#FF3B4E';
                } else if (/(2FA|KORU|GÜVEN|DOĞRULAMA|ŞİMDİ|VPN|UYANIK|KALKAN|AVCI|GÜÇLÜ|TAKİP|AÇ|PAROLA)/.test(wText)) {
                  activeColor = '#00FF66';
                } else {
                  activeColor = '#00F0FF';
                }
              }
              ctx.save();
              ctx.shadowColor = activeColor;
              ctx.shadowBlur = 16;
              ctx.fillStyle = activeColor;
              ctx.fillText(wText, curX, drawY);
              ctx.restore();
            } else {
              ctx.fillStyle = '#FFFFFF';
              ctx.fillText(wText, curX, drawY);
            }
            curX += wWidth + spaceW;
          }
        } else {
          ctx.textAlign = 'center';
          ctx.strokeStyle = '#000000';
          ctx.lineWidth = 6;
          ctx.strokeText(textToDisplay, 640, cy);
          ctx.fillStyle = p.captionColor || (isCyber ? '#00FF66' : '#FFE500');
          ctx.fillText(textToDisplay, 640, cy);
        }
      } else if (p.style === 'B2') {
        const width = Math.max(...lines.map(s => ctx.measureText(s).width));
        ctx.fillStyle = 'rgba(249,246,235,.94)'; ctx.beginPath(); ctx.roundRect(640 - width / 2 - 22, cy - 33, width + 44, 66 + (lines.length - 1) * 43, 9); ctx.fill();
        lines.forEach((s, i) => { ctx.fillStyle = '#38382f'; ctx.fillText(s, 640, cy + i * 43); });
      } else lines.forEach((s, i) => {
        const y = cy + i * 43;
        ctx.strokeStyle = p.style === 'B4' ? '#303234' : p.style === 'B3' ? '#111316' : '#FDFDFC';
        ctx.lineWidth = ['B3', 'B4'].includes(p.style) ? 3 : 7;
        if (p.style === 'B4') { ctx.shadowColor = 'rgba(0,0,0,.45)'; ctx.shadowBlur = 2; ctx.shadowOffsetY = 2; }
        ctx.strokeText(s, 640, y); ctx.fillStyle = p.style === 'B4' ? '#f8f5ed' : p.style === 'B3' ? '#f4f5f5' : '#20231F'; ctx.fillText(s, 640, y);
        ctx.shadowBlur = 0; ctx.shadowOffsetY = 0;
      });
      ctx.textAlign = 'left';
    }
    return canvas;
  };
}
async function encode(p, frame, start, frames, target, onFrame) {
  const child = spawn(FFMPEG, ['-v', 'error', '-y', '-f', 'rawvideo', '-pixel_format', 'rgba', '-video_size', `${p.width}x${p.height}`, '-framerate', String(p.fps), '-i', 'pipe:0', '-an', '-c:v', 'libx264', '-preset', 'fast', '-crf', '20', '-pix_fmt', 'yuv420p', target], {stdio: ['pipe', 'ignore', 'pipe']});
  let error = ''; child.stderr.on('data', b => { error = (error + b).slice(-2400); });
  child.stdin.on('error', () => {});
  const completion = new Promise((resolve, reject) => { child.on('error', reject); child.on('close', code => code === 0 ? resolve() : reject(new Error(`Encode failed: ${error}`))); });
  completion.catch(() => {});
  try {
    for (let i = 0; i < frames; i++) {
      const canvas = frame((start + i) / p.fps);
      if (onFrame) onFrame(canvas, start + i + 1);
      if (!child.stdin.write(canvas.data())) await Promise.race([once(child.stdin, 'drain'), completion.then(() => { throw new Error('Encoder closed early'); })]);
    }
    child.stdin.end(); await completion;
  } catch (e) { child.kill(); await completion.catch(() => {}); throw e; }
}
function makeSfx(p, file, duration) {
  const rate = 48000, n = Math.ceil(duration * rate), samples = new Float32Array(n);
  let seed = 8731; const noise = () => { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed / 4294967296 * 2 - 1; };
  for (const s of p.sfx || []) {
    if (s.kind === 'file') continue;
    const len = s.kind === 'jingle' ? 0.52 : s.kind === 'chime' ? 0.55 : s.kind === 'alert' ? 0.42 : s.kind === 'keyboard' ? 0.12 : s.kind === 'glitch' ? 0.28 : s.kind === 'lock' ? 0.35 : s.kind === 'paper' ? 0.22 : 0.12;
    for (let i = 0; i < len * rate; i++) {
      const index = Math.round(s.time * rate) + i; if (index >= n) break;
      const t = i / rate;
      let value = 0;
      if (s.kind === 'paper') value = noise();
      else if (s.kind === 'tap') value = noise() * 0.6 + Math.sin(t * 2 * Math.PI * 170) * 0.4;
      else if (s.kind === 'chime') value = Math.sin(2 * Math.PI * (880 + 200 * t) * t) * Math.exp(-t * 8);
      else if (s.kind === 'pop') value = Math.sin(2 * Math.PI * (500 - 900 * t) * t) * Math.exp(-t * 25);
      else if (s.kind === 'alert') {
        const f = Math.floor(t * 14) % 2 === 0 ? 880 : 1320;
        value = Math.sin(2 * Math.PI * f * t) * 0.7 * Math.exp(-t * 5);
      }
      else if (s.kind === 'keyboard') {
        value = (noise() * 0.6 + Math.sin(2 * Math.PI * 3400 * t) * 0.4) * Math.exp(-t * 45);
      }
      else if (s.kind === 'glitch') {
        value = noise() * Math.sin(2 * Math.PI * 360 * t) * Math.exp(-Math.abs(t - 0.14) * 16);
      }
      else if (s.kind === 'lock') {
        value = (Math.sin(2 * Math.PI * 115 * t) * 0.6 + noise() * 0.4) * Math.exp(-t * 12);
      }
      else if (s.kind === 'jingle') {
        value = 0;
        for (const [delay, f] of [[0, 2460], [0.085, 3110], [0.21, 2080]]) {
          const q = t - delay;
          if (q >= 0) value += (Math.sin(2 * Math.PI * f * q) * 0.55 + Math.sin(2 * Math.PI * f * 1.73 * q) * 0.25 + Math.sin(2 * Math.PI * f * 2.31 * q) * 0.12) * Math.exp(-q * 18);
        }
      }
      samples[index] += value * s.gain;
    }
  }
  const bytes = Buffer.alloc(44 + n * 2); bytes.write('RIFF'); bytes.writeUInt32LE(bytes.length - 8, 4); bytes.write('WAVEfmt ', 8); bytes.writeUInt32LE(16, 16); bytes.writeUInt16LE(1, 20); bytes.writeUInt16LE(1, 22); bytes.writeUInt32LE(rate, 24); bytes.writeUInt32LE(rate * 2, 28); bytes.writeUInt16LE(2, 32); bytes.writeUInt16LE(16, 34); bytes.write('data', 36); bytes.writeUInt32LE(n * 2, 40);
  samples.forEach((s, i) => bytes.writeInt16LE(Math.round(Math.max(-1, Math.min(1, s)) * 32767), 44 + i * 2)); fs.writeFileSync(file, bytes);
}
function stamp(seconds, vtt = false) {
  const ms = Math.round(seconds * 1000); return `${String(Math.floor(ms / 3600000)).padStart(2, '0')}:${String(Math.floor(ms / 60000) % 60).padStart(2, '0')}:${String(Math.floor(ms / 1000) % 60).padStart(2, '0')}${vtt ? '.' : ','}${String(ms % 1000).padStart(3, '0')}`;
}
export async function render(p, {full = false, seconds = 20, replace = false, live = false} = {}) {
  // A preview never loads media or paints frames belonging only to the unfinished tail.
  if (!full) p = loadProject(p.file, {seconds});
  const out = outputDir(p), lock = path.join(out, 'render.lock');
  if (fs.existsSync(lock)) {
    const prior = JSON.parse(fs.readFileSync(lock, 'utf8'));
    try { process.kill(prior.pid, 0); throw new Error('Another renderer is running'); }
    catch (e) { if (e.code !== 'ESRCH') throw e; fs.unlinkSync(lock); }
  }
  const fd = fs.openSync(lock, 'wx'); fs.writeFileSync(fd, JSON.stringify({pid: process.pid})); fs.closeSync(fd);
  const kind = full ? 'film' : 'preview', target = path.join(out, `${kind}.mp4`), reportPath = path.join(out, `${kind}-report.json`);
  const duration = Math.min(p.duration, full ? p.duration : seconds), count = Math.ceil(duration * p.fps - 1e-6);
  let lastUpdate = 0;
  const thumbnail = live ? createCanvas(360, Math.round(360 * p.height / p.width)) : null;
  const onFrame = live ? (canvas, current, force = false) => {
    if (!force && current !== count && Date.now() - lastUpdate < 500) return;
    lastUpdate = Date.now();
    thumbnail.getContext('2d').drawImage(canvas, 0, 0, thumbnail.width, thumbnail.height);
    const destination = path.join(out, 'live-frame.jpg'), temporary = destination + '.tmp';
    fs.writeFileSync(temporary, thumbnail.toBuffer('image/jpeg'));
    fs.renameSync(temporary, destination);
    process.stdout.write('TUBIE_EVENT ' + JSON.stringify({kind, frame: current, frames: count, seconds: current / p.fps, duration}) + '\n');
  } : null;
  const done = value => save(path.join(out, 'checkpoint.json'), {stage: value, kind, fingerprint: p.fingerprint, pid: process.pid, updated_at: new Date().toISOString()});
  try {
    if (fs.existsSync(target) && !replace) {
      if (fs.existsSync(reportPath)) { const previous = JSON.parse(fs.readFileSync(reportPath)); if (previous.fingerprint === p.fingerprint && previous.frames === count && previous.output_sha256 === hashFile(target)) return {...previous, reused: true}; }
      throw new Error('Output exists with different/unverified contents; pass --replace to replace renderer-owned output');
    }
    const cache = path.join(out, 'cache'); fs.mkdirSync(cache, {recursive: true});
    const frame = await painter(p), clips = []; let reused = 0;
    const rendererHash = sha(['autopilot-render.mjs', 'b3-graphics.mjs'].map(f => hashFile(path.join(ROOT, 'scripts', f))).join(':'));
    for (const [i, shot] of p.shots.entries()) {
      const first = Math.round(shot.start * p.fps), last = Math.min(count, i === p.shots.length - 1 ? count : Math.round(shot.end * p.fps)); if (last <= first) continue;
      const files = new Set(Object.values(fontPaths(p)));
      for (const layer of walkLayers(shot.layers)) {
        if (layer.type === 'image') files.add(mediaPath(p.base, layer.path));
        if (layer.type === 'pose') files.add(mediaPath(p.base, '@b1/character-pose-sheet-v6.png'));
      }
      const key = sha(JSON.stringify({shot, first, last, style: p.style, television: p.television, width: p.width, height: p.height, fps: p.fps, captions: p.captions.filter(c => c.end > first / p.fps && c.start < last / p.fps), captionSize: p.captionSize, grid: p.grid, inputs: p.inputs.filter(f => files.has(f.path)), rendererHash}));
      const clip = path.join(cache, `${key}.mp4`); let valid = false;
      if (fs.existsSync(clip)) try { const info = probe(clip); valid = Number(info.streams[0].nb_frames) === last - first; } catch {}
      if (valid) {
        reused++;
        if (onFrame) onFrame(frame((last - 1) / p.fps), last, true);
      } else {
        done(`rendering shot ${i + 1}/${p.shots.length}`);
        const tmp = path.join(cache, `${key}.partial.mp4`);
        await encode(p, frame, first, last - first, tmp, onFrame);
        fs.renameSync(tmp, clip);
      }
      clips.push(clip); process.stdout.write(JSON.stringify({shot: i + 1, total: p.shots.length, cache_hit: valid}) + '\n');
    }
    done('assembling original-speed voice, captions and SFX');
    const list = path.join(cache, 'concat.txt'); fs.writeFileSync(list, clips.map(concatFileLine).join('\n') + '\n');
    const args = ['-v', 'error', '-y', '-f', 'concat', '-safe', '0', '-i', list, '-i', p.voicePath];
    const trim = p.voiceTrimStart ? `atrim=start=${p.voiceTrimStart},asetpts=PTS-STARTPTS,` : '';
    const filters = [`[1:a]${trim}adelay=${Math.round(p.lead * 1000)}:all=1,apad,atrim=0:${count / p.fps}[voice]`], mixes = ['[voice]'], backgrounds = [];
    const effects = p.style === 'B4' ? backgrounds : mixes; let idx = 2;
    if (p.sfx?.some(s => s.kind !== 'file')) { const f = path.join(cache, 'sfx.wav'); makeSfx(p, f, count / p.fps); args.push('-i', f); effects.push(`[${idx++}:a]`); }
    if (p.style === 'B4') idx = appendB4Audio(p, args, filters, mixes, backgrounds, idx);
    if (p.music) { args.push('-stream_loop', '-1', '-i', mediaPath(p.base, p.music.path)); filters.push(`[${idx++}:a]volume=${p.music.gain}[music]`); effects.push('[music]'); }
    if (backgrounds.length) {
      const cuts = p.audio_cuts.map(cut => `volume=0:enable='gte(t,${cut.start})*lt(t,${cut.end})'`);
      filters.push(`${backgrounds.join('')}amix=inputs=${backgrounds.length}:normalize=0:duration=longest${cuts.length ? ',' + cuts.join(',') : ''}[background]`);
      mixes.push('[background]');
    }
    filters.push(`${mixes.join('')}amix=inputs=${mixes.length}:normalize=0:duration=first,alimiter=limit=0.97:level=false:latency=true[a]`);
    const partial = path.join(out, `${kind}.partial.mp4`);
    args.push('-filter_complex', filters.join(';'), '-map', '0:v:0', '-map', '[a]', '-c:v', 'copy', '-c:a', 'aac', '-b:a', '192k', '-t', String(count / p.fps), '-movflags', '+faststart', partial);
    run(FFMPEG, args); run(FFMPEG, ['-v', 'error', '-i', partial, '-f', 'null', '-']); fs.renameSync(partial, target);
    const cues = p.captions.filter(c => c.start < duration).map(c => ({...c, end: Math.min(c.end, duration)}));
    fs.writeFileSync(path.join(out, `${kind}.srt`), cues.map((c, i) => `${i + 1}\n${stamp(c.start)} --> ${stamp(c.end)}\n${c.text}\n`).join('\n'));
    fs.writeFileSync(path.join(out, `${kind}.vtt`), 'WEBVTT\n\n' + cues.map(c => `${stamp(c.start, true)} --> ${stamp(c.end, true)}\n${c.text}\n`).join('\n'));
    const sheet = createCanvas(960, 540), sc = sheet.getContext('2d');
    for (let i = 0; i < 9; i++) sc.drawImage(frame(Math.min(duration - 1 / p.fps, (i + 0.5) * duration / 9)), (i % 3) * 320, Math.floor(i / 3) * 180, 320, 180);
    fs.writeFileSync(path.join(out, `${kind}-contact-sheet.png`), sheet.toBuffer('image/png'));
    const info = probe(target), report = {fingerprint: p.fingerprint, preview_seconds: full ? null : duration, output: target, output_sha256: hashFile(target), original_voice_sha256: hashFile(p.voicePath), original_voice_duration: p.voiceDuration, voice_rate: 'unchanged; leading/trailing padding only', frames: count, duration: Number(info.format.duration), cache_hits: reused, shot_count: clips.length, full_decode: 'pass', caption_count: cues.length, human_review: 'required; technical checks are not acceptance', inputs: p.inputs};
    if (p.style === 'B4') {
      report.voice_trim_start_seconds = p.voiceTrimStart;
      report.voice_rate = 'unchanged; explicit leading source trim/padding only';
      report.opening_audio = checkOpeningFile(target, FFMPEG);
      report.voice_identity = {name: p.voice.name, id: p.voice.id, evidence: 'project metadata; source identity must be verified before generation'};
      report.audio_tracks = {ambience: p.ambience.length, dialogue: p.dialogue.length, file_sfx: p.sfx.filter(s => s.kind === 'file').length, cuts: p.audio_cuts.length};
      save(path.join(out, `${kind}-opening-audio.json`), report.opening_audio);
    }
    save(reportPath, report); done('complete'); return report;
  } catch (e) { done(`failed: ${e.message}`); throw e; }
  finally { fs.unlinkSync(lock); }
}
