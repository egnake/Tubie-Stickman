const fs = require('fs');
const path = require('path');

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {
    title: 'SİBER GÜVENLİK UYARISI',
    topic: '',
    prop: '',
    pose: 1,
    aspect: '9:16',
    variant: 'a',
    mood: 'danger',
    out: ''
  };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--title' && args[i + 1]) out.title = args[++i];
    else if (args[i] === '--topic' && args[i + 1]) out.topic = args[++i];
    else if (args[i] === '--prop' && args[i + 1]) out.prop = args[++i];
    else if (args[i] === '--pose' && args[i + 1]) out.pose = parseInt(args[++i], 10) || 1;
    else if (args[i] === '--aspect' && args[i + 1]) out.aspect = args[++i];
    else if (args[i] === '--variant' && args[i + 1]) out.variant = args[++i].toLowerCase();
    else if (args[i] === '--mood' && args[i + 1]) out.mood = args[++i].toLowerCase();
    else if (args[i] === '--out' && args[i + 1]) out.out = args[++i];
  }
  return out;
}

function rrect(ctx, x, y, w, h, r) {
  const rad = Math.max(0, Math.min(r, w / 2, h / 2));
  ctx.beginPath();
  if (typeof ctx.roundRect === 'function') ctx.roundRect(x, y, w, h, rad);
  else ctx.rect(x, y, w, h);
}

function drawVectorWarningTriangle(ctx, cx, cy, r, col) {
  ctx.save();
  ctx.strokeStyle = col;
  ctx.fillStyle = col;
  ctx.lineWidth = Math.max(2, r * 0.18);
  ctx.lineJoin = 'round';
  ctx.beginPath();
  ctx.moveTo(cx, cy - r);
  ctx.lineTo(cx - r * 0.92, cy + r * 0.78);
  ctx.lineTo(cx + r * 0.92, cy + r * 0.78);
  ctx.closePath();
  ctx.stroke();
  ctx.fillRect(cx - r * 0.09, cy - r * 0.34, r * 0.18, r * 0.56);
  ctx.fillRect(cx - r * 0.09, cy + r * 0.35, r * 0.18, r * 0.18);
  ctx.restore();
}

