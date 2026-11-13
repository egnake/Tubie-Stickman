'use strict';
const $ = selector => document.querySelector(selector);
const $$ = selector => [...document.querySelectorAll(selector)];
const topics = ['Wi-Fi güvenliği', 'Parola güvenliği', 'SQL injection', 'Phishing'];
const descriptions = ['Halka açık ağların görünmeyen yüzü', 'Brute force ve güçlü parolalar', 'Veritabanına açılan arka kapı', 'Oltalama tuzaklarını tanı'];
let presets = [], state = null, busy = false, pending = false, disconnected = false;
let recorder = null, recordingStream = null, recordingTimer = null, recordingDeadline = 0, recordingChunks = [];
let lastVideo = '', lastFrame = '', lastOutputs = '', lastLog = '', lastReference = false, syncedJob = 0, toastTimer;
const media = name => 'media/' + name.split('/').map(encodeURIComponent).join('/');
function element(tag, text, className) {
  const node = document.createElement(tag);
  if (text !== undefined) node.textContent = text;
  if (className) node.className = className;
  return node;
}
function toast(message) {
  $('#toast').textContent = message; $('#toast').hidden = false;
  clearTimeout(toastTimer); toastTimer = setTimeout(() => { $('#toast').hidden = true; }, 10000);
}
async function api(path, options = {}) {
  const response = await fetch('api/' + path, {cache: 'no-store', ...options});
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || 'İşlem tamamlanamadı.');
  return data;
}
async function action(path, body = {}) {
  if (pending) return;
  pending = true; updateControls();
  try {
    await api(path, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)});
    await refresh();
  } catch (error) { toast(error.message); }
  finally { pending = false; updateControls(); }
}
function tab(name) {
  if (!['produce', 'voice', 'outputs', 'system'].includes(name)) name = 'produce';
  $$('.page').forEach(node => { node.hidden = node.id !== name; });
  $$('.nav').forEach(node => node.classList.toggle('active', node.dataset.tab === name));
  $('#breadcrumb').textContent = {produce: 'Üretim', voice: 'Ses', outputs: 'Çıktılar', system: 'Ortam'}[name];
  history.replaceState(null, '', '#' + name);
}
$$('.nav').forEach(button => button.addEventListener('click', () => tab(button.dataset.tab)));
$('.brand').addEventListener('click', event => { event.preventDefault(); tab('produce'); });
window.addEventListener('hashchange', () => tab(location.hash.slice(1)));
let currentTopicMode = 'ai';
let currentScenes = [];

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

function loadPresetIntoEditor() {
  const p = presets.find(item => item.id === selected()[0]);
  if (!p) return;
  $('#custom-title').value = p.title;
  $('#custom-narration').value = p.narration;
  currentScenes = p.scenes || [];
  renderSceneBadges(currentScenes);
  updateWordCount();
}

