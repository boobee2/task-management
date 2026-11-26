"use strict";

/* TaskFlow - frontend Task Management System
   HTML + CSS + JavaScript biasa (tanpa framework).
   Ubah API_URL kalau backend berjalan di alamat lain. */
const API_URL = "http://localhost:8080/api/v1";

/* ---------- util ---------- */
const $ = (sel, root = document) => root.querySelector(sel);

const esc = (v) =>
  String(v ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

const qs = (obj) => {
  const p = new URLSearchParams();
  Object.entries(obj).forEach(([k, v]) => {
    if (v !== "" && v != null) p.set(k, v);
  });
  const s = p.toString();
  return s ? "?" + s : "";
};

const fmtDate = (s) => {
  if (!s) return "-";
  const d = new Date(s);
  return isNaN(d) ? s : d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
};

const fmtDue = (s) => {
  if (!s) return "-";
  const [y, m, d] = s.split("-");
  return `${d}/${m}/${y}`;
};

const debounce = (fn, ms = 350) => {
  let t;
  return (...a) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...a), ms);
  };
};

const options = (pairs, selected) =>
  pairs.map(([v, l]) => `<option value="${esc(v)}" ${String(selected) === String(v) ? "selected" : ""}>${esc(l)}</option>`).join("");

const STATUS_LABEL = { TODO: "To Do", IN_PROGRESS: "Dikerjakan", DONE: "Selesai", LOW: "Rendah", MEDIUM: "Sedang", HIGH: "Tinggi" };
const STATUS_OPTS = [["TODO", "To Do"], ["IN_PROGRESS", "Dikerjakan"], ["DONE", "Selesai"]];
const PRIORITY_OPTS = [["LOW", "Rendah"], ["MEDIUM", "Sedang"], ["HIGH", "Tinggi"]];

const badge = (kind, val) => `<span class="badge ${kind}-${esc(val)}">${esc(STATUS_LABEL[val] || val)}</span>`;

/* ---------- toast ---------- */
let toastTimer;
function toast(msg, type = "ok") {
  const el = $("#toast");
  el.textContent = msg;
  el.className = "toast " + type;
  el.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (el.hidden = true), 3500);
}

/* ---------- penyimpanan sesi ---------- */
const store = {
  get token() {
    try { return localStorage.getItem("token"); } catch { return null; }
  },
  get user() {
    try { return JSON.parse(localStorage.getItem("user") || "null"); } catch { return null; }
  },
  save(token, user) {
    try {
      localStorage.setItem("token", token);
      localStorage.setItem("user", JSON.stringify(user));
    } catch { /* abaikan */ }
  },
  clear() {
    try {
      localStorage.removeItem("token");
      localStorage.removeItem("user");
    } catch { /* abaikan */ }
  },
};

