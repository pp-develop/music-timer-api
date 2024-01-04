create table tracks(
    `uri` VARCHAR(255) primary key,
    `artists_name` json,
    `duration_ms` int,
    `isrc` VARCHAR(255),
    `created_at` datetime,
    `updated_at` datetime
);

create table artists(
    `user_id` VARCHAR(255),
    `name` text,
    FOREIGN KEY fk_user_id(user_id) REFERENCES users(id),
    UNIQUE `user_id_name_index` (user_id, `name`(255))
);

create table playlists(
    `id` VARCHAR(255),
    index id_index (id),
    `user_id` VARCHAR(255),
    FOREIGN KEY fk_user_id(user_id) REFERENCES users(id)
);

create table users(
    `id` VARCHAR(255) primary key,
    `access_token` text,
    `refresh_token` text,
    `token_expiration` int,
    `session` VARCHAR(255),
    `created_at` datetime,
    `updated_at` datetime
);