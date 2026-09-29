CREATE TABLE IF NOT EXISTS invitations (
    token bytea PRIMARY KEY,
    user_id bigint NOT NULL,
);