'use strict';
const $ = selector => document.querySelector(selector);
const $$ = selector => [...document.querySelectorAll(selector)];

const STYLE_LABELS = {
  hybrid: 'Hibrit Viral Akış',
  cinematic: 'Sinematik Gerilim',
  informative: 'Yoğun Bilgi Akışı',
  mythbuster: 'Efsane Yıkıcı',
  challenge: '5 Saniye Testi',
  redvsblue: 'Saldırı vs Savunma'
};

const FORMAT_LABELS = {
  shorts: { header: '720 × 1280', corner: '9 : 16' },
  'shorts-hd': { header: '1080 × 1920', corner: '9 : 16 HD' },
  youtube: { header: '1280 × 720', corner: '16 : 9' },
  'youtube-hd': { header: '1920 × 1080', corner: '16 : 9 HD' }
};

const MOOD_OPTIONS = [
  ['hook', 'Kanca (Hook)'],
  ['danger', 'Tehlike (Danger)'],
  ['breach', 'İhlal (Breach)'],
  ['twist', 'Kırılma (Twist)'],
  ['defense', 'Savunma (Defense)'],
  ['cta', 'Çağrı (CTA)']
];

const WIDGET_OPTIONS = [
  ['notification', 'Bildirim Kartı'],
  ['terminal', 'Canlı Terminal'],
  ['url_inspect', 'Link / URL Tarayıcı'],
  ['network_mitm', 'Ağ / Ortadaki Adam'],
  ['password_crack', 'Şifre Kırma Sayacı'],
  ['threat_meter', 'Tehdit Ölçer'],
  ['checklist', '3 Altın Kural'],
  ['versus', 'Hatalı vs Güvenli']
];

const CAMERA_OPTIONS = [
  ['wide', 'Geniş Plan'],
  ['punch_in', 'Hızlı Yakınlaş'],
  ['zoom_slow', 'Yavaş Zoom'],
  ['shake', 'Sarsıntı (Shake)'],
  ['glitch_cut', 'Glitch Kesme']
];

let presets = [], ideas = [], activeIdeaCategory = 'Tümü';
let state = null, busy = false, pending = false, disconnected = false, stopped = false;
let recorder = null, recordingStream = null, recordingTimer = null, recordingDeadline = 0, recordingChunks = [];
let lastVideo = '', lastFrame = '', lastOutputs = '', lastLog = '', lastReference = false, lastRefRevision = 0, syncedJob = 0;
let toastTimer, analyzeTimer;

const media = name => 'media/' + name.split('/').map(encodeURIComponent).join('/');

function element(tag, text, className) {
  const node = document.createElement(tag);
  if (text !== undefined) node.textContent = text;
  if (className) node.className = className;
  return node;
}

function toast(message) {
  $('#toast').textContent = message;
  $('#toast').hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { $('#toast').hidden = true; }, 10000);
}

async function api(path, options = {}) {
  const response = await fetch('api/' + path, { cache: 'no-store', ...options });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || 'İşlem tamamlanamadı.');
  return data;
}

