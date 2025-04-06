package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// SnipeRequest defines the structure for incoming snipe requests
type SnipeRequest struct {
	Username string `json:"username"`
	Delay    int    `json:"delay"` 
}

// ConfigRequest defines the structure for incoming config save requests
type ConfigRequest struct {
	GCAccounts string `json:"gcAccounts"`
	GPAccounts string `json:"gpAccounts"`
	MSAccounts string `json:"msAccounts"`
}

// ConfigResponse defines the structure for returning loaded config
type ConfigResponse struct {
	GCAccounts string `json:"gcAccounts"`
	GPAccounts string `json:"gpAccounts"`
	MSAccounts string `json:"msAccounts"`
}

// Helper function to read config data from a file
func readConfigFile(filename string) string {
	// Files are in the parent directory relative to where the server runs (web/)
	fpath := filepath.Join("..", filename)
	data, err := os.ReadFile(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Info: Config file '%s' not found. Creating.", filename)
			// Create the file if it doesn't exist
			emptyData := []byte("")
			if writeErr := os.WriteFile(fpath, emptyData, 0644); writeErr != nil {
				log.Printf("Error: Failed to create config file '%s': %v", filename, writeErr)
			}
		} else {
			log.Printf("Warning: Could not read config file '%s': %v", filename, err)
		}
		return ""
	}
	return string(data)
}

// Helper function to write config data to a file
func writeConfigFile(filename string, data string) error {
	fpath := filepath.Join("..", filename)
	return os.WriteFile(fpath, []byte(data), 0644)
}

// handleConfigSave handles saving account configurations
func handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req ConfigRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding config request body: %v", err)
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	// Write each account type to its file
	err = writeConfigFile("gc.txt", req.GCAccounts)
	if err != nil {
		log.Printf("Error writing gc.txt: %v", err)
		http.Error(w, fmt.Sprintf("Failed to write gc.txt: %v", err), http.StatusInternalServerError)
		return
	}

	err = writeConfigFile("gp.txt", req.GPAccounts)
	if err != nil {
		log.Printf("Error writing gp.txt: %v", err)
		http.Error(w, fmt.Sprintf("Failed to write gp.txt: %v", err), http.StatusInternalServerError)
		return
	}

	err = writeConfigFile("ms.txt", req.MSAccounts)
	if err != nil {
		log.Printf("Error writing ms.txt: %v", err)
		http.Error(w, fmt.Sprintf("Failed to write ms.txt: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully updated account configuration files.")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account configurations saved successfully!",
	})
}

// handleConfigLoad handles loading current account configurations
func handleConfigLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	
	resp := ConfigResponse{
		GCAccounts: readConfigFile("gc.txt"),
		GPAccounts: readConfigFile("gp.txt"),
		MSAccounts: readConfigFile("ms.txt"),
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding config load response: %v", err)
	}
}

func handleSnipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SnipeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf("Received snipe request for username: %s (Delay: %dms - currently ignored)", req.Username, req.Delay)

	// In this simplified version, we'll execute the CLI directly
	// This assumes the MCsniperGO executable is already built
	cmd := exec.Command("../mcsnipergo", "-u", req.Username)
	cmd.Dir = ".." // Run from project root
	
	// We'll just collect standard output and error for logging
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running MCsniperGO: %v", err)
		errorMsg := fmt.Sprintf("Error running MCsniperGO: %v\nOutput: %s", err, string(outputBytes))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": errorMsg,
		})
		return
	}
	
	log.Printf("MCsniperGO Output: %s", string(outputBytes))

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Snipe request processed for '%s'. Check server logs for output.", req.Username),
	})
}

func main() {
	// Serve static files from dist directory
	fs := http.FileServer(http.Dir("dist"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/snipe", handleSnipe)
	http.HandleFunc("/api/config/save", handleConfigSave)
	http.HandleFunc("/api/config/load", handleConfigLoad)

	log.Println("Starting web server on http://localhost:8080")
	log.Println("To use the sniper functionality, make sure MCsniperGO is built in the parent directory.")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
} 