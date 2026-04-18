# IDEA - Grain MCP -> X-Pilot Production Plan Review

## Scope

This document compares **three implementation plans** for turning **Grain transcripts** into **X-Pilot doc-to-video outputs** using an **agent that attaches to a user-started Chrome session**.

The goal is not just to list options. The goal is to:

1. propose 3 realistic plans,
2. test them against the current official documentation,
3. use role-based reviewer agents to challenge the plans,
4. reach consensus on the **most accurate** and **most business-safe** plan,
5. agree on the **ordered steps** of the final plan.

## Non-negotiables

- The default mode must allow the agent to complete the full process end to end.
- A `dry_run` mode with a human approval gate must exist, but it is optional.
- The agent must attach to a **user-started Chrome session** for X-Pilot sign-in and upload flows.
- The plan must not rely on a public X-Pilot API unless the account actually has confirmed access.
- The plan must work with **Grain transcripts first**, because that is the source material.

## Evidence-locked facts

These are the facts the plans must respect.

### Grain

- Grain v2 exposes **List Recordings**, **Get Recording**, **Get Recording Transcript (json)**, and **Get Recording Transcript** in `.txt`, `.vtt`, and `.srt` formats.
- Transcript JSON returns `speaker`, `start`, `end`, and `text`, which is the right shape for sectioning, source traceability, and clip mapping.
- Grain also exposes hooks and documents `recording_added`, `recording_updated`, and `upload_status` among the hook types.

### Chrome / browser control

- Chrome DevTools MCP supports browser control tools including `click`, `fill`, `upload_file`, `new_page`, `select_page`, `wait_for`, `evaluate_script`, `take_snapshot`, and `take_screenshot`.
- Chrome DevTools MCP can attach to an already running browser using `--browserUrl` or `--wsEndpoint`.
- Chrome DevTools MCP also supports `--autoConnect`, but only for supported Chrome versions and with user-enabled remote debugging.
- Chrome 136 changed remote debugging behavior: debugging switches are **not respected for the default Chrome data directory**. A non-default `--user-data-dir` is required.
- Playwright supports `connectOverCDP(...)` for attaching to an existing Chromium browser, but Playwright warns that CDP attachment is lower fidelity than native Playwright protocol.
- Puppeteer supports reconnecting to an existing browser through `Puppeteer.connect(...)` and a WebSocket endpoint.
- Chrome extensions can use `chrome.debugger` as an alternate transport for the Chrome DevTools Protocol, but that API has restricted domains.

### X-Pilot

- X-Pilot accepts `.pdf`, `.ppt/.pptx`, `.doc/.docx`, `.md/.markdown`, and `.txt` inputs.
- X-Pilot says headings improve results and H1/H2/H3 structure becomes video chapters automatically.
- X-Pilot documents a **50,000 character maximum per video** on the text-to-video workflow.
- X-Pilot says Pro and above can batch process up to **50 documents simultaneously**.
- X-Pilot exports MP4 and SCORM, and its FAQ also references xAPI on Enterprise.
- Public X-Pilot materials are **inconsistent** about API access. One FAQ pricing table says Pro includes API, while the technical FAQ says API and custom integrations are Enterprise-only. Because of that inconsistency, the safest implementation assumption is: **do not depend on X-Pilot API access for the production MVP**.

### Browserbase / cloud recovery

- Browserbase contexts persist cookies and auth state across sessions.
- Browserbase `keepAlive` lets a session survive disconnects and reconnect later.
- Browserbase also explicitly notes that websites can still expire sessions, revoke tokens, or force logouts.

## What I could not verify from this host

I could **not** introspect the exact live Grain MCP method names from this ChatGPT environment.

Because of that, this plan uses a **canonical Grain adapter interface**. During implementation, the first task is to map the real Grain MCP method names in your host to that interface.

That is a setup task, not a product risk.

---

# Version 1 - Lean Local Pipeline

## Thesis

Build the smallest possible production system:

`Grain MCP -> normalize -> batch folder -> Chrome DevTools MCP -> X-Pilot`

Use only one browser driver, one execution path, and one recovery mode.

## Why this plan exists

This is the fastest way to get from idea to usable output. It minimizes engineering overhead and validates business value quickly.

## Architecture

```text
Grain MCP
  -> transcript fetch
  -> normalize + chunk
  -> docs/*.md + manifest.json
  -> attach to user-started Chrome via Chrome DevTools MCP
  -> upload to X-Pilot
  -> parse/review
  -> export MP4/SCORM
  -> save outputs
```

## Required Grain capability surface

```ts
list_recordings()
get_recording(recording_id)
get_recording_transcript_json(recording_id)
get_recording_transcript_txt(recording_id)
create_hook?(hook_url, hook_type)
list_hooks?()
delete_hook?(hook_id)
```