function drawShockedStickman(ctx, x, y, w, h, accentColor) {
  const cx = x + w * 0.5;
  const strokeW = Math.max(6, w * 0.032);
  const headR = w * 0.21;
  const headY = y + h * 0.26;
  const neckY = headY + headR;
  const shY = neckY + h * 0.05;
  const hipY = shY + h * 0.30;
  const footY = y + h * 0.96;

  ctx.save();
  // Grounded studio contact shadow + soft rim
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)';
  ctx.beginPath();
  ctx.ellipse(cx, footY + 4, w * 0.28, Math.max(8, h * 0.032), 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.strokeStyle = accentColor === '#10B981' ? 'rgba(16, 185, 129, 0.26)' : 'rgba(56, 189, 248, 0.28)';
  ctx.lineWidth = 1.6;
  ctx.stroke();

  const aura = ctx.createRadialGradient(cx, y + h * 0.44, w * 0.08, cx, y + h * 0.44, w * 0.56);
  aura.addColorStop(0, accentColor === '#10B981' ? 'rgba(16, 185, 129, 0.14)' : 'rgba(56, 189, 248, 0.15)');
  aura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = aura;
  ctx.beginPath();
  ctx.arc(cx, y + h * 0.44, w * 0.56, 0, Math.PI * 2);
  ctx.fill();

  ctx.strokeStyle = '#F1F5F9';
  ctx.lineWidth = strokeW;
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.shadowColor = 'rgba(0, 0, 0, 0.5)';
  ctx.shadowBlur = 6;

  // Curved organic legs
  ctx.beginPath();
  ctx.moveTo(cx, hipY); ctx.quadraticCurveTo(cx - w * 0.16, (hipY + footY) * 0.50, cx - w * 0.22, footY);
  ctx.moveTo(cx, hipY); ctx.quadraticCurveTo(cx + w * 0.16, (hipY + footY) * 0.50, cx + w * 0.22, footY);
  ctx.stroke();

  // Curved spine
  ctx.beginPath();
  ctx.moveTo(cx, neckY);
  ctx.quadraticCurveTo(cx - w * 0.025, (neckY + hipY) * 0.5, cx, hipY);
  ctx.stroke();

  // Smooth curved hands-on-cheeks shocked arms
  ctx.beginPath();
  ctx.moveTo(cx, shY); ctx.quadraticCurveTo(cx - w * 0.31, headY + headR * 0.42, cx - headR * 0.96, headY);
  ctx.moveTo(cx, shY); ctx.quadraticCurveTo(cx + w * 0.31, headY + headR * 0.42, cx + headR * 0.96, headY);
  ctx.stroke();

  // Head circle
  ctx.fillStyle = '#09111E';
  ctx.beginPath();
  ctx.arc(cx, headY, headR, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Signature two hair tufts
  ctx.beginPath();
  ctx.moveTo(cx - headR * 0.08, headY - headR);
  ctx.quadraticCurveTo(cx + headR * 0.12, headY - headR * 1.42, cx + headR * 0.38, headY - headR * 1.32);
  ctx.moveTo(cx + headR * 0.06, headY - headR);
  ctx.quadraticCurveTo(cx + headR * 0.30, headY - headR * 1.30, cx + headR * 0.50, headY - headR * 1.14);
  ctx.stroke();

  // Expressive eyes, raised shocked brows & open screaming mouth
  ctx.shadowBlur = 0;
  ctx.fillStyle = '#FFFFFF';
  ctx.beginPath();
  ctx.ellipse(cx - headR * 0.34, headY - headR * 0.14, headR * 0.12, headR * 0.20, 0, 0, Math.PI * 2);
  ctx.ellipse(cx + headR * 0.34, headY - headR * 0.14, headR * 0.12, headR * 0.20, 0, 0, Math.PI * 2);
  ctx.fill();

  ctx.strokeStyle = '#FFFFFF';
  ctx.lineWidth = Math.max(3, strokeW * 0.42);
  ctx.beginPath();
  ctx.moveTo(cx - headR * 0.52, headY - headR * 0.42); ctx.lineTo(cx - headR * 0.16, headY - headR * 0.52);
  ctx.moveTo(cx + headR * 0.52, headY - headR * 0.42); ctx.lineTo(cx + headR * 0.16, headY - headR * 0.52);
  ctx.stroke();

  ctx.fillStyle = '#04070E';
  ctx.beginPath();
  ctx.ellipse(cx, headY + headR * 0.38, headR * 0.27, headR * 0.32, 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Shock rays & exclamation spikes
  ctx.strokeStyle = '#FBBF24';
  ctx.fillStyle = '#FBBF24';
  ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
  ctx.shadowBlur = 5;
  ctx.lineWidth = Math.max(3, strokeW * 0.36);
  ctx.beginPath();
  ctx.moveTo(cx - headR * 1.35, headY - headR * 0.45); ctx.lineTo(cx - headR * 1.65, headY - headR * 0.65);
  ctx.moveTo(cx - headR * 1.40, headY - headR * 0.05); ctx.lineTo(cx - headR * 1.75, headY - headR * 0.10);
  ctx.moveTo(cx + headR * 1.35, headY - headR * 0.45); ctx.lineTo(cx + headR * 1.65, headY - headR * 0.65);
  ctx.moveTo(cx + headR * 1.40, headY - headR * 0.05); ctx.lineTo(cx + headR * 1.75, headY - headR * 0.10);
  ctx.stroke();

  ctx.font = `900 ${Math.round(w * 0.13)}px NotoSans, sans-serif`;
  ctx.textAlign = 'center';
  ctx.fillText('!!', cx - headR * 1.1, headY - headR * 1.05);
  ctx.restore();
}

function drawHoodedHacker(ctx, x, y, w, h) {
  const cx = x + w * 0.5;
  const headR = w * 0.20;
  const headY = y + h * 0.26;
  const shY = headY + headR * 1.15;
  const hipY = shY + h * 0.36;
  const baseY = y + h * 0.95;

  ctx.save();
  // Grounded studio contact shadow
  ctx.fillStyle = 'rgba(0, 0, 0, 0.48)';
  ctx.beginPath();
  ctx.ellipse(cx, baseY, w * 0.32, Math.max(8, h * 0.032), 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.strokeStyle = 'rgba(244, 63, 94, 0.25)';
  ctx.lineWidth = 1.6;
  ctx.stroke();

  const redAura = ctx.createRadialGradient(cx, y + h * 0.42, w * 0.06, cx, y + h * 0.42, w * 0.56);
  redAura.addColorStop(0, 'rgba(244, 63, 94, 0.16)');
  redAura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = redAura;
  ctx.beginPath();
  ctx.arc(cx, y + h * 0.42, w * 0.56, 0, Math.PI * 2);
  ctx.fill();

  // Hooded cloak silhouette
  ctx.fillStyle = '#111827';
  ctx.strokeStyle = '#F43F5E';
  ctx.lineWidth = Math.max(4.5, w * 0.022);
  ctx.shadowColor = 'rgba(0, 0, 0, 0.55)';
  ctx.shadowBlur = 8;
  ctx.beginPath();
  ctx.moveTo(cx, headY - headR * 1.50);
  ctx.quadraticCurveTo(cx + headR * 1.50, headY - headR * 0.35, cx + w * 0.35, hipY + h * 0.14);
  ctx.lineTo(cx - w * 0.35, hipY + h * 0.14);
  ctx.quadraticCurveTo(cx - headR * 1.50, headY - headR * 0.35, cx, headY - headR * 1.50);
  ctx.closePath();
  ctx.fill();
  ctx.stroke();

  // Dark head face void
  ctx.fillStyle = '#070A12';
  ctx.beginPath();
  ctx.arc(cx, headY, headR * 0.94, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Sharp coral visor eyes
  ctx.fillStyle = '#F43F5E';
  ctx.shadowColor = '#F43F5E';
  ctx.shadowBlur = 6;
  ctx.beginPath();
  ctx.moveTo(cx - headR * 0.54, headY - headR * 0.16);
  ctx.lineTo(cx - headR * 0.10, headY - headR * 0.02);
  ctx.lineTo(cx - headR * 0.44, headY + headR * 0.10);
  ctx.closePath();
  ctx.fill();

  ctx.beginPath();
  ctx.moveTo(cx + headR * 0.54, headY - headR * 0.16);
  ctx.lineTo(cx + headR * 0.10, headY - headR * 0.02);
  ctx.lineTo(cx + headR * 0.44, headY + headR * 0.10);
  ctx.closePath();
  ctx.fill();

  // Menacing curved reach arms
  ctx.shadowBlur = 4;
  ctx.strokeStyle = '#F43F5E';
  ctx.lineWidth = Math.max(4.5, w * 0.026);
  ctx.lineCap = 'round';
  ctx.beginPath();
  ctx.moveTo(cx - w * 0.15, shY); ctx.quadraticCurveTo(cx - w * 0.34, shY + h * 0.10, cx - w * 0.42, shY + h * 0.02);
  ctx.moveTo(cx + w * 0.15, shY); ctx.quadraticCurveTo(cx + w * 0.30, shY + h * 0.14, cx + w * 0.18, shY + h * 0.24);
  ctx.stroke();
  ctx.restore();
}

function drawFocalThreatCard(ctx, x, y, w, h, titleLower, variant) {
  let cardHeader = 'KRİTİK İHLAL';
  let cardMain = 'HESAP ELE GEÇİRİLDİ!';
  let accent = variant === 'b' ? '#FBBF24' : '#F43F5E';

  if (titleLower.includes('yapay zeka') || titleLower.includes('dijital iz') || titleLower.includes('gizlilik')) {
    cardHeader = 'VERİ PROFİLİ';
    cardMain = 'HER ŞEY KAYITLI!';
  } else if (titleLower.includes('kripto') || titleLower.includes('cüzdan') || titleLower.includes('airdrop')) {
    cardHeader = 'CÜZDAN BAKİYESİ';
    cardMain = '0.00 TL (BOŞALTILDI)';
  } else if (titleLower.includes('crack') || titleLower.includes('apk') || titleLower.includes('infostealer') || titleLower.includes('zararlı')) {
    cardHeader = 'SAHTE CRACK & APK';
    cardMain = 'ÇEREZ & ŞİFRELER SIZDI!';
  } else if (titleLower.includes('wifi') || titleLower.includes('wi-fi') || titleLower.includes('ortak ağ') || titleLower.includes('kafe') || /\bağ(ı|da|dan|lar)?\b/i.test(titleLower)) {
    cardHeader = 'EVIL TWIN AĞI';
    cardMain = 'ŞİFRELER ÇALINIYOR!';
  } else if (titleLower.includes('sql')) {
    cardHeader = "' OR 1=1 --";
    cardMain = 'ADMİN YETKİSİ AÇILDI!';
  } else if (titleLower.includes('link') || titleLower.includes('oltalama') || titleLower.includes('phishing') || /\bqr\b/i.test(titleLower)) {
    cardHeader = 'SAHTE BAĞLANTI';
    cardMain = 'HTTPS != GÜVENLİ!';
  } else if (titleLower.includes('deepfake') || /\bses(inle|ini|i)?\b/i.test(titleLower)) {
    cardHeader = 'YAPAY ZEKA SES';
    cardMain = '3 SN İLE KLONLANDI!';
  } else if (titleLower.includes('whatsapp') || /\bsms\b|\bsim\b/i.test(titleLower)) {
    cardHeader = '6 HANELİ KOD';
    cardMain = 'ASLA PAYLAŞMA!';
  } else if (titleLower.includes('şifre') || titleLower.includes('parola') || titleLower.includes('brute')) {
    cardHeader = 'KABA KUVVET';
    cardMain = '0.4 SANİYEDE KIRILDI!';
  } else if (variant === 'b' && (titleLower.includes('koru') || titleLower.includes('rehber'))) {
    cardHeader = 'SİBER KALKAN';
    cardMain = '3 ALTIN GÜVENLİK KURALI';
    accent = '#10B981';
  }

  ctx.save();
  ctx.shadowColor = 'rgba(0, 0, 0, 0.52)';
  ctx.shadowBlur = 18;
  ctx.shadowOffsetY = 6;
  ctx.fillStyle = 'rgba(17, 24, 39, 0.96)';
  ctx.strokeStyle = accent;
  ctx.lineWidth = 2.2;
  rrect(ctx, x, y, w, h, 16);
  ctx.fill();
  ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.shadowOffsetY = 0;
  drawVectorWarningTriangle(ctx, x + 36, y + h * 0.34, 15, '#FBBF24');
  ctx.fillStyle = '#FBBF24';
  ctx.font = `800 ${Math.max(15, Math.round(h * 0.17))}px NotoSans, sans-serif`;
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  ctx.fillText(cardHeader, x + 64, y + h * 0.34);

  ctx.fillStyle = '#F8FAFC';
  ctx.font = `900 ${Math.max(18, Math.round(h * 0.22))}px NotoSans, sans-serif`;
  ctx.fillText(cardMain, x + 22, y + h * 0.72, w - 44);
  ctx.restore();
}

function buildHeadlineLines(title, maxWidth, ctx, fontSize, maxLines = 3) {
  const cleaned = String(title || 'SİBER GÜVENLİK UYARISI')
    .toLocaleUpperCase('tr-TR')
    .replace(/[?!.,'"]/g, '')
    .trim();
  const words = cleaned.split(/\s+/).filter(Boolean);
  ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;

  const lines = [];
  let cur = '';
  for (const word of words) {
    const cand = cur ? `${cur} ${word}` : word;
    if (ctx.measureText(cand).width <= maxWidth) {
      cur = cand;
    } else if (cur) {
      lines.push(cur);
      cur = word;
    } else {
      lines.push(word);
      cur = '';
    }
    if (lines.length === maxLines) {
      cur = '';
      break;
    }
  }
  if (cur && lines.length < maxLines) lines.push(cur);
  return lines.slice(0, maxLines);
}

function loadCanvasModule(repoRoot) {
  const candidates = [
    path.resolve(repoRoot, 'skills/directing-stickman-videos/node_modules/@napi-rs/canvas'),
    path.resolve(repoRoot, 'node_modules/@napi-rs/canvas'),
    '@napi-rs/canvas',
  ];
  for (const candidate of candidates) {
    if (candidate.startsWith('@') || fs.existsSync(candidate)) {
      try { return require(candidate); } catch {}
    }
  }
  return require('@napi-rs/canvas');
}

function writeFileAtomic(target, buf) {
  const tmp = `${target}.${process.pid}.tmp`;
  fs.writeFileSync(tmp, buf);
  const retryCodes = new Set(['EPERM', 'EBUSY', 'EACCES']);
  for (let attempt = 0; attempt < 10; attempt++) {
    try {
      fs.renameSync(tmp, target);
      return;
    } catch (e) {
      if (!retryCodes.has(e && e.code)) throw e;
      try { Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, Math.min(100, 15 * (attempt + 1))); } catch {}
    }
  }
  fs.copyFileSync(tmp, target);
  try { fs.unlinkSync(tmp); } catch {}
}

async function main() {
  const opts = parseArgs();
  if (!opts.out) {
    console.error('Missing --out parameter');
    process.exit(1);
  }

  const repoRoot = path.resolve(__dirname, '../..');
  const { createCanvas, loadImage, GlobalFonts } = loadCanvasModule(repoRoot);

  const isLandscape = opts.aspect === '16:9';
  const width = isLandscape ? 1280 : 720;
  const height = isLandscape ? 720 : 1280;
  const canvas = createCanvas(width, height);
  const ctx = canvas.getContext('2d');

  const notoSans = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/NotoSans.ttf');
  const caveat = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/Caveat.ttf');
  if (fs.existsSync(notoSans)) GlobalFonts.registerFromPath(notoSans, 'NotoSans');
  if (fs.existsSync(caveat)) GlobalFonts.registerFromPath(caveat, 'Caveat');

  const isVariantB = opts.variant === 'b';
  const lowerTitle = `${opts.title} ${opts.topic}`.toLocaleLowerCase('tr-TR');

  // 1. Dark-Editorial Slate Studio Backdrop
  const bgGrad = ctx.createLinearGradient(0, 0, width, height);
  bgGrad.addColorStop(0, isVariantB ? '#0E1726' : '#0F172A');
  bgGrad.addColorStop(0.55, '#0B0F19');
  bgGrad.addColorStop(1, '#070A12');
  ctx.fillStyle = bgGrad;
  ctx.fillRect(0, 0, width, height);

  // 2. Subtle Architectural Dot-Grid (no noisy scanlines)
  ctx.save();
  ctx.fillStyle = 'rgba(148, 163, 184, 0.08)';
  const gridSize = 48;
  for (let x = 24; x < width; x += gridSize) {
    for (let y = 24; y < height; y += gridSize) {
      ctx.fillRect(x, y, 1.5, 1.5);
    }
  }
  ctx.restore();

  // 3. Clean Editorial Top Header & Divider
  ctx.save();
  const primaryAccent = isVariantB ? '#10B981' : '#38BDF8';
  const pad = 28;
  ctx.strokeStyle = 'rgba(148, 163, 184, 0.18)';
  ctx.lineWidth = 1.5;
  ctx.beginPath(); ctx.moveTo(pad, pad + 32); ctx.lineTo(width - pad, pad + 32); ctx.stroke();

  ctx.font = '800 15px NotoSans, sans-serif';
  ctx.fillStyle = primaryAccent;
  ctx.textAlign = 'left';
  ctx.fillText(isVariantB ? 'KORUNMA REHBERİ' : 'SİBER GÜVENLİK ANALİZİ', pad, pad + 18);

  ctx.textAlign = 'right';
  ctx.fillStyle = 'rgba(148, 163, 184, 0.78)';
  ctx.fillText('TUBIE STUDIO', width - pad, pad + 18);
  ctx.restore();

  // Badge label
  let badgeText = isVariantB ? '3 ALTIN KORUNMA KURALI' : 'SİBER GÜVENLİK UYARISI';
  if (lowerTitle.includes('yapay zeka') || lowerTitle.includes('gizlilik') || lowerTitle.includes('sohbet')) {
    badgeText = isVariantB ? 'DİJİTAL GİZLİLİK REHBERİ' : 'YAPAY ZEKA VERİ RİSKİ';
  } else if (lowerTitle.includes('instagram') || lowerTitle.includes('dm')) {
    badgeText = 'HESAP GÜVENLİĞİ ALARMI';
  } else if (lowerTitle.includes('kripto') || lowerTitle.includes('cüzdan')) {
    badgeText = 'KRİPTO CÜZDAN UYARISI';
  } else if (lowerTitle.includes('whatsapp') || lowerTitle.includes('sms') || lowerTitle.includes('sim')) {
    badgeText = 'SMS & HESAP TUZAĞI';
  } else if (lowerTitle.includes('wifi') || lowerTitle.includes('wi-fi') || lowerTitle.includes('şifre')) {
    badgeText = 'KRİTİK GÜVENLİK AÇIĞI';
  }

  // Optional focal prop icon if explicitly provided and inside assets/b1
  let propImg = null;
  if (opts.prop) {
    let resolvedProp = opts.prop;
    if (resolvedProp.startsWith('@b1/')) {
      resolvedProp = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1', resolvedProp.slice(4));
    }
    if (fs.existsSync(resolvedProp)) {
      try { propImg = await loadImage(resolvedProp); } catch {}
    }
  }

  if (isLandscape) {
    // 16:9 YouTube Widescreen Composition (Left Visual Duel, Right Giant Text)
    drawShockedStickman(ctx, 25, 105, 300, 395, isVariantB ? '#10B981' : '#38BDF8');
    drawHoodedHacker(ctx, 285, 85, 310, 415);

    if (propImg) {
      ctx.save();
      ctx.shadowColor = 'rgba(0, 0, 0, 0.55)';
      ctx.shadowBlur = 14;
      ctx.drawImage(propImg, 470, 95, 115, 115 * propImg.height / propImg.width);
      ctx.restore();
    }

    drawFocalThreatCard(ctx, 48, 535, 535, 124, lowerTitle, opts.variant);

    // Right stage: Badge + Giant 3-Line Headline + Bottom CTA
    ctx.save();
    ctx.font = '800 20px NotoSans, sans-serif';
    const bW = ctx.measureText(badgeText).width + 64;
    const bX = 635;
    const bY = 78;
    ctx.fillStyle = 'rgba(17, 24, 39, 0.96)';
    ctx.strokeStyle = isVariantB ? '#10B981' : '#F43F5E';
    ctx.lineWidth = 2;
    rrect(ctx, bX, bY, bW, 44, 22);
    ctx.fill();
    ctx.stroke();
    drawVectorWarningTriangle(ctx, bX + 26, bY + 22, 10, '#FBBF24');
    ctx.fillStyle = '#FBBF24';
    ctx.textBaseline = 'middle';
    ctx.fillText(badgeText, bX + 46, bY + 22);
    ctx.restore();

    const fontSize = 58;
    const maxHeadlineW = width - 660;
    const lines = buildHeadlineLines(opts.title, maxHeadlineW, ctx, fontSize, 3);
    const startY = 152;
    const lineH = 74;
    ctx.save();
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
    lines.forEach((line, i) => {
      const yPos = startY + i * lineH;
      const col = i === lines.length - 1 ? (isVariantB ? '#34D399' : '#F43F5E') : '#F8FAFC';
      ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;
      ctx.shadowColor = 'rgba(0,0,0,0.85)';
      ctx.shadowBlur = 14;
      ctx.strokeStyle = '#070A12';
      ctx.lineWidth = 10;
      ctx.strokeText(line, 635, yPos);
      ctx.fillStyle = col;
      ctx.fillText(line, 635, yPos);
    });

    const lineY = startY + lines.length * lineH + 12;
    ctx.strokeStyle = isVariantB ? '#10B981' : '#38BDF8';
    ctx.lineWidth = 4;
    ctx.beginPath(); ctx.moveTo(635, lineY); ctx.lineTo(width - 48, lineY); ctx.stroke();

    // Right Bottom Clean Editorial CTA Pill
    const ctaX = 635, ctaY = 568, ctaW = width - 680, ctaH = 74;
    const grad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
    grad.addColorStop(0, isVariantB ? '#34D399' : '#38BDF8');
    grad.addColorStop(1, isVariantB ? '#10B981' : '#34D399');
    ctx.fillStyle = grad;
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 16;
    ctx.shadowOffsetY = 6;
    rrect(ctx, ctaX, ctaY, ctaW, ctaH, 18);
    ctx.fill();

    ctx.shadowBlur = 0;
    ctx.shadowOffsetY = 0;
    ctx.fillStyle = '#070B14';
    ctx.font = '900 24px NotoSans, sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(isVariantB ? 'HEMEN ÖNLEMİNİ AL >>' : '30 SANİYEDE KORUNMAYI ÖĞREN >>', ctaX + ctaW / 2, ctaY + ctaH / 2);
    ctx.restore();
  } else {
    // 9:16 Vertical Shorts / Reels Composition (720x1280)
    // Top Hook Badge
    ctx.save();
    ctx.font = '800 21px NotoSans, sans-serif';
    const badgeW = Math.min(width - 80, ctx.measureText(badgeText).width + 74);
    const badgeX = (width - badgeW) / 2;
    const badgeY = 74;
    const badgeH = 46;
    ctx.fillStyle = 'rgba(17, 24, 39, 0.96)';
    ctx.strokeStyle = isVariantB ? '#10B981' : '#F43F5E';
    ctx.lineWidth = 2;
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 12;
    rrect(ctx, badgeX, badgeY, badgeW, badgeH, 23);
    ctx.fill();
    ctx.stroke();

    ctx.shadowBlur = 0;
    drawVectorWarningTriangle(ctx, badgeX + 26, badgeY + badgeH / 2, 10, '#FBBF24');
    ctx.fillStyle = '#FBBF24';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(badgeText, width / 2 + 12, badgeY + badgeH / 2);
    ctx.restore();

    // Giant 3-Line Hook Headline
    const fontSize = 58;
    const lines = buildHeadlineLines(opts.title, width - 84, ctx, fontSize, 3);
    const startY = 148;
    const lineH = 74;
    ctx.save();
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    lines.forEach((line, i) => {
      const yPos = startY + i * lineH;
      const color = i === lines.length - 1 ? (isVariantB ? '#34D399' : '#F43F5E') : '#F8FAFC';
      ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;
      ctx.shadowColor = 'rgba(0,0,0,0.85)';
      ctx.shadowBlur = 14;
      ctx.shadowOffsetY = 6;
      ctx.strokeStyle = '#070A12';
      ctx.lineWidth = 10;
      ctx.strokeText(line, width / 2, yPos);

      if (i === lines.length - 1) {
        const tw = ctx.measureText(line).width;
        ctx.fillStyle = isVariantB ? 'rgba(16, 185, 129, 0.16)' : 'rgba(244, 63, 94, 0.16)';
        rrect(ctx, (width - tw - 36) / 2, yPos + 4, tw + 36, fontSize + 10, 12);
        ctx.fill();
      }
      ctx.fillStyle = color;
      ctx.fillText(line, width / 2, yPos);
    });

    const lastLineY = startY + lines.length * lineH + 8;
    const sepCol = isVariantB ? '#10B981' : '#38BDF8';
    ctx.strokeStyle = sepCol;
    ctx.lineWidth = 4;
    ctx.shadowBlur = 0;
    ctx.shadowOffsetY = 0;
    ctx.beginPath(); ctx.moveTo(86, lastLineY); ctx.lineTo(width - 86, lastLineY); ctx.stroke();
    ctx.restore();

    // Center Confrontation: Vector Shocked Stickman vs Hooded Hacker
    drawShockedStickman(ctx, 24, 435, 340, 445, isVariantB ? '#10B981' : '#38BDF8');
    drawHoodedHacker(ctx, 355, 415, 345, 465);

    if (propImg) {
      ctx.save();
      ctx.shadowColor = 'rgba(0, 0, 0, 0.55)';
      ctx.shadowBlur = 14;
      const pw = 135;
      ctx.drawImage(propImg, (width - pw) / 2, 430, pw, pw * propImg.height / propImg.width);
      ctx.restore();
    }

    // Threat Arrow pointing from Hacker to Victim
    ctx.save();
    ctx.strokeStyle = '#F43F5E';
    ctx.lineWidth = 5.5;
    ctx.lineCap = 'round';
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 6;
    ctx.beginPath();
    ctx.moveTo(410, 620); ctx.lineTo(310, 660);
    ctx.moveTo(310, 660); ctx.lineTo(338, 642);
    ctx.moveTo(310, 660); ctx.lineTo(334, 678);
    ctx.stroke();
    ctx.restore();

    // Focal Cyber Evidence Status Card
    drawFocalThreatCard(ctx, 48, 920, width - 96, 134, lowerTitle, opts.variant);

    // Bottom Editorial CTA Plate
    ctx.save();
    const ctaY = 1105;
    const ctaH = 78;
    const ctaW = width - 80;
    const ctaX = 40;
    const ctaGrad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
    ctaGrad.addColorStop(0, isVariantB ? '#34D399' : '#38BDF8');
    ctaGrad.addColorStop(1, isVariantB ? '#10B981' : '#34D399');
    ctx.fillStyle = ctaGrad;
    ctx.shadowColor = 'rgba(0, 0, 0, 0.45)';
    ctx.shadowBlur = 18;
    ctx.shadowOffsetY = 6;
    rrect(ctx, ctaX, ctaY, ctaW, ctaH, 18);
    ctx.fill();

    ctx.shadowBlur = 0;
    ctx.shadowOffsetY = 0;
    ctx.fillStyle = '#070B14';
    ctx.font = '900 26px NotoSans, sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(isVariantB ? 'BUNU BİLMEDEN TIKLAMA >>' : '30 SANİYEDE KORUNMAYI ÖĞREN >>', width / 2, ctaY + ctaH / 2);
    ctx.restore();
  }

  const outDir = path.dirname(opts.out);
  if (!fs.existsSync(outDir)) {
    fs.mkdirSync(outDir, { recursive: true });
  }
  const buf = canvas.toBuffer('image/jpeg', { quality: 96 });
  writeFileAtomic(opts.out, buf);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
