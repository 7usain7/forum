package handler

import (
	"forum/database"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// renderPage renders a template with common data
func renderPage(w http.ResponseWriter, r *http.Request, templateName string, data any) {
	file := "templates/" + templateName + ".html"
	tpl, err := template.ParseFiles(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to parse template: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Create base data structure
	pageData := struct {
		Username string
		Data     any
	}{
		Username: CurrentUsername(r),
		Data:     data,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, pageData); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// IndexHandler handles GET requests for the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleCreatePost(w, r)
		return
	}

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		renderPage(w, r, "error", NotFound)
		return
	}

	posts, err := fetchPostsWithComments()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to fetch posts: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Check for filter parameters
	filter := r.URL.Query().Get("sort")

	if filter != "" {
		switch filter {
		case "newest", "oldest":
			posts, err = filterByCreationDate(filter)
		case "most_popular":
			posts, err = filterByPopularity()
		case "most_liked":
			posts, err = filterBymostliked()
		default:
			// Invalid filter, ignore or handle as needed
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			WriteErrorLog("error.log", "failed to filter posts: "+err.Error())
			renderPage(w, r, "error", InternalServerError)
			return
		}
	}

	// Fetch categories for the form
	categories, err := fetchAllCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to fetch categories: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Check for error parameters in URL
	errorType := r.URL.Query().Get("error")

	data := struct {
		Posts      []Post
		Categories []Category
		Filter     string
		Error      string
	}{
		Posts:      posts,
		Categories: categories,
		Error:      errorType,
		Filter:     filter,
	}

	renderPage(w, r, "index", data)
}

// handleCreatePost handles POST requests for creating posts
func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	username := CurrentUsername(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	if title == "" || body == "" {
		http.Redirect(w, r, "/?error=empty", http.StatusSeeOther)
		return
	}

	// Parse form to get categories
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		WriteErrorLog("error.log", "failed to parse form: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	categories := r.Form["categories"]
	if len(categories) == 0 {
		http.Redirect(w, r, "/?error=empty_categories", http.StatusSeeOther)
		return
	}

	uid, err := getUserIDbyusername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to get user ID: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Get the post ID after insertion
	res, err := database.DB.Exec(
		`INSERT INTO posts(user_id, title, body) VALUES (?, ?, ?)`,
		uid, title, body,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to create post: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Handle categories
	RawPostID, err := res.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to get post ID: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}
	postID := int(RawPostID)

	for _, categorieName := range categories {
		var categorieID int

		err := database.DB.QueryRow("SELECT id FROM categories WHERE name = ?", categorieName).Scan(&categorieID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			WriteErrorLog("error.log", "failed to query category: "+err.Error())
			return
		}
		// Link post to category
		_, err = database.DB.Exec("INSERT OR IGNORE INTO post_categories(post_id, category_id) VALUES (?, ?)", postID, categorieID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			WriteErrorLog("error.log", "failed to link category: "+err.Error())
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LoginHandler handles login page
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		LoginPOST(w, r)
		return
	}

	renderPage(w, r, "login", nil)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		RegisterPOST(w, r)
		return
	}
	renderPage(w, r, "register", nil)
}

func CurrentUsername(r *http.Request) string {
	c, err := r.Cookie("session_id")
	if err != nil || c.Value == "" {
		return ""
	}

	var (
		username string
		expires  time.Time
	)
	err = database.DB.QueryRow(`
        SELECT u.username, s.expires_at
        FROM sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.id = ?`,
		c.Value,
	).Scan(&username, &expires)
	if err != nil {
		return ""
	}
	if !expires.IsZero() && time.Now().After(expires) {
		// optional cleanup
		database.DB.Exec(`DELETE FROM sessions WHERE id = ?`, c.Value)
		return ""
	}
	return username
}

func HandleLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := CurrentUsername(r) // you already have session logic
	if username == "" {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	userID, err := getUserIDbyusername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	targetType := r.FormValue("target_type") // "post" or "comment"
	targetID := r.FormValue("target_id")
	likeType := r.FormValue("like_type") // "1" for like, "-1" for dislike

	_, err = database.DB.Exec(`
        INSERT INTO likes (user_id, target_type, target_id, like_type)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(user_id, target_type, target_id)
        DO UPDATE SET like_type = excluded.like_type
    `, userID, targetType, targetID, likeType)

	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func SubforumHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if r.Method == http.MethodPost {
		category := r.FormValue("category")
		if category != "" {
			http.Redirect(w, r, "/r/"+category, http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if path == "/r/" || path == "/r" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// Expecting path like /r/{subforum}
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "r" || parts[2] == "" {
		http.NotFound(w, r)
		return
	}
	subforum := parts[2]
	renderSubforum(w, r, subforum)
}

func renderSubforum(w http.ResponseWriter, r *http.Request, subforum string) {
	posts, err := fetchPostsWithComments()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to fetch posts: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Filter posts by subforum (category)
	var filteredPosts []Post
	for _, post := range posts {
		for _, cat := range post.Categories {
			if strings.EqualFold(cat.Name, subforum) {
				filteredPosts = append(filteredPosts, post)
				break
			}
		}
	}

	// Fetch categories for the form
	categories, err := fetchAllCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to fetch categories: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	// Check for error parameters in URL
	errorType := r.URL.Query().Get("error")
	filter := r.URL.Query().Get("sort")

	data := struct {
		Posts      []Post
		Categories []Category
		Filter     string
		Error      string
		Subforum   string
	}{
		Posts:      filteredPosts,
		Categories: categories,
		Filter:     filter,
		Error:      errorType,
		Subforum:   subforum,
	}

	renderPage(w, r, "index", data)
}
