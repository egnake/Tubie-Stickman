package tubie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const g4fRunnerTimeout = 65 * time.Second

// g4fBridgeScript queries free cloud LLM providers via Python's g4f library.
// Input and output use UTF-8/ASCII-safe JSON over stdin/stdout so Windows codepages
// never corrupt Turkish characters and third-party prints cannot pollute stdout.
const g4fBridgeScript = `import asyncio, contextlib, importlib, io, json, os, re, sys, urllib.parse, warnings

warnings.filterwarnings("ignore")

def strip_json_fences(raw):
    text = (raw or "").strip()
    if text.startswith("` + "```" + `"):
        nl = text.find("\n")
        if nl != -1:
            text = text[nl + 1:]
        end_fence = text.rfind("` + "```" + `")
        if end_fence != -1:
            text = text[:end_fence]
    start = text.find("{")
    end = text.rfind("}")
    if start != -1 and end > start:
        return text[start:end + 1]
    return text.strip()

REFUSAL_MARKERS = (
    "copilot moved to a new app",
    "no cake credits",
    "bake proof-of-work",
    "sign up at g4f.dev",
    "kaydolun ve isteğinizi",
    "bu görevin koşullarını yerine getiremiyorum",
    "özür dilerim, ancak",
    "yapay zeka modeliyim",
    "verify you are human",
    "turnstile_session_token",
    "missing authorization",
    "invalid api key",
)

PLACEHOLDER_MARKERS = (
    "[isim]",
    "[adınız]",
    "[konu]",
    "sahne 1 başlığı",
    "siber güvenlik senaryosu",
)

def has_bad_markers(text):
    low = (text or "").lower()
    if any(m in low for m in REFUSAL_MARKERS):
        return True
    if any(m in low for m in PLACEHOLDER_MARKERS):
        return True
    return False

META_LABEL_RE = re.compile(r"(?i)(?:^|[\s.!?…])(?:efsane|kanca|köprü|final|giriş|açılış|senaryo|açıklama|kapanış|sahne\s*\d+)\s*:")
FOREIGN_SCRIPT_RE = re.compile(r"[\u0400-\u04FF\u4E00-\u9FFF\u0600-\u06FF\u0900-\u097F]")
SPACED_PERCENT_RE = re.compile(r"\b\d+\s+%")

def has_duplicate_sentence(text):
    parts = re.split(r"[.!?…]+", turkish_lower(text))
    seen = set()
    for part in parts:
        words = re.findall(r"[a-zçğıöşü0-9]+", part)
        if len(words) < 5:
            continue
        key = " ".join(words)
        if key in seen:
            return True
        seen.add(key)
    return False

def turkish_lower(s):
    return (s or "").replace("I", "ı").replace("İ", "i").lower()

def extract_topic_tokens(text):
    stop = {
        "nasıl", "neden", "nedir", "neler", "hakkında", "konusunda", "ile", "veya",
        "için", "gibi", "daha", "kadar", "sizi", "seni", "sizin", "senin", "beni",
        "bizi", "olan", "olur", "olabilir", "siber", "güvenlik", "güvenliği",
        "tuzağı", "tuzakları", "yöntemleri", "korunma", "rehberi", "video",
        "dikkat", "kritik", "önemli", "yeni", "neleri", "hangi", "bunu",
    }
    short_tech = {"vpn", "2fa", "xss", "sql", "qr", "usb", "iot", "ai", "llm", "api", "ssh", "dns", "sms", "otp", "pin"}
    words = re.findall(r"[a-zA-ZçğıöşüÇĞİÖŞÜ0-9]+", turkish_lower(text))
    return [w for w in words if (len(w) >= 4 or w in short_tech) and w not in stop]

def check_relevance(reference_text, candidate_text):
    tokens = extract_topic_tokens(reference_text)
    if not tokens:
        return True
    cand = turkish_lower(candidate_text)
    synonyms = {
        "2fa": ("iki aşamalı", "doğrulama", "authenticator", "2fa"),
        "ai": ("yapay zeka", "model", "ai"),
        "vpn": ("vpn", "şifreli tünel", "ortak ağ"),
        "qr": ("qr", "karekod", "quishing"),
        "otp": ("tek kullanımlık", "onay kodu", "otp"),
    }
    for tok in tokens:
        if tok in synonyms:
            if any(syn in cand for syn in synonyms[tok]):
                return True
        stem = tok[:5] if len(tok) > 5 else tok
        if stem in cand:
            return True
    return False

def validate_payload(task, data, req):
    if not isinstance(data, dict):
        return False, "not a json object"
    raw_dump = json.dumps(data, ensure_ascii=False)
    if has_bad_markers(raw_dump):
        return False, "refusal or placeholder detected"

    if task == "script":
        title = str(data.get("title") or "").strip()
        narration = str(data.get("narration") or "").strip()
        scenes = data.get("scenes")
        if not title or len(title) < 6 or turkish_lower(title) == "siber güvenlik senaryosu":
            return False, "invalid title"
        if FOREIGN_SCRIPT_RE.search(title + " " + narration):
            return False, "foreign script detected"
        if META_LABEL_RE.search(narration) or "senaryo yazıyorum" in turkish_lower(narration):
            return False, "meta label leak in narration"
        if SPACED_PERCENT_RE.search(narration):
            return False, "malformed spaced percentage"
        if has_duplicate_sentence(narration):
            return False, "duplicate sentence in narration"
        words = narration.split()
        if len(words) < 75:
            return False, f"narration too short ({len(words)} words)"
        if "**" in narration or "##" in narration:
            return False, "markdown inside narration"
        if not isinstance(scenes, list) or len(scenes) < 4:
            return False, "insufficient scenes"
        valid_scenes = 0
        for sc in scenes:
            if isinstance(sc, dict) and str(sc.get("title") or "").strip():
                valid_scenes += 1
        if valid_scenes < 4:
            return False, "incomplete scenes"
        if not check_relevance(req.get("topic", ""), title + " " + narration):
            return False, "off-topic script"
        return True, ""

    if task == "viral":
        ta = str(data.get("title_a") or "").strip()
        tb = str(data.get("title_b") or "").strip()
        tc = str(data.get("title_c") or "").strip()
        desc = str(data.get("description") or "").strip()
        tags = data.get("hashtags")
        if not (ta and tb and tc):
            return False, "missing A/B/C titles"
        if len(desc.split()) < 45:
            return False, f"description too short ({len(desc.split())} words)"
        if not isinstance(tags, list):
            return False, "hashtags not list"
        clean_tags = []
        for t in tags:
            ts = str(t or "").strip()
            if "%" in ts:
                ts = urllib.parse.unquote(ts)
            if ts and not ts.startswith("#"):
                ts = "#" + ts.lstrip("#")
            if re.match(r"^#[0-9a-zA-ZçğıöşüÇĞİÖŞÜ_]+$", ts):
                clean_tags.append(ts)
        if len(clean_tags) < 6:
            return False, f"insufficient valid hashtags ({len(clean_tags)})"
        data["hashtags"] = clean_tags
        ref = req.get("title", "") + " " + req.get("narration", "")[:220]
        combined = " ".join([ta, tb, tc, desc] + clean_tags)
        if not check_relevance(ref, combined):
            return False, "off-topic metadata"
        return True, ""

    return False, "unknown task"

def load_g4f_provider(P, name):
    try:
        attr = getattr(P, name, None)
        if isinstance(attr, type):
            return attr
    except Exception:
        pass
    loader = getattr(P, "ProviderLoader", None)
    if loader is not None:
        try:
            return loader.from_name(name)
        except Exception:
            pass
    try:
        mod = importlib.import_module(f"g4f.Provider.{name}")
        cls = getattr(mod, name, None)
        if isinstance(cls, type):
            return cls
    except Exception:
        pass
    return None

async def call_g4f_provider(client, prov, model, messages, timeout_sec):
    resp = await asyncio.wait_for(
        client.chat.completions.create(model=model, provider=prov, messages=messages),
        timeout=timeout_sec,
    )
    if resp and getattr(resp, "choices", None):
        return resp.choices[0].message.content or ""
    return ""

async def call_pollinations_direct(user_prompt, timeout_sec):
    import aiohttp
    payload = {
        "model": "openai",
        "messages": [{"role": "user", "content": user_prompt}],
        "response_format": {"type": "json_object"},
    }
    timeout = aiohttp.ClientTimeout(total=timeout_sec)
    async with aiohttp.ClientSession(timeout=timeout) as session:
        async with session.post("https://text.pollinations.ai/openai", json=payload) as resp:
            if resp.status != 200:
                raise RuntimeError(f"http {resp.status}")
            body = await resp.json(content_type=None)
            choices = body.get("choices") or []
            if choices:
                return choices[0].get("message", {}).get("content", "") or ""
            return ""

async def run_cascade(req):
    import g4f
    import g4f.Provider as P
    from g4f.client import AsyncClient

    loop = asyncio.get_running_loop()
    loop.set_exception_handler(lambda _loop, _ctx: None)

    task = req.get("task", "script")
    sys_prompt = req.get("system_prompt", "")
    user_prompt = req.get("user_prompt", "")
    no_browser = bool(req.get("no_browser", False))

    if sys.platform.startswith("linux") and not os.environ.get("DISPLAY") and not os.environ.get("WAYLAND_DISPLAY"):
        no_browser = True

    client = AsyncClient()
    errors = []

    candidates = [
        ("yqcloud", "Yqcloud", "gpt-4", 16.0, False),
    ]
    if os.environ.get("TUBIE_G4F_ALLOW_POLLINATIONS") == "1":
        candidates.append(("pollinations", None, "openai", 15.0, False))
    if not no_browser:
        candidates.append(("chatgpt", "ChatGPT", "auto", 22.0, False))

    for label, prov_name, model, t_sec, use_sys in candidates:
        try:
            if label == "pollinations":
                raw_text = await call_pollinations_direct((sys_prompt + "\n\n" if sys_prompt else "") + user_prompt, t_sec)
                engine_name = "g4f:Pollinations"
            else:
                prov = load_g4f_provider(P, prov_name)
                if prov is None:
                    errors.append(f"{label}: unavailable")
                    continue
                if use_sys and sys_prompt:
                    msgs = [{"role": "system", "content": sys_prompt}, {"role": "user", "content": user_prompt}]
                else:
                    merged = (sys_prompt + "\n\n" + user_prompt).strip() if sys_prompt else user_prompt
                    msgs = [{"role": "user", "content": merged}]
                raw_text = await call_g4f_provider(client, prov, model, msgs, t_sec)
                engine_name = f"g4f:{prov_name}"

            cleaned = strip_json_fences(raw_text)
            parsed = json.loads(cleaned)
            ok, reason = validate_payload(task, parsed, req)
            if ok:
                return {
                    "ok": True,
                    "engine": engine_name,
                    "json": json.dumps(parsed, ensure_ascii=False),
                }
            errors.append(f"{label}: rejected ({reason})")
        except Exception as e:
            errors.append(f"{label}: {type(e).__name__}: {str(e)[:100]}")

    return {"ok": False, "error": "; ".join(errors)}

def main():
    orig_stdout = sys.stdout
    req = json.loads(sys.stdin.buffer.read().decode("utf-8"))
    sink = io.StringIO()
    try:
        with contextlib.redirect_stdout(sink):
            result = asyncio.run(run_cascade(req))
    except Exception as e:
        result = {"ok": False, "error": f"{type(e).__name__}: {e}"}
    orig_stdout.write(json.dumps(result, ensure_ascii=True))
    orig_stdout.flush()

if __name__ == "__main__":
    main()
`