async function action(path, body = {}) {
  if (pending) return;
  pending = true;
  updateControls();
  try {
    await api(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    await refresh();
  } catch (error) {
    toast(error.message);
  } finally {
    pending = false;
    updateControls();
  }
}

function tab(name) {
  if (!['produce', 'voice', 'outputs', 'system'].includes(name)) name = 'produce';
  $$('.page').forEach(node => { node.hidden = node.id !== name; });
  $$('.nav').forEach(node => node.classList.toggle('active', node.dataset.tab === name));
  $('#breadcrumb').textContent = { produce: 'Üretim', voice: 'Ses', outputs: 'Çıktılar', system: 'Ortam' }[name];
  history.replaceState(null, '', '#' + name);
}

$$('.nav').forEach(button => button.addEventListener('click', () => tab(button.dataset.tab)));
$('.brand').addEventListener('click', event => { event.preventDefault(); tab('produce'); });
window.addEventListener('hashchange', () => tab(location.hash.slice(1)));

let currentTopicMode = 'ai';
let currentScenes = [];
let scenesSource = { title: '', narration: '' };
let scenesManuallyEdited = false;

function setTopicMode(mode) {
  currentTopicMode = mode;
  $('#mode-tab-ai').classList.toggle('active', mode === 'ai');
  $('#mode-tab-ai').setAttribute('aria-selected', mode === 'ai');
  $('#mode-tab-presets').classList.toggle('active', mode === 'presets');
  $('#mode-tab-presets').setAttribute('aria-selected', mode === 'presets');
  $('#ai-panel').hidden = mode !== 'ai';
  $('#presets-panel').hidden = mode !== 'presets';
}

$('#mode-tab-ai').addEventListener('click', () => setTopicMode('ai'));
$('#mode-tab-presets').addEventListener('click', () => {
  setTopicMode('presets');
  loadPresetIntoEditor();
});

function updateFormatLabels() {
  const fmt = $('#video-format')?.value || 'shorts';
  const spec = FORMAT_LABELS[fmt] || FORMAT_LABELS.shorts;
  const headerEl = $('.format-label');
  if (headerEl) {
    headerEl.replaceChildren(
      document.createTextNode(spec.header + ' '),
      element('span', '·'),
      document.createTextNode(' 30 FPS '),
      element('span', '·'),
      document.createTextNode(' MP4')
    );
  }
  const cornerEl = $('.stage-corner');
  if (cornerEl) cornerEl.textContent = spec.corner;
}
$('#video-format')?.addEventListener('change', updateFormatLabels);

function loadPresetIntoEditor() {
  const p = presets.find(item => item.id === selected()[0]);
  if (!p) return;
  $('#custom-title').value = p.title;
  $('#custom-narration').value = p.narration;
  currentScenes = (p.scenes || []).map(sc => ({ ...sc }));
  scenesSource = { title: p.title, narration: p.narration };
  scenesManuallyEdited = false;
  renderSceneBadges(currentScenes);
  renderSceneDirector(currentScenes);
  updateWordCount();
  scheduleViralAnalyze();
}

function scenesMatchCurrentScript(title, narration) {
  if (!currentScenes || !currentScenes.length) return false;
  if (scenesManuallyEdited || !scenesSource.narration) return true;
  const norm = s => s.toLowerCase().split(/\s+/).filter(w => w.length >= 3);
  const srcWords = new Set(norm(scenesSource.narration));
  const curWords = norm(narration);
  if (!srcWords.size || !curWords.length) return false;
  let shared = 0;
  curWords.forEach(w => { if (srcWords.has(w)) shared++; });
  const overlap = shared / Math.max(curWords.length, 1);
  const titleChanged = title.trim().toLowerCase() !== (scenesSource.title || '').trim().toLowerCase();
  if (overlap < 0.25 || (titleChanged && overlap < 0.4)) return false;
  return true;
}

function updateWordCount() {
  const text = $('#custom-narration').value.trim();
  const words = text ? text.split(/\s+/).length : 0;
  $('#char-word-count').textContent = `${words} kelime · ~${Math.round(words / 2.6)} sn`;
}

function buildSelect(options, selectedVal, onChange) {
  const sel = document.createElement('select');
  sel.className = 'text-input';
  options.forEach(([val, label]) => {
    const opt = document.createElement('option');
    opt.value = val;
    opt.textContent = label;
    if (val === selectedVal) opt.selected = true;
    sel.append(opt);
  });
  sel.addEventListener('change', () => onChange(sel.value));
  return sel;
}

function renderSceneBadges(scenes) {
  const container = $('#scenes-preview-container');
  const badgesBox = $('#scenes-badges');
  badgesBox.replaceChildren();
  if (!scenes || !scenes.length) {
    container.hidden = true;
    return;
  }
  container.hidden = false;
  scenes.forEach((sc, idx) => {
    const badge = element('span', `${idx + 1}. ${sc.title || 'Sahne'}`, 'scene-badge');
    if (sc.color) badge.style.borderColor = sc.color;
    badgesBox.append(badge);
  });
}

function renderSceneDirector(scenes) {
  const box = $('#scenes-director');
  if (!box) return;
  box.replaceChildren();
  if (!scenes || !scenes.length) return;
  scenes.forEach((sc, idx) => {
    const card = element('div', undefined, 'director-card');
    const head = element('div', undefined, 'director-card-head');
    head.append(element('span', `SAHNE ${idx + 1}`), element('span', sc.mood || 'hook'));
    card.append(head);

    const titleInput = document.createElement('input');
    titleInput.type = 'text';
    titleInput.className = 'text-input';
    titleInput.value = sc.title || '';
    titleInput.placeholder = 'Sahne başlığı';
    titleInput.addEventListener('input', () => {
      sc.title = titleInput.value;
      scenesManuallyEdited = true;
      renderSceneBadges(currentScenes);
    });
    card.append(titleInput);

    const grid = element('div', undefined, 'director-grid');
    const moodSel = buildSelect(MOOD_OPTIONS, sc.mood || 'hook', val => {
      sc.mood = val;
      scenesManuallyEdited = true;
      head.lastChild.textContent = val;
    });
    const widgetSel = buildSelect(WIDGET_OPTIONS, sc.visual_type || 'notification', val => {
      sc.visual_type = val;
      scenesManuallyEdited = true;
    });
    const camSel = buildSelect(CAMERA_OPTIONS, sc.camera || 'punch_in', val => {
      sc.camera = val;
      scenesManuallyEdited = true;
    });
    const calloutIn = document.createElement('input');
    calloutIn.type = 'text';
    calloutIn.className = 'text-input';
    calloutIn.value = sc.callout || '';
    calloutIn.placeholder = 'Çağrı etiketi (ör. DİKKAT!)';
    calloutIn.addEventListener('input', () => {
      sc.callout = calloutIn.value;
      scenesManuallyEdited = true;
    });

    grid.append(moodSel, widgetSel, camSel, calloutIn);
    card.append(grid);
    box.append(card);
  });
}

$('#btn-toggle-director')?.addEventListener('click', () => {
  const box = $('#scenes-director');
  const btn = $('#btn-toggle-director');
  if (!box || !btn) return;
  const open = box.hidden;
  box.hidden = !open;
  btn.textContent = open ? 'Gizle ▴' : 'Düzenle ▾';
  if (open) renderSceneDirector(currentScenes);
});

function renderViralReport(report) {
  const card = $('#viral-analyzer-card');
  if (!card || !report) return;
  card.hidden = false;
  $('#viral-total-score').textContent = `${report.total_score}/100`;
  const pill = $('#viral-grade-pill');
  pill.textContent = report.grade || '-';
  pill.className = 'grade-pill grade-' + String(report.grade || 'c').toLowerCase();
  $('#viral-meta-line').textContent = `${report.threat_category || ''} · ~${report.estimated_seconds || 0} sn (${report.word_count || 0} kelime)${report.loop_ready ? ' · Loop Hazır' : ''}`;

  const dimsBox = $('#viral-dimensions');
  dimsBox.replaceChildren();
  (report.dimensions || []).forEach(dim => {
    const row = element('div', undefined, 'viral-dim-row');
    const top = element('div', undefined, 'viral-dim-top');
    top.append(element('span', `${dim.name} (${dim.feedback})`), element('span', `${dim.score}/${dim.max_score}`));
    const bar = element('div', undefined, 'viral-bar');
    const fill = document.createElement('span');
    fill.style.width = `${Math.max(4, Math.min(100, Math.round((dim.score / Math.max(1, dim.max_score)) * 100)))}%`;
    bar.append(fill);
    row.append(top, bar);
    dimsBox.append(row);
  });

  const tipsBox = $('#viral-tips');
  tipsBox.replaceChildren();
  (report.tips || []).forEach(tip => tipsBox.append(element('li', tip)));
}

async function runViralAnalyze() {
  const title = $('#custom-title')?.value.trim() || '';
  const narration = $('#custom-narration')?.value.trim() || '';
  if (!title && !narration) {
    $('#viral-analyzer-card').hidden = true;
    return;
  }
  const loopEnding = Boolean($('#loop-ending')?.checked);
  try {
    const res = await api('viral-analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, narration, scenes: currentScenes, loop_ending: loopEnding })
    });
    if (res.ok && res.report) renderViralReport(res.report);
  } catch {
    // Ignore transient errors while user types rapidly
  }
}