/* ---------- pemanggil API ---------- */
async function api(path, { method = "GET", body, auth = true } = {}) {
  const headers = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (auth && store.token) headers.Authorization = "Bearer " + store.token;

  let res;
  try {
    res = await fetch(API_URL + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    return { ok: false, status: 0, data: { error: "Tidak dapat terhubung ke server. Pastikan backend berjalan dan CORS sudah diatur." } };
  }

  let data = null;
  try { data = await res.json(); } catch { /* body kosong */ }

  // 401 pada endpoint privat: token tidak valid atau kedaluwarsa -> kembali ke login
  if (res.status === 401 && auth) {
    store.clear();
    location.hash = "#/login";
    toast("Sesi berakhir, silakan masuk lagi.", "error");
  }
  return { ok: res.ok, status: res.status, data: data || {} };
}

const errMsg = (res) =>
  res.status === 403 ? res.data.error || "Anda tidak punya akses." : res.data.error || "Terjadi kesalahan.";

function showError(res) {
  toast(errMsg(res), "error");
  const m = $(".main");
  if (m) m.innerHTML = `<div class="empty">${esc(errMsg(res))}</div>`;
}

/* ---------- komponen umum ---------- */
let viewId = 0;
const guard = () => {
  const id = viewId;
  return () => id !== viewId; // true kalau halaman sudah berpindah
};

const ICONS = {
  grid: '<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
  tasks: '<rect x="3" y="4" width="18" height="16" rx="3"/><path d="M8 12l3 3 5-6"/>',
  tag: '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>',
  users: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  eye: '<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>',
  logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>',
  menu: '<line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>',
  chevrons: '<polyline points="11 17 6 12 11 7"/><polyline points="18 17 13 12 18 7"/>',
  alert: '<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>',
  help: '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.82 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/>',
};

const icon = (name) => `<svg class="ico" viewBox="0 0 24 24" aria-hidden="true">${ICONS[name]}</svg>`;

const NAV = {
  USER: [["#/dashboard", "Dashboard", "grid"], ["#/tasks", "Task", "tasks"], ["#/categories", "Kategori", "tag"]],
  ADMIN: [["#/admin", "Dashboard", "grid"], ["#/admin/users", "Kelola user", "users"], ["#/admin/tasks", "Monitoring task", "eye"]],
};

function sidebarCollapsed() {
  try { return localStorage.getItem("sidebar") === "collapsed"; } catch { return false; }
}

function shell(active, content) {
  const u = store.user;
  const collapsed = sidebarCollapsed();
  const links = NAV[u.role]
    .map(([h, l, ic]) => `<a href="${h}" class="${h === active ? "active" : ""}" title="${esc(l)}">${icon(ic)}<span class="nav-label">${esc(l)}</span></a>`)
    .join("");

  $("#app").innerHTML = `
    <div class="layout${collapsed ? " collapsed" : ""}" id="layout">
      <aside class="sidebar" id="sidebar">
        <div class="side-head">
          <span class="logo-mark" aria-hidden="true">T</span>
          <span class="brand nav-label">TaskFlow</span>
          <button class="icon-btn collapse-btn" id="collapse" aria-label="Ciutkan atau lebarkan menu" title="Ciutkan atau lebarkan menu">${icon("chevrons")}</button>
        </div>
        <nav aria-label="Menu utama">${links}</nav>
        <div class="me">
          <div class="who">
            <span class="avatar" aria-hidden="true">${esc((u.name || "?").trim().charAt(0).toUpperCase())}</span>
            <span class="who-text nav-label"><strong>${esc(u.name)}</strong><span class="badge role-${esc(u.role)}">${esc(u.role)}</span></span>
          </div>
          <button class="logout" id="logout" title="Keluar">${icon("logout")}<span class="nav-label">Keluar</span></button>
        </div>
      </aside>
      <div class="backdrop" id="backdrop"></div>
      <div class="content">
        <header class="topbar">
          <button class="icon-btn" id="burger" aria-label="Buka menu">${icon("menu")}</button>
          <span class="brand">TaskFlow</span>
        </header>
        <main class="main">${content}</main>
      </div>
    </div>`;

  const layout = $("#layout");
  $("#collapse").onclick = () => {
    const now = layout.classList.toggle("collapsed");
    try { localStorage.setItem("sidebar", now ? "collapsed" : "expanded"); } catch { /* abaikan */ }
  };
  $("#burger").onclick = () => layout.classList.add("open");
  $("#backdrop").onclick = () => layout.classList.remove("open");
  document.querySelectorAll(".sidebar nav a").forEach((a) => (a.onclick = () => layout.classList.remove("open")));
  $("#logout").onclick = () => {
    store.clear();
    location.hash = "#/login";
  };
}

function modal(title, fields, onSave) {
  const el = document.createElement("div");
  el.className = "overlay";
  el.innerHTML = `
    <div class="modal" role="dialog" aria-modal="true" aria-label="${esc(title)}">
      <h2>${esc(title)}</h2>
      <form id="mform" class="form">
        ${fields}
        <p class="error" id="merr"></p>
        <div class="actions">
          <button type="button" class="btn ghost" id="mcancel">Batal</button>
          <button type="submit" class="btn">Simpan</button>
        </div>
      </form>
    </div>`;
  document.body.appendChild(el);
  const close = () => el.remove();
  $("#mcancel", el).onclick = close;
  el.addEventListener("mousedown", (e) => { if (e.target === el) close(); });
  const first = $("input, select, textarea", el);
  if (first) first.focus();

  $("#mform", el).onsubmit = async (e) => {
    e.preventDefault();
    const submit = $("button[type=submit]", el);
    submit.disabled = true;
    const err = await onSave(new FormData(e.target));
    submit.disabled = false;
    if (err) $("#merr", el).textContent = err;
    else close();
  };
}

// Kotak konfirmasi bergaya aplikasi (pengganti confirm bawaan browser). Mengembalikan Promise<boolean>.
function confirmBox(message, { title = "Konfirmasi", confirmText = "Ya, lanjutkan", danger = false } = {}) {
  return new Promise((resolve) => {
    const el = document.createElement("div");
    el.className = "overlay";
    el.innerHTML = `
      <div class="modal confirm${danger ? " is-danger" : ""}" role="alertdialog" aria-modal="true" aria-labelledby="ctitle" aria-describedby="cmsg">
        <div class="confirm-icon" aria-hidden="true">${icon(danger ? "alert" : "help")}</div>
        <h2 id="ctitle">${esc(title)}</h2>
        <p id="cmsg" class="muted">${esc(message)}</p>
        <div class="actions">
          <button type="button" class="btn ghost" id="cno">Batal</button>
          <button type="button" class="btn" id="cyes">${esc(confirmText)}</button>
        </div>
      </div>`;

    const onKey = (e) => { if (e.key === "Escape") done(false); };
    const done = (value) => {
      document.removeEventListener("keydown", onKey);
      el.remove();
      resolve(value);
    };

    document.addEventListener("keydown", onKey);
    document.body.appendChild(el);
    $("#cno", el).onclick = () => done(false);
    $("#cyes", el).onclick = () => done(true);
    el.addEventListener("mousedown", (e) => { if (e.target === el) done(false); });
    $(danger ? "#cno" : "#cyes", el).focus(); // aksi berbahaya: fokus awal di tombol Batal
  });
}

function pagerHtml(p) {
  const pages = Math.max(p.total_pages, 1);
  return `<div class="pager">
    <span>${p.total} data, halaman ${p.page} dari ${pages}</span>
    <div>
      <button class="btn ghost small" data-page="${p.page - 1}" ${p.page <= 1 ? "disabled" : ""}>Sebelumnya</button>
      <button class="btn ghost small" data-page="${p.page + 1}" ${p.page >= p.total_pages ? "disabled" : ""}>Berikutnya</button>
    </div>
  </div>`;
}

function bindPager(root, onPage) {
  root.querySelectorAll("[data-page]").forEach((b) => (b.onclick = () => onPage(Number(b.dataset.page))));
}

const statCard = (label, value, tone = "red") =>
  `<div class="card stat tone-${tone}"><strong>${esc(value)}</strong><span>${esc(label)}</span></div>`;

// Grafik donat dengan wajah yang berubah sesuai persentase task selesai
function donutChart(t, friendly = true) {
  const R = 15.9155; // keliling lingkaran = 100, jadi panjang dash = persen
  const total = t.total;
  const donePct = total ? Math.round((t.done / total) * 100) : 0;
  const segs = [
    { n: t.todo, cls: "todo", label: "To Do" },
    { n: t.in_progress, cls: "progress", label: "Dikerjakan" },
    { n: t.done, cls: "done", label: "Selesai" },
  ];
  const active = segs.filter((s) => s.n > 0);

  let rings = "";
  let offset = 0;
  active.forEach((s) => {
    const pct = (s.n / total) * 100;
    if (active.length === 1) {
      rings += `<circle class="seg seg-${s.cls}" cx="21" cy="21" r="${R}"/>`;
    } else {
      const dash = Math.max(pct - 7, 0.01); // sisakan ruang untuk ujung bulat dan celah
      rings += `<circle class="seg seg-${s.cls}" cx="21" cy="21" r="${R}" stroke-linecap="round" stroke-dasharray="${dash.toFixed(2)} ${(100 - dash).toFixed(2)}" stroke-dashoffset="${(-(offset + 3.5)).toFixed(2)}"/>`;
    }
    offset += pct;
  });

  const cheeks = `<circle class="cheek" cx="13.8" cy="24.2" r="1.9"/><circle class="cheek" cx="28.2" cy="24.2" r="1.9"/>`;
  const dotEyes = `<circle class="fill" cx="17" cy="19.5" r="1.5"/><circle class="fill" cx="25" cy="19.5" r="1.5"/>`;
  let face;
  if (!total) face = `<path d="M14.5 19.5q2.2 1.8 4.4 0M23.1 19.5q2.2 1.8 4.4 0"/><path d="M19.5 25.5h3"/>`; // mengantuk
  else if (donePct >= 70) face = `<path d="M14.8 20q2.2-2.8 4.4 0M22.8 20q2.2-2.8 4.4 0"/><path class="fill" d="M16.5 23.6q4.5 6 9 0z"/>`; // senang
  else if (donePct >= 30) face = `${dotEyes}<path d="M17 24q4 3.8 8 0"/>`; // tersenyum
  else face = `${dotEyes}<ellipse class="fill" cx="21" cy="25.4" rx="1.7" ry="2"/>`; // bersemangat

  const msg = !total
    ? "Belum ada task, yuk mulai!"
    : donePct === 100
      ? "Semua beres, kerja bagus!"
      : donePct >= 70
        ? "Hampir selesai, semangat!"
        : donePct >= 30
          ? "Terus lanjut, kamu bisa!"
          : "Ayo kerjakan satu per satu.";

  const legend = segs.map((s) => `<li><span class="dot dot-${s.cls}"></span><span>${s.label}</span><b>${s.n}</b></li>`).join("");

  return `
    <div class="chart">
      <svg class="donut" viewBox="0 0 42 42" role="img" aria-label="Grafik progres task: ${donePct}% selesai">
        <circle class="track" cx="21" cy="21" r="${R}"/>
        <g transform="rotate(-90 21 21)">${rings}</g>
        <g class="face">${cheeks}${face}</g>
      </svg>
      <div class="donut-caption"><strong>${donePct}%</strong><span>selesai</span></div>
      <p class="muted center small">${friendly ? msg : `dari ${total} task di sistem`}</p>
      <ul class="legend2">${legend}</ul>
    </div>`;
}

// Nama lama dipertahankan agar halaman lain yang masih memanggil progressBar tetap jalan
const progressBar = (t) => donutChart(t, false);

/* =====================================================
   Router + route guard
   (menyembunyikan menu hanya kenyamanan UI; keamanan tetap di backend)
   ===================================================== */
const home = () => (store.user && store.user.role === "ADMIN" ? "#/admin" : "#/dashboard");

const routes = {
  "#/login": { view: loginView, public: true },
  "#/register": { view: registerView, public: true },
  "#/dashboard": { view: userDashboard, role: "USER" },
  "#/tasks": { view: tasksView, role: "USER" },
  "#/categories": { view: categoriesView, role: "USER" },
  "#/admin": { view: adminDashboard, role: "ADMIN" },
  "#/admin/users": { view: adminUsersView, role: "ADMIN" },
  "#/admin/tasks": { view: adminTasksView, role: "ADMIN" },
};

function router() {
  viewId++;
  let r = routes[location.hash || "#/login"];
  const m = (location.hash || "").match(/^#\/admin\/users\/(\d+)$/);
  if (!r && m) r = { view: () => adminUserDetail(Number(m[1])), role: "ADMIN" };
  const user = store.user;
  const loggedIn = !!(store.token && user);

  if (!r) {
    location.hash = loggedIn ? home() : "#/login";
    return;
  }
  if (r.public) {
    if (loggedIn) { location.hash = home(); return; }
    return r.view();
  }
  if (!loggedIn) { location.hash = "#/login"; return; }
  if (r.role !== user.role) {
    toast("Anda tidak punya akses ke halaman itu.", "error");
    location.hash = home();
    return;
  }
  r.view();
}

window.addEventListener("hashchange", router);

/* =====================================================
   Login & registrasi
   ===================================================== */
function authShell(title, sub, formHtml, footer) {
  $("#app").innerHTML = `
    <div class="auth">
      <div class="card auth-card">
        <div class="brand big">TaskFlow</div>
        <h1>${title}</h1>
        <p class="muted">${sub}</p>
        ${formHtml}
        <p class="muted center">${footer}</p>
      </div>
    </div>`;
}

function loginView() {
  authShell(
    "Masuk",
    "Kelola tugasmu dengan rapi.",
    `<form id="f" class="form">
      <label>Email<input type="email" name="email" required autocomplete="username"></label>
      <label>Password<input type="password" name="password" required autocomplete="current-password"></label>
      <p class="error" id="err"></p>
      <button class="btn" type="submit">Masuk</button>
    </form>`,
    `Belum punya akun? <a href="#/register">Daftar</a>`
  );

  $("#f").onsubmit = async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const btn = $("button", e.target);
    btn.disabled = true;
    const r = await api("/auth/login", {
      method: "POST",
      auth: false,
      body: { email: fd.get("email").trim(), password: fd.get("password") },
    });
    btn.disabled = false;
    if (!r.ok) {
      $("#err").textContent = errMsg(r);
      return;
    }
    store.save(r.data.token, r.data.user);
    location.hash = home(); // USER ke dashboard pribadi, ADMIN ke dashboard admin
  };
}

function registerView() {
  authShell(
    "Buat akun",
    "Daftar untuk mulai mencatat tugas.",
    `<form id="f" class="form">
      <label>Nama<input name="name" maxlength="100" required></label>
      <label>Email<input type="email" name="email" maxlength="150" required></label>
      <label>Password<input type="password" name="password" minlength="8" maxlength="72" required autocomplete="new-password"></label>
      <p class="muted small">Minimal 8 karakter.</p>
      <p class="error" id="err"></p>
      <button class="btn" type="submit">Daftar</button>
    </form>`,
    `Sudah punya akun? <a href="#/login">Masuk</a>`
  );

  $("#f").onsubmit = async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const btn = $("button", e.target);
    btn.disabled = true;
    // role tidak dikirim: backend selalu menetapkan USER
    const r = await api("/auth/register", {
      method: "POST",
      auth: false,
      body: { name: fd.get("name").trim(), email: fd.get("email").trim(), password: fd.get("password") },
    });
    btn.disabled = false;
    if (!r.ok) {
      $("#err").textContent = errMsg(r);
      return;
    }
    toast("Akun dibuat. Silakan masuk.");
    location.hash = "#/login";
  };
}

