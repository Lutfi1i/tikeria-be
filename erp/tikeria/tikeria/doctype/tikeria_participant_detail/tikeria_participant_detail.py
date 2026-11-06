# Copyright (c) 2026, Tikeria Team and contributors
try:
    import frappe
    from frappe.model.document import Document
except ImportError:
    class Document:
        pass

def mask_nik(nik_val: str) -> str:
    if not nik_val or len(nik_val) < 8:
        return nik_val or ""
    return f"{nik_val[:4]}{'*' * (len(nik_val) - 8)}{nik_val[-4:]}"

class TikeriaParticipantDetail(Document):
    def before_save(self):
        if hasattr(self, 'nik') and self.nik:
            self.masked_nik = mask_nik(self.nik)
