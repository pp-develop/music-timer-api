DROP TABLE IF EXISTS users CASCADE;

create table users(
    "id" VARCHAR(255) primary key,
    "country" VARCHAR(255),
    "access_token" text,
    "refresh_token" text,
    "token_expiration" int,
    "session" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "playlist_count" INTEGER DEFAULT 0
);

DROP TABLE IF EXISTS tracks CASCADE;

create table tracks(
    "uri" VARCHAR(255) primary key,
    "duration_ms" int,
    "isrc" VARCHAR(255),
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP
);

DROP TABLE IF EXISTS playlists CASCADE;

create table playlists(
    "id" VARCHAR(255) PRIMARY KEY,
    index id_index (id),
    "user_id" VARCHAR(255),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
);

DROP TABLE IF EXISTS favorite_tracks CASCADE;

CREATE TABLE favorite_tracks (
    "user_id" VARCHAR(255) PRIMARY KEY,
    "tracks" JSONB,
    "updated_at" TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

DROP TABLE IF EXISTS artists CASCADE;

CREATE TABLE artists (
    "id" VARCHAR(255) PRIMARY KEY,
    index id_index (id),
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
    CONSTRAINT fk_jwt_refresh_token_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);