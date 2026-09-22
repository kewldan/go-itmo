# Changelog

## Unreleased

- `itmoid`: ITMO.ID login with password, one-time codes and SSO; PKCE;
  refreshing token source that reports rotated tokens; logout, userinfo,
  ID-token claims, refresh-token expiry, strict callback parsing, token
  exchange (impersonation).
- `myitmo`: client for every route of the my.itmo.ru cabinet — 30
  services, 809 methods, student and staff sections — plus the qr.itmo.su pass.
- `bars`: client for the full BARS REST API (91 methods) with silent session
  renewal through ITMO.ID SSO, period selection, journals, marks, approvals,
  checkpoint plans, deadlines and reports.
