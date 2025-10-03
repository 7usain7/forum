package handler

func filterByCreationDate(order string) ([]Post, error) {
	posts, err := fetchPostsWithComments()
	if err != nil {
		return nil, err
	}
	switch order {
	case "newest":
		// Sort posts by CreatedAt newstest first (descending order)
		for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
			posts[i], posts[j] = posts[j], posts[i]
		}
		return posts, nil
	case "oldest":
		// Posts are already ordered by created_at DESC in fetchPostsWithComments
		return posts, nil
	}
	return posts, nil
}

func filterByPopularity(order string) ([]Post, error) {
	posts, err := fetchPostsWithComments()
	if err != nil {
		return nil, err
	}
	switch order {
	case "most_popular":
		// Sort posts by number of comments (descending order)
		for i := 0; i < len(posts)-1; i++ {
			for j := 0; j < len(posts)-i-1; j++ {
				if len(posts[j].Comments) < len(posts[j+1].Comments) {
					posts[j], posts[j+1] = posts[j+1], posts[j]
				}
			}
		}
		return posts, nil
	case "least_popular":
		// Sort posts by number of comments (ascending order)
		for i := 0; i < len(posts)-1; i++ {
			for j := 0; j < len(posts)-i-1; j++ {
				if len(posts[j].Comments) > len(posts[j+1].Comments) {
					posts[j], posts[j+1] = posts[j+1], posts[j]
				}
			}
		}
		return posts, nil
	}
	return posts, nil
}

func filterBymostliked(order string) ([]Post, error) {
	posts, err := fetchPostsWithComments()
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
