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
	// return fmt.Sprintf(`
	//         <html>
	//             <body>
	//                 <h1>Authentication Required</h1>
	//                 <p>Please <a href="/login?lead_id=%s">log in</a> with your @reddsummit.com email to continue.</p>
	//             </body>
	//         </html>
	//     `, leadID)
}

// Function to get the HTML content
func GetConfirmHTML(leadID string) string {
	return fmt.Sprintf(`
			<html>
				<body>
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
