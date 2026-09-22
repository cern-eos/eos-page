let catalog = { settings: {}, pages: [], cards: [], news: [], people: [], workshops: [], docs: [], years: [], index: {} };

const $ = (sel, root = document) => root.querySelector(sel);

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function setting(key) {
  return catalog.settings?.[key] || "";
}

function page(id) {
  return (catalog.pages || []).find((p) => p.id === id) || { title: "", body: "" };
}

function cards(kind) {
  return (catalog.cards || []).filter((c) => c.kind === kind);
}

function ordinal(n) {
  const v = Number(n);
  if (!v) return "";
  const s = ["th", "st", "nd", "rd"];
  const m = v % 100;
  return v + (s[(m - 20) % 10] || s[m] || s[0]);
}

function path() {
  return location.pathname.replace(/\/+$/, "") || "/";
}

function samePlace(href) {
  const u = new URL(href, location.origin);
  const next = u.pathname.replace(/\/+$/, "") || "/";
  return next === path() && u.search === location.search;
}

let pageBusy = false;
let sprayRAF = 0;
let growthRAF = 0;
let growthHold = 0;
let growthGen = 0;
let termTimer = 0;
let aboutLitTimer = 0;
let aboutLitHover = 0;
const TERM_CLONE = "git clone https://github.com/cern-eos/eos.git";
function prefersQuiet() {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

function go(href) {
  const u = new URL(href, location.origin);
  if (u.origin !== location.origin) {
    location.href = href;
    return;
  }
  document.querySelector(".top")?.classList.remove("is-menu");
  const next = u.pathname + u.search + u.hash;
  if (samePlace(u.pathname + u.search)) {
    history.pushState({}, "", next);
    scrollToHash();
    return;
  }
  if (pageBusy) return;
  if (prefersQuiet()) {
    history.pushState({}, "", next);
    render();
    scrollToHash();
    return;
  }
  fadeTo(() => {
    history.pushState({}, "", next);
    render();
    scrollToHash();
  });
}

function scrollToHash() {
  const id = decodeURIComponent((location.hash || "").replace(/^#/, ""));
  if (!id) {
    window.scrollTo(0, 0);
    return;
  }
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: prefersQuiet() ? "auto" : "smooth", block: "start" });
  else window.scrollTo(0, 0);
}

function fadeTo(swap) {
  const app = document.getElementById("app");
  if (!app) {
    swap();
    return;
  }
  pageBusy = true;
  app.classList.add("fade-out");
  let done = false;
  const afterOut = (e) => {
    if (e && (e.target !== app || (e.propertyName && e.propertyName !== "opacity"))) return;
    if (done) return;
    done = true;
    app.removeEventListener("transitionend", afterOut);
    swap();
    app.classList.add("fade-prep");
    app.classList.remove("fade-out");
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        app.classList.remove("fade-prep");
        app.classList.add("fade-in");
        let inDone = false;
        const afterIn = (ev) => {
          if (ev && (ev.target !== app || (ev.propertyName && ev.propertyName !== "opacity"))) return;
          if (inDone) return;
          inDone = true;
          app.removeEventListener("transitionend", afterIn);
          app.classList.remove("fade-in");
          pageBusy = false;
        };
        app.addEventListener("transitionend", afterIn);
        window.setTimeout(afterIn, 420);
      });
    });
  };
  app.addEventListener("transitionend", afterOut);
  window.setTimeout(afterOut, 380);
}

function diopsideFigure() {
  return `
    <figure class="diopside" title="Faceted Diopside, Madagascar — Didier Descouens, Wikimedia Commons, CC BY-SA 4.0">
      <div class="diopside-rays" aria-hidden="true"></div>
      <div class="diopside-glow" aria-hidden="true"></div>
      <div class="diopside-stone">
        <img src="/static/media/diopside.png?v=forest" alt="Faceted Diopside gemstone" width="920" height="519" />
      </div>
    </figure>`;
}

function letterizeTitle(text) {
  return String(text).split(/\s+/).filter(Boolean).map((part) => {
    const open = part.toLowerCase() === "open";
    const storage = part.toLowerCase() === "storage";
    const letters = [...part].map((ch, i) => {
      const openO = open && i === 0 && ch.toLowerCase() === "o";
      const cls = `title-letter${open ? " is-open" : ""}${openO ? " is-o" : ""}`;
      return `<span class="${cls}">${esc(ch)}${openO ? diopsideFigure() : ""}</span>`;
    }).join("");
    return `<span class="title-word${storage ? " is-storage" : ""}">${letters}</span>`;
  }).join("");
}

function easeInOutQuad(t) {
  return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
}

function stopTitleSpray() {
  cancelAnimationFrame(sprayRAF);
  sprayRAF = 0;
  document.querySelectorAll(".title-spray, .title-spark").forEach((el) => el.remove());
}

const EOS_CAPACITY = [
  { year: 2010, pb: 5, approx: true },
  { year: 2011, pb: 15, approx: true },
  { year: 2012, pb: 30, approx: true },
  { year: 2013, pb: 45, approx: true },
  { year: 2014, pb: 70, approx: true },
  { year: 2015, pb: 100, approx: true },
  { year: 2016, pb: 135, approx: false, mark: true },
  { year: 2017, pb: 180, approx: true },
  { year: 2018, pb: 220, approx: true },
  { year: 2019, pb: 280, approx: true },
  { year: 2020, pb: 350, approx: true },
  { year: 2021, pb: 500, approx: true },
  { year: 2022, pb: 780, approx: false, mark: true },
  { year: 2023, pb: 900, approx: true },
  { year: 2024, pb: 1050, approx: true, mark: true },
  { year: 2025, pb: 1200, approx: true, mark: true },
  { year: 2030, pb: 2500, target: true, mark: true },
];

const EOS_HINGE_YEAR = 2025;

function formatPB(pb, approx) {
  const n = Math.round(pb);
  const s = n >= 1000 ? n.toLocaleString("en-US") : String(n);
  return `${approx ? "~" : ""}${s} PB`;
}

function formatCapacity(pb, meta = {}) {
  if (meta.target || Math.round(pb) >= 2500) return "2.5 EB";
  return formatPB(pb, meta.approx !== false);
}

function capacityChartMarkup() {
  return `
    <figure class="eos-growth">
      <figcaption class="eos-growth-cap">
        <span>Raw disk at CERN</span>
        <strong aria-live="polite"><em data-growth-year>2010</em><i data-growth-pb>~5 PB</i></strong>
      </figcaption>
      <svg class="eos-growth-svg" viewBox="0 0 720 240" role="img" aria-label="EOS raw capacity at CERN from about 5 PB in 2010 to a 2.5 EB target in 2030">
        <defs>
          <linearGradient id="growth-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#e65c00" stop-opacity="0.4"/>
            <stop offset="100%" stop-color="#e65c00" stop-opacity="0"/>
          </linearGradient>
          <linearGradient id="growth-target-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#e65c00" stop-opacity="0.16"/>
            <stop offset="100%" stop-color="#e65c00" stop-opacity="0"/>
          </linearGradient>
          <linearGradient id="growth-stroke" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#f0a060"/>
            <stop offset="100%" stop-color="#fff6e8"/>
          </linearGradient>
        </defs>
        <g class="eos-growth-plot">
          <path class="eos-growth-area"></path>
          <path class="eos-growth-area-target"></path>
        </g>
        <g class="eos-growth-grid"></g>
        <g class="eos-growth-plot-line">
          <path class="eos-growth-line"></path>
          <path class="eos-growth-line-dash"></path>
        </g>
        <g class="eos-growth-marks"></g>
        <circle class="eos-growth-dot" r="4.6" cx="0" cy="0"></circle>
      </svg>
      <p class="eos-growth-note">One point per year through 2025. Values with ~ are approximate. The dashed line extrapolates to the 2.5 EB target in 2030.</p>
    </figure>`;
}

function growthLayout() {
  const W = 720, H = 240;
  const pad = { l: 54, r: 28, t: 36, b: 34 };
  const innerW = W - pad.l - pad.r;
  const innerH = H - pad.t - pad.b;
  const minYear = EOS_CAPACITY[0].year;
  const maxYear = EOS_CAPACITY[EOS_CAPACITY.length - 1].year;
  const maxPB = 2500;
  const xOf = (year) => pad.l + ((year - minYear) / (maxYear - minYear)) * innerW;
  const yOf = (pb) => pad.t + (1 - Math.max(0, pb) / maxPB) * innerH;
  const pts = EOS_CAPACITY.map((d) => ({ ...d, x: xOf(d.year), y: yOf(d.pb) }));
  return { W, H, pad, innerW, innerH, maxPB, xOf, yOf, pts };
}

function curvePath(pts) {
  if (pts.length < 2) return "";
  let d = `M ${pts[0].x.toFixed(2)} ${pts[0].y.toFixed(2)}`;
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i - 1] || pts[i];
    const p1 = pts[i];
    const p2 = pts[i + 1];
    const p3 = pts[i + 2] || p2;
    const c1x = p1.x + (p2.x - p0.x) / 8;
    const c1y = p1.y + (p2.y - p0.y) / 8;
    const c2x = p2.x - (p3.x - p1.x) / 8;
    const c2y = p2.y - (p3.y - p1.y) / 8;
    d += ` C ${c1x.toFixed(2)} ${c1y.toFixed(2)} ${c2x.toFixed(2)} ${c2y.toFixed(2)} ${p2.x.toFixed(2)} ${p2.y.toFixed(2)}`;
  }
  return d;
}

