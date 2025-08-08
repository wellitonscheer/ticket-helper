# Ticket Helper

A smart ticket management system that helps you capture, organize, and find support tickets more efficiently. Instead of manually sorting through hundreds of tickets, this tool uses AI to understand what your tickets are about and helps you find similar issues instantly.

## What It Does

**Intelligent Search**: Find related tickets by describing what you're looking for in plain English - no need to remember exact keywords or ticket numbers.

**Email Notifications**: Get notified when similar tickets are found or when action is needed.

**Secure Access**: Only authorized team members can access the system through email verification.

## Why Use This?

- **Save Time**: Stop manually searching through old tickets - just describe the issue and find similar cases instantly
- **Better Support**: Quickly find how similar problems were solved before
- **Team Efficiency**: Share knowledge across your support team automatically
- **Easy Capture**: Grab tickets from any website without copy-pasting

## Prerequisites

- Go 1.23.2 or later
- Docker (for PostgreSQL and embedding service)
- Chrome browser (for extension)
- Make (for build automation)

## Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/wellitonscheer/ticket-helper.git
cd ticket-helper
```

### 2. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your specific configuration:

```bash
# Network Configuration
MY_IP=192.168.0.5
BASE_URL=127.0.0.1
APP_ENV=development
GIN_PORT=8080

# Authentication
VERIFIC_CODE_LIFETIME=900
SESSION_LIFETIME=10800
AUTH_EMAILS_PATH=./data_source/authorized_emails.json

# Embedding Service
EMBED_PORT=5000
EMBED_CONTAINER_NAME=embedding-endpoint

# Email Configuration (SMTP)
EMAIL_SERVER_USER=your_smtp_user
EMAIL_SERVER_PASSWORD=your_smtp_password
EMAIL_SERVER_HOST=your_smtp_host
EMAIL_SERVER_PORT=587
EMAIL_FROM=your_email@domain.com

# PostgreSQL Configuration
POSTGRES_CONTAINER_NAME=postgres
POSTGRES_USER=postgres
POSTGRES_DB=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_PORT=5432
```

### 3. Initial Setup

Run the setup command to install dependencies and prepare scripts:

```bash
make setup
```

This will:

- Install Air for hot reloading
- Make shell scripts executable
- Set up development dependencies

### 4. Start Development Environment

```bash
make dev
```

This command will:

- Start PostgreSQL with pgvector extension
- Launch the embedding service
- Start the Go application with hot reload

The application will be available at `http://localhost:8080`

## Development

### Available Make Commands

```bash
make help          # Display available commands
make setup         # Install dependencies and prepare environment
make dev           # Start development environment with hot reload
```

### Project Structure

```
ticket-helper/
├── cmd/app/           # Application entry point
├── internal/          # Internal application code
├── web/              # Web assets and templates
│   ├── static/       # CSS, JS, images
│   └── templates/    # HTML templates
├── chrome_extension/ # Chrome extension files
├── data_source/      # Data files and configurations
├── scripts/          # Utility scripts
├── .env.example      # Environment configuration template
├── Makefile          # Build automation
└── README.md         # This file
```

### Database Migrations

The application automatically runs database migrations on startup:

- SQLite migrations for application data
- PostgreSQL migrations for vector storage

### Hot Reload

The development environment uses [Air](https://github.com/air-verse/air) for hot reloading. Configuration is in `.air.toml`.

## Chrome Extension

### Installation

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the `chrome_extension` directory

### Usage

The extension provides browser integration for ticket capture and management. It includes:

- Content script injection
- Background service worker
- Popup interface
- Omnibox integration (keyword: "api")

## Email Configuration

The application supports email notifications through SMTP. Configure your email settings in the `.env` file:

- Use your email provider's SMTP settings
- For Gmail, you may need to use App Passwords
- For AWS SES, use your SES credentials

## Authorized Users

Add authorized email addresses to `./data_source/authorized_emails.json` to control access to the application.

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8080, 5000, and 5432 are available
2. **Database connection**: Verify PostgreSQL is running and accessible
3. **Environment variables**: Check that all required variables are set in `.env`
4. **Permissions**: Ensure shell scripts have execute permissions

### Logs

- Application logs are displayed in the terminal when running `make dev`
- Check Docker logs for database and embedding service issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the terms specified in the repository.
