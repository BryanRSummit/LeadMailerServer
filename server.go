package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BryanRSummit/LeadMailerServer/templates"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Define a struct to hold our rate limiter and last hit times
type IPRateLimiter struct {
	ips   *cache.Cache
	mu    sync.Mutex
	rate  rate.Limit
	burst int
}

// Create a new rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:   cache.New(2*time.Minute, 5*time.Minute),
		rate:  r,
		burst: b,
	}

	return i
}

// Get and create limiter for an IP address if it doesn't exist
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.ips.Get(ip)
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		i.ips.Set(ip, limiter, cache.DefaultExpiration)
		return limiter
	}

	limiter := v.(*rate.Limiter)
	return limiter
}

// Create a new rate limiter allowing 5 requests per second with a burst of 10
var limiter = NewIPRateLimiter(5, 10)

// Middleware function to handle rate limiting
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := limiter.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

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

	//--------SHEETS_CREDS PROD----------------------------------------------------------
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
	//--------END SHEETS_CREDS PROD----------------------------------------------------------

	// //---------SHEETS CREDS LOCAL---------------------------------------------------------------
	// credBytes, err := os.ReadFile("agentcontactcount-01c64e5317e2.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read credentials file: %v", err)
	// }

	// config, err := google.JWTConfigFromJSON(credBytes, sheets.SpreadsheetsScope)
	// if err != nil {
	// 	log.Fatalf("Unable to parse credentials: %v", err)
	// }
	// //---------END SHEETS_CREDS LOCAL-----------------------------------------------------------

	ctx := context.Background()
	client := config.Client(ctx)

	sheetsService, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	// //-------LOCAL ENV - COMMENT OUT FOR PROD----------------------------------------
	// // Load .env.local file
	// if err := godotenv.Load(".env.local"); err != nil {
	// 	log.Println("Error loading .env.local file. Falling back to .env")
	// 	// Attempt to load .env file as fallback
	// 	if err := godotenv.Load(); err != nil {
	// 		log.Println("Error loading .env file. Using system environment variables.")
	// 	}
	// }
	// //-------LOCAL ENV - COMMENT OUT FOR PROD----------------------------------------

	//auth config
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("CALLBACK_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
} // end init

func loginHandler(w http.ResponseWriter, r *http.Request) {
	leadID := r.URL.Query().Get("lead_id")
	if leadID == "" {
		missingIdHTML := templates.GetMissingIdMessage()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, missingIdHTML)
		return
	}

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
		unauthorizedHTML := templates.GetUnauthorizedEmailDomain()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, unauthorizedHTML)
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
		missingIdHTML := templates.GetMissingIdMessage()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, missingIdHTML)
		return
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	session, _ := store.Get(r, "auth-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		unauthenticatedHTML := templates.GetUnauthenticatedHTML(leadID)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, unauthenticatedHTML)
		return
	}

	if r.Method == "GET" {
		leadID := r.URL.Query().Get("lead_id")
		if leadID == "" {
			missingIdHTML := templates.GetMissingIdMessage()
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, missingIdHTML)
			return
		}

		confirmHTML := templates.GetConfirmHTML(leadID)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, confirmHTML)

	} else if r.Method == "POST" {
		leadID := r.URL.Query().Get("lead_id")
		postReqHTML := templates.GetPostRequestHTML(leadID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, postReqHTML)
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
			missingIdHTML := templates.GetMissingIdMessage()
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, missingIdHTML)
			return
		}
	}

	err := updateLeadInSheet(leadID)
	if err != nil {
		http.Error(w, fmt.Sprintf("An error occurred: %v", err), http.StatusInternalServerError)
		return
	}

	leadUpdatedHTML := templates.GetLeadUpdatedHTML(leadID)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, leadUpdatedHTML)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		alreadyLoggedOutHTML := templates.GetAlreadyLoggedOut()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, alreadyLoggedOutHTML)
		return
	}

	//clear session
	session.Options.MaxAge = -1                        // Set MaxAge to -1 to delete the cookie
	session.Values = make(map[interface{}]interface{}) // Clear all session values

	// Ensure cookie is secure and HTTP-only
	session.Options.Secure = true // Only use this if your site is HTTPS
	session.Options.HttpOnly = true

	// Save session (which will delete the cookie)
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// loggedOutHTML := templates.GetLoggedOutMessage()
	// w.Header().Set("Content-Type", "text/html")
	// w.WriteHeader(http.StatusOK)
	// fmt.Fprint(w, loggedOutHTML)

	// Redirect to login page or home page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func dateSelectionHandler(w http.ResponseWriter, r *http.Request) {
	leadID := r.URL.Query().Get("lead_id")
	if leadID == "" {
		http.Error(w, "Missing lead ID", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "auth-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, fmt.Sprintf("/login?lead_id=%s", leadID), http.StatusTemporaryRedirect)
		//http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dateSelectionHTML := templates.GetDateSelectHTML(leadID)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, dateSelectionHTML)
}

func updateDateInSheet(leadID string, date string) error {
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

	// Update the 20th column (T) with the date
	updateColumn := "T"
	updateRange := fmt.Sprintf("%s!%s%d", sheetName, updateColumn, rowIndex+1)
	values := [][]interface{}{
		{date},
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

func submitDateHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data struct {
		Date   string `json:"date"`
		LeadID string `json:"leadId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := updateDateInSheet(data.LeadID, data.Date)
	if err != nil {
		http.Error(w, fmt.Sprintf("An error occurred: %v", err), http.StatusInternalServerError)
		return
	}

	dateChangedHTML := templates.GetDateChangedHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, dateChangedHTML)
}

func main() {
	// http.HandleFunc("/", handleLeadMailer)
	// http.HandleFunc("/update-lead", updateLeadHandler)
	// http.HandleFunc("/login", loginHandler)
	// http.HandleFunc("/auth/google/callback", callbackHandler)
	// http.HandleFunc("/logout", logoutHandler)
	// http.HandleFunc("/select-date", dateSelectionHandler)
	// http.HandleFunc("/submit-date", submitDateHandler)

	// Apply rate limiting middleware to your handlers
	http.HandleFunc("/", rateLimitMiddleware(handleLeadMailer))
	http.HandleFunc("/update-lead", rateLimitMiddleware(updateLeadHandler))
	http.HandleFunc("/login", rateLimitMiddleware(loginHandler))
	http.HandleFunc("/auth/google/callback", rateLimitMiddleware(callbackHandler))
	http.HandleFunc("/logout", rateLimitMiddleware(logoutHandler))
	http.HandleFunc("/select-date", rateLimitMiddleware(dateSelectionHandler))
	http.HandleFunc("/submit-date", rateLimitMiddleware(submitDateHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
