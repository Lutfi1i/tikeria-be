# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

from datetime import datetime

try:
    import frappe
    whitelist = frappe.whitelist
except ImportError:
    frappe = None

    def whitelist(*args, **kwargs):
        def decorator(fn):
            return fn
        return decorator


@whitelist(allow_guest=True)
def health():
    return {
        "app": "tikeria",
        "status": "ok",
        "doctypes": 17,
    }


@whitelist()
def submit_event_approval(event_name: str):
    if not frappe:
        raise RuntimeError("Frappe environment required")
    doc = frappe.get_doc("Tikeria Event", event_name)
    if doc.status not in ("Draft", "Reject"):
        frappe.throw("Event can only be submitted for approval from Draft or Reject status")
    categories = frappe.get_all("Tikeria Ticket Category", filters={"event": event_name})
    if not categories:
        frappe.throw("Event must have at least 1 ticket category before submitting for approval")
    doc.status = "Waiting Approval"
    doc.reject_reason = ""
    doc.save()
    return {"name": doc.name, "status": doc.status}


@whitelist()
def approve_event(event_name: str):
    if not frappe:
        raise RuntimeError("Frappe environment required")
    doc = frappe.get_doc("Tikeria Event", event_name)
    if doc.status != "Waiting Approval":
        frappe.throw("Only events in Waiting Approval status can be approved")
    now = datetime.now().isoformat()
    doc.status = "Published"
    doc.reviewed_by = frappe.session.user
    doc.reviewed_at = now
    doc.published_at = now
    doc.save()
    return {"name": doc.name, "status": doc.status, "published_at": doc.published_at}


@whitelist()
def reject_event(event_name: str, reject_reason: str):
    if not str(reject_reason or "").strip():
        if frappe:
            frappe.throw("Reject reason is required when rejecting an event")
        raise ValueError("Reject reason is required when rejecting an event")
    if not frappe:
        raise RuntimeError("Frappe environment required")
    doc = frappe.get_doc("Tikeria Event", event_name)
    if doc.status != "Waiting Approval":
        frappe.throw("Only events in Waiting Approval status can be rejected")
    doc.status = "Reject"
    doc.reject_reason = reject_reason
    doc.reviewed_by = frappe.session.user
    doc.reviewed_at = datetime.now().isoformat()
    doc.save()
    return {"name": doc.name, "status": doc.status, "reject_reason": doc.reject_reason}
