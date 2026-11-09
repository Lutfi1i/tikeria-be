#!/usr/bin/env python3
"""
Tikeria Frappe ERP Runtime Server (Docker Standalone & REST API Compatible)
Loads all 17 Tikeria DocTypes and fixtures, exposes Frappe REST API (/api/resource/* & /api/method/*),
and provides the Tikeria ERP Desk Web UI at http://localhost:8000.
"""

import json
import os
import sqlite3
import urllib.parse
from datetime import datetime
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent
DOCTYPE_ROOT = BASE_DIR / "tikeria" / "tikeria" / "doctype"
FIXTURES_ROOT = BASE_DIR / "tikeria" / "fixtures"
DB_PATH = os.environ.get("ERP_DB_PATH", str(BASE_DIR / "frappe_erp.db"))
PORT = int(os.environ.get("PORT", "8000"))
DEFAULT_API_KEY = os.environ.get("FRAPPE_API_KEY", "tikeria_admin_key")
DEFAULT_API_SECRET = os.environ.get("FRAPPE_API_SECRET", "tikeria_admin_secret")


def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


def init_db_and_load_schemas():
    conn = get_db()
    cur = conn.cursor()
    cur.execute("""
        CREATE TABLE IF NOT EXISTS doctypes (
            name TEXT PRIMARY KEY,
            module TEXT,
            issingle INTEGER DEFAULT 0,
            schema_json TEXT
        )
    """)
    cur.execute("""
        CREATE TABLE IF NOT EXISTS documents (
            doctype TEXT NOT NULL,
            name TEXT NOT NULL,
            data_json TEXT NOT NULL,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL,
            PRIMARY KEY (doctype, name)
        )
    """)

    # Load 17 DocType JSON schemas
    if DOCTYPE_ROOT.exists():
        for dt_json in sorted(DOCTYPE_ROOT.glob("*/*.json")):
            try:
                raw = json.loads(dt_json.read_text(encoding="utf-8"))
                dt_name = raw.get("name")
                if dt_name:
                    cur.execute(
                        "INSERT OR REPLACE INTO doctypes (name, module, issingle, schema_json) VALUES (?, ?, ?, ?)",
                        (dt_name, raw.get("module", "Tikeria"), int(raw.get("issingle", 0)), json.dumps(raw)),
                    )
            except Exception as exc:
                print(f"[ERP] Warning loading doctype {dt_json}: {exc}")

    # Load fixtures if not yet inserted
    if FIXTURES_ROOT.exists():
        for fx_json in sorted(FIXTURES_ROOT.glob("*.json")):
            try:
                items = json.loads(fx_json.read_text(encoding="utf-8"))
                if isinstance(items, list):
                    for idx, item in enumerate(items, start=1):
                        dt = item.get("doctype")
                        if not dt:
                            continue
                        doc_name = (
                            item.get("name")
                            or item.get("package_code")
                            or item.get("event_type_code")
                            or item.get("role_name")
                            or item.get("workflow_name")
                            or (dt if dt == "Tikeria Platform Setting" else f"{dt}-{idx}")
                        )
                        item["name"] = str(doc_name)
                        now = datetime.now().isoformat()
                        cur.execute(
                            "INSERT OR IGNORE INTO documents (doctype, name, data_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                            (dt, str(doc_name), json.dumps(item), now, now),
                        )
            except Exception as exc:
                print(f"[ERP] Warning loading fixture {fx_json}: {exc}")

    conn.commit()
    conn.close()


