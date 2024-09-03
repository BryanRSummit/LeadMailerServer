package templates

import "fmt"

// Function to get the HTML content
func GetLeadUpdatedHTML(leadID string) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Lead Updated</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 36px;
                color: #333333;
            }
            p {
                font-size: 18px;
                color: #666666;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Lead Updated</h1>
            <p>Emails will no longer be sent for Lead ID: %s</p>
        </div>
    </body>
    </html>
    `, leadID)
}

func GetUnauthenticatedHTML(leadID string) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Authentication Required</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Authentication Required</h1>
            <p>Please <a href="/login?lead_id=%s">log in</a> with your Redd Summit email to continue.</p>
        </div>
    </body>
    </html>
    `, leadID)
}

// Function to get the HTML content
func GetConfirmHTML(leadID string) string {
	return fmt.Sprintf(`
			    <!DOCTYPE html>
                <html>
                <head>
                    <title>Lead Updated</title>
                    <style>
                        body {
                            font-family: Arial, sans-serif;
                            text-align: center;
                            padding-top: 50px;
                            background-color: #f4f4f4;
                        }
                        .container {
                            max-width: 600px;
                            margin: 0 auto;
                            padding: 20px;
                            background-color: #ffffff;
                            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                            border-radius: 8px;
                        }
                        h1 {
                            font-size: 36px;
                            color: #333333;
                        }
                        p {
                            font-size: 18px;
                            color: #666666;
                        }
                    </style>
                </head>
				<body>
                    <div class="container">
                        <h1 style="font-size: 24px;">Confirm That You Wish to Give Up Lead!</h1>
                        <p><b style="font-size: 20px;">This is a permanent Action! The lead will be killed or reassigned! Lead: %s</b></p>
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
                                            'Accept': 'text/html'
                                        }
                                    })
                                    .then(response => {
                                        if (!response.ok) {
                                            throw new Error('Network response was not ok');
                                        }
                                        return response.text(); // Use text() instead of json() for HTML response
                                    })
                                    .then(html => {
                                        // Replace the entire page content with the new HTML
                                        document.open();
                                        document.write(html);
                                        document.close();
                                    })
                                    .catch(error => {
                                        alert('An error occurred: ' + error);
                                    });
                                }
                            </script>
                    </div>
				</body>
			</html>
		`, leadID, leadID)
}

func GetPostRequestHTML(leadID string) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Post Request</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 36px;
                color: #333333;
            }
            p {
                font-size: 18px;
                color: #666666;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Post Request Recieved</h1>
            <p>to update lead use GET Request with lead_id = %s</p>
        </div>
    </body>
    </html>
    `, leadID)
}

// Exported function to get the HTML content for "Logged Out" message
func GetLoggedOutMessage() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Logged Out</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>You Have Been Logged Out</h1>
            <p>Thank you for visiting. Created by Bryan Edman.</p>
        </div>
    </body>
    </html>
    `
}

func GetMissingIdMessage() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Missing Lead Id</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 800px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>You Must Provide a Lead Id!</h1>
            <h2>Close this Window and try again, something has gone wrong!</h2>
            <p>Thank you for visiting. Created by Bryan Edman.</p>
        </div>
    </body>
    </html>
    `
}

func GetUnauthorizedEmailDomain() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Unauthorized Email Domain</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 800px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Unauthorized Email Domain!</h1>
            <h3>This application is restricted to members of Redd Summit.</h3>
            <p>Thank you for visiting. Created by Bryan Edman.</p>
        </div>
    </body>
    </html>
    `
}

func GetAlreadyLoggedOut() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Already Logged Out</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 800px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Already Logged Out!</h1>
            <p>Thank you for visiting. Created by Bryan Edman.</p>
        </div>
    </body>
    </html>
    `
}

func GetDateSelectHTML(leadID string) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 800px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
            button {
                font-size: 24px;
                padding: 10px 20px;
                margin-top: 20px;
                background-color: #4CAF50;
                color: white;
                border: none;
                border-radius: 4px;
                cursor: pointer;
            }
            button:hover {
                background-color: #45a049;
            }
            input[type="date"] {
                font-size: 24px;
                padding: 10px;
                width: 100%%;
                max-width: 400px;
                box-sizing: border-box;
                border-radius: 4px;
                border: 1px solid #ccc;
                margin-top: 20px;
            }
        </style>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Select Date</title>
        <script src="https://cdn.jsdelivr.net/npm/axios/dist/axios.min.js"></script>
    </head>
        <body>
            <div class="container">
                <h1>Select a Date</h1>
                    <form id="dateForm">
                        <input type="date" id="selectedDate" required>
                        <button type="submit">Submit</button>
                    </form>
                <script>
                    const leadId = "%s";
                    document.getElementById('dateForm').addEventListener('submit', function(e) {
                        e.preventDefault();
                        const date = document.getElementById('selectedDate').value;
                        
                        axios.post('/submit-date', { date, leadId })
                            .then(response => {
                                document.open();
                                document.write(response.data);
                                document.close();
                            })
                            .catch(error => {
                                alert('Error submitting date: ' + error.response.data);
                            });
                    });
                </script>
            </div>
        </body>
    </html>
    `, leadID)
}

func GetDateChangedHTML() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Date Changed</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                padding-top: 50px;
                background-color: #f4f4f4;
            }
            .container {
                max-width: 800px;
                margin: 0 auto;
                padding: 20px;
                background-color: #ffffff;
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                border-radius: 8px;
            }
            h1 {
                font-size: 48px;
                color: #333333;
            }
            p {
                font-size: 20px;
                color: #666666;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Reminder date Set!</h1>
            <h2>Go ahead and Close this tab now.</h2>
            <p>Thank you for visiting. Created by Bryan Edman.</p>
        </div>
    </body>
    </html>
    `
}
