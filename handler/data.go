package handler

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
