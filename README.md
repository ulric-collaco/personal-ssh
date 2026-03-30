# Personal SSH TUI on GCP VM (Sanitized Setup Guide)

This document captures the full setup for running a public SSH-based TUI portfolio on a Google Cloud VM, with a custom Go SSH wrapper that supports anonymous access and per-IP rate limiting.

All sensitive details are intentionally replaced with placeholders.

## Architecture

- Public visitors connect to custom SSH server on port `22`.
- Admin access uses system OpenSSH on port `2222`.
- Custom SSH server launches your TUI binary in a PTY.
- Connection attempts are logged for analytics.
- A systemd service keeps the SSH wrapper running.

## Placeholder Conventions

Replace these with your real values when deploying:

- `<vm-user>`: Linux username on the VM
- `<vm-ip>`: Public VM IP
- `<public-domain>`: DNS name for the public SSH endpoint
- `<repo-url>`: Git repository URL
- `<go-bin>`: Go binary path (for example `/usr/local/go/bin/go`)

## Files Created on VM

### 1) SSH Server Wrapper (Go)
- Path: `/home/<vm-user>/ssh-wrapper/ssh-server.go`
- Purpose: Custom SSH server for anonymous sessions
- Features:
  - `NoClientAuth` (passwordless)
  - per-IP rate limiting (10 connections/minute)
  - session handoff to TUI binary

### 2) Go Module Files
- Path: `/home/<vm-user>/ssh-wrapper/go.mod` and `/home/<vm-user>/ssh-wrapper/go.sum`
- Purpose: Dependency management for wrapper server

### 3) Compiled SSH Server Binary
- Path: `/home/<vm-user>/ssh-wrapper/ssh-server`
- Purpose: Executable run by systemd

### 4) Analytics Log
- Path: `/home/<vm-user>/ssh-wrapper/connections.log`
- Purpose: Store connection attempts and metadata
- Example format:
  - `TIMESTAMP | IP: x.x.x.x | Country: XXX | Region: XXX | City: XXX | ISP: XXX`

### 5) Analytics Script
- Path: `/home/<vm-user>/ssh-wrapper/stats.sh`
- Purpose: Render traffic stats (all-time, 24h, 30d, top locations, recent activity)

### 6) TUI Application Source
- Path: `/home/<vm-user>/personal-ssh/main.go`
- Purpose: Portfolio TUI app source code

### 7) Compiled TUI Binary
- Path: `/home/<vm-user>/personal-ssh/tui-app`
- Purpose: Program launched for each visitor SSH session

### 8) Deploy Script
- Path: `/home/<vm-user>/personal-ssh/deploy.sh`
- Purpose: Pull latest code, rebuild, and restart service

### 9) systemd Service
- Path: `/etc/systemd/system/personal-ssh.service`
- Purpose: Keep SSH wrapper alive and restart on failure

### 10) Sudoers Entry
- Path: `/etc/sudoers.d/<vm-user>`
- Purpose: Allow selected operational commands without password prompts

### 11) OpenSSH Daemon Config
- Path: `/etc/ssh/sshd_config`
- Purpose: Move admin SSH to port `2222`

### 12) Authorized Keys
- Path: `/home/<vm-user>/.ssh/authorized_keys`
- Purpose: Store admin public key(s)

## SSH Wrapper Code (Rate-Limited)

The following is the SSH server wrapper with per-IP rate limiting. It has been sanitized to remove personal paths.

