-- Users
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);

-- User-Tenant roles
CREATE TABLE user_tenants (
    user_id    UUID NOT NULL REFERENCES users(id),
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    role       VARCHAR(30) NOT NULL DEFAULT 'developer',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, tenant_id)
);

-- API Keys
CREATE TABLE api_keys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    name       VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    key_hash   VARCHAR(64) NOT NULL UNIQUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Threat events
CREATE TABLE threat_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    type         VARCHAR(50) NOT NULL,
    severity     VARCHAR(20) NOT NULL,
    source_ip    VARCHAR(100),
    target_scene VARCHAR(50),
    details      TEXT,
    event_count  INTEGER NOT NULL DEFAULT 0,
    first_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen    TIMESTAMPTZ NOT NULL DEFAULT now(),
    status       VARCHAR(20) NOT NULL DEFAULT 'active'
);

CREATE INDEX idx_threats_tenant ON threat_events(tenant_id);
CREATE INDEX idx_threats_status ON threat_events(status);

-- Webhook subscriptions
CREATE TABLE webhook_subscriptions (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id    UUID NOT NULL REFERENCES apps(id),
    url       TEXT NOT NULL,
    secret    VARCHAR(128),
    events    TEXT[] NOT NULL DEFAULT '{}',
    active    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