## Required Chrome DevTools MCP surface

```ts
new_page()
navigate_page()
select_page()
list_pages()
wait_for()
click()
fill()
press_key()
upload_file()
evaluate_script()
take_snapshot()
take_screenshot()
```

## Run modes

### `auto`

Default.

The agent:

1. fetches transcripts,
2. normalizes them,
3. writes docs,
4. attaches to Chrome,
5. uploads docs,
6. waits for parse,
7. exports outputs,
8. downloads artifacts,
9. marks the batch complete.

### `dry_run`

Optional.

The agent:

1. fetches transcripts,
2. normalizes them,
3. writes docs and a review packet,
4. stops before upload,
5. waits for human approval.

## Strengths

- Fastest to build.
- Most direct path to value.
- Uses only current, documented public capabilities.
- Avoids relying on uncertain X-Pilot API access.

## Weaknesses

- Single browser driver dependency.
- If Chrome DevTools MCP attach fails, the whole run stalls.
- Weakest resilience against auth disruptions.
- Least suitable for business-critical uptime.

## Accuracy verdict

**Accurate but incomplete.**

This plan matches the public documentation, but it underestimates the need for attach fallbacks and auth recovery if the workflow becomes business-critical.

---

# Version 2 - Resilient Attach Abstraction

## Thesis

Keep the local-user-started-Chrome requirement, but introduce an explicit **Attach Manager** so the agent can keep working if one attach client fails.

`Grain MCP -> normalize -> batch folder -> Attach Manager -> X-Pilot`

## Why this plan exists

This is the best balance of speed, accuracy, and resilience.

It still treats the **user-started dedicated Chrome session** as the source of truth for authentication, but it avoids depending on a single browser automation client.

## Architecture

```text
Grain MCP
  -> transcript fetch
  -> normalize + chunk
  -> docs/*.md + manifest.json
  -> Attach Manager
       1. Chrome DevTools MCP via --browserUrl / --wsEndpoint
       2. Chrome DevTools MCP via --autoConnect
       3. Playwright connectOverCDP(...)
       4. Puppeteer.connect(...)
       5. Optional companion extension via chrome.debugger
  -> X-Pilot upload / parse / export
  -> outputs + state store
```

## Dedicated Chrome rule

The user should start a **dedicated local Chrome automation profile** for X-Pilot, not their everyday browsing profile.

That profile should:

- use a **non-default user-data-dir**,
- keep a stable signed-in X-Pilot session,
- expose a remote debugging endpoint,
- be the only Chrome instance the agent is allowed to control.

This is the most accurate interpretation of the current Chrome security model and the safest way to preserve sign-in continuity.

## Key design idea

Treat browser control as a **pluggable transport layer**, not as the product itself.

If DevTools MCP is down, the agent should not lose the run. It should reattach through another client to the **same already running Chrome session**.

## State machine

### Per run

```text
DISCOVERED
-> FETCHED
-> NORMALIZED
-> BATCH_READY
-> AWAITING_APPROVAL [only if dry_run or forced review]
-> ATTACHING
-> ATTACHED
-> AUTHENTICATED
-> UPLOADING
-> PARSED
-> EXPORTING
-> DOWNLOADING
-> COMPLETE
```

### Per item

```text
NEW
-> TRANSCRIPT_READY
-> DOC_READY
-> READY_FOR_UPLOAD
-> UPLOADED
-> PARSED
-> EXPORTED
-> DOWNLOADED
-> FAILED
```

## Batch folder contract

```text
runs/
  2026-04-17/
    run_001/
      manifest.json
      review.md
      docs/
        lesson-001.md
        lesson-002.md
      source/
        lesson-001.transcript.json
        lesson-001.transcript.txt
      exports/
      logs/
        orchestrator.jsonl
        browser.jsonl
```

## `manifest.json` keys

```json
{
  "run_id": "run_001",
  "mode": "auto",
  "status": "BATCH_READY",
  "browser_profile": "xpilot-automation",
  "attach_path": null,
  "items": []
}
```

## Normalization rules

For each Grain transcript:

1. ingest transcript JSON as the system of record,
2. use TXT only as audit-friendly fallback,
3. group transcript blocks into topic sections,
4. add explicit H1/H2/H3 structure,
5. keep each output document under **42,000 characters** as a safety margin below X-Pilot's 50,000-character limit,
6. preserve source traceability through timecode comments.

### Markdown template

```markdown
# Lesson Title

## Objective
One clear learner outcome.

## Context
Short framing section.

## Core Idea 1
Explanation.

## Example
Concrete example.

## Common Mistake
Likely learner error.

## Core Idea 2
Explanation.

## Recap
- Bullet 1
- Bullet 2
- Bullet 3

<!-- grain_recording_id: ... -->
<!-- start_ms: ... -->
<!-- end_ms: ... -->
```

