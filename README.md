# ![](https://raw.githubusercontent.com/kost/tty2web/master/resources/favicon.png) tty2web - Share your terminal as a web application (bind/reverse)

[![Build Status](https://travis-ci.org/kost/tty2web.png)](https://travis-ci.org/kost/tty2web)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[release]: https://github.com/kost/tty2web/releases
[license]: https://github.com/kost/tty2web/blob/master/LICENSE

tty2web is a simple command line tool that turns your CLI tools into web applications. it is based on [Gotty](https://github.com/yudai/gotty), but heavily improved.
Improvements include implementation of bind and reverse mode (useful for penetration testing/NAT traversal), basic support for Windows, etc.

![Screenshot](https://raw.githubusercontent.com/kost/tty2web/master/screenshot.gif)

# Installation

Download the latest stable binary file from the [Releases](https://github.com/kost/tty2web/releases) page. Note that the release marked `Pre-release` is built for testing purpose, which can include unstable or breaking changes. Download a release marked [Latest release](https://github.com/kost/tty2web/releases/latest) for a stable build.

(Files named with `darwin_amd64` are for Mac OS X users)

## `go get` Installation (Development)

If you have a Go language environment, you can install tty2web with the `go get` command. However, this command builds a binary file from the latest master branch, which can include unstable or breaking changes. tty2web requires go1.9 or later.

```sh
$ go get github.com/kost/tty2web
```

# Usage

```
Usage: tty2web [options] <command> [<arguments...>]
```

Run `tty2web` with your preferred command as its arguments (e.g. `tty2web top`).

By default, tty2web starts a web server at port 8080. Open the URL on your web browser and you can see the running command as if it were running on your terminal.

# Example usage


Bind mode is simple, specify port to listen and command to run (add -w if you want to interact):
```
tty2web --port 8081 top
```
Point your web browser to IP and port 8081 in order to see output of top. Add -w if you want to interact directly in the browser.


For reverse mode, you need to start listener first:
```
tty2web --listen :4444 --server 127.0.0.1:8000 --password test
```

After having listener running, you can start client to connect to the listener:
```
tty2web --connect 192.168.1.1:4444 --password test -w /bin/sh
```

Point your web browser to http://127.0.0.1:8000

## Options

```
   --address value, -a value     IP address to listen (default: "0.0.0.0") [$TTY2WEB_ADDRESS]
   --port value, -p value        Port number to listen (default: "8080") [$TTY2WEB_PORT]
   --permit-write, -w            Permit clients to write to the TTY (BE CAREFUL) [$TTY2WEB_PERMIT_WRITE]
   --credential value, -c value  Credential for Basic Authentication (ex: user:pass, default disabled) [$TTY2WEB_CREDENTIAL]
   --random-url, -r              Add a random string to the URL [$TTY2WEB_RANDOM_URL]
   --random-url-length value     Random URL length (default: 8) [$TTY2WEB_RANDOM_URL_LENGTH]
   --tls, -t                     Enable TLS/SSL [$TTY2WEB_TLS]
   --tls-crt value               TLS/SSL certificate file path (default: "~/.tty2web.crt") [$TTY2WEB_TLS_CRT]
   --tls-key value               TLS/SSL key file path (default: "~/.tty2web.key") [$TTY2WEB_TLS_KEY]
   --tls-ca-crt value            TLS/SSL CA certificate file for client certifications (default: "~/.tty2web.ca.crt") [$TTY2WEB_TLS_CA_CRT]
   --index value                 Custom index.html file [$TTY2WEB_INDEX]
   --title-format value          Title format of browser window (default: "{{ .command }}@{{ .hostname }}") [$TTY2WEB_TITLE_FORMAT]
   --listen value                Listen for reverse connection (ex. 0.0.0.0:4444) [$TTY2WEB_LISTEN]
   --listencert value            Certificate and key for listen server (ex. mycert) [$TTY2WEB_LISTENCERT]
   --server value                Server for forwarding reverse connections (ex. 127.0.0.1:6000) (default: "127.0.0.1:6000") [$TTY2WEB_SERVER]
   --password value              Password for reverse server connection [$TTY2WEB_PASSWORD]
   --connect value               Connect to host for reverse connection (ex. 192.168.1.1:4444) [$TTY2WEB_CONNECT]
   --proxy value                 Use proxy for reverse server connection (ex. 192.168.1.1:8080) [$TTY2WEB_PROXY]
   --proxyauth value             Use proxy authentication for reverse server connection (ex. DOMAIN/user:password) [$TTY2WEB_PROXYAUTH]
   --useragent value             Use user agent for reverse server connection (ex. Mozilla) [$TTY2WEB_USERAGENT]
   --reconnect                   Enable reconnection [$TTY2WEB_RECONNECT]
   --reconnect-time value        Time to reconnect (default: 10) [$TTY2WEB_RECONNECT_TIME]
   --max-connection value        Maximum connection to tty2web (default: 0) [$TTY2WEB_MAX_CONNECTION]
   --once                        Accept only one client and exit on disconnection [$TTY2WEB_ONCE]
   --timeout value               Timeout seconds for waiting a client(0 to disable) (default: 0) [$TTY2WEB_TIMEOUT]
   --permit-arguments            Permit clients to send command line arguments in URL (e.g. http://example.com:8080/?arg=AAA&arg=BBB) [$TTY2WEB_PERMIT_ARGUMENTS]
   --width value                 Static width of the screen, 0(default) means dynamically resize (default: 0) [$TTY2WEB_WIDTH]
   --height value                Static height of the screen, 0(default) means dynamically resize (default: 0) [$TTY2WEB_HEIGHT]
   --ws-origin value             A regular expression that matches origin URLs to be accepted by WebSocket. No cross origin requests are acceptable by default [$TTY2WEB_WS_ORIGIN]
   --term value                  Terminal name to use on the browser, one of xterm or hterm. (default: "xterm") [$TTY2WEB_TERM]
   --close-signal value          Signal sent to the command process when tty2web close it (default: SIGHUP) (default: 1) [$TTY2WEB_CLOSE_SIGNAL]
   --close-timeout value         Time in seconds to force kill process after client is disconnected (default: -1) (default: -1) [$TTY2WEB_CLOSE_TIMEOUT]
   --config value                Config file path (default: "~/.tty2web") [$TTY2WEB_CONFIG]
   --version, -v                 print the version
```

### Config File

You can customize default options and your terminal (hterm) by providing a config file to the `tty2web` command. tty2web loads a profile file at `~/.tty2web` by default when it exists.

```
// Listen at port 9000 by default
port = "9000"

// Enable TSL/SSL by default
enable_tls = true

// hterm preferences
// Smaller font and a little bit bluer background color
preferences {
    font_size = 5
    background_color = "rgb(16, 16, 32)"
}
```

See the [`.tty2web`](https://github.com/kost/tty2web/blob/master/.tty2web) file in this repository for the list of configuration options.

### Security Options

By default, tty2web doesn't allow clients to send any keystrokes or commands except terminal window resizing. When you want to permit clients to write input to the TTY, add the `-w` option. However, accepting input from remote clients is dangerous for most commands. When you need interaction with the TTY for some reasons, consider starting tty2web with tmux or GNU Screen and run your command on it (see "Sharing with Multiple Clients" section for detail).

To restrict client access, you can use the `-c` option to enable the basic authentication. With this option, clients need to input the specified username and password to connect to the tty2web server. Note that the credentical will be transmitted between the server and clients in plain text. For more strict authentication, consider the SSL/TLS client certificate authentication described below.

The `-r` option is a little bit casualer way to restrict access. With this option, tty2web generates a random URL so that only people who know the URL can get access to the server.

All traffic between the server and clients are NOT encrypted by default. When you send secret information through tty2web, we strongly recommend you use the `-t` option which enables TLS/SSL on the session. By default, tty2web loads the crt and key files placed at `~/.tty2web.crt` and `~/.tty2web.key`. You can overwrite these file paths with the `--tls-crt` and `--tls-key` options. When you need to generate a self-signed certification file, you can use the `openssl` command.

```sh
openssl req -x509 -nodes -days 9999 -newkey rsa:2048 -keyout ~/.tty2web.key -out ~/.tty2web.crt
```

(NOTE: For Safari uses, see [how to enable self-signed certificates for WebSockets](http://blog.marcon.me/post/24874118286/secure-websockets-safari) when use self-signed certificates)

For additional security, you can use the SSL/TLS client certificate authentication by providing a CA certificate file to the `--tls-ca-crt` option (this option requires the `-t` or `--tls` to be set). This option requires all clients to send valid client certificates that are signed by the specified certification authority.

## Sharing with Multiple Clients

tty2web starts a new process with the given command when a new client connects to the server. This means users cannot share a single terminal with others by default. However, you can use terminal multiplexers for sharing a single process with multiple clients.

For example, you can start a new tmux session named `tty2web` with `top` command by the command below.

```sh
$ tty2web tmux new -A -s tty2web top
```

This command doesn't allow clients to send keystrokes, however, you can attach the session from your local terminal and run operations like switching the mode of the `top` command. To connect to the tmux session from your terminal, you can use following command.

```sh
$ tmux new -A -s tty2web
```

By using terminal multiplexers, you can have the control of your terminal and allow clients to just see your screen.

### Quick Sharing on tmux

To share your current session with others by a shortcut key, you can add a line like below to your `.tmux.conf`.

```
# Start tty2web in a new window with C-t
bind-key C-t new-window "tty2web tmux attach -t `tmux display -p '#S'`"
```

### Screen

Install screen:
```
apt-get install screen
```

Start a new session with `screen -S session-name` and connect to it with tty2web in another terminal window/tab through `screen -x session-name`.
All commands and activities being done in the first terminal tab/window will now be broadcasted by tty2web.

## Playing with Docker

When you want to create a jailed environment for each client, you can use Docker containers like following:

```sh
$ tty2web -w docker run -it --rm busybox
```

## Development

You can build a binary using the following commands. There is basic Windows support, but it is limited. go1.9 is required.

```sh
# Install tools
go get github.com/jteeuwen/go-bindata/...
go get github.com/tools/godep

# Build
make
```

To build the frontend part (JS files and other static files), you need `npm`.

## Windows support

There is limited Windows support because there is no proper pty support on Windows. Currently, only listening (bind) mode works and you will have most luck with specific programs. In testing, cmd.exe did not work, but powershell.exe works:

```DOS .bat
tty2web.exe -w powershell.exe
```

Some of the applications work better if you run it with winpty. Windows support is not my focus, but if you send me pull request, I would gladly accept it.

## Architecture

tty2web uses [xterm.js](https://xtermjs.org/) and [hterm](https://groups.google.com/a/chromium.org/forum/#!forum/chromium-hterm) to run a JavaScript based terminal on web browsers. tty2web itself provides a websocket server that simply relays output from the TTY to clients and receives input from clients and forwards it to the TTY. This hterm + websocket idea is inspired by [Wetty](https://github.com/krishnasrinivas/wetty).

## Alternatives

### Command line client

* [gotty-client](https://github.com/moul/gotty-client): If you want to connect to tty2web or GoTTY server from your terminal

### Terminal/SSH on Web Browsers

* [gotty](https://github.com/yudai/gotty): Original gotty on which tty2web is based
* [Secure Shell (Chrome App)](https://chrome.google.com/webstore/detail/secure-shell/pnhechapfaindjhompbnflcldabbghjo): If you are a chrome user and need a "real" SSH client on your web browser, perhaps the Secure Shell app is what you want
* [Wetty](https://github.com/krishnasrinivas/wetty): Node based web terminal (SSH/login)
* [ttyd](https://tsl0922.github.io/ttyd): C port of GoTTY with CJK and IME support

### Terminal Sharing

* [tmate](http://tmate.io/): Forked-Tmux based Terminal-Terminal sharing
* [termshare](https://termsha.re): Terminal-Terminal sharing through a HTTP server
* [tmux](https://tmux.github.io/): Tmux itself also supports TTY sharing through SSH)

# Credits

tty2web is based on [gotty](https://github.com/yudai/gotty). To be specific, it is based on latest master branch at the time of the fork [commit on 13 Dec 2017](https://github.com/yudai/gotty/commit/a080c85cbc59226c94c6941ad8c395232d72d517)

# License

The MIT License

