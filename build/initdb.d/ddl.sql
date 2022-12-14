create table tracks(
    `uri` VARCHAR(255) primary key,
    `duration_ms` int,
    `isrc` VARCHAR(255),
    `created_at` datetime,
    `updated_at` datetime
);

create table playlists(
    `id` VARCHAR(255),
    index id_index (id),
);
-- create table playlist(
--     `id` int not null auto_increment primary key,
--     `msec` int
-- );

create table users(
    `id` VARCHAR(255) primary key,
    `access_token` text,
    `refresh_token` text,
    `token_expiration` int,
    `session` VARCHAR(255),
    `created_at` datetime,
    `updated_at` datetime
);