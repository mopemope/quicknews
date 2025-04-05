# quicknews

quicknews is a terminal-based RSS reader. It can summarize article content using LLMs and read it aloud.

This tool is a personal project. Its primary purpose is for my own use.

## Features

- Add, edit, and delete RSS feeds from the command line.
- Fetch and update RSS feeds.
- Browse feeds and articles using a TUI (Terminal User Interface).
- Summarize articles using LLMs (Large Language Models).
- Convert summaries to audio using Google Text-to-Speech.

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
- `play`: Read aloud unlistened feeds.

For detailed options, refer to `./quicknews --help`.

## Configuration

Some features require specific environment variables to be set:

- **Google Text-to-Speech:**
    - If you want to use the text-to-speech feature, you need to set up Google Cloud authentication.
    - Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of your service account key file.
    - Example: `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/keyfile.json"`

- **Gemini Summarization:**
    - If you want to use the article summarization feature with Google Gemini, you need an API key.
    - Set the `GEMINI_API_KEY` environment variable to your API key.
    - Example: `export GEMINI_API_KEY="YOUR_API_KEY"`

These features are optional. The core RSS reading functionality works without these environment variables.
