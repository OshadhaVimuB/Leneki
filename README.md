# Leneki

**Offline desktop transcription.** Drop in a recording, get an accurate editable transcript, export SRT. Powered by [whisper.cpp](https://github.com/ggerganov/whisper.cpp), nothing leaves your machine.

> **Status: early development.** There is no installable release yet and no application code in this repository. What exists today is the design: the [scope](documents/initial-plan.md), the [architecture](documents/architecture/v1.0.0.md), and a [build plan](documents/development-plan/phase-1.md) broken into 19 parts. If you want to help shape it before the first line of Go is written, now is the good moment. See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## What it is

Leneki turns audio and video recordings into text you can read, correct, and export. It is a desktop application, not a service. You install it, and it works with the network cable unplugged.

It does this by wrapping two well established command line tools, whisper.cpp for speech recognition and [FFmpeg](https://ffmpeg.org/) for media decoding, in an interface a non-technical person can use without ever seeing a terminal.

## Why it exists

Transcription is either expensive, or it is a privacy problem, or it requires you to be comfortable with a terminal. Usually two of the three.

Cloud transcription services charge by the minute and require you to upload the recording. For a journalist with a confidential source, a doctor with patient audio, a lawyer with privileged material, or a researcher under an ethics agreement, uploading is not a pricing question. It is disqualifying.

The offline alternatives are real and good, and most of them are command line tools. That is a fine answer for developers and no answer at all for the people who most need the privacy.

Leneki is aimed at the gap: the accuracy and privacy of running Whisper locally, in an application you install by double clicking.

## What version 1 will do

None of this is built yet. This is the scope of the first release, and it is deliberately small.

- **Transcribe** audio and video in over 90 languages, with automatic language detection and a translate-to-English option
- **Handle the formats people actually have**, MP4, M4A, MP3, WAV and anything else FFmpeg decodes, with no manual conversion
- **Show text as it arrives** rather than a spinner, so a long job visibly progresses
- **Edit the transcript inline**, with autosave, undo, and find and replace across the whole document
- **Navigate by clicking text** to jump the audio to that moment
- **Export** to TXT, SRT, VTT, and JSON with word level timings, with subtitles available raw or properly formatted for subtitling work
- **Queue many files** and leave it running, with the queue surviving a crash or a closed laptop
- **Recommend a model honestly**, by benchmarking your actual machine and telling you the tradeoff in minutes rather than gigabytes

Explicitly **not** in version 1: speaker diarization, a searchable transcript library, custom vocabulary, live microphone capture, summarization, and anything cloud.

## How it works

```
Your file  ->  FFmpeg  ->  whisper.cpp  ->  transcript in a local database
              decode       recognize        edit and export
```

Leneki is a single Go program with a user interface rendered by the operating system's own webview. FFmpeg and whisper.cpp are bundled inside the installer and extracted on first run, so there is nothing else to install. Model weights are downloaded once, on your terms, and never bundled into the installer, which is why the download is tens of megabytes rather than gigabytes.

Everything is described in detail in the [architecture document](documents/architecture/v1.0.0.md).

## Platforms

Windows x64, macOS on Apple Silicon, macOS on Intel, and Linux x64.

Linux additionally needs WebKitGTK, which most desktop distributions already have. The exact package name will be documented with the first release.

## Privacy

This is the point of the project, so it is worth being specific.

- No account, no sign in, no telemetry, no analytics, no crash reporting.
- The only network request Leneki ever makes is downloading a speech model, over HTTPS, from a URL compiled into the application. You can see them in the source.
- Your recordings and transcripts are read and written on your own disk. There is no code path that sends them anywhere.
- Source media is opened read only. Leneki never modifies your files.

## Documentation

| Document | What it covers |
|---|---|
| [Phase 1 scope](documents/initial-plan.md) | What version 1 includes, and what it deliberately leaves out |
| [Architecture v1.0.0](documents/architecture/v1.0.0.md) | Components, package layout, data model, contracts, build pipeline |
| [Phase 1 build plan](documents/development-plan/phase-1.md) | The work split into 19 parts, in dependency order |
| [Contributing](CONTRIBUTING.md) | How to get set up and how the work is organized |

## Contributing

Contributions are welcome, and the project is at the stage where they have the most leverage. The [build plan](documents/development-plan/phase-1.md) divides the work into parts you can pick up one at a time, each with a clear goal and a checkable definition of done.

Start with [CONTRIBUTING.md](CONTRIBUTING.md).

## Licence

Leneki is licensed under the [Apache License 2.0](LICENSE).

It ships two third party programs, each under its own licence and each invoked as a separate executable rather than linked in:

- **whisper.cpp**, MIT licence
- **FFmpeg**, LGPL 2.1 or GPL 2.0 depending on the build

Speech model weights are downloaded separately and are MIT licensed. Full attribution and licence texts ship with the application and will live in `assets/licenses/`.
