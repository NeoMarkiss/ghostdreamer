/*
 * Go Streaming Window Portal Boilerplate - Ghostdreamer v0.1.0 BETA
 * Features:
 * - Serves an embedded, responsive PWA-ready HTML/JS GUI.
 * - Dynamic PWA Manifest endpoint (/manifest.json).
 * - Self-healing disk manifest generator if local manifest.json is 0 or 2 bytes.
 * - Real-time state management (VLC RTMP cross-connect, YouTube profile history log, private access toggle).
 * - RESTful JSON API endpoints for configuration updates.
 * - Built-in low-fidelity database history tracker.
 * Usage:
 * go run main.go
 * Then open http://127.0.0.1:8080 or install via Edge ("Install site as app").
 */

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AppState struct {
	sync.RWMutex
	VLCStreamURL      string    `json:"vlc_stream_url"`
	YouTubeStreamKey  string    `json:"youtube_stream_key"`
	PrivateFeedToggle bool      `json:"private_feed_toggle"`
	VisualizerPreset  string    `json:"visualizer_preset"`
	StreamStatus      string    `json:"stream_status"`
	TargetFPS         int       `json:"target_fps"`
	History           []History `json:"history"`
}

type History struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Profile   string    `json:"profile"`
	Metric    string    `json:"metric"`
	Passed    bool      `json:"passed"`
}

var globalState = &AppState{
	VLCStreamURL:      "rtmp://127.0.0.1:1935/live/vlc_input",
	YouTubeStreamKey:  "xxxx-xxxx-xxxx-xxxx",
	PrivateFeedToggle: true,
	VisualizerPreset:  "Cybernetic Audio Spectrum",
	StreamStatus:      "IDLE / READY",
	TargetFPS:         15,
	History: []History{
		{
			ID:        "hist-001",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Profile:   "Initial Edge PWA Handshake",
			Metric:    "12 FPS / Standard Sync",
			Passed:    true,
		},
	},
}

type Manifest struct {
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	StartURL        string `json:"start_url"`
	Display         string `json:"display"`
	BackgroundColor string `json:"background_color"`
	ThemeColor      string `json:"theme_color"`
}

func getManifestData() Manifest {
	return Manifest{
		Name:            "Ghostdreamer v0.1.0 BETA",
		ShortName:       "Ghostdreamer",
		StartURL:        "/?mode=standalone",
		Display:         "standalone",
		BackgroundColor: "#0f172a",
		ThemeColor:      "#3b82f6",
	}
}

func sanitizeDiskManifest() {
	fileName := "manifest.json"
	info, err := os.Stat(fileName)
	if err == nil && info.Size() <= 2 {
		log.Printf("[Fix] Local manifest.json size is invalid (%d bytes). Repairing...", info.Size())
		data, err := json.MarshalIndent(getManifestData(), "", "  ")
		if err == nil {
			_ = os.WriteFile(fileName, data, 0644)
			log.Println("[Fix] Successfully wrote valid JSON to manifest.json")
		}
	}
}

func main() {
	sanitizeDiskManifest()

	mux := http.NewServeMux()

	staticDir := filepath.Join(".", "static")
	if _, err := os.Stat(staticDir); err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	}

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/manifest.json", handleManifest)

	mux.HandleFunc("/api/state", handleGetState)
	mux.HandleFunc("/api/update", handleUpdateState)
	mux.HandleFunc("/api/history/add", handleAddHistory)

	desiredPort := "8080"
	listener, err := net.Listen("tcp", "127.0.0.1:"+desiredPort)
	if err != nil {
		log.Printf("[Warning] Port %s is busy. Asking OS for free port...", desiredPort)
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("Failed to bind network listener: %v", err)
		}
	}

	assignedAddr := listener.Addr().String()
	log.Printf("=========================================================")
	log.Printf("Ghostdreamer v0.1.0 BETA Initialized on http://%s", assignedAddr)
	log.Printf("Edge PWA Mode ready via manifest at http://%s/manifest.json", assignedAddr)
	log.Printf("Open Microsoft Edge and navigate to: http://%s", assignedAddr)
	log.Printf("=========================================================")

	if err := http.Serve(listener, mux); err != nil {
		log.Fatalf("Server launch failure: %v", err)
	}
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getManifestData())
}