function growthSample(t) {
  const { pts } = growthLayout();
  if (t >= 1) return { ...pts[pts.length - 1], t: 1 };
  const u = t * (pts.length - 1);
  const i = Math.min(pts.length - 2, Math.floor(u));
  const f = u - i;
  const a = pts[i];
  const b = pts[i + 1];
  return {
    x: a.x + (b.x - a.x) * f,
    y: a.y + (b.y - a.y) * f,
    year: a.year + (b.year - a.year) * f,
    pb: a.pb + (b.pb - a.pb) * f,
    approx: f < 0.035 ? a.approx : true,
    t,
  };
}

function growthVisible(t) {
  const { pts, pad, innerH } = growthLayout();
  const sample = growthSample(t);
  const hist = pts.filter((p) => !p.target);
  const hinge = hist[hist.length - 1];
  const u = t * (pts.length - 1);
  const lastIdx = Math.min(pts.length - 1, Math.floor(u + 1e-6));
  const vis = pts.slice(0, lastIdx + 1);
  if (t < 1 && (!vis.length || Math.abs(vis[vis.length - 1].x - sample.x) > 0.05)) {
    vis.push(sample);
  }
  const visHist = vis.filter((p) => !p.target);
  if (sample.year > EOS_HINGE_YEAR && hinge && (!visHist.length || visHist[visHist.length - 1].year < hinge.year)) {
    visHist.push(hinge);
  }
  const line = visHist.length >= 2 ? curvePath(visHist) : `M ${sample.x.toFixed(2)} ${sample.y.toFixed(2)}`;
  const dash = sample.year > EOS_HINGE_YEAR
    ? `M ${hinge.x.toFixed(2)} ${hinge.y.toFixed(2)} L ${sample.x.toFixed(2)} ${sample.y.toFixed(2)}`
    : "";
  const base = (pad.t + innerH).toFixed(2);
  const areaEnd = sample.year > EOS_HINGE_YEAR ? hinge : sample;
  const area = visHist.length
    ? `${line} L ${areaEnd.x.toFixed(2)} ${base} L ${pts[0].x.toFixed(2)} ${base} Z`
    : "";
  const areaTarget = dash
    ? `M ${hinge.x.toFixed(2)} ${hinge.y.toFixed(2)} L ${sample.x.toFixed(2)} ${sample.y.toFixed(2)} L ${sample.x.toFixed(2)} ${base} L ${hinge.x.toFixed(2)} ${base} Z`
    : "";
  return { line, dash, area, areaTarget, sample };
}

function applyGrowth(root, t) {
  const { pad, innerH } = growthLayout();
  const { line, dash, area, areaTarget, sample } = growthVisible(t);
  const lineEl = root.querySelector(".eos-growth-line");
  const dashEl = root.querySelector(".eos-growth-line-dash");
  const areaEl = root.querySelector(".eos-growth-area");
  const targetEl = root.querySelector(".eos-growth-area-target");
  const dot = root.querySelector(".eos-growth-dot");
  const yearEl = root.querySelector("[data-growth-year]");
  const pbEl = root.querySelector("[data-growth-pb]");
  if (lineEl) lineEl.setAttribute("d", line);
  if (dashEl) dashEl.setAttribute("d", dash);
  if (areaEl) areaEl.setAttribute("d", area);
  if (targetEl) targetEl.setAttribute("d", areaTarget);
  if (dot) {
    dot.setAttribute("cx", sample.x.toFixed(2));
    dot.setAttribute("cy", sample.y.toFixed(2));
    dot.classList.toggle("is-target", sample.year > EOS_HINGE_YEAR);
  }
  if (yearEl) yearEl.textContent = String(Math.round(sample.year));
  if (pbEl) pbEl.textContent = formatCapacity(sample.pb, sample);
  const baseline = pad.t + innerH;
  root.querySelectorAll(".eos-growth-yline").forEach((el) => {
    const pb = Number(el.getAttribute("data-pb"));
    const rise = Math.max(0, Math.min(1, (sample.pb + 60 - pb) / 180));
    el.setAttribute("x2", sample.x.toFixed(2));
    el.style.opacity = rise > 0 ? String(0.25 + 0.75 * rise) : "0";
  });
  root.querySelectorAll(".eos-growth-grid text[data-pb]").forEach((el) => {
    const pb = Number(el.getAttribute("data-pb"));
    el.style.opacity = sample.pb + 60 >= pb ? "1" : "0";
  });
  root.querySelectorAll(".eos-growth-xline").forEach((el) => {
    const year = Number(el.getAttribute("data-year"));
    const local = Math.max(0, Math.min(1, (sample.year + 0.2 - year) / 0.55));
    el.setAttribute("y1", baseline.toFixed(2));
    el.setAttribute("y2", (baseline - local * innerH).toFixed(2));
    el.style.opacity = local > 0 ? "1" : "0";
  });
  root.querySelectorAll(".eos-growth-marks [data-year], .eos-growth-grid text.is-mark, .eos-growth-tick").forEach((el) => {
    const year = Number(el.getAttribute("data-year"));
    el.style.opacity = sample.year + 0.15 >= year ? "1" : "0";
  });
}

function buildCapacityChart(root) {
  const { pad, innerH, yOf, pts } = growthLayout();
  const baseline = pad.t + innerH;
  const yTicks = [0, 500, 1000, 1500, 2000, 2500];
  const grid = yTicks.map((v) => {
    const y = yOf(v);
    const label = v === 2500 ? "2.5 EB" : v >= 1000 ? v.toLocaleString("en-US") : String(v);
    return `<line class="eos-growth-yline" data-pb="${v}" x1="${pad.l}" y1="${y.toFixed(1)}" x2="${pad.l}" y2="${y.toFixed(1)}" style="opacity:0"/><text data-pb="${v}" x="${pad.l - 6}" y="${(y + 3).toFixed(1)}" text-anchor="end" style="opacity:0">${label}</text>`;
  }).join("") + pts.map((p) => {
    const x = p.x;
    const labeled = p.year % 2 === 0 || p.year === 2025 || p.target;
    const major = labeled ? " is-major" : "";
    const xline = `<line class="eos-growth-xline${major}" data-year="${p.year}" x1="${x.toFixed(1)}" y1="${baseline.toFixed(1)}" x2="${x.toFixed(1)}" y2="${baseline.toFixed(1)}" style="opacity:0"/>`;
    const tick = `<line class="eos-growth-tick" data-year="${p.year}" x1="${x.toFixed(1)}" y1="${baseline.toFixed(1)}" x2="${x.toFixed(1)}" y2="${(baseline + (labeled ? 5 : 3)).toFixed(1)}" style="opacity:0"/>`;
    if (!labeled) return `${xline}${tick}`;
    return `${xline}${tick}<text class="is-mark" data-year="${p.year}" x="${x.toFixed(1)}" y="${(baseline + 18).toFixed(1)}" text-anchor="middle" style="opacity:0">${p.year}</text>`;
  }).join("");
  const gridEl = root.querySelector(".eos-growth-grid");
  if (gridEl) gridEl.innerHTML = grid;

  const marks = root.querySelector(".eos-growth-marks");
  if (marks) {
    marks.innerHTML = pts.map((p) => {
      const labeled = p.year % 2 === 0 || p.year === 2025 || p.target;
      const r = p.target ? 5.2 : p.mark || labeled ? 3.6 : 2.2;
      const klass = p.target ? "eos-growth-target" : p.mark ? "eos-growth-mark" : "eos-growth-pt";
      const anchor = p.target || p.year >= 2024 ? "end" : p.year === 2010 ? "start" : "middle";
      const lx = p.target || p.year >= 2024 ? p.x - 6 : p.year === 2010 ? p.x + 5 : p.x;
      const ly = p.y - (p.target ? 10 : 12);
      const text = p.target ? "2.5 EB" : formatPB(p.pb, p.approx);
      const label = labeled
        ? `<text class="eos-growth-val${p.target ? " is-target" : ""}" data-year="${p.year}" x="${lx.toFixed(2)}" y="${ly.toFixed(2)}" text-anchor="${anchor}" style="opacity:0">${text}</text>`
        : "";
      return `${label}<circle class="${klass}" r="${r}" cx="${p.x.toFixed(2)}" cy="${p.y.toFixed(2)}" data-year="${p.year}" style="opacity:0"></circle>`;
    }).join("");
  }
}

function stopTermType() {
  window.clearTimeout(termTimer);
  termTimer = 0;
}

function startTermType() {
  stopTermType();
  const el = document.querySelector("[data-term]");
  if (!el) return;
  if (prefersQuiet()) {
    el.textContent = TERM_CLONE;
    return;
  }
  let i = 0;
  const tick = () => {
    if (!document.querySelector("[data-term]")) return;
    i += 1;
    el.textContent = TERM_CLONE.slice(0, i);
    if (i < TERM_CLONE.length) {
      termTimer = window.setTimeout(tick, i < 10 ? 55 : 28);
      return;
    }
    termTimer = window.setTimeout(() => {
      el.textContent = "";
      i = 0;
      termTimer = window.setTimeout(tick, 420);
    }, 2600);
  };
  termTimer = window.setTimeout(tick, 280);
}

function stopCapacityChart() {
  growthGen += 1;
  cancelAnimationFrame(growthRAF);
  growthRAF = 0;
  window.clearTimeout(growthHold);
  growthHold = 0;
}

