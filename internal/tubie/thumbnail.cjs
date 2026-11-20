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
  // Cyan / emerald halo
  const aura = ctx.createRadialGradient(cx, y + h * 0.44, w * 0.08, cx, y + h * 0.44, w * 0.62);
  aura.addColorStop(0, accentColor === '#00FF66' ? 'rgba(0, 255, 102, 0.32)' : 'rgba(0, 240, 255, 0.34)');
  aura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = aura;
  ctx.beginPath();
  ctx.arc(cx, y + h * 0.44, w * 0.62, 0, Math.PI * 2);
  ctx.fill();

  ctx.strokeStyle = '#FFFFFF';
  ctx.lineWidth = strokeW;
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.shadowColor = accentColor;
  ctx.shadowBlur = 18;

  // Legs
  ctx.beginPath();
  ctx.moveTo(cx, hipY); ctx.lineTo(cx - w * 0.16, (hipY + footY) * 0.52); ctx.lineTo(cx - w * 0.22, footY);
  ctx.moveTo(cx, hipY); ctx.lineTo(cx + w * 0.16, (hipY + footY) * 0.52); ctx.lineTo(cx + w * 0.22, footY);
  ctx.stroke();

  // Spine
  ctx.beginPath();
  ctx.moveTo(cx, neckY);
  ctx.lineTo(cx, hipY);
  ctx.stroke();

  // Hands-on-cheeks shocked arms
  ctx.beginPath();
  ctx.moveTo(cx, shY); ctx.lineTo(cx - w * 0.30, headY + headR * 0.32); ctx.lineTo(cx - headR * 0.96, headY);
  ctx.moveTo(cx, shY); ctx.lineTo(cx + w * 0.30, headY + headR * 0.32); ctx.lineTo(cx + headR * 0.96, headY);
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
  ctx.strokeStyle = '#FFE600';
  ctx.fillStyle = '#FFE600';
  ctx.shadowColor = '#FFE600';
  ctx.shadowBlur = 12;
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

  ctx.save();
  const redAura = ctx.createRadialGradient(cx, y + h * 0.42, w * 0.06, cx, y + h * 0.42, w * 0.60);
  redAura.addColorStop(0, 'rgba(255, 42, 68, 0.36)');
  redAura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = redAura;
  ctx.beginPath();
  ctx.arc(cx, y + h * 0.42, w * 0.60, 0, Math.PI * 2);
  ctx.fill();

  // Hooded cloak silhouette
  ctx.fillStyle = '#070A12';
  ctx.strokeStyle = '#FF2A44';
  ctx.lineWidth = Math.max(5, w * 0.025);
  ctx.shadowColor = '#FF2A44';
  ctx.shadowBlur = 20;
  ctx.beginPath();
  ctx.moveTo(cx, headY - headR * 1.50);
  ctx.quadraticCurveTo(cx + headR * 1.50, headY - headR * 0.35, cx + w * 0.35, hipY + h * 0.14);
  ctx.lineTo(cx - w * 0.35, hipY + h * 0.14);
  ctx.quadraticCurveTo(cx - headR * 1.50, headY - headR * 0.35, cx, headY - headR * 1.50);
  ctx.closePath();
  ctx.fill();
  ctx.stroke();

  // Dark head face void
  ctx.fillStyle = '#030509';
  ctx.beginPath();
  ctx.arc(cx, headY, headR * 0.94, 0, Math.PI * 2);
  ctx.fill();
  ctx.stroke();

  // Glowing crimson narrow visor eyes
  ctx.fillStyle = '#FF2A44';
  ctx.shadowBlur = 18;
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

  // Menacing reach arms
  ctx.strokeStyle = '#FF3B4E';
  ctx.lineWidth = Math.max(5, w * 0.028);
  ctx.lineCap = 'round';
  ctx.beginPath();
  ctx.moveTo(cx - w * 0.15, shY); ctx.lineTo(cx - w * 0.34, shY + h * 0.10); ctx.lineTo(cx - w * 0.42, shY + h * 0.02);
  ctx.moveTo(cx + w * 0.15, shY); ctx.lineTo(cx + w * 0.30, shY + h * 0.14); ctx.lineTo(cx + w * 0.18, shY + h * 0.24);
  ctx.stroke();
  ctx.restore();
}