function scheduleViralAnalyze() {
  clearTimeout(analyzeTimer);
  analyzeTimer = setTimeout(runViralAnalyze, 350);
}

$('#custom-title').addEventListener('input', scheduleViralAnalyze);
$('#custom-narration').addEventListener('input', () => {
  updateWordCount();
  scheduleViralAnalyze();
});
$('#loop-ending')?.addEventListener('change', scheduleViralAnalyze);

function renderIdeasDeck() {
  const catBox = $('#ideas-categories');
  const listBox = $('#ideas-list');
  if (!catBox || !listBox) return;

  const categories = ['Tümü', ...new Set(ideas.map(i => i.category).filter(Boolean))];
  if (!categories.includes(activeIdeaCategory)) activeIdeaCategory = 'Tümü';

  catBox.replaceChildren();
  categories.forEach(cat => {
    const btn = element('button', cat, 'chip' + (cat === activeIdeaCategory ? ' active' : ''));
    btn.type = 'button';
    btn.addEventListener('click', () => {
      activeIdeaCategory = cat;
      renderIdeasDeck();
    });
    catBox.append(btn);
  });

  const filtered = activeIdeaCategory === 'Tümü' ? ideas : ideas.filter(i => i.category === activeIdeaCategory);
  $('#ideas-count').textContent = `${filtered.length} fikir`;
  listBox.replaceChildren();
  filtered.forEach(idea => {
    const card = element('button', undefined, 'idea-card');
    card.type = 'button';
    const meta = element('div', undefined, 'idea-meta');
    meta.append(element('span', idea.category), element('span', STYLE_LABELS[idea.recommended_style] || idea.recommended_style));
    card.append(
      element('strong', idea.title),
      element('small', idea.hook_preview || idea.prompt),
      meta
    );
    card.addEventListener('click', () => {
      setTopicMode('ai');
      if ($('#ai-style') && idea.recommended_style) $('#ai-style').value = idea.recommended_style;
      triggerAIGenerate(idea.prompt);
    });
    listBox.append(card);
  });
}

