let state = null;
let tab = "home";

const $ = (id) => document.getElementById(id);

async function api(path, opt = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json", ...(opt.headers || {}) },
    ...opt,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function field(name, value, label, extra = "") {
  return `<label>${label}<input name="${name}" value="${esc(value || "")}" ${extra}></label>`;
}

function area(name, value, label) {
  return `<label>${label}<textarea name="${name}">${esc(value || "")}</textarea></label>`;
}

function showDash(on) {
  $("gate").hidden = on;
  $("dash").hidden = !on;
}

async function refresh() {
  state = await api("/api/controller/state");
  render();
}

function formObj(form) {
  const data = {};
  form.querySelectorAll("[name]").forEach((el) => {
    if (el.type === "checkbox") data[el.name] = el.checked;
    else if (el.type === "number") data[el.name] = Number(el.value || 0);
    else data[el.name] = el.value;
  });
  return data;
}

function cardsOf(kind) {
  return (state.cards || []).filter((c) => c.kind === kind);
}

function cardForms(kind, title) {
  return `
    <h2>${title}</h2>
    ${cardsOf(kind).map((c) => `
      <form class="card grid card-form" data-id="${esc(c.id)}">
        <input type="hidden" name="kind" value="${esc(c.kind)}" />
        <div class="row">${field("title", c.title, "Title")}${field("meta", c.meta, "Meta")}</div>
        ${area("body", c.body, "Text")}
        <div class="row">${field("href", c.href, "Link")}${field("sort", c.sort, "Sort", 'type="number"')}</div>
        <label><input type="checkbox" name="visible" ${c.visible ? "checked" : ""}> visible</label>
        <div><button type="submit">Save</button> <button type="button" data-del-card="${esc(c.id)}">Delete</button></div>
      </form>`).join("")}
    <form class="card grid card-new" data-kind="${esc(kind)}">
      <strong>New ${esc(kind)}</strong>
      <input type="hidden" name="kind" value="${esc(kind)}" />
      <div class="row">${field("title", "", "Title")}${field("meta", "", "Meta")}</div>
      ${area("body", "", "Text")}
      <div class="row">${field("href", "", "Link")}${field("sort", "10", "Sort", 'type="number"')}</div>
      <label><input type="checkbox" name="visible" checked> visible</label>
      <button type="submit">Add</button>
    </form>`;
}

function renderHome() {
  const s = state.settings || {};
  return `
    <h1>Home texts</h1>
    <form class="grid" id="settings-form">
      ${field("hero_kicker", s.hero_kicker, "Kicker")}
      ${field("hero_title", s.hero_title, "Title")}
      ${area("hero_lede", s.hero_lede, "Lead")}
      ${field("hero_video", s.hero_video, "YouTube video ID (hero background)")}
      <div class="row">
        ${field("latest_version", s.latest_version, "Latest version")}
        ${field("latest_version_url", s.latest_version_url, "Release notes URL")}
      </div>
      <div class="row">
        ${field("workshop_label", s.workshop_label, "Workshop badge")}
        ${field("workshop_url", s.workshop_url, "Workshop URL")}
      </div>
      <div class="row">
        ${field("stat_volume", s.stat_volume, "Volume")}
        ${field("stat_io", s.stat_io, "IO")}
        ${field("stat_disks", s.stat_disks, "Disks")}
        ${field("stat_files", s.stat_files, "Files")}
        ${field("stat_clients", s.stat_clients, "Clients")}
      </div>
      <div class="row">
        ${field("contact_email", s.contact_email, "Support email")}
        ${field("address", s.address, "Address")}
      </div>
      ${area("contributors", s.contributors, "Contributors")}
      ${field("indico_event_ids", s.indico_event_ids, "Indico event IDs (comma-separated)")}
      <button class="primary" type="submit">Save settings</button>
    </form>
    <h2>Pages</h2>
    ${(state.pages || []).map((p) => `
      <form class="card grid page-form" data-id="${esc(p.id)}">
        <strong>${esc(p.id)}</strong>
        ${field("title", p.title, "Title")}
        ${area("body", p.body, "Text")}
        <button type="submit">Save page</button>
      </form>`).join("")}`;
}

function renderNews() {
  return `
    <h1>News</h1>
    ${(state.news || []).map((n) => `
      <form class="card grid news-form" data-id="${esc(n.id)}">
        <div class="row">${field("title", n.title, "Title")}${field("dateLabel", n.dateLabel, "Date label")}</div>
        ${area("body", n.body, "Text")}
        <div class="row">${field("href", n.href, "Link")}${field("sort", n.sort, "Sort", 'type="number"')}</div>
        <label><input type="checkbox" name="visible" ${n.visible ? "checked" : ""}> visible</label>
        <div><button type="submit">Save</button> <button type="button" data-del-news="${esc(n.id)}">Delete</button></div>
      </form>`).join("")}
    <form class="card grid" id="news-new">
      <strong>New item</strong>
      <div class="row">${field("title", "", "Title")}${field("dateLabel", "", "Date label")}</div>
      ${area("body", "", "Text")}
      <div class="row">${field("href", "", "Link")}${field("sort", "10", "Sort", 'type="number"')}</div>
      <button type="submit">Add</button>
    </form>`;
}

function renderTeam() {
  return `
    <h1>Team</h1>
    ${(state.people || []).map((p) => `
      <form class="card grid person-form" data-id="${esc(p.id)}">
        <div class="row">${field("name", p.name, "Name")}${field("role", p.role, "Role")}</div>
        <div class="row">${field("email", p.email, "Email")}${field("sort", p.sort, "Sort", 'type="number"')}</div>
        <label><input type="checkbox" name="visible" ${p.visible ? "checked" : ""}> visible</label>
        <div><button type="submit">Save</button> <button type="button" data-del-person="${esc(p.id)}">Delete</button></div>
      </form>`).join("")}
    <form class="card grid" id="person-new">
      <strong>New person</strong>
      <div class="row">${field("name", "", "Name")}${field("role", "", "Role")}</div>
      ${field("email", "", "Email")}
      <button type="submit">Add</button>
    </form>
    ${cardForms("collab", "Collaborations")}`;
}

function renderIndex() {
  const idx = state.index || {};
  const ws = (state.workshops || []).map((w) => `
    <tr><td>${esc(String(w.year))}</td><td>${esc(w.title)}</td><td>${esc(w.location)}</td><td><a href="${esc(w.url)}" target="_blank" rel="noreferrer">Indico</a></td></tr>`).join("");
  return `
    <h1>Search index</h1>
    <p class="muted">${esc(idx.queryHelp || "")}</p>
    <p>
      <button type="button" id="btn-indico">Refresh Indico</button>
      <button type="button" id="btn-docs">Refresh docs</button>
      <button type="button" id="btn-seed">Reload bundled seed</button>
    </p>
    <p id="index-msg" class="muted"></p>
    <h2>Workshops</h2>
    <table><thead><tr><th>Year</th><th>Title</th><th>Place</th><th></th></tr></thead><tbody>${ws}</tbody></table>`;
}

function renderInbox() {
  const rows = (state.inbox || []).map((m) => `
    <tr><td>${esc(m.createdAt)}</td><td>${esc(m.name)}</td><td>${esc(m.email)}</td><td>${esc(m.message)}</td></tr>`).join("");
  const subs = (state.subscribers || []).map((s) => `
    <tr>
      <td>${esc(s.email)}</td><td>${esc(s.status)}</td><td>${esc(s.createdAt)}</td>
      <td><button type="button" data-sub="${esc(s.id)}" data-status="confirmed">Confirm</button></td>
    </tr>`).join("");
  return `
    <h1>Inbox</h1>
    <table><thead><tr><th>When</th><th>Name</th><th>Email</th><th>Message</th></tr></thead><tbody>${rows || `<tr><td colspan="4" class="muted">Empty</td></tr>`}</tbody></table>
    <h2>Newsletter</h2>
    <table><thead><tr><th>Email</th><th>Status</th><th>When</th><th></th></tr></thead><tbody>${subs || `<tr><td colspan="4" class="muted">None</td></tr>`}</tbody></table>`;
}

function render() {
  const panel = $("panel");
  document.querySelectorAll(".tab").forEach((b) => b.classList.toggle("on", b.dataset.tab === tab));
  if (tab === "home") panel.innerHTML = renderHome();
  if (tab === "about") panel.innerHTML = cardForms("feature", "About cards") + cardForms("help", "Docs help") + cardForms("resource", "Resources") + cardForms("service", "Services");
  if (tab === "tech") panel.innerHTML = cardForms("component", "Architecture") + cardForms("link", "Code & trackers");
  if (tab === "news") panel.innerHTML = renderNews();
  if (tab === "team") panel.innerHTML = renderTeam();
  if (tab === "index") panel.innerHTML = renderIndex();
  if (tab === "inbox") panel.innerHTML = renderInbox();
}

function showGateError(msg) {
  const el = $("gate-error");
  el.hidden = !msg;
  el.textContent = msg || "";
}

$("gate-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    await api("/api/controller/login", { method: "POST", body: JSON.stringify({ secret: $("secret").value }) });
    showGateError("");
    showDash(true);
    await refresh();
  } catch (err) {
    showGateError(err.message);
  }
});

