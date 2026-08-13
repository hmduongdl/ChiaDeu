# Update log

## 2026-08-12

- Added Figma-based authentication routes:
  - `/login` — Emerald Minimalist login screen (node `25:795`).
  - `/register` — account creation screen (node `25:848`).
  - `/forgot-password` — password recovery screen (node `25:917`).
- Added shared auth field and Figma asset helpers under `frontend/src/components/auth/`.
- Added password visibility toggle on login and registration forms.
- Added navigation links between login, registration, and password recovery screens.
- Bottom Nav is hidden on all three auth routes via `usePathname`, so it does not overlap the Figma auth layouts.
- Forms currently prevent submission while backend authentication endpoints are not implemented.
- Figma SVG assets are referenced through the connector's temporary URLs for this implementation preview; they should be downloaded into `frontend/public` before a production release.

## 2026-08-12 — AI workflow rule

- Added root `AGENTS.md` with a mandatory rule requiring every AI-assisted code, configuration, dependency, documentation, or tooling change to append an entry to `update_log.md`.
- The rule requires the date, summary, changed paths, validation, and known limitations/follow-up notes.
- Validation: verified the rule file and this log entry are both present.
