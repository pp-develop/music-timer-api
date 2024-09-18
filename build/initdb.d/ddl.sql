DROP TABLE IF EXISTS users CASCADE;
create table users(
    "id" VARCHAR(255) primary key,
    "country" VARCHAR(255),
    "access_token" text,
    "refresh_token" text,
    "token_expiration" int,
    "session" VARCHAR(255),
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

DROP TABLE IF EXISTS favorite_tracks CASCADE;
CREATE TABLE favorite_tracks (
    "user_id" VARCHAR(255) PRIMARY KEY,   
    "tracks" JSONB,                  
    "updated_at" TIMESTAMP,                 
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);