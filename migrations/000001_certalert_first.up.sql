-- users table in order to store user data
CREATE TABLE IF NOT EXISTS "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "first_name" VARCHAR NOT NULL,
    "last_name" VARCHAR NOT NULL,
    "email" VARCHAR NOT NULL,
    "password" VARCHAR NOT NULL,
    "created_at" TIMESTAMP DEFAULT timezone('Asia/Tashkent', CURRENT_TIMESTAMP) NOT NULL
);

CREATE TABLE IF NOT EXISTS "sessions" (
    "id" UUID PRIMARY KEY,
    "user_id" BIGINT REFERENCES users(id) ON DELETE CASCADE,
    "access_token" TEXT NOT NULL,
    "expires_at" TIMESTAMP NOT NULL,
    "ip_address" VARCHAR NOT NULL,
    "user_agent" VARCHAR NOT NULL,
    "city" VARCHAR NOT NULL,
    "region" VARCHAR NOT NULL,
    "country" VARCHAR NOT NULL,
    "timezone" VARCHAR NOT NULL,
    "is_blocked" BOOLEAN NOT NULL DEFAULT false,
    "last_login" TIMESTAMP NOT NULL,
    "created_at"  TIMESTAMP DEFAULT timezone('Asia/Tashkent', CURRENT_TIMESTAMP) NOT NULL
);