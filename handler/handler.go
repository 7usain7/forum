package handler

import (
	"html/template"
	"net/http"
	"time"

	"forum/database"
)

func IndexHandler(w http.ResponseWriter, r *http.Request, name string) {
	file := "templates/" + name + ".html"

	tpl, err := template.ParseFiles(file)
	if err != nil {
		http.Error(w, "template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
	}{
		Username: CurrentUsername(r),
	}

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