## Attach ladder

The agent should attempt attach in this order:

1. **Chrome DevTools MCP with `--browserUrl` or `--wsEndpoint`**
2. **Chrome DevTools MCP with `--autoConnect`**
3. **Playwright `connectOverCDP(...)`**
4. **Puppeteer `connect(...)`**
5. **Optional companion extension using `chrome.debugger`**

If all five fail, the run enters `AUTH_OR_ATTACH_RECOVERY_REQUIRED`.

## Auth rule

The system should never assume the session is valid just because the browser attached.

It must check an **auth sentinel** on X-Pilot, for example:

- dashboard shell exists,
- upload CTA exists,
- no login form is visible,
- no redirect to auth routes,
- current user marker is present.

If any check fails:

- mark run as `AUTH_REQUIRED`,
- pause,
- request user takeover,
- resume the same run after sign-in.

## Strengths

- Strongest balance of accuracy and implementation realism.
- Keeps the user-started Chrome requirement intact.
- Avoids overcommitting to unverified X-Pilot API access.
- Allows recovery from attach client failures without discarding the run.
- Business-safe enough to use as the production baseline.

## Weaknesses

- More engineering work than Version 1.
- Requires clear attach instrumentation and logs.
- Needs a dedicated Chrome profile policy.

## Accuracy verdict

**Most accurate and most defensible.**

This plan aligns with the documentation and avoids making unverified promises.

---

# Version 3 - Dual-Lane Production + Disaster Recovery

## Thesis

Build Version 2, but also add a separate **cloud recovery lane** so the process can continue even if the local machine disconnects.

## Why this plan exists

This plan maximizes continuity for a high-value business workflow.

## Architecture

```text
Primary lane:
  Grain MCP
    -> normalize
    -> batch folder
    -> user-started local Chrome
    -> X-Pilot

Recovery lane:
  Browserbase context + keepAlive
    -> reconnectable cloud session
    -> same X-Pilot workflow
```

## Recovery logic

If the local run fails because of:

- laptop sleep,
- local network loss,
- local browser crash,
- MCP transport failure,

then the orchestrator may continue from a recovery lane.

## Strengths

- Best continuity.
- Good for overnight jobs.
- Good for reconnect workflows.

## Weaknesses

- Violates the spirit of "always attach to the user-started Chrome session" once the recovery lane takes over.
- Adds cost, infra, and another security surface.
- Browserbase itself warns that sites can still expire cookies, revoke tokens, or force logout.
- Too much complexity for the first production version.

## Accuracy verdict

**Accurate as an optional add-on, not as the default plan.**

This plan is useful, but only after the local attach plan is working.

---

# Reviewer agents

These are **role-based reviewer agents**, not external tools. Each reviewer evaluated the three plans against the documented constraints and business needs.

## Agent A - Browser Automation Engineer

### Priorities

- attach reliability
- auth continuity
- browser-client fallback depth
- resumability

### Assessment

- **Version 1** is too thin for revenue-critical execution.
- **Version 2** is the strongest because it keeps a single auth truth source while allowing multiple attach clients.
- **Version 3** is valuable later, but it complicates the sign-in story and adds infrastructure too early.

### Ranking

1. Version 2
2. Version 1
3. Version 3

## Agent B - Reliability and Security Architect

### Priorities

- Chrome 136+ remote debugging safety
- separation from the default browsing profile
- recoverability
- avoiding undocumented vendor dependencies

### Assessment

- **Version 1** is accurate, but fragile.
- **Version 2** is the best because it uses a dedicated non-default profile and does not depend on unverified X-Pilot API access.
- **Version 3** has legitimate recovery value, but it should not be the default because it changes the trust boundary.

### Ranking

1. Version 2
2. Version 3
3. Version 1

## Agent C - Content Ops and Business Owner

### Priorities

- fastest path to useful output
- lowest operational surprise
- ease of support
- clear human checkpoint when auth breaks

### Assessment

- **Version 1** is tempting because it is simple.
- **Version 2** is the best business choice because it stays simple enough while solving the failure modes that would hurt a real business.
- **Version 3** is too much for the first commit.

### Ranking

1. Version 2
2. Version 1
3. Version 3

---

# Consensus

## Consensus result

All three reviewer agents agree that the **best plan is Version 2**, with:

- the **simplicity of Version 1** for the MVP milestone,
- the **attach fallback ladder of Version 2** as the production baseline,
- and the **cloud recovery ideas from Version 3** kept as a later enhancement, not day-one scope.

## Consensus statement on accuracy

The most accurate plan is the one that:

