# Grain → Normalize → Batch Folder → Chrome MCP → X-Pilot

## Goal

Build a restart-safe production pipeline that turns Grain transcripts into X-Pilot-ready documents and then uploads/renders them through a browser agent.

Two modes:

- `auto` (default): agent completes the full process end to end.
- `dry_run`: agent prepares the full batch and pauses for explicit human approval before upload/render.

## Non-negotiables

- The primary browser path always attaches to a **user-started Chrome session**.
- Use a **dedicated Chrome automation profile**, not the user's everyday default profile.
- The system must be able to resume from failure without duplicating work.
- Human approval is optional by default, but can be forced by mode or by policy.
- Authentication failure is treated as a resumable state, not a fatal dead end.

## Why a dedicated user-started Chrome profile

Chrome changed remote debugging behavior in 2025: from Chrome 136 onward, `--remote-debugging-port` and `--remote-debugging-pipe` are not honored for the default Chrome data directory. If remote debugging is needed, Chrome must be started with a non-default `--user-data-dir`.

That means the correct production pattern is:

1. Create a dedicated local profile, for example `~/chrome-xpilot-automation`.
2. Start Chrome manually with that profile.
3. Sign into X-Pilot in that profile once.
4. Attach the agent to that exact running session.

## Browser control fallback ladder

### Primary

**Chrome DevTools MCP attached to a user-started Chrome automation profile**

Use one of:

- `--browserUrl=http://127.0.0.1:9222`
- `--wsEndpoint=ws://127.0.0.1:9222/devtools/browser/...`

This preserves the warm logged-in session and avoids the sign-in problems that occur when automation launches its own controlled browser.

### Fallback 1

**Chrome DevTools MCP with `--autoConnect`**

Use when the user already has Chrome running and has enabled remote debugging attach permission in Chrome 144+.

### Fallback 2

**Playwright `chromium.connectOverCDP(...)`**

Attach to the same running Chrome instance if Chrome DevTools MCP is unavailable or unstable.

### Fallback 3

**Puppeteer `connect(...)`**

Also attach to the same running Chrome instance by `browserURL` or `browserWSEndpoint`.

### Fallback 4

**Local Chrome extension using `chrome.debugger`**

Emergency attach path for tab-scoped CDP control from the same user-started browser.

## Hard truth

No architecture can honestly guarantee "always able to sign in no matter what."

Sites can:

- expire cookies
- revoke sessions
- force MFA
- challenge suspicious sessions
- rotate auth flows

The practical business target is:

- no fresh-login requirement in steady state
- a fast human reauth lane that resumes the same job

## Grain adapter contract

This is the canonical interface the pipeline should depend on.

```ts
interface GrainAdapter {
  listRecordings(args?: {
    cursor?: string
    filter?: Record<string, unknown>
    include?: Record<string, unknown>
  }): Promise<{ cursor?: string; recordings: Recording[] }>

  getRecording(args: {
    recordingId: string
    include?: Record<string, unknown>
  }): Promise<Recording>

  getRecordingTranscriptJson(args: {
    recordingId: string
  }): Promise<TranscriptSegment[]>

  getRecordingTranscriptTxt(args: {
    recordingId: string
  }): Promise<string>

  getRecordingTranscriptVtt?(args: {
    recordingId: string
  }): Promise<string>

  getRecordingTranscriptSrt?(args: {
    recordingId: string
  }): Promise<string>

  downloadRecording?(args: {
    recordingId: string
  }): Promise<Blob | Buffer>

  createHook?(args: {
    hookUrl: string
    hookType:
      | "recording_added"
      | "recording_updated"
      | "recording_deleted"
      | "highlight_added"
      | "highlight_updated"
      | "highlight_deleted"
      | "story_added"
      | "story_updated"
      | "story_deleted"
      | "upload_status"
    include?: Record<string, unknown>
  }): Promise<Hook>

  listHooks?(args?: {
    filter?: Record<string, unknown>
  }): Promise<{ hooks: Hook[] }>

  deleteHook?(args: {
    hookId: string
  }): Promise<{ success: true }>
}
```

## Minimum Grain functions for v1

1. `listRecordings`
2. `getRecording`
3. `getRecordingTranscriptJson`
4. `getRecordingTranscriptTxt`
5. `createHook` or a scheduled polling fallback

## Normalize stage contract

Input:

- Grain transcript JSON
- Grain transcript TXT
- recording metadata

Output:

- one or more X-Pilot-ready Markdown files
- `manifest.json`
- `review.md`
- normalized transcript artifacts

Rules:

- remove filler and repeated false starts
- infer H1/H2/H3 structure
- preserve timecode traceability
- split long inputs into chunks under a safe ceiling below X-Pilot's 50,000-char per-video limit
- prefer 42,000 characters max per document to leave headroom

