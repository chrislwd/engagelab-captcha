-- EngageLab CAPTCHA - Initial Database Schema
-- PostgreSQL 15+

-- Enable UUID generation.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Tenants
-- ============================================================
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    api_key     VARCHAR(255) NOT NULL UNIQUE,
    plan        VARCHAR(50)  NOT NULL DEFAULT 'free'
                CHECK (plan IN ('free', 'pro', 'enterprise')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_api_key ON tenants (api_key);

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL DEFAULT '',
    role          VARCHAR(50)  NOT NULL DEFAULT 'member'
                  CHECK (role IN ('admin', 'member', 'viewer')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_tenant_id ON users (tenant_id);
CREATE INDEX idx_users_email ON users (email);

-- ============================================================
-- Apps
-- ============================================================
CREATE TABLE apps (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255)  NOT NULL,
    site_key        VARCHAR(255)  NOT NULL UNIQUE,
    secret_key      VARCHAR(255)  NOT NULL UNIQUE,
    allowed_domains TEXT[]        NOT NULL DEFAULT '{}',
    status          VARCHAR(50)   NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive')),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_apps_tenant_id ON apps (tenant_id);
CREATE INDEX idx_apps_site_key ON apps (site_key);
CREATE INDEX idx_apps_secret_key ON apps (secret_key);

-- ============================================================
-- Policies
-- ============================================================
CREATE TABLE policies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scene_type      VARCHAR(50)   NOT NULL
                    CHECK (scene_type IN ('register', 'login', 'activity', 'comment', 'api')),
    threshold_low   DOUBLE PRECISION NOT NULL DEFAULT 20,
    threshold_high  DOUBLE PRECISION NOT NULL DEFAULT 60,
    action_low      VARCHAR(50)   NOT NULL DEFAULT 'pass'
                    CHECK (action_low IN ('pass', 'invisible')),
    action_mid      VARCHAR(50)   NOT NULL DEFAULT 'challenge'
                    CHECK (action_mid IN ('challenge', 'invisible')),
    action_high     VARCHAR(50)   NOT NULL DEFAULT 'deny'
                    CHECK (action_high IN ('challenge', 'deny')),
    ip_whitelist    TEXT[]        NOT NULL DEFAULT '{}',
    ip_blacklist    TEXT[]        NOT NULL DEFAULT '{}',
    rate_limit_rpm  INTEGER       NOT NULL DEFAULT 30,
    rate_limit_rph  INTEGER       NOT NULL DEFAULT 300,
    enabled         BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policies_scene_type ON policies (scene_type);

-- ============================================================
-- Scenes
-- ============================================================
CREATE TABLE scenes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id      UUID         NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    scene_type  VARCHAR(50)  NOT NULL
                CHECK (scene_type IN ('register', 'login', 'activity', 'comment', 'api')),
    policy_id   UUID         REFERENCES policies(id) ON DELETE SET NULL,
    status      VARCHAR(50)  NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'inactive')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scenes_app_id ON scenes (app_id);
CREATE INDEX idx_scenes_policy_id ON scenes (policy_id);

-- ============================================================
-- Challenge Sessions
-- ============================================================
CREATE TABLE challenge_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id          UUID            NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    scene_id        UUID            REFERENCES scenes(id) ON DELETE SET NULL,
    session_id      VARCHAR(255)    NOT NULL,
    ip              INET,
    ua_hash         VARCHAR(64),
    fingerprint_id  VARCHAR(255),
    challenge_type  VARCHAR(50)     NOT NULL
                    CHECK (challenge_type IN ('invisible', 'slider', 'click', 'puzzle')),
    risk_score      DOUBLE PRECISION NOT NULL DEFAULT 0,
    risk_label      TEXT            NOT NULL DEFAULT '',
    status          VARCHAR(50)     NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'passed', 'failed', 'expired')),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ     NOT NULL DEFAULT (NOW() + INTERVAL '5 minutes')
);

CREATE INDEX idx_challenge_sessions_app_id ON challenge_sessions (app_id);
CREATE INDEX idx_challenge_sessions_scene_id ON challenge_sessions (scene_id);
CREATE INDEX idx_challenge_sessions_session_id ON challenge_sessions (session_id);
CREATE INDEX idx_challenge_sessions_status ON challenge_sessions (status);
CREATE INDEX idx_challenge_sessions_created_at ON challenge_sessions (created_at DESC);
CREATE INDEX idx_challenge_sessions_ip ON challenge_sessions (ip);

-- ============================================================
-- Verification Results
-- ============================================================
CREATE TABLE verification_results (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenge_id  UUID            NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    verified      BOOLEAN         NOT NULL DEFAULT FALSE,
    score         DOUBLE PRECISION NOT NULL DEFAULT 0,
    labels        TEXT[]          NOT NULL DEFAULT '{}',
    action        VARCHAR(50)    NOT NULL DEFAULT '',
    reason_code   VARCHAR(100)   NOT NULL DEFAULT '',
    completed_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_results_challenge_id ON verification_results (challenge_id);

-- ============================================================
-- Event Feedback
-- ============================================================
CREATE TABLE event_feedback (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenge_id   UUID           NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    feedback_type  VARCHAR(50)    NOT NULL
                   CHECK (feedback_type IN ('false_positive', 'false_negative', 'abuse', 'other')),
    comment        TEXT           NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_feedback_challenge_id ON event_feedback (challenge_id);

-- ============================================================
-- Trigger: auto-update updated_at columns
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_apps_updated_at
    BEFORE UPDATE ON apps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_scenes_updated_at
    BEFORE UPDATE ON scenes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Seed data
-- ============================================================
INSERT INTO tenants (id, name, api_key, plan) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Demo Corp', 'ak_demo_key_123456', 'pro');

INSERT INTO users (id, tenant_id, email, password_hash, name, role) VALUES
    ('00000000-0000-0000-0000-000000000002',
     '00000000-0000-0000-0000-000000000001',
     'admin@democorp.com',
     crypt('demo_password', gen_salt('bf')),
     'Admin User',
     'admin');

INSERT INTO apps (id, tenant_id, name, site_key, secret_key, allowed_domains) VALUES
    ('00000000-0000-0000-0000-000000000003',
     '00000000-0000-0000-0000-000000000001',
     'Demo Website',
     'sk_demo_site_key_abc',
     'sec_demo_secret_key_xyz',
     ARRAY['localhost', 'demo.engagelab.cc']);

INSERT INTO policies (id, scene_type, threshold_low, threshold_high, action_low, action_mid, action_high, rate_limit_rpm, rate_limit_rph) VALUES
    ('00000000-0000-0000-0000-000000000004', 'login', 20, 60, 'pass', 'challenge', 'deny', 30, 300),
    ('00000000-0000-0000-0000-000000000005', 'register', 15, 50, 'invisible', 'challenge', 'deny', 10, 60);

INSERT INTO scenes (id, app_id, scene_type, policy_id) VALUES
    ('00000000-0000-0000-0000-000000000006',
     '00000000-0000-0000-0000-000000000003',
     'login',
     '00000000-0000-0000-0000-000000000004'),
    ('00000000-0000-0000-0000-000000000007',
     '00000000-0000-0000-0000-000000000003',
     'register',
     '00000000-0000-0000-0000-000000000005');