function startHeroBackground() {
  const wrap = document.querySelector(".hero-video");
  const iframe = wrap?.querySelector("iframe");
  if (!wrap || !iframe) return;
  let src = iframe.dataset.src;
  if (!src) return;
  if (!src.includes("enablejsapi=1")) {
    src += (src.includes("?") ? "&" : "?") + "enablejsapi=1&origin=" + encodeURIComponent(location.origin);
  }
  const reveal = () => {
    if (!document.contains(iframe)) return;
    wrap.classList.add("is-live");
  };
  if (iframe.getAttribute("src") === src) {
    reveal();
    return;
  }
  const onMsg = (e) => {
    if (!/youtube\.com$/.test(e.origin || "") && !/youtube-nocookie\.com$/.test(e.origin || "")) return;
    let data = e.data;
    if (typeof data === "string") {
      try { data = JSON.parse(data); } catch { return; }
    }
    if (!data || typeof data !== "object") return;
    if (data.event === "onReady") {
      try {
        iframe.contentWindow.postMessage(JSON.stringify({
          event: "command", func: "addEventListener", args: ["onStateChange"],
        }), "*");
      } catch (_) {}
    }
    const playing = data.event === "onStateChange" && (data.info === 1 || data.info === 3);
    const delivery = data.event === "infoDelivery" && data.info && data.info.playerState === 1;
    if (playing || delivery) {
      window.removeEventListener("message", onMsg);
      reveal();
    }
  };
  window.addEventListener("message", onMsg);
  iframe.addEventListener("load", () => {
    try {
      iframe.contentWindow.postMessage(JSON.stringify({ event: "listening", id: "hero" }), "*");
    } catch (_) {}
    window.setTimeout(reveal, prefersQuiet() ? 200 : 1600);
  }, { once: true });
  iframe.src = src;
}

function startOrbitMovie() {
  const video = document.querySelector(".hero-orbit");
  if (!video) return;
  if (!video.querySelector("source")) {
    const source = document.createElement("source");
    source.src = "/static/media/orbit.mp4";
    source.type = "video/mp4";
    video.appendChild(source);
    video.load();
  }
  video.currentTime = 0;
  video.play().catch(() => {});
}

function stopAboutHighlight() {
  window.clearTimeout(aboutLitTimer);
  window.clearInterval(aboutLitTimer);
  aboutLitTimer = 0;
  aboutLitHover = 0;
  document.querySelectorAll(".about-features .is-lit").forEach((el) => el.classList.remove("is-lit"));
}

function startAboutHighlight() {
  stopAboutHighlight();
  const root = document.querySelector(".about-features");
  if (!root) return;
  const boxes = [...root.querySelectorAll(".set-card")];
  if (!boxes.length) return;
  let i = 0;
  let on = false;
  const held = () => aboutLitHover > 0 || boxes.some((el) => el.matches(":hover, :focus-within"));
  const paint = () => {
    const hold = held();
    boxes.forEach((el, n) => el.classList.toggle("is-lit", on && !hold && n === i));
  };
  boxes.forEach((el) => {
    el.addEventListener("pointerenter", () => {
      aboutLitHover += 1;
      paint();
    });
    el.addEventListener("pointerleave", () => {
      aboutLitHover = Math.max(0, aboutLitHover - 1);
      paint();
    });
  });
  root.addEventListener("focusin", paint);
  root.addEventListener("focusout", paint);
  aboutLitTimer = window.setTimeout(() => {
    if (!document.contains(root)) return;
    on = true;
    i = 0;
    paint();
    aboutLitTimer = window.setInterval(() => {
      if (held()) return;
      i = (i + 1) % boxes.length;
      paint();
    }, 2000);
  }, 2000);
}

function startCapacityChart() {
  stopCapacityChart();
  const root = document.querySelector(".eos-growth");
  if (!root) return;
  buildCapacityChart(root);
  applyGrowth(root, 0);
  const duration = prefersQuiet() ? 4800 : 7800;
  const hold = 2200;
  const gen = growthGen;
  const t0 = performance.now();
  const step = (now) => {
    if (gen !== growthGen) return;
    const t = Math.min(1, (now - t0) / duration);
    applyGrowth(root, t);
    if (t < 1) {
      growthRAF = requestAnimationFrame(step);
      return;
    }
    applyGrowth(root, 1);
    if (prefersQuiet()) return;
    growthHold = window.setTimeout(() => {
      if (gen === growthGen) startCapacityChart();
    }, hold);
  };
  growthRAF = requestAnimationFrame(step);
}

function spawnTitleSpark(h1, x, y) {
  const s = document.createElement("i");
  const kind = Math.random();
  s.className = "title-spark" + (kind < 0.4 ? " star5" : "") + (kind > 0.72 ? " is-ember" : "");
  const size = 5 + Math.random() * 8;
  s.style.cssText = `left:${x.toFixed(1)}px;top:${y.toFixed(1)}px;width:${size}px;height:${size}px`;
  h1.appendChild(s);
  window.setTimeout(() => s.remove(), 900);
}

function paintTitleMetal() {
  document.querySelectorAll(".title-letter").forEach((el) => {
    el.classList.add("sprayed");
    el.classList.remove("wet");
  });
  const h1 = document.querySelector(".title-loupe");
  if (h1) h1.classList.add("is-metal");
  addTitleGlints(h1);
}

function addTitleGlints(h1) {
  if (!h1 || h1.querySelector(".title-glints")) return;
  const n = 28;
  const bits = Array.from({ length: n }, (_, i) => {
    const x = 3 + Math.random() * 94;
    const y = 6 + Math.random() * 82;
    const delay = (Math.random() * 2.8).toFixed(2);
    const dur = (1.3 + Math.random() * 1.8).toFixed(2);
    const wide = i % 5 === 0 ? " class=\"wide\"" : "";
    return `<i${wide} style="left:${x}%;top:${y}%;animation-delay:${delay}s;animation-duration:${dur}s"></i>`;
  }).join("");
  const wrap = document.createElement("span");
  wrap.className = "title-glints";
  wrap.setAttribute("aria-hidden", "true");
  wrap.innerHTML = bits;
  h1.appendChild(wrap);
}

function startTitleSpray() {
  stopTitleSpray();
  const h1 = document.querySelector(".title-loupe");
  const letters = [...document.querySelectorAll(".title-letter")];
  if (!h1 || !letters.length) return;
  if (prefersQuiet()) {
    paintTitleMetal();
    return;
  }
  const spray = document.createElement("span");
  spray.className = "title-spray";
  spray.setAttribute("aria-hidden", "true");
  spray.innerHTML = `<span class="tail"></span><span class="bar"></span>`;
  h1.appendChild(spray);
  const start = performance.now();
  const duration = 2600;
  const painted = new Set();
  let lastSpark = 0;
  const letterBox = (j) => {
    const last = letters[letters.length - 1].getBoundingClientRect();
    if (j < letters.length) return letters[j].getBoundingClientRect();
    return {
      left: last.right + last.width * 0.08,
      top: last.top,
      width: last.width,
      height: last.height,
    };
  };
  const tick = (now) => {
    if (!document.body.contains(h1)) return;
    const t = Math.min(1, (now - start) / duration);
    const e = easeInOutQuad(t);
    const steps = letters.length + 1;
    const idx = e * (steps - 0.001);
    const i = Math.min(steps - 1, Math.floor(idx));
    const frac = idx - i;
    const h1r = h1.getBoundingClientRect();
    const a = letterBox(i);
    const b = letterBox(Math.min(steps - 1, i + 1));
    const ax = a.left - h1r.left + a.width * 0.72;
    const bx = b.left - h1r.left + b.width * 0.18;
    const x = ax + (bx - ax) * frac;
    const y = a.top - h1r.top + (b.top - a.top) * frac;
    const h = a.height + (b.height - a.height) * frac;
    spray.style.height = `${h}px`;
    spray.style.transform = `translate3d(${x}px, ${y}px, 0)`;
    spray.style.opacity = t < 0.04 ? t / 0.04 : t > 0.97 ? (1 - t) / 0.03 : 1;
    if (now - lastSpark > 26) {
      lastSpark = now;
      const n = 1 + (Math.random() < 0.45 ? 1 : 0);
      for (let k = 0; k < n; k++) {
        spawnTitleSpark(
          h1,
          x - 6 - Math.random() * 42,
          y + h * (0.12 + Math.random() * 0.76)
        );
      }
    }
    letters.forEach((el, j) => {
      if (j > idx || painted.has(el)) return;
      painted.add(el);
      el.classList.add("sprayed", "wet");
      window.setTimeout(() => el.classList.remove("wet"), 320);
    });
    if (t < 1) {
      sprayRAF = requestAnimationFrame(tick);
      return;
    }
    spray.remove();
    paintTitleMetal();
  };
  sprayRAF = requestAnimationFrame(tick);
}

function setNav() {
  const p = path();
  document.querySelectorAll(".nav a").forEach((a) => {
    const href = a.getAttribute("href");
    a.classList.toggle("on", href === p || (href !== "/" && p.startsWith(href)));
  });
}

