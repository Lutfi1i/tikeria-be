# Copyright (c) 2026, Tikeria Team and contributors
# For license information, please see license.txt

try:
    import frappe
    from frappe.model.document import Document
except ImportError:
    frappe = None

    class Document:
        pass


class TikeriaOrder(Document):
    def validate(self):
        subtotal = float(getattr(self, "subtotal", 0) or 0)
        service_fee = float(getattr(self, "service_fee", 0) or 0)
        if subtotal == 0:
            service_fee = 0.0
            self.service_fee = 0.0
        self.total_amount = subtotal + service_fee
