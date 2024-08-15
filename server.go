package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	sheetsService *sheets.Service
	spreadsheetID = "1QKE4Gm54fyNA-ArMmd3MJgMkXG0pPfGfKLiIxpcsMtg"
	sheetName     = "Sheet1"
)

func init() {
	// Initialize Google Sheets API client
	credBytes, err := os.ReadFile("./agentleadmailer-8cc577104ac3.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(credBytes, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Use the context when creating the client
	client := config.Client(ctx)

	sheetsService, err = sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	data, err := getSheetData()
	if err != nil {
		log.Printf("Error reading sheet: %v", err)
		http.Error(w, "Error reading sheet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body><table border='1'>")

	if len(data) < 6 {
		fmt.Fprintf(w, "<tr><td>Not enough data in the sheet</td></tr>")
	} else {
		// Iterate through the data starting from row 6
		for i := 6; i < len(data); i++ {
			row := data[i]

			// Check if the row has less than 2 columns or if the second column is empty
			if len(row) < 2 || row[1] == "" {
				break
			}

			fmt.Fprintf(w, "<tr>")
			for _, cell := range row {
				fmt.Fprintf(w, "<td>%v</td>", cell)
			}
			fmt.Fprintf(w, "</tr>")
		}
	}

	fmt.Fprintf(w, "</table></body></html>")
}

func getSheetData() ([][]interface{}, error) {
	// Specify the range to read
	readRange := fmt.Sprintf("%s!A1:Z", sheetName) // This will read all columns from row 1 to 26

	// Perform the read operation
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// Print the data to the console for debugging
	for _, row := range resp.Values {
		fmt.Println(row)
	}

	return resp.Values, nil
}

func updateSheet() error {
	// Define the values to be inserted
	values := [][]interface{}{
		{"✓"}, // This creates a checkmark. You can change this to any value you want.
	}

	// Create a ValueRange object with these values
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Define the range where you want to update
	// This example updates cell A1. Modify as needed.
	updateRange := fmt.Sprintf("%s!A1", sheetName)

	// Perform the update operation
	_, err := sheetsService.Spreadsheets.Values.Update(spreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	return err
}

func main() {
	http.HandleFunc("/", handleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
