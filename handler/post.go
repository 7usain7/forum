package handler

import (
	"forum/database"
)

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID         int
	UserID     int
	Title      string
	Body       string
	CreatedAt  string
	Comments   []Comment
	Categories []Category
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Body      string
	CreatedAt string
}

func getUserIDbyusername(username string) (int, error) {
	var id int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	return id, err
}

func fetchPostsWithComments() ([]Post, error) {
	rows, err := database.DB.Query(`
        SELECT id, user_id, title, body, created_at
        FROM posts
        ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Body, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	for i := range posts {
		cr, err := database.DB.Query(`
            SELECT id, post_id, user_id, body, created_at
            FROM comments
            WHERE post_id = ?
            ORDER BY created_at ASC`, posts[i].ID)
		if err != nil {
			return nil, err
		}
		var cs []Comment
		for cr.Next() {
			var c Comment
			if err := cr.Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
				cr.Close()
				return nil, err
			}
			cs = append(cs, c)
		}
		cr.Close()
		posts[i].Comments = cs

		// Post categories fetch
		categorieRow, err := database.DB.Query(`
            SELECT c.id, c.name 
            FROM categories c
            JOIN post_categories pc ON c.id = pc.category_id
            WHERE pc.post_id = ?`, posts[i].ID)

		if err != nil {
			return nil, err
		}

		var categories []Category
		for categorieRow.Next() {
			var categ Category
			if err := categorieRow.Scan(&categ.ID, &categ.Name); err != nil {
				categorieRow.Close()
				return nil, err
			}
			categories = append(categories, categ)
		}
		categorieRow.Close()
		posts[i].Categories = categories
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