$("btn-logout").addEventListener("click", async () => {
  await api("/api/controller/logout", { method: "POST", body: "{}" });
  showDash(false);
});

document.querySelector(".tabs").addEventListener("click", (e) => {
  const b = e.target.closest("[data-tab]");
  if (!b) return;
  tab = b.dataset.tab;
  render();
});

$("panel").addEventListener("submit", async (e) => {
  const form = e.target;
  if (!(form instanceof HTMLFormElement)) return;
  e.preventDefault();
  const data = formObj(form);
  try {
    if (form.id === "settings-form") {
      await api("/api/controller/settings", { method: "PATCH", body: JSON.stringify(data) });
    } else if (form.classList.contains("page-form")) {
      await api("/api/controller/pages/" + form.dataset.id, { method: "PUT", body: JSON.stringify(data) });
    } else if (form.classList.contains("card-form")) {
      await api("/api/controller/cards/" + form.dataset.id, { method: "PATCH", body: JSON.stringify(data) });
    } else if (form.classList.contains("card-new")) {
      data.visible = data.visible !== false;
      await api("/api/controller/cards", { method: "POST", body: JSON.stringify(data) });
    } else if (form.classList.contains("news-form")) {
      await api("/api/controller/news/" + form.dataset.id, { method: "PATCH", body: JSON.stringify(data) });
    } else if (form.id === "news-new") {
      data.visible = true;
      await api("/api/controller/news", { method: "POST", body: JSON.stringify(data) });
    } else if (form.classList.contains("person-form")) {
      await api("/api/controller/people/" + form.dataset.id, { method: "PATCH", body: JSON.stringify(data) });
    } else if (form.id === "person-new") {
      data.visible = true;
      await api("/api/controller/people", { method: "POST", body: JSON.stringify(data) });
    }
    await refresh();
  } catch (err) {
    alert(err.message);
  }
});

