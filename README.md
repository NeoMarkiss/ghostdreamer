Ghostdreamer v0.1.0 BETA - Standalone Embedding & Build Architecture

1. Is there a separate index.html file required?

No. Ghostdreamer operates on a strict Single-File Mandate for source organization during distribution, while remaining modular for local development:

Embedded Production Mode: Inside main.go, the frontend interface is stored as a raw multiline string constant (htmlTemplate) and parsed via template.Must(template.New("index").Parse(htmlTemplate)). When you compile with go build, Go's //go:embed or string compilation embeds the entire HTML, CSS, and PWA canvas directly into the binary.

Zero Loose Files: End-users downloading your binary (ghostdreamerv1b.exe or ghostdreamerv1b-linux) do not need to carry a separate index.html file alongside it. The executable is entirely self-contained.

2. GitHub Repository Build Specs & Optional Flags

When publishing or compiling your Go binary from your repository source, you can use these optional build flags depending on your target environment:

Standard Build (Default with Terminal Output)

go build -o ghostdreamerv1b main.go


Production Release Build (Windows GUI Mode + Stripped Symbols)

Suppresses the background console command prompt window on Windows and strips debugging symbols to minimize file size:

# For Windows
go build -ldflags="-H windowsgui -s -w" -o ghostdreamerv1b.exe main.go

# For Linux Server / Edge Nodes
go build -ldflags="-s -w" -o ghostdreamerv1b main.go


3. Local Runtime & PWA Verification

Run your compiled binary or execute go run main.go.

Open Microsoft Edge or Google Chrome and navigate to http://127.0.0.1:8080.

Click the browser menu (three dots) and select "Install Ghostdreamer as an app" to run your streaming portal in its own dedicated PWA window container.