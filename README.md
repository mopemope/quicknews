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
- `bookmark <URL>`: Adds a new bookmark.

### Global Options

These options can be used with any subcommand:

- `--db <path>`: Path to the SQLite database file (default: `~/quicknews.db`).
- `--config <path>`: Path to the configuration file (default: `~/.config/quicknews/config.toml`).
- `--log <path>`: Path to the log file (default: `~/quicknews.log`). If not specified, logs to standard output.
- `-V`, `--version`: Show version information.
- `-d`, `--debug`: Enable debug logging.

For detailed options for each subcommand, refer to `./quicknews <subcommand> --help`.

## Configuration

quicknews uses a configuration file located at `$HOME/.config/quicknews/config.toml`. Create this file if it doesn't exist.

Here's an example `config.toml` with available options:

```toml
# Enable overriding settings with environment variables (default: false)
# If true, environment variables like GOOGLE_APPLICATION_CREDENTIALS, GEMINI_API_KEY, etc.,
# will take precedence over the values in this file.
# enable_env_override = false

# Default speaking rate for TTS (default: 1.3, set in code if not specified here or by env)
speaking_rate = 1.3

# Require confirmation before performing certain actions (e.g., deleting)
require_confirm = true

# Google Text-to-Speech settings
# Path to your Google Cloud service account key file.
# Required if using Google TTS and not using environment variables (when override is enabled).
google_application_credentials = "/path/to/your/keyfile.json"

# Gemini API settings
# Your Google Gemini API key.
# Required for article summarization if not using environment variables (when override is enabled).
gemini_api_key = "YOUR_API_KEY"

# Org Mode Export settings (Optional)
# Directory path to export summaries as Org mode files.
export_org = "/path/to/your/org/files"

# VoiceVox settings (Optional)
[voicevox]
# Default speaker ID for VoiceVox.
# Can be overridden by the --speaker flag or VOICEVOX_SPEAKER env var.
speaker = 10

```


The core RSS reading functionality works without configuring these optional features.