function hero(extra = "") {
  const vid = setting("hero_video") || "ttSjYYBOlsM";
  return `
    <section class="hero">
      <div class="hero-video" aria-hidden="true">
        <img class="hero-poster" src="/static/media/hero-poster.jpg" alt="" />
        <iframe
          data-src="https://www.youtube-nocookie.com/embed/${esc(vid)}?autoplay=1&mute=1&controls=0&loop=1&playlist=${esc(vid)}&start=10&playsinline=1&rel=0&modestbranding=1&iv_load_policy=3"
          title="EOS background video"
          allow="autoplay; encrypted-media; picture-in-picture"
          tabindex="-1"></iframe>
      </div>
      <div class="hero-overlay"></div>
      <div class="hero-inner">
        <p class="hero-term" aria-label="${esc(TERM_CLONE)}">
          <span class="hero-term-prompt">$</span>
          <span class="hero-term-text" data-term></span>
          <span class="hero-term-caret" aria-hidden="true"></span>
        </p>
        <p class="kicker">${esc(setting("hero_kicker"))}</p>
        <div class="hero-title-row">
          <div class="hero-copy">
            <h1 class="title-loupe" aria-label="EOS Open Storage">${letterizeTitle("EOS Open Storage")}</h1>
            ${capacityChartMarkup()}
            <p class="lede">${esc(setting("hero_lede"))}</p>
            <div class="actions">
              <a class="btn" href="${esc(setting("workshop_url") || "/workshops")}">${esc(setting("workshop_label") || "Workshops")}</a>
              <a class="btn ghost" href="${esc(setting("latest_version_url") || "/docs")}">Latest v${esc(setting("latest_version") || "5")}</a>
              <a class="btn ghost" data-nav href="/docs">Install</a>
              <a class="btn ghost" data-nav href="/search">Search talks</a>
            </div>
          </div>
          <div class="orbit-frame">
            <video class="hero-orbit" muted loop playsinline preload="none" aria-label="EOS orbit"></video>
          </div>
        </div>
        ${extra}
      </div>
    </section>`;
}

function stats(extra = "") {
  const items = [
    [setting("stat_volume"), setting("stat_volume_label")],
    [setting("stat_disks"), setting("stat_disks_label")],
    [setting("stat_files"), setting("stat_files_label")],
    [setting("stat_clients"), setting("stat_clients_label")],
  ];
  return `<div class="stats${extra ? ` ${extra}` : ""}">${items.map(([n, l]) => `<div class="stat"><b>${esc(n)}</b><span>${esc(l)}</span></div>`).join("")}</div>`;
}

function cardIcon(id) {
  const stroke = `fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"`;
  const icons = {
    "f-flex": `<svg viewBox="0 0 24 24" ${stroke}><rect x="3.5" y="3.5" width="7" height="7" rx="1.4"/><rect x="13.5" y="3.5" width="7" height="7" rx="1.4"/><rect x="3.5" y="13.5" width="7" height="7" rx="1.4"/><path d="M17 13.5v7M13.5 17h7"/></svg>`,
    "f-scale": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4 19V10M10 19V6M16 19v-8M20 19H3"/><path d="M14.5 5.5 16 4l1.5 1.5M16 4v5"/></svg>`,
    "f-cern": `<svg viewBox="0 0 24 24" ${stroke}><ellipse cx="12" cy="12" rx="8.2" ry="3.3" transform="rotate(-28 12 12)"/><ellipse cx="12" cy="12" rx="8.2" ry="3.3" transform="rotate(28 12 12)"/><circle cx="12" cy="12" r="1.4" fill="currentColor" stroke="none"/></svg>`,
    "f-sec": `<svg viewBox="0 0 24 24" ${stroke}><path d="M12 3.6 5.4 6.2v5.3c0 4.2 2.8 7.2 6.6 8.5 3.8-1.3 6.6-4.3 6.6-8.5V6.2L12 3.6z"/><path d="m9.2 12.1 1.9 1.9 3.8-3.9"/></svg>`,
    "f-sync": `<svg viewBox="0 0 24 24" ${stroke}><path d="M7.2 8.2A6.2 6.2 0 0 1 18 10.4"/><path d="M16.6 7.2 18.2 10.4 15 11"/><path d="M16.8 15.8A6.2 6.2 0 0 1 6 13.6"/><path d="M7.4 16.8 5.8 13.6 9 13"/></svg>`,
    "f-tape": `<svg viewBox="0 0 24 24" ${stroke}><rect x="3.2" y="5.2" width="17.6" height="13.6" rx="2.2"/><circle cx="8.4" cy="12" r="2.6"/><circle cx="15.6" cy="12" r="2.6"/><path d="M11 12h2"/></svg>`,
    "c-mgm": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4 8.2 12 4l8 4.2v7.6L12 20l-8-4.2z"/><path d="M12 12.2 20 8.2M12 12.2V20M12 12.2 4 8.2"/></svg>`,
    "c-mq": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4 8h6v8H4zM14 8h6v8h-6z"/><path d="M10 12h4"/></svg>`,
    "c-qdb": `<svg viewBox="0 0 24 24" ${stroke}><circle cx="12" cy="6.5" r="2.4"/><circle cx="6.4" cy="16.4" r="2.4"/><circle cx="17.6" cy="16.4" r="2.4"/><path d="M10.2 8.2 7.8 14.2M13.8 8.2l2.4 6M8.8 16.4h6.4"/></svg>`,
    "c-fst": `<svg viewBox="0 0 24 24" ${stroke}><ellipse cx="12" cy="7" rx="7.2" ry="2.6"/><path d="M4.8 7v7.4c0 1.5 3.2 2.6 7.2 2.6s7.2-1.1 7.2-2.6V7"/><ellipse cx="12" cy="14.4" rx="7.2" ry="2.6"/></svg>`,
    "c-cli": `<svg viewBox="0 0 24 24" ${stroke}><path d="M5 8.2 9.2 12 5 15.8"/><path d="M12.2 16.4H19"/></svg>`,
    "s-orbit": `<svg viewBox="0 0 24 24" ${stroke}><circle cx="12" cy="12" r="3.1"/><ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(-18 12 12)"/><path d="M18.6 8.2a9 9 0 1 1-1.4-1.8"/><circle cx="18.8" cy="7.6" r="1.15" fill="currentColor" stroke="none"/></svg>`,
    "s-tower": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4 18V8.5M9 18V5M14 18v-7M19 18V7"/><path d="M3.5 19h17"/></svg>`,
    "s-box": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4.4 8.4 12 4.6l7.6 3.8v7.2L12 19.4 4.4 15.6z"/><path d="M12 12.2 19.6 8.4M12 12.2V19.4M12 12.2 4.4 8.4"/></svg>`,
    "s-swan": `<svg viewBox="0 0 24 24" ${stroke}><rect x="4" y="4.5" width="16" height="15" rx="2"/><path d="M8 9h8M8 12.5h8M8 16h5"/></svg>`,
    "s-cta": `<svg viewBox="0 0 24 24" ${stroke}><rect x="3.2" y="5.2" width="17.6" height="13.6" rx="2.2"/><circle cx="8.4" cy="12" r="2.6"/><circle cx="15.6" cy="12" r="2.6"/><path d="M11 12h2"/></svg>`,
  };
  const svg = icons[id];
  return svg ? `<span class="card-ico" aria-hidden="true">${svg}</span>` : "";
}

function cardPreview(c) {
  const fallbacks = {
    "s-tower": "/static/media/control-tower.jpg",
    "s-orbit": "/static/media/eos-orbit.jpg",
    "s-box": "/static/media/cernbox-web.jpg",
    "s-swan": "/static/media/swan-web.jpg",
    "s-cta": "/static/media/cta-web.jpg",
  };
  if (c.image) return c.image;
  if (fallbacks[c.id]) return fallbacks[c.id];
  if (/\.(png|jpe?g|webp|gif)$/i.test(c.meta || "")) return c.meta;
  return "";
}

function cardGrid(list, cols = "grid-3", extra = "") {
  return `<div class="grid ${cols}${extra ? ` ${extra}` : ""}">${list.map((c) => {
    const preview = cardPreview(c);
    const meta = preview && c.meta === preview ? "" : c.meta;
    return `
    <article class="card set-card${preview ? " has-preview" : ""}">
      ${preview ? `<a class="card-preview" ${c.href && c.href.startsWith("/") ? "data-nav" : 'target="_blank" rel="noreferrer"'} href="${esc(c.href || preview)}"><img src="${esc(preview)}" alt="${esc(c.title || "")}" /></a>` : ""}
      ${meta ? `<p class="meta">${esc(meta)}</p>` : ""}
      <h3 class="card-title">${cardIcon(c.id)}${esc(c.title)}</h3>
      <p>${esc(c.body)}</p>
      ${c.href ? `<a class="more" ${c.href.startsWith("/") ? "data-nav" : 'target="_blank" rel="noreferrer"'} href="${esc(c.href)}">Open →</a>` : ""}
    </article>`;
  }).join("")}</div>`;
}

function home() {
  return `
    ${hero()}
    ${stats()}
    <div class="wrap">
      <div class="section-head">
        <h2>${esc(page("about").title || "About EOS")}</h2>
        <p class="muted">${highlightLead(page("about").body)}</p>
      </div>
      ${cardGrid(cards("feature"), "grid-3", "about-features")}
      <div class="section-head" style="margin-top:2.4rem">
        <h2>Help from the docs</h2>
        <p class="muted">Extracted from eos-docs so you can find setup, architecture, and operations without leaving the homepage.</p>
      </div>
      ${cardGrid(cards("help").length ? cards("help") : (catalog.docs || []).slice(0, 6).map((d) => ({ title: d.title, body: d.summary, href: d.url, meta: d.section })), "grid-3")}
      <p style="margin-top:1rem"><a data-nav href="/docs">All documentation topics →</a></p>
    </div>`;
}

function about() {
  return `
    <div class="wrap">
      <p class="kicker">About</p>
      <h2>${esc(page("about").title)}</h2>
      <div class="prose"><p>${highlightLead(page("about").body)}</p></div>
      <div style="margin-top:1.6rem">${cardGrid(cards("feature"), "grid-3", "about-features")}</div>
    </div>`;
}