function updateWordCount() {
  const text = $('#custom-narration').value.trim();
  const words = text ? text.split(/\s+/).length : 0;
  $('#char-word-count').textContent = `${words} kelime · ~${Math.round(words / 2.6)} sn`;
}
$('#custom-narration').addEventListener('input', updateWordCount);

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
  btn.disabled = true;
  status.textContent = '✨ Yapay zeka seçili tarzda senaryo ve sahneleri hazırlıyor...';
  status.hidden = false;

  try {
    const res = await api('ai-generate', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({prompt, style})
    });
    if (res.ok && res.result) {
      $('#custom-title').value = res.result.title;
      $('#custom-narration').value = res.result.narration;
      currentScenes = res.result.scenes || [];
      renderSceneBadges(currentScenes);
      updateWordCount();
      const styleNames = {
        hybrid: 'Sinematik + Bilgi Akışı',
        cinematic: 'Sinematik',
        informative: 'Yoğun Bilgi Akışı'
      };
      const styleLabel = styleNames[res.result.style || style] || 'Hibrit';
      const engineLabel = res.result.engine ? ` · ${res.result.engine}` : '';
      status.textContent = `✨ ${styleLabel} senaryo oluşturuldu (${currentScenes.length} sahne${engineLabel}). Aşağıdan düzenleyebilir veya doğrudan üretime geçebilirsiniz.`;
    }
  } catch (err) {
    status.textContent = 'Hata: ' + err.message;
    toast(err.message);
  } finally {
    btn.disabled = false;
  }
}
$('#btn-ai-generate').addEventListener('click', () => triggerAIGenerate());
$$('.chip').forEach(chip => {
  chip.addEventListener('click', () => {
    setTopicMode('ai');
    triggerAIGenerate(chip.dataset.prompt);
  });
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
  ['generate', 'voice-style', 'bgm-choice', 'ai-style', 'all', 'voice-file', 'doctor', 'btn-ai-generate', 'custom-title', 'custom-narration', 'ai-prompt'].forEach(id => {
    const node = $('#' + id);
    if (node) node.disabled = blocked;
  });
  $$('#presets input').forEach(input => { input.disabled = blocked || $('#all').checked; });
  $$('.chip').forEach(chip => { chip.disabled = blocked; });
  $('#test-voice').disabled = blocked || !state?.reference;
  $('#record').disabled = busy || pending || disconnected;
  $('#cancel').hidden = !busy;
  $('#cancel').disabled = pending || state?.job.status === 'stopping' || disconnected;
  $('#approve').disabled = pending || disconnected;
}
$('#generate').addEventListener('click', () => {
  const voiceStyle = $('#voice-style')?.value || 'viral-male';
  const bgm = $('#bgm-choice')?.value || 'cyber';
  const mode = voiceStyle.startsWith('cloned') ? 'cloned' : 'pro';

  if (currentTopicMode === 'ai' || !$('#all').checked) {
    const title = $('#custom-title').value.trim();
    const narration = $('#custom-narration').value.trim();
    if (!narration) {
      toast('Lütfen bir seslendirme metni girin veya yapay zeka ile oluşturun.');
      return;
    }
    action('job', {
      action: 'generate',
      mode: mode,
      voice_style: voiceStyle,
      bgm: bgm,
      custom: {
        title: title || 'Siber Güvenlik Videosu',
        narration: narration,
        scenes: currentScenes
      }
    });
  } else {
    action('job', {
      action: 'generate',
      presets: selected(),
      mode: mode,
      voice_style: voiceStyle,
      bgm: bgm
    });
  }
});
$('#doctor').addEventListener('click', async () => { $('#log-panel').open = true; await action('job', {action: 'doctor'}); });
$('#cancel').addEventListener('click', () => action('cancel'));
$('#approve').addEventListener('click', () => action('approve'));
$('#test-voice').addEventListener('click', async () => {
  $('#log-panel').open = true;
  const selectedStyle = $('#voice-style')?.value || '';
  const voiceStyle = selectedStyle.startsWith('cloned') ? selectedStyle : 'cloned';
  await action('job', {action: 'test-voice', voice_style: voiceStyle});
});
async function shutdown() {
  if (busy && !window.confirm('Devam eden işlem durdurulacak. Uygulama kapatılsın mı?')) return;
  if (recorder) { recorder.onstop = null; recorder.stop(); releaseMicrophone(); }
  try {
    await api('shutdown', {method: 'POST', body: '{}'});
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
    await api('voice', {method: 'POST', headers: {'Content-Type': 'application/octet-stream'}, body: blob});
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
    recordingStream = await navigator.mediaDevices.getUserMedia({audio: true});
    recorder = new MediaRecorder(recordingStream);
    recordingChunks = [];
    recorder.ondataavailable = event => { if (event.data.size) recordingChunks.push(event.data); };
    recorder.onerror = () => { releaseMicrophone(); toast('Mikrofon kaydı tamamlanamadı.'); };
    recorder.onstop = async () => {
      const blob = new Blob(recordingChunks, {type: recorder?.mimeType || 'audio/webm'});
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
function renderOutputs(outputs) {
  const signature = JSON.stringify(outputs);
  if (signature === lastOutputs) return;
  lastOutputs = signature;
  $('#output-count').textContent = outputs.length;
  const list = $('#output-list'); list.replaceChildren();
  if (!outputs.length) { list.append(element('div', 'Henüz bir video yok. İlk üretiminiz burada yerini alacak.', 'empty-list')); return; }
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
    info.append(element('strong', output.name), element('small', `MP4 · ${(output.size / 1024 / 1024).toFixed(1)} MB`));
    row.append(info);

    const actions = element('div', undefined, 'output-actions');
    const play = element('button', 'İzle ↗');
    play.addEventListener('click', () => { const video = $('#output-player video'); video.src = media(output.file); $('#output-player').hidden = false; video.play().catch(() => {}); });
    actions.append(play);

    const download = element('a', 'Video ↓'); download.href = media(output.file); download.download = output.file.split('/').pop(); actions.append(download);
    if (output.subtitle) { const sub = element('a', 'Altyazı (.srt) ↓'); sub.href = media(output.subtitle); sub.download = output.subtitle.split('/').pop(); actions.append(sub); }
    if (output.thumbnail) { const thumbLink = element('a', 'Thumbnail (.jpg) ↓'); thumbLink.href = media(output.thumbnail); thumbLink.download = 'thumbnail.jpg'; actions.append(thumbLink); }
    if (output.description) { const descLink = element('a', 'aciklama.txt ↓'); descLink.href = media(output.description); descLink.download = 'aciklama.txt'; actions.append(descLink); }
    if (output.description_text) {
      const copyBtn = element('button', 'Açıklama & Tag Kopyala 📋');
      copyBtn.addEventListener('click', async () => {
        try {
          await navigator.clipboard.writeText(output.description_text);
          toast('✨ Instagram açıklaması, A/B/C başlıkları ve hashtagler panoya kopyalandı!');
        } catch {
          toast('Panoya kopyalanamadı.');
        }
      });
      actions.append(copyBtn);
    }
    row.append(actions);
    card.append(row);

    if (output.description_text) {
      const details = document.createElement('details');
      details.className = 'output-desc-details';
      const summary = document.createElement('summary');
      summary.textContent = 'A/B/C Başlıkları, Açıklama ve Hashtagleri Gör ↗';
      const pre = document.createElement('pre');
      pre.textContent = output.description_text;
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
    $('#all').checked = job.presets.length === 4;
    $$('#presets input').forEach(input => { input.checked = input.value === job.presets[0]; });
    if (job.voice_style && $('#voice-style')) $('#voice-style').value = job.voice_style;
    loadPresetIntoEditor(); modeHint();
  }
  $('#job-label').textContent = job.label;
  $('#activity-dot').className = 'status-dot ' + job.status;
  $('#queue-label').textContent = job.items ? `${job.item} / ${job.items} video` : '';
  $('#activity-dot').parentElement.title = job.title || '';
  $('#approval').hidden = job.status !== 'approval';
  $('#preview-badge').textContent = job.status === 'approval' ? 'ONAY BEKLİYOR' : busy ? 'CANLI' : job.status === 'complete' ? 'TAMAMLANDI' : job.status === 'failed' ? 'HATA' : 'BEKLEMEDE';
  $('#preview-badge').classList.toggle('live', busy);
  const stageIndex = {voice: 0, captions: 1, storyboard: 1, preview: 2, approval: 2, render: 3, export: 4, complete: 5}[job.stage] ?? -1;
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

      const vidLink = element('a', `Video ${index + 1} (.mp4) ↓`); vidLink.href = media(file); vidLink.download = baseName;
      const srtLink = element('a', 'Altyazı (.srt) ↓'); srtLink.href = media(folder + '/' + slug + '.srt'); srtLink.download = slug + '.srt';
      const thumbLink = element('a', 'Thumbnail (.jpg) ↓'); thumbLink.href = media(folder + '/thumbnail.jpg'); thumbLink.download = 'thumbnail.jpg';
      const descLink = element('a', 'aciklama.txt ↓'); descLink.href = media(folder + '/aciklama.txt'); descLink.download = 'aciklama.txt';

      box.append(vidLink, srtLink, thumbLink, descLink);
      $('#completed').append(box);
    });
  }
  $('#completed').hidden = !job.results?.length;
  $('#reference-badge').textContent = data.reference ? 'REFERANS HAZIR' : 'REFERANS YOK';
  if (data.reference && (!lastReference || (job.action === 'voice' && job.status === 'complete' && $('#reference-audio').dataset.job !== String(job.started)))) {
    $('#reference-audio').src = media('sesim.wav') + '?v=' + (job.ended || Date.now());
    $('#reference-audio').dataset.job = String(job.started); $('#reference-audio').hidden = false;
  }
  lastReference = data.reference;
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
let stopped = false;
async function poll() {
  if (stopped) return;
  try { await refresh(); }
  catch (error) { if (stopped) return; disconnected = true; $('#connection').textContent = 'Bağlantı kesildi · yeniden deneniyor'; updateControls(); }
  if (!stopped) setTimeout(poll, 800);
}
async function init() {
  tab(location.hash.slice(1));
  try {
    presets = await api('presets');
    presets.forEach((preset, index) => {
      const label = element('label', undefined, 'preset');
      const number = element('span', '0' + (index + 1), 'preset-number');
      const text = element('span'); text.append(element('strong', topics[index]), element('small', descriptions[index]));
      const input = element('input'); input.type = 'radio'; input.name = 'preset'; input.value = preset.id; input.checked = index === 0;
      input.addEventListener('change', loadPresetIntoEditor); label.append(number, text, input); $('#presets').append(label);
    });
    $('#ai-prompt').value = 'Sahte linkleri 5 saniyede nasıl anlarsın? Oltalama tuzakları';
    modeHint(); loadPresetIntoEditor(); poll(); setInterval(() => { if (!stopped) tick(); }, 1000);
  } catch (error) { toast('Arayüz yüklenemedi. Sayfayı yenileyin. ' + error.message); }
}
init();