/* =====================================================
   USER: dashboard
   ===================================================== */
async function userDashboard() {
  const stale = guard();
  shell("#/dashboard", `<h1>Dashboard</h1><p class="muted">Memuat...</p>`);
  const r = await api("/dashboard");
  if (stale()) return;
  if (!r.ok) return showError(r);

  const d = r.data.data;
  const t = d.tasks;
  const list = (items, empty) =>
    items.length
      ? `<ul class="plain">${items
          .map(
            (x) => `<li>
              <div><strong>${esc(x.title)}</strong><div class="muted small">Deadline: ${fmtDue(x.due_date)}</div></div>
              <div class="badges">${badge("status", x.status)} ${badge("priority", x.priority)}</div>
            </li>`
          )
          .join("")}</ul>`
      : `<p class="muted">${empty}</p>`;

  shell(
    "#/dashboard",
    `<div class="page-head">
      <div><h1>Halo, ${esc(store.user.name)}</h1><p class="muted">Ringkasan task milikmu.</p></div>
      <a class="btn" href="#/tasks">Lihat semua task</a>
    </div>
    <div class="stats">
      ${statCard("Total task", t.total, "red")}
      ${statCard("To Do", t.todo, "orange")}
      ${statCard("Dikerjakan", t.in_progress, "yellow")}
      ${statCard("Selesai", t.done, "green")}
    </div>
    <div class="dash-grid">
      <section class="card panel chart-panel"><h2>Progres task</h2>${donutChart(t, true)}</section>
      <section class="card panel"><h2>Deadline terdekat</h2><div class="scroll">${list(d.upcoming_tasks, "Tidak ada task yang punya deadline.")}</div></section>
      <section class="card panel"><h2>Task terbaru</h2><div class="scroll">${list(d.recent_tasks, "Belum ada task.")}</div></section>
    </div>`
  );
}

