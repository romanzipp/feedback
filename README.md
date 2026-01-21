# Feedback

File sharing and commenting application built with Go.

> [!WARNING]
> This project is a complete vibe-coded hellhole of garbage. Read code at your own caution. Redirect frustration at [Anthropic](https://support.claude.com/en/articles/9015913-how-to-get-support).

## Features

- Admin panel for creating shares and uploading files
- Public share links with commenting functionality
- Image viewing in fullscreen modal
- Clean, Nextcloud-inspired design
- Mobile responsive layout
- SQLite database with WAL mode
- Docker deployment ready

## Quick Start

### Local Development

1. Install dependencies:

```bash
go mod download
npm install
```

1. Create `.env` file:

```bash
cp .env.example .env
# Edit .env and set secure values for ADMIN_TOKEN and SESSION_SECRET
```

1. Build Tailwind CSS:

```bash
npx tailwindcss -i web/static/css/input.css -o web/static/css/output.css --watch
```

1. Run the application:

```bash
go run cmd/feedback/main.go
```

1. Access admin panel:

```
http://localhost:8080/admin/{ADMIN_TOKEN}
```

### Docker

Build and run with Docker:

```bash
docker build -t feedback:latest .
docker run -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e ADMIN_TOKEN=your-secure-token \
  -e SESSION_SECRET=your-secure-secret \
  feedback:latest
```

Or use Docker Compose:

```bash
# Create .env file with ADMIN_TOKEN and SESSION_SECRET
docker-compose up
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 8080 |
| HOST | Server host | 0.0.0.0 |
| ADMIN_TOKEN | Admin authentication token (required) | - |
| SESSION_SECRET | Cookie signing secret (required) | - |
| DATA_DIR | Data storage directory | ./data |
| MAX_UPLOAD_SIZE | Max upload size in bytes | 52428800 (50MB) |
| DB_PATH | SQLite database path | ./data/feedback.db |

## Usage

### Admin Workflow

1. Access admin panel at `/admin/{ADMIN_TOKEN}`
2. Create a new share with name and description
3. Upload files to the share
4. Copy the public share link (`/share/{hash}`)
5. Share the link with users

### User Workflow

1. Access public share via `/share/{hash}`
2. Enter your name (stored in cookie)
3. View files and images
4. Click images to view in fullscreen modal
5. Post comments on files
6. See comments from other users in real-time

## Project Structure

```
feedback/
├── cmd/feedback/          # Application entrypoint
├── internal/              # Internal packages
│   ├── config/           # Configuration loading
│   ├── database/         # Database models and migrations
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   └── services/         # Business logic
├── web/                  # Frontend assets
│   ├── static/          # Static files (CSS, JS)
│   └── templates/       # HTML templates
├── data/                # Runtime data (gitignored)
├── Dockerfile           # Container build
└── docker-compose.yml   # Local development
```

## Development

### Build Production Binary

```bash
CGO_ENABLED=1 go build -o feedback cmd/feedback/main.go
```

### Build Tailwind CSS for Production

```bash
npx tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify
```

### Run Tests

```bash
go test ./...
```

## Deployment

### GitHub Container Registry

Push a semver tag (no v-prefix) to trigger automated build:

```bash
git tag 1.0.0
git push origin 1.0.0
```

The GitHub Actions workflow will:

- Build Docker image for amd64 and arm64
- Push to `ghcr.io/romanzipp/feedback:{version}`
- Tag as `latest`

Pull and run:

```bash
docker pull ghcr.io/romanzipp/feedback:latest
docker run -p 8080:8080 \
  -v /path/to/data:/data \
  -e ADMIN_TOKEN=xxx \
  -e SESSION_SECRET=xxx \
  ghcr.io/romanzipp/feedback:latest
```

## Security

- Admin token is high-entropy random string
- Session cookies are HMAC-signed
- All SQL queries use parameterized statements
- HTML templates auto-escape output
- File paths validated to prevent directory traversal
- Rate limiting on comment endpoints
- Upload size limits enforced

## License

MIT