const LEAD_KEYS = [
  "erasure coding",
  "CERNBox",
  "Diopside",
  "QuarkDB",
  "XRootD",
  "eosxd",
  "WLCG",
  "CTA",
  "LHC",
];
const LEAD_KEY_RE = new RegExp(
  `(${LEAD_KEYS.slice().sort((a, b) => b.length - a.length).map((k) => k.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|")})`,
  "g"
);

function highlightLead(text) {
  return esc(text).replace(LEAD_KEY_RE, `<mark class="kw">$1</mark>`);
}

function paragraphs(text) {
  return String(text || "")
    .split(/\n\n+/)
    .map((p) => p.trim())
    .filter(Boolean)
    .map((p) => `<p>${highlightLead(p)}</p>`)
    .join("");
}

function architectureFigure() {
  return `
    <figure class="arch-figure">
      <img class="arch-draw" src="/static/media/eos-architecture.svg?v=nomq" width="1280" height="820" alt="EOS Diopside architecture: clients reach the MGM over XRootD or HTTP; metadata is persisted in a three-node QuarkDB RAFT cluster; QuarkDB pub-sub carries MGM–FST messages; file data lives on FST disks, with CERNBox, S3, CIFS, SFTP and CTA at the edge." />
      <div class="arch-stack" aria-hidden="true">
        <p class="arch-layer">01 Access</p>
        <div class="arch-pills">
          <span>eos CLI</span><span>eosxd FUSE</span><span>HTTP / WebDAV</span><span>gRPC</span><span>xrdcp</span><span>XrdCl</span>
        </div>
        <div class="arch-proto">XRootD · auth · redirect · vector I/O · TPC</div>
        <p class="arch-layer">02 Control plane</p>
        <div class="arch-ctrl">
          <article class="set-card"><b>MGM</b><span>Namespace · LRU cache · active + standby</span></article>
          <article class="set-card"><b>QuarkDB</b><span>RAFT · 3 nodes · RocksDB · pub-sub</span></article>
        </div>
        <p class="arch-layer">03 Data plane</p>
        <div class="arch-pills">
          <span>FST replica</span><span>FST drain / fsck</span><span>FST erasure coding</span><span>FST scale-out</span>
        </div>
        <p class="arch-layer">04 Ecosystem</p>
        <div class="arch-pills">
          <span>CERNBox</span><span>Samba / CIFS</span><span>S3 / MinIO</span><span>SFTP</span><span>CTA tape</span>
        </div>
      </div>
      <figcaption class="arch-cap">Three core services — MGM, FST and QuarkDB. Messaging is QuarkDB pub-sub, not a separate MQ. Clients open on the MGM; data I/O is redirected to FSTs.</figcaption>
    </figure>`;
}

function tech() {
  const p = page("tech");
  const docs = setting("docs_url") || "https://eos-docs.web.cern.ch/diopside/";
  const archDocs = "https://eos-docs.web.cern.ch/diopside/architecture/index.html";
  return `
    <div class="wrap tech-page">
      <p class="kicker">Design & architecture</p>
      <h2>${esc(p.title || "Design & architecture")}</h2>
      <div class="prose">${paragraphs(p.body)}</div>
      ${architectureFigure()}
      <ol class="arch-flow">
        <li><b>1</b><div><strong>Open on the MGM</strong><span>Authenticate and resolve the path in the hierarchical namespace.</span></div></li>
        <li><b>2</b><div><strong>Redirect</strong><span>XRootD sends the client to an FST chosen by placement, GEO tags and policy.</span></div></li>
        <li><b>3</b><div><strong>Read or write data</strong><span>I/O stays on the FST — replica or erasure-coded stripes, with checksums.</span></div></li>
        <li><b>4</b><div><strong>Persist metadata</strong><span>The MGM write-back queue commits namespace changes to QuarkDB.</span></div></li>
      </ol>
      <div class="section-head">
        <h2>Core services</h2>
        <p class="muted">MGM, FST and QuarkDB — pub-sub lives in QuarkDB — plus the clients that speak XRootD, HTTP and FUSE.</p>
      </div>
      ${cardGrid(cards("component").filter((c) => c.id !== "c-mq"), "arch")}
      <div class="section-head" style="margin-top:2.4rem">
        <h2>How the cluster is organised</h2>
        <p class="muted">From the Diopside design chapter: views, virtual identities, policies and GEO scheduling.</p>
      </div>
      <div class="grid grid-2">
        <article class="card set-card">
          <p class="meta">Views</p>
          <h3>Space, node, group, filesystem</h3>
          <p>The <em>fs</em> view lists every configured filesystem. The <em>node</em> view groups them by FST host. The <em>group</em> view is the scheduling (placement) group. The <em>space</em> view is the pool. A filesystem is a storage path, a host, an internal ID from 1 to 65534, and a UUID written on the disk.</p>
        </article>
        <article class="card set-card">
          <p class="meta">Mapping</p>
          <h3>Virtual identities</h3>
          <p>Each client is mapped, from its authentication method and vid rules, to a virtual identity — a uid/gid pair that owns files and directories. Roles can be attached so a person or service may act on behalf of everyone, or of a subset of identities.</p>
        </article>
        <article class="card set-card">
          <p class="meta">Policies</p>
          <h3>How a file is stored</h3>
          <p>Layout, checksums and placement can be set on a space, on an application, group or user, on a directory, or on a single URL. A directory can force erasure coding — for example <code>sys.forced.layout=raid6</code> with 12 stripes in an <code>erasure</code> space.</p>
        </article>
        <article class="card set-card">
          <p class="meta">GEO</p>
          <h3>Placement close to the client</h3>
          <p>GEO tags can be assigned to client IPs and to FST nodes. Placement policies match the two, so a file can be stored as close as possible to the reader — or kept apart for resilience.</p>
        </article>
      </div>
      <div class="section-head" style="margin-top:2.4rem">
        <h2>Microservices</h2>
        <p class="muted">Every instance ships a set of configurable engines that keep the cluster balanced and consistent.</p>
      </div>
      <div class="micro-grid">
        <article class="set-card"><h3>Balancers</h3><p>Filesystem balancer inside a group, group balancer inside a space, and geo balancer across locations.</p></article>
        <article class="set-card"><h3>Converter</h3><p>Queued jobs that change how a file is stored — for example from one replica to two-fold replication.</p></article>
        <article class="set-card"><h3>Lifecycle</h3><p>Automation for disk and node replacement: empty a filesystem that should leave production.</p></article>
        <article class="set-card"><h3>LRU engine</h3><p>Scans the namespace to apply clean-up or conversion policies, such as a scratch space that expires in 30 days.</p></article>
        <article class="set-card"><h3>Inspector</h3><p>Accounting of how files are stored — how much of an instance is replicated versus erasure-coded.</p></article>
        <article class="set-card"><h3>Consistency</h3><p>Distributed check and repair of data and metadata inconsistencies.</p></article>
        <article class="set-card"><h3>Workflow</h3><p>Event queue for external systems — typically CTA, when a new file should migrate to tape.</p></article>
      </div>
      <p class="tech-docs"><a href="${esc(archDocs)}" target="_blank" rel="noreferrer">Read the Design &amp; Architecture chapter</a> in the <a href="${esc(docs)}" target="_blank" rel="noreferrer">Diopside documentation</a>.</p>
      <div style="margin-top:1.6rem">${cardGrid(cards("link"), "grid-3")}</div>
    </div>`;
}

function docsPage() {
  return `
    <div class="wrap">
      <p class="kicker">Documentation</p>
      <h2>Find your way around EOS</h2>
      <p class="muted">Curated sections from <a href="${esc(setting("docs_url"))}" target="_blank" rel="noreferrer">eos-docs.web.cern.ch</a>. Search the full extracted manual below or on the talks page.</p>
      <form class="search-box" data-search="docs">
        <input name="q" placeholder="quota, FUSE, erasure coding, helm…" />
        <button class="btn" type="submit">Search docs</button>
      </form>
      <div id="doc-results"></div>
      <div class="section-head"><h2>Start here</h2></div>
      ${cardGrid(cards("help").length ? cards("help") : (catalog.docs || []).map((d) => ({ title: d.title, body: d.summary, href: d.url, meta: d.section })), "grid-2")}
    </div>`;
}

function workshops() {
  const list = (catalog.workshops || []).filter((w) => w.kind !== "conference");
  const rows = list.map((w) => {
    const note = w.description || "";
    return `
      <article class="pub-row">
        <span class="pub-year">${esc(String(w.year))}</span>
        <div class="pub-main">
          <h3>${w.url ? `<a href="${esc(w.url)}" target="_blank" rel="noreferrer">${esc(w.title)}</a>` : esc(w.title)}${w.edition ? ` <span class="muted">· ${esc(ordinal(w.edition))}</span>` : ""}</h3>
          <p class="pub-venue">${esc(w.start)} – ${esc(w.end)} · ${esc(w.location)}${w.kind ? ` · ${esc(w.kind)}` : ""}</p>
          ${note ? `<p class="pub-note">${esc(note)}</p>` : ""}
          ${w.url ? `<p class="pub-links"><a href="${esc(w.url)}" target="_blank" rel="noreferrer">Indico</a></p>` : ""}
        </div>
      </article>`;
  }).join("");
  return `
    <div class="wrap">
      <p class="kicker">Community</p>
      <h2>EOS workshops</h2>
      <p class="muted">${esc(catalog.index?.queryHelp || "")}. Search titles, abstracts, and speakers — then open slides or the CERN recording.</p>
      <p><a class="btn" data-nav href="/search">Search presentations</a></p>
      ${rows ? `<div class="pub-table" style="margin-top:1.2rem">${rows}</div>` : `<p class="empty">No workshops indexed yet.</p>`}
    </div>`;
}