/* =====================================================
   USER: task
   ===================================================== */
const tState = { page: 1, limit: 8, search: "", status: "", priority: "", category_id: "", sort_by: "created_at", sort_order: "desc", items: [], cats: [] };
const SORT_OPTS = [
  ["created_at:desc", "Terbaru"],
  ["created_at:asc", "Terlama"],
  ["due_date:asc", "Deadline terdekat"],
  ["title:asc", "Judul A-Z"],
  ["title:desc", "Judul Z-A"],
];

async function tasksView() {
  const stale = guard();
  shell("#/tasks", `<h1>Task</h1><p class="muted">Memuat...</p>`);
  const cr = await api("/categories");
  if (stale()) return;
  if (!cr.ok) return showError(cr);
  tState.cats = cr.data.data;

  shell(
    "#/tasks",
    `<h1>Task</h1>
    <p class="muted">Cari, saring, dan atur task milikmu.</p>
    <div class="toolbar">
      <input id="fSearch" type="search" placeholder="Cari judul task" value="${esc(tState.search)}" aria-label="Cari judul">
      <select id="fStatus" aria-label="Status">${options([["", "Semua status"], ...STATUS_OPTS], tState.status)}</select>
      <select id="fPriority" aria-label="Prioritas">${options([["", "Semua prioritas"], ...PRIORITY_OPTS], tState.priority)}</select>
      <select id="fCategory" aria-label="Kategori">${options([["", "Semua kategori"], ...tState.cats.map((c) => [c.id, c.name])], tState.category_id)}</select>
      <select id="fSort" aria-label="Urutan">${options(SORT_OPTS, tState.sort_by + ":" + tState.sort_order)}</select>
      <button class="btn" id="addTask">Tambah task</button>
    </div>
    <div id="list"></div>`
  );

  const reload = () => { tState.page = 1; loadTasks(stale); };
  $("#fSearch").oninput = debounce((e) => { tState.search = e.target.value.trim(); reload(); });
  $("#fStatus").onchange = (e) => { tState.status = e.target.value; reload(); };
  $("#fPriority").onchange = (e) => { tState.priority = e.target.value; reload(); };
  $("#fCategory").onchange = (e) => { tState.category_id = e.target.value; reload(); };
  $("#fSort").onchange = (e) => {
    [tState.sort_by, tState.sort_order] = e.target.value.split(":");
    reload();
  };
  $("#addTask").onclick = () => taskForm(null, stale);
  loadTasks(stale);
}

