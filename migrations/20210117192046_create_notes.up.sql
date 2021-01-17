CREATE TABLE notes (
    id bigserial not null primary key,
    author_id bigint REFERENCES users (id) ON DELETE CASCADE,
    header varchar not null,
    body text not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);