function searchPage() {
  const params = new URLSearchParams(location.search);
  const q = params.get("q") || "";
  const kind = params.get("kind") || "all";
  const year = params.get("year") || "";
  const years = (catalog.years || []).map((y) => `<option value="${y}" ${String(y) === year ? "selected" : ""}>${y}</option>`).join("");
  return `
    <div class="wrap">
      <p class="kicker">Search</p>
      <h2>Presentations and documentation</h2>
      <p class="muted">${esc(catalog.index?.queryHelp || "Index of EOS workshop talks and eos-docs.")}</p>
      <form class="search-box" data-search="all" id="search-form">
        <input name="q" value="${esc(q)}" placeholder="FUSE, QuarkDB, CTA, site report, recycle bin…" />
        <select name="kind">
          <option value="all" ${kind === "all" ? "selected" : ""}>Talks + docs</option>
          <option value="talks" ${kind === "talks" ? "selected" : ""}>Talks only</option>
          <option value="docs" ${kind === "docs" ? "selected" : ""}>Docs only</option>
        </select>
        <select name="year">
          <option value="">All years</option>
          ${years}
        </select>
        <button class="btn" type="submit">Search</button>
      </form>
      <div id="search-results"><p class="muted">Loading talks…</p></div>
    </div>`;
}

const publications = [
  { set: "eos", year: 2025, title: "Advancements and Operations for LHC Run-3 and beyond",
    authors: "G. Amadio, M. Arsuaga Rios, C. Caffy, G. Del Monte, A. Lekshmanan, L. Mascetti, A.J. Peters, E.A. Sindrilaru, D. Smith, I. Vrachnaki",
    venue: "EPJ Web of Conferences 337, 01337 (2025)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202533701337" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2945900" },
    ] },
  { set: "eos", year: 2025, title: "ROOT RNTuple and EOS: The Next Generation of Event Data I/O",
    authors: "J. Blomer, A.J. Peters, G. Amadio et al.",
    venue: "EPJ Web of Conferences 337, 01324 (2025)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202533701324" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2946566" },
    ] },
  { set: "eos", year: 2024, title: "Operation of the CERN disk storage infrastructure during LHC Run-3",
    authors: "C. Caffy, G. Amadio, M. Arsuaga Rios et al.",
    venue: "EPJ Web of Conferences 295, 01041 (2024)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202429501041" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2919576" },
    ] },
  { set: "eos", year: 2024, title: "XRootD Client: A robust technology for LHC Run-3 and beyond",
    authors: "G. Amadio, C. Caffy, A.B. Hanushevsky, M.K. Simon, D. Smith",
    venue: "EPJ Web of Conferences 295, 01056 (2024)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202429501056" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2919257" },
    ] },
  { set: "eos", year: 2024, title: "I/O performance studies of analysis workloads on production and dedicated resources at CERN",
    authors: "",
    venue: "CHEP 2023 proceedings (2024)",
    note: "Study of analysis I/O workloads involving EOS, XCache, XRootD and FUSE.",
    links: [
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2919561" },
    ] },
  { set: "eos", year: 2020, title: "EOS architectural evolution and strategic development directions",
    authors: "G. Bitzes, F. Luchetti, A. Manzi, M. Patrascoiu, A.J. Peters, M.K. Simon, E.A. Sindrilaru",
    venue: "EPJ Web of Conferences 245, 04009 (2020)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202024504009" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2758818" },
    ] },
  { set: "eos", year: 2020, title: "Erasure Coding for production in the EOS Open Storage system",
    authors: "A.J. Peters, M.K. Simon, E.A. Sindrilaru",
    venue: "EPJ Web of Conferences 245, 04008 (2020)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202024504008" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2758817" },
      { label: "EPJ", href: "https://www.epj-conferences.org/articles/epjconf/abs/2020/21/epjconf_chep2020_04008/epjconf_chep2020_04008.html" },
    ] },
  { set: "eos", year: 2019, title: "Scaling the EOS namespace — new developments and performance optimizations",
    authors: "",
    venue: "EPJ Web of Conferences 214, 04019 (2019)",
    note: "Focuses on scaling the EOS namespace and the architecture that led to QuarkDB.",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/201921404019" },
    ] },
  { set: "eos", year: 2019, title: "Sharing server nodes for storage and compute",
    authors: "",
    venue: "EPJ Web of Conferences 214, 08025 (2019)",
    note: "Work involving the CERN EOS production infrastructure and combined storage/compute resources.",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/201921408025" },
      { label: "CERN", href: "https://repository.cern/legacy/record/2653012/files/" },
    ] },
  { set: "eos", year: 2019, title: "EOS Open Storage — evolution of an ecosystem for scientific data repositories",
    authors: "A.J. Peters, G. Bitzes, M.K. Simon, J. Makai, E.A. Sindrilaru",
    venue: "CHEP 2018",
    note: "The evolution of EOS from CERN’s original disk storage system toward a more general scientific data-storage platform.",
    links: [
      { label: "Indico", href: "https://indico.cern.ch/event/587955/contributions/3012718/" },
    ] },
  { set: "eos", year: 2017, title: "EOS Developments",
    authors: "A.J. Peters, E.A. Sindrilaru, G.M. Adde, P.H. Lensing",
    venue: "CHEP 2016",
    note: "Namespace redesign, CERN-wide filesystem access and object-storage integration.",
    links: [
      { label: "Indico", href: "https://indico.cern.ch/event/505613/contributions/2230904/" },
    ] },
  { set: "eos", year: 2015, title: "EOS as the present and future solution for data storage at CERN",
    authors: "A.J. Peters, E.A. Sindrilaru, G. Adde",
    venue: "J. Phys.: Conf. Ser. 664, 042042 (2015)",
    note: "High-level description of EOS for LHC Run II. In production at CERN since 2011, aimed at low-latency analysis and multi-PB installations.",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/664/4/042042" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2134573" },
      { label: "INSPIRE", href: "https://inspirehep.net/literature/1413872" },
    ] },
  { set: "eos", year: 2015, title: "CERNBox + EOS: end-user storage for science",
    authors: "L. Mascetti, H. González Labrador, M. Lamanna, J.T. Mościcki, A.J. Peters",
    venue: "J. Phys.: Conf. Ser. 664, 062037 (2015)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/664/6/062037" },
    ] },
  { set: "eos", year: 2015, title: "Latest evolution of EOS filesystem",
    authors: "G. Adde et al.",
    venue: "J. Phys.: Conf. Ser. 608, 012009 (2015)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/608/1/012009" },
    ] },
  { set: "eos", year: 2014, title: "Disk storage at CERN: Handling LHC data and beyond",
    authors: "X. Espinal et al.",
    venue: "J. Phys.: Conf. Ser. 513 (2014)",
    note: "CERN disk storage infrastructure, including EOS as the disk-only system for large-scale physics analysis.",
    links: [
      { label: "Indico", href: "https://indico.cern.ch/event/214784/contributions/1512846/" },
    ] },
  { set: "eos", year: 2012, title: "Evaluation of software based redundancy algorithms for the EOS storage system at CERN",
    authors: "A.J. Peters, E.A. Sindrilaru, P. Zigann",
    venue: "J. Phys.: Conf. Ser. 396, 042046 (2012)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/396/4/042046" },
    ] },
  { set: "eos", year: 2011, title: "Exabyte Scale Storage at CERN",
    authors: "Andreas J. Peters, Lukasz Janyst",
    venue: "J. Phys.: Conf. Ser. 331, 052015 (2011)",
    note: "A foundational paper for EOS: scaling LHC disk storage toward hundreds of millions of files and potentially exabyte-scale capacity.",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/331/5/052015" },
      { label: "INSPIRE", href: "https://inspirehep.net/files/fee406743e55f4af9b6a82703fb23d6d" },
    ] },
  { set: "xrootd", year: 2024, title: "XRootD Client: A robust technology for LHC Run-3 and beyond",
    authors: "G. Amadio, C. Caffy, A.B. Hanushevsky, M.K. Simon, D. Smith",
    venue: "EPJ Web of Conferences 295, 01056 (2024)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1051/epjconf/202429501056" },
      { label: "CERN CDS", href: "https://cds.cern.ch/record/2919257" },
    ] },
  { set: "xrootd", year: 2017, title: "XrootdFS: A Posix Filesystem for Xrootd",
    authors: "",
    venue: "2017",
    note: "POSIX-style filesystem access to XRootD storage." },
  { set: "xrootd", year: 2014, title: "Xrootd, disk-based, caching proxy for optimization of data access, data placement and data replication",
    authors: "",
    venue: "2014",
    note: "XRootD as a caching proxy and distributed data-access technology." },
  { set: "xrootd", year: 2012, title: "Using Xrootd to Federate Regional Storage",
    authors: "",
    venue: "2012",
    note: "XRootD-based federation of geographically distributed storage resources." },
  { set: "xrootd", year: 2010, title: "Scalla/Xrootd WAN globalization tools: Where we are",
    authors: "Fabrizio Furano, Andrew Hanushevsky",
    venue: "J. Phys.: Conf. Ser. 219, 072005 (2010)",
    links: [
      { label: "DOI", href: "https://doi.org/10.1088/1742-6596/219/7/072005" },
    ] },
  { set: "xrootd", year: 2005, title: "XROOTD/TXNetFile: a highly scalable architecture for data access in the ROOT environment",
    authors: "A. Dorigo, P. Elmer, F. Furano, A. Hanushevsky",
    venue: "2005",
    note: "An early foundational paper describing the architecture that became modern XRootD." },
];

