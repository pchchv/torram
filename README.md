# torram — Torrent Telegram Downloader
[![CI](https://github.com/pchchv/express/workflows/CI/badge.svg)](https://github.com/pchchv/express/actions?query=workflow%3ACI+event%3Apush)

A lightweight, concurrent, and robust Go-based service that automatically downloads torrent files or direct torrent links sent via a Telegram bot.   
The service handles files asynchronously using a resilient worker pool architecture and ensures graceful shutdown on system interrupts.

## 🚀 Features

- **Telegram Bot Interface:** Send a `.torrent` file or a direct HTTP/HTTPS link to the bot to trigger a download.
- **Concurrent Worker Pool:** Dedicated, configurable workers pick up files asynchronously to download torrent payloads.
- **Robust Folder Monitoring:** A scheduled background scanner safely polls a directory for incoming torrent tasks.
- **Graceful Shutdown:** Full application context lifecycle management (`SIGINT`/`SIGTERM`) ensuring that ongoing file operations or active downloads terminate cleanly without corrupting state.
- **Cross-Platform Pathing:** Built entirely using native Go pathing mechanisms (`filepath.Join`), ensuring safe executions across Linux, macOS, and Windows.

```
       [ User ]
          │  (Sends .torrent file / URL link)
          ▼
   ┌──────────────┐
   │ Telegram Bot │
   └──────┬───────┘
          │ (Downloads file & saves to WatchDir)
          ▼
   ┌──────────────┐
   │   WatchDir   │ ◄───┐
   └──────┬───────┘     │
          │             │ (Scans every 5s)
          ▼             │
   ┌──────────────┐     │
   │ Directory    ├─────┘
   │   Scanner    │
   └──────┬───────┘
          │ (Dispatches Task)
          ▼
   ┌──────────────┐
   │ Worker Pool  │
   └──────┬───────┘
          │ (Downloads data via anacrolix/torrent)
          ▼
   ┌──────────────┐
   │  OutputDir   │ (Final Payload Location)
   └──────────────┘
```

## ⚙️ Configuration

The application is configured using environment variables. Create a `.env` file in the root directory of your project:

```env
# Telegram Bot Token obtained from @BotFather
TG_BOT_TOKEN=123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ

# Root directory path to monitor and output downloaded payloads
DOWNLOAD_DIR=/path/to/your/torrent/workspace

# Maximum number of concurrent torrent downloads allowed at the same time
MAX_NUM_WORKERS=3
```

### Directory Structure Generated Automatically:
- **Watch Directory (`DOWNLOAD_DIR`):** The landing zone where incoming `.torrent` files are saved.
- **Target Directory (`DOWNLOAD_DIR/Downloaded`):** The archive folder where `.torrent` files are safely moved *after* successful data payload completion.
- **Output Directory (`DOWNLOAD_DIR/Files`):** The destination where the actual torrent contents (movies, music, software) are written during download.

## 🏃 Getting Started

1. **Clone the repository:**
   ```bash
   git clone https://github.com/pchchv/torram
   cd torram
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set up your environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your real credentials
   ```

4. **Run the service:**
   ```bash
   go run .
   ```

## 🔒 Error Handling & Resiliency

- **Self-Healing Scanner:** If the target directory becomes temporarily unavailable (e.g., a disconnected network-attached storage device), the directory scanner logs the anomaly and continues running. It retries cleanly on the next 5-second interval without dropping the application runtime.
- **Duplicate Prevention:** Using a thread-safe `sync.Map` workflow state machine, files that are currently in the queue or being actively downloaded will not be fed to workers twice.
- **Failure Recovery:** If a worker encounters a network drop or corrupted file payload, it automatically flushes the tracking footprint out of the memory matrix, allowing the application to attempt processing the file again on subsequent scan loops.
- **Resource Leak Safeguards:** Active progress tracking tickers are strictly coupled to contextual lifetimes. When a download completes or the application undergoes an interrupt, execution tracks immediately return, preventing goroutine leaks.