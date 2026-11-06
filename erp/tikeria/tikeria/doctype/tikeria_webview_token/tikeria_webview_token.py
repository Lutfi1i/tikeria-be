# Copyright (c) 2026, Tikeria Team and contributors
try:
    import frappe
    from frappe.model.document import Document
    from frappe.utils import now_datetime
except ImportError:
    class Document:
        pass
    def now_datetime():
        from datetime import datetime
        return datetime.now()

class TikeriaWebViewToken(Document):
    pass

def cleanup_expired_tokens():
    """Deletes webview tokens older than 1 hour."""
    now = now_datetime()
    frappe.db.delete("Tikeria WebView Token", {
        "expires_at": ["<", now]
    })
    frappe.db.commit()
