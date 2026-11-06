# Copyright (c) 2026, Tikeria Team and contributors
try:
    import frappe
    from frappe.model.document import Document
    from frappe.utils import nowdate
except ImportError:
    class Document:
        pass
    def nowdate():
        from datetime import date
        return date.today().isoformat()

class TikeriaTicket(Document):
    pass

def expire_past_event_tickets():
    """Expire issued tickets where event date is in the past."""
    today = nowdate()
    past_events = frappe.get_all(
        "Tikeria Event",
        filters={"event_date": ["<", today]},
        pluck="name"
    )
    if past_events:
        tickets = frappe.get_all(
            "Tikeria Ticket",
            filters={"event": ["in", past_events], "status": "Issued"},
            pluck="name"
        )
        for t in tickets:
            frappe.db.set_value("Tikeria Ticket", t, "status", "Expired")
        if tickets:
            frappe.db.commit()
