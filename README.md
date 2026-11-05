# Ghostdreamer (v0.1.0-beta)

A self-contained, high-performance Go application engineered with `go:embed` to bundle the frontend `index.html` and script assets directly into the compiled binary execution stream.

## Architectural Rationale: Why Go Embed Over External Assets?

In traditional Go web architectures, serving frontend interfaces requires either maintaining external static directories (`/public`, `/static`) alongside the binary or embedding massive multi-line HTML/JS payloads inside Go source files using raw string backticks (` `` `). 

Ghostdreamer utilizes `go:embed` to bypass these limitations through a native single-file build workflow:

1. **Elimination of Backtick & Line-Limit Friction:** Storing complex, scripted HTML/JS interfaces inside Go source files via raw string literals quickly hits readability walls, syntax highlighting breakdown, and string-escaping overhead. Embedding native `.html` files keeps markup in its native domain while satisfying strict static analysis and code validation tolerances.
2. **Zero Runtime Path Drift:** Non-embedded deployments introduce fragility where production binaries fail to locate relative asset paths or static folders upon execution. Embedding the frontend guarantees atomic version synchronization—the binary and the interface are inextricably bound to the exact same commit hash.
3. **Atomic Single-File Portability:** Eliminates the need for auxiliary zip distributions, folder structures, or web server reverse proxies. A single executable (`ghostdreamerv1b.exe` or `ghostdreamerv1b-linux`) spins up both the REST/RTMP routing layer and the complete interactive UI locally with zero external dependencies.

---

## Quick Start & Verification

### 1. Download Pre-Built Releases
Grab the latest stable binary for your platform from the [GitHub Releases](../../releases) page:
* **Windows:** `ghostdreamerv1b.exe`
* **Linux:** `ghostdreamerv1b-linux`

### 2. Execute Locally
Run the standalone binary in your environment:
```powershell
./ghostdreamerv1b.exe
