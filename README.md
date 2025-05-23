# RethinkDB API Server

## Overview

This project is a RESTful API server built using Go and Echo framework, with RethinkDB as the database and Redis for session management. It provides user and book management features with JWT-based authentication.

## Features

- User authentication and session management with JWT and Redis
- CRUD operations for users and books
- RethinkDB as the primary database
- Echo framework for handling HTTP requests
- Static file serving

## Prerequisites

Ensure you have the following installed:

- Go 1.18 or later
- RethinkDB
- Redis

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/N-Jangra/rethink-db.git
   cd rethink-db
   ```
2. Install dependencies:
   ```sh
   go mod init rethink
   go mod tidy
   ```
3. Configure the application:
   - Create a `config.json` file with the following structure:
     ```json
     {
       "DB_HOST":"your_database_address",
       "DB_PORT":"your_database_port",
       "DB_NAME":"your_database_name",
       
       "JWT_SECRET":"******"
     }
     ```
   - Replace all these with your actual values.

## Running the Application

1. Start RethinkDB and Redis servers:
   ```sh
   rethinkdb &
   redis-server &
   ```
2. Run the Go application:
   ```sh
   go run main.go
   ```

The server will start on port `8090`.

## API Endpoints

### Authentication

- `POST /api/login` - User login
- `POST /api/register` - User registration

### Users

- `GET /api/users/:id` - Get user details
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### Books

- `GET /api/books` - List books
- `POST /api/books` - Add a new book
- `GET /api/books/:id` - Get book details
- `PUT /api/books/:id` - Update book
- `DELETE /api/books/:id` - Delete book

## Usage

You can use `curl` or Postman to test the API endpoints. Example:

```sh
curl -X POST http://localhost:8090/api/login -d '{"username": "test", "password": "pass"}' -H "Content-Type: application/json"
```

## Project Structure

```
rethinkdb/
│── api/
│   ├── db/                   # Database connection
│   │   ├── db.go                            
│   │   └── pass.go                         
│   ├── handlers/             # Request handlers
│   │   ├── books.go              
│   │   ├── jwt.go                
│   │   ├── root.go               
│   │   ├── users.go              
│   │   └── web.go               
│   ├── middleware/           # Middlewares for authentication 
│   │   ├── auth.go                           
│   │   └── books.go          
│   ├── models/               # Request handlers
│   │   ├── access.go                 
│   │   ├── appusers.go                
│   │   ├── books.go                
│   │   ├── privileges.go                  
│   │   └── roles.go          
│   ├── repo/                 # Repository layer
│   │   ├── books.go                           
│   │   └── users.go          
│   ├── routes/               # API route definitions
│   │   ├── books.go                          
│   │   └── users.go          
│   ├── web/                  # API route definitions
│   │   └── frontend files    # (templates, html, css)          
│── config.json               # Configuration file
│── main.go                   # Application entry point
│── go.mod                    # Go module dependencies
│── go.sum                    # Dependency checksums
│── seed.txt                  # RethinkDB insertions for roles
│── rethinkdb_data/           # RethinkDB storage
```

## License

This project is licensed under the GPL-3.0 License.

## Copyright

© 2025 Nitin - itznitinjangra@gmail.com . All rights reserved.
