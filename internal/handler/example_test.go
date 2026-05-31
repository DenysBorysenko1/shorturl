package handler_test

import (
	"fmt"
	_ "shorturl/internal/handler"
)

// ExampleGenerate demonstrates the Generate handler creation.
//
// This example shows how to create a handler for shortening URLs using plain text format.
// The handler accepts POST requests with a URL in the request body and returns a short URL.
//
// Request format:
//   POST /shorten
//   Content-Type: text/plain
//   Body: https://example.com
//
// Response format:
//   Status: 201 Created
//   Content-Type: text/plain
//   Body: http://localhost:8080/abc12345
func ExampleGenerate() {
	fmt.Println("Generate handler creates short URLs from plain text requests")
	fmt.Println("Endpoint: POST /shorten")
	fmt.Println("Request Body: URL in plain text")
	fmt.Println("Response: Short URL with 201 status")
	
	// Output:
	// Generate handler creates short URLs from plain text requests
	// Endpoint: POST /shorten
	// Request Body: URL in plain text
	// Response: Short URL with 201 status
}

// ExampleGenerateJSON demonstrates the GenerateJSON handler creation.
//
// This example shows how to create a handler for shortening URLs using JSON format.
// The handler accepts POST requests with a JSON body containing the URL and returns
// a JSON response with the short URL.
//
// Request format:
//   POST /api/shorten
//   Content-Type: application/json
//   Body: {"url": "https://example.com"}
//
// Response format:
//   Status: 201 Created
//   Content-Type: application/json
//   Body: {"result": "http://localhost:8080/abc12345"}
func ExampleGenerateJSON() {
	fmt.Println("GenerateJSON handler creates short URLs from JSON requests")
	fmt.Println("Endpoint: POST /api/shorten")
	fmt.Println("Request Body: JSON with 'url' field")
	fmt.Println("Response: JSON with 'result' field containing short URL")
	
	// Output:
	// GenerateJSON handler creates short URLs from JSON requests
	// Endpoint: POST /api/shorten
	// Request Body: JSON with 'url' field
	// Response: JSON with 'result' field containing short URL
}

// ExampleGenerateBatch demonstrates the GenerateBatch handler creation.
//
// This example shows how to create a handler for batch shortening of URLs.
// The handler accepts POST requests with a JSON array of URLs and returns
// a JSON array with corresponding short URLs.
//
// Request format:
//   POST /api/shorten/batch
//   Content-Type: application/json
//   Body: [
//     {"correlation_id": "id1", "original_url": "https://example1.com"},
//     {"correlation_id": "id2", "original_url": "https://example2.com"}
//   ]
//
// Response format:
//   Status: 201 Created
//   Content-Type: application/json
//   Body: [
//     {"correlation_id": "id1", "short_url": "http://localhost:8080/id1"},
//     {"correlation_id": "id2", "short_url": "http://localhost:8080/id2"}
//   ]
func ExampleGenerateBatch() {
	fmt.Println("GenerateBatch handler creates multiple short URLs in one request")
	fmt.Println("Endpoint: POST /api/shorten/batch")
	fmt.Println("Request Body: JSON array of correlation_id and original_url pairs")
	fmt.Println("Response: JSON array with correlation_id and short_url pairs")
	
	// Output:
	// GenerateBatch handler creates multiple short URLs in one request
	// Endpoint: POST /api/shorten/batch
	// Request Body: JSON array of correlation_id and original_url pairs
	// Response: JSON array with correlation_id and short_url pairs
}

// ExampleRetrieve demonstrates the Retrieve handler creation.
//
// This example shows how to create a handler for retrieving original URLs from short URLs.
// The handler accepts GET requests with a short URL ID and redirects to the original URL.
//
// Request format:
//   GET /{id}
//
// Response format:
//   Status: 307 Temporary Redirect
//   Location: https://example.com (original URL)
func ExampleRetrieve() {
	fmt.Println("Retrieve handler redirects short URLs to original URLs")
	fmt.Println("Endpoint: GET /{id}")
	fmt.Println("Response: 307 redirect to original URL")
	
	// Output:
	// Retrieve handler redirects short URLs to original URLs
	// Endpoint: GET /{id}
	// Response: 307 redirect to original URL
}

// ExampleListUrls demonstrates the ListUrls handler creation.
//
// This example shows how to create a handler for retrieving all URLs created by
// an authenticated user. The handler accepts GET requests and returns a JSON array
// of all the user's URLs with their short URLs.
//
// Request format:
//   GET /api/user/urls
//
// Response format (200 OK):
//   Content-Type: application/json
//   Body: [
//     {"original_url": "https://example1.com", "short_url": "http://localhost:8080/abc123"},
//     {"original_url": "https://example2.com", "short_url": "http://localhost:8080/def456"}
//   ]
//
// Response format (204 No Content):
//   User has no URLs
func ExampleListUrls() {
	fmt.Println("ListUrls handler retrieves all user's URLs")
	fmt.Println("Endpoint: GET /api/user/urls")
	fmt.Println("Response: JSON array of original_url and short_url pairs")
	fmt.Println("Empty result: 204 No Content status")
	
	// Output:
	// ListUrls handler retrieves all user's URLs
	// Endpoint: GET /api/user/urls
	// Response: JSON array of original_url and short_url pairs
	// Empty result: 204 No Content status
}

// ExampleRemoveListUrls demonstrates the RemoveListUrls handler creation.
//
// This example shows how to create a handler for deleting multiple URLs asynchronously.
// The handler accepts DELETE requests with a JSON array of URL IDs and enqueues
// them for background deletion.
//
// Request format:
//   DELETE /api/user/urls
//   Content-Type: application/json
//   Body: ["abc123", "def456", "ghi789"]
//
// Response format:
//   Status: 202 Accepted
// Note: Deletion is processed asynchronously in the background
func ExampleRemoveListUrls() {
	fmt.Println("RemoveListUrls handler deletes multiple URLs asynchronously")
	fmt.Println("Endpoint: DELETE /api/user/urls")
	fmt.Println("Request Body: JSON array of short URL IDs to delete")
	fmt.Println("Response: 202 Accepted (deletion processed in background)")
	
	// Output:
	// RemoveListUrls handler deletes multiple URLs asynchronously
	// Endpoint: DELETE /api/user/urls
	// Request Body: JSON array of short URL IDs to delete
	// Response: 202 Accepted (deletion processed in background)
}

// ExamplePing demonstrates the Ping handler creation.
//
// This example shows how to create a handler for checking database connectivity.
// The handler accepts GET requests and returns the status of the database connection.
//
// Request format:
//   GET /ping
//
// Response formats:
//   Status: 200 OK - Database connection is healthy
//   Status: 500 Internal Server Error - Database connection failed
func ExamplePing() {
	fmt.Println("Ping handler checks database connectivity")
	fmt.Println("Endpoint: GET /ping")
	fmt.Println("Response: 200 OK if database is accessible, 500 otherwise")
	
	// Output:
	// Ping handler checks database connectivity
	// Endpoint: GET /ping
	// Response: 200 OK if database is accessible, 500 otherwise
}