type g4fRequest struct {
	Task         string `json:"task"`
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
	Topic        string `json:"topic,omitempty"`
	Title        string `json:"title,omitempty"`
	Narration    string `json:"narration,omitempty"`
	NoBrowser    bool   `json:"no_browser,omitempty"`
}

type g4fResponse struct {
	OK     bool   `json:"ok"`
	Engine string `json:"engine"`
	JSON   string `json:"json"`
	Error  string `json:"error"`
}

func resolvePythonBinary() string {
	if override := strings.TrimSpace(os.Getenv("PYTHON")); override != "" {
		return override
	}
	py := BinaryPath("python")
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath(py); err != nil {
			if p3, err3 := exec.LookPath("python3"); err3 == nil {
				return p3
			}
		}
	}
	return py
}

// CheckG4FAvailable reports whether Python can import g4f cleanly without network calls.
func CheckG4FAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, resolvePythonBinary(), "-c", "import g4f")
	cmd.Dir = os.TempDir()
	return cmd.Run() == nil
}

func invokeG4F(ctx context.Context, req g4fRequest) (*g4fResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, resolvePythonBinary(), "-c", g4fBridgeScript)
	cmd.Dir = os.TempDir()
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("g4f execution failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	var resp g4fResponse
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &resp); err != nil {
		return nil, fmt.Errorf("g4f output parse failed: %w", err)
	}
	if !resp.OK || strings.TrimSpace(resp.JSON) == "" {
		return nil, fmt.Errorf("g4f providers failed: %s", resp.Error)
	}
	return &resp, nil
}

