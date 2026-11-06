# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

import re
from datetime import date, datetime

try:
    import frappe
    from frappe.model.document import Document
except ImportError:
    frappe = None

    class Document:
        pass


def _throw(msg: str):
    if frappe and hasattr(frappe, "throw"):
        frappe.throw(msg)
    raise ValueError(msg)


class TikeriaEvent(Document):
    def validate(self):
        # 1. Auto-generate slug from title if empty
        if getattr(self, "title", None) and not getattr(self, "slug", None):
            slug = re.sub(r"[^a-z0-9]+", "-", str(self.title).lower()).strip("-")
            self.slug = slug

        # 2. Default event_template to 'pensi' if empty
        if not getattr(self, "event_template", None):
            self.event_template = "pensi"

        # 3. Validate event_date is in the future when creating or editing Draft/Reject
        event_date_val = getattr(self, "event_date", None)
        if event_date_val:
            if isinstance(event_date_val, str):
                parsed_date = datetime.strptime(event_date_val[:10], "%Y-%m-%d").date()
            elif isinstance(event_date_val, datetime):
                parsed_date = event_date_val.date()
            else:
                parsed_date = event_date_val
            if parsed_date <= date.today() and getattr(self, "status", "Draft") in ("Draft", "Reject"):
                _throw("Event date must be greater than today")

        # 4. Reject status requires reject_reason
        if getattr(self, "status", None) == "Reject":
            if not str(getattr(self, "reject_reason", "") or "").strip():
                _throw("Reject reason is required when rejecting an event")
