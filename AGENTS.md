# Project AI Rules

## Mandatory change log

Every AI-assisted code change in this repository MUST update the root
`update_log.md` file before the task is considered complete.

The log entry MUST include:
- Always comment code in Vietnamese
- Date of the change.
- A short description of what was implemented, fixed, removed, or configured.
- The main files or directories changed.
- Validation performed (tests, type-check, build, lint, or a clear reason when not run).
- Any known limitation, follow-up work, or deployment note.

Append a new dated entry; do not delete or rewrite previous entries unless the
user explicitly asks for history cleanup. This requirement applies to frontend,
backend, configuration, dependency, documentation, and tooling changes.

Before the final response, verify that both the requested code change and its
corresponding `update_log.md` entry are present in the working tree.
