# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

try:
    import frappe
    from frappe.model.document import Document
except ImportError:
    class Document:
        pass

class TikeriaBuyerProfile(Document):
    pass
