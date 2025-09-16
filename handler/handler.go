package handler

import (
	"html/template"
	"net/http"
	"time"

	"forum/database"
)

func IndexHandler(w http.ResponseWriter, r *http.Request, name string) {
	// if it's POST, create a post then redirect
	if r.Method == http.MethodPost {
		username := CurrentUsername(r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		title := r.FormValue("title")
		body := r.FormValue("body")
		if title == "" || body == "" {
			http.Redirect(w, r, "/?err=empty", http.StatusSeeOther)
			return
		}

		uid, err := getUserIDbyusername(username)
		if err != nil {
			http.Error(w, "user not found", http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec(
			`INSERT INTO posts(user_id, title, body) VALUES (?, ?, ?)`,
			uid, title, body,
		)
		if err != nil {
			http.Error(w, "cannot create post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// otherwise GET → render the page
	file := "templates/" + name + ".html"
	tpl, err := template.ParseFiles(file)
	if err != nil {
		http.Error(w, "template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := fetchPostsWithComments()
	if err != nil {
		http.Error(w, "load posts error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Posts    []Post
	}{
		Username: CurrentUsername(r),
		Posts:    posts,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
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