$("panel").addEventListener("click", async (e) => {
  const t = e.target;
  if (!(t instanceof HTMLElement)) return;
  try {
    if (t.dataset.delCard) await api("/api/controller/cards/" + t.dataset.delCard, { method: "DELETE" });
    else if (t.dataset.delNews) await api("/api/controller/news/" + t.dataset.delNews, { method: "DELETE" });
    else if (t.dataset.delPerson) await api("/api/controller/people/" + t.dataset.delPerson, { method: "DELETE" });
    else if (t.dataset.sub) await api("/api/controller/subscribers/" + t.dataset.sub, { method: "PATCH", body: JSON.stringify({ status: t.dataset.status }) });
    else if (t.id === "btn-indico" || t.id === "btn-docs" || t.id === "btn-seed") {
      const msg = $("index-msg");
      msg.textContent = "Working…";
      const path = t.id === "btn-indico" ? "/api/controller/refresh/indico" : t.id === "btn-docs" ? "/api/controller/refresh/docs" : "/api/controller/refresh/seed";
      const out = await api(path, { method: "POST", body: "{}" });
      msg.textContent = JSON.stringify(out.index || out);
      await refresh();
      return;
    } else return;
    await refresh();
  } catch (err) {
    alert(err.message);
  }
});

api("/api/controller/state").then((data) => {
  state = data;
  showDash(true);
  render();
}).catch(() => showDash(false));