async function loadTasks(stale) {
  const box = $("#list");
  if (!box) return;
  box.innerHTML = `<p class="muted">Memuat...</p>`;

  const q = { page: tState.page, limit: tState.limit, search: tState.search, status: tState.status, priority: tState.priority, category_id: tState.category_id, sort_by: tState.sort_by, sort_order: tState.sort_order };
  const r = await api("/tasks" + qs(q));
  if (stale()) return;
  if (!r.ok) { box.innerHTML = `<div class="empty">${esc(errMsg(r))}</div>`; return; }

  const items = r.data.data;
  const p = r.data.pagination;
  if (!items.length && tState.page > 1 && p.total > 0) {
    tState.page = p.total_pages; // halaman terakhir sudah kosong setelah hapus
    return loadTasks(stale);
  }
  tState.items = items;
  if (!items.length) {
    box.innerHTML = `<div class="empty">Belum ada task yang cocok. Ubah filter atau tambah task baru.</div>`;
    return;
  }

  const catName = (id) => (tState.cats.find((c) => c.id === id) || {}).name || "-";
  box.innerHTML = `
    <div class="table-wrap"><table>
      <thead><tr><th>Judul</th><th>Kategori</th><th>Prioritas</th><th>Status</th><th>Deadline</th><th>Aksi</th></tr></thead>
      <tbody>${items
        .map(
          (t) => `<tr>
            <td><strong>${esc(t.title)}</strong>${t.description ? `<div class="desc">${esc(t.description)}</div>` : ""}</td>
            <td>${esc(catName(t.category_id))}</td>
            <td>${badge("priority", t.priority)}</td>
            <td>${badge("status", t.status)}</td>
            <td>${fmtDue(t.due_date)}</td>
            <td class="actions">
              <button class="btn ghost small" data-edit="${t.id}">Ubah</button>
              <button class="btn danger small" data-del="${t.id}">Hapus</button>
            </td>
          </tr>`
        )
        .join("")}</tbody>
    </table></div>
    ${pagerHtml(p)}`;

  bindPager(box, (pg) => { tState.page = pg; loadTasks(stale); });
  box.querySelectorAll("[data-edit]").forEach((b) => {
    b.onclick = () => taskForm(tState.items.find((t) => t.id === Number(b.dataset.edit)), stale);
  });
  box.querySelectorAll("[data-del]").forEach((b) => {
    b.onclick = async () => {
      const task = tState.items.find((x) => x.id === Number(b.dataset.del));
      const ok = await confirmBox(`Task "${task ? task.title : "ini"}" akan dihapus permanen dan tidak bisa dikembalikan.`, { title: "Hapus task?", confirmText: "Ya, hapus", danger: true });
      if (!ok) return;
      const res = await api("/tasks/" + b.dataset.del, { method: "DELETE" });
      if (!res.ok) return toast(errMsg(res), "error");
      toast("Task dihapus.");
      loadTasks(stale);
    };
  });
}

function taskForm(task, stale) {
  const t = task || { title: "", description: "", category_id: null, priority: "MEDIUM", status: "TODO", due_date: "" };
  const fields = `
    <label>Judul<input name="title" maxlength="200" required value="${esc(t.title)}"></label>
    <label>Deskripsi<textarea name="description" rows="3">${esc(t.description || "")}</textarea></label>
    <div class="grid2">
      <label>Kategori<select name="category_id">${options([["", "Tanpa kategori"], ...tState.cats.map((c) => [c.id, c.name])], t.category_id ?? "")}</select></label>
      <label>Deadline<input type="date" name="due_date" value="${esc(t.due_date || "")}"></label>
      <label>Prioritas<select name="priority">${options(PRIORITY_OPTS, t.priority)}</select></label>
      <label>Status<select name="status">${options(STATUS_OPTS, t.status)}</select></label>
    </div>`;

  modal(task ? "Ubah task" : "Task baru", fields, async (fd) => {
    const body = {
      title: fd.get("title").trim(),
      description: fd.get("description").trim() || null,
      category_id: fd.get("category_id") ? Number(fd.get("category_id")) : null,
      priority: fd.get("priority"),
      status: fd.get("status"),
      due_date: fd.get("due_date") || null,
    };
    const r = await api(task ? "/tasks/" + task.id : "/tasks", { method: task ? "PUT" : "POST", body });
    if (!r.ok) return errMsg(r);
    toast(task ? "Task diperbarui." : "Task dibuat.");
    loadTasks(stale);
    return null;
  });
}

/* =====================================================
   USER: kategori
   ===================================================== */
async function categoriesView() {
  const stale = guard();
  shell("#/categories", `<h1>Kategori</h1><p class="muted">Memuat...</p>`);

  const load = async () => {
    const r = await api("/categories");
    if (stale()) return;
    if (!r.ok) return showError(r);
    const cats = r.data.data;

    shell(
      "#/categories",
      `<h1>Kategori</h1>
      <p class="muted">Kelompokkan task-mu. Menghapus kategori tidak menghapus task di dalamnya.</p>
      <div class="toolbar"><button class="btn" id="addCat">Tambah kategori</button></div>
      ${
        cats.length
          ? `<div class="table-wrap"><table>
              <thead><tr><th>Nama</th><th>Dibuat</th><th>Aksi</th></tr></thead>
              <tbody>${cats
                .map(
                  (c) => `<tr>
                    <td><strong>${esc(c.name)}</strong></td>
                    <td>${fmtDate(c.created_at)}</td>
                    <td class="actions">
                      <button class="btn ghost small" data-edit="${c.id}">Ubah</button>
                      <button class="btn danger small" data-del="${c.id}">Hapus</button>
                    </td>
                  </tr>`
                )
                .join("")}</tbody>
            </table></div>`
          : `<div class="empty">Belum ada kategori. Tambahkan yang pertama, misalnya Kuliah atau Organisasi.</div>`
      }`
    );

    $("#addCat").onclick = () => catForm(null);
    document.querySelectorAll("[data-edit]").forEach((b) => {
      b.onclick = () => catForm(cats.find((c) => c.id === Number(b.dataset.edit)));
    });
    document.querySelectorAll("[data-del]").forEach((b) => {
      b.onclick = async () => {
        const cat = cats.find((c) => c.id === Number(b.dataset.del));
        const ok = await confirmBox(`Kategori "${cat ? cat.name : "ini"}" akan dihapus. Task di dalamnya tetap ada tanpa kategori.`, { title: "Hapus kategori?", confirmText: "Ya, hapus", danger: true });
        if (!ok) return;
        const res = await api("/categories/" + b.dataset.del, { method: "DELETE" });
        if (!res.ok) return toast(errMsg(res), "error");
        toast("Kategori dihapus.");
        load();
      };
    });
  };

  const catForm = (cat) => {
    modal(cat ? "Ubah kategori" : "Kategori baru", `<label>Nama<input name="name" maxlength="100" required value="${esc(cat ? cat.name : "")}"></label>`, async (fd) => {
      const body = { name: fd.get("name").trim() };
      const r = await api(cat ? "/categories/" + cat.id : "/categories", { method: cat ? "PUT" : "POST", body });
      if (!r.ok) return errMsg(r);
      toast(cat ? "Kategori diperbarui." : "Kategori dibuat.");
      load();
      return null;
    });
  };

  load();
}

