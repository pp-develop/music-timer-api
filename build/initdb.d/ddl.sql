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

CREATE TABLE spotify_tracks (
    "uri" VARCHAR(255) PRIMARY KEY,
    "duration_ms" INT,
    "isrc" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP
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


DROP TABLE IF EXISTS jwt_refresh_tokens CASCADE;

CREATE TABLE jwt_refresh_tokens (
    "jti" VARCHAR(255) PRIMARY KEY,
    "user_id" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP NOT NULL,
    "revoked" BOOLEAN DEFAULT FALSE,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    CONSTRAINT fk_jwt_refresh_token_user FOREIGN KEY (user_id) REFERENCES spotify_users(id) ON DELETE CASCADE
);


-- =====================================================
-- YouTube Music Tables
-- =====================================================

DROP TABLE IF EXISTS ytmusic_users CASCADE;

CREATE TABLE ytmusic_users (
    "id" VARCHAR(255) PRIMARY KEY,
    "email" VARCHAR(255),
    "access_token" TEXT,
    "refresh_token" TEXT,
    "token_expiration" BIGINT,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS ytmusic_favorite_tracks CASCADE;

CREATE TABLE ytmusic_favorite_tracks (
    "user_id" VARCHAR(255) PRIMARY KEY,
    "tracks" JSONB,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ytmusic_favorite_tracks_user FOREIGN KEY (user_id) REFERENCES ytmusic_users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS ytmusic_playlists CASCADE;

CREATE TABLE ytmusic_playlists (
    "id" VARCHAR(255) PRIMARY KEY,
    INDEX id_index (id),
    "user_id" VARCHAR(255),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ytmusic_playlist_user FOREIGN KEY (user_id) REFERENCES ytmusic_users(id)
);