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
  - `-i`, `--interval <duration>`: Fetch feeds repeatedly at the specified interval (e.g., `1h`, `30m`). If 0 or not specified, fetches only once.
- `read`: Launches the TUI to browse feeds and articles.
  - `--no-fetch`: Disables background fetching of articles while the TUI is running.
  - `-s`, `--speaking-rate <rate>`: Sets the speaking rate for TTS (default: 1.3, or value from config).
  - `--voicevox`: Uses the VoiceVox engine for TTS (requires VoiceVox configuration).
  - `--speaker <id>`: Sets the VoiceVox speaker ID (default: 10, or value from config).
  - `--non-interactive`: Run in non-interactive mode without TUI (useful for systemd services).
- `play`: Read aloud unlistened summaries.
  - `--no-fetch`: Disables background fetching of articles while playing audio.
  - `--date <YYYY-MM-DD>`: Plays summaries published on the specified date.
  - `-s`, `--speaking-rate <rate>`: Sets the speaking rate for TTS (default: 1.3, or value from config).
  - `--voicevox`: Uses the VoiceVox engine for TTS (requires VoiceVox configuration).
  - `--speaker <id>`: Sets the VoiceVox speaker ID (default: 10, or value from config).
- `import <opmlfile>`: Import feeds from an OPML file.
- `bookmark <URL>`: Adds a new bookmark (web page) to a special feed.
- `publish [YYYY-MM-DD]`: Processes articles for the specified date (defaults to today) and the preceding two days. For each day and each feed, it merges the audio files of the summaries published on that day into a single MP3 file (named `YYYY-MM-DD_FeedTitle.mp3`). These merged MP3 files, along with an updated podcast RSS feed (`rss.xml`), are then uploaded to Cloudflare R2. This command requires the `AudioPath` and `Podcast` sections to be configured in the `config.toml` file.
- `export-audio`: Regenerates and saves audio files for all existing summaries based on current TTS settings. This is useful if you change TTS engines or settings and want to update previously generated audio.

### Global Options

These options can be used with any subcommand:

- `--db <path>`: Path to the SQLite database file (default: `~/quicknews.db`).
- `--config <path>`: Path to the configuration file (default: `~/.config/quicknews/config.toml`).
- `--log <path>`: Path to the log file (default: `~/quicknews.log`). If not specified, logs to standard output.
- `-V`, `--version`: Show version information.
- `-d`, `--debug`: Enable debug logging.

For detailed options for each subcommand, refer to `./quicknews <subcommand> --help`.

## Running as a Service

The following commands can run in non-TTY environments such as systemd services:

- `read` - Requires `--non-interactive` flag when running without TTY
- `play` - Automatically detects TTY absence and runs in non-interactive mode
- `fetch` - Automatically detects TTY absence and runs in non-interactive mode  
- `publish` - Console output only, works without TTY
- `add`, `import`, `bookmark`, `export-audio` - All work without TTY

### Examples for systemd services:

For continuous feed fetching with UI (requires real TTY):
```
[Unit]
Description=Quicknews TUI Service
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=/path/to/quicknews read --interval=1h
Restart=always
RestartSec=10
Environment=HOME=%h
TTYPath=/dev/tty1
TTYReset=yes

[Install]
WantedBy=graphical-session.target
```

For background feed fetching (no TTY needed):
```
[Unit]
Description=Quicknews Background Fetch Service
After=network.target
Wants=network.target

[Service]
Type=oneshot
ExecStart=/path/to/quicknews fetch
RemainAfterExit=yes
Environment=HOME=%h

[Install]
WantedBy=timers.target
```

For systemd timer-based fetching (recommended for periodic background fetching):
1. Create `~/.config/systemd/user/quicknews-fetch.service`:
```
[Unit]
Description=Quicknews Fetch Articles
Wants=network.target

[Service]
Type=oneshot
ExecStart=/path/to/quicknews fetch
Environment=HOME=%h
```

2. Create `~/.config/systemd/user/quicknews-fetch.timer`:
```
[Unit]
Description=Run quicknews fetch every hour
Requires=quicknews-fetch.service

[Timer]
OnBootSec=10min
OnUnitActiveSec=1h
Persistent=true

[Install]
WantedBy=timers.target
```

3. Enable and start the timer:
```bash
systemctl --user daemon-reload
systemctl --user enable quicknews-fetch.timer
systemctl --user start quicknews-fetch.timer
```

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

# Databse File settings
db = "/path/to/your/quicknews.db"

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

# Podcast settings (Optional)
# If you want to distribute your feed as a podcast,
# please configure cloudflare to host the files and set up the podcast feed information.
[cloudflare]
access_key_id="your cloudflare access_key_id"
secret_access_key="your cloudflare secret_access_key"
bucket_name="r2 bucket name"
endpoint_url="r2 endpoint url"

[podcast]
channel_title = "podcast channel title"
channel_link = "podcast channel link"
channel_desc = "podcast channel desc"
author = "podcast author"
publish_url = "podcast publish url"

```

The core RSS reading functionality works without configuring these optional features.
