-- 003_create_code_versions_and_feedback.up.sql
-- Data flywheel foundation: code version history and user feedback

-- Code versions: every code change (generate or chat) creates a version
CREATE TABLE code_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    version_number INT NOT NULL DEFAULT 1,
    html_code TEXT NOT NULL,
    diff_from_previous TEXT,
    source VARCHAR(20) NOT NULL DEFAULT 'generate',  -- 'generate' | 'chat'
    trigger_message TEXT,  -- the user message that triggered this version (for chat)
    agent_pipeline JSONB,  -- snapshot of which agents ran and their configs
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_code_versions_session_id ON code_versions (session_id);
CREATE INDEX idx_code_versions_session_version ON code_versions (session_id, version_number);

-- User feedback: ratings and comments on generated code
CREATE TABLE user_feedback (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    code_version_id UUID REFERENCES code_versions(id) ON DELETE SET NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    feedback_text TEXT,
    feedback_type VARCHAR(30) NOT NULL DEFAULT 'general',  -- 'general' | 'visual' | 'interaction' | 'code_quality'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_feedback_session_id ON user_feedback (session_id);
CREATE INDEX idx_user_feedback_rating ON user_feedback (rating);
CREATE INDEX idx_user_feedback_type ON user_feedback (feedback_type);
