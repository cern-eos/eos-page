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
        window.setTimeout(afterIn, 210);
      });
    });
  };
  app.addEventListener("transitionend", afterOut);
  window.setTimeout(afterOut, 190);
}

function diopsideFigure() {
  return `
    <figure class="diopside" title="Faceted Diopside, Madagascar - Didier Descouens, Wikimedia Commons, CC BY-SA 4.0">
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
      <svg class="eos-growth-svg" viewBox="0 0 720 240" role="img" aria-label="EOS raw capacity at CERN from about 5 PB in 2010 to a 2.5 EB target in 2030. After 2025 the path is an open quantum-like band that only pins the 2030 arrival.">
        <defs>
          <linearGradient id="growth-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#e65c00" stop-opacity="0.4"/>
            <stop offset="100%" stop-color="#e65c00" stop-opacity="0"/>
          </linearGradient>
          <linearGradient id="growth-stroke" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#f0a060"/>
            <stop offset="100%" stop-color="#fff6e8"/>
          </linearGradient>
          <linearGradient id="growth-quantum-fill" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#ffb06a" stop-opacity="0.28"/>
            <stop offset="45%" stop-color="#ffc48a" stop-opacity="0.58"/>
            <stop offset="100%" stop-color="#e65c00" stop-opacity="0.22"/>
          </linearGradient>
          <filter id="growth-quantum-fuzz" x="-18%" y="-40%" width="136%" height="180%">
            <feTurbulence type="fractalNoise" baseFrequency="0.04 0.12" numOctaves="2" seed="11" result="n"/>
            <feDisplacementMap in="SourceGraphic" in2="n" scale="4.5" xChannelSelector="R" yChannelSelector="G"/>
            <feGaussianBlur stdDeviation="1.1"/>
          </filter>
          <filter id="growth-quantum-soft" x="-10%" y="-20%" width="120%" height="140%">
            <feGaussianBlur stdDeviation="1.4"/>
          </filter>
        </defs>
        <g class="eos-growth-plot">
          <path class="eos-growth-area"></path>
          <path class="eos-growth-quantum-outer"></path>
          <path class="eos-growth-quantum-inner"></path>
          <path class="eos-growth-quantum-ghosts"></path>
        </g>
        <g class="eos-growth-grid"></g>
        <g class="eos-growth-plot-line">
          <path class="eos-growth-line"></path>
        </g>
        <g class="eos-growth-marks"></g>
        <circle class="eos-growth-dot" r="4.6" cx="0" cy="0"></circle>
      </svg>
      <p class="eos-growth-note">One point per year through 2025. Values with ~ are approximate. After 2025 the path is unknown - only the 2.5 EB arrival in 2030 is fixed.</p>
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

function sampleAlong(pts, t) {
  if (!pts.length) return { x: 0, y: 0, year: 0, pb: 0, t };
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

function bandPath(up, lo) {
  if (up.length < 2 || lo.length < 2) return "";
  let d = `M ${up[0].x.toFixed(2)} ${up[0].y.toFixed(2)}`;
  for (let i = 1; i < up.length; i++) d += ` L ${up[i].x.toFixed(2)} ${up[i].y.toFixed(2)}`;
  for (let i = lo.length - 1; i >= 0; i--) d += ` L ${lo[i].x.toFixed(2)} ${lo[i].y.toFixed(2)}`;
  return `${d} Z`;
}

function quantumEnvelope(s, xOf, yOf, floorPB = 1200, ceilPB = 2500) {
  if (s <= 0.001) return { outer: "", inner: "", ghosts: "", lo: floorPB, hi: floorPB };
  const clampPB = (pb) => Math.min(ceilPB, Math.max(floorPB, pb));
  const n = Math.max(12, Math.round(32 * Math.max(s, 0.12)));
  const upper = [];
  const lower = [];
  const innerU = [];
  const innerL = [];
  const ghosts = [[], [], [], []];
  let lo = ceilPB;
  let hi = floorPB;
  const span = ceilPB - floorPB;
  for (let i = 0; i <= n; i++) {
    const u = (i / n) * s;
    const x = xOf(EOS_HINGE_YEAR + 5 * u);
    const mid = floorPB + span * u;
    const bulge = Math.sin(Math.PI * u);
    const top = clampPB(mid + 740 * bulge);
    const bot = clampPB(mid - 700 * bulge);
    lo = Math.min(lo, bot);
    hi = Math.max(hi, top);
    upper.push({ x, y: yOf(top) });
    lower.push({ x, y: yOf(bot) });
    innerU.push({ x, y: yOf(clampPB(mid + 300 * bulge)) });
    innerL.push({ x, y: yOf(clampPB(mid - 270 * bulge)) });
    ghosts[0].push({ x, y: yOf(clampPB(mid + 520 * bulge)) });
    ghosts[1].push({ x, y: yOf(clampPB(mid - 470 * bulge)) });
    ghosts[2].push({ x, y: yOf(clampPB(floorPB + span * (u * u) + 90 * bulge)) });
    ghosts[3].push({ x, y: yOf(clampPB(floorPB + span * (1 - (1 - u) * (1 - u)) - 70 * bulge)) });
  }
  return {
    outer: bandPath(upper, lower),
    inner: bandPath(innerU, innerL),
    ghosts: ghosts.map((pts) => curvePath(pts)).filter(Boolean).join(" "),
    lo,
    hi,
  };
}

function formatQuantumRange(lo, hi) {
  if (hi - lo < 90) return "2.5 EB";
  const fmt = (pb) => (pb >= 1000 ? `${(pb / 1000).toFixed(1)} EB` : `${Math.round(pb)} PB`);
  return `~${fmt(lo)}–${fmt(hi)}`;
}

const GROWTH_HIST_SHARE = 0.68;

function growthVisible(t) {
  const { pts, pad, innerH, xOf, yOf } = growthLayout();
  const hist = pts.filter((p) => !p.target);
  const hinge = hist[hist.length - 1];
  const target = pts[pts.length - 1];
  const done = t >= 1;
  const inQuantum = done || t > GROWTH_HIST_SHARE;
  const histT = done || inQuantum ? 1 : t / GROWTH_HIST_SHARE;
  const s = done ? 1 : inQuantum ? (t - GROWTH_HIST_SHARE) / (1 - GROWTH_HIST_SHARE) : 0;
  const walk = sampleAlong(hist, histT);
  const quantum = quantumEnvelope(s, xOf, yOf, hinge.pb, target.pb);
  const vis = [];
  const u = histT * (hist.length - 1);
  const lastIdx = Math.min(hist.length - 1, Math.floor(u + 1e-6));
  vis.push(...hist.slice(0, lastIdx + 1));
  if (histT < 1 && (!vis.length || Math.abs(vis[vis.length - 1].x - walk.x) > 0.05)) vis.push(walk);
  const line = vis.length >= 2 ? curvePath(vis) : `M ${walk.x.toFixed(2)} ${walk.y.toFixed(2)}`;
  const areaEnd = histT >= 1 ? hinge : walk;
  const base = (pad.t + innerH).toFixed(2);
  const area = vis.length
    ? `${line} L ${areaEnd.x.toFixed(2)} ${base} L ${hist[0].x.toFixed(2)} ${base} Z`
    : "";
  const year = inQuantum ? EOS_HINGE_YEAR + 5 * s : walk.year;
  const xNow = inQuantum ? xOf(year) : walk.x;
  const pbNow = inQuantum ? (done ? 2500 : quantum.hi) : walk.pb;
  return {
    line,
    area,
    quantum,
    sample: walk,
    hinge,
    target,
    done,
    inQuantum,
    s,
    year,
    xNow,
    pbNow,
  };
}

function applyGrowth(root, t) {
  const { pad, innerH } = growthLayout();
  const view = growthVisible(t);
  const lineEl = root.querySelector(".eos-growth-line");
  const areaEl = root.querySelector(".eos-growth-area");
  const outerEl = root.querySelector(".eos-growth-quantum-outer");
  const innerEl = root.querySelector(".eos-growth-quantum-inner");
  const ghostEl = root.querySelector(".eos-growth-quantum-ghosts");
  const dot = root.querySelector(".eos-growth-dot");
  const yearEl = root.querySelector("[data-growth-year]");
  const pbEl = root.querySelector("[data-growth-pb]");
  if (lineEl) lineEl.setAttribute("d", view.line);
  if (areaEl) areaEl.setAttribute("d", view.area);
  if (outerEl) outerEl.setAttribute("d", view.quantum.outer);
  if (innerEl) innerEl.setAttribute("d", view.quantum.inner);
  if (ghostEl) ghostEl.setAttribute("d", view.quantum.ghosts);
  if (dot) {
    const pin = view.done ? view.target : view.sample;
    dot.setAttribute("cx", pin.x.toFixed(2));
    dot.setAttribute("cy", pin.y.toFixed(2));
    dot.classList.toggle("is-target", view.done);
  }
  if (yearEl) yearEl.textContent = String(Math.round(view.year));
  if (pbEl) {
    pbEl.textContent = view.done
      ? "2.5 EB"
      : view.inQuantum
        ? formatQuantumRange(view.quantum.lo, view.quantum.hi)
        : formatCapacity(view.sample.pb, view.sample);
  }
  const baseline = pad.t + innerH;
  root.querySelectorAll(".eos-growth-yline").forEach((el) => {
    const pb = Number(el.getAttribute("data-pb"));
    const rise = Math.max(0, Math.min(1, (view.pbNow + 60 - pb) / 180));
    el.setAttribute("x2", view.xNow.toFixed(2));
    el.style.opacity = rise > 0 ? String(0.25 + 0.75 * rise) : "0";
  });
  root.querySelectorAll(".eos-growth-grid text[data-pb]").forEach((el) => {
    const pb = Number(el.getAttribute("data-pb"));
    el.style.opacity = view.pbNow + 60 >= pb ? "1" : "0";
  });
  root.querySelectorAll(".eos-growth-xline").forEach((el) => {
    const year = Number(el.getAttribute("data-year"));
    const local = Math.max(0, Math.min(1, (view.year + 0.2 - year) / 0.55));
    el.setAttribute("y1", baseline.toFixed(2));
    el.setAttribute("y2", (baseline - local * innerH).toFixed(2));
    el.style.opacity = local > 0 ? "1" : "0";
  });
  root.querySelectorAll(".eos-growth-marks [data-year], .eos-growth-grid text.is-mark, .eos-growth-tick").forEach((el) => {
    const year = Number(el.getAttribute("data-year"));
    const ready = year <= EOS_HINGE_YEAR
      ? view.year + 0.15 >= year
      : view.s >= 0.96 || view.done;
    el.style.opacity = ready ? "1" : "0";
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

function bindTermCopy() {
  const term = document.querySelector("[data-copy-term]");
  if (!term || term.dataset.copyBound) return;
  term.dataset.copyBound = "1";
  term.addEventListener("click", () => copyTermCommand(term));
}

async function copyTermCommand(term) {
  if (!(await copyText(TERM_CLONE))) return;
  term.classList.remove("is-copied");
  void term.offsetWidth;
  term.classList.add("is-copied");
  window.clearTimeout(Number(term.dataset.copyTimer || 0));
  term.dataset.copyTimer = String(window.setTimeout(() => {
    term.classList.remove("is-copied");
  }, 1600));
}

async function copyText(text) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch (_) {}
  try {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    const ok = document.execCommand("copy");
    ta.remove();
    return ok;
  } catch (_) {
    return false;
  }
}

function startTermType() {
  stopTermType();
  bindTermCopy();
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

let logoSpinRaf = 0;
let logoSpinAc = null;

function stopLogoSpin() {
  if (logoSpinRaf) cancelAnimationFrame(logoSpinRaf);
  logoSpinRaf = 0;
  logoSpinAc?.abort();
  logoSpinAc = null;
}

let collideAc = null;

function stopCollision() {
  collideAc?.abort();
  collideAc = null;
}

let orbitAlignAc = null;

function stopOrbitAlign() {
  orbitAlignAc?.abort();
  orbitAlignAc = null;
}

function alignOrbitHex() {
  const logo = document.querySelector(".orbit-logo");
  const title = document.querySelector(".title-letter") || document.querySelector(".title-loupe");
  const slot = document.querySelector(".orbit-slot");
  if (!logo || !title || !slot) return;
  const titleR = title.getBoundingClientRect();
  const slotR = slot.getBoundingClientRect();
  const titleCy = titleR.top + titleR.height / 2;
  const slotCy = slotR.top + slotR.height / 2;
  const dy = titleCy - slotCy - logo.offsetHeight / 2;
  logo.style.translate = `-50% ${dy.toFixed(1)}px`;
}

function startOrbitAlign() {
  stopOrbitAlign();
  if (!document.querySelector(".orbit-logo")) return;
  orbitAlignAc = new AbortController();
  const run = () => {
    alignOrbitHex();
  };
  run();
  requestAnimationFrame(run);
  window.addEventListener("resize", run, { signal: orbitAlignAc.signal });
  document.fonts?.ready?.then(() => {
    if (document.querySelector(".orbit-logo")) run();
  });
}

function aimCollision() {
  const a = document.querySelector(".collide-aim.is-a");
  const b = document.querySelector(".collide-aim.is-b");
  if (!a || !b) return;
  const deg = Math.random() * 360;
  const skew = (Math.random() - 0.5) * 48;
  a.setAttribute("transform", `rotate(${deg.toFixed(1)} 200 100)`);
  b.setAttribute("transform", `rotate(${(deg + skew).toFixed(1)} 200 100)`);
}

function startCollision() {
  stopCollision();
  const bunch = document.querySelector(".collide-bunch.is-l");
  if (!bunch || prefersQuiet()) return;
  collideAc = new AbortController();
  aimCollision();
  bunch.addEventListener("animationiteration", aimCollision, { signal: collideAc.signal });
}

function collidePulse(anim) {
  if (!anim || anim.currentTime == null) return 1;
  const t = (anim.currentTime % 10000) / 1000 - 2.76;
  if (t < 0 || t > 0.82) return 1;
  return 1 - 0.48 * Math.exp(-4.4 * t) * Math.sin(20 * t);
}

function startLogoSpin() {
  stopLogoSpin();
  const img = document.querySelector(".orbit-logo img");
  const host = document.querySelector(".orbit-logo");
  if (!img || prefersQuiet()) return;
  const periodStart = 240;
  const periodEnd = 10;
  const rampSec = 10;
  const deg0 = 360 / periodStart;
  const deg1 = 360 / periodEnd;
  let angle = 0;
  let last = 0;
  let hovering = false;
  let boost = 1;
  let collideAnim = null;
  const t0 = performance.now();
  logoSpinAc = new AbortController();
  if (host) {
    host.addEventListener("mouseenter", () => { hovering = true; }, { signal: logoSpinAc.signal });
    host.addEventListener("mouseleave", () => { hovering = false; }, { signal: logoSpinAc.signal });
  }
  const tick = (now) => {
    if (!document.contains(img)) {
      stopLogoSpin();
      return;
    }
    if (!last) last = now;
    const dt = Math.min(0.05, (now - last) / 1000);
    last = now;
    const u = Math.min(1, (now - t0) / 1000 / rampSec);
    const ease = u * u;
    const target = hovering ? 4 : 1;
    boost += (target - boost) * (1 - Math.exp(-dt / 0.42));
    const degPerSec = (deg0 + (deg1 - deg0) * ease) * boost;
    angle = (angle + degPerSec * dt) % 360;
    if (!collideAnim || collideAnim.playState === "idle") {
      const bunch = document.querySelector(".collide-bunch.is-l");
      collideAnim = bunch?.getAnimations?.()[0] || null;
    }
    const pulse = collidePulse(collideAnim);
    img.style.transform = `rotate(${angle.toFixed(3)}deg) scale(${(0.15 * pulse).toFixed(4)})`;
    const hex = document.querySelector(".orbit-wind-hexmask");
    if (hex) hex.setAttribute("transform", `rotate(${angle.toFixed(3)} 100 100)`);
    logoSpinRaf = requestAnimationFrame(tick);
  };
  logoSpinRaf = requestAnimationFrame(tick);
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
  if (document.hidden) {
    const kick = () => {
      document.removeEventListener("visibilitychange", kick);
      if (path() === "/" && document.querySelector(".eos-growth")) startCapacityChart();
    };
    document.addEventListener("visibilitychange", kick, { once: true });
    buildCapacityChart(root);
    applyGrowth(root, 0);
    return;
  }
  buildCapacityChart(root);
  applyGrowth(root, 0);
  const duration = prefersQuiet() ? 4800 : 8600;
  const hold = 5200;
  const gen = growthGen;
  let t0 = 0;
  const step = (now) => {
    if (gen !== growthGen || !document.contains(root)) return;
    if (document.hidden) {
      growthRAF = 0;
      const resume = () => {
        document.removeEventListener("visibilitychange", resume);
        if (gen === growthGen && path() === "/") startCapacityChart();
      };
      document.addEventListener("visibilitychange", resume, { once: true });
      return;
    }
    if (!t0) t0 = now;
    const t = Math.min(1, (now - t0) / duration);
    applyGrowth(root, t);
    if (t < 1) {
      growthRAF = requestAnimationFrame(step);
      return;
    }
    applyGrowth(root, 1);
    growthRAF = 0;
    if (prefersQuiet()) return;
    growthHold = window.setTimeout(() => {
      if (gen === growthGen) startCapacityChart();
    }, hold);
  };
  growthRAF = requestAnimationFrame(step);
}

function ensureCapacityChart(force) {
  if (path() !== "/") return;
  if (!document.querySelector(".eos-growth") || document.hidden) return;
  if (!force && (growthRAF || growthHold)) return;
  startCapacityChart();
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
  const extra = 1;
  const duration = 2600;
  const painted = new Set();
  let lastSpark = 0;
  const letterBox = (j) => {
    const last = letters[letters.length - 1].getBoundingClientRect();
    if (j < letters.length) return letters[j].getBoundingClientRect();
    return {
      left: last.right + last.width * 0.06,
      top: last.top,
      width: last.width,
      height: last.height,
    };
  };
  const tick = (now) => {
    if (!document.body.contains(h1)) return;
    const t = Math.min(1, (now - start) / duration);
    const e = easeInOutQuad(t);
    const steps = letters.length + extra;
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

function hoverNavEnabled() {
  return window.matchMedia("(hover: hover) and (pointer: fine)").matches
    && !window.matchMedia("(max-width: 860px)").matches;
}

function bindNavHover() {
  const nav = document.querySelector(".nav");
  if (!nav || nav.dataset.hoverBound) return;
  nav.dataset.hoverBound = "1";
  let timer = 0;
  let pending = "";
  const clear = () => {
    window.clearTimeout(timer);
    timer = 0;
    pending = "";
  };
  const destOf = (a) => {
    if (!a || a.target === "_blank") return "";
    const href = a.getAttribute("href");
    if (!href || /^(mailto:|tel:|#)/.test(href)) return "";
    try {
      const u = new URL(href, location.origin);
      if (u.origin !== location.origin) return "";
      if (/^\/(api|static|media|controller)(\/|$)/.test(u.pathname)) return "";
      return u.pathname + u.search + u.hash;
    } catch (_) {
      return "";
    }
  };
  nav.addEventListener("mouseover", (e) => {
    if (!hoverNavEnabled()) return;
    const a = e.target.closest("a");
    if (!a || !nav.contains(a)) return;
    if (e.relatedTarget && a.contains(e.relatedTarget)) return;
    const next = destOf(a);
    if (!next || samePlace(next)) return;
    clear();
    pending = next;
    timer = window.setTimeout(() => {
      const dest = pending;
      clear();
      if (dest && hoverNavEnabled() && !samePlace(dest)) go(dest);
    }, 180);
  });
  nav.addEventListener("mouseout", (e) => {
    const a = e.target.closest("a");
    if (!a || !nav.contains(a)) return;
    if (e.relatedTarget && a.contains(e.relatedTarget)) return;
    if (destOf(a) === pending) clear();
  });
}

function latestNewsTicker() {
  const item = (catalog.news || [])[0];
  if (!item || !item.title) return "";
  const label = [item.dateLabel, item.title].filter(Boolean).join("  ·  ");
  const href = item.href || "/news";
  const nav = href.startsWith("/") ? " data-nav" : ` target="_blank" rel="noreferrer"`;
  const bit = `<span>${esc(label)}</span>`;
  return `<div class="news-ticker" aria-label="Latest news">
    <a class="news-ticker-link" href="${esc(href)}"${nav}>
      <span class="news-ticker-track">${bit}${bit}</span>
    </a>
  </div>`;
}

function windSwirlMarkup() {
  const spiralPts = (turns, rInner, rOuter, steps, phase) => {
    const pts = [];
    for (let i = 0; i <= steps; i++) {
      const u = i / steps;
      const t = phase + u * turns * Math.PI * 2;
      const r = rOuter + (rInner - rOuter) * u;
      pts.push({ u, r, x: 100 + r * Math.cos(t), y: 100 + r * Math.sin(t) });
    }
    return pts;
  };
  const hurricaneWidth = (u, outer, inner) => {
    const band = outer * (1 - u) ** 1.45;
    const eye = 2.8 * Math.exp(-(((u - 0.82) / 0.11) ** 2));
    return Math.max(0.7, band + inner + eye);
  };
  const spiralBand = (turns, rInner, rOuter, steps, phase, outerW, innerW) => {
    const pts = spiralPts(turns, rInner, rOuter, steps, phase);
    const left = [];
    const right = [];
    for (let i = 0; i < pts.length; i++) {
      const prev = pts[Math.max(0, i - 1)];
      const next = pts[Math.min(pts.length - 1, i + 1)];
      let tx = next.x - prev.x;
      let ty = next.y - prev.y;
      const len = Math.hypot(tx, ty) || 1;
      const nx = -ty / len;
      const ny = tx / len;
      const w = hurricaneWidth(pts[i].u, outerW, innerW) * 0.5;
      left.push([pts[i].x + nx * w, pts[i].y + ny * w]);
      right.push([pts[i].x - nx * w, pts[i].y - ny * w]);
    }
    const fmt = ([x, y]) => `${x.toFixed(2)} ${y.toFixed(2)}`;
    return `M${fmt(left[0])}${left.slice(1).map((p) => `L${fmt(p)}`).join("")}${right.reverse().map((p) => `L${fmt(p)}`).join("")}Z`;
  };
  const spiral = (turns, rInner, rOuter, steps, phase) => {
    return spiralPts(turns, rInner, rOuter, steps, phase).map((p, i) =>
      `${i ? "L" : "M"}${p.x.toFixed(2)} ${p.y.toFixed(2)}`
    ).join("");
  };
  const bands = [
    spiralBand(2.25, 10, 90, 110, 0.15, 11.5, 1.1),
    spiralBand(2.05, 12, 94, 100, 2.2, 8.4, 0.9),
    spiralBand(1.8, 14, 88, 88, 4.05, 6.2, 0.75),
    spiralBand(2.35, 16, 96, 104, 5.3, 9.2, 0.85),
  ];
  const arms = [
    spiral(2.15, 10, 88, 96, 0.2),
    spiral(1.95, 12, 92, 90, 2.25),
    spiral(1.7, 14, 96, 82, 4.15),
  ];
  const motes = [
    [148, 68, 1.15], [54, 128, 0.9], [158, 138, 1.05],
    [70, 52, 0.75], [128, 162, 0.95], [44, 78, 0.7],
    [170, 94, 0.85], [88, 36, 0.65],
  ];
  const hex = Array.from({ length: 6 }, (_, i) => {
    const t = (-90 + i * 60) * Math.PI / 180;
    return `${(100 + 45 * Math.cos(t)).toFixed(1)},${(100 + 45 * Math.sin(t)).toFixed(1)}`;
  }).join(" ");
  return `
    <i class="orbit-wind-halo"></i>
    <svg class="orbit-wind" viewBox="0 0 200 200" aria-hidden="true">
      <defs>
        <filter id="orbit-wind-cloud" x="-40%" y="-40%" width="180%" height="180%">
          <feTurbulence type="fractalNoise" baseFrequency="0.038" numOctaves="3" seed="4" result="n"/>
          <feDisplacementMap in="SourceGraphic" in2="n" scale="3.4" xChannelSelector="R" yChannelSelector="G"/>
          <feGaussianBlur stdDeviation="1.6"/>
        </filter>
        <filter id="orbit-wind-softmask" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="1.6"/>
        </filter>
        <mask id="orbit-wind-touch" maskUnits="userSpaceOnUse">
          <rect width="200" height="200" fill="white"/>
          <g filter="url(#orbit-wind-softmask)">
            <polygon class="orbit-wind-hexmask" points="${hex}" fill="black"/>
            <circle cx="100" cy="100" r="17" fill="white"/>
          </g>
        </mask>
      </defs>
      <g class="orbit-wind-clip" mask="url(#orbit-wind-touch)">
        <g class="orbit-wind-flow" filter="url(#orbit-wind-cloud)">
          ${bands.map((d, i) => `<path class="orbit-wind-band is-${i}" d="${d}"/>`).join("")}
          ${arms.map((d, i) => `<path class="orbit-wind-arm is-${i}" d="${d}"/>`).join("")}
        </g>
        <g class="orbit-wind-rings" filter="url(#orbit-wind-cloud)">
          <ellipse cx="100" cy="100" rx="62" ry="54"/>
          <ellipse cx="100" cy="100" rx="76" ry="66"/>
          <ellipse cx="100" cy="100" rx="90" ry="78"/>
        </g>
        <g class="orbit-wind-motes">
          ${motes.map(([x, y, r]) => `<circle cx="${x}" cy="${y}" r="${r}"/>`).join("")}
        </g>
      </g>
    </svg>`;
}

function orbitSkyMarkup() {
  let seed = 91;
  const rnd = () => {
    seed = (seed * 16807) % 2147483647;
    return (seed - 1) / 2147483646;
  };
  const dots = [];
  for (let i = 0; i < 1100; i++) {
    const ang = rnd() * Math.PI * 2;
    const rad = 18 + rnd() ** 0.72 * 80;
    const fade = Math.max(0.12, (1 - rad / 102) ** 1.15);
    const kind = rnd() < 0.28 ? "is-ice" : rnd() < 0.42 ? "is-hot" : "";
    const r = rnd() < 0.08 ? 0.42 + rnd() * 0.28 : 0.12 + rnd() * 0.22;
    dots.push(
      `<circle class="orbit-sky-star ${kind}" cx="${(100 + rad * Math.cos(ang)).toFixed(2)}" cy="${(100 + rad * Math.sin(ang)).toFixed(2)}" r="${r.toFixed(2)}" style="--o:${fade.toFixed(2)};animation-delay:${(rnd() * 3.2).toFixed(2)}s"/>`
    );
  }
  return `<svg class="orbit-sky" viewBox="0 0 200 200" aria-hidden="true">${dots.join("")}</svg>`;
}

function collisionMarkup() {
  const rnd = (seed0) => {
    let seed = seed0;
    return () => {
      seed = (seed * 16807) % 2147483647;
      return (seed - 1) / 2147483646;
    };
  };
  const bunch = (cls, seed0) => {
    const r = rnd(seed0);
    const dots = [];
    for (let i = 0; i < 26; i++) {
      const x = (r() - 0.5) * 36;
      const y = (r() - 0.5) * 8.5;
      const rad = 0.45 + r() * 0.9;
      dots.push(`<circle class="collide-proton" cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${rad.toFixed(2)}"/>`);
    }
    return `
      <g class="collide-bunch ${cls}">
        <ellipse class="collide-bunch-trail" cx="-18" cy="0" rx="22" ry="5.2"/>
        <ellipse class="collide-bunch-halo" cx="0" cy="0" rx="26" ry="8.5"/>
        <ellipse class="collide-bunch-core" cx="0" cy="0" rx="14" ry="3.8"/>
        ${dots.join("")}
      </g>`;
  };
  const shower = (axisDeg, cls, seed0) => {
    const r = rnd(seed0);
    const ax = axisDeg * Math.PI / 180;
    const parts = [];
    const addPhoton = (x1, y1, ang, len, depth) => {
      const x2 = x1 + Math.cos(ang) * len;
      const y2 = y1 + Math.sin(ang) * len;
      parts.push(`<line class="collide-photon d${depth}" x1="${x1.toFixed(1)}" y1="${y1.toFixed(1)}" x2="${x2.toFixed(1)}" y2="${y2.toFixed(1)}"/>`);
      if (depth < 3) {
        addElectron(x2, y2, ang + 0.16 + r() * 0.08, len * (0.55 + r() * 0.12), depth + 1, 1);
        addElectron(x2, y2, ang - 0.16 - r() * 0.08, len * (0.52 + r() * 0.1), depth + 1, -1);
      }
    };
    const addElectron = (x1, y1, ang, len, depth, bend) => {
      const x2 = x1 + Math.cos(ang) * len;
      const y2 = y1 + Math.sin(ang) * len;
      const mx = x1 + Math.cos(ang) * len * 0.48 - Math.sin(ang) * (4.2 + depth) * bend;
      const my = y1 + Math.sin(ang) * len * 0.48 + Math.cos(ang) * (4.2 + depth) * bend;
      parts.push(`<path class="collide-electron d${depth}" d="M${x1.toFixed(1)} ${y1.toFixed(1)} Q${mx.toFixed(1)} ${my.toFixed(1)} ${x2.toFixed(1)} ${y2.toFixed(1)}"/>`);
      if (depth < 3) {
        addPhoton(x1 + (x2 - x1) * 0.42, y1 + (y2 - y1) * 0.42, ang + 0.2 * bend, len * 0.52, depth + 1);
        if (depth < 2) addPhoton(x2, y2, ang - 0.1 * bend, len * 0.38, depth + 1);
      }
    };
    addPhoton(200, 100, ax - 0.09, 40, 0);
    addPhoton(200, 100, ax + 0.11, 44, 0);
    addElectron(200, 100, ax + 0.24, 32, 0, 1);
    addElectron(200, 100, ax - 0.26, 30, 0, -1);
    const gx = 200 + Math.cos(ax) * 38;
    const gy = 100 + Math.sin(ax) * 38;
    parts.unshift(`<ellipse class="collide-shower-glow" cx="${gx.toFixed(1)}" cy="${gy.toFixed(1)}" rx="40" ry="15" transform="rotate(${axisDeg.toFixed(1)} ${gx.toFixed(1)} ${gy.toFixed(1)})"/>`);
    for (let i = 0; i < 7; i++) {
      const a = ax + (r() - 0.5) * 0.7;
      const d = 18 + r() * 48;
      parts.push(`<circle class="collide-spark d${i % 3}" cx="${(200 + Math.cos(a) * d).toFixed(1)}" cy="${(100 + Math.sin(a) * d).toFixed(1)}" r="${(0.55 + r() * 0.7).toFixed(2)}"/>`);
    }
    return `<g class="collide-shower ${cls}">${parts.join("")}</g>`;
  };
  const rays = Array.from({ length: 10 }, (_, i) => {
    const a = (i / 10) * Math.PI * 2;
    return `<line class="collide-flash-ray" x1="${(200 + Math.cos(a) * 4).toFixed(1)}" y1="${(100 + Math.sin(a) * 4).toFixed(1)}" x2="${(200 + Math.cos(a) * 22).toFixed(1)}" y2="${(100 + Math.sin(a) * 22).toFixed(1)}"/>`;
  }).join("");
  return `
    <svg class="orbit-collide" viewBox="0 0 400 200" aria-hidden="true">
      <defs>
        <filter id="collide-soft" x="-50%" y="-80%" width="200%" height="260%">
          <feGaussianBlur stdDeviation="1.3"/>
        </filter>
        <radialGradient id="collide-flash-g" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="#fff8e6"/>
          <stop offset="35%" stop-color="#ffe08a"/>
          <stop offset="100%" stop-color="#ff8a3a" stop-opacity="0"/>
        </radialGradient>
      </defs>
      <line class="collide-beam" x1="8" y1="100" x2="392" y2="100"/>
      <g transform="translate(200 100)">
        ${bunch("is-l", 17)}
        ${bunch("is-r", 41)}
      </g>
      <g class="collide-flash">
        <circle class="collide-flash-core" cx="200" cy="100" r="16" fill="url(#collide-flash-g)"/>
        <g filter="url(#collide-soft)">${rays}</g>
      </g>
      <g class="collide-event">
        <g class="collide-aim is-a">${shower(-38, "is-a", 23)}</g>
        <g class="collide-aim is-b">${shower(142, "is-b", 59)}</g>
      </g>
    </svg>`;
}

