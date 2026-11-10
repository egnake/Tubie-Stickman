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
function selected() { return $('#all').checked ? presets.map(p => p.id) : [$('#presets input:checked')?.value || '1']; }
function narration() {
  $('#narration').textContent = presets.find(p => p.id === selected()[0])?.narration || '';
}
$('#all').addEventListener('change', () => { updateControls(); narration(); });
function modeHint() {
  $('#mode-hint').textContent = {
    pro: 'Türkçe anlatıcı. Ses üretimi internet gerektirir.',
    cloned: 'Ses merkezindeki referansınız kullanılır. XTTS kurulumu gerekir.',
    auto: 'Referans varsa kendi sesiniz, yoksa profesyonel ses kullanılır.'
  }[$('#mode').value];
}
$('#mode').addEventListener('change', modeHint);
function updateControls() {
  const blocked = busy || pending || disconnected || !!recorder;
  ['generate', 'mode', 'all', 'voice-file', 'doctor'].forEach(id => { $('#' + id).disabled = blocked; });
  $$('#presets input').forEach(input => { input.disabled = blocked || $('#all').checked; });
  $('#test-voice').disabled = blocked || !state?.reference;
  $('#record').disabled = busy || pending || disconnected;
  $('#cancel').hidden = !busy;
  $('#cancel').disabled = pending || state?.job.status === 'stopping' || disconnected;
  $('#approve').disabled = pending || disconnected;
}
$('#generate').addEventListener('click', () => action('job', {action: 'generate', presets: selected(), mode: $('#mode').value}));
$('#doctor').addEventListener('click', async () => { $('#log-panel').open = true; await action('job', {action: 'doctor'}); });
$('#cancel').addEventListener('click', () => action('cancel'));
$('#approve').addEventListener('click', () => action('approve'));
$('#test-voice').addEventListener('click', async () => { $('#log-panel').open = true; await action('job', {action: 'test-voice'}); });
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
    const row = element('div', undefined, 'output-row');
    const info = element('div');
    info.append(element('strong', output.name), element('small', `MP4 · ${(output.size / 1024 / 1024).toFixed(1)} MB`));
    const actions = element('div', undefined, 'output-actions');
    const play = element('button', 'İzle ↗');
    play.addEventListener('click', () => { const video = $('#output-player video'); video.src = media(output.file); $('#output-player').hidden = false; video.play().catch(() => {}); });
    actions.append(play);
    const download = element('a', 'İndir ↓'); download.href = media(output.file); download.download = output.file.split('/').pop(); actions.append(download);
    if (output.subtitle) { const sub = element('a', 'Altyazı ↓'); sub.href = media(output.subtitle); sub.download = output.subtitle.split('/').pop(); actions.append(sub); }
    row.append(info, actions); list.append(row);
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
    $('#mode').value = job.mode;
    narration(); modeHint();
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
    (job.results || []).forEach((file, index) => { const link = element('a', `Video ${index + 1} · indir ↓`); link.href = media(file); link.download = file.split('/').pop(); $('#completed').append(link); });
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
      input.addEventListener('change', narration); label.append(number, text, input); $('#presets').append(label);
    });
    narration(); poll(); setInterval(() => { if (!stopped) tick(); }, 1000);
  } catch (error) { toast('Arayüz yüklenemedi. Sayfayı yenileyin. ' + error.message); }
}
init();
