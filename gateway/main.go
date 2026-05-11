package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type ProxyTarget struct {
	Name string
	URL  string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Health check requested from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	resp := HealthResponse{
		Service:   "gateway",
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

func proxyHealthHandler(target ProxyTarget) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Proxying health check to %s at %s", target.Name, target.URL)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(target.URL + "/health")
		if err != nil {
			log.Printf("[ERROR] Failed to reach %s: %v", target.Name, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   fmt.Sprintf("cannot reach %s", target.Name),
				"details": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		json.NewEncoder(w).Encode(body)
	}
}

func statusHandler(targets []ProxyTarget) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Aggregated status requested from %s", r.RemoteAddr)
		client := &http.Client{Timeout: 5 * time.Second}
		results := make(map[string]interface{})

		results["gateway"] = HealthResponse{
			Service:   "gateway",
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		for _, t := range targets {
			resp, err := client.Get(t.URL + "/health")
			if err != nil {
				log.Printf("[WARN] %s unreachable: %v", t.Name, err)
				results[t.Name] = map[string]string{"status": "unreachable", "error": err.Error()}
				continue
			}
			defer resp.Body.Close()
			var body map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&body)
			results[t.Name] = body
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func main() {
	port := getEnv("GATEWAY_PORT", "8080")
	monitorURL := getEnv("MONITOR_URL", "http://localhost:8081")
	dashboardURL := getEnv("DASHBOARD_URL", "http://localhost:8082")

	targets := []ProxyTarget{
		{Name: "monitor", URL: monitorURL},
		{Name: "dashboard", URL: dashboardURL},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/monitor/health", proxyHealthHandler(targets[0]))
	mux.HandleFunc("/api/dashboard/health", proxyHealthHandler(targets[1]))
	mux.HandleFunc("/api/status", statusHandler(targets))

	log.Printf("[INFO] Gateway starting on port %s", port)
	log.Printf("[INFO] Monitor URL: %s", monitorURL)
	log.Printf("[INFO] Dashboard URL: %s", dashboardURL)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] Gateway failed to start: %v", err)
	}
}