1. assumes **Grain transcript ingest is stable**,
2. assumes **X-Pilot browser automation is the primary path**,
3. assumes **X-Pilot API access is not reliable enough to plan against**,
4. uses a **dedicated user-started Chrome profile with non-default user-data-dir**,
5. treats browser attachment as **replaceable transport**, not as the whole system,
6. accepts that **human re-authentication remains a real recovery case**.

That is Version 2.

## Agreed steps of the plan

These are the steps all reviewer agents agreed on.

### Phase 0 - Setup

1. **Create a dedicated X-Pilot automation Chrome profile**.
2. Launch it with a **non-default user-data-dir** and remote debugging enabled.
3. Sign in to X-Pilot manually once in that profile.
4. Store no passwords or 2FA codes in chat or logs.

### Phase 1 - Capability mapping

5. Map the real Grain MCP tool names to the canonical Grain adapter surface.
6. Verify the required transcript methods exist.
7. Verify at least one hook path exists for event-driven triggering, or fall back to scheduled polling.

### Phase 2 - Ingest and normalize

8. List target recordings from Grain.
9. Fetch transcript JSON and TXT.
10. Normalize transcripts into structured Markdown with H1/H2/H3 headings.
11. Chunk long documents below the X-Pilot character budget.
12. Write `docs/`, `source/`, `manifest.json`, and `review.md` into a batch folder.

### Phase 3 - Execution modes

13. If `mode = dry_run`, stop here and require approval.
14. If `mode = auto`, continue automatically unless a forced-review rule triggers.

### Phase 4 - Attach and auth

15. Attach to the running Chrome via the Attach Manager.
16. Run the X-Pilot auth sentinel.
17. If auth is broken, request user takeover and resume the same run.

### Phase 5 - Upload and export

18. Upload documents to X-Pilot.
19. Wait for parse completion.
20. Validate title, structure, and scene generation.
21. Export MP4 and SCORM where required.
22. Download artifacts into `exports/`.

### Phase 6 - Closeout

23. Mark item states complete.
24. Write a run report with successes, failures, and retry reasons.
25. Keep enough logs to resume without starting over.

## Forced-review rules

Even in `auto`, the run must stop for review if:

- transcript quality is poor,
- PII or regulated content is detected,
- topic boundaries are low confidence,
- the attach ladder exhausts all local options,
- X-Pilot parse results are malformed,
- auth requires reentry,
- export fails repeatedly.

## Final recommendation

Build **Version 2** first.

That means:

- local dedicated Chrome profile,
- explicit attach manager,
- Grain transcript normalization,
- batch-folder execution,
- X-Pilot browser automation,
- optional `dry_run`,
- human reauth lane,
- no API dependency for MVP.

Do **not** start with Version 3.
Do **not** ship Version 1 as the long-term architecture.

---

# Implementation notes

## Canonical files

```text
runs/
  <date>/
    <run_id>/
      manifest.json
      review.md
      docs/
      source/
      exports/
      logs/
```

## Recommended run config

```yaml
mode: auto
require_human_approval: false
forced_review_on_auth_failure: true
browser_profile: xpilot-automation
attach_order:
  - chrome_devtools_mcp_browser_url
  - chrome_devtools_mcp_auto_connect
  - playwright_cdp
  - puppeteer_connect
  - chrome_debugger_extension
char_budget_per_doc: 42000
batch_size: 25
```

## Why batch size 25, not 50

X-Pilot says up to 50 docs simultaneously, but 25 is the safer operational default for the first production version. It gives better failure isolation and easier reruns.

---

# Reference links

Official sources used for the review:

- Grain API: https://developers.grain.com/
- Chrome DevTools MCP blog: https://developer.chrome.com/blog/chrome-devtools-mcp
- Chrome DevTools MCP existing-session blog: https://developer.chrome.com/blog/chrome-devtools-mcp-debug-your-browser-session
- Chrome remote debugging change: https://developer.chrome.com/blog/remote-debugging-port
- Chrome DevTools MCP GitHub: https://github.com/mcp/ChromeDevTools/chrome-devtools-mcp
- Playwright BrowserType docs: https://playwright.dev/docs/api/class-browsertype
- Puppeteer connect docs: https://pptr.dev/api/puppeteer.puppeteer.connect
- Puppeteer wsEndpoint docs: https://pptr.dev/api/puppeteer.browser.wsendpoint
- Chrome debugger API: https://developer.chrome.com/docs/extensions/reference/api/debugger
- Browserbase Contexts: https://docs.browserbase.com/features/contexts
- Browserbase Keep Alive: https://docs.browserbase.com/features/keep-alive
- X-Pilot FAQ: https://www.x-pilot.ai/resources/faq
- X-Pilot Text-to-Video: https://www.x-pilot.ai/products/text-to-video
