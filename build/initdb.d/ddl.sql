DROP TABLE IF EXISTS users CASCADE;
create table users(
    "id" VARCHAR(255) primary key,
    "access_token" text,
    "refresh_token" text,
    "token_expiration" int,
    "session" VARCHAR(255),
    "favorite_track" jsonb,
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP
);

DROP TABLE IF EXISTS tracks CASCADE;
create table tracks(
    "uri" VARCHAR(255) primary key,
    "artists_id" jsonb,
    "artists_name" jsonb,
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