```go
package main

import (
    "encoding/binary"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "os/exec"
    "sync"
    "time"

    "github.com/creack/pty"
    "golang.org/x/crypto/ssh"
)

// Rate limiter
type rateLimiter struct {
    mu          sync.Mutex
    connections map[string]*connInfo
    maxPerIP    int
    windowSize  time.Duration
}

type connInfo struct {
    count     int
    firstSeen time.Time
    lastSeen  time.Time
}

func newRateLimiter(maxPerIP int, windowSize time.Duration) *rateLimiter {
    rl := &rateLimiter{
        connections: make(map[string]*connInfo),
        maxPerIP:    maxPerIP,
        windowSize:  windowSize,
    }
    // Cleanup old entries every minute
    go rl.cleanup()
    return rl
}

func (rl *rateLimiter) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        for ip, info := range rl.connections {
            if now.Sub(info.lastSeen) > rl.windowSize {
                delete(rl.connections, ip)
            }
        }
        rl.mu.Unlock()
    }
}

func (rl *rateLimiter) allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    info, exists := rl.connections[ip]

    if !exists {
        rl.connections[ip] = &connInfo{
            count:     1,
            firstSeen: now,
            lastSeen:  now,
        }
        return true
    }

    // Reset if window expired
    if now.Sub(info.firstSeen) > rl.windowSize {
        info.count = 1
        info.firstSeen = now
        info.lastSeen = now
        return true
    }

    // Check limit
    if info.count >= rl.maxPerIP {
        log.Printf("Rate limit exceeded for IP: %s (%d connections in window)", ip, info.count)
        return false
    }

    info.count++
    info.lastSeen = now
    return true
}

var limiter *rateLimiter

func main() {
    // Allow 10 connections per IP per minute
    limiter = newRateLimiter(10, 1*time.Minute)

    config := &ssh.ServerConfig{
        NoClientAuth: true,
    }

    hostKeyBytes, err := os.ReadFile("/etc/ssh/ssh_host_rsa_key")
    if err != nil {
        log.Fatalf("Failed to read host key: %v", err)
    }

    hostKey, err := ssh.ParsePrivateKey(hostKeyBytes)
    if err != nil {
        log.Fatalf("Failed to parse host key: %v", err)
    }

    config.AddHostKey(hostKey)

    listener, err := net.Listen("tcp", "0.0.0.0:22")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Println("SSH Server listening on port 22 with rate limiting (10 conn/min per IP)")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Accept error: %v", err)
            continue
        }

        // Extract IP
        ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()

        // Rate limit check
        if !limiter.allow(ip) {
            log.Printf("Blocking connection from %s (rate limit)", ip)
            conn.Close()
            continue
        }

        go handleConnection(conn, config)
    }
}

func handleConnection(netConn net.Conn, config *ssh.ServerConfig) {
    sshConn, chans, reqs, err := ssh.NewServerConn(netConn, config)
    if err != nil {
        return
    }
    defer sshConn.Close()

    go ssh.DiscardRequests(reqs)

    for newChannel := range chans {
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            continue
        }

        channel, requests, err := newChannel.Accept()
        if err != nil {
            continue
        }

        go runTUI(channel, requests)
    }
}

func runTUI(channel ssh.Channel, requests <-chan *ssh.Request) {
    defer channel.Close()

    cmd := exec.Command("/home/<vm-user>/personal-ssh/tui-app")
    cmd.Env = append(os.Environ(),
        "TERM=xterm-256color",
    )

    f, err := pty.Start(cmd)
    if err != nil {
        channel.Write([]byte(fmt.Sprintf("Failed to start: %v\r\n", err)))
        return
    }
    defer f.Close()

    // Handle PTY requests
    go func() {
        for req := range requests {
            switch req.Type {
            case "shell":
                req.Reply(true, nil)
            case "pty-req":
                termLen := req.Payload[3]
                w, h := parseDims(req.Payload[termLen+4:])
                pty.Setsize(f, &pty.Winsize{Rows: uint16(h), Cols: uint16(w)})
                req.Reply(true, nil)
            case "window-change":
                w, h := parseDims(req.Payload)
                pty.Setsize(f, &pty.Winsize{Rows: uint16(h), Cols: uint16(w)})
            }
        }
    }()

    // Copy data both ways
    go io.Copy(f, channel)
    io.Copy(channel, f)
    cmd.Wait()
}

func parseDims(b []byte) (uint32, uint32) {
    if len(b) < 8 {
        return 80, 24
    }
    w := binary.BigEndian.Uint32(b)
    h := binary.BigEndian.Uint32(b[4:])
    return w, h
}
```

## Deploy Script Example

```bash
#!/usr/bin/env bash
set -euo pipefail

cd /home/<vm-user>/personal-ssh
git pull origin main
<go-bin> build -o tui-app .
sudo systemctl restart personal-ssh.service
sudo systemctl status personal-ssh.service --no-pager
```

## systemd Service Example

```ini
[Unit]
Description=Personal SSH TUI Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/<vm-user>/ssh-wrapper
ExecStart=/home/<vm-user>/ssh-wrapper/ssh-server
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## OpenSSH Config (Admin Port)

In `/etc/ssh/sshd_config`, set:

```sshconfig
Port 2222
```

Then reload:

```bash
sudo systemctl restart ssh
```

## Service Management Commands

```bash
# start / stop / restart / status
sudo systemctl start personal-ssh.service
sudo systemctl stop personal-ssh.service
sudo systemctl restart personal-ssh.service
sudo systemctl status personal-ssh.service

# enable / disable on boot
sudo systemctl enable personal-ssh.service
sudo systemctl disable personal-ssh.service

# logs
sudo journalctl -u personal-ssh.service -f
```

## Port Checks

```bash
sudo ss -tlnp | grep :22
sudo ss -tlnp | grep :2222
sudo ss -tlnp | grep -E ':(22|2222)'
```

## Analytics Commands

```bash
cd /home/<vm-user>/ssh-wrapper && ./stats.sh
cat /home/<vm-user>/ssh-wrapper/connections.log
wc -l /home/<vm-user>/ssh-wrapper/connections.log
awk '{print $5}' /home/<vm-user>/ssh-wrapper/connections.log | sort | uniq | wc -l
```

## Access Reference (Sanitized)

- Public visitor endpoint: `ssh <public-domain>`
- Admin endpoint: `ssh <vm-user>@<vm-ip> -p 2222`

## Security Notes

- Do not publish private keys, full user identities, raw hostnames, or static IPs in public docs.
- Prefer restricting sudoers to only required commands instead of broad `NOPASSWD: ALL`.
- Keep admin SSH on a non-default port and ideally firewall to known source IPs.
- Consider `fail2ban` and/or cloud firewall rate limits in addition to app-level limiting.
