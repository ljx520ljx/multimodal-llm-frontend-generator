## MODIFIED Requirements

### Requirement: Session Persistence
The system SHALL persist all session data (images, history, generated code, design state) in PostgreSQL instead of in-memory storage. Session data SHALL survive server restarts.

#### Scenario: Session survives restart
- **WHEN** the server restarts
- **THEN** all active sessions SHALL be recoverable from PostgreSQL

#### Scenario: Session expiration
- **WHEN** a session has not been accessed for longer than the configured TTL
- **THEN** the system SHALL automatically clean up the session and associated data

#### Scenario: Concurrent session access
- **WHEN** multiple requests access the same session concurrently
- **THEN** the system SHALL ensure data consistency through database-level transactions

### Requirement: Session Store Interface
The system SHALL define a SessionStore interface that abstracts the storage backend. Both MemoryStore (for testing) and PostgresStore (for production) SHALL implement this interface.

#### Scenario: Storage backend configuration
- **WHEN** the DATABASE_URL environment variable is set
- **THEN** the system SHALL use PostgresStore as the session backend

#### Scenario: Fallback to memory store
- **WHEN** the DATABASE_URL environment variable is not set
- **THEN** the system SHALL fall back to MemoryStore with a warning log