func generateScriptViaG4F(topic string, opts ScriptOptions) (*ScriptResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g4fRunnerTimeout)
	defer cancel()

	noBrowser := strings.TrimSpace(os.Getenv("TUBIE_G4F_NO_BROWSER")) == "1"
	sysPrompt := "Sen ana dili Türkçe olan profesyonel bir siber güvenlik kısa video senaristisin. Yalnızca dilbilgisi kusursuz, doğal Türkçe kullan; uydurma kelime, İngilizce yüzde ('70 %' değil '%70') veya 'Efsane:', 'Kanca:' gibi bölüm etiketleri ASLA yazma. Sadece tek bir geçerli JSON objesi döndür."
	userPrompt := buildScriptSystemPromptWithOptions(topic, opts)

	resp, err := invokeG4F(ctx, g4fRequest{
		Task:         "script",
		SystemPrompt: sysPrompt,
		UserPrompt:   userPrompt,
		Topic:        topic,
		NoBrowser:    noBrowser,
	})
	if err != nil {
		return nil, err
	}

	return parseAndNormalizeScript(resp.JSON, topic, opts.Style, resp.Engine)
}

func generateViralViaG4F(title, narration string) (*ViralMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g4fRunnerTimeout)
	defer cancel()

	noBrowser := strings.TrimSpace(os.Getenv("TUBIE_G4F_NO_BROWSER")) == "1"
	sysPrompt := "You are a strict JSON API that writes Turkish social-media publishing packs matching the exact video topic. Reply ONLY with one valid JSON object. Never output markdown fences or unrelated topics."
	userPrompt := buildViralMetadataPrompt(title, narration)

	resp, err := invokeG4F(ctx, g4fRequest{
		Task:         "viral",
		SystemPrompt: sysPrompt,
		UserPrompt:   userPrompt,
		Title:        title,
		Narration:    narration,
		NoBrowser:    noBrowser,
	})
	if err != nil {
		return nil, err
	}

	return parseAndNormalizeViralMetadata(resp.JSON)
}