/* =====================================================
   ADMIN: dashboard global
   ===================================================== */
async function adminDashboard() {
  const stale = guard();
  shell("#/admin", `<h1>Dashboard admin</h1><p class="muted">Memuat...</p>`);
  const r = await api("/admin/dashboard");
  if (stale()) return;
  if (!r.ok) return showError(r);

  const { users: u, tasks: t, recent_users: recent } = r.data.data;
  shell(
    "#/admin",
    `<div class="page-head">
      <div><h1>Dashboard admin</h1><p class="muted">Statistik seluruh sistem.</p></div>
      <a class="btn" href="#/admin/users">Kelola user</a>
    </div>
    <div class="stats">
      ${statCard("Total user", u.total, "red")}
      ${statCard("User aktif", u.active, "green")}
      ${statCard("Nonaktif", u.inactive, "orange")}
      ${statCard("Admin", u.admins, "yellow")}
      ${statCard("Baru 7 hari terakhir", u.new_last_7_days, "red")}
    </div>
    <div class="dash-grid admin">
      <section class="card panel chart-panel"><h2>Progres seluruh task</h2>${donutChart(t, false)}</section>
      <section class="card panel">
        <h2>User terbaru</h2>
        <div class="table-wrap flat"><table>
          <thead><tr><th>Nama</th><th>Email</th><th>Role</th><th>Status</th><th>Bergabung</th></tr></thead>
          <tbody>${recent
            .map(
              (x) => `<tr>
                <td><a href="#/admin/users/${x.id}"><strong>${esc(x.name)}</strong></a></td><td>${esc(x.email)}</td>
                <td><span class="badge role-${esc(x.role)}">${esc(x.role)}</span></td>
                <td><span class="badge ${x.is_active ? "on" : "off"}">${x.is_active ? "Aktif" : "Nonaktif"}</span></td>
                <td>${fmtDate(x.created_at)}</td>
              </tr>`
            )
            .join("")}</tbody>
        </table></div>
      </section>
    </div>`
  );
}

/* =====================================================
   ADMIN: kelola user
   ===================================================== */
const uState = { page: 1, limit: 10, search: "", role: "", is_active: "", sort_by: "created_at", sort_order: "desc" };

async function adminUsersView() {
  const stale = guard();
  shell(
    "#/admin/users",
    `<h1>Kelola user</h1>
    <p class="muted">Ubah role atau aktifkan dan nonaktifkan akun.</p>
    <div class="toolbar">
      <input id="fSearch" type="search" placeholder="Cari nama atau email" value="${esc(uState.search)}" aria-label="Cari user">
      <select id="fRole" aria-label="Role">${options([["", "Semua role"], ["USER", "USER"], ["ADMIN", "ADMIN"]], uState.role)}</select>
      <select id="fActive" aria-label="Status akun">${options([["", "Semua status"], ["true", "Aktif"], ["false", "Nonaktif"]], uState.is_active)}</select>
      <select id="fSort" aria-label="Urutan">${options([["created_at:desc", "Terbaru"], ["created_at:asc", "Terlama"], ["name:asc", "Nama A-Z"], ["name:desc", "Nama Z-A"]], uState.sort_by + ":" + uState.sort_order)}</select>
    </div>
    <div id="list"></div>`
  );

  const reload = () => { uState.page = 1; loadUsers(stale); };
  $("#fSearch").oninput = debounce((e) => { uState.search = e.target.value.trim(); reload(); });
  $("#fRole").onchange = (e) => { uState.role = e.target.value; reload(); };
  $("#fActive").onchange = (e) => { uState.is_active = e.target.value; reload(); };
  $("#fSort").onchange = (e) => { [uState.sort_by, uState.sort_order] = e.target.value.split(":"); reload(); };
  loadUsers(stale);
}

