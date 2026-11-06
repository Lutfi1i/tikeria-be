app_name = "tikeria"
app_title = "Tikeria"
app_publisher = "Tikeria Team"
app_description = "ERP and Admin module for Tikeria ticketing platform"
app_email = "dev@tikeria.id"
app_license = "mit"

fixtures = [
    {"dt": "Role", "filters": [["name", "in", ["Tikeria Admin", "Tikeria Organizer", "Tikeria Buyer", "Tikeria Scanner"]]]},
    {"dt": "Tikeria Organizer Package"},
    {"dt": "Tikeria Event Type"},
    {"dt": "Tikeria Platform Setting"},
    {"dt": "Workflow", "filters": [["name", "in", ["Tikeria Event Approval"]]]},
    {"dt": "Workflow State"},
    {"dt": "Workflow Action Master"}
]

scheduler_events = {
    "cron": {
        # Expire temporary holds every 1 minute
        "*/1 * * * *": [
            "tikeria.tikeria.doctype.tikeria_ticket_hold.tikeria_ticket_hold.expire_active_holds"
        ],
        # Daily job: expire past tickets and cleanup webview tokens
        "0 0 * * *": [
            "tikeria.tikeria.doctype.tikeria_ticket.tikeria_ticket.expire_past_event_tickets",
            "tikeria.tikeria.doctype.tikeria_webview_token.tikeria_webview_token.cleanup_expired_tokens"
        ]
    }
}