func handleGetState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	globalState.RLock()
	defer globalState.RUnlock()
	json.NewEncoder(w).Encode(globalState)
}

func handleUpdateState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		VLCStreamURL      string `json:"vlc_stream_url"`
		YouTubeStreamKey  string `json:"youtube_stream_key"`
		PrivateFeedToggle bool   `json:"private_feed_toggle"`
		VisualizerPreset  string `json:"visualizer_preset"`
		TargetFPS         int    `json:"target_fps"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	globalState.Lock()
	globalState.VLCStreamURL = payload.VLCStreamURL
	globalState.YouTubeStreamKey = payload.YouTubeStreamKey
	globalState.PrivateFeedToggle = payload.PrivateFeedToggle
	globalState.VisualizerPreset = payload.VisualizerPreset
	if payload.TargetFPS > 0 {
		globalState.TargetFPS = payload.TargetFPS
	}
	globalState.StreamStatus = "CONFIGURED / PASS-THROUGH READY"
	globalState.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func handleAddHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Profile string `json:"profile"`
		Metric  string `json:"metric"`
		Passed  bool   `json:"passed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	globalState.Lock()
	newHist := History{
		ID:        fmt.Sprintf("hist-%03d", len(globalState.History)+1),
		Timestamp: time.Now(),
		Profile:   req.Profile,
		Metric:    req.Metric,
		Passed:    req.Passed,
	}
	globalState.History = append([]History{newHist}, globalState.History...)
	globalState.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newHist)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))
	globalState.RLock()
	defer globalState.RUnlock()
	tmpl.Execute(w, globalState)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Ghostdreamer v0.1.0 BETA</title>
    <link rel="manifest" href="/manifest.json">
    <style>
        :root {
            --bg-color: #0b0f19;
            --card-bg: #111827;
            --border: #1f2937;
            --text-main: #f3f4f6;
            --text-muted: #9ca3af;
            --accent: #3b82f6;
            --accent-hover: #2563eb;
            --success: #10b981;
            --warning: #f59e0b;
        }
        body {
            margin: 0;
            padding: 0;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-main);
            display: flex;
            flex-direction: column;
            height: 100vh;
            overflow: hidden;
        }
        header {
            background-color: var(--card-bg);
            border-bottom: 1px solid var(--border);
            padding: 12px 24px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        header h1 {
            margin: 0;
            font-size: 1.1rem;
            letter-spacing: 0.5px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .badge {
            background: var(--accent);
            color: #fff;
            font-size: 0.75rem;
            padding: 2px 8px;
            border-radius: 12px;
            font-weight: normal;
        }
        .main-container {
            display: grid;
            grid-template-columns: 1fr 340px;
            gap: 16px;
            padding: 16px;
            flex: 1;
            overflow: hidden;
        }
        .viewport-panel {
            background-color: #000;
            border-radius: 8px;
            border: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            position: relative;
            overflow: hidden;
        }
        canvas {
            width: 100%;
            height: 100%;
            display: block;
        }
        .overlay-controls {
            position: absolute;
            top: 12px;
            left: 12px;
            background: rgba(15, 23, 42, 0.85);
            backdrop-filter: blur(8px);
            padding: 8px 16px;
            border-radius: 6px;
            border: 1px solid var(--border);
            font-size: 0.85rem;
        }
        .sidebar {
            display: flex;
            flex-direction: column;
            gap: 16px;
            overflow-y: auto;
        }
        .card {
            background-color: var(--card-bg);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 16px;
        }
        .card h2 {
            margin-top: 0;
            font-size: 0.95rem;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 12px;
        }
        .form-group {
            margin-bottom: 12px;
        }
        label {
            display: block;
            font-size: 0.8rem;
            color: var(--text-muted);
            margin-bottom: 4px;
        }
        input[type="text"], input[type="number"], select {
            width: 100%;
            padding: 8px 10px;
            background-color: var(--bg-color);
            border: 1px solid var(--border);
            border-radius: 4px;
            color: var(--text-main);
            box-sizing: border-box;
            font-size: 0.85rem;
        }
        .toggle-container {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-top: 8px;
        }
        .switch {
            position: relative;
            display: inline-block;
            width: 44px;
            height: 22px;
        }
        .switch input {
            opacity: 0;
            width: 0;
            height: 0;
        }
        .slider {
            position: absolute;
            cursor: pointer;
            top: 0; left: 0; right: 0; bottom: 0;
            background-color: #334155;
            transition: .3s;
            border-radius: 22px;
        }
        .slider:before {
            position: absolute;
            content: "";
            height: 16px;
            width: 16px;
            left: 3px;
            bottom: 3px;
            background-color: white;
            transition: .3s;
            border-radius: 50%;
        }
        input:checked + .slider {
            background-color: var(--success);
        }
        input:checked + .slider:before {
            transform: translateX(22px);
        }
        button {
            width: 100%;
            background-color: var(--accent);
            color: white;
            border: none;
            padding: 10px;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 600;
            font-size: 0.85rem;
            transition: background-color 0.2s;
        }
        button:hover {
            background-color: var(--accent-hover);
        }
        .history-list {
            list-style: none;
            padding: 0;
            margin: 0;
            max-height: 180px;
            overflow-y: auto;
            font-size: 0.8rem;
        }
        .history-item {
            padding: 8px 0;
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
        }
        .history-item:last-child {
            border-bottom: none;
        }
    </style>
</head>
<body>
    <header>
        <h1>Ghostdreamer <span class="badge">v0.1.0 BETA</span></h1>
        <div style="font-size: 0.85rem; color: var(--text-muted);">
            Status: <span id="statusDisplay" style="color: var(--success);">{{.StreamStatus}}</span>
        </div>
    </header>

    <div class="main-container">
        <div class="viewport-panel">
            <canvas id="visualizerCanvas"></canvas>
            <div class="overlay-controls">
                <div><strong>Source:</strong> VLC Local Capture Stream</div>
                <div><strong>Pass-Through:</strong> <span id="passThroughStatus">{{if .PrivateFeedToggle}}ENABLED (PRIVATE){{else}}PUBLIC INGEST{{end}}</span></div>
            </div>
        </div>

        <div class="sidebar">
            <div class="card">
                <h2>VLC / YouTube Cross-Connect</h2>
                <div class="form-group">
                    <label>VLC RTMP Ingest Endpoint</label>
                    <input type="text" id="vlcUrl" value="{{.VLCStreamURL}}">
                </div>
                <div class="form-group">
                    <label>YouTube Live Stream Key</label>
                    <input type="text" id="ytKey" value="{{.YouTubeStreamKey}}">
                </div>
                <div class="form-group">
                    <label>Visualizer Preset</label>
                    <select id="preset">
                        <option value="Cybernetic Audio Spectrum">Cybernetic Audio Spectrum</option>
                        <option value="Low-FPS Dream Wave Loop">Low-FPS Dream Wave Loop</option>
                        <option value="Minimalist Obsidian Monolith">Minimalist Obsidian Monolith</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Frame Rate Limit (FPS)</label>
                    <input type="number" id="fps" value="{{.TargetFPS}}" min="5" max="60">
                </div>
                <div class="toggle-container">
                    <span style="font-size: 0.85rem;">Private Access Toggle</span>
                    <label class="switch">
                        <input type="checkbox" id="privateToggle" {{if .PrivateFeedToggle}}checked{{end}}>
                        <span class="slider"></span>
                    </label>
                </div>
                <div style="margin-top: 16px;">
                    <button onclick="updateConfig()">Sync Configuration</button>
                </div>
            </div>

            <div class="card">
                <h2>Low-Fi Profile History</h2>
                <ul class="history-list" id="historyList">
                    {{range .History}}
                    <li class="history-item">
                        <div>
                            <strong>{{.Profile}}</strong><br>
                            <span style="color: var(--text-muted);">{{.Metric}}</span>
                        </div>
                        <div style="color: {{if .Passed}}var(--success){{else}}var(--warning){{end}};">
                            {{if .Passed}}PASSED{{else}}TESTING{{end}}
                        </div>
                    </li>
                    {{end}}
                </ul>
                <div style="margin-top: 12px;">
                    <button style="background: var(--card-bg); border: 1px solid var(--border);" onclick="logTestMetric()">+ Log Test Metric</button>
                </div>
            </div>
        </div>
    </div>

    <script>
        const canvas = document.getElementById('visualizerCanvas');
        const ctx = canvas.getContext('2d');
        let animationId;
        let fps = {{.TargetFPS}};

        function resizeCanvas() {
            canvas.width = canvas.parentElement.clientWidth;
            canvas.height = canvas.parentElement.clientHeight;
        }
        window.addEventListener('resize', resizeCanvas);
        resizeCanvas();

        let step = 0;
        function renderVisualizer() {
            ctx.fillStyle = '#090d16';
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            const numBars = 32;
            const barWidth = canvas.width / numBars;

            for (let i = 0; i < numBars; i++) {
                const barHeight = (Math.sin(step + i * 0.2) * 0.5 + 0.5) * (canvas.height * 0.6);
                const x = i * barWidth;
                const y = canvas.height - barHeight;

                const gradient = ctx.createLinearGradient(0, canvas.height, 0, 0);
                gradient.addColorStop(0, '#3b82f6');
                gradient.addColorStop(1, '#10b981');

                ctx.fillStyle = gradient;
                ctx.fillRect(x, y, barWidth - 4, barHeight);
            }

            step += 0.05;
            setTimeout(() => {
                animationId = requestAnimationFrame(renderVisualizer);
            }, 1000 / fps);
        }
        renderVisualizer();

        async function updateConfig() {
            const payload = {
                vlc_stream_url: document.getElementById('vlcUrl').value,
                youtube_stream_key: document.getElementById('ytKey').value,
                visualizer_preset: document.getElementById('preset').value,
                target_fps: parseInt(document.getElementById('fps').value, 10),
                private_feed_toggle: document.getElementById('privateToggle').checked
            };

            fps = payload.target_fps;

            const res = await fetch('/api/update', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                document.getElementById('statusDisplay').innerText = "CONFIGURED / PASS-THROUGH READY";
                document.getElementById('passThroughStatus').innerText = payload.private_feed_toggle ? "ENABLED (PRIVATE)" : "PUBLIC INGEST";
                alert("Configuration synchronized to Go backend.");
            }
        }

        async function logTestMetric() {
            const reqData = {
                profile: "VLC Pass-through Benchmark",
                metric: fps + " FPS / RTMP Ingest Sync",
                passed: true
            };

            const res = await fetch('/api/history/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(reqData)
            });

            if (res.ok) {
                const item = await res.json();
                const list = document.getElementById('historyList');
                const li = document.createElement('li');
                li.className = 'history-item';
                li.innerHTML = `
                    <div>
                        <strong>${item.profile}</strong><br>
                        <span style="color: var(--text-muted);">${item.metric}</span>
                    </div>
                    <div style="color: var(--success);">PASSED</div>
                `;
                list.insertBefore(li, list.firstChild);
            }
        }
    </script>
</body>
</html>`