function hero(extra = "") {
  const vid = setting("hero_video") || "ttSjYYBOlsM";
  return `
    <section class="hero">
      <!-- test: title background movie off
      <div class="hero-video" aria-hidden="true">
        <img class="hero-poster" src="/static/media/hero-poster.jpg" alt="" />
        <iframe
          data-src="https://www.youtube-nocookie.com/embed/${esc(vid)}?autoplay=1&mute=1&controls=0&loop=1&playlist=${esc(vid)}&start=10&playsinline=1&rel=0&modestbranding=1&iv_load_policy=3"
          title="EOS background video"
          allow="autoplay; encrypted-media; picture-in-picture"
          tabindex="-1"></iframe>
      </div>
      -->
      <div class="hero-overlay"></div>
      <div class="hero-inner">
        <button type="button" class="hero-term" data-copy-term title="Copy command" aria-label="Copy ${esc(TERM_CLONE)}">
          <span class="hero-term-prompt">$</span>
          <span class="hero-term-text" data-term></span>
          <span class="hero-term-caret" aria-hidden="true"></span>
          <span class="hero-term-pop" aria-live="polite">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m5.5 12.5 4.2 4.2 8.8-9.4"/></svg>
            Copied
          </span>
        </button>
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
            ${latestNewsTicker()}
          </div>
          <div class="orbit-slot">
            <div class="orbit-logo" aria-hidden="true">
              ${orbitSkyMarkup()}
              ${windSwirlMarkup()}
              <img src="/static/media/eos-hex.png" alt="" />
              ${collisionMarkup()}
            </div>
            <div class="orbit-frame">
              <video class="hero-orbit" muted loop playsinline preload="none" aria-label="EOS orbit"></video>
            </div>
          </div>
        </div>
        ${extra}
      </div>
    </section>`;
}

