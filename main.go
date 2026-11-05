/*
 * Go Streaming Window Portal Boilerplate - Ghostdreamer v0.1.0 BETA
 * Features:
 * - Serves an embedded, responsive PWA-ready HTML/JS GUI from index.html using go:embed.
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
	"embed"
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

//go:embed index.html
var contentFiles embed.FS

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
	tmpl, err := template.ParseFS(contentFiles, "index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}
	globalState.RLock()
	defer globalState.RUnlock()
	tmpl.Execute(w, globalState)
}