## Batch folder contract

```text
runs/
  2026-04-17/
    run_20260417_001/
      manifest.json
      queue.json
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

## Manifest shape

```json
{
  "run_id": "run_20260417_001",
  "mode": "auto",
  "status": "READY_FOR_UPLOAD",
  "require_human_approval": false,
  "browser_strategy": "browserUrl",
  "items": [
    {
      "recording_id": "grain_rec_123",
      "slug": "lesson-001",
      "title": "Prompt Engineering Basics",
      "doc_path": "docs/lesson-001.md",
      "char_count": 38120,
      "qa": {
        "score": 0.93,
        "pii": false,
        "forced_review": false
      },
      "state": "READY_FOR_UPLOAD",
      "artifacts": {
        "mp4": null,
        "scorm": null
      }
    }
  ]
}
```

## Run modes

### auto

Default. Agent continues end to end unless a forced-review condition is hit.

### dry_run

Does real work:

- ingest
- normalize
- chunk
- emit final docs
- write manifest
- verify browser attach path

Then stops before upload/render.

## Forced human review triggers

Even in `auto`, stop if any of these fire:

- QA score below threshold
- PII/redaction issue
- transcript segmentation is low confidence
- browser auth challenge appears
- X-Pilot parser output looks malformed
- upload is rejected

## Global state machine

```text
DISCOVERED
-> TRANSCRIPTS_FETCHED
-> NORMALIZED
-> BATCH_FOLDER_READY
-> AWAITING_APPROVAL        [dry_run only, or forced review]
-> READY_FOR_UPLOAD
-> BROWSER_ATTACHED
-> AUTH_VERIFIED
-> XPILOT_UPLOAD_IN_PROGRESS
-> XPILOT_PARSE_COMPLETE
-> XPILOT_RENDER_IN_PROGRESS
-> XPILOT_EXPORT_READY
-> ASSETS_DOWNLOADED
-> COMPLETE
```

## Per-item state machine

```text
NEW
-> FETCHED
-> NORMALIZED
-> DOC_READY
-> REVIEW_REQUIRED | READY_FOR_UPLOAD
-> UPLOADED
-> PARSED
-> RENDERED
-> EXPORTED
-> DOWNLOADED
-> FAILED
```

## Resume rules

- state is persisted after every step
- reruns resume from the first incomplete state
- reprocessing is hash-aware to avoid duplicate uploads or exports
- exported assets are versioned, never overwritten blindly

## X-Pilot browser worker behavior

### Upload strategy

Assume sequential uploads, even if the UI supports batch selection.

Reason:

- the browser tool surface exposes a single `upload_file` action
- existing-session attachment modes can be less predictable with multi-file chooser flows
- sequential uploads are slower but more deterministic

### Pseudocode

```python
def run_xpilot(batch):
    attach_browser()
    verify_auth()
    open_xpilot_upload_page()

    upload_target = find_upload_target()

    for item in batch.ready_items():
        upload_file(upload_target, item.doc_path)
        wait_until_queue_contains(item.slug)
        mark_state(item, "UPLOADED")

    wait_until_parser_finishes()
    mark_run("XPILOT_PARSE_COMPLETE")

    for item in batch.items_in_state("UPLOADED"):
        open_item(item)
        assert_title_present()
        assert_storyboard_present()
        if batch.mode == "auto" and not item.qa.forced_review:
            trigger_render(item)
            mark_state(item, "RENDERED")
        else:
            mark_state(item, "PARSED")

    wait_until_exports_ready()

    for item in batch.items_in_state("RENDERED"):
        download_mp4(item)
        download_scorm_if_enabled(item)
        mark_state(item, "DOWNLOADED")
```

## Auth handling

If login/MFA/challenge is detected:

1. set run status to `AUTH_REQUIRED`
2. notify human
3. human completes auth in the same open automation Chrome profile
4. agent re-checks the signed-in dashboard marker
5. run resumes from the paused step

## Recommended defaults

```yaml
mode: auto
require_human_approval: false
pause_before_upload: false
pause_before_render: false
browser_attach_strategy:
  primary: browserUrl
  fallback_1: autoConnect
  fallback_2: playwright_cdp
  fallback_3: puppeteer_connect
  fallback_4: chrome_debugger_extension
batch_size: 25
char_budget_per_doc: 42000
export_formats:
  - mp4
  - scorm
```

## Immediate next implementation step

Build the project in three workers:

1. `grain_ingest`
2. `normalize_docs`
3. `xpilot_browser_worker`

Do not start with full orchestration complexity. Start by proving these three workers independently, then add the run-level state machine.
