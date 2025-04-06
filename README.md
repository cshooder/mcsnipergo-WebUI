<h3 align="center">
  <img src="https://i.imgur.com/ShMq72J.png" alt="MCsniperGO"></img>
</h3>

<p align="center">
    <a href="https://github.com/cshooder/MCsniperGO/releases/"><img alt="downloads" src="https://img.shields.io/github/downloads/cshooder/MCsniperGO/total?color=%233889c4" height="22"></a>
</p>

MCsniperGO is a Minecraft username sniper with a modern web-based user interface. It allows you to automatically claim Minecraft usernames when they become available. This version features a completely redesigned interface and significant improvements to the sniping process.

## Features

- **Modern Web UI**: Easy-to-use interface for sniping usernames without command-line knowledge
- **Account Management**: Support for both Gift Code accounts (no usernames) and Microsoft accounts
- **Proxy Support**: Use proxies to avoid rate limits
- **Efficient Sniping**: Written in Go for high performance
- **Enhanced Experience**: Streamlined workflow and improved reliability

## Running the Web Interface

Follow these simple steps to get started:

1. **Install Go:**
   - Download and install Go from [go.dev/dl](https://go.dev/dl/)
   - Make sure Go is properly installed by running `go version` in your terminal/command prompt

2. **Get the MCsniperGO code:**
   - Download the latest release from GitHub or clone the repository
   - Extract the files to a folder of your choice

3. **Start the web server:**
   - Open your terminal/command prompt
   - Navigate to the MCsniperGO folder: `cd path/to/MCsniperGO`
   - Run the web server with: `go run ./web/start_server.py`
   - If you encounter any issues, try: `python3 ./web/start_server.py`

4. **Access the web interface:**
   - Open your browser and go to `http://localhost:8081`
   - You should see the MCsniperGO web interface

5. **Configure and use:**
   - Enter your accounts in the configuration section
   - Add proxies if needed
   - Enter the target username and snipe delay
   - Start sniping with a single click

## Accounts Formatting

- You can add your accounts directly through the web interface
- For accounts without usernames, use the GC Accounts section
- For accounts that already have usernames, use the MS Accounts section
- Accounts can be in either of these formats:
  ```
  EMAIL:PASSWORD
  BEARER_TOKEN
  ```

## Understanding Logs

Each request made to change your username will return a 3 digit HTTP status code, the meanings are as follows:

- 400 / 403: Failed to claim username (will continue trying)
- 401: Unauthorized (restart claimer if it appears)
- 429: Too many requests (add more proxies if this occurs frequently)
