CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGSERIAL NOT NULL UNIQUE,
    user_id BIGSERIAL NOT NULL UNIQUE,
    content TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);