function pubTable(set) {
  const rows = publications.filter((p) => p.set === set);
  let lastYear = 0;
  return `<div class="pub-table">${rows.map((p) => {
    const yearHead = p.year !== lastYear ? `<div class="pub-year-band">${p.year}</div>` : "";
    lastYear = p.year;
    const links = (p.links || []).map((l) => `<a href="${esc(l.href)}" target="_blank" rel="noreferrer">${esc(l.label)}</a>`).join("");
    return `${yearHead}
      <article class="pub-row">
        <span class="pub-year">${p.year}</span>
        <div class="pub-main">
          <h3>${esc(p.title)}</h3>
          ${p.authors ? `<p class="pub-authors">${esc(p.authors)}</p>` : ""}
          <p class="pub-venue">${esc(p.venue)}</p>
          ${p.note ? `<p class="pub-note">${esc(p.note)}</p>` : ""}
          ${links ? `<p class="pub-links">${links}</p>` : ""}
        </div>
      </article>`;
  }).join("")}</div>`;
}

function resources() {
  return `
    <div class="wrap">
      <p class="kicker">Resources</p>
      <h2>Documentation, papers, status</h2>
      ${cardGrid(cards("resource"), "grid-3")}
      <section class="pubs" id="publications">
        <div class="section-head">
          <h2>Publications</h2>
          <p class="muted">Reverse-chronological papers on EOS at CERN — architecture, operations — and XRootD where it is part of the EOS data path.</p>
        </div>
        <h3 class="pub-set">EOS storage at CERN</h3>
        ${pubTable("eos")}
        <h3 class="pub-set">XRootD</h3>
        <p class="muted pub-set-lede">Useful when considering XRootD on its own, and its later integration with EOS and WLCG storage.</p>
        ${pubTable("xrootd")}
      </section>
    </div>`;
}

function service() {
  return `
    <div class="wrap">
      <p class="kicker">CERN services</p>
      <h2>EOS in production</h2>
      ${cardGrid(cards("service"), "grid-2")}
    </div>
    <section class="stats-band">
      <div class="wrap">
        <div class="section-head">
          <p class="kicker">Scale</p>
          <h2>At CERN in numbers</h2>
        </div>
        ${stats("stats-inset")}
      </div>
    </section>`;
}

function news() {
  const items = (catalog.news || []).map((n) => `
    <article class="card set-card">
      <p class="meta">${esc(n.dateLabel)}</p>
      <h3>${esc(n.title)}</h3>
      <p>${esc(n.body)}</p>
      ${n.href ? `<a class="more" ${n.href.startsWith("/") ? "data-nav" : 'target="_blank" rel="noreferrer"'} href="${esc(n.href)}">Read more →</a>` : ""}
    </article>`).join("");
  return `<div class="wrap"><p class="kicker">News</p><h2>Latest</h2><div class="grid grid-2">${items}</div></div>`;
}

function personMarks(p) {
  const role = p.role || "";
  const marks = [];
  if (/core/i.test(role)) marks.push("Core");
  if (/project lead/i.test(role)) marks.push("Project Lead");
  if (/operations/i.test(role)) marks.push("Operations");
  if (/openlab/i.test(role)) marks.push("openlab");
  return marks;
}

function personFace(p) {
  const marks = personMarks(p);
  return `
    <h3>${esc(p.name)}</h3>
    <p class="muted">${esc(p.role)}</p>
    ${marks.length ? `<p class="person-marks">${marks.map((m) => `<span>${esc(m)}</span>`).join("")}</p>` : ""}
    ${p.email ? `<a href="mailto:${esc(p.email)}">${esc(p.email)}</a>` : ""}`;
}

function personCard(p) {
  const face = personFace(p);
  return `
    <div class="flip">
      <div class="flip-inner">
        <article class="card person flip-face">${face}</article>
        <article class="card person flip-face is-back" aria-hidden="true">${face}</article>
      </div>
    </div>`;
}

function isCorePerson(p) {
  return /core/i.test(p.role || "");
}

function community() {
  const people = catalog.people || [];
  const core = people.filter(isCorePerson).map(personCard).join("");
  const ops = people.filter((p) => !isCorePerson(p)).map(personCard).join("");
  return `
    <div class="wrap">
      <p class="kicker">Community</p>
      <h2>${esc(page("support").title || "Get in contact")}</h2>
      <div class="prose"><p>${esc(page("support").body)}</p></div>
      <p><a class="btn" href="mailto:${esc(setting("contact_email"))}">${esc(setting("contact_email"))}</a></p>
      <div class="section-head" style="margin-top:2rem"><h2>Core development team</h2></div>
      <div class="grid grid-3 team-grid">${core}</div>
      <div class="section-head" style="margin-top:2rem"><h2>Operations Lead &amp; openlab</h2></div>
      <div class="grid grid-3 team-grid">${ops}</div>
      <div class="section-head" style="margin-top:2rem"><h2>Collaborations</h2></div>
      ${cardGrid(cards("collab"), "grid-2")}
      <form class="card" id="contact-form" style="margin-top:2rem;display:grid;gap:0.7rem">
        <h3>Write to the project</h3>
        <input name="name" required placeholder="Name" />
        <input name="email" type="email" required placeholder="Email" />
        <textarea name="message" required placeholder="Message" rows="4"></textarea>
        <button class="btn" type="submit">Send</button>
        <p id="contact-msg" class="muted" hidden></p>
      </form>
    </div>`;
}

function roadmapAreas() {
  return [
    {
      title: "Storage architecture",
      lede: "Flexible hierarchies inside EOS — cache, disk, erasure coding and tape as one platform.",
      items: [
        { title: "Read-through spaces & tiering", body: "Move data between spaces by policy — SSD to HDD, replica to erasure coding, hot to cold — and treat one space as a transparent cache in front of another, with admission, eviction, pinning and prefetch." },
        { title: "Native Mirage I/O", body: "Simulate file contents, checksums and the network path without touching disks, so protocol, gateway and scalability tests no longer depend on physical storage." },
        { title: "Native tape", body: "Investigate tape as an EOS storage tier rather than an external CTA buffer: lifecycle, archive and retrieve from EOS, with a path for today’s EOSCTA sites." },
        { title: "Erasure coding in production", body: "File updates, small files, degraded reads, faster repair and online conversion — so EC can sit in a hot → replica → EC → tape lifecycle." },
      ],
    },
    {
      title: "High-performance access",
      lede: "Standards and transports for HPC, Kubernetes and GPU-heavy analysis.",
      items: [
        { title: "Production NFSv4.1", body: "Locking, caching, identity mapping and monitoring where eosxd is hard to deploy — Kubernetes, HPC and sites that need a standard filesystem." },
        { title: "XrdHttp, GPU and RDMA", body: "Kernel clients, low-copy paths into GPU memory and RDMA, with common capability discovery for AI/ML and high-throughput analysis." },
        { title: "S3 as another EOS protocol", body: "Re-think the gateway model so object access shares more of EOS instead of a separate Versity-shaped stack." },
      ],
    },
    {
      title: "Operations & reliability",
      lede: "One picture of the instance, and data movement that protects client traffic.",
      items: [
        { title: "Controller UI", body: "A control plane for health, spaces, nodes, drain, quotas and alerts — safe workflows with RBAC, and later a fleet view across instances." },
        { title: "Traffic shaping & draining", body: "Load-aware limits for drain, balance, convert and repair so foreground I/O stays protected and failing hardware can still be emptied first." },
        { title: "XRootD 6.3 monitoring", body: "Adopt the new XRootD telemetry, keep EOS-specific metrics beside it, and retire obsolete dashboards." },
      ],
    },
    {
      title: "Infrastructure & efficiency",
      lede: "Data locality and energy as first-class placement inputs.",
      items: [
        { title: "Hyper-converged nodes", body: "Run FSTs on compute nodes and prefer local or rack-local data for analysis, HPC and AI/ML instead of treating storage and compute as strictly separate." },
        { title: "PROMISE spin-down", body: "Concentrate active data so whole groups of disks can idle. Placement that understands power state, with hysteresis so drives are not woken for background noise." },
      ],
    },
    {
      title: "Platform & ecosystem",
      lede: "One identity model, a stable management API, and EOS that fits Kubernetes.",
      items: [
        { title: "Common OIDC", body: "Push XrdSecOIDC into XRootD and use it for FUSE, HTTP and gateways — one token and identity story instead of a different one per access path." },
        { title: "Management API", body: "A documented API for the Controller UI, automation and monitoring, so tools stop scraping the CLI." },
        { title: "Cloud-native EOS", body: "Production CSI, container-friendly auth and simpler charts so smaller sites and Kubernetes workloads can consume EOS cleanly." },
      ],
    },
  ];
}

function roadmapMilestones() {
  return [
    { year: "2009", label: "R&D year", note: "Architecture study before the project" },
    { year: "2010", label: "Project starts", note: "LHC analysis use-case, 2 FTEs" },
    { year: "2011", label: "LHC instances", note: "ATLAS, CMS, ALICE, LHCb" },
    { year: "2015", label: "CERNBox", note: "Sync-and-share on EOS" },
    { year: "2018", label: "QuarkDB", note: "Namespace persistency replaces in-memory" },
    { year: "2021", label: "Diopside", note: "EOS 5 production line" },
    { year: "2023", label: "Exabyte", note: "CERN disk volume, 500× growth" },
    { year: "2026", label: "Now", note: "The programme below" },
  ];
}

