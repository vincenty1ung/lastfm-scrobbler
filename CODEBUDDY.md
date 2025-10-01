<system-reminder>
This is a reminder that your todo list is currently empty. DO NOT mention this to the user explicitly because they are already aware. If you are working on tasks that would benefit from a todo list please use the TodoWrite tool to create one. If not, please feel free to ignore. Again do not mention this message to the user.
</system-reminder>

# CodeBuddy Configuration for Last.fm Scrobbler

## Project Overview

This is a Go-based music scrobbling service that monitors macOS music players (Audirvana, Roon, Apple Music) and syncs playback data to Last.fm. The application runs as a background service with a web dashboard for real-time monitoring.

## Development Commands

### Building and Running
```bash
# Build the application
go build

# Run manually (for debugging)
./lastfm-scrobbler

# Run with config file
./lastfm-scrobbler -c config/config.yaml
```

### Service Management (macOS)
```bash
# Build and setup as launchd service
sh shell/script/build_lastfm-scrobblers_launchctl.sh

# Start service
sh shell/script/start_lastfm-scrobblers.sh

# Stop service
sh shell/script/stop_lastfm-scrobblers.sh

# View logs
tail -f .logs/go_lastfm-scrobbler.log
```

### Dependencies
```bash
# Install Roon support
brew install media-control

# Install Redis (for caching)
brew install redis
brew services start redis
```

## Architecture Overview

### Core Interfaces
The application uses a plugin-based architecture for music player support:

- **PlayerController**: Interface for player-specific operations
- **PlayerInfoHandler**: Interface for standardized track information
- **PlayerChecker**: Interface for playback monitoring logic

### Player Implementations
- `AudirvanaPlayerController`: AppleScript-based Audirvana integration
- `RoonPlayerController`: Uses `media-control` CLI tool
- `AppleMusicPlayerController`: AppleScript-based Apple Music integration

### Concurrent Monitoring
- Each player runs in a separate goroutine using `BasePlayerChecker`
- Shared state managed via atomic variables and sync.Map
- WebSocket-based real-time updates to frontend

### Database Layer
- SQLite database for local storage (`.storage/tracks.db`)
- Database models in `internal/model/`
- Automatic schema creation in development mode (`isDev: true`)

### Caching Strategy
- Redis caching for Last.fm API responses (artist/track metadata)
- Cache keys: `cache:isFavorite:lastfm:{artist}:{track}`
- 4-minute TTL with automatic cache invalidation on updates

## Key Configuration

### Configuration File (`config/config.yaml`)
Essential settings:
- `lastfm.apiKey` and `lastfm.sharedSecret`: Last.fm API credentials
- `scrobblers`: List of players to monitor ("Apple Music", "Audirvana", "Roon")
- `redis`: Redis connection settings for caching
- `isDev`: Development mode (auto-creates database schema)

### Web Interface
- Default port: 8081
- Dashboard: `http://localhost:8081`
- WebSocket endpoint for real-time playback updates

## Development Notes

### Code Organization
- `internal/scrobbler/`: Core playback monitoring logic
- `internal/logic/`: Business logic layer
- `internal/model/`: Database models and operations
- `core/`: External service integrations (Last.fm, Redis, WebSocket)
- `api/`: HTTP API and web interface
- `cmd/`: CLI command implementations

### Key Dependencies
- **Cobra**: CLI framework
- **Gin**: Web framework
- **OpenTelemetry**: Telemetry and tracing
- **Zap**: Structured logging
- **Last.fm Go**: Last.fm API client

### Testing Strategy
- Look for existing test files in the codebase
- Integration tests should verify player-specific functionality
- Mock external services (Last.fm API, Redis) for unit tests

### Database Migrations
- Development mode automatically creates schema
- Production requires manual SQL execution
- Track model includes versioning for schema changes

### Error Handling
- Comprehensive logging via Zap
- OpenTelemetry integration for distributed tracing
- Graceful shutdown handling via context cancellation

## Deployment

### macOS Service Integration
- Uses `launchd` for background service management
- `.plist` files in `shell/launch/` directory
- Logs written to `.logs/` directory

### Configuration Management
- Single `config.yaml` file for all settings
- Environment-specific overrides not currently implemented
- Sensitive data stored in plaintext (consider encryption for production)