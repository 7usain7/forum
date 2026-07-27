# Forum Project

A full-featured web forum built with Go, SQLite, and Docker that allows users to communicate, share posts, and interact with content similar to Reddit.

## Features

### 🔐 Authentication & User Management
- User registration with email, username, and password
- Secure login sessions with cookies
- Password encryption (bcrypt)
- UUID-based session management
- Email uniqueness validation

### 💬 Communication
- Create and view posts
- Comment on posts
- Category-based post organization
- Public visibility for all content (registered users can interact, everyone can view)

### 👍 Interactions
- Like and dislike posts and comments
- Visible engagement metrics for all users
- Interactive voting system

### 🔍 Filtering & Discovery
- Filter posts by categories
- View created posts (registered users)
- View liked posts (registered users)
- Category-based subforums

## Tech Stack

- **Backend**: Go (Golang)
- **Database**: SQLite3
- **Authentication**: Cookie-based sessions with UUID
- **Security**: bcrypt for password hashing
- **Containerization**: Docker
- **Frontend**: Native HTML/CSS (No frameworks)

## Database Schema

The application uses SQLite with the following key entities:

- **Users** (id, email, username, password_hash, created_at)
- **Sessions** (id, user_id, expires_at)
- **Categories** (id, name)
- **Posts** (id, user_id, title, body, created_at)
- **Post_Categories** (post_id, category_id)
- **Comments** (id, post_id, user_id, body, created_at)
- **Likes** (id, user_id, target_type, target_id, like_type)

## Prerequisites

- Docker

## Installation & Setup

1. **Clone the repository**
   ```bash
   git clone https://learn.reboot01.com/git/mohkadhem/forum
   cd forum
   ```

2. **Build and run with Docker**
   ```bash
   docker build -t forum-app .

   docker run -d -p 8080:8080 --name forum-container forum-app

   docker start forum-container
   ```

3. **Access the application**
   ```
   http://localhost:8080
   ```

## Security Features

- Password hashing with bcrypt
- Session management with UUID tokens
- Cookie expiration management
- Input validation and sanitization
- SQL injection prevention
- XSS protection

## Error Handling

- Comprehensive HTTP status code responses
- User-friendly error messages
- Database connection error handling
- Input validation errors
- Session management errors

## Development

### Project Structure
```
forum-project/
├── database/
├── handler/
├── web/
│   ├── templates/
│   └── static/
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Learning Outcomes

This project demonstrates understanding of:
- Web development fundamentals (HTML, HTTP)
- Session and cookie management
- Docker containerization
- SQL database design and operations
- Data encryption and security
- Error handling and validation

---

For questions or support, please open an issue in the project repository.

## License

This project is proprietary. The code is publicly visible for portfolio and viewing purposes only. No unauthorized copying, modification, or distribution is permitted. See [LICENSE](LICENSE) for details.
