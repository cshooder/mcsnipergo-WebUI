<h3 align="center">
  <img src="https://i.imgur.com/ShMq72J.png" alt="MCsniperGO"></img>

  
  Originally by kqzz. Expanded upon by cshooder
</h3>

<p align="center">
    <a href="https://github.com/Kqzz/MCsniperGO/releases/"><img alt="downloads" src="https://img.shields.io/github/downloads/Kqzz/MCsniperGO/total?color=%233889c4" height="22"></a>
</p>

MCsniperGO is a Minecraft username sniper with both a web-based user interface and a command-line interface. It allows you to automatically claim Minecraft usernames when they become available. This version has been significantly expanded by cshooder, adding a modern web UI and various improvements to the original codebase by kqzz.

## Features

- **Modern Web UI**: Easy-to-use interface created by cshooder for sniping usernames without command-line knowledge
- **Account Management**: Support for both Gift Code accounts (no usernames) and Microsoft accounts
- **Proxy Support**: Use proxies to avoid rate limits
- **Efficient Sniping**: Written in Go for high performance
- **Enhanced Experience**: Streamlined workflow and improved reliability

## Usage - Web Interface

The new web interface by cshooder makes sniping usernames easier than ever:

1. [Install go](https://go.dev/dl/)
2. Download or clone the MCsniperGO repository
3. Open MCsniperGO folder in your terminal/cmd
4. Run `go run ./web/server.go` to start the web server
5. Open your browser and navigate to `http://localhost:8080` (or the configured port)
6. Use the web interface to:
   - Enter your accounts in the configuration section
   - Add proxies
   - Enter the target username and snipe delay
   - Start sniping with a single click

## Usage - Command Line Interface

For those who prefer the command line:

- [Install go](https://go.dev/dl/)
- Download or clone MCsniperGO repository 
- Open MCsniperGO folder in your terminal / cmd
- Put your prename accounts (no claimed username) in [`gc.txt`](#accounts-formatting) and your normal accounts in [`ms.txt`](#accounts-formatting)
- Put proxies into `proxies.txt` in the format `user:pass@ip:port` (there should NOT be 4 `:` in it as many proxy providers provide it as)
- Run `go run ./cmd/cli`
- Enter username + [claim range](#claim-range)
- Wait, and hope you claim the username!

## Claim Range
Use the following Javascript bookmarklet in your browser to obtain the droptime while on `namemc.com/search?q=<username>`:

```js
javascript:(function(){function parseIsoDatetime(dtstr) {
    return new Date(dtstr);
};

startElement = document.getElementById('availability-time');
endElement = document.getElementById('availability-time2');

start = parseIsoDatetime(startElement.getAttribute('datetime'));
end = parseIsoDatetime(endElement.getAttribute('datetime'));

para = document.createElement("p");
para.innerText = Math.floor(start.getTime() / 1000) + '-' + Math.ceil(end.getTime() / 1000);

endElement.parentElement.appendChild(para);})();

```

If 3name.xyz has a lower length claim range for a username I would recommend using that, you can get the unix droptime range with this bookmarklet on `3name.xyz/name/<name>`

```js
javascript: (function() {
    startElement = document.getElementById('lower-bound-update');
    endElement = document.getElementById('upper-bound-update');
  
  	if (startElement === null) {
    	startElement = 0;
    } else {
      startElement = startElement.getAttribute('data-lower-bound')
    }
  
  
    para = document.createElement("p");
    para.innerText = Math.floor(Number(startElement) / 1000) + '-' + Math.ceil(Number(endElement.getAttribute('data-upper-bound')) / 1000);
    endElement.parentElement.appendChild(para)
})()
```

## Accounts Formatting

- Place in `gc.txt` or `ms.txt` depending on their account type.
  - `gc.txt` is for accounts without usernames
  - `ms.txt` is for accounts that already have usernames on them
  - **Web UI users**: You can add these directly through the interface
- A bearer token can be obtained by following [this guide](https://kqzz.github.io/mc-bearer-token/)

```txt
EMAIL:PASSWORD
BEARER
```

### Example accounts file

```txt
kqzz@gmail.com:SecurePassword3
teun@example.com:SafePassword!
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```
> This will load 3 accounts into the sniper, two of which are supplied with email / password. The last is loaded by bearer token, and will last 24 hours (the sniper will show the remaining time).

> Their account types are determined by if they are placed in `gc.txt` or `ms.txt`.

## Understanding Logs

Each request made to change your username will return a 3 digit HTTP status code, the meanings are as follows:

- 400 / 403: Failed to claim username (will continue trying)
- 401: Unauthorized (restart claimer if it appears)
- 429: Too many requests (add more proxies if this occurs frequently)