DESK_HTML = """<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Tikeria ERP Desk</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: #F8FAFC;
      color: #0F172A;
      display: flex;
      min-height: 100vh;
    }
    aside {
      width: 280px;
      background: #0F172A;
      color: #F8FAFC;
      display: flex;
      flex-direction: column;
      border-right: 1px solid #1E293B;
      flex-shrink: 0;
    }
    .brand {
      padding: 20px 24px;
      border-bottom: 1px solid #1E293B;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .brand h1 { font-size: 16px; font-weight: 700; letter-spacing: 0.5px; }
    .badge-live {
      font-size: 11px;
      background: #10B981;
      color: #064E3B;
      padding: 2px 8px;
      border-radius: 4px;
      font-weight: 700;
    }
    .nav-section {
      padding: 16px 12px;
      overflow-y: auto;
      flex: 1;
    }
    .nav-label {
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.8px;
      color: #64748B;
      padding: 0 12px 8px 12px;
      font-weight: 600;
    }
    .dt-btn {
      width: 100%;
      text-align: left;
      background: transparent;
      border: none;
      color: #CBD5E1;
      padding: 9px 12px;
      border-radius: 6px;
      font-size: 13px;
      cursor: pointer;
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2px;
    }
    .dt-btn:hover, .dt-btn.active {
      background: #1E293B;
      color: #FFFFFF;
    }
    .dt-count {
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 11px;
      background: #334155;
      color: #E2E8F0;
      padding: 1px 6px;
      border-radius: 4px;
    }
    main {
      flex: 1;
      padding: 28px 36px;
      overflow-y: auto;
    }
    .top-bar {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 24px;
      gap: 16px;
      flex-wrap: wrap;
    }
    .title-group h2 { font-size: 22px; font-weight: 700; color: #0F172A; }
    .title-group p { font-size: 13px; color: #64748B; margin-top: 4px; }
    .cred-box {
      background: #FFFFFF;
      border: 1px solid #E2E8F0;
      border-radius: 6px;
      padding: 10px 14px;
      font-size: 12px;
      color: #334155;
    }
    .cred-box code {
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      background: #F1F5F9;
      padding: 2px 5px;
      border-radius: 4px;
      color: #0F172A;
    }
    .card {
      background: #FFFFFF;
      border: 1px solid #E2E8F0;
      border-radius: 8px;
      overflow: hidden;
      margin-bottom: 24px;
    }
    .card-header {
      padding: 14px 20px;
      border-bottom: 1px solid #E2E8F0;
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #F8FAFC;
    }
    .card-header h3 { font-size: 14px; font-weight: 600; color: #1E293B; }
    table { width: 100%; border-collapse: collapse; font-size: 13px; }
    th, td {
      text-align: left;
      padding: 12px 20px;
      border-bottom: 1px solid #F1F5F9;
      vertical-align: top;
    }
    th {
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      color: #64748B;
      background: #F8FAFC;
      font-weight: 600;
    }
    pre {
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 12px;
      background: #F8FAFC;
      padding: 8px 10px;
      border-radius: 4px;
      border: 1px solid #E2E8F0;
      max-width: 620px;
      overflow-x: auto;
      white-space: pre-wrap;
    }
    .btn {
      background: #2563EB;
      color: #FFFFFF;
      border: none;
      padding: 7px 12px;
      border-radius: 5px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
    }
    .btn:hover { background: #1D4ED8; }
    .empty { padding: 32px 20px; text-align: center; color: #64748B; font-size: 14px; }
  </style>
</head>
<body>
  <aside>
    <div class="brand">
      <h1>TIKERIA ERP</h1>
      <span class="badge-live">ONLINE</span>
    </div>
    <div class="nav-section">
      <div class="nav-label">17 DocTypes</div>
      <div id="doctype-list"></div>
    </div>
  </aside>
  <main>
    <div class="top-bar">
      <div class="title-group">
        <h2 id="current-dt-title">Tikeria Event</h2>
        <p>Frappe REST API Server &bull; Sinkron dengan Backend Go (Port 8080)</p>
      </div>
      <div class="cred-box">
        <div><strong>.env Konfigurasi Go:</strong></div>
        <div style="margin-top:4px;">
          <code>FRAPPE_SYNC_ENABLED=true</code> &nbsp;
          <code>FRAPPE_API_KEY=__API_KEY__</code> &nbsp;
          <code>FRAPPE_API_SECRET=__API_SECRET__</code>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 id="table-heading">Daftar Dokumen</h3>
        <button class="btn" onclick="reloadCurrent()">Muat Ulang Data</button>
      </div>
      <div id="table-container"></div>
    </div>
  </main>

  <script>
    let currentDoctype = "Tikeria Event";

    async function loadSummary() {
      const res = await fetch("/api/method/tikeria.api.summary");
      const json = await res.json();
      const listEl = document.getElementById("doctype-list");
      listEl.innerHTML = "";
      (json.doctypes || []).forEach(dt => {
        const btn = document.createElement("button");
        btn.className = "dt-btn" + (dt.name === currentDoctype ? " active" : "");
        btn.innerHTML = `<span>${dt.name}</span><span class="dt-count">${dt.count}</span>`;
        btn.onclick = () => {
          currentDoctype = dt.name;
          loadSummary();
          loadDocs(dt.name);
        };
        listEl.appendChild(btn);
      });
    }

    async function loadDocs(dt) {
      document.getElementById("current-dt-title").textContent = dt;
      document.getElementById("table-heading").textContent = `Dokumen: ${dt}`;
      const res = await fetch(`/api/resource/${encodeURIComponent(dt)}`);
      const json = await res.json();
      const rows = json.data || [];
      const container = document.getElementById("table-container");
      if (rows.length === 0) {
        container.innerHTML = `<div class="empty">Belum ada dokumen di <strong>${dt}</strong>. Buat transaksi/event dari API Go untuk melihat sinkronisasi otomatis.</div>`;
        return;
      }
      let html = `<table><thead><tr><th>ID Dokumen (name)</th><th>Status / Info Utama</th><th>Payload JSON</th><th>Diperbarui</th></tr></thead><tbody>`;
      rows.forEach(r => {
        const status = r.status || r.order_status || r.package_name || r.event_type_name || (r.default_service_fee !== undefined ? `Fee: Rp ${r.default_service_fee}` : "-");
        const updated = r._updated_at ? r._updated_at.replace("T", " ").slice(0, 19) : "-";
        const copy = Object.assign({}, r);
        delete copy._updated_at;
        html += `<tr>
          <td><strong><code>${r.name}</code></strong></td>
          <td>${status}</td>
          <td><pre>${JSON.stringify(copy, null, 2)}</pre></td>
          <td>${updated}</td>
        </tr>`;
      });
      html += `</tbody></table>`;
      container.innerHTML = html;
    }

    function reloadCurrent() {
      loadSummary();
      loadDocs(currentDoctype);
    }

    loadSummary();
    loadDocs(currentDoctype);
  </script>
</body>
</html>
"""


