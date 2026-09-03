# Contributing to Leneki

Thanks for looking. Leneki is an offline desktop transcription application, and it is at the earliest possible stage: the design is written, the code is not. That makes this a good moment to contribute, because the decisions are still cheap to change.

This document covers how the work is organized, how to get set up, and the conventions the project follows. Read the [architecture document](documents/architecture/v1.0.0.md) before writing code, since it answers most "why is it done this way" questions.

---

## The best way to help right now

There is no application code yet, so the useful contributions today are not pull requests full of features.

1. **Review the design.** The [architecture](documents/architecture/v1.0.0.md) commits to specific choices, and several are worth arguing about before they become load bearing. Open an issue if you disagree with one.
2. **Claim a part.** The [Phase 1 build plan](documents/development-plan/phase-1.md) splits the work into 19 parts in dependency order. Each part states its goal, what it depends on, the files it creates, and a checkable definition of done. Comment on the tracking issue for a part to claim it.
3. **Test on your platform.** Once builds exist, the four targets need real users on real machines. Linux under WebKitGTK especially, since it is the least forgiving of the three webview engines.

If you are new to the project, the parts marked **S** in the build plan are the smallest self-contained pieces of work.

---

## Project layout

```
documents/          Scope, architecture, and build plan. Read these first.
internal/           Go packages. One responsibility each, see the architecture.
frontend/           Svelte user interface.
assets/             Bundled binaries, licences, benchmark sample.
build/              Packaging configuration per platform.
```

The `internal/` tree does not exist yet. It is created part by part, following the build plan.

---

## Getting set up

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.23 or later |
| Node.js | 20 LTS |
| Wails CLI | v2.9 or later, `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |

Plus your platform's native toolchain: Xcode command line tools on macOS, the WebView2 SDK on Windows, `libwebkit2gtk-4.0-dev` and `build-essential` on Linux.

Run `wails doctor` first. It reports missing native dependencies clearly and will save you an hour of confusing build errors.

### Running it

```
wails dev      # development build with hot reload
wails build    # production build for your platform
go test ./...  # the Go test suite
```

Wails cannot cross compile, because it binds to each platform's native webview. You can only build for the machine you are on.

---

## How the work is organized

Every piece of work in Phase 1 maps to a part in the [build plan](documents/development-plan/phase-1.md). A part is a self-contained unit with:

- a **goal**, one sentence
- a **depends on** line, which is the real ordering constraint
- the **files** it creates
- a **done when**, which is a check you can actually run

A part is finished when its done-when holds, not when the code compiles. If a part's done-when turns out to be wrong or untestable, say so in the issue rather than quietly working around it.

### Scope discipline

Phase 1 is deliberately small, and keeping it small is what gets it shipped. The following are out of scope, and pull requests adding them will be asked to wait, however good they are:

speaker diarization, searchable transcript library, custom vocabulary, voice activity detection, live microphone capture, summarization, anything cloud, automatic updates, interface localization.

The scope boundary is in the [Phase 1 scope document](documents/initial-plan.md). If you think something belongs in Phase 1 that is not there, open an issue and make the case before writing it.

---

## Conventions

These are project rules, not preferences. They exist so the codebase reads as though one person wrote it.

### Code

- **No em dashes** anywhere, in code, comments, documentation, or commit messages. Use a comma, colon, or period.
- **Comments are short and rare.** One line, only where the code is not self-explanatory. No block comments, no docstring headers, no comment restating what the line below does.
- **Match the surrounding style.** If you would do it differently, do it the way the file already does it.
- **Touch only what your change requires.** Do not reformat, rename, or improve adjacent code in the same pull request. If you spot something unrelated that is wrong, open an issue.
- **Simplest thing that solves the problem.** No abstraction for a single use. No configurability nobody asked for. No error handling for impossible states.

### Documentation

Leneki is public, so every document is written for a reader with no prior context. Lead with what the thing is and why it exists. Use headings and short paragraphs. Define project-specific terms on first use. Update the affected documents in the same pull request rather than leaving stale instructions behind.

### Commit messages

One short subject line, then a description. No `fix:`, `feat:`, `chore:` or any other prefix tag, and no trailing period on the subject.

The description covers three things, in this order: **Why** the commit exists, **What changed**, and **To test**, meaning what someone should check to confirm it works. A sentence or two each.

```
Resume model downloads instead of restarting them

Why: the large model is 3GB and users on unreliable connections were
losing the whole download on a dropped connection.

What changed: downloads write to a .part file and reopen with an HTTP
Range request. SHA256 verification now runs before the file is moved
into place, so a partial file can never be used as a model.

To test: start a large model download, kill the network mid transfer,
reconnect, and confirm it resumes from where it stopped rather than
starting at zero.
```

---

## Branches and pull requests

- `main` holds releases.
- `dev` is the integration branch. **Branch from `dev` and open your pull request against `dev`.**
- Name branches after the work, for example `part-5-media-intake` or `fix-srt-timestamps`.

### Before you open a pull request

- [ ] `go test ./...` passes
- [ ] The part's done-when actually holds, and you checked it rather than assuming
- [ ] No em dashes, no long comments, no unrelated changes in the diff
- [ ] Documentation updated in the same pull request if behaviour changed
- [ ] Commit messages follow the format above

Keep pull requests scoped to one part where possible. A reviewer can hold one part in their head. Three parts at once, they cannot.

---

## Testing

The architecture's [testing strategy](documents/architecture/v1.0.0.md#17-testing-strategy) defines three tiers, and knowing which one your change belongs to tells you what to write.

1. **Unit tests, no external dependencies.** State transitions, export writers, output parsing, time estimates, error mapping, migrations. This is most of the suite and it runs everywhere in under a second. New logic belongs here by default.
2. **Integration tests, need the bundled binaries.** Probing and converting real fixture media. Fixtures live in `testdata/` and should stay small, a few seconds each.
3. **Manual, needs a real machine.** Benchmark accuracy, GPU detection, playback in each webview, installer behaviour. Describe what you checked in the pull request, since a reviewer cannot reproduce your hardware.

The backend is deliberately built so that a whole transcription job can run inside `go test` with no window on screen. Services take an event emitter interface rather than importing the Wails runtime. Keep it that way: if you find yourself importing the Wails runtime into a package under `internal/`, that is a sign the logic belongs in the binding layer instead.

---

## Reporting bugs

Once there are releases, a useful bug report includes your operating system and version, the Leneki version, the model you were using, what you did, what happened, and what you expected. If it involves a specific file, the output of `ffprobe` on that file is worth more than a description of it.

Do not attach recordings that are confidential. That is the entire reason this project exists.

---

## Licence

Leneki is licensed under the [Apache License 2.0](LICENSE). By contributing, you agree that your contributions are licensed under the same terms.

If you add a third party dependency, check that its licence is compatible and say so in the pull request. The project bundles whisper.cpp and FFmpeg as separate executables rather than linking them, deliberately, and that boundary should not be crossed without discussion.