function roadmapSketch() {
  const marks = roadmapMilestones();
  return `
    <figure class="road-sketch">
      <figcaption>
        <strong>Sixteen years of history</strong>
        <span>Major marks from the 2026 workshop timeline — then the work ahead.</span>
      </figcaption>
      <ol class="road-line">
        ${marks.map((m, i) => `
          <li class="${i === marks.length - 1 ? "is-now" : ""}">
            <b>${esc(m.year)}</b>
            <i></i>
            <strong>${esc(m.label)}</strong>
            <span>${esc(m.note)}</span>
          </li>`).join("")}
      </ol>
    </figure>`;
}

function roadmap() {
  const p = page("roadmap");
  const lead = p.body || "The EOS development programme covers production improvements, architectural evolution and selected R&D.";
  return `
    <div class="wrap roadmap-page">
      <p class="kicker">Development programme</p>
      <h2>${esc(p.title || "Roadmap")}</h2>
      <div class="prose">${paragraphs(lead)}</div>
      <p class="roadmap-aim">${highlightLead("The direction is a single EOS that caches, places and archives by policy — disk, erasure coding and tape together — and that can exploit locality and energy on future CERN computing farms.")}</p>
      ${roadmapSketch()}
      ${roadmapAreas().map((area) => `
        <section class="roadmap-area">
          <div class="section-head">
            <h2>${esc(area.title)}</h2>
            <p class="muted">${esc(area.lede)}</p>
          </div>
          <div class="grid grid-2">${area.items.map((item) => `
            <article class="card set-card">
              <h3>${esc(item.title)}</h3>
              <p>${esc(item.body)}</p>
            </article>`).join("")}</div>
        </section>`).join("")}
    </div>`;
}

function notFound() {
  return `<div class="wrap"><h2>Not found</h2><p class="muted">That page is not part of the EOS site.</p><p><a data-nav href="/">Back home</a></p></div>`;
}

function view() {
  switch (path()) {
    case "/": return home();
    case "/about": return about();
    case "/tech": return tech();
    case "/roadmap": return roadmap();
    case "/docs": return docsPage();
    case "/workshops": return workshops();
    case "/search": return searchPage();
    case "/resources": return resources();
    case "/service": return service();
    case "/news": return news();
    case "/community": return community();
    default: return notFound();
  }
}

function talkTable(talks) {
  let lastYear = 0;
  return `<div class="pub-table">${(talks || []).map((t) => {
    const year = Number(t.year) || 0;
    const yearHead = year && year !== lastYear ? `<div class="pub-year-band">${year}</div>` : "";
    lastYear = year;
    const links = [
      t.url ? `<a href="${esc(t.url)}" target="_blank" rel="noreferrer">Indico</a>` : "",
      t.slides ? `<a href="${esc(t.slides)}" target="_blank" rel="noreferrer">Slides</a>` : "",
      t.recording ? `<a href="${esc(t.recording)}" target="_blank" rel="noreferrer">Recording</a>` : "",
    ].filter(Boolean).join("");
    const note = t.abstract || t.session || "";
    return `${yearHead}
      <article class="pub-row">
        <span class="pub-year">${year || "—"}</span>
        <div class="pub-main">
          <h3>${t.url ? `<a href="${esc(t.url)}" target="_blank" rel="noreferrer">${esc(t.title)}</a>` : esc(t.title)}</h3>
          ${t.speakers ? `<p class="pub-authors">${esc(t.speakers)}</p>` : ""}
          <p class="pub-venue">${esc(t.workshopTitle || "EOS workshop")}${t.session ? ` · ${esc(t.session)}` : ""}</p>
          ${note ? `<p class="pub-note">${esc(note)}</p>` : ""}
          ${links ? `<p class="pub-links">${links}</p>` : ""}
        </div>
      </article>`;
  }).join("")}</div>`;
}

function renderHits(data) {
  const talks = data.talks || [];
  const docs = data.docs || [];
  if (!talks.length && !docs.length) {
    return `<p class="empty">${data.query ? `No matches for “${esc(data.query)}”.` : "No talks indexed yet."}</p>`;
  }
  const talkBlock = talks.length ? `
    <div class="section-head" style="margin-top:0.4rem">
      <h2>${data.query ? "Matching talks" : "All workshop talks"}</h2>
      <p class="muted">${talks.length} ${talks.length === 1 ? "talk" : "talks"}${data.query ? ` for “${esc(data.query)}”` : ", newest first. Hover a row for the abstract."}</p>
    </div>
    ${talkTable(talks)}` : "";
  const docBlock = docs.length ? `
    <h3 class="pub-set">Documentation</h3>
    ${docs.map((d) => `
      <article class="hit">
        <p class="src">${esc(d.section)}</p>
        <h3>${d.url ? `<a href="${esc(d.url)}" target="_blank" rel="noreferrer">${esc(d.title)}</a>` : esc(d.title)}</h3>
        <p>${esc(d.summary || d.body || "")}</p>
      </article>`).join("")}` : "";
  return `${talkBlock}${docBlock}`;
}

async function runSearch(scope, form) {
  const fd = new FormData(form);
  const q = String(fd.get("q") || "").trim();
  const kind = String(fd.get("kind") || scope || "all");
  const year = String(fd.get("year") || "");
  if (scope === "all") {
    const next = new URL("/search", location.origin);
    next.searchParams.set("q", q);
    next.searchParams.set("kind", kind);
    if (year) next.searchParams.set("year", year);
    history.replaceState({}, "", next.pathname + next.search);
  }
  const params = new URLSearchParams({ q, kind });
  if (year) params.set("year", year);
  const res = await fetch("/api/search?" + params.toString());
  const data = await res.json();
  const box = scope === "docs" ? $("#doc-results") : $("#search-results");
  if (box) box.innerHTML = renderHits(data);
}

function bindPage() {
  $("#app").querySelectorAll("[data-search]").forEach((form) => {
    form.addEventListener("submit", (e) => {
      e.preventDefault();
      runSearch(form.dataset.search, form).catch((err) => {
        const box = $("#search-results") || $("#doc-results");
        if (box) box.innerHTML = `<p class="empty">${esc(err.message)}</p>`;
      });
    });
  });
  if (path() === "/search") {
    const form = $("#search-form");
    if (form) runSearch("all", form);
  }
  const contact = $("#contact-form");
  if (contact) {
    contact.addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(contact);
      const res = await fetch("/api/inbox", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: fd.get("name"), email: fd.get("email"), message: fd.get("message") }),
      });
      const data = await res.json().catch(() => ({}));
      const msg = $("#contact-msg");
      msg.hidden = false;
      msg.textContent = res.ok ? "Thanks — the message is in the controller inbox." : (data.error || "Could not send");
      msg.className = res.ok ? "ok" : "muted";
    });
  }
}

function render() {
  stopTitleSpray();
  stopCapacityChart();
  stopTermType();
  stopAboutHighlight();
  setNav();
  document.title = path() === "/" ? "EOS Open Storage" : `EOS · ${path().slice(1)}`;
  $("#app").innerHTML = view();
  bindPage();
  document.querySelector(".top")?.classList.remove("is-menu");
  if (location.hash) requestAnimationFrame(scrollToHash);
  if (path() === "/" || path() === "/about") startAboutHighlight();
  if (path() !== "/") return;
  startTermType();
  startHeroBackground();
  startOrbitMovie();
  startCapacityChart();
  const kick = () => {
    if (path() === "/") startTitleSpray();
  };
  if (document.fonts?.ready) document.fonts.ready.then(kick);
  else kick();
}

async function load() {
  const res = await fetch("/api/catalog");
  catalog = await res.json();
  $("#foot-address").textContent = setting("address");
  const mail = $("#foot-mail");
  mail.href = "mailto:" + setting("contact_email");
  mail.textContent = setting("contact_email");
  const gl = $("#repo-gitlab");
  const gh = $("#repo-github");
  if (gl && setting("gitlab")) gl.href = setting("gitlab");
  if (gh && setting("github")) gh.href = setting("github");
  render();
}

document.addEventListener("click", (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a");
  if (!a || a.target === "_blank" || a.hasAttribute("download")) return;
  const href = a.getAttribute("href");
  if (!href || /^(mailto:|tel:|#)/.test(href)) return;
  const u = new URL(href, location.origin);
  if (u.origin !== location.origin) return;
  if (/^\/(api|static|media|controller)(\/|$)/.test(u.pathname)) return;
  e.preventDefault();
  document.querySelector(".top")?.classList.remove("is-menu");
  go(u.pathname + u.search + u.hash);
});

$("[data-menu]")?.addEventListener("click", () => {
  const top = document.querySelector(".top");
  top.classList.toggle("is-menu");
  $("[data-menu]").setAttribute("aria-expanded", top.classList.contains("is-menu") ? "true" : "false");
});

$("#news-form")?.addEventListener("submit", async (e) => {
  e.preventDefault();
  const email = $("#news-email").value;
  const res = await fetch("/api/newsletter", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  const data = await res.json().catch(() => ({}));
  const msg = $("#news-msg");
  msg.hidden = false;
  msg.textContent = res.ok ? "Subscribed — we will confirm from the controller." : (data.error || "Could not subscribe");
});

window.addEventListener("popstate", () => {
  if (prefersQuiet() || pageBusy) {
    render();
    scrollToHash();
    return;
  }
  fadeTo(() => {
    render();
    scrollToHash();
  });
});
load().catch((err) => {
  $("#app").innerHTML = `<div class="wrap"><p class="empty">Could not load the site: ${esc(err.message)}</p></div>`;
});