async function triggerAIGenerate(customPrompt) {
  const prompt = customPrompt || $('#ai-prompt').value.trim();
  if (!prompt) {
    toast('Lütfen bir konu veya fikir yazın.');
    $('#ai-prompt').focus();
    return;
  }
  $('#ai-prompt').value = prompt;
  const btn = $('#btn-ai-generate');
  const status = $('#ai-status');
  const style = $('#ai-style')?.value || 'hybrid';
  const loopEnding = Boolean($('#loop-ending')?.checked);
  btn.disabled = true;
  status.textContent = '✨ Yapay zeka seçili formatta viral kanca, sahne planı ve senaryo hazırlıyor...';
  status.hidden = false;

  try {
    const res = await api('ai-generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt, style, loop_ending: loopEnding })
    });
    if (res.ok && res.result) {
      $('#custom-title').value = res.result.title;
      $('#custom-narration').value = res.result.narration;
      currentScenes = (res.result.scenes || []).map(sc => ({ ...sc }));
      scenesSource = { title: res.result.title || '', narration: res.result.narration || '' };
      scenesManuallyEdited = false;
      renderSceneBadges(currentScenes);
      renderSceneDirector(currentScenes);
      updateWordCount();
      if (res.result.viral_score) renderViralReport(res.result.viral_score);
      else scheduleViralAnalyze();

      const styleLabel = STYLE_LABELS[res.result.style || style] || 'Hibrit';
      const engineLabel = res.result.engine ? ` · ${res.result.engine}` : '';
      status.textContent = `✨ ${styleLabel} senaryo oluşturuldu (${currentScenes.length} sahne${engineLabel}). Sahneleri ve metni doğrudan düzenleyebilirsiniz.`;
    }
  } catch (err) {
    status.textContent = 'Hata: ' + err.message;
    toast(err.message);
  } finally {
    btn.disabled = false;
  }
}
$('#btn-ai-generate').addEventListener('click', () => triggerAIGenerate());

$('#btn-thumb-preview')?.addEventListener('click', async () => {
  const btn = $('#btn-thumb-preview');
  const title = $('#custom-title').value.trim() || 'BU LİNKE TIKLARSAN NE OLUR?';
  const topic = $('#custom-narration').value.trim() || $('#ai-prompt').value.trim() || title;
  const mood = currentScenes[0]?.mood || 'danger';
  btn.disabled = true;
  btn.textContent = 'Kapaklar oluşturuluyor...';
  try {
    const res = await api('thumbnail-preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, topic, mood })
    });
    if (res.ok) {
      $('#thumb-prev-9x16').src = media(res.preview_9x16) + '?v=' + res.revision;
      $('#thumb-prev-16x9').src = media(res.preview_16x9) + '?v=' + res.revision;
      $('#thumb-preview-strip').hidden = false;
    }
  } catch (err) {
    toast(err.message);
  } finally {
    btn.disabled = false;
    btn.textContent = '🖼️ Kapak Önizle (9:16 + 16:9)';
  }
});

function selected() { return $('#all').checked ? presets.map(p => p.id) : [$('#presets input:checked')?.value || '1']; }
$('#all').addEventListener('change', () => { updateControls(); loadPresetIntoEditor(); });

function modeHint() {
  const hints = {
    'viral-male': 'Viral tempoda, broadcast filtreli Ahmet seslendirmesi (+15% hız).',
    'viral-fast': 'Çok seri ve enerjik Ahmet seslendirmesi (+22% hız).',
    'viral-female': 'Canlı ve akıcı Emel seslendirmesi (+14% hız).',
    'deep-male': 'Daha derin ve karizmatik Ahmet seslendirmesi (-4% perde).',
    'cloned': 'Ses merkezindeki sesim.wav referansı XTTS ile +15% hızlandırılıp akıcı hâle getirilir.',
    'cloned-fast': 'Ses merkezindeki sesim.wav referansı XTTS ile +25% seri tempoda hızlandırılır.'
  };
  const style = $('#voice-style')?.value || 'viral-male';
  $('#mode-hint').textContent = hints[style] || hints['viral-male'];
}
$('#voice-style')?.addEventListener('change', modeHint);

function updateControls() {
  const blocked = busy || pending || disconnected || !!recorder;
  ['generate', 'voice-style', 'bgm-choice', 'video-format', 'loop-ending', 'ai-style', 'all', 'voice-file', 'doctor', 'btn-ai-generate', 'btn-thumb-preview', 'custom-title', 'custom-narration', 'ai-prompt'].forEach(id => {
    const node = $('#' + id);
    if (node) node.disabled = blocked;
  });
  $$('#presets input').forEach(input => { input.disabled = blocked || $('#all').checked; });
  $$('.chip, .idea-card').forEach(chip => { chip.disabled = blocked; });
  $('#test-voice').disabled = blocked || !state?.reference;
  $('#record').disabled = busy || pending || disconnected;
  $('#cancel').hidden = !busy;
  $('#cancel').disabled = pending || state?.job.status === 'stopping' || disconnected;
  $('#approve').disabled = pending || disconnected;
}

