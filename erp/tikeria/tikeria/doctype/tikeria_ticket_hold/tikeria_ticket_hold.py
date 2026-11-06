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

class TikeriaTicketHold(Document):
    pass

def expire_active_holds():
    """Scheduled job to expire active holds that passed expired_at."""
    now = now_datetime()
    holds = frappe.get_all(
        "Tikeria Ticket Hold",
        filters={"status": "Active", "expired_at": ["<", now]},
        fields=["name", "order"]
    )
    for h in holds:
        frappe.db.set_value("Tikeria Ticket Hold", h.name, {
            "status": "Expired",
            "released_at": now
        })
        if h.order:
            frappe.db.set_value("Tikeria Order", h.order, "status", "Expired")
    if holds:
        frappe.db.commit()
