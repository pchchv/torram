# torram — Torrent Telegram Downloader

A lightweight, concurrent, and robust Go-based service that automatically downloads torrent files or direct torrent links sent via a Telegram bot.   
The service handles files asynchronously using a resilient worker pool architecture and ensures graceful shutdown on system interrupts.