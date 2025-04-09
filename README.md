# quicknews

quicknews is a terminal-based RSS reader. It can summarize article content using LLMs and read it aloud.

This tool is a personal project. Its primary purpose is for my own use.

## Features

- Add RSS feeds from the command line (`add`).
- Import feeds from an OPML file (`import`).
- Fetch and update RSS feeds (`fetch`), optionally at regular intervals (`fetch --interval`).
- Browse feeds and articles using a TUI (Terminal User Interface) (`read`).
- Summarize articles using LLMs (Large Language Models) like Google Gemini.
- Convert summaries to audio using Google Text-to-Speech.
- Play unlistened summaries aloud (`play`).
- Export summaries to Org mode files (optional, requires `EXPORT_ORG` environment variable).

## How to Compile

Requires Go 1.24 or later.

```bash
go build -o quicknews .
```

## How to Run

After building, you can run the program with the following command:

```bash
./quicknews <subcommand> [options]
```

### Main Subcommands

- `add <URL>`: Adds a new RSS feed.
- `fetch`: Fetches and updates registered feeds.
- `read`: Launches the TUI to browse feeds and articles.
    - `--speaking-rate` / `-s`: Sets the speaking rate (default: 1.2).
    - `--voicevox`: Uses the VoiceVox engine for TTS.
    - `--speaker`: Sets the VoiceVox speaker ID (default: 10).
- `play`: Read aloud unlistened feeds.
    - `--speaking-rate` / `-s`: Sets the speaking rate (default: 1.2).
    - `--voicevox`: Uses the VoiceVox engine for TTS.
    - `--speaker`: Sets the VoiceVox speaker ID (default: 10).
- `import`: Import feeds from an OPML file.

For detailed options, refer to `./quicknews --help`.

## Configuration

Some features require specific environment variables to be set:

- **Google Text-to-Speech:**
    - If you want to use the text-to-speech feature, you need to set up Google Cloud authentication.
    - Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of your service account key file.
    - Example: `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/keyfile.json"`
    - Alternatively, you can use the VoiceVox engine by running the `play` or `read` command with the `--voicevox` flag. Ensure the VoiceVox engine is running locally (usually at `http://localhost:50021`). You can specify the speaker ID using the `--speaker` flag (e.g., `--speaker 10`).

- **Gemini Summarization:**
    - If you want to use the article summarization feature with Google Gemini, you need an API key.
    - Set the `GEMINI_API_KEY` environment variable to your API key.
    - Example: `export GEMINI_API_KEY="YOUR_API_KEY"`

- **Org Mode Export (Optional):**
    - To export summaries as Org mode files, set the `EXPORT_ORG` environment variable to the desired destination directory path.
    - Example: `export EXPORT_ORG="/path/to/your/org/files"`

These features are optional. The core RSS reading functionality works without these environment variables.
