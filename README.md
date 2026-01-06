# DocVault - Document Management System

A secure document management system built with Go, PostgreSQL, and NFS storage. Features comprehensive security validation, chunked file uploads, and HTTP Range request support for resume downloads.

## Features

### Public Features
- 🔍 Search documents by title, order number, and date
- 📂 Filter by category with sidebar navigation
- 📄 Display documents in table format
- 👁️ Preview files before download (PDF, DOCX, XLSX)
- ⬇️ Download files with resume support (HTTP Range)

### Admin Features
- 🔐 Login/Logout with JWT authentication
- ✏️ CRUD operations for documents and categories
- 📤 Upload files with streaming support
- 🔄 Chunked upload for large files
- 📎 Manage multiple file attachments per document

### File Streaming
- 🔧 Buffer pool management to reduce GC pressure
- 📦 Chunked upload/download
- ⚡ HTTP Range request support (resume download)
- ✅ SHA-256 checksum verification

### Security Features

#### Layer 1: Magic Bytes Validation
- Validates file signatures (magic numbers)
- Prevents file extension spoofing
- Supports PDF, DOCX, XLSX, and other Office formats

#### Layer 2: MIME Type Verification
- Deep MIME detection (doesn't trust Content-Type header)
- Cross-checks with magic bytes

#### Layer 3: Malicious Content Scanner
- JavaScript/VBScript detection in PDFs
- VBA macro detection in Office documents
- Embedded executable detection
- Suspicious pattern detection
- PowerShell and shell script detection

#### Layer 4: Document Structure Validation
- PDF structure integrity check
- OOXML (Office) structure validation
- ZIP bomb detection (compression ratio > 100:1)

#### Layer 5: Antivirus Scan (Optional)
- ClamAV integration
- Real-time virus scanning

#### Security Actions
- Automatically rejects files that fail validation
- Quarantines suspicious files
- Logs detailed scan reports

## Tech Stack

- **Backend**: Go (Golang) + Fiber Framework
- **Database**: PostgreSQL
- **Storage**: NFS (Network File System) - Local/Network Filesystem
- **Authentication**: JWT

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker & Docker Compose (optional, for containerized deployment)

## Installation

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/namkangwaan/docvault-go.git
cd docvault-go
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Edit `.env` file and configure your settings:
```bash
# Update database connection
DATABASE_URL=postgres://postgres:postgres@localhost:5432/docvault?sslmode=disable

# Update JWT secret (important for production!)
JWT_SECRET=your-super-secret-key-change-in-production

# Configure storage paths
STORAGE_BASE_PATH=/data/documents
```

4. Install dependencies:
```bash
make deps
```

5. Run database migrations:
```bash
# Start PostgreSQL (if using Docker)
docker run -d --name docvault-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=docvault \
  -p 5432:5432 \
  postgres:15-alpine

# Run migrations
psql -h localhost -U postgres -d docvault -f internal/database/migrations/001_initial_schema.up.sql
```

6. Run the application:
```bash
make run
```

The application will start on `http://localhost:8080`

### Docker Deployment

1. Clone the repository:
```bash
git clone https://github.com/namkangwaan/docvault-go.git
cd docvault-go
```

2. Start with Docker Compose:
```bash
make docker-up
```

This will start both the application and PostgreSQL database.

3. View logs:
```bash
make docker-logs
```

4. Stop containers:
```bash
make docker-down
```

## API Documentation

### Health Check

```
GET /health
```

Returns server health status.

### Public Endpoints

#### List Documents
```
GET /api/documents?limit=30&offset=0&search=query&category_id=1
```

Query Parameters:
- `limit` (optional): Number of results per page (default: 30, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `search` (optional): Search query for title/order number
- `category_id` (optional): Filter by category
- `order_number` (optional): Filter by order number
- `date_from` (optional): Filter by effective date (YYYY-MM-DD)
- `date_to` (optional): Filter by effective date (YYYY-MM-DD)

#### Get Document Details
```
GET /api/documents/:id
```

Returns document with files and category information.

#### List Categories
```
GET /api/categories
```

Returns all active categories with document counts.

#### Download File
```
GET /api/files/:id/download
```

Downloads file with HTTP Range support for resume.

Headers:
- `Range: bytes=0-1023` (optional): Request specific byte range

#### Preview File
```
GET /api/files/:id/preview
```

Serves file inline for preview in browser.

### Authentication

#### Login
```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

Returns:
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

#### Logout
```
POST /api/auth/logout
Authorization: Bearer <token>
```

### Admin Endpoints

All admin endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <token>
```

#### Create Document
```
POST /api/admin/documents
Content-Type: application/json

{
  "category_id": 1,
  "order_number": "ORDER-001",
  "title": "Document Title",
  "description": "Document description",
  "effective_date": "2024-01-01",
  "is_published": true
}
```

#### Update Document
```
PUT /api/admin/documents/:id
Content-Type: application/json

{
  "category_id": 1,
  "order_number": "ORDER-001",
  "title": "Updated Title",
  "is_published": true
}
```

#### Delete Document
```
DELETE /api/admin/documents/:id
```

#### Upload File (Single)
```
POST /api/admin/files/upload
Content-Type: multipart/form-data

document_id: 1
file: <file>
```

Returns file information and security scan report.

#### Chunked Upload

1. Initialize session:
```
POST /api/admin/files/upload/chunked/init
Content-Type: application/json

{
  "file_name": "large-file.pdf",
  "total_size": 104857600,
  "chunk_size": 5242880,
  "document_id": 1
}
```

2. Upload chunks:
```
POST /api/admin/files/upload/chunked/chunk
Content-Type: multipart/form-data

session_id: <session_id>
chunk_index: 0
chunk: <chunk_data>
```

3. Finalize upload:
```
POST /api/admin/files/upload/chunked/finalize
Content-Type: application/json

{
  "session_id": "<session_id>",
  "document_id": 1
}
```

4. Check status:
```
GET /api/admin/files/upload/chunked/status/:session_id
```

#### Delete File
```
DELETE /api/admin/files/:id
```

## Database Schema

### Categories Table
- `id`: Primary key
- `name`: Category name
- `slug`: URL-friendly slug
- `parent_id`: Parent category (for hierarchical structure)
- `sort_order`: Display order
- `is_active`: Active status
- `created_at`, `updated_at`: Timestamps

### Documents Table
- `id`: Primary key
- `category_id`: Foreign key to categories
- `order_number`: Document order/reference number
- `title`: Document title
- `description`: Document description
- `effective_date`: Date when document becomes effective
- `is_published`: Publication status
- `view_count`: Number of views
- `download_count`: Number of downloads
- `created_at`, `updated_at`: Timestamps

### Files Table
- `id`: Primary key
- `document_id`: Foreign key to documents
- `file_name`: Stored filename
- `original_name`: Original filename
- `mime_type`: MIME type
- `file_extension`: File extension
- `file_size`: Size in bytes
- `storage_path`: Path on storage
- `checksum`: SHA-256 checksum
- `chunk_size`, `total_chunks`: For chunked uploads
- `file_category`: File category (attachment, etc.)
- `sort_order`: Display order
- `preview_generated`, `preview_path`: Preview information
- `created_at`, `updated_at`: Timestamps

### Admin Users Table
- `id`: Primary key
- `username`: Unique username
- `email`: Unique email
- `password_hash`: Bcrypt password hash
- `is_active`: Active status
- `last_login_at`: Last login timestamp
- `created_at`, `updated_at`: Timestamps

### Audit Logs Table
- `id`: Primary key
- `user_id`: Foreign key to admin users
- `action`: Action performed
- `entity_type`: Type of entity
- `entity_id`: ID of entity
- `old_values`, `new_values`: JSONB fields for change tracking
- `ip_address`: Request IP address
- `created_at`: Timestamp

## Configuration

All configuration is done through environment variables. See `.env.example` for all available options.

### Key Configuration Options

- `SERVER_PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `STORAGE_BASE_PATH`: Base path for file storage
- `STORAGE_MAX_FILE_SIZE`: Maximum file size in bytes
- `JWT_SECRET`: Secret key for JWT (change in production!)
- `JWT_EXPIRY`: Token expiration time
- `ENABLE_MAGIC_BYTE_CHECK`: Enable magic bytes validation
- `ENABLE_MIME_CHECK`: Enable MIME type verification
- `ENABLE_MALWARE_SCANNER`: Enable malware scanning
- `ENABLE_DOCUMENT_CHECK`: Enable document structure validation
- `ENABLE_ANTIVIRUS_SCAN`: Enable ClamAV scanning
- `ENABLE_QUARANTINE`: Enable file quarantine

## Development

### Build
```bash
make build
```

### Run
```bash
make run
```

### Test
```bash
make test
```

### Test with Coverage
```bash
make test-coverage
```

### Format Code
```bash
make fmt
```

### Clean Build Artifacts
```bash
make clean
```

## Security Best Practices

1. **Always change the JWT secret** in production
2. **Enable all security layers** for maximum protection
3. **Use HTTPS** in production
4. **Set appropriate file size limits**
5. **Regularly update dependencies**
6. **Monitor quarantine directory** for suspicious files
7. **Review security scan logs** regularly
8. **Use strong passwords** for admin accounts

## Creating an Admin User

Connect to PostgreSQL and run:

```sql
INSERT INTO admin_users (username, email, password_hash, is_active)
VALUES (
  'admin',
  'admin@example.com',
  '$2a$10$example_hash_here', -- Use bcrypt to hash your password
  true
);
```

Or use Go code to generate password hash:
```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "your_password"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
```

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please open an issue on GitHub.
