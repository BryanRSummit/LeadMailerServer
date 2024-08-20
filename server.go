package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	sheetsService *sheets.Service
	spreadsheetID = "1QKE4Gm54fyNA-ArMmd3MJgMkXG0pPfGfKLiIxpcsMtg"
	sheetName     = "Sheet1"
)

func init() {
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
		leadID := r.URL.Query().Get("lead_id")
		if leadID == "" {
			http.Error(w, "Missing lead_id", http.StatusBadRequest)
			return
		}

		// err := updateLeadInSheet(leadID)
		// if err != nil {
		// 	http.Error(w, fmt.Sprintf("An error occurred: %v", err), http.StatusInternalServerError)
		// 	return
		// }

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<html>
				<body>
					<h1>Confirm That you wish to Give Up Lead!</h1>
					<p>This is a permanent Action! Lead: %s</p>
					<button onclick="confirmUpdate('%s')">Confirm</button>
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
								
								// Check if the button element is available
								if (buttonElement) {
									// Remove the button and show a success message
									buttonElement.parentNode.removeChild(buttonElement);
									
									const successMessage = document.createElement('p');
									successMessage.textContent = 'Lead ' + leadID + ' has been updated.';
									buttonElement.parentNode.appendChild(successMessage);
								} else {
									// Button element not found, display a message
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
	http.HandleFunc("/", handleLeadMailer)
	http.HandleFunc("/update-lead", updateLeadHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
