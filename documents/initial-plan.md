# Phase 1 Build Plan

**Project:** Offline desktop transcription app
**Stack:** Go + Wails v2 + Svelte, bundled whisper.cpp and ffmpeg
**Goal of Phase 1:** A complete, shippable app that a non-technical person can install and use to transcribe a file end to end.

---

## Scope boundary

Phase 1 is done when someone can install the app, drop in a Zoom recording, get an accurate editable transcript, fix a few words, and export an SRT. Nothing more.

Explicitly **not** in Phase 1: speaker diarization, the searchable library, custom vocabulary, VAD, live microphone capture, summarization, cloud anything.

---

## Milestone 0: Skeleton and distribution

Solve distribution before writing features. Leaving it to the end is how projects stall at 90 percent complete.

### Tasks

- Wails v2 project with Svelte frontend, one Go method bound and callable from JS
- CI matrix building all four targets: Windows x64, macOS ARM64, macOS x64, Linux x64
- macOS: codesign and notarize the bundle
- Windows: installer (NSIS or WiX), test on a clean VM for SmartScreen warnings
- Linux: AppImage, plus document the `libwebkit2gtk` dependency

### Binary bundling strategy

Recommended approach:

1. `go:embed` the platform's `whisper-cli` and `ffmpeg` binaries into the Go binary
2. On first run, extract to the user cache directory
3. `chmod +x` on Unix targets
4. Verify by SHA256 on every launch, re-extract if the hash mismatches so a corrupted extract self-heals

This keeps distribution to a single file, which is the main reason for choosing Go over Python here.

**Tradeoff:** installer size grows to roughly 60 to 80MB with ffmpeg included. Acceptable. The alternative, downloading binaries on first run, adds a failure mode at the worst possible moment in the user's experience.

### Done when

A non-technical person can install on all four platforms with no security warnings and see a window.

---

## Milestone 1: Media intake

Package: `internal/media`

### Operations

**Probe.** Run `ffprobe`, return duration, stream list, codecs, sample rate. Called before anything else.

**Convert.** Transcode to 16kHz mono WAV in a temp directory keyed by job ID. This is what whisper expects, and users will hand you everything except this.

### Error cases to handle explicitly

- No audio stream in the file
- Corrupt or truncated container
- DRM protected media
- Multiple audio tracks, need to select one, default to the first
- Insufficient disk space for the converted WAV (a 3 hour file is roughly 350MB as 16kHz mono WAV)

Every one of these should produce a readable message in the UI, not a crash three steps later.

### Cleanup

Delete temp WAVs on job completion. Also sweep the temp directory on app start, since a crash leaves orphans behind.

### Done when

You can hand it an MP4, an M4A, an MP3, and a corrupt file, and get correct behaviour from all four.

---

## Milestone 2: Model management

Package: `internal/models`

### Catalog

Ship a static catalog of available models with, for each: name, download URL, file size, SHA256, approximate RAM requirement, and relative speed factor.

| Model | Size | RAM needed | Relative speed |
|---|---|---|---|
| tiny | 75MB | ~400MB | 1.0x baseline |
| base | 140MB | ~500MB | ~0.6x |
| small | 465MB | ~1GB | ~0.25x |
| medium | 1.5GB | ~2.5GB | ~0.1x |
| large-v3 | 3GB | ~4GB | ~0.05x |

Speed factors are rough. They get replaced by real measurements in Milestone 3.

### Operations

- List installed and available models
- Download with resumable transfer, progress events to the frontend, and SHA256 verification on completion
- Delete a model, with a guard against deleting the one currently in use
- Report free disk space before starting a download

### Do not

Bundle model weights in the installer. This is the single biggest reason competing tools have 2GB downloads.

### Done when

A user can download, verify, switch, and delete models, and an interrupted download resumes rather than restarting.

---

## Milestone 3: Hardware benchmark and recommendation

Package: `internal/hardware`

### Detection

- **RAM:** `gopsutil` for total and available. Available is the real constraint, not total.
- **CPU cores:** `runtime.NumCPU()` for thread count. Poor speed predictor, do not use it for recommendations.
- **GPU:** platform-specific, three code paths:
  - macOS: `sysctl` for chip name, detect Apple Silicon, enable Metal
  - Windows and Linux: look for `nvidia-smi`, parse VRAM from its output
  - Fallback: assume CPU only

### Benchmark

Detection alone is a weak predictor. An 8-core laptop from 2019 and an 8-core machine from 2025 are not comparable.

1. Ship a 30 second audio sample in the binary
2. On first run, transcribe it with the `base` model
3. Measure real-time factor for this specific machine
4. Extrapolate to other models using the relative speed factors

Twenty seconds of setup produces a far better answer than any spec sheet reading.

### Presentation

Show the tradeoff in the user's terms, not yours:

> On your machine, one hour of audio takes about **6 minutes** with this model, or **25 minutes** with the more accurate one.

Not "8GB RAM detected, recommending small." Most users do not know what a model size means.

### Rules

- Recommendation, never a lock. Show all models with their estimated time on this hardware, preselect yours, allow override.
- Cache the benchmark result. Re-run if hardware changes, and expose a manual re-run in settings for users who install a GPU.
- Note in the UI that laptops on battery may throttle and run slower than the benchmark suggests.

