package handler

import (
	"database/sql"
	"forum/database"
	"net/http"
)

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID           int
	UserID       int
	Title        string
	Body         string
	CreatedAt    string
	Comments     []Comment
	Categories   []Category
	LikeCount    int // total likes
	DislikeCount int // optional, for like_type = -1
	Username     string
}

type Comment struct {
	ID           int
	PostID       int
	UserID       int
	Username     string
	Body         string
	CreatedAt    string
	LikeCount    int
	DislikeCount int
}

func getUsernameFromSession(r *http.Request) string {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func getUserIDbyusername(username string) (int, error) {
	var id int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	return id, err
}

func fetchPostsWithComments(fetchType, username string) ([]Post, error) {
	var rows *sql.Rows
	var err error
	switch fetchType {
	case "all":
		rows, err = database.DB.Query(`
        SELECT p.id, p.user_id, u.username, p.title, p.body, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
    `)
	case "liked":
		rows, err = database.DB.Query(`
		SELECT p.id, p.user_id, u.username, p.title, p.body, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN likes l ON p.id = l.target_id
		JOIN users liking_user ON l.user_id = liking_user.id
		WHERE l.like_type = 1 AND l.target_type = 'post' AND liking_user.username = ?
		GROUP BY p.id
		ORDER BY COUNT(l.id) DESC
		`, username)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Body, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Categories = []Category{}
		p.Comments = []Comment{}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range posts {
		// --- Fetch comments with username ---
		commentRows, err := database.DB.Query(`
			SELECT c.id, c.post_id, c.user_id, u.username, c.body, c.created_at
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.created_at ASC
        `, posts[i].ID)
		if err != nil {
			return nil, err
		}

		var comments []Comment
		for commentRows.Next() {
			var c Comment
			if err := commentRows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Body, &c.CreatedAt); err != nil {
				commentRows.Close()
				return nil, err
			}
			comments = append(comments, c)
		}
		commentRows.Close()
		if err := commentRows.Err(); err != nil {
			return nil, err
		}
		posts[i].Comments = comments

		// --- Categories ---
		catRows, err := database.DB.Query(`
            SELECT c.id, c.name 
            FROM categories c
            JOIN post_categories pc ON c.id = pc.category_id
            WHERE pc.post_id = ?
        `, posts[i].ID)
		if err != nil {
			return nil, err
		}

		var categories []Category
		for catRows.Next() {
			var cat Category
			if err := catRows.Scan(&cat.ID, &cat.Name); err != nil {
				catRows.Close()
				return nil, err
			}
			categories = append(categories, cat)
		}
		catRows.Close()
		if err := catRows.Err(); err != nil {
			return nil, err
		}
		posts[i].Categories = categories

		// --- Post Likes ---
		if err := database.DB.QueryRow(`
            SELECT 
                COALESCE(SUM(CASE WHEN like_type = 1 THEN 1 ELSE 0 END), 0),
                COALESCE(SUM(CASE WHEN like_type = -1 THEN 1 ELSE 0 END), 0)
            FROM likes
            WHERE target_type = 'post' AND target_id = ?
        `, posts[i].ID).Scan(&posts[i].LikeCount, &posts[i].DislikeCount); err != nil {
			return nil, err
		}

		// --- Comment Likes ---
		for j := range posts[i].Comments {
			if err := database.DB.QueryRow(`
                SELECT 
                    COALESCE(SUM(CASE WHEN like_type = 1 THEN 1 ELSE 0 END), 0),
                    COALESCE(SUM(CASE WHEN like_type = -1 THEN 1 ELSE 0 END), 0)
                FROM likes
                WHERE target_type = 'comment' AND target_id = ?
            `, posts[i].Comments[j].ID).Scan(&posts[i].Comments[j].LikeCount, &posts[i].Comments[j].DislikeCount); err != nil {
				return nil, err
			}
		}
	}

	return posts, nil
}

// Getting all categories for the form
func fetchAllCategories() ([]Category, error) {
	rows, err := database.DB.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var c Category

		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
