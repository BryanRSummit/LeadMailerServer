package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"encoding/json"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	sheetsService *sheets.Service
	// USE FOR TESTING
	//spreadsheetID = "1QKE4Gm54fyNA-ArMmd3MJgMkXG0pPfGfKLiIxpcsMtg"
	// PRODUCTION
	spreadsheetID = "1HQfk8kbZuUdP7__nQ6CxQKw_BN9i-4P2so3iplLOU8U"
	sheetName     = "Sheet1"

	oauthConfig *oauth2.Config
	store       *sessions.CookieStore
)

func init() {

	// // // Load .env file
	// // if err := godotenv.Load(); err != nil {
	// // 	fmt.Println("Error loading .env file")
	// // }
	// Initialize Google Sheets API client
	credJSON := os.Getenv("SHEETS_CREDS")
	if credJSON == "" {
		log.Fatal("SHEETS_CREDS environment variable is not set")
	}

	credBytes := []byte(credJSON)

	config, err := google.JWTConfigFromJSON(credBytes, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	ctx := context.Background()
	client := config.Client(ctx)

	sheetsService, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	//auth config
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		//RedirectURL:  "http://localhost:8080/auth/google/callback", // Update this with your domain
		RedirectURL: os.Getenv("CALLBACK_URL"), // Update this with your domain
		Scopes:      []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:    google.Endpoint,
	}

	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	leadID := r.URL.Query().Get("lead_id")
	state := generateStateToken(leadID) // We'll implement this function
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateStateToken(leadID string) string {
	return base64.URLEncoding.EncodeToString([]byte(leadID))
}

func getLeadIDFromState(state string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	//needed to pass lead_id
	state := r.URL.Query().Get("state")
	leadID, err := getLeadIDFromState(state)
	if err != nil {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	if !strings.HasSuffix(userInfo.Email, "@reddsummit.com") {
		http.Error(w, "Unauthorized email domain", http.StatusUnauthorized)
		return
	}

	session, _ := store.Get(r, "auth-session")
	session.Values["authenticated"] = true
	session.Values["email"] = userInfo.Email
	session.Values["lead_id"] = leadID // Store the lead_id in the session
	session.Save(r, w)

	// Redirect to the update page with the lead_id
	http.Redirect(w, r, fmt.Sprintf("/?lead_id=%s", leadID), http.StatusTemporaryRedirect)
}

func handleLeadMailer(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	leadID := r.URL.Query().Get("lead_id")
	if leadID == "" {
		http.Error(w, "Missing lead_id", http.StatusBadRequest)
		return
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	session, _ := store.Get(r, "auth-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
            <html>
                <body>
                    <h1>Authentication Required</h1>
                    <p>Please <a href="/login?lead_id=%s">log in</a> with your @reddsummit.com email to continue.</p>
                </body>
            </html>
        `, leadID)
		return
	}

	if r.Method == "GET" {
		leadID := r.URL.Query().Get("lead_id")
		if leadID == "" {
			http.Error(w, "Missing lead_id", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<html>
				<body>
				<h1 style="font-size: 24px;">Confirm That You Wish to Give Up Lead!</h1>
				<p><b style="font-size: 20px;">This is a permanent Action! Lead: %s</b></p>
				<div>Identity Platform Quickstart</div>
				<div id="message">Loading...</div>
				<button 
					onclick="confirmUpdate('%s')" 
					style="font-size: 36px; padding: 20px 40px; border-radius: 10px; background-color: #007BFF; color: white; border: none; cursor: pointer;">
					Confirm
				</button>
					<script>
						function confirmUpdate(leadID) {
							// Make an AJAX request to the server to update the lead
							fetch('/update-lead?lead_id=' + leadID, {
								method: 'GET',
								headers: {
									'Content-Type': 'application/json'
								}
							})
							.then(response => response.json())
							.then(data => {
								// Find the button element
								const buttonElement = document.querySelector('button[onclick="confirmUpdate(\'' + leadID + '\')"]');

								if (buttonElement) {
									// Remove the button and show a success message
									const successMessage = document.createElement('p');
									successMessage.style.fontSize = '18px';
									successMessage.textContent = 'Lead ' + leadID + ' has been updated.';
									buttonElement.parentNode.replaceChild(successMessage, buttonElement);
								} else {
									// Button element not found, display a message without modifying the DOM
									alert('Lead ' + leadID + ' has been updated.');
								}
							})
							.catch(error => {
								alert('An error occurred: ' + error);
							});
						}
					</script>
				</body>
			</html>
		`, leadID, leadID)

	} else if r.Method == "POST" {
		leadID := r.URL.Query().Get("lead_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "POST request received, to update lead use GET %s"}`, leadID)
	} else {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func updateLeadInSheet(leadID string) error {
	// Get all values from the sheet
	readRange := fmt.Sprintf("%s!A1:Z", sheetName)
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	// Find the row with the matching lead_id
	rowIndex := -1
	for i, row := range resp.Values {
		for _, cell := range row {
			cellStr, ok := cell.(string)
			if !ok {
				continue // Skip non-string cells
			}
			if strings.Contains(cellStr, leadID) {
				rowIndex = i
				break
			}
		}
		if rowIndex != -1 {
			break
		}
	}

	if rowIndex == -1 {
		return fmt.Errorf("Lead ID %s not found", leadID)
	}

	// Update the 13th column (index 12 in zero-based indexing)
	updateColumn := "M"
	updateRange := fmt.Sprintf("%s!%s%d", sheetName, updateColumn, rowIndex+1)
	values := [][]interface{}{
		{"TRUE"},
	}
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err = sheetsService.Spreadsheets.Values.Update(spreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	if err != nil {
		return fmt.Errorf("unable to update sheet: %v", err)
	}

	return nil
}

// New endpoint to handle the AJAX request
func updateLeadHandler(w http.ResponseWriter, r *http.Request) {
	//passing on lead_id
	session, _ := store.Get(r, "auth-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	leadID := r.URL.Query().Get("lead_id")
	if leadID == "" {
		// If lead_id is not in the URL, try to get it from the session
		leadID, ok := session.Values["lead_id"].(string)
		if !ok || leadID == "" {
			http.Error(w, "Missing lead_id", http.StatusBadRequest)
			return
		}
	}

	err := updateLeadInSheet(leadID)
	if err != nil {
		http.Error(w, fmt.Sprintf("An error occurred: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Lead updated, emails will no longer be sent! %s"}`, leadID)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")
	session.Options.MaxAge = -1                        // Set MaxAge to -1 to delete the cookie
	session.Values = make(map[interface{}]interface{}) // Clear all session values
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "You have been logged out!"}`)
}

func main() {
	http.HandleFunc("/", handleLeadMailer)
	http.HandleFunc("/update-lead", updateLeadHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth/google/callback", callbackHandler)
	http.HandleFunc("/logout", logoutHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
