package webserver

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time" // Will likely be needed for drop range

	// Adjust these imports based on actual MCsniperGO package structure
	"github.com/Kqzz/MCsniperGO/claimer"
	cliutils "github.com/Kqzz/MCsniperGO/cmd/cli" // Import the package containing getAccounts
	mc "github.com/Kqzz/MCsniperGO/pkg/mc"
	"github.com/Kqzz/MCsniperGO/pkg/parser"
)

//go:embed all:../../web/dist
var embeddedFiles embed.FS

// SnipeRequest defines the structure for incoming snipe requests
type SnipeRequest struct {
	Username string `json:"username"`
	Delay    int    `json:"delay"` // Still not directly used by core logic shown previously
}

// ConfigRequest defines the structure for incoming config save requests
type ConfigRequest struct {
	GCAccounts string `json:"gcAccounts"`
	GPAccounts string `json:"gpAccounts"`
	MSAccounts string `json:"msAccounts"`
	// TODO: Add Proxies field later
}

// ConfigResponse defines the structure for returning loaded config
type ConfigResponse struct {
	GCAccounts string `json:"gcAccounts"`
	GPAccounts string `json:"gpAccounts"`
	MSAccounts string `json:"msAccounts"`
	// TODO: Add Proxies field later
}

// --- Helper Functions ---

// Reads config relative to executable's CWD (project root)
func readConfigFile(filename string) string {
	// No longer need filepath.Join("..", filename) as executable runs from root
	data, err := os.ReadFile(filename) 
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Info: Config file '%s' not found. Creating.", filename)
			// Create the file if it doesn't exist
            emptyData := []byte("")
            if writeErr := os.WriteFile(filename, emptyData, 0644); writeErr != nil {
                 log.Printf("Error: Failed to create config file '%s': %v", filename, writeErr)
            }
		} else {
			log.Printf("Warning: Could not read config file '%s': %v", filename, err)
		}
		return ""
	}
	return string(data)
}

// Writes config relative to executable's CWD (project root)
func writeConfigFile(filename string, data string) error {
	// No longer need filepath.Join("..", filename)
	return os.WriteFile(filename, []byte(data), 0644)
}

// Simplified version of getAccounts from cmd/cli/util.go
// Assumes config files are in the current working directory (project root)
func getAccounts() ([]*mc.MCaccount, error) {
	giftCodeLines, _ := parser.ReadLines("gc.txt")
	gamepassLines, _ := parser.ReadLines("gp.txt")
	microsoftLines, _ := parser.ReadLines("ms.txt")

	gcs, gcParseErrors := parser.ParseAccounts(giftCodeLines, mc.MsPr)
	logErrors(gcParseErrors)

	microsofts, msParseErrors := parser.ParseAccounts(microsoftLines, mc.Ms)
	logErrors(msParseErrors)
	
	gamepasses, gpParseErrors := parser.ParseAccounts(gamepassLines, mc.MsGp)
	logErrors(gpParseErrors)

	accounts := append(gcs, microsofts...)
	accounts = append(accounts, gamepasses...)

	if len(accounts) == 0 {
		return accounts, fmt.Errorf("no accounts found or parsed successfully in gc.txt, ms.txt, gp.txt")
	}
	return accounts, nil
}

func logErrors(errors []error) {
	for _, err := range errors {
		if err != nil {
			log.Printf("Config Parse Error: %v", err) // Use server logger
		}
	}
}

// --- HTTP Handlers ---

func handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	// ... (rest of the save logic is largely the same, using the updated writeConfigFile)
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


func handleConfigLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
    // ... (rest of the load logic is largely the same, using the updated readConfigFile)
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

	// --- Direct Call Logic ---
	go func() { // Run snipe logic in a goroutine to avoid blocking the HTTP response
		username := req.Username
		// TODO: Properly get/calculate dropRange. Using placeholder for now.
		// Need to import and potentially adapt log.GetDropRange() or use time.Now() logic
		dropTime := time.Now().Add(1 * time.Second) // Example: Target 1 second from now
		dropRange := mc.DropRange{Start: dropTime, End: dropTime.Add(500 * time.Millisecond)} // Example range
		
		log.Printf("Loading accounts for snipe...")
		// Use getAccounts from the cliutils package
		accounts, accErr := cliutils.GetAccounts("gc.txt", "gp.txt", "ms.txt") 
		if accErr != nil {
			log.Printf("Snipe Failed for %s: Could not load accounts: %v", username, accErr)
			// TODO: Communicate this failure back to the user (e.g., WebSocket/SSE)
			return
		}
		log.Printf("Found %d accounts.", len(accounts))

		log.Printf("Loading proxies for snipe...")
		// TODO: Handle proxies properly (load from proxies.txt or web UI config)
		proxies, proxyErr := parser.ReadLines("proxies.txt")
		if proxyErr != nil {
			log.Printf("Warning: Could not load proxies.txt: %v. Proceeding without proxies.", proxyErr)
			proxies = []string{} // Ensure it's an empty slice, not nil
		} else {
			log.Printf("Found %d proxies.", len(proxies))
		}


		log.Printf("Starting snipe for %s at ~%s...", username, dropRange.Start.Format(time.RFC3339))

		// Call the core claimer function directly
		claimErr := claimer.ClaimWithinRange(username, dropRange, accounts, proxies)

		if claimErr != nil {
			log.Printf("Snipe completed for %s with error: %v", username, claimErr)
			// TODO: Communicate this failure back to the user
		} else {
			log.Printf("Snipe completed successfully for %s (or claim attempt finished). Check logs for claim status.", username)
			// TODO: Communicate success/completion back to the user
		}
	}()

	// Respond immediately that the snipe process has been initiated
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Snipe process initiated for username '%s'. Check server logs for progress and results.", req.Username),
	})
}


// StartWebServer starts the integrated web server
func StartWebServer(port string) {
	mux := http.NewServeMux()

	// Create a sub-filesystem rooted at "dist" within the embedded files
	distFS, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	
	// Serve static files from the embedded filesystem
	// Handle index.html explicitly
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" || r.URL.Path == "" {
             r.URL.Path = "/index.html" // Serve index.html for root path
        } else if _, err := distFS.Open(r.URL.Path[1:]); err != nil {
			 // If file doesn't exist in distFS, serve index.html (for SPA routing, if needed later)
			 // Or return 404 if preferred: http.NotFound(w, r); return
             // For now, just let FileServer handle it (will likely 404 if not found)
		}

        // FileServer needs to be created here to use the modified path
		fs := http.FileServer(http.FS(distFS))
        fs.ServeHTTP(w, r)
    })


	// API Handlers
	mux.HandleFunc("/api/snipe", handleSnipe)
	mux.HandleFunc("/api/config/save", handleConfigSave)
	mux.HandleFunc("/api/config/load", handleConfigLoad)

	log.Printf("Starting integrated web server on http://localhost%s", port)
	err = http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
} 