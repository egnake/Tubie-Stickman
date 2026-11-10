const fs = require('fs');
const path = require('path');

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {
    title: 'SİBER GÜVENLİK UYARISI',
    topic: '',
    prop: '',
    pose: 1, // Default to shocked/panicked stickman
    out: ''
  };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--title' && args[i + 1]) out.title = args[++i];
    else if (args[i] === '--topic' && args[i + 1]) out.topic = args[++i];
    else if (args[i] === '--prop' && args[i + 1]) out.prop = args[++i];
    else if (args[i] === '--pose' && args[i + 1]) out.pose = parseInt(args[++i], 10) || 1;
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

  // Register fonts
  const notoSans = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/NotoSans.ttf');
  const caveat = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/fonts/Caveat.ttf');
  if (fs.existsSync(notoSans)) GlobalFonts.registerFromPath(notoSans, 'NotoSans');
  if (fs.existsSync(caveat)) GlobalFonts.registerFromPath(caveat, 'Caveat');

  // -------------------------------------------------------------
  // 1. Cinematic Deep Cyber Gradient Background
  // -------------------------------------------------------------
  const bgGrad = ctx.createLinearGradient(0, 0, width, height);
  bgGrad.addColorStop(0, '#090D17');
  bgGrad.addColorStop(0.45, '#04070E');
  bgGrad.addColorStop(1, '#020306');
  ctx.fillStyle = bgGrad;
  ctx.fillRect(0, 0, width, height);

  // -------------------------------------------------------------
  // 2. High-Tech Hex/Grid Lines & Glowing Horizon
  // -------------------------------------------------------------
  ctx.save();
  ctx.strokeStyle = 'rgba(0, 240, 255, 0.065)';
  ctx.lineWidth = 1.2;
  const gridSize = 50;
  for (let x = 20; x < width; x += gridSize) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }
  for (let y = 20; y < height; y += gridSize) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }

  // Crosshair dots at grid intersections
  ctx.fillStyle = 'rgba(0, 240, 255, 0.18)';
  for (let x = 20; x < width; x += gridSize * 2) {
    for (let y = 120; y < height - 120; y += gridSize * 2) {
      ctx.fillRect(x - 1.5, y - 1.5, 3, 3);
    }
  }
  ctx.restore();

  // -------------------------------------------------------------
  // 3. Dual Radial Atmospheric Glows (Cyber Cyan + Emergency Red)
  // -------------------------------------------------------------
  ctx.save();
  // Hero Cyan Aura behind stickman
  const cyanAura = ctx.createRadialGradient(230, 720, 40, 230, 720, 360);
  cyanAura.addColorStop(0, 'rgba(0, 240, 255, 0.28)');
  cyanAura.addColorStop(0.5, 'rgba(0, 255, 128, 0.12)');
  cyanAura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = cyanAura;
  ctx.fillRect(0, 420, width, 600);

  // Threat Red Aura behind threat prop
  const redAura = ctx.createRadialGradient(530, 720, 30, 530, 720, 300);
  redAura.addColorStop(0, 'rgba(255, 42, 85, 0.26)');
  redAura.addColorStop(0.6, 'rgba(255, 0, 85, 0.08)');
  redAura.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = redAura;
  ctx.fillRect(300, 450, width - 300, 550);
  ctx.restore();

  // -------------------------------------------------------------
  // 4. Subtle Scanline Texture for Terminal Aesthetic
  // -------------------------------------------------------------
  ctx.save();
  ctx.fillStyle = 'rgba(0, 240, 255, 0.015)';
  for (let y = 0; y < height; y += 4) {
    ctx.fillRect(0, y, width, 1.5);
  }
  ctx.restore();

  // -------------------------------------------------------------
  // 5. Corner Sci-Fi Brackets & Top HUD Status Bar (Zero Emojis)
  // -------------------------------------------------------------
  ctx.save();
  ctx.strokeStyle = '#00F0FF';
  ctx.lineWidth = 3.5;
  ctx.shadowColor = '#00F0FF';
  ctx.shadowBlur = 10;
  const bSize = 34;
  const pad = 24;

  // Top Left Bracket
  ctx.beginPath(); ctx.moveTo(pad, pad + bSize); ctx.lineTo(pad, pad); ctx.lineTo(pad + bSize, pad); ctx.stroke();
  // Top Right Bracket
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, pad); ctx.lineTo(width - pad, pad); ctx.lineTo(width - pad, pad + bSize); ctx.stroke();
  // Bottom Left Bracket
  ctx.beginPath(); ctx.moveTo(pad, height - pad - bSize); ctx.lineTo(pad, height - pad); ctx.lineTo(pad + bSize, height - pad); ctx.stroke();
  // Bottom Right Bracket
  ctx.beginPath(); ctx.moveTo(width - pad - bSize, height - pad); ctx.lineTo(width - pad, height - pad); ctx.lineTo(width - pad, height - pad - bSize); ctx.stroke();

  // Top HUD Bar - Clean vector text without any emojis
  ctx.shadowBlur = 0;
  ctx.font = '700 16px NotoSans, sans-serif';
  ctx.fillStyle = '#00F0FF';
  ctx.textAlign = 'left';
  ctx.fillText('[ SİBER İSTİHBARAT ]', pad + 40, pad + 18);

  // Top Right REC indicator with canvas-drawn red dot (no emoji)
  ctx.textAlign = 'right';
  ctx.fillStyle = '#FFFFFF';
  ctx.fillText('REC', width - pad - 120, pad + 18);

  // Canvas-drawn glowing red REC circle
  ctx.fillStyle = '#FF3B30';
  ctx.shadowColor = '#FF3B30';
  ctx.shadowBlur = 10;
  ctx.beginPath();
  ctx.arc(width - pad - 105, pad + 13, 5, 0, Math.PI * 2);
  ctx.fill();

  ctx.shadowBlur = 0;
  ctx.fillStyle = '#00F0FF';
  ctx.fillText('4K UHD', width - pad - 36, pad + 18);
  ctx.restore();

  // -------------------------------------------------------------
  // 6. Professional High-Impact Hook Badge (Clean Text, No Emojis)
  // -------------------------------------------------------------
  let badgeText = "SİBER GÜVENLİK UYARISI";
  const lowerTitle = (opts.title + " " + opts.topic).toLowerCase();
  if (lowerTitle.includes("instagram") || lowerTitle.includes("sosyal") || lowerTitle.includes("dm")) {
    badgeText = "HESAP GÜVENLİĞİ PROTOKOLÜ";
  } else if (lowerTitle.includes("kripto") || lowerTitle.includes("bitcoin") || lowerTitle.includes("cüzdan")) {
    badgeText = "KRİPTO GÜVENLİK UYARISI";
  } else if (lowerTitle.includes("whatsapp") || lowerTitle.includes("sms")) {
    badgeText = "WHATSAPP GÜVENLİK ANALİZİ";
  } else if (lowerTitle.includes("wifi") || lowerTitle.includes("şifre")) {
    badgeText = "KRİTİK GÜVENLİK AÇIĞI";
  }

  ctx.save();
  ctx.font = '800 22px NotoSans, sans-serif';
  const badgeWidth = ctx.measureText(badgeText).width + 56;
  const badgeX = (width - badgeWidth) / 2;
  const badgeY = 66;
  const badgeH = 46;

  // Badge background plate
  ctx.fillStyle = 'rgba(10, 16, 28, 0.94)';
  ctx.strokeStyle = '#FF3B30';
  ctx.lineWidth = 2.2;
  ctx.shadowColor = 'rgba(255, 59, 48, 0.55)';
  ctx.shadowBlur = 14;
  ctx.beginPath();
  ctx.roundRect(badgeX, badgeY, badgeWidth, badgeH, 23);
  ctx.fill();
  ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.fillStyle = '#FFE600';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(badgeText, width / 2, badgeY + badgeH / 2);
  ctx.restore();

  // -------------------------------------------------------------
  // 7. Bold High-Impact Hook Headline (3-Line Clean Staging)
  // -------------------------------------------------------------
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
  const startY = 150;
  const lineHeight = 76;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';

  displayLines.forEach((line, i) => {
    const yPos = startY + i * lineHeight;
    let textColor = '#FFFFFF';
    let fontSize = 58;

    if (i === 0) {
      textColor = '#00F0FF';
      fontSize = 62;
    } else if (i === displayLines.length - 1) {
      textColor = '#FF3344';
      fontSize = 64;
    }

    ctx.font = `900 ${fontSize}px NotoSans, sans-serif`;

    // Multi-layer drop shadow for maximum 3D pop
    ctx.shadowColor = 'rgba(0, 0, 0, 0.96)';
    ctx.shadowBlur = 24;
    ctx.shadowOffsetY = 8;

    // Dark outline stroke
    ctx.strokeStyle = '#04070D';
    ctx.lineWidth = 14;
    ctx.strokeText(line, width / 2, yPos);

    // If final line, add subtle red warning underplate
    if (i === displayLines.length - 1) {
      ctx.save();
      const txtWidth = ctx.measureText(line).width;
      ctx.fillStyle = 'rgba(255, 42, 85, 0.18)';
      ctx.beginPath();
      ctx.roundRect((width - txtWidth - 36) / 2, yPos + 6, txtWidth + 36, fontSize + 12, 12);
      ctx.fill();
      ctx.restore();
    }

    // Front fill
    ctx.fillStyle = textColor;
    ctx.fillText(line, width / 2, yPos);
  });

  // Glowing neon underline with diamond terminal endpoints
  const lastLineY = startY + displayLines.length * lineHeight;
  ctx.strokeStyle = '#00FF66';
  ctx.lineWidth = 5.5;
  ctx.shadowColor = '#00FF66';
  ctx.shadowBlur = 16;
  ctx.beginPath();
  ctx.moveTo(90, lastLineY + 12);
  ctx.lineTo(width - 90, lastLineY + 12);
  ctx.stroke();

  // Terminal Diamond Endpoints
  ctx.fillStyle = '#00FF66';
  ctx.fillRect(86, lastLineY + 8, 8, 8);
  ctx.fillRect(width - 94, lastLineY + 8, 8, 8);
  ctx.restore();

  // -------------------------------------------------------------
  // 8. Visual Scene Staging: Shocked Stickman + Clean Threat Props
  // -------------------------------------------------------------
  const poseAtlasPath = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/character-pose-sheet-v6.png');
  let atlas = null;
  if (fs.existsSync(poseAtlasPath)) {
    atlas = await loadImage(poseAtlasPath);
  }

  // Hero Stickman: Slice 1 is the shocked/panicked stickman (hands on head, open mouth)
  if (atlas) {
    const sw = atlas.width / 3;
    // Slice 1 in character-pose-sheet-v6.png is the dramatic shocked pose
    const pIdx = 1;

    ctx.save();
    ctx.shadowColor = 'rgba(0, 240, 255, 0.65)';
    ctx.shadowBlur = 32;

    const smW = 460;
    const smH = smW * atlas.height / sw;
    const smX = 35;
    const smY = 470;
    ctx.drawImage(atlas, pIdx * sw, 0, sw, atlas.height, smX, smY, smW, smH);
    ctx.restore();

    // Shock / Panic Accents around head (pure text punctuation, no emojis)
    ctx.save();
    ctx.font = '900 38px NotoSans, sans-serif';
    ctx.fillStyle = '#FFE600';
    ctx.shadowColor = '#FFE600';
    ctx.shadowBlur = 14;
    ctx.fillText('! !', smX + 110, smY + 50);
    ctx.fillText('! !', smX + smW - 140, smY + 50);

    // Speed / panic rays near head
    ctx.strokeStyle = '#FFE600';
    ctx.lineWidth = 3.5;
    ctx.beginPath();
    ctx.moveTo(smX + 80, smY + 90); ctx.lineTo(smX + 50, smY + 80);
    ctx.moveTo(smX + 90, smY + 60); ctx.lineTo(smX + 65, smY + 40);
    ctx.moveTo(smX + smW - 90, smY + 90); ctx.lineTo(smX + smW - 60, smY + 80);
    ctx.moveTo(smX + smW - 100, smY + 60); ctx.lineTo(smX + smW - 75, smY + 40);
    ctx.stroke();
    ctx.restore();
  }

  // Hardware Cyber Laptop
  const laptopFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-laptop.png');
  if (fs.existsSync(laptopFile)) {
    const laptopImg = await loadImage(laptopFile);
    ctx.save();
    ctx.shadowColor = 'rgba(0, 240, 255, 0.55)';
    ctx.shadowBlur = 25;
    const lpW = 280;
    const lpH = lpW * laptopImg.height / laptopImg.width;
    const lpX = 405;
    const lpY = 690;
    ctx.drawImage(laptopImg, lpX, lpY, lpW, lpH);
    ctx.restore();
  }

  // Floating Threat Prop (Skull / Shield / Padlock)
  let threatFile = opts.prop;
  if (!threatFile || !fs.existsSync(threatFile)) {
    if (lowerTitle.includes("kripto")) {
      threatFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-padlock-unlocked.png');
    } else if (lowerTitle.includes("koru") || lowerTitle.includes("2fa")) {
      threatFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-shield.png');
    } else if (lowerTitle.includes("wifi")) {
      threatFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-wifi.png');
    } else {
      threatFile = path.resolve(repoRoot, 'skills/directing-stickman-videos/assets/b1/cyber-skull.png');
    }
  }

  if (fs.existsSync(threatFile)) {
    const threatImg = await loadImage(threatFile);
    ctx.save();
    ctx.shadowColor = 'rgba(255, 42, 85, 0.85)';
    ctx.shadowBlur = 32;

    const trW = 190;
    const trH = trW * threatImg.height / threatImg.width;
    const trX = 465;
    const trY = 525;
    ctx.drawImage(threatImg, trX, trY, trW, trH);
    ctx.restore();
  }

  // Dynamic High-Voltage Threat Arrow pointing to victim
  ctx.save();
  ctx.strokeStyle = '#FF2A55';
  ctx.lineWidth = 6.5;
  ctx.shadowColor = '#FF2A55';
  ctx.shadowBlur = 18;
  ctx.beginPath();
  ctx.moveTo(460, 640);
  ctx.lineTo(360, 690);
  ctx.stroke();

  // Arrow Head
  ctx.beginPath();
  ctx.moveTo(360, 690);
  ctx.lineTo(385, 670);
  ctx.moveTo(360, 690);
  ctx.lineTo(385, 710);
  ctx.stroke();
  ctx.restore();

  // Warning Status Box next to laptop (Zero Emojis - Vector Warning Triangle)
  ctx.save();
  const alertBoxX = 425;
  const alertBoxY = 960;
  const alertBoxW = 250;
  const alertBoxH = 44;

  ctx.fillStyle = 'rgba(255, 42, 85, 0.16)';
  ctx.strokeStyle = '#FF3B30';
  ctx.lineWidth = 2.0;
  ctx.beginPath();
  ctx.roundRect(alertBoxX, alertBoxY, alertBoxW, alertBoxH, 10);
  ctx.fill();
  ctx.stroke();

  // Canvas-drawn warning triangle (vector)
  const triX = alertBoxX + 22;
  const triY = alertBoxY + alertBoxH / 2;
  ctx.strokeStyle = '#FFE600';
  ctx.fillStyle = '#FFE600';
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(triX, triY - 11);
  ctx.lineTo(triX - 10, triY + 9);
  ctx.lineTo(triX + 10, triY + 9);
  ctx.closePath();
  ctx.stroke();
  // Exclamation dot in triangle
  ctx.fillRect(triX - 1, triY - 5, 2, 6);
  ctx.fillRect(triX - 1, triY + 4, 2, 2);

  // Status text
  ctx.font = '800 16px NotoSans, sans-serif';
  ctx.fillStyle = '#FFE600';
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  ctx.fillText('ŞİFRE ELE GEÇİRİLDİ!', alertBoxX + 42, alertBoxY + alertBoxH / 2);
  ctx.restore();

  // -------------------------------------------------------------
  // 9. High-Converting Bottom Call-To-Action (CTA) Plate (Zero Emojis)
  // -------------------------------------------------------------
  ctx.save();
  const ctaY = 1115;
  const ctaH = 78;
  const ctaW = width - 80;
  const ctaX = 40;

  // Gradient fill
  const ctaGrad = ctx.createLinearGradient(ctaX, 0, ctaX + ctaW, 0);
  ctaGrad.addColorStop(0, '#00FF66');
  ctaGrad.addColorStop(0.5, '#00F0FF');
  ctaGrad.addColorStop(1, '#00FF88');
  ctx.fillStyle = ctaGrad;

  // Intense neon bloom
  ctx.shadowColor = 'rgba(0, 255, 102, 0.55)';
  ctx.shadowBlur = 28;
  ctx.beginPath();
  ctx.roundRect(ctaX, ctaY, ctaW, ctaH, 39);
  ctx.fill();

  // High contrast bold label
  ctx.shadowBlur = 0;
  ctx.fillStyle = '#04070D';
  ctx.font = '900 27px NotoSans, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText("30 SANİYEDE KORUNMAYI ÖĞREN", width / 2 - 18, ctaY + ctaH / 2);

  // Canvas-drawn sharp vector arrow ->
  const arrX = width / 2 + 235;
  const arrY = ctaY + ctaH / 2;
  ctx.strokeStyle = '#04070D';
  ctx.lineWidth = 4;
  ctx.lineCap = 'round';
  ctx.beginPath();
  ctx.moveTo(arrX - 18, arrY);
  ctx.lineTo(arrX, arrY);
  ctx.lineTo(arrX - 7, arrY - 7);
  ctx.moveTo(arrX, arrY);
  ctx.lineTo(arrX - 7, arrY + 7);
  ctx.stroke();
  ctx.restore();

  // -------------------------------------------------------------
  // 10. Master Output Encoding
  // -------------------------------------------------------------
  const outDir = path.dirname(opts.out);
  if (!fs.existsSync(outDir)) {
    fs.mkdirSync(outDir, { recursive: true });
  }

  // Quality 98 JPEG for pristine clarity without artifacting
  const buf = canvas.toBuffer('image/jpeg', { quality: 98 });
  fs.writeFileSync(opts.out, buf);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
