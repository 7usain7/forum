package handler

import (
	"net/http"
)

func filterByCreationDate(order string) ([]Post, error) {
	posts, err := fetchPostsWithComments("all", "")
	if err != nil {
		return nil, err
	}
	switch order {
	case "oldest":
		// Sort posts by CreatedAt newstest first (descending order)
		for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
			posts[i], posts[j] = posts[j], posts[i]
		}
		return posts, nil
	case "newest":
		// Posts are already ordered by created_at DESC in fetchPostsWithComments
		return posts, nil
	}
	return posts, nil
}

func filterByPopularity() ([]Post, error) {
	posts, err := fetchPostsWithComments("all", "")
	if err != nil {
		return nil, err
	}

	// Sort posts by number of comments (descending order)
	for i := 0; i < len(posts)-1; i++ {
		for j := 0; j < len(posts)-i-1; j++ {
			if len(posts[j].Comments) < len(posts[j+1].Comments) {
				posts[j], posts[j+1] = posts[j+1], posts[j]
			}
		}
	}

	return posts, nil
}

func filterBymostliked() ([]Post, error) {
	posts, err := fetchPostsWithComments("all", "")
	if err != nil {
		return nil, err
	}

	// Sort posts by number of likes (descending order)
	for i := 0; i < len(posts)-1; i++ {
		for j := 0; j < len(posts)-i-1; j++ {
			if posts[j].LikeCount < posts[j+1].LikeCount {
				posts[j], posts[j+1] = posts[j+1], posts[j]
			}
		}
	}
	return posts, nil
}

func fetchAndFilterPosts(r *http.Request) ([]Post, string, error) {
	posts, err := fetchPostsWithComments("all", "")
	if err != nil {
		return nil, "", err
	}

	filter := r.URL.Query().Get("sort")
	if filter == "" {
		return posts, "", nil
	}

	switch filter {
	case "newest", "oldest":
		posts, err = filterByCreationDate(filter)
	case "most_popular":
		posts, err = filterByPopularity()
	case "most_liked":
		posts, err = filterBymostliked()
	default:
		// unknown filter: ignore and return original slice
		filter = ""
	}
	return posts, filter, err
}

// prepareIndexData builds the payload passed to the index template.
func prepareIndexData(posts []Post, categories []Category, filter, errorType string, validationErrorList []string, currentPage ...string) any {
	page := "home"
	if len(currentPage) > 0 && currentPage[0] != "" {
		page = currentPage[0]
	}

	return struct {
		Posts            []Post
		Categories       []Category
		Filter           string
		Error            string
		ValidationErrors []string
		CurrentPage      string
	}{
		Posts:            posts,
		Categories:       categories,
		Filter:           filter,
		Error:            errorType,
		ValidationErrors: validationErrorList,
		CurrentPage:      page,
	}
}

func CreatedPostsHandler(w http.ResponseWriter, r *http.Request) {
	username := getUsernameFromSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	posts, err := fetchPostsWithComments("created_by", username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "Failed to fetch posts based on created by: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	categories, err := fetchAllCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "Failed to fetch categories: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	data := prepareIndexData(posts, categories, "", "", nil, "created")
	renderPage(w, r, "index", data)
}

func errorString(err string) string {
	switch err {
	case "empty":
		return "Please fill in all fields."
	case "empty_categories":
		return "Please select a category."
	case "title_too_long":
		return "Title exceeds maximum length of 128 characters."
	case "body_too_long":
		return "Body exceeds maximum length of 512 characters."
	case "comment_too_long":
		return "Comment exceeds maximum length of 512 characters."
	default:
		return ""
	}
}