async function loadUsers(stale) {
  const box = $("#list");
  if (!box) return;
  box.innerHTML = `<p class="muted">Memuat...</p>`;

  const r = await api("/admin/users" + qs({ page: uState.page, limit: uState.limit, search: uState.search, role: uState.role, is_active: uState.is_active, sort_by: uState.sort_by, sort_order: uState.sort_order }));
  if (stale()) return;
  if (!r.ok) { box.innerHTML = `<div class="empty">${esc(errMsg(r))}</div>`; return; }

  const users = r.data.data;
  const p = r.data.pagination;
  if (!users.length) {
    box.innerHTML = `<div class="empty">Tidak ada user yang cocok dengan filter.</div>`;
    return;
  }

  const meId = Number(store.user.id);
  box.innerHTML = `
    <div class="table-wrap"><table>
      <thead><tr><th>Nama</th><th>Email</th><th>Role</th><th>Status</th><th>Bergabung</th><th>Aksi</th></tr></thead>
      <tbody>${users
        .map((u) => {
          const self = Number(u.id) === meId;
          return `<tr>
            <td><a href="#/admin/users/${u.id}"><strong>${esc(u.name)}</strong></a>${self ? ` <span class="muted small">(Anda)</span>` : ""}</td>
            <td>${esc(u.email)}</td>
            <td><span class="badge role-${esc(u.role)}">${esc(u.role)}</span></td>
            <td><span class="badge ${u.is_active ? "on" : "off"}">${u.is_active ? "Aktif" : "Nonaktif"}</span></td>
            <td>${fmtDate(u.created_at)}</td>
            <td class="actions">
              <button class="btn ghost small" data-role="${u.id}" data-new="${u.role === "ADMIN" ? "USER" : "ADMIN"}" ${self ? "disabled title=\"Tidak bisa mengubah akun sendiri\"" : ""}>${u.role === "ADMIN" ? "Jadikan USER" : "Jadikan ADMIN"}</button>
              <button class="btn ${u.is_active ? "danger" : "ghost"} small" data-status="${u.id}" data-new="${u.is_active ? "false" : "true"}" ${self ? "disabled title=\"Tidak bisa mengubah akun sendiri\"" : ""}>${u.is_active ? "Nonaktifkan" : "Aktifkan"}</button>
            </td>
          </tr>`;
        })
        .join("")}</tbody>
    </table></div>
    ${pagerHtml(p)}`;

  bindPager(box, (pg) => { uState.page = pg; loadUsers(stale); });

  box.querySelectorAll("[data-role]").forEach((b) => {
    b.onclick = async () => {
      const role = b.dataset.new;
      const ok = await confirmBox(`Role user ini akan diubah menjadi ${role}.`, { title: "Ubah role?", confirmText: "Ya, ubah" });
      if (!ok) return;
      const res = await api(`/admin/users/${b.dataset.role}/role`, { method: "PATCH", body: { role } });
      if (!res.ok) return toast(errMsg(res), "error");
      toast("Role diubah.");
      loadUsers(stale);
    };
  });
  box.querySelectorAll("[data-status]").forEach((b) => {
    b.onclick = async () => {
      const active = b.dataset.new === "true";
      const ok = await confirmBox(
        active ? "Akun ini akan diaktifkan kembali dan user bisa masuk lagi." : "User tidak akan bisa masuk lagi sampai akunnya diaktifkan.",
        { title: active ? "Aktifkan akun?" : "Nonaktifkan akun?", confirmText: active ? "Ya, aktifkan" : "Ya, nonaktifkan", danger: !active }
      );
      if (!ok) return;
      const res = await api(`/admin/users/${b.dataset.status}/status`, { method: "PATCH", body: { is_active: active } });
      if (!res.ok) return toast(errMsg(res), "error");
      toast(active ? "Akun diaktifkan." : "Akun dinonaktifkan.");
      loadUsers(stale);
    };
  });
}

/* =====================================================
   ADMIN: monitoring task (read-only)
   ===================================================== */
const aState = { page: 1, limit: 10, search: "", user_id: "", status: "", priority: "", sort_by: "created_at", sort_order: "desc" };

async function adminTasksView() {
  const stale = guard();
  shell(
    "#/admin/tasks",
    `<h1>Monitoring task</h1>
    <p class="muted">Seluruh task dari semua user. Admin tidak bisa mengubah task, hanya menghapus task yang melanggar aturan.</p>
    <div class="toolbar">
      <input id="fSearch" type="search" placeholder="Cari judul task" value="${esc(aState.search)}" aria-label="Cari judul">
      <input id="fUser" type="number" min="1" placeholder="ID user" value="${esc(aState.user_id)}" aria-label="Filter ID user">
      <select id="fStatus" aria-label="Status">${options([["", "Semua status"], ...STATUS_OPTS], aState.status)}</select>
      <select id="fPriority" aria-label="Prioritas">${options([["", "Semua prioritas"], ...PRIORITY_OPTS], aState.priority)}</select>
      <select id="fSort" aria-label="Urutan">${options(SORT_OPTS, aState.sort_by + ":" + aState.sort_order)}</select>
    </div>
    <div id="list"></div>`
  );

  const reload = () => { aState.page = 1; loadAdminTasks(stale); };
  $("#fSearch").oninput = debounce((e) => { aState.search = e.target.value.trim(); reload(); });
  $("#fUser").oninput = debounce((e) => { aState.user_id = e.target.value.trim(); reload(); });
  $("#fStatus").onchange = (e) => { aState.status = e.target.value; reload(); };
  $("#fPriority").onchange = (e) => { aState.priority = e.target.value; reload(); };
  $("#fSort").onchange = (e) => { [aState.sort_by, aState.sort_order] = e.target.value.split(":"); reload(); };
  loadAdminTasks(stale);
}