function drawFocalThreatCard(ctx, x, y, w, h, titleLower, variant) {
  let cardHeader = 'KRİTİK İHLAL';
  let cardMain = 'HESAP ELE GEÇİRİLDİ!';
  let accent = variant === 'b' ? '#FFE600' : '#FF3B4E';

  if (titleLower.includes('yapay zeka') || titleLower.includes('dijital iz') || titleLower.includes('gizlilik')) {
    cardHeader = 'VERİ PROFİLİ';
    cardMain = 'HER ŞEY KAYITLI!';
  } else if (titleLower.includes('kripto') || titleLower.includes('cüzdan') || titleLower.includes('airdrop')) {
    cardHeader = 'CÜZDAN BAKİYESİ';
    cardMain = '0.00 TL (BOŞALTILDI)';
  } else if (titleLower.includes('wifi') || titleLower.includes('wi-fi') || titleLower.includes('ağ')) {
    cardHeader = 'EVİL TWİN AĞI';
    cardMain = 'ŞİFRELER ÇALINIYOR!';
  } else if (titleLower.includes('sql')) {
    cardHeader = "' OR 1=1 --";
    cardMain = 'ADMİN YETKİSİ AÇILDI!';
  } else if (titleLower.includes('şifre') || titleLower.includes('parola') || titleLower.includes('brute')) {
    cardHeader = 'KABA KUVVET';
    cardMain = '0.4 SANİYEDE KIRILDI!';
  } else if (titleLower.includes('link') || titleLower.includes('oltalama') || titleLower.includes('phishing') || titleLower.includes('qr')) {
    cardHeader = 'SAHTE BAĞLANTI';
    cardMain = 'HTTPS != GÜVENLİ!';
  } else if (titleLower.includes('ses') || titleLower.includes('deepfake')) {
    cardHeader = 'YAPAY ZEKA SES';
    cardMain = '3 SN İLE KLONLANDI!';
  } else if (titleLower.includes('whatsapp') || titleLower.includes('sms') || titleLower.includes('sim')) {
    cardHeader = '6 HANELİ KOD';
    cardMain = 'ASLA PAYLAŞMA!';
  } else if (variant === 'b' && (titleLower.includes('koru') || titleLower.includes('rehber'))) {
    cardHeader = 'SİBER KALKAN';
    cardMain = '3 ALTIN GÜVENLİK KURALI';
    accent = '#00FF66';
  }

  ctx.save();
  ctx.shadowColor = accent;
  ctx.shadowBlur = 22;
  ctx.fillStyle = 'rgba(7, 12, 22, 0.95)';
  ctx.strokeStyle = accent;
  ctx.lineWidth = 3;
  rrect(ctx, x, y, w, h, 16);
  ctx.fill();
  ctx.stroke();

  ctx.shadowBlur = 0;
  drawVectorWarningTriangle(ctx, x + 36, y + h * 0.34, 15, '#FFE600');
  ctx.fillStyle = '#FFE600';
  ctx.font = `800 ${Math.max(15, Math.round(h * 0.17))}px NotoSans, sans-serif`;
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  ctx.fillText(cardHeader, x + 64, y + h * 0.34);

  ctx.fillStyle = '#FFFFFF';
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

async function main() {
  const opts = parseArgs();
  if (!opts.out) {
    console.error('Missing --out parameter');
    process.exit(1);
  }

  const repoRoot = path.resolve(__dirname, '../..');
  const canvasPkg = path.resolve(repoRoot, 'skills/directing-stickman-videos/node_modules/@napi-rs/canvas');
  const { createCanvas, loadImage, GlobalFonts } = require(canvasPkg);

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

  // 1. Deep Cyber Backdrop
  const bgGrad = ctx.createLinearGradient(0, 0, width, height);
  bgGrad.addColorStop(0, isVariantB ? '#07111A' : '#0B0F1B');
  bgGrad.addColorStop(0.5, '#04070E');
  bgGrad.addColorStop(1, '#020306');
  ctx.fillStyle = bgGrad;
  ctx.fillRect(0, 0, width, height);

  // 2. Subtle Neon Grid & Scanlines
  ctx.save();
  ctx.strokeStyle = isVariantB ? 'rgba(0, 255, 128, 0.06)' : 'rgba(0, 240, 255, 0.065)';
  ctx.lineWidth = 1.2;
  const gridSize = 48;
  for (let x = 24; x < width; x += gridSize) {
    ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, height); ctx.stroke();
  }
  for (let y = 24; y < height; y += gridSize) {
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(width, y); ctx.stroke();
  }
  ctx.fillStyle = 'rgba(0, 240, 255, 0.015)';
  for (let y = 0; y < height; y += 4) ctx.fillRect(0, y, width, 1.5);
  ctx.restore();

  // 3. Corner Brackets + Top Broadcast Tag
  ctx.save();
  const primaryGlow = isVariantB ? '#00FF66' : '#00F0FF';
  ctx.strokeStyle = primaryGlow;
  ctx.lineWidth = 3.5;
  ctx.shadowColor = primaryGlow;
  ctx.shadowBlur = 10;
  const bSize = 32;
  const pad = 24;
  ctx.beginPath(); ctx.moveTo(pad, pad + bSize); ctx.lineTo(pad, pad); ctx.lineTo(pad + bSize, pad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, pad); ctx.lineTo(width - pad, pad); ctx.lineTo(width - pad, pad + bSize); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(pad, height - pad - bSize); ctx.lineTo(pad, height - pad); ctx.lineTo(pad + bSize, height - pad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, height - pad); ctx.lineTo(width - pad, height - pad); ctx.lineTo(width - pad, height - pad - bSize); ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.font = '800 16px NotoSans, sans-serif';
  ctx.fillStyle = primaryGlow;
  ctx.textAlign = 'left';
  ctx.fillText(isVariantB ? '[ SİBER SAVUNMA RAPORU ]' : '[ SİBER İSTİHBARAT ]', pad + 38, pad + 18);

  ctx.textAlign = 'right';
  ctx.fillStyle = '#FFFFFF';
  ctx.fillText('REC', width - pad - 115, pad + 18);
  ctx.fillStyle = '#FF3B30';
  ctx.beginPath();
  ctx.arc(width - pad - 100, pad + 13, 5, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = primaryGlow;
  ctx.fillText('4K UHD', width - pad - 34, pad + 18);
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
    // Left stage: Shocked Stickman + Hooded Hacker + Focal Threat Card
    drawShockedStickman(ctx, 25, 105, 300, 395, isVariantB ? '#00FF66' : '#00F0FF');
    drawHoodedHacker(ctx, 285, 85, 310, 415);

    if (propImg) {
      ctx.save();
      ctx.shadowColor = 'rgba(255, 59, 48, 0.7)';
      ctx.shadowBlur = 22;
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
    ctx.fillStyle = 'rgba(10, 16, 28, 0.95)';
    ctx.strokeStyle = isVariantB ? '#00FF66' : '#FF3B4E';
    ctx.lineWidth = 2.4;
    rrect(ctx, bX, bY, bW, 44, 22);
    ctx.fill();
    ctx.stroke();
    drawVectorWarningTriangle(ctx, bX + 26, bY + 22, 10, '#FFE600');
    ctx.fillStyle = '#FFE600';
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
      const col = i === 0 ? (isVariantB ? '#00FF66' : '#00F0FF') : i === lines.length - 1 ? (isVariantB ? '#FFE600' : '#FF3B4E') : '#FFFFFF';
      ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;
      ctx.shadowColor = 'rgba(0,0,0,0.95)';
      ctx.shadowBlur = 18;
      ctx.strokeStyle = '#04070D';
      ctx.lineWidth = 12;
      ctx.strokeText(line, 635, yPos);
      ctx.fillStyle = col;
      ctx.fillText(line, 635, yPos);
    });

    const lineY = startY + lines.length * lineH + 12;
    ctx.strokeStyle = isVariantB ? '#FFE600' : '#00FF66';
    ctx.lineWidth = 5;
    ctx.beginPath(); ctx.moveTo(635, lineY); ctx.lineTo(width - 48, lineY); ctx.stroke();

    // Right Bottom CTA Pill
    const ctaX = 635, ctaY = 568, ctaW = width - 680, ctaH = 74;
    const grad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
    grad.addColorStop(0, isVariantB ? '#FFE600' : '#00FF66');
    grad.addColorStop(1, '#00F0FF');
    ctx.fillStyle = grad;
    ctx.shadowColor = isVariantB ? 'rgba(255,230,0,0.45)' : 'rgba(0,255,102,0.5)';
    ctx.shadowBlur = 22;
    rrect(ctx, ctaX, ctaY, ctaW, ctaH, 37);
    ctx.fill();

    ctx.shadowBlur = 0;
    ctx.fillStyle = '#04070D';
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
    const badgeY = 64;
    const badgeH = 46;
    ctx.fillStyle = 'rgba(10, 16, 28, 0.94)';
    ctx.strokeStyle = isVariantB ? '#00FF66' : '#FF3B30';
    ctx.lineWidth = 2.4;
    ctx.shadowColor = isVariantB ? 'rgba(0,255,102,0.45)' : 'rgba(255,59,48,0.55)';
    ctx.shadowBlur = 14;
    rrect(ctx, badgeX, badgeY, badgeW, badgeH, 23);
    ctx.fill();
    ctx.stroke();

    ctx.shadowBlur = 0;
    drawVectorWarningTriangle(ctx, badgeX + 26, badgeY + badgeH / 2, 10, '#FFE600');
    ctx.fillStyle = '#FFE600';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(badgeText, width / 2 + 12, badgeY + badgeH / 2);
    ctx.restore();

    // Giant 3-Line Hook Headline
    const fontSize = 58;
    const lines = buildHeadlineLines(opts.title, width - 84, ctx, fontSize, 3);
    const startY = 144;
    const lineH = 74;
    ctx.save();
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    lines.forEach((line, i) => {
      const yPos = startY + i * lineH;
      const color = i === 0 ? (isVariantB ? '#00FF66' : '#00F0FF') : i === lines.length - 1 ? (isVariantB ? '#FFE600' : '#FF3B4E') : '#FFFFFF';
      ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;
      ctx.shadowColor = 'rgba(0,0,0,0.96)';
      ctx.shadowBlur = 22;
      ctx.shadowOffsetY = 7;
      ctx.strokeStyle = '#04070D';
      ctx.lineWidth = 14;
      ctx.strokeText(line, width / 2, yPos);

      if (i === lines.length - 1) {
        const tw = ctx.measureText(line).width;
        ctx.fillStyle = isVariantB ? 'rgba(255, 230, 0, 0.16)' : 'rgba(255, 42, 85, 0.20)';
        rrect(ctx, (width - tw - 36) / 2, yPos + 4, tw + 36, fontSize + 10, 12);
        ctx.fill();
      }
      ctx.fillStyle = color;
      ctx.fillText(line, width / 2, yPos);
    });

    const lastLineY = startY + lines.length * lineH + 8;
    const sepCol = isVariantB ? '#FFE600' : '#00FF66';
    ctx.strokeStyle = sepCol;
    ctx.lineWidth = 5;
    ctx.shadowColor = sepCol;
    ctx.shadowBlur = 14;
    ctx.beginPath(); ctx.moveTo(86, lastLineY); ctx.lineTo(width - 86, lastLineY); ctx.stroke();
    ctx.restore();

    // Center Confrontation: Vector Shocked Stickman vs Hooded Hacker
    drawShockedStickman(ctx, 24, 435, 340, 445, isVariantB ? '#00FF66' : '#00F0FF');
    drawHoodedHacker(ctx, 355, 415, 345, 465);

    if (propImg) {
      ctx.save();
      ctx.shadowColor = 'rgba(0, 240, 255, 0.6)';
      ctx.shadowBlur = 20;
      const pw = 135;
      ctx.drawImage(propImg, (width - pw) / 2, 430, pw, pw * propImg.height / propImg.width);
      ctx.restore();
    }

    // Threat Arrow pointing from Hacker to Victim
    ctx.save();
    ctx.strokeStyle = '#FF2A55';
    ctx.lineWidth = 6.5;
    ctx.lineCap = 'round';
    ctx.shadowColor = '#FF2A55';
    ctx.shadowBlur = 16;
    ctx.beginPath();
    ctx.moveTo(410, 620); ctx.lineTo(310, 660);
    ctx.moveTo(310, 660); ctx.lineTo(338, 642);
    ctx.moveTo(310, 660); ctx.lineTo(334, 678);
    ctx.stroke();
    ctx.restore();

    // Focal Cyber Evidence Status Card
    drawFocalThreatCard(ctx, 48, 920, width - 96, 134, lowerTitle, opts.variant);

    // Bottom Neon CTA Plate
    ctx.save();
    const ctaY = 1105;
    const ctaH = 78;
    const ctaW = width - 80;
    const ctaX = 40;
    const ctaGrad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
    ctaGrad.addColorStop(0, isVariantB ? '#FFE600' : '#00FF66');
    ctaGrad.addColorStop(0.5, '#00F0FF');
    ctaGrad.addColorStop(1, isVariantB ? '#00FF66' : '#00FF88');
    ctx.fillStyle = ctaGrad;
    ctx.shadowColor = isVariantB ? 'rgba(255, 230, 0, 0.5)' : 'rgba(0, 255, 102, 0.55)';
    ctx.shadowBlur = 28;
    rrect(ctx, ctaX, ctaY, ctaW, ctaH, 39);
    ctx.fill();

    ctx.shadowBlur = 0;
    ctx.fillStyle = '#04070D';
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
  fs.writeFileSync(opts.out, buf);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
