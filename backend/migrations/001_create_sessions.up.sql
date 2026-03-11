-- 001_create_sessions.up.sql
-- Core session tables with images and history split into separate tables

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT NOT NULL DEFAULT '',
    framework VARCHAR(50) NOT NULL DEFAULT '',
    design_state JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 minutes'
);

CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
CREATE INDEX idx_sessions_updated_at ON sessions (updated_at);

-- Session images (split from sessions to avoid large JSONB updates)
CREATE TABLE session_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL DEFAULT '',
    mime_type VARCHAR(100) NOT NULL DEFAULT '',
    base64_data TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_images_session_id ON session_images (session_id);
CREATE INDEX idx_session_images_sort_order ON session_images (session_id, sort_order);

-- Session history (split for efficient append + LIMIT queries)
CREATE TABLE session_history (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    entry_type VARCHAR(20) NOT NULL DEFAULT 'text',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_history_session_id ON session_history (session_id);
CREATE INDEX idx_session_history_created_at ON session_history (session_id, created_at DESC);

-- Shared previews
CREATE TABLE shared_previews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    short_code VARCHAR(16) NOT NULL UNIQUE,
    html_snapshot TEXT NOT NULL,
    view_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shared_previews_short_code ON shared_previews (short_code) WHERE is_active = TRUE;
CREATE INDEX idx_shared_previews_session_id ON shared_previews (session_id);
