# Quick Start Guide

## Prerequisites

- Go 1.21+ installed
- PostgreSQL 15+ installed and running
- Docker & Docker Compose (optional)

## Option 1: Quick Start with Docker Compose

1. Clone the repository:
```bash
git clone https://github.com/namkangwaan/docvault-go.git
cd docvault-go
```

2. Start the services:
```bash
docker-compose up -d
```

3. Wait for services to start (about 30 seconds), then check health:
```bash
curl http://localhost:8080/health
```

4. Create an admin user (see "Creating an Admin User" below)

5. Access the API at `http://localhost:8080`

## Option 2: Local Development Setup

1. Clone and setup:
```bash
git clone https://github.com/namkangwaan/docvault-go.git
cd docvault-go
cp .env.example .env
```

2. Edit `.env` and configure settings

3. Install dependencies:
```bash
go mod download
```

4. Start PostgreSQL (if not running):
```bash
docker run -d --name docvault-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=docvault \
  -p 5432:5432 \
  postgres:15-alpine
```

5. Run migrations:
```bash
psql -h localhost -U postgres -d docvault -f internal/database/migrations/001_initial_schema.up.sql
```

6. Start the application:
```bash
go run cmd/server/main.go
```

## Creating an Admin User

You need to create an admin user to access protected endpoints. Here are the options:

### Option 1: Using psql

```bash
# Generate password hash with Go
go run -e '
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("your_password"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
'

# Insert admin user with the generated hash
psql -h localhost -U postgres -d docvault << EOF
INSERT INTO admin_users (username, email, password_hash, is_active)
VALUES ('admin', 'admin@example.com', '\$2a\$10\$YOUR_GENERATED_HASH_HERE', true);
EOF
```

### Option 2: Using a Helper Script

Create `scripts/create-admin.go`:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run create-admin.go <username> <email> <password>")
		os.Exit(1)
	}

	username := os.Args[1]
	email := os.Args[2]
	password := os.Args[3]

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/docvault?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Insert admin user
	_, err = pool.Exec(context.Background(),
		"INSERT INTO admin_users (username, email, password_hash, is_active) VALUES ($1, $2, $3, true)",
		username, email, string(hash),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Admin user '%s' created successfully!\n", username)
}
```

Run it:
```bash
go run scripts/create-admin.go admin admin@example.com mypassword
```

## Testing the API

### 1. Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"mypassword"}'
```

Response:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com"
  }
}
```

### 2. Create a Category
```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8080/api/admin/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Policies",
    "is_active": true,
    "sort_order": 1
  }'
```

### 3. Create a Document
```bash
curl -X POST http://localhost:8080/api/admin/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1,
    "order_number": "DOC-001",
    "title": "Company Policy",
    "description": "Annual company policy document",
    "is_published": true
  }'
```

### 4. Upload a File
```bash
curl -X POST http://localhost:8080/api/admin/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "document_id=1" \
  -F "file=@/path/to/document.pdf"
```

### 5. List Documents (Public)
```bash
curl http://localhost:8080/api/documents
```

### 6. Download a File (Public)
```bash
curl -O http://localhost:8080/api/files/1/download
```

### 7. Resume Download with Range Header
```bash
# Download first 1MB
curl -H "Range: bytes=0-1048575" http://localhost:8080/api/files/1/download -o partial.pdf

# Resume from 1MB
curl -H "Range: bytes=1048576-" http://localhost:8080/api/files/1/download >> partial.pdf
```

## Security Features in Action

### Testing File Validation

Try uploading a malicious PDF with JavaScript:
```bash
# This will be rejected
curl -X POST http://localhost:8080/api/admin/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "document_id=1" \
  -F "file=@malicious.pdf"
```

Response will include security report:
```json
{
  "error": "file validation failed",
  "security_report": {
    "malware_scan_passed": false,
    "threats": ["PDF contains JavaScript or actions: /JavaScript"]
  }
}
```

## Chunked Upload for Large Files

For files larger than 100MB, use chunked upload:

### 1. Initialize Upload Session
```bash
curl -X POST http://localhost:8080/api/admin/files/upload/chunked/init \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "large-video.mp4",
    "total_size": 524288000,
    "chunk_size": 5242880,
    "document_id": 1
  }'
```

Response:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_chunks": 100
}
```

### 2. Upload Chunks
```bash
SESSION_ID="550e8400-e29b-41d4-a716-446655440000"

for i in {0..99}; do
  curl -X POST http://localhost:8080/api/admin/files/upload/chunked/chunk \
    -H "Authorization: Bearer $TOKEN" \
    -F "session_id=$SESSION_ID" \
    -F "chunk_index=$i" \
    -F "chunk=@chunk_$i.bin"
done
```

### 3. Check Progress
```bash
curl http://localhost:8080/api/admin/files/upload/chunked/status/$SESSION_ID \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Finalize Upload
```bash
curl -X POST http://localhost:8080/api/admin/files/upload/chunked/finalize \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "'$SESSION_ID'",
    "document_id": 1
  }'
```

## Monitoring

### Check Application Health
```bash
curl http://localhost:8080/health
```

### View Logs
```bash
# Docker Compose
docker-compose logs -f app

# Local
# Logs are output to stdout
```

## Troubleshooting

### Connection Refused
- Check if PostgreSQL is running
- Verify DATABASE_URL in .env
- Check firewall settings

### Permission Denied
- Ensure storage directories have correct permissions
- Check STORAGE_BASE_PATH exists and is writable

### JWT Token Invalid
- Token may have expired (default: 24h)
- Login again to get a new token

### File Upload Failed
- Check STORAGE_MAX_FILE_SIZE setting
- Verify file passes security validation
- Check available disk space

## Production Deployment

### Important Security Settings

1. **Change JWT Secret**:
```env
JWT_SECRET=use-a-strong-random-secret-at-least-32-characters
```

2. **Enable HTTPS** (use reverse proxy like Nginx)

3. **Set Strong Passwords** for admin accounts

4. **Configure Firewall** to restrict access

5. **Regular Backups**:
```bash
# Database backup
pg_dump -h localhost -U postgres docvault > backup.sql

# File storage backup
tar -czf storage-backup.tar.gz /data/documents
```

6. **Monitor Quarantine Directory**:
```bash
ls -la /data/quarantine
```

## Support

- GitHub Issues: https://github.com/namkangwaan/docvault-go/issues
- Documentation: See README.md

## License

MIT License - see LICENSE file