$('#generate').addEventListener('click', () => {
  const voiceStyle = $('#voice-style')?.value || 'viral-male';
  const bgm = $('#bgm-choice')?.value || 'cyber';
  const format = $('#video-format')?.value || 'shorts';
  const loopEnding = Boolean($('#loop-ending')?.checked);
  const mode = voiceStyle.startsWith('cloned') ? 'cloned' : 'pro';

  if (currentTopicMode === 'ai' || !$('#all').checked) {
    const title = $('#custom-title').value.trim();
    const narration = $('#custom-narration').value.trim();
    if (!narration) {
      toast('Lütfen bir seslendirme metni girin veya yapay zeka ile oluşturun.');
      return;
    }
    const scenesToSend = scenesMatchCurrentScript(title, narration) ? currentScenes : [];
    action('job', {
      action: 'generate',
      mode,
      voice_style: voiceStyle,
      bgm,
      format,
      loop_ending: loopEnding,
      custom: {
        title: title || 'Siber Güvenlik Videosu',
        narration,
        scenes: scenesToSend
      }
    });
  } else {
    action('job', {
      action: 'generate',
      presets: selected(),
      mode,
      voice_style: voiceStyle,
      bgm,
      format,
      loop_ending: loopEnding
    });
  }
});

$('#doctor').addEventListener('click', async () => { $('#log-panel').open = true; await action('job', { action: 'doctor' }); });
$('#cancel').addEventListener('click', () => action('cancel'));
$('#approve').addEventListener('click', () => action('approve'));
$('#test-voice').addEventListener('click', async () => {
  $('#log-panel').open = true;
  const selectedStyle = $('#voice-style')?.value || '';
  const voiceStyle = selectedStyle.startsWith('cloned') ? selectedStyle : 'cloned';
  await action('job', { action: 'test-voice', voice_style: voiceStyle });
});

async function shutdown() {
  if (busy && !window.confirm('Devam eden işlem durdurulacak. Uygulama kapatılsın mı?')) return;
  if (recorder) { recorder.onstop = null; recorder.stop(); releaseMicrophone(); }
  try {
    await api('shutdown', { method: 'POST', body: '{}' });
    stopped = true;
    $('#connection').textContent = 'Uygulama kapatıldı';
    document.body.replaceChildren(element('p', 'Tubie Studio kapatıldı. Bu sekmeyi kapatabilirsiniz.', 'empty-list'));
  } catch (error) { toast(error.message); }
}
$('#shutdown').addEventListener('click', shutdown);
$('#shutdown-footer').addEventListener('click', shutdown);

async function upload(blob, name) {
  if (busy || pending || disconnected) return;
  pending = true; updateControls();
  $('#upload-label').textContent = name + ' · hazırlanıyor…';
  try {
    if (blob.size > 50 * 1024 * 1024) throw new Error('Ses dosyası en fazla 50 MB olabilir.');
    await api('voice', { method: 'POST', headers: { 'Content-Type': 'application/octet-stream' }, body: blob });
    $('#upload-label').textContent = 'Referans ses hazır (sesim.wav). Üretim ekranından “Kendi sesim” seçebilirsiniz.';
    $('#log-panel').open = true;
    await refresh();
  } catch (error) { toast(error.message); $('#upload-label').textContent = ''; }
  finally { pending = false; $('#voice-file').value = ''; updateControls(); }
}
$('#voice-file').addEventListener('change', event => {
  const file = event.target.files[0]; if (file) upload(file, file.name);
});

function releaseMicrophone() {
  clearTimeout(recordingTimer);
  recordingStream?.getTracks().forEach(track => track.stop());
  recorder = null; recordingStream = null;
  $('#record').classList.remove('recording');
  $('#record').replaceChildren(element('span', undefined, 'record-dot'), document.createTextNode('15 saniyelik kayıt başlat'));
  updateControls();
}