async function loadAdminTasks(stale) {
  const box = $("#list");
  if (!box) return;
  box.innerHTML = `<p class="muted">Memuat...</p>`;

  const r = await api("/admin/tasks" + qs({ page: aState.page, limit: aState.limit, search: aState.search, user_id: aState.user_id, status: aState.status, priority: aState.priority, sort_by: aState.sort_by, sort_order: aState.sort_order }));
  if (stale()) return;
  if (!r.ok) { box.innerHTML = `<div class="empty">${esc(errMsg(r))}</div>`; return; }

  const items = r.data.data;
  const p = r.data.pagination;
  if (!items.length && aState.page > 1 && p.total > 0) {
    aState.page = p.total_pages; // halaman terakhir sudah kosong setelah hapus
    return loadAdminTasks(stale);
  }
  if (!items.length) {
    box.innerHTML = `<div class="empty">Tidak ada task yang cocok dengan filter.</div>`;
    return;
  }

  box.innerHTML = `
    <div class="table-wrap"><table>
      <thead><tr><th>Judul</th><th>Pemilik</th><th>Prioritas</th><th>Status</th><th>Deadline</th><th>Dibuat</th><th>Aksi</th></tr></thead>
      <tbody>${items
        .map(
          (t) => `<tr>
            <td><strong>${esc(t.title)}</strong>${t.description ? `<div class="desc">${esc(t.description)}</div>` : ""}</td>
            <td>${esc(t.owner_name)}<div class="muted small">${esc(t.owner_email)}</div></td>
            <td>${badge("priority", t.priority)}</td>
            <td>${badge("status", t.status)}</td>
            <td>${fmtDue(t.due_date)}</td>
            <td>${fmtDate(t.created_at)}</td>
            <td class="actions"><button class="btn danger small" data-del="${t.id}" data-title="${esc(t.title)}" data-owner="${esc(t.owner_name)}">Hapus</button></td>
          </tr>`
        )
        .join("")}</tbody>
    </table></div>
    ${pagerHtml(p)}`;

  bindPager(box, (pg) => { aState.page = pg; loadAdminTasks(stale); });

  box.querySelectorAll("[data-del]").forEach((b) => {
    b.onclick = async () => {
      const ok = await confirmBox(`Task "${b.dataset.title}" milik ${b.dataset.owner} akan dihapus permanen dan tidak bisa dikembalikan.`, { title: "Hapus task user?", confirmText: "Ya, hapus", danger: true });
      if (!ok) return;
      const res = await api("/admin/tasks/" + b.dataset.del, { method: "DELETE" });
      if (!res.ok) return toast(errMsg(res), "error");
      toast("Task dihapus.");
      loadAdminTasks(stale);
    };
  });
}

/* =====================================================
   ADMIN: detail user
   ===================================================== */
async function adminUserDetail(id) {
  const stale = guard();
  const back = `<a class="back" href="#/admin/users">Kembali ke daftar user</a>`;
  shell("#/admin/users", `${back}<p class="muted">Memuat...</p>`);

  const r = await api("/admin/users/" + id);
  if (stale()) return;
  if (!r.ok) {
    shell("#/admin/users", `${back}<div class="empty">${esc(r.status === 404 ? "User tidak ditemukan." : errMsg(r))}</div>`);
    return;
  }

  const u = r.data.data;
  const t = u.tasks;
  const self = Number(u.id) === Number(store.user.id);
  const lock = self ? ` disabled title="Tidak bisa mengubah akun sendiri"` : "";

  shell(
    "#/admin/users",
    `${back}
    <div class="head">
      <div>
        <h1>${esc(u.name)}${self ? ` <span class="muted small">(Anda)</span>` : ""}</h1>
        <p class="muted">${esc(u.email)}</p>
      </div>
      <div class="head-badges">
        <span class="badge role-${esc(u.role)}">${esc(u.role)}</span>
        <span class="badge ${u.is_active ? "on" : "off"}">${u.is_active ? "Aktif" : "Nonaktif"}</span>
      </div>
    </div>

    <div class="card panel" style="margin:18px 0">
      <h2>Informasi akun</h2>
      <dl class="info">
        <dt>ID</dt><dd>${esc(u.id)}</dd>
        <dt>Email</dt><dd>${esc(u.email)}</dd>
        <dt>Role</dt><dd>${esc(u.role)}</dd>
        <dt>Status</dt><dd>${u.is_active ? "Aktif" : "Nonaktif"}</dd>
        <dt>Bergabung</dt><dd>${fmtDate(u.created_at)}</dd>
        <dt>Terakhir diubah</dt><dd>${fmtDate(u.updated_at)}</dd>
      </dl>
    </div>

    <h2>Task milik user ini</h2>
    <div class="stats">
      ${statCard("Total task", t.total)}
      ${statCard("To Do", t.todo)}
      ${statCard("Dikerjakan", t.in_progress)}
      ${statCard("Selesai", t.done)}
    </div>
    <div class="card panel" style="margin-bottom:18px">${t.total ? progressBar(t) : `<p class="muted">User ini belum punya task.</p>`}</div>

    <div class="toolbar">
      <button class="btn" id="tasksBtn">Lihat task user ini</button>
      <button class="btn ghost" id="roleBtn"${lock}>${u.role === "ADMIN" ? "Jadikan USER" : "Jadikan ADMIN"}</button>
      <button class="btn ${u.is_active ? "danger" : "ghost"}" id="statusBtn"${lock}>${u.is_active ? "Nonaktifkan akun" : "Aktifkan akun"}</button>
    </div>`
  );

  $("#tasksBtn").onclick = () => {
    aState.user_id = String(u.id);
    aState.page = 1;
    location.hash = "#/admin/tasks";
  };

  $("#roleBtn").onclick = async () => {
    const role = u.role === "ADMIN" ? "USER" : "ADMIN";
    const ok = await confirmBox(`Role user ini akan diubah menjadi ${role}.`, { title: "Ubah role?", confirmText: "Ya, ubah" });
    if (!ok) return;
    const res = await api(`/admin/users/${u.id}/role`, { method: "PATCH", body: { role } });
    if (!res.ok) return toast(errMsg(res), "error");
    toast("Role diubah.");
    adminUserDetail(id);
  };

  $("#statusBtn").onclick = async () => {
    const active = !u.is_active;
    const ok = await confirmBox(
      active ? "Akun ini akan diaktifkan kembali dan user bisa masuk lagi." : "User tidak akan bisa masuk lagi sampai akunnya diaktifkan.",
      { title: active ? "Aktifkan akun?" : "Nonaktifkan akun?", confirmText: active ? "Ya, aktifkan" : "Ya, nonaktifkan", danger: !active }
    );
    if (!ok) return;
    const res = await api(`/admin/users/${u.id}/status`, { method: "PATCH", body: { is_active: active } });
    if (!res.ok) return toast(errMsg(res), "error");
    toast(active ? "Akun diaktifkan." : "Akun dinonaktifkan.");
    adminUserDetail(id);
  };
}

/* ---------- mulai ---------- */
router();