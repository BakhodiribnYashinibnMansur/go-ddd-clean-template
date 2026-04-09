# ADR-0001: Architecture Decision Record Template

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

As the GCT backend grows, architectural decisions are made informally in conversations
and code reviews. Without a written record, new team members cannot understand *why*
the codebase is structured a certain way, and past decisions get revisited without
the original context.

We need a lightweight, version-controlled format for recording significant
architecture choices so they remain discoverable alongside the code they govern.

## Decision

We adopt Architecture Decision Records (ADRs) stored in
`docs/architecture/adr/NNNN-slug.md`. Each ADR follows this template:

```
# ADR-NNNN: [Title]

**Status:** PROPOSED | ACCEPTED | DEPRECATED | SUPERSEDED  
**Date:** YYYY-MM-DD  

## Context
## Decision
## Consequences
### Positive
### Negative
## Alternatives Considered
```

Numbering is sequential, zero-padded to four digits. An ADR is never deleted; if a
decision is reversed, the original is marked SUPERSEDED with a link to the
replacement.

Any change to directory layout, communication patterns, infrastructure choices, or
security architecture requires a new ADR before implementation begins.

## Consequences

### Positive
- Decisions are discoverable via `git log` and file browsing.
- Onboarding is faster because rationale lives next to the code.
- Prevents re-litigating settled decisions without new information.

### Negative
- Adds a small overhead to each significant architecture change.
- Stale ADRs may confuse readers if not marked SUPERSEDED promptly.

## Alternatives Considered

- **Wiki pages** -- drift out of sync with code and lack review workflow.
- **Inline code comments** -- too scattered; hard to find the full picture.
- **No formal record** -- the status quo that motivated this ADR.