$('#record').addEventListener('click', async () => {
  if (recorder) { recorder.stop(); return; }
  if (!navigator.mediaDevices?.getUserMedia || !window.MediaRecorder) {
    toast('Bu tarayıcı mikrofon kaydını desteklemiyor. Ses dosyası yükleyebilirsiniz.'); return;
  }
  pending = true; updateControls();
  try {
    recordingStream = await navigator.mediaDevices.getUserMedia({ audio: true });
    recorder = new MediaRecorder(recordingStream);
    recordingChunks = [];
    recorder.ondataavailable = event => { if (event.data.size) recordingChunks.push(event.data); };
    recorder.onerror = () => { releaseMicrophone(); toast('Mikrofon kaydı tamamlanamadı.'); };
    recorder.onstop = async () => {
      const blob = new Blob(recordingChunks, { type: recorder?.mimeType || 'audio/webm' });
      releaseMicrophone();
      $('#record-status').textContent = 'Kayıt alındı. Referans sese dönüştürülüyor.';
      await upload(blob, 'Mikrofon kaydı');
    };
    recorder.start(); recordingDeadline = Date.now() + 15000;
    recordingTimer = setTimeout(() => { if (recorder?.state === 'recording') recorder.stop(); }, 15000);
    $('#record').classList.add('recording');
    $('#record').replaceChildren(element('span', undefined, 'record-dot'), document.createTextNode('Kaydı bitir ve kullan'));
  } catch (error) {
    releaseMicrophone();
    toast('Mikrofon açılamadı. Tarayıcı iznini ve giriş aygıtını kontrol edin. ' + error.message);
  } finally { pending = false; updateControls(); }
});

function setVideo(file) {
  const version = media(file) + '?v=' + state.job.started;
  if (version !== lastVideo) {
    $('#preview-video').pause(); $('#preview-video').src = version; lastVideo = version;
  }
  $('#preview-video').hidden = false; $('#live-frame').hidden = true; $('#preview-empty').hidden = true;
}

