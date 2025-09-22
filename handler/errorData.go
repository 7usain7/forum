package handler

import (
	"fmt"
	"os"
	"time"
)

type HTTPError struct {
	Code    int
	Title   string
	Message string
}

func NewHTTPError(code int, title, message string) HTTPError {
	return HTTPError{
		Code:    code,
		Title:   title,
		Message: message,
	}
}

// Predefined errors
var (
	BadRequest          = NewHTTPError(400, "Bad Request", "Your request could not be understood by the server. Please check your input and try again.")
	Unauthorized        = NewHTTPError(401, "Unauthorized", "You must be logged in to access this page. Please log in and try again.")
	Forbidden           = NewHTTPError(403, "Forbidden", "You do not have permission to view this page or perform this action.")
	NotFound            = NewHTTPError(404, "Page Not Found", "The page you are looking for does not exist or may have been moved.")
	MethodNotAllowed    = NewHTTPError(405, "Method Not Allowed", "The method you used is not allowed for this resource.")
	Conflict            = NewHTTPError(409, "Conflict", "There was a conflict with your request. Please try again.")
	InternalServerError = NewHTTPError(500, "Internal Server Error", "An unexpected error occurred on the server. Please try again later.")
)

// log file for errors
func WriteErrorLog(filename, errorMessage string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] ERROR: %s\n", timestamp, errorMessage)

	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}
