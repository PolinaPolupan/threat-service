CREATE TABLE threats (
    md5         VARCHAR(32) PRIMARY KEY
                CHECK (md5 ~ '^[a-fA-F0-9]{32}$'),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    file_name   VARCHAR(255),
    file_size   BIGINT CHECK (file_size >= 0),
    first_seen  TIMESTAMPTZ,
    last_seen   TIMESTAMPTZ,
    severity    VARCHAR(20)
                CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    sha256      VARCHAR(64)
                CHECK (sha256 IS NULL OR sha256 ~ '^[a-fA-F0-9]{64}$'),
    source      VARCHAR(255),
    tags        TEXT[],
    type        VARCHAR(20)
                CHECK (type IN (
                    'trojan', 'ransomware', 'worm', 'virus',
                    'spyware', 'rootkit', 'adware', 'other'
                )),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
