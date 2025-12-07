DROP TABLE IF EXISTS spotify_users CASCADE;

CREATE TABLE spotify_users (
    "id" VARCHAR(255) PRIMARY KEY,
    "country" VARCHAR(255),
    "access_token" TEXT,
    "refresh_token" TEXT,
    "token_expiration" INT,
    "session" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "playlist_count" INTEGER DEFAULT 0
);

DROP TABLE IF EXISTS spotify_tracks CASCADE;

-- Row-Level TTL (CockroachDB機能):
-- - updated_atから180日経過した行を自動削除
-- - 毎日AM3時(UTC)にバックグラウンドで実行
-- - バッチサイズ1000行ずつ処理（MVCC負荷軽減のため）
-- - 参考: https://www.cockroachlabs.com/docs/stable/row-level-ttl
CREATE TABLE spotify_tracks (
    "uri" VARCHAR(255) PRIMARY KEY,
    "duration_ms" INT,
    "isrc" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    INDEX idx_spotify_tracks_updated_at (updated_at ASC)
) WITH (
    ttl_expiration_expression = '(updated_at::TIMESTAMPTZ + INTERVAL ''180 days'')',
    ttl_job_cron = '0 3 * * *',
    ttl_delete_batch_size = 1000,
    ttl_select_batch_size = 1000
);

DROP TABLE IF EXISTS spotify_playlists CASCADE;

CREATE TABLE spotify_playlists (
    "id" VARCHAR(255) PRIMARY KEY,
    INDEX id_index (id),
    "user_id" VARCHAR(255),
    CONSTRAINT fk_spotify_playlist_user FOREIGN KEY (user_id) REFERENCES spotify_users(id)
);

DROP TABLE IF EXISTS spotify_favorite_tracks CASCADE;

CREATE TABLE spotify_favorite_tracks (
    "user_id" VARCHAR(255) PRIMARY KEY,
    "tracks" JSONB,
    "updated_at" TIMESTAMP,
    CONSTRAINT fk_spotify_favorite_tracks_user FOREIGN KEY (user_id) REFERENCES spotify_users(id)
);

DROP TABLE IF EXISTS spotify_artists CASCADE;

CREATE TABLE spotify_artists (
    "id" VARCHAR(255) PRIMARY KEY,
    INDEX id_index (id),
    "tracks" JSONB,
    "updated_at" TIMESTAMP
);


DROP TABLE IF EXISTS spotify_jwt_refresh_tokens CASCADE;

CREATE TABLE spotify_jwt_refresh_tokens (
    "jti" VARCHAR(255) PRIMARY KEY,
    "user_id" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP NOT NULL,
    "revoked" BOOLEAN DEFAULT FALSE,
    INDEX idx_spotify_jwt_user_id (user_id),
    INDEX idx_spotify_jwt_expires_at (expires_at),
    CONSTRAINT fk_spotify_jwt_refresh_token_user FOREIGN KEY (user_id) REFERENCES spotify_users(id) ON DELETE CASCADE
);


-- SoundCloud Tables

DROP TABLE IF EXISTS soundcloud_users CASCADE;

CREATE TABLE soundcloud_users (
    "id" VARCHAR(255) PRIMARY KEY,
    "username" VARCHAR(255),
    "access_token" TEXT,
    "refresh_token" TEXT,
    "token_expiration" INT,
    "session" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "playlist_count" INTEGER DEFAULT 0
);

DROP TABLE IF EXISTS soundcloud_favorite_tracks CASCADE;

CREATE TABLE soundcloud_favorite_tracks (
    "user_id" VARCHAR(255) PRIMARY KEY,
    "tracks" JSONB,
    "updated_at" TIMESTAMP,
    CONSTRAINT fk_soundcloud_favorite_tracks_user FOREIGN KEY (user_id) REFERENCES soundcloud_users(id)
);

DROP TABLE IF EXISTS soundcloud_playlists CASCADE;

CREATE TABLE soundcloud_playlists (
    "id" VARCHAR(255) PRIMARY KEY,
    INDEX id_index (id),
    "user_id" VARCHAR(255),
    CONSTRAINT fk_soundcloud_playlist_user FOREIGN KEY (user_id) REFERENCES soundcloud_users(id)
);

DROP TABLE IF EXISTS soundcloud_artists CASCADE;

CREATE TABLE soundcloud_artists (
    "id" VARCHAR(255) PRIMARY KEY,
    INDEX id_index (id),
    "tracks" JSONB,
    "updated_at" TIMESTAMP
);

DROP TABLE IF EXISTS soundcloud_jwt_refresh_tokens CASCADE;

CREATE TABLE soundcloud_jwt_refresh_tokens (
    "jti" VARCHAR(255) PRIMARY KEY,
    "user_id" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP NOT NULL,
    "revoked" BOOLEAN DEFAULT FALSE,
    INDEX idx_soundcloud_jwt_user_id (user_id),
    INDEX idx_soundcloud_jwt_expires_at (expires_at),
    CONSTRAINT fk_soundcloud_jwt_refresh_token_user FOREIGN KEY (user_id) REFERENCES soundcloud_users(id) ON DELETE CASCADE
);