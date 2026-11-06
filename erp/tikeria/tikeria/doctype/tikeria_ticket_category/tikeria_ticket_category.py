# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

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


class TikeriaTicketCategory(Document):
    def validate(self):
        price = float(getattr(self, "price", 0) or 0)
        if price < 0:
            _throw("Price must be greater than or equal to 0")

        quota = int(getattr(self, "quota", 0) or 0)
        if quota <= 0:
            _throw("Quota must be greater than 0")

        max_tx = int(getattr(self, "max_ticket_per_transaction", 0) or 0)
        if max_tx <= 0:
            self.max_ticket_per_transaction = 5
        elif max_tx > 5:
            _throw("Max ticket per transaction cannot exceed 5 in MVP")
