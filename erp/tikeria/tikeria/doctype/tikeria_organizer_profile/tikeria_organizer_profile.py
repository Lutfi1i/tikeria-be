# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

from datetime import date

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


class TikeriaOrganizerProfile(Document):
    def validate(self):
        if not getattr(self, "package", None):
            self.package = "basic"
        if not getattr(self, "package_since", None):
            self.package_since = date.today().isoformat()
        if not getattr(self, "organizer_type", None):
            self.organizer_type = "school_campus"

        if getattr(self, "organizer_type", None) == "public_later":
            _throw("organizer_type 'public_later' is reserved for post-MVP; use 'school_campus'")

        if getattr(self, "organizer_type", None) == "school_campus":
            if not str(getattr(self, "institution", "") or "").strip():
                _throw("Institution is required for school_campus organizer")
            if not str(getattr(self, "city", "") or "").strip():
                _throw("City is required for school_campus organizer")