function extractPinnedComment(descText) {
  if (!descText) return '';
  const lines = String(descText).replace(/\r\n/g, '\n').split('\n');
  let idx = lines.findIndex(line => line.includes('SABİTLENECEK YORUM KANCASI') || line.includes('[SABİT YORUM'));
  if (idx === -1) return '';
  idx++;
  if (idx < lines.length && /^-+$/.test(lines[idx].trim())) idx++;
  const collected = [];
  while (idx < lines.length) {
    const trimmed = lines[idx].trim();
    if (!trimmed) {
      if (collected.length) break;
      idx++;
      continue;
    }
    if (/^(?:\[|📝|🚀|🏷️|📌|📊)/.test(trimmed)) break;
    collected.push(trimmed);
    idx++;
  }
  return collected.join('\n').trim();
}

function makeDownloadLink(label, mediaPath, fileName) {
  const a = element('a', label);
  a.href = media(mediaPath);
  a.download = fileName || mediaPath.split('/').pop();
  return a;
}

function renderOutputs(outputs) {
  const signature = JSON.stringify(outputs);
  if (signature === lastOutputs) return;
  lastOutputs = signature;
  $('#output-count').textContent = outputs.length;
  const list = $('#output-list');
  list.replaceChildren();
  if (!outputs.length) {
    list.append(element('div', 'Henüz bir video yok. İlk üretiminiz burada yerini alacak.', 'empty-list'));
    return;
  }
  outputs.forEach(output => {
    const card = element('div', undefined, 'output-card');
    const row = element('div', undefined, 'output-row');

    if (output.thumbnail) {
      const thumb = document.createElement('img');
      thumb.src = media(output.thumbnail);
      thumb.className = 'output-thumb';
      thumb.alt = output.name;
      row.append(thumb);
    }

    const info = element('div', undefined, 'output-info');
    const titleWrap = element('strong', output.name);
    if (output.viral_score !== undefined) {
      titleWrap.append(element('span', `${output.viral_score}/100 · ${output.viral_grade || ''}`, 'output-score-pill'));
    }
    const fmtText = output.format ? ` · ${output.format}` : '';
    info.append(titleWrap, element('small', `MP4 · ${(output.size / 1024 / 1024).toFixed(1)} MB${fmtText}`));
    row.append(info);

    const actions = element('div', undefined, 'output-actions');
    const play = element('button', 'İzle ↗');
    play.addEventListener('click', () => {
      const video = $('#output-player video');
      video.src = media(output.file);
      $('#output-player').hidden = false;
      video.play().catch(() => { });
    });
    actions.append(play);

    actions.append(makeDownloadLink('Video ↓', output.file));
    if (output.subtitle) actions.append(makeDownloadLink('SRT ↓', output.subtitle));
    if (output.vtt) actions.append(makeDownloadLink('VTT ↓', output.vtt));
    if (output.thumbnail) actions.append(makeDownloadLink('Kapak A ↓', output.thumbnail, 'thumbnail.jpg'));
    if (output.thumbnail_b) actions.append(makeDownloadLink('Kapak B ↓', output.thumbnail_b, 'thumbnail_b.jpg'));
    if (output.thumbnail_16x9) actions.append(makeDownloadLink('16:9 ↓', output.thumbnail_16x9, 'thumbnail_16x9.jpg'));
    if (output.description) actions.append(makeDownloadLink('aciklama.txt ↓', output.description, 'aciklama.txt'));

    if (output.description_text) {
      const copyBtn = element('button', 'Açıklama Kopyala 📋');
      copyBtn.addEventListener('click', async () => {
        try {
          await navigator.clipboard.writeText(output.description_text);
          toast('✨ A/B/C/D/E başlıkları, açıklama ve hashtagler panoya kopyalandı!');
        } catch { toast('Panoya kopyalanamadı.'); }
      });
      actions.append(copyBtn);

      const pinned = extractPinnedComment(output.description_text);
      if (pinned) {
        const pinBtn = element('button', 'Sabit Yorum Kopyala 📌');
        pinBtn.addEventListener('click', async () => {
          try {
            await navigator.clipboard.writeText(pinned);
            toast('📌 Sabit yorum metni panoya kopyalandı!');
          } catch { toast('Panoya kopyalanamadı.'); }
        });
        actions.append(pinBtn);
      }
    }
    row.append(actions);
    card.append(row);

    if (output.description_text) {
      const details = document.createElement('details');
      details.className = 'output-desc-details';
      const summary = element('summary', 'A/B/C/D/E Başlıkları, Sabit Yorum ve Hashtagleri Gör ↗');
      const pre = element('pre', output.description_text);
      details.append(summary, pre);
      card.append(details);
    }

    list.append(card);
  });
}

function render(data) {
  state = data;
  const job = data.job;
  busy = ['running', 'approval', 'stopping'].includes(job.status);
  if (busy && job.action === 'generate' && job.presets?.length && syncedJob !== job.started) {
    syncedJob = job.started;
    $('#all').checked = presets.length > 0 && job.presets.length === presets.length;
    $$('#presets input').forEach(input => { input.checked = input.value === job.presets[0]; });
    if (job.voice_style && $('#voice-style')) $('#voice-style').value = job.voice_style;
    if (job.format && $('#video-format')) { $('#video-format').value = job.format; updateFormatLabels(); }
    loadPresetIntoEditor(); modeHint();
  }
  $('#job-label').textContent = job.label;
  $('#activity-dot').className = 'status-dot ' + job.status;
  $('#queue-label').textContent = job.items ? `${job.item} / ${job.items} video` : '';
  $('#activity-dot').parentElement.title = job.title || '';
  $('#approval').hidden = job.status !== 'approval';
  $('#preview-badge').textContent = job.status === 'approval' ? 'ONAY BEKLİYOR' : busy ? 'CANLI' : job.status === 'complete' ? 'TAMAMLANDI' : job.status === 'failed' ? 'HATA' : 'BEKLEMEDE';
  $('#preview-badge').classList.toggle('live', busy);
  const stageIndex = { voice: 0, captions: 1, storyboard: 1, preview: 2, approval: 2, render: 3, export: 4, complete: 5 }[job.stage] ?? -1;
  $$('.steps [data-stage]').forEach((step, index) => {
    step.classList.toggle('done', index < stageIndex);
    step.classList.toggle('current', index === stageIndex);
  });
  if (busy && !job.frame && job.status !== 'approval') $('#render-progress').removeAttribute('value');
  else $('#render-progress').value = ['complete', 'approval'].includes(job.status) ? 100 : job.frames ? job.frame / job.frames * 100 : 0;
  const result = job.results?.at(-1);
  if (job.status === 'approval' && job.preview) setVideo(job.preview);
  else if (job.status === 'complete' && result) setVideo(result);
  else if (job.live && job.live_revision) {
    const src = media(job.live) + '?v=' + job.live_revision;
    if (src !== lastFrame) { $('#live-frame').src = src; lastFrame = src; }
    $('#preview-video').pause(); $('#preview-video').hidden = true; $('#live-frame').hidden = false; $('#preview-empty').hidden = true;
  } else if (busy && job.action === 'generate') {
    $('#preview-video').pause(); $('#preview-video').hidden = true; $('#live-frame').hidden = true; $('#preview-empty').hidden = false;
  }
  $('#frame-label').textContent = job.frame ? `${job.kind === 'preview' ? 'Önizleme' : 'Render'} · ${job.frame} / ${job.frames} kare · ${Number(job.seconds).toFixed(1)} sn` : job.title || (busy ? job.label : 'Henüz bir üretim yok');
  $('#job-error').hidden = !job.error; $('#job-error').textContent = job.error || '';
  if (job.error) $('#log-panel').open = true;

  const signature = JSON.stringify(job.results || []);
  if ($('#completed').dataset.signature !== signature) {
    $('#completed').dataset.signature = signature; $('#completed').replaceChildren();
    (job.results || []).forEach((file, index) => {
      const box = element('div', undefined, 'completed-item');
      const baseName = file.split('/').pop();
      const folder = file.includes('/') ? file.substring(0, file.lastIndexOf('/')) : 'output';
      const slug = baseName.replace('.mp4', '');

      box.append(
        makeDownloadLink(`Video ${index + 1} (.mp4) ↓`, file, baseName),
        makeDownloadLink('SRT ↓', folder + '/' + slug + '.srt', slug + '.srt'),
        makeDownloadLink('VTT ↓', folder + '/' + slug + '.vtt', slug + '.vtt'),
        makeDownloadLink('Kapak A ↓', folder + '/thumbnail.jpg', 'thumbnail.jpg'),
        makeDownloadLink('Kapak B ↓', folder + '/thumbnail_b.jpg', 'thumbnail_b.jpg'),
        makeDownloadLink('16:9 ↓', folder + '/thumbnail_16x9.jpg', 'thumbnail_16x9.jpg'),
        makeDownloadLink('aciklama.txt ↓', folder + '/aciklama.txt', 'aciklama.txt')
      );
      $('#completed').append(box);
    });
  }
  $('#completed').hidden = !job.results?.length;
  $('#reference-badge').textContent = data.reference ? 'REFERANS HAZIR' : 'REFERANS YOK';
  const refRev = data.reference_revision || 0;
  if (data.reference && (!lastReference || (refRev && refRev !== lastRefRevision) || (job.action === 'voice' && job.status === 'complete' && $('#reference-audio').dataset.job !== String(job.started)))) {
    $('#reference-audio').src = media('sesim.wav') + '?v=' + (refRev || job.ended || Date.now());
    $('#reference-audio').dataset.job = String(job.started || refRev); $('#reference-audio').hidden = false;
  }
  lastReference = data.reference;
  lastRefRevision = refRev;
  if (job.action === 'voice' && job.status === 'complete') $('#upload-label').textContent = 'Referans ses hazır. Üretim ekranından “Kendi sesim” seçebilirsiniz.';
  if (job.audio) {
    const src = media(job.audio) + '?v=' + job.started;
    if ($('#test-audio').getAttribute('src') !== src) $('#test-audio').src = src;
    $('#test-audio').hidden = false;
  }
  renderOutputs(data.outputs);
  const logSignature = (data.logs.at(-1)?.id || '') + ':' + data.logs.length;
  if (logSignature !== lastLog) {
    const logs = $('#logs'); const follow = logs.scrollTop + logs.clientHeight >= logs.scrollHeight - 40;
    logs.textContent = data.logs.map(line => `${line.time}  ${line.text}`).join('\n') || 'Üretim başladığında ayrıntılar burada görünecek.';
    if (follow) logs.scrollTop = logs.scrollHeight;
    $('#log-count').textContent = data.logs.length; lastLog = logSignature;
  }
  updateControls(); tick();
}

function tick() {
  const job = state?.job;
  const seconds = job?.started ? Math.max(0, Math.floor(((job.ended || Date.now() / 1000) - job.started))) : 0;
  $('#elapsed').textContent = `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`;
  if (recorder) $('#record-status').textContent = `Kayıt alınıyor · ${Math.max(0, Math.ceil((recordingDeadline - Date.now()) / 1000))} saniye kaldı`;
}

async function refresh() {
  const data = await api('state');
  if (stopped) return;
  disconnected = false;
  $('#connection').replaceChildren(element('i', undefined, 'local-dot'), document.createTextNode(' Yerel bağlantı'));
  render(data);
}

async function poll() {
  if (stopped) return;
  try { await refresh(); }
  catch { if (stopped) return; disconnected = true; $('#connection').textContent = 'Bağlantı kesildi · yeniden deneniyor'; updateControls(); }
  if (!stopped) setTimeout(poll, 800);
}

async function init() {
  tab(location.hash.slice(1));
  updateFormatLabels();
  try {
    const [loadedPresets, ideasRes] = await Promise.all([
      api('presets'),
      api('ideas').catch(() => ({ ideas: [] }))
    ]);
    presets = loadedPresets || [];
    presets.forEach((preset, index) => {
      const label = element('label', undefined, 'preset');
      const number = element('span', String(index + 1).padStart(2, '0'), 'preset-number');
      const text = element('span');
      const sub = [preset.category, preset.summary || preset.hook_preview].filter(Boolean).join(' · ') || 'Siber güvenlik senaryosu';
      text.append(element('strong', preset.title), element('small', sub));
      const input = document.createElement('input');
      input.type = 'radio';
      input.name = 'preset';
      input.value = preset.id;
      input.checked = index === 0;
      input.addEventListener('change', loadPresetIntoEditor);
      label.append(number, text, input);
      $('#presets').append(label);
    });

    ideas = ideasRes?.ideas || [];
    renderIdeasDeck();

    $('#ai-prompt').value = 'Sahte linkleri 5 saniyede nasıl anlarsın? Oltalama tuzakları';
    modeHint(); loadPresetIntoEditor(); poll();
    setInterval(() => { if (!stopped) tick(); }, 1000);
  } catch (error) {
    toast('Arayüz yüklenemedi. Sayfayı yenileyin. ' + error.message);
  }
}
init();
