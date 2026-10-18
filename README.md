# torram — Torrent Telegram Downloader

A lightweight, concurrent, and robust Go-based service that automatically downloads torrent files or direct torrent links sent via a Telegram bot.   
The service handles files asynchronously using a resilient worker pool architecture and ensures graceful shutdown on system interrupts.

## 🚀 Features

- **Telegram Bot Interface:** Send a `.torrent` file or a direct HTTP/HTTPS link to the bot to trigger a download.
- **Concurrent Worker Pool:** Dedicated, configurable workers pick up files asynchronously to download torrent payloads.
- **Robust Folder Monitoring:** A scheduled background scanner safely polls a directory for incoming torrent tasks.
- **Graceful Shutdown:** Full application context lifecycle management (`SIGINT`/`SIGTERM`) ensuring that ongoing file operations or active downloads terminate cleanly without corrupting state.
- **Cross-Platform Pathing:** Built entirely using native Go pathing mechanisms (`filepath.Join`), ensuring safe executions across Linux, macOS, and Windows.