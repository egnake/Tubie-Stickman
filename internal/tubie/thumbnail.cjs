const fs = require('fs');
const path = require('path');

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {
    title: 'SİBER GÜVENLİK UYARISI',
    topic: '',
    prop: '',
    pose: 2,
    out: ''
  };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--title' && args[i + 1]) out.title = args[++i];
    else if (args[i] === '--topic' && args[i + 1]) out.topic = args[++i];
    else if (args[i] === '--prop' && args[i + 1]) out.prop = args[++i];
    else if (args[i] === '--pose' && args[i + 1]) out.pose = parseInt(args[++i], 10) || 0;
    else if (args[i] === '--out' && args[i + 1]) out.out = args[++i];
  }
  return out;
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

  const width = 720;
  const height = 1280;
  const canvas = createCanvas(width, height);
  const ctx = canvas.getContext('2d');

  // Register fonts if present
  const notoSans = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/NotoSans.ttf');
  const caveat = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/Caveat.ttf');
  if (fs.existsSync(notoSans)) GlobalFonts.registerFromPath(notoSans, 'NotoSans');
  if (fs.existsSync(caveat)) GlobalFonts.registerFromPath(caveat, 'Caveat');

  // 1. Dark Cyber Gradient Background
  const bgGrad = ctx.createLinearGradient(0, 0, 0, height);
  bgGrad.addColorStop(0, '#0B0F19');
  bgGrad.addColorStop(0.35, '#070A12');
  bgGrad.addColorStop(1, '#020407');
  ctx.fillStyle = bgGrad;
  ctx.fillRect(0, 0, width, height);

  // 2. High-Tech Cyber Grid
  ctx.save();
  ctx.strokeStyle = 'rgba(0, 240, 255, 0.08)';
  ctx.lineWidth = 1.5;
  for (let x = 40; x < width; x += 60) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }
  for (let y = 40; y < height; y += 60) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }
  ctx.restore();

  // 3. Central Neon Radial Aura
  ctx.save();
  const radGrad = ctx.createRadialGradient(width / 2, 730, 60, width / 2, 730, 440);
  radGrad.addColorStop(0, 'rgba(0, 240, 255, 0.22)');
  radGrad.addColorStop(0.55, 'rgba(255, 0, 85, 0.13)');
  radGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = radGrad;
  ctx.fillRect(0, 360, width, 750);
  ctx.restore();

  // 4. Futuristic Cyber Corner Brackets
  ctx.save();
  ctx.strokeStyle = '#00F0FF';
  ctx.lineWidth = 4;
  const bSize = 38;
  const pad = 24;
  // Top Left
  ctx.beginPath(); ctx.moveTo(pad, pad + bSize); ctx.lineTo(pad, pad); ctx.lineTo(pad + bSize, pad); ctx.stroke();
  // Top Right
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, pad); ctx.lineTo(width - pad, pad); ctx.lineTo(width - pad, pad + bSize); ctx.stroke();
  // Bottom Left
  ctx.beginPath(); ctx.moveTo(pad, height - pad - bSize); ctx.lineTo(pad, height - pad); ctx.lineTo(pad + bSize, height - pad); ctx.stroke();
  // Bottom Right
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, height - pad); ctx.lineTo(width - pad, height - pad); ctx.lineTo(width - pad, height - pad - bSize); ctx.stroke();
  ctx.restore();

  // 5. Contextual Top Hook Badge
  let badgeText = "🚨 SİBER GÜVENLİK UYARISI";
  const lowerTitle = (opts.title + " " + opts.topic).toLowerCase();
  if (lowerTitle.includes("instagram") || lowerTitle.includes("sosyal") || lowerTitle.includes("dm")) {
    badgeText = "🚨 HESAP GÜVENLİĞİ UYARISI";
  } else if (lowerTitle.includes("kripto") || lowerTitle.includes("bitcoin") || lowerTitle.includes("cüzdan")) {
    badgeText = "💰 KRİPTO DOLANDIRICILIĞI";
  } else if (lowerTitle.includes("whatsapp") || lowerTitle.includes("sms")) {
    badgeText = "💬 WHATSAPP HESAP TUZAĞI";
  } else if (lowerTitle.includes("wifi") || lowerTitle.includes("şifre")) {
    badgeText = "🔓 KRİTİK GÜVENLİK AÇIĞI";
  }

  ctx.save();
  ctx.font = '800 24px NotoSans, sans-serif';
  const badgeWidth = ctx.measureText(badgeText).width + 52;
  const badgeX = (width - badgeWidth) / 2;
  const badgeY = 88;
  const badgeH = 48;

  ctx.fillStyle = 'rgba(17, 24, 39, 0.95)';
  ctx.strokeStyle = '#FF3B30';
  ctx.lineWidth = 2.5;
  ctx.beginPath();
  ctx.roundRect(badgeX, badgeY, badgeWidth, badgeH, 24);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = '#FFE600';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(badgeText, width / 2, badgeY + badgeH / 2);
  ctx.restore();

  // 6. Massive, Clean Multi-Line Hook Headline
  const rawWords = opts.title.toUpperCase().replace(/[?!.,]/g, '').split(/\s+/).filter(Boolean);
  const lines = [];
  let cur = "";
  for (const w of rawWords) {
    if ((cur + " " + w).trim().length > 13 && cur) {
      lines.push(cur.trim());
      cur = w;
    } else {
      cur = (cur + " " + w).trim();
    }
  }
  if (cur) lines.push(cur.trim());
  const displayLines = lines.slice(0, 3);

  ctx.save();
  const startY = 185;
  const lineHeight = 72;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';

  displayLines.forEach((line, i) => {
    const yPos = startY + i * lineHeight;
    let textColor = '#FFFFFF';
    if (i === 0) textColor = '#00F0FF';
    else if (i === displayLines.length - 1) textColor = '#FF3344';

    // Drop shadow
    ctx.shadowColor = 'rgba(0, 0, 0, 0.92)';
    ctx.shadowBlur = 18;
    ctx.shadowOffsetY = 6;

    // Stroke
    ctx.font = '900 58px NotoSans, sans-serif';
    ctx.strokeStyle = '#05080E';
    ctx.lineWidth = 12;
    ctx.strokeText(line, width / 2, yPos);

    // Fill
    ctx.fillStyle = textColor;
    ctx.fillText(line, width / 2, yPos);
  });

  // Glowing neon underline
  const lastLineY = startY + displayLines.length * lineHeight;
  ctx.strokeStyle = '#00FF66';
  ctx.lineWidth = 6;
  ctx.shadowColor = '#00FF66';
  ctx.shadowBlur = 14;
  ctx.beginPath();
  ctx.moveTo(110, lastLineY + 6);
  ctx.lineTo(width - 110, lastLineY + 6);
  ctx.stroke();
  ctx.restore();

  // 7. Hero Stickman in Action
  const poseAtlasPath = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/character-pose-sheet-v6.png');
  if (fs.existsSync(poseAtlasPath)) {
    const atlas = await loadImage(poseAtlasPath);
    const sw = atlas.width / 3;
    const pIdx = Math.max(0, Math.min(2, opts.pose));

    ctx.save();
    ctx.shadowColor = 'rgba(0, 240, 255, 0.45)';
    ctx.shadowBlur = 28;

    const smW = 460;
    const smH = smW * atlas.height / sw;
    const smX = 60;
    const smY = 490;
    ctx.drawImage(atlas, pIdx * sw, 0, sw, atlas.height, smX, smY, smW, smH);
    ctx.restore();
  }

  // 8. Cyber Prop Icon
  let propFile = opts.prop;
  if (!propFile || !fs.existsSync(propFile)) {
    if (lowerTitle.includes("kripto")) {
      propFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-padlock-unlocked.png');
    } else if (lowerTitle.includes("koru") || lowerTitle.includes("2fa")) {
      propFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-shield.png');
    } else if (lowerTitle.includes("wifi")) {
      propFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-wifi.png');
    } else {
      propFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-skull.png');
    }
  }

  if (fs.existsSync(propFile)) {
    const propImg = await loadImage(propFile);
    ctx.save();
    ctx.shadowColor = 'rgba(255, 59, 48, 0.75)';
    ctx.shadowBlur = 30;

    const propW = 230;
    const propH = propW * propImg.height / propImg.width;
    const propX = 435;
    const propY = 620;
    ctx.drawImage(propImg, propX, propY, propW, propH);
    ctx.restore();
  }

  // 9. Warning Danger Arrow pointing between prop and stickman
  ctx.save();
  ctx.strokeStyle = '#FF3B30';
  ctx.lineWidth = 6;
  ctx.shadowColor = '#FF3B30';
  ctx.shadowBlur = 12;
  ctx.beginPath();
  ctx.moveTo(435, 680);
  ctx.lineTo(375, 720);
  ctx.stroke();
  ctx.beginPath();
  ctx.moveTo(375, 720);
  ctx.lineTo(395, 705);
  ctx.moveTo(375, 720);
  ctx.lineTo(395, 735);
  ctx.stroke();
  ctx.restore();

  // 10. High-Converting Call to Action Banner
  ctx.save();
  const ctaY = 1110;
  const ctaH = 70;
  const ctaW = width - 100;
  const ctaX = 50;

  const ctaGrad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
  ctaGrad.addColorStop(0, '#00FF66');
  ctaGrad.addColorStop(1, '#00F0FF');
  ctx.fillStyle = ctaGrad;
  ctx.shadowColor = 'rgba(0, 255, 102, 0.45)';
  ctx.shadowBlur = 22;
  ctx.beginPath();
  ctx.roundRect(ctaX, ctaY, ctaW, ctaH, 35);
  ctx.fill();

  ctx.shadowColor = 'transparent';
  ctx.fillStyle = '#05080E';
  ctx.font = '900 30px NotoSans, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText("⚡ 30 SANİYEDE ÖĞREN · İZLE ↗", width / 2, ctaY + ctaH / 2);
  ctx.restore();

  // Ensure output directory exists and write JPEG
  const outDir = path.dirname(opts.out);
  if (!fs.existsSync(outDir)) {
    fs.mkdirSync(outDir, { recursive: true });
  }

  const buf = canvas.toBuffer('image/jpeg', { quality: 95 });
  fs.writeFileSync(opts.out, buf);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