function stats(extra = "") {
  const items = [
    [setting("stat_volume"), setting("stat_volume_label")],
    [setting("stat_io") || "1–2 TB/s", setting("stat_io_label") || "IO"],
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
    "s-status": `<svg viewBox="0 0 24 24" ${stroke}><path d="M5 18V9.2M10 18V6M15 18v-5.2M20 18V7.4"/><path d="M3.6 19h17"/></svg>`,
    "r-commits": `<svg viewBox="0 0 24 24" ${stroke}><circle cx="7.2" cy="7" r="2.1"/><circle cx="16.8" cy="12" r="2.1"/><circle cx="7.2" cy="17" r="2.1"/><path d="M9.3 7h3.4c2 0 2.9.9 2.9 2.6V12M9.3 17h3.4c2 0 2.9-.9 2.9-2.6V12"/></svg>`,
    "r-pres": `<svg viewBox="0 0 24 24" ${stroke}><rect x="3.6" y="5" width="16.8" height="11.2" rx="1.6"/><path d="M8 19.2h8M12 16.2v3"/></svg>`,
    "r-docs": `<svg viewBox="0 0 24 24" ${stroke}><path d="M4.4 6.1c2.5-1.2 5.3-1.1 7.6.5v11.8c-2.3-1.6-5.1-1.7-7.6-.5z"/><path d="M19.6 6.1c-2.5-1.2-5.3-1.1-7.6.5v11.8c2.3-1.6 5.1-1.7 7.6-.5z"/></svg>`,
    "r-search": `<svg viewBox="0 0 24 24" ${stroke}><circle cx="12" cy="7.1" r="2.15"/><path d="M7.1 17.8c.35-3.15 2.25-4.85 4.9-4.85s4.55 1.7 4.9 4.85"/><circle cx="6.05" cy="8.35" r="1.65"/><path d="M3.3 17.8c.2-2.15 1.3-3.4 2.85-3.7"/><circle cx="17.95" cy="8.35" r="1.65"/><path d="M20.7 17.8c-.2-2.15-1.3-3.4-2.85-3.7"/></svg>`,
    "r-pubs": `<svg viewBox="0 0 24 24" ${stroke}><rect x="6.4" y="4.6" width="12.2" height="14.6" rx="1.2"/><path d="M4.8 6.8v12c0 .7.55 1.25 1.25 1.25H16"/><path d="M9 8.8h6.6M9 12h6.6M9 15.2h4.4"/></svg>`,
    "r-rel": `<svg viewBox="0 0 24 24" ${stroke}><path d="M12.5 4.5 19.5 11.5a1.5 1.5 0 0 1 0 2.1l-5.9 5.9a1.5 1.5 0 0 1-2.1 0L4.5 12.5V4.5h8z"/><circle cx="9.1" cy="9.1" r="1.2"/></svg>`,
    "n-ws27": `<svg viewBox="0 0 24 24" ${stroke}><rect x="3.6" y="5.2" width="16.8" height="14.4" rx="1.8"/><path d="M8 3.6v3.4M16 3.6v3.4M3.6 9.4h16.8"/><path d="M12 12.4v3.4l2.3 1.3"/></svg>`,
    "n-ws26": `<svg viewBox="0 0 24 24" ${stroke}><path d="M8.2 10.6 12 7.8l3.8 2.8v1.8H8.2z"/><path d="M6.2 19.2 8.4 12.6h7.2l2.2 6.6z"/><path d="M11.15 7.6V5.3h1.7v2.3"/><circle cx="12" cy="4.4" r="0.95"/></svg>`,
    "n-551": `<svg viewBox="0 0 24 24" ${stroke}><path d="M12 3.6 19.4 10.4 12 20.4 4.6 10.4z"/><path d="M4.6 10.4h14.8M8.1 10.4 12 3.6l3.9 6.8"/></svg>`,
    "n-exa": `<svg viewBox="0 0 24 24" ${stroke}><ellipse cx="12" cy="6.4" rx="7" ry="2.35"/><path d="M5 6.4v3.2c0 1.3 3.1 2.35 7 2.35s7-1.05 7-2.35V6.4"/><ellipse cx="12" cy="14.6" rx="7" ry="2.35"/><path d="M5 14.6v3c0 1.3 3.1 2.35 7 2.35s7-1.05 7-2.35v-3"/></svg>`,
    "n-generic": `<svg viewBox="0 0 24 24" ${stroke}><rect x="4.2" y="4.2" width="15.6" height="15.6" rx="1.5"/><path d="M7.2 8.2h9.6M7.2 12h9.6M7.2 15.8h6.2"/></svg>`,
    "sup-forum": `<svg viewBox="0 0 24 24" ${stroke}><path d="M5 6.2h14v9.2H9.2L5 18.8z"/></svg>`,
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
  "energy-aware placement",
  "policy-driven platform",
  "disk, erasure coding and tape",
  "data locality",
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
      <figcaption class="arch-cap">Three core services - MGM, FST and QuarkDB. Messaging is QuarkDB pub-sub, not a separate MQ. Clients open on the MGM; data I/O is redirected to FSTs.</figcaption>
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
        <li><b>3</b><div><strong>Read or write data</strong><span>I/O stays on the FST - replica or erasure-coded stripes, with checksums.</span></div></li>
        <li><b>4</b><div><strong>Persist metadata</strong><span>The MGM write-back queue commits namespace changes to QuarkDB.</span></div></li>
      </ol>
      <div class="section-head">
        <h2>Core services</h2>
        <p class="muted">MGM, FST and QuarkDB - pub-sub lives in QuarkDB - plus the clients that speak XRootD, HTTP and FUSE.</p>
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
          <p>Each client is mapped, from its authentication method and vid rules, to a virtual identity - a uid/gid pair that owns files and directories. Roles can be attached so a person or service may act on behalf of everyone, or of a subset of identities.</p>
        </article>
        <article class="card set-card">
          <p class="meta">Policies</p>
          <h3>How a file is stored</h3>
          <p>Layout, checksums and placement can be set on a space, on an application, group or user, on a directory, or on a single URL. A directory can force erasure coding - for example <code>sys.forced.layout=raid6</code> with 12 stripes in an <code>erasure</code> space.</p>
        </article>
        <article class="card set-card">
          <p class="meta">GEO</p>
          <h3>Placement close to the client</h3>
          <p>GEO tags can be assigned to client IPs and to FST nodes. Placement policies match the two, so a file can be stored as close as possible to the reader - or kept apart for resilience.</p>
        </article>
      </div>
      <div class="section-head" style="margin-top:2.4rem">
        <h2>Microservices</h2>
        <p class="muted">Every instance ships a set of configurable engines that keep the cluster balanced and consistent.</p>
      </div>
      <div class="micro-grid">
        <article class="set-card"><h3>Balancers</h3><p>Filesystem balancer inside a group, group balancer inside a space, and geo balancer across locations.</p></article>
        <article class="set-card"><h3>Converter</h3><p>Queued jobs that change how a file is stored - for example from one replica to two-fold replication.</p></article>
        <article class="set-card"><h3>Lifecycle</h3><p>Automation for disk and node replacement: empty a filesystem that should leave production.</p></article>
        <article class="set-card"><h3>LRU engine</h3><p>Scans the namespace to apply clean-up or conversion policies, such as a scratch space that expires in 30 days.</p></article>
        <article class="set-card"><h3>Inspector</h3><p>Accounting of how files are stored - how much of an instance is replicated versus erasure-coded.</p></article>
        <article class="set-card"><h3>Consistency</h3><p>Distributed check and repair of data and metadata inconsistencies.</p></article>
        <article class="set-card"><h3>Workflow</h3><p>Event queue for external systems - typically CTA, when a new file should migrate to tape.</p></article>
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
      <p class="muted">${esc(catalog.index?.queryHelp || "")}. Search titles, abstracts, and speakers - then open slides or the CERN recording.</p>
      <p><a class="btn" data-nav href="/search">Search presentations</a></p>
      ${rows ? `<div class="pub-table" style="margin-top:1.2rem">${rows}</div>` : `<p class="empty">No workshops indexed yet.</p>`}
    </div>`;
}

function commitsPage() {
  const q = new URLSearchParams(location.search).get("q") || "";
  return `
    <div class="wrap">
      <p class="kicker">Resources</p>
      <h2>Search commits</h2>
      <p class="muted">Live log from the EOS <code>master</code> branch. Headings, messages, and authors - refreshed from GitHub.</p>
      <form class="search-box" data-search="commits" id="commit-form">
        <input name="q" value="${esc(q)}" placeholder="author, fix, QuarkDB, SHA…" />
        <button class="btn" type="submit">Search commits</button>
      </form>
      <div id="commit-results"><p class="muted">Loading the live master log…</p></div>
    </div>`;
}

function searchKind(raw) {
  if (raw === "talks") return "workshop";
  if (raw === "all") return "workshop-docs";
  return raw || "presentations";
}

function searchPage() {
  const params = new URLSearchParams(location.search);
  const q = params.get("q") || "";
  const kind = searchKind(params.get("kind"));
  const year = params.get("year") || "";
  const years = (catalog.years || []).map((y) => `<option value="${y}" ${String(y) === year ? "selected" : ""}>${y}</option>`).join("");
  return `
    <div class="wrap">
      <p class="kicker">Search</p>
      <h2>Presentations and documentation</h2>
      <p class="muted">${esc(catalog.index?.queryHelp || "Index of EOS workshop talks, conference presentations, and eos-docs.")}</p>
      <form class="search-box" data-search="all" id="search-form">
        <input name="q" value="${esc(q)}" placeholder="FUSE, QuarkDB, CTA, site report, recycle bin…" />
        <select name="kind">
          <option value="presentations" ${kind === "presentations" ? "selected" : ""}>All Presentations</option>
          <option value="workshop-docs" ${kind === "workshop-docs" ? "selected" : ""}>Workshop Presentations + Docs</option>
          <option value="workshop" ${kind === "workshop" ? "selected" : ""}>Workshop Presentations</option>
          <option value="external" ${kind === "external" ? "selected" : ""}>External Presentations</option>
          <option value="docs" ${kind === "docs" ? "selected" : ""}>Docs</option>
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
  { set: "eos", year: 2019, title: "Scaling the EOS namespace - new developments and performance optimizations",
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
  { set: "eos", year: 2019, title: "EOS Open Storage - evolution of an ecosystem for scientific data repositories",
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
      <h2>Documentation - Publications - Sourcecode</h2>
      ${cardGrid(cards("resource"), "grid-3")}
      <section class="pubs" id="publications">
        <div class="section-head">
          <h2>Publications</h2>
          <p class="muted">Reverse-chronological papers on EOS at CERN - architecture, operations - and XRootD where it is part of the EOS data path.</p>
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
      <h3 class="card-title">${cardIcon(n.id) || cardIcon("n-generic")}${esc(n.title)}</h3>
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
      <div class="section-head" style="margin-top:2rem"><h2>Physics Data Service Lead &amp; openlab</h2></div>
      <div class="grid grid-3 team-grid">${ops}</div>
      <div class="section-head" style="margin-top:2rem"><h2>Collaborations</h2></div>
      ${cardGrid(cards("collab"), "grid-2")}
      <div class="section-head" style="margin-top:2rem"><h2>Support</h2></div>
      ${cardGrid([{
        id: "sup-forum",
        title: "Community Forum",
        href: (setting("community") || "https://eos-community.web.cern.ch/").replace(/^http:/, "https:"),
        body: "Discourse for EOS sites, operators, and users.",
      }], "grid-2")}
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
      lede: "Flexible hierarchies inside EOS - cache, disk, erasure coding and tape as one platform.",
      items: [
        { title: "Read-through spaces & tiering", body: "Move data between spaces by policy - SSD to HDD, replica to erasure coding, hot to cold - and treat one space as a transparent cache in front of another, with admission, eviction, pinning and prefetch." },
        { title: "Native Mirage I/O", body: "Simulate file contents, checksums and the network path without touching disks, so protocol, gateway and scalability tests no longer depend on physical storage." },
        { title: "Native tape", body: "Investigate tape as an EOS storage tier rather than an external CTA buffer: lifecycle, archive and retrieve from EOS, with a path for today’s EOSCTA sites." },
        { title: "Erasure coding in production", body: "File updates, small files, degraded reads, faster repair and online conversion - so EC can sit in a hot → replica → EC → tape lifecycle." },
      ],
    },
    {
      title: "High-performance access",
      lede: "Standards and transports for HPC, Kubernetes and GPU-heavy analysis.",
      items: [
        { title: "Production NFSv4.1", body: "Locking, caching, identity mapping and monitoring where eosxd is hard to deploy - Kubernetes, HPC and sites that need a standard filesystem." },
        { title: "XrdHttp, GPU and RDMA", body: "Kernel clients, low-copy paths into GPU memory and RDMA, with common capability discovery for AI/ML and high-throughput analysis." },
        { title: "S3 as another EOS protocol", body: "Re-think the gateway model so object access shares more of EOS instead of a separate Versity-shaped stack." },
      ],
    },
    {
      title: "Operations & reliability",
      lede: "One picture of the instance, and data movement that protects client traffic.",
      items: [
        { title: "Controller UI", body: "A control plane for health, spaces, nodes, drain, quotas and alerts - safe workflows with RBAC, and later a fleet view across instances." },
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
        { title: "Common OIDC", body: "Push XrdSecOIDC into XRootD and use it for FUSE, HTTP and gateways - one token and identity story instead of a different one per access path." },
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
        <span>Major marks from the 2026 workshop timeline - then the work ahead.</span>
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
      <p class="roadmap-aim">${esc("The direction is a single EOS that caches, places and archives by policy - disk, erasure coding and tape together - and that can exploit locality and energy on future CERN computing farms.")}</p>
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
    case "/commits": return commitsPage();
    case "/resources": return resources();
    case "/service": return service();
    case "/news": return news();
    case "/community": return community();
    default: return notFound();
  }
}

function commitWhen(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return String(iso).slice(0, 10);
  return d.toISOString().slice(0, 10);
}

function commitTable(commits) {
  let lastYear = 0;
  return `<div class="pub-table">${(commits || []).map((c) => {
    const year = Number(c.year) || 0;
    const yearHead = year && year !== lastYear ? `<div class="pub-year-band">${year}</div>` : "";
    lastYear = year;
    const note = c.body || "";
    const when = commitWhen(c.date);
    const sha = c.short || (c.sha || "").slice(0, 7);
    return `${yearHead}
      <article class="pub-row">
        <span class="pub-year">${year || "-"}</span>
        <div class="pub-main">
          <h3>${c.url ? `<a href="${esc(c.url)}" target="_blank" rel="noreferrer">${esc(c.title)}</a>` : esc(c.title)}</h3>
          ${c.author ? `<p class="pub-authors">${esc(c.author)}</p>` : ""}
          <p class="pub-venue">${esc(sha)}${when ? ` · ${esc(when)}` : ""} · master</p>
          ${note ? `<p class="pub-note">${esc(note)}</p>` : ""}
          ${c.url ? `<p class="pub-links"><a href="${esc(c.url)}" target="_blank" rel="noreferrer">Commit</a></p>` : ""}
        </div>
      </article>`;
  }).join("")}</div>`;
}

function renderCommits(data) {
  const commits = data.commits || [];
  const when = data.fetchedAt ? commitWhen(data.fetchedAt) : "";
  const src = data.source || "master";
  if (!commits.length) {
    return `<p class="empty">${data.query ? `No matches for “${esc(data.query)}”.` : "No commits loaded from master yet."}</p>`;
  }
  return `
    <div class="section-head" style="margin-top:0.4rem">
      <h2>${data.query ? "Matching commits" : "Latest commits on master"}</h2>
      <p class="muted">${commits.length} ${commits.length === 1 ? "commit" : "commits"}${data.query ? ` for “${esc(data.query)}”` : ""}${data.total && !data.query ? ` · ${data.total} indexed` : ""} · ${esc(src)}${when ? ` · updated ${esc(when)}` : ""}. Hover a row for the message.</p>
    </div>
    ${commitTable(commits)}`;
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
        <span class="pub-year">${year || "-"}</span>
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

function talkHeading(data) {
  const k = data.kind || "";
  if (data.query) {
    if (k === "external") return "Matching external presentations";
    if (k === "workshop" || k === "workshop-docs") return "Matching workshop presentations";
    return "Matching presentations";
  }
  if (k === "external") return "External presentations";
  if (k === "workshop" || k === "workshop-docs") return "Workshop presentations";
  return "All presentations";
}

function renderHits(data) {
  const talks = data.talks || [];
  const docs = data.docs || [];
  if (!talks.length && !docs.length) {
    return `<p class="empty">${data.query ? `No matches for “${esc(data.query)}”.` : "No talks indexed yet."}</p>`;
  }
  const talkBlock = talks.length ? `
    <div class="section-head" style="margin-top:0.4rem">
      <h2>${talkHeading(data)}</h2>
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
  const kind = searchKind(String(fd.get("kind") || (scope === "docs" ? "docs" : "presentations")));
  const year = String(fd.get("year") || "");
  if (scope === "commits") {
    const next = new URL("/commits", location.origin);
    if (q) next.searchParams.set("q", q);
    history.replaceState({}, "", next.pathname + next.search);
    const res = await fetch("/api/commits?" + new URLSearchParams({ q }).toString());
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Could not load commits");
    const box = $("#commit-results");
    if (box) box.innerHTML = renderCommits(data);
    return;
  }
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
        const box = $("#search-results") || $("#doc-results") || $("#commit-results");
        if (box) box.innerHTML = `<p class="empty">${esc(err.message)}</p>`;
      });
    });
  });
  if (path() === "/search") {
    const form = $("#search-form");
    if (form) runSearch("all", form);
  }
  if (path() === "/commits") {
    const form = $("#commit-form");
    if (form) runSearch("commits", form);
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
      msg.textContent = res.ok ? "Thanks - the message is in the controller inbox." : (data.error || "Could not send");
      msg.className = res.ok ? "ok" : "muted";
    });
  }
}

function render() {
  if (path() === "/presentations") {
    history.replaceState({}, "", "/search?kind=external");
  }
  stopTitleSpray();
  stopCapacityChart();
  stopTermType();
  stopAboutHighlight();
  stopLogoSpin();
  stopCollision();
  stopOrbitAlign();
  setNav();
  document.title = path() === "/" ? "EOS Open Storage" : `EOS · ${path().slice(1)}`;
  $("#app").innerHTML = view();
  bindPage();
  document.querySelector(".top")?.classList.remove("is-menu");
  if (location.hash) requestAnimationFrame(scrollToHash);
  if (path() === "/" || path() === "/about") startAboutHighlight();
  if (path() !== "/") return;
  startTermType();
  // startHeroBackground(); // test: title background movie off
  startOrbitMovie();
  startLogoSpin();
  startCollision();
  startOrbitAlign();
  startCapacityChart();
  const kick = () => {
    if (path() === "/") startTitleSpray();
  };
  if (document.fonts?.ready) document.fonts.ready.then(kick);
  else kick();
}

let chatHistory = [];
let chatBusy = false;
let chatPinned = false;
let chatLeaveTimer = 0;
let chatTypeTimer = 0;

function setChatOpen(on, focus) {
  const root = $("#eos-chat");
  const panel = $("#eos-chat-panel");
  const tog = $("#eos-chat-toggle");
  if (!root || !panel || !tog) return;
  root.classList.toggle("is-open", on);
  tog.setAttribute("aria-expanded", on ? "true" : "false");
  if (on && focus) panel.querySelector("textarea")?.focus();
}

function clearChat() {
  window.clearTimeout(chatTypeTimer);
  chatHistory = [];
  const log = $("#eos-chat-log");
  if (log) {
    log.innerHTML = `<p class="eos-chat-hello">Ask about EOS, CERNBox, CTA, docs or workshops.</p>`;
  }
}

function setChatLarge(on) {
  const root = $("#eos-chat");
  const grow = root?.querySelector("[data-chat-grow]");
  if (!root) return;
  root.classList.toggle("is-large", on);
  if (grow) {
    grow.setAttribute("aria-label", on ? "Shrink chat" : "Enlarge chat");
    grow.setAttribute("aria-pressed", on ? "true" : "false");
  }
  if (on) {
    chatPinned = true;
    setChatOpen(true, false);
  }
}

function appendChat(role, text) {
  const log = $("#eos-chat-log");
  if (!log) return null;
  log.querySelector(".eos-chat-hello")?.remove();
  const el = document.createElement("div");
  el.className = "eos-chat-msg is-" + role;
  el.textContent = text;
  log.appendChild(el);
  log.scrollTop = log.scrollHeight;
  return el;
}

function attachChatSources(el, sources) {
  if (!el || !sources || !sources.length) return;
  const box = document.createElement("p");
  box.className = "eos-chat-sources";
  sources.slice(0, 6).forEach((s) => {
    const a = document.createElement("a");
    a.href = s.url;
    a.target = "_blank";
    a.rel = "noreferrer";
    a.textContent = s.title || s.url;
    box.appendChild(a);
  });
  el.appendChild(box);
}

function typeChatAnswer(el, text, sources) {
  window.clearTimeout(chatTypeTimer);
  const log = $("#eos-chat-log");
  const full = String(text || "");
  if (prefersQuiet() || !full) {
    el.textContent = full;
    attachChatSources(el, sources);
    if (log) log.scrollTop = log.scrollHeight;
    return;
  }
  let i = 0;
  const tick = () => {
    i += 1;
    el.textContent = full.slice(0, i);
    if (log) log.scrollTop = log.scrollHeight;
    if (i < full.length) {
      chatTypeTimer = window.setTimeout(tick, i < 12 ? 18 : 11);
      return;
    }
    attachChatSources(el, sources);
  };
  tick();
}

async function sendChat(form) {
  if (chatBusy) return;
  const input = form.querySelector("textarea");
  const q = String(input?.value || "").trim();
  if (!q) return;
  input.value = "";
  appendChat("user", q);
  chatHistory.push({ role: "user", text: q });
  chatBusy = true;
  chatPinned = true;
  form.querySelector("button").disabled = true;
  const wait = appendChat("wait", "Looking that up…");
  try {
    const ac = new AbortController();
    const abortTimer = window.setTimeout(() => ac.abort(), 60000);
    const res = await fetch("/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: q, history: chatHistory.slice(0, -1).slice(-8) }),
      signal: ac.signal,
    });
    window.clearTimeout(abortTimer);
    const data = await res.json().catch(() => ({}));
    wait?.remove();
    if (!res.ok) {
      appendChat("error", data.error || "Ask EOS is unavailable.");
      return;
    }
    const bot = appendChat("bot", "");
    typeChatAnswer(bot, data.text || "", data.sources);
    chatHistory.push({ role: "assistant", text: data.text || "" });
  } catch (err) {
    wait?.remove();
    const timedOut = err && (err.name === "AbortError" || /abort/i.test(String(err.message || "")));
    appendChat("error", timedOut ? "Ask EOS timed out. Try again." : (err.message || "Ask EOS is unavailable."));
  } finally {
    chatBusy = false;
    form.querySelector("button").disabled = false;
    input?.focus();
  }
}

function bindChat() {
  const root = $("#eos-chat");
  if (!root || root.dataset.bound) return;
  root.dataset.bound = "1";
  fetch("/api/chat").catch(() => {});
  root.addEventListener("pointerenter", () => {
    window.clearTimeout(chatLeaveTimer);
    setChatOpen(true, false);
  });
  root.addEventListener("pointerleave", () => {
    window.clearTimeout(chatLeaveTimer);
    if (chatPinned || chatBusy || root.querySelector("textarea") === document.activeElement) return;
    chatLeaveTimer = window.setTimeout(() => setChatOpen(false, false), 280);
  });
  $("#eos-chat-toggle")?.addEventListener("click", () => {
    chatPinned = !root.classList.contains("is-open") || !chatPinned;
    setChatOpen(true, true);
  });
  root.querySelector("[data-chat-close]")?.addEventListener("click", () => {
    chatPinned = false;
    setChatLarge(false);
    setChatOpen(false, false);
  });
  root.querySelector("[data-chat-clear]")?.addEventListener("click", () => {
    if (chatBusy) return;
    clearChat();
  });
  root.querySelector("[data-chat-grow]")?.addEventListener("click", () => {
    setChatLarge(!root.classList.contains("is-large"));
  });
  $("#eos-chat-form")?.addEventListener("submit", (e) => {
    e.preventDefault();
    sendChat(e.currentTarget);
  });
  $("#eos-chat-form textarea")?.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      e.currentTarget.form?.requestSubmit();
    }
  });
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
  bindChat();
  bindDocsViewer();
  bindNavHover();
  render();
}

const DOCS_HOST = "eos-docs.web.cern.ch";
let docsStack = [];
let docsBusy = false;

function isEosDocsURL(href) {
  try {
    return new URL(href, location.href).hostname === DOCS_HOST;
  } catch (_) {
    return false;
  }
}

function docsViewer() {
  return document.querySelector(".docs-viewer");
}

function closeDocsViewer() {
  const root = docsViewer();
  if (!root) return;
  root.hidden = true;
  root.classList.remove("is-open");
  document.body.classList.remove("is-docs-open");
  docsStack = [];
}

async function openDocsViewer(raw, push = true) {
  const root = docsViewer();
  if (!root) return;
  let href = raw;
  try {
    href = new URL(raw, location.href).href;
  } catch (_) {}
  if (push) {
    if (docsStack[docsStack.length - 1] !== href) docsStack.push(href);
  }
  root.hidden = false;
  root.classList.add("is-open");
  document.body.classList.add("is-docs-open");
  const titleEl = root.querySelector("[data-docs-title]");
  const bodyEl = root.querySelector("[data-docs-body]");
  const orig = root.querySelector("[data-docs-orig]");
  const prev = root.querySelector("[data-docs-prev]");
  const next = root.querySelector("[data-docs-next]");
  const back = root.querySelector("[data-docs-back]");
  const pager = root.querySelector(".docs-viewer-pager");
  if (titleEl) titleEl.textContent = "Opening docs…";
  if (bodyEl) bodyEl.innerHTML = `<p class="docs-viewer-status">Loading the EOS documentation…</p>`;
  if (orig) orig.href = href;
  if (back) back.hidden = docsStack.length < 2;
  if (prev) { prev.hidden = true; prev.removeAttribute("href"); }
  if (next) { next.hidden = true; next.removeAttribute("href"); }
  if (pager) pager.hidden = true;
  docsBusy = true;
  try {
    const res = await fetch("/api/docs/view?url=" + encodeURIComponent(href));
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || "Could not open that docs page");
    if (titleEl) titleEl.textContent = data.title || "EOS documentation";
    if (orig) orig.href = data.url || href;
    if (bodyEl) bodyEl.innerHTML = data.html || "<p>That page had no article text.</p>";
    if (prev && data.prev) {
      prev.hidden = false;
      prev.href = data.prev.url;
      prev.textContent = "← " + (data.prev.title || "Previous");
    }
    if (next && data.next) {
      next.hidden = false;
      next.href = data.next.url;
      next.textContent = (data.next.title || "Next") + " →";
    }
    if (pager) pager.hidden = !(data.prev || data.next);
    prepareDocsCodeBlocks(bodyEl);
    const hash = (() => { try { return new URL(href).hash; } catch (_) { return ""; } })();
    if (hash && bodyEl) {
      const id = decodeURIComponent(hash.replace(/^#/, ""));
      const target = bodyEl.querySelector("#" + CSS.escape(id)) || bodyEl.querySelector(`[id="${id}"]`);
      if (target) target.scrollIntoView({ block: "start" });
      else bodyEl.scrollTop = 0;
    } else if (bodyEl) {
      bodyEl.scrollTop = 0;
    }
  } catch (err) {
    if (titleEl) titleEl.textContent = "Documentation";
    if (bodyEl) {
      bodyEl.innerHTML = `<p class="docs-viewer-status">Could not load that page here. <a href="${esc(href)}" target="_blank" rel="noreferrer">Open it on eos-docs</a>.</p><p class="muted">${esc(err.message)}</p>`;
    }
  } finally {
    docsBusy = false;
  }
}

function docsCodeTarget(target, body) {
  if (!body?.contains(target)) return null;
  const table = target.closest(".highlighttable");
  if (table) return table.querySelector(".highlight") || table.querySelector("td.code pre");
  const hi = target.closest(".highlight");
  if (hi) return hi;
  const pre = target.closest("pre");
  if (pre && !pre.closest(".linenos, .linenodiv")) return pre;
  return null;
}

function prepareDocsCodeBlocks(body) {
  if (!body) return;
  body.querySelectorAll(".highlight, pre").forEach((el) => {
    if (el.closest(".linenos, .linenodiv")) return;
    if (el.matches("pre") && el.closest(".highlight")) return;
    el.classList.add("is-copyable");
    el.setAttribute("title", "Click to copy");
    el.setAttribute("role", "button");
    el.setAttribute("tabindex", "0");
  });
}

async function copyDocsCode(el) {
  const block = el.closest(".highlight") || el.closest("pre") || el;
  const pre = block.matches("pre") ? block : (block.querySelector("pre") || block);
  const text = (pre.innerText || "").replace(/\u00a0/g, " ").replace(/\n$/, "");
  if (!text.trim()) return;
  if (!(await copyText(text))) return;
  block.classList.add("is-copied");
  window.clearTimeout(Number(block.dataset.copyTimer || 0));
  block.dataset.copyTimer = String(window.setTimeout(() => block.classList.remove("is-copied"), 1400));
}

function bindDocsViewer() {
  const root = docsViewer();
  if (!root || root.dataset.bound) return;
  root.dataset.bound = "1";
  root.addEventListener("click", (e) => {
    if (e.target.closest("[data-docs-close]")) {
      e.preventDefault();
      closeDocsViewer();
      return;
    }
    if (e.target.closest("[data-docs-back]")) {
      e.preventDefault();
      if (docsStack.length > 1) {
        docsStack.pop();
        openDocsViewer(docsStack[docsStack.length - 1], false);
      }
      return;
    }
    const code = docsCodeTarget(e.target, root.querySelector("[data-docs-body]"));
    if (code) {
      e.preventDefault();
      copyDocsCode(code);
      return;
    }
    const a = e.target.closest("a");
    if (!a || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    const href = a.getAttribute("href") || "";
    if (href.startsWith("#")) {
      e.preventDefault();
      const id = decodeURIComponent(href.slice(1));
      const target = root.querySelector("#" + CSS.escape(id)) || root.querySelector(`[id="${id}"]`);
      target?.scrollIntoView({ block: "start" });
      return;
    }
    if (isEosDocsURL(a.href)) {
      e.preventDefault();
      openDocsViewer(a.href);
    }
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && root.classList.contains("is-open")) closeDocsViewer();
    if ((e.key === "Enter" || e.key === " ") && root.classList.contains("is-open")) {
      const code = docsCodeTarget(e.target, root.querySelector("[data-docs-body]"));
      if (code) {
        e.preventDefault();
        copyDocsCode(code);
      }
    }
  });
}

document.addEventListener("click", (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a");
  if (!a || a.hasAttribute("download")) return;
  if (isEosDocsURL(a.href) && !a.closest(".docs-viewer")) {
    e.preventDefault();
    openDocsViewer(a.href);
    return;
  }
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
  msg.textContent = res.ok ? "Subscribed - we will confirm from the controller." : (data.error || "Could not subscribe");
});

window.addEventListener("pageshow", (e) => {
  ensureCapacityChart(e.persisted);
});
document.addEventListener("visibilitychange", () => {
  if (!document.hidden) ensureCapacityChart(true);
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
