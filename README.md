# Ghostdreamer (v0.1.0-beta)

A self-contained, high-performance Go application utilizing Go's `go:embed` package to bundle the frontend `index.html` interface directly into the executable binary.

## Architecture Overview

Ghostdreamer combines a lightweight HTTP server and real-time canvas stream router into a single static binary. 

* **`main.go`**: Core server, HTTP routing, API endpoints, and static file serving via `//go:embed index.html`.
* **`index.html`**: Embedded frontend dashboard featuring real-time telemetry, canvas rendering loops, and configuration state synchronization.

---

## Getting Started

### Option 1: Download Pre-Built Releases (Recommended)
Grab the latest stable binary for your platform directly from the [GitHub Releases](../../releases) page:
* **Windows (GUI/Target):** `ghostdreamerv1b.exe`
* **Linux:** `ghostdreamerv1b-linux`

Run the binary directly in your environment:
```powershell
./ghostdreamerv1b.exe