class FrappeRequestHandler(BaseHTTPRequestHandler):
    def _send_json(self, status_code: int, payload: dict):
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status_code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _read_json_body(self) -> dict:
        length = int(self.headers.get("Content-Length", "0"))
        if length <= 0:
            return {}
        raw = self.rfile.read(length).decode("utf-8")
        try:
            return json.loads(raw)
        except Exception:
            return {}

    def do_OPTIONS(self):
        self._send_json(200, {"ok": True})

    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        path = urllib.parse.unquote(parsed.path)

        if path in ("/", "/app", "/desk"):
            html = DESK_HTML.replace("__API_KEY__", DEFAULT_API_KEY).replace("__API_SECRET__", DEFAULT_API_SECRET)
            body = html.encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        if path == "/api/method/tikeria.api.health":
            conn = get_db()
            cnt = conn.execute("SELECT count(*) FROM doctypes").fetchone()[0]
            conn.close()
            self._send_json(200, {"message": {"app": "tikeria", "status": "ok", "doctypes": cnt}})
            return

        if path == "/api/method/tikeria.api.summary":
            conn = get_db()
            dts = conn.execute("SELECT name, module, issingle FROM doctypes ORDER BY name").fetchall()
            result = []
            for row in dts:
                c = conn.execute("SELECT count(*) FROM documents WHERE doctype = ?", (row["name"],)).fetchone()[0]
                result.append({"name": row["name"], "module": row["module"], "count": c})
            conn.close()
            self._send_json(200, {"doctypes": result})
            return

        if path.startswith("/api/resource/"):
            rest = path[len("/api/resource/"):].strip("/")
            parts = rest.split("/", 1)
            doctype = parts[0]

            conn = get_db()
            if len(parts) == 1:
                rows = conn.execute(
                    "SELECT name, data_json, updated_at FROM documents WHERE doctype = ? ORDER BY updated_at DESC",
                    (doctype,),
                ).fetchall()
                conn.close()
                docs = []
                for r in rows:
                    d = json.loads(r["data_json"])
                    d["name"] = r["name"]
                    d["_updated_at"] = r["updated_at"]
                    docs.append(d)
                self._send_json(200, {"data": docs})
                return
            else:
                doc_name = parts[1]
                row = conn.execute(
                    "SELECT data_json FROM documents WHERE doctype = ? AND name = ?",
                    (doctype, doc_name),
                ).fetchone()
                conn.close()
                if not row:
                    self._send_json(404, {"message": f"{doctype} {doc_name} not found"})
                    return
                self._send_json(200, {"data": json.loads(row["data_json"])})
                return

        self._send_json(404, {"message": "Not Found"})

    def do_POST(self):
        parsed = urllib.parse.urlparse(self.path)
        path = urllib.parse.unquote(parsed.path)
        body = self._read_json_body()

        if path.startswith("/api/resource/"):
            doctype = path[len("/api/resource/"):].strip("/").split("/")[0]
            conn = get_db()
            cnt = conn.execute("SELECT count(*) FROM documents WHERE doctype = ?", (doctype,)).fetchone()[0]
            doc_name = (
                body.get("name")
                or body.get("hold_code")
                or body.get("order_code")
                or body.get("invoice_number")
                or body.get("payment_code")
                or f"{doctype[:3].upper()}-{datetime.now().year}-{cnt + 1:05d}"
            )
            body["doctype"] = doctype
            body["name"] = str(doc_name)
            now = datetime.now().isoformat()
            conn.execute(
                "INSERT OR REPLACE INTO documents (doctype, name, data_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                (doctype, str(doc_name), json.dumps(body), now, now),
            )
            conn.commit()
            conn.close()
            self._send_json(200, {"data": body})
            return

        if path.startswith("/api/method/tikeria.api."):
            method_name = path.split("/api/method/tikeria.api.", 1)[1]
            event_name = body.get("event_name") or body.get("name")
            conn = get_db()
            row = conn.execute(
                "SELECT data_json FROM documents WHERE doctype = 'Tikeria Event' AND name = ?",
                (str(event_name),),
            ).fetchone()
            if not row:
                conn.close()
                self._send_json(404, {"message": f"Event {event_name} not found"})
                return
            doc = json.loads(row["data_json"])
            now = datetime.now().isoformat()
            if method_name == "submit_event_approval":
                doc["status"] = "Waiting Approval"
            elif method_name == "approve_event":
                doc["status"] = "Published"
                doc["published_at"] = now
            elif method_name == "reject_event":
                doc["status"] = "Reject"
                doc["reject_reason"] = body.get("reject_reason", "")
            conn.execute(
                "UPDATE documents SET data_json = ?, updated_at = ? WHERE doctype = 'Tikeria Event' AND name = ?",
                (json.dumps(doc), now, str(event_name)),
            )
            conn.commit()
            conn.close()
            self._send_json(200, {"message": doc})
            return

        self._send_json(404, {"message": "Unknown endpoint"})

    def do_PUT(self):
        parsed = urllib.parse.urlparse(self.path)
        path = urllib.parse.unquote(parsed.path)
        body = self._read_json_body()

        if path.startswith("/api/resource/"):
            rest = path[len("/api/resource/"):].strip("/")
            parts = rest.split("/", 1)
            if len(parts) < 2:
                self._send_json(400, {"message": "Document name required for PUT"})
                return
            doctype, doc_name = parts[0], parts[1]
            conn = get_db()
            row = conn.execute(
                "SELECT data_json, created_at FROM documents WHERE doctype = ? AND name = ?",
                (doctype, doc_name),
            ).fetchone()
            now = datetime.now().isoformat()
            if row:
                existing = json.loads(row["data_json"])
                existing.update(body)
                existing["name"] = doc_name
                conn.execute(
                    "UPDATE documents SET data_json = ?, updated_at = ? WHERE doctype = ? AND name = ?",
                    (json.dumps(existing), now, doctype, doc_name),
                )
                result = existing
            else:
                body["doctype"] = doctype
                body["name"] = doc_name
                conn.execute(
                    "INSERT INTO documents (doctype, name, data_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                    (doctype, doc_name, json.dumps(body), now, now),
                )
                result = body
            conn.commit()
            conn.close()
            self._send_json(200, {"data": result})
            return

        self._send_json(404, {"message": "Not Found"})


if __name__ == "__main__":
    init_db_and_load_schemas()
    print(f"Tikeria Frappe ERP Server running at http://0.0.0.0:{PORT}")
    server = HTTPServer(("0.0.0.0", PORT), FrappeRequestHandler)
    server.serve_forever()
