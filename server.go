package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var firebaseApp *firebase.App

var (
	sheetsService *sheets.Service
	// USE FOR TESTING
	//spreadsheetID = "1QKE4Gm54fyNA-ArMmd3MJgMkXG0pPfGfKLiIxpcsMtg"
	// PRODUCTION
	spreadsheetID = "1HQfk8kbZuUdP7__nQ6CxQKw_BN9i-4P2so3iplLOU8U"
	sheetName     = "Sheet1"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		client, err := firebaseApp.Auth(ctx)
		if err != nil {
			http.Error(w, "Error getting Auth client", http.StatusInternalServerError)
			return
		}

		idToken := r.Header.Get("Authorization")
		if idToken == "" {
			http.Error(w, "No ID token provided", http.StatusUnauthorized)
			return
		}

		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			http.Error(w, "Invalid ID token", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(r.Context(), "userID", token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func init() {

	// // // Load .env file
	// // if err := godotenv.Load(); err != nil {
	// // 	fmt.Println("Error loading .env file")
	// // }
	// Initialize Google Sheets API client
	credJSON := os.Getenv("SHEETS_CREDS")
	// Use the sheetsCreds value in your code
	fmt.Println("SHEETS_CREDS:", credJSON)
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

	// Initialize Firebase
	opt := option.WithCredentialsFile("path/to/your/firebase-adminsdk.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	firebaseApp = app
}

func handleLeadMailer(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == "GET" {
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
					<head>
						<script src="https://www.gstatic.com/firebasejs/8.10.0/firebase-app.js"></script>
						<script src="https://www.gstatic.com/firebasejs/8.10.0/firebase-auth.js"></script>
					</head>
					<body>
						<h1 style="font-size: 24px;">Confirm That You Wish to Give Up Lead!</h1>
						<p><b style="font-size: 20px;">This is a permanent Action! Lead: %s</b></p>
						<div id="message">Loading...</div>
						<button id="signInButton" style="font-size: 24px; padding: 10px 20px; display: none;">Sign In with Google</button>
						<button id="confirmButton" 
							style="font-size: 36px; padding: 20px 40px; border-radius: 10px; background-color: #007BFF; color: white; border: none; cursor: pointer; display: none;">
							Confirm
						</button>
						<script>
							// Initialize Firebase
							const firebaseConfig = {
								// Your web app's Firebase configuration
								// You'll need to replace this with your actual Firebase config
								apiKey: "YOUR_API_KEY",
								authDomain: "YOUR_AUTH_DOMAIN",
								projectId: "YOUR_PROJECT_ID",
								// ... other config options ...
							};
							firebase.initializeApp(firebaseConfig);
	
							const signInButton = document.getElementById('signInButton');
							const confirmButton = document.getElementById('confirmButton');
							const messageDiv = document.getElementById('message');
	
							firebase.auth().onAuthStateChanged(function(user) {
								if (user) {
									signInButton.style.display = 'none';
									confirmButton.style.display = 'block';
									messageDiv.textContent = 'Signed in as ' + user.email;
								} else {
									signInButton.style.display = 'block';
									confirmButton.style.display = 'none';
									messageDiv.textContent = 'Please sign in to continue';
								}
							});
	
							signInButton.onclick = function() {
								const provider = new firebase.auth.GoogleAuthProvider();
								firebase.auth().signInWithPopup(provider);
							};
	
							confirmButton.onclick = function() {
								firebase.auth().currentUser.getIdToken(true).then(function(idToken) {
									confirmUpdate('%s', idToken);
								}).catch(function(error) {
									console.error('Error getting ID token:', error);
								});
							};
	
							function confirmUpdate(leadID, idToken) {
								fetch('/update-lead?lead_id=' + leadID, {
									method: 'GET',
									headers: {
										'Content-Type': 'application/json',
										'Authorization': idToken
									}
								})
								.then(response => response.json())
								.then(data => {
									messageDiv.textContent = 'Lead ' + leadID + ' has been updated.';
									confirmButton.style.display = 'none';
								})
								.catch(error => {
									messageDiv.textContent = 'An error occurred: ' + error;
								});
							}
						</script>
					</body>
				</html>
			`, leadID, leadID)
		}

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
	leadID := r.URL.Query().Get("lead_id")
	if leadID == "" {
		http.Error(w, "Missing lead_id", http.StatusBadRequest)
		return
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

func main() {
	http.HandleFunc("/", handleLeadMailer) // This doesn't need auth as it serves the login page
	http.HandleFunc("/update-lead", authMiddleware(updateLeadHandler))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