### Done when

First launch produces a sensible recommendation with honest time estimates, and the user can override it.

---

## Milestone 4: Transcription pipeline

Package: `internal/transcribe`

This is the core. Everything else hangs off it.

### Job state machine

```
queued -> converting -> transcribing -> done
                    \-> failed
                    \-> cancelled
```

Persist state to disk so a user who drops fifteen files and closes the laptop does not lose the queue.

### Process execution

- `exec.CommandContext` per job, cancel func stored in the job record
- Invoke `whisper-cli` with JSON output and segment-level progress
- Read stdout line by line, parse each completed segment
- Emit each segment to the frontend via `runtime.EventsEmit` as it arrives

Progressive fill matters. A 20 minute job that shows text appearing feels alive; one that shows a spinner feels frozen.

### Why shelling out rather than cgo bindings

- Keeps the project pure Go, so `GOOS=darwin GOARCH=arm64 go build` works from any CI runner
- Process isolation: an inference crash does not take the UI with it
- Cancellation is just killing a process

The cost is coarser control over the model context. Acceptable for Phase 1, revisit if needed.

### Queue behaviour

- One job at a time by default. Parallel jobs compete for the same RAM and GPU and usually make things slower.
- Pause, resume, cancel per job
- Reorder pending jobs
- On cancel: kill process, clean temp files, mark cancelled

### Done when

A queue of mixed files runs to completion unattended, and cancelling mid job leaves no orphan processes or temp files.

---

## Milestone 5: Transcript view

This screen is the product. Budget more time for it than feels reasonable.

### Requirements

- Segment list with timestamps
- Click any segment to seek audio to that point
- Current segment highlighted during playback
- Fully editable text, inline
- Autosave on edit, debounced
- Find and replace across the whole transcript
- Keyboard shortcuts: play/pause, skip back 5s, next/previous segment
- Undo and redo

### Why find and replace is not optional

Fixing a name that whisper spelled wrong 40 times is the single most common post-transcription task. A tool without it pushes users into a text editor, and then they stop coming back.

### Data model note

Store segments as individual records with `start`, `end`, `text`, `job_id`. Not as one text blob.

This is a Phase 1 decision with Phase 2 consequences: FTS5 search in the library feature needs to return a timestamp, not a document. Retrofitting this after users have data is unpleasant.

### Done when

Someone can transcribe an interview, fix every misspelled name, and navigate the audio by clicking text.

---

## Milestone 6: Export

### Formats

- **TXT** plain text, optional timestamps
- **SRT** subtitles
- **VTT** web subtitles
- **JSON** full structure with word-level timings

The JSON export is what makes you useful to developers building on top of your output. Do not skip it.

### Also

- Copy full transcript to clipboard
- Remember last used export directory

### Phase 1 subtitle caveat

Proper subtitle formatting (character-per-line caps, reading speed limits, sensible break points) is Phase 2. Phase 1 SRT is raw segments.

**Reconsider this if your audience is video editors rather than journalists.** Raw segment SRT is close to worthless for subtitling work, and if that is your core user, this feature moves into Phase 1.

### Done when

All four formats export correctly and the SRT loads in VLC without errors.

---

## Milestone 7: Language handling

- Auto-detect by default, using whisper's own detection
- Manual override with a searchable language list
- Translate-to-English toggle, which whisper does natively
- Show the detected language in the UI so users can catch a wrong detection

Ninety plus language support is a genuine advantage over English-first commercial tools. Make it visible.

---

## Milestone 8: Polish before release

- Empty states and first-run guidance
- Every error message readable by a non-technical person
- Crash recovery: restore the queue on next launch
- Settings: model directory, temp directory, thread count, re-run benchmark
- About screen with version and license
- README with screenshots, install instructions per platform, and the Linux webkit dependency spelled out

---

## Build order summary

| Order | Milestone | Rough weight |
|---|---|---|
| 1 | Skeleton and distribution | Medium |
| 2 | Media intake | Small |
| 3 | Model management | Medium |
| 4 | Hardware benchmark | Small |
| 5 | Transcription pipeline | Large |
| 6 | Transcript view | Large |
| 7 | Export | Small |
| 8 | Language handling | Small |
| 9 | Polish | Medium |

The two large items are the pipeline and the transcript view. Everything else is comparatively mechanical.

---

## Decisions to lock before starting

1. **Binary bundling:** embed and extract, versus ship alongside in the app bundle. Recommendation is embed.
2. **Segment storage schema:** individual records, not blobs. Locks in Phase 2 search.
3. **Subtitle formatting placement:** Phase 1 or Phase 2, depending on whether video editors are a core audience.
4. **Default concurrency:** one job at a time. Changing later is easy; assuming parallel from the start is not.

---

## Open risks

**WebKitGTK on Linux.** The webview differs meaningfully from WebView2 and macOS WebKit. Test the transcript view there early, not at the end.

**Windows SmartScreen.** Unsigned installers get flagged, and your non-technical audience will not click through the warning. Code signing certificates cost money. Budget for it or accept the friction.

**Long file memory.** A 4 hour recording produces a large segment list. Virtualize the transcript list rather than rendering every segment.

**Model download failures.** Users on slow or unreliable connections downloading 1.5GB. Resumable transfer is not optional.