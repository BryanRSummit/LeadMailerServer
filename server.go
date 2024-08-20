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
	// // Set the environment variable directly in the code
	// os.Setenv("SHEETS_CREDS", `{
	// 	"type": "service_account",
	// 	"project_id": "agentleadmailer",
	// 	"private_key_id": "8cc577104ac3f36b7a4bac187e5172e529dc8d9e",
	// 	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC1JJQhLA75qKa5\n3uEfMZm0Kz3uMg72lH3Chs53L/f/vTjWItwqGfdapyLCUKEwzwAkuTyFLh0ralk1\nTnkukuCtsZ3uXkfJmEOOlQLsJPRi0DeKm/nDsVa/iVUtE51xx1doZFjELJ6CydPn\nkf+VHzuxScSU+zDlBuMtid8qb9vWKwF+DSE3azex2UwflqdngjH/TuzcqaWM09s9\nJE+mEmEM4FZI7TeR6EDSsHZM5NGn5qWqX2zzLVnbrfYkL3YjXsXJZy3dgc4L3Sev\nfaCIhlQNwhL3VaPORQYj9yZ+mVJHbn6zpSN5QjHtvR4A3M0pPlY+nzx32PrMtM5P\nftZAEva1AgMBAAECggEAIEPrNLtt0W+Gfx4hmFZT7AE1z0tQWgCaI/+yIA3FzWJN\nkOr1r3QfmKCjstv80j5U5rWt/4T2wih3ymR3dmHILngwSui1PcXm5qtJMXnlpAI1\nmnVs+DwK2SQjrVtMlJsuyRPyscLG20ILAjkBvvSow8wBfY3+qBThe1ePDjaNgGis\nck12coD95pBVUuUcQFF5+3kRyMyrs1Q7hJegd25zeWA7ogO+o2OJ4smQfhfhUxyU\nTM0XWdYPBH1dfLlXNr3c6U+U92ZDd03LB9u2+nBNzKh3GiCIU6oLAFm+TleTTZ18\nFfLiO2vrq9C3gg2XgYWy1jWaxQfUmqV+4FNwM2IRIQKBgQDwxbV+XGuBf+f7P5SJ\nq5GKoxcMMZN2dOzGxdtdynVv93bDHzOmY3nqHEMgkYsIHXmUVC5zID/HcABYaa0L\nxitiXWj+iOW+yEKqQ9FNjV9nOUu+0RoKU5p4gTJL0QcjNK3v2VNMtuspQ50pgPCH\ndkKz16zogvqlxATrTMdQAb1AMQKBgQDAmWxJ60W0E98YM++VA1bUfkkQqj5LQGMN\nvL69k0/Fp7Eney/EO0RkPHyBGx8z+my60H9UFubuLLsMe++YbR7IIbG1d0EFJyH0\nioX2D2uPksl4fEAV0WnKUYjqgnWP4Q1t5LnkkX+dAUQkupViXsGZhcJxuTpFUlNr\nGel1XH5hxQKBgGNDb4bv/VZ/cBmSZd+4PyGkCV16luwQWom8iqsJTA9kO69IDtg7\nTMjq6/XiaypmVHiFmDzYf9LuZwYMU052Xe6Iyj+eGvHjyDBAE2tgrIN3CLZbqNu3\nCglCYoUFYWbvUgJ/W6tWAm+Zs5Kn2QJQDEHu2hdl4IY04T5NAiMHBIoRAoGBAJnU\nsPplcVoAmSsiqFRTw3GboE4wO+ss9TDOtWaDl66eXs/TA3bvg5OwAB26hPSmK1wX\nFewbEr3felLhVqBfX7unteHj60nrVKKWVaMP8/BL5KFYVHNYvO98qifspWuS7H/+\ntT9LuyqzDTNs184nMuilPoZI1LLzq28a1i4H/2WlAoGAR6jdam9PanzV8lQVmtOo\nEADY3FMZ+cMRferPFKAopxu6NaLA0LjEOreSub1DnyBoH+flqIJ9O/byPfB6kGNc\n7Co4b1/01FlA0gdlGv84rGTUQ9C3eHZROZIJ2ZGlkvBd+IV2c+OWjSsMUGL2Wqmy\n8I4LlkY/mPWiH6yGrK/TORA=\n-----END PRIVATE KEY-----",
	// 	"client_email": "agentleadmailerservice@agentleadmailer.iam.gserviceaccount.com",
	// 	"client_id": "109660796532972752753",
	// 	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	// 	"token_uri": "https://oauth2.googleapis.com/token",
	// 	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	// 	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/agentleadmailerservice%40agentleadmailer.iam.gserviceaccount.com",
	// 	"universe_domain": "googleapis.com"
	//   }`)

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
				<h1 style="font-size: 24px;">Confirm That You Wish to Give Up Lead!</h1>
				<p><b style="font-size: 20px;">This is a permanent Action! Lead: %s</b></p>
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
