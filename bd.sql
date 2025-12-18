-- Включаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email          CITEXT      UNIQUE NOT NULL,
  password_hash  TEXT        NOT NULL,
  is_blocked     BOOLEAN     NOT NULL DEFAULT false,
  last_login_at  TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS refresh_sessions (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  BYTEA       NOT NULL,
  user_agent  TEXT,
  ip          INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_token_hash ON refresh_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_active ON refresh_sessions(user_id)
  WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS password_resets (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  otp_hash    TEXT        NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pwreset_user ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_pwreset_active ON password_resets(user_id)
  WHERE used_at IS NULL;

-- Таблица тегов упражнений
CREATE TABLE "tag"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "type" VARCHAR(255) NOT NULL
);

-- Таблица упражнений
CREATE TABLE "exercise"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "description" TEXT NOT NULL,
    "href" TEXT NOT NULL
);

-- Связующая таблица упражнений и тегов
CREATE TABLE "exercise_to_tag"(
    "exercise_id" BIGINT NOT NULL,
    "tag_id" BIGINT NOT NULL,
    PRIMARY KEY ("exercise_id", "tag_id")
);

-- Таблица тренировок
CREATE TABLE "training"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "is_done" BOOLEAN NOT NULL DEFAULT FALSE,
    "planned_date" DATE NOT NULL,
    "actual_date" DATE NULL,
    "started_at" TIMESTAMP NULL,
    "finished_at" TIMESTAMP NULL,
    "total_duration" INTERVAL NULL,
    "total_rest_time" INTERVAL NULL,
    "total_exercise_time" INTERVAL NULL,
    "rating" INTEGER CHECK(rating >= 1 AND rating <= 5) NULL
);

-- Таблица выполненных упражнений в тренировке
CREATE TABLE "trained_exercise"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "training_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "weight" DECIMAL(5,2) NULL,
    "approaches" INTEGER NULL,
    "reps" INTEGER NULL,
    "time" INTERVAL NULL,
    "doing" INTERVAL NULL,
    "rest" INTERVAL NULL,
    "notes" TEXT NULL
);

-- Таблица глобальных тренировок
CREATE TABLE "global_training"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "level" VARCHAR(50) NOT NULL CHECK(level IN('beginner', 'intermediate', 'advanced'))
);

-- Связующая таблица глобальных тренировок и упражнений
CREATE TABLE "global_training_exercise"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "global_training_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL
);

-- Таблица рекомендаций
CREATE TABLE "recommendation"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "training_id" BIGINT NOT NULL,
    "approach" INTEGER NULL,
    "weight" DECIMAL(5,2) NULL,
    "time" INTERVAL NOT NULL,
    "reason" TEXT NOT NULL
);

-- Таблица информации о пользователе
CREATE TABLE "user_info"(
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "weight" DECIMAL(5,2) NOT NULL,
    "height" INTEGER NOT NULL,
    "date" DATE NOT NULL,
    "age" INTEGER NOT NULL,
    "user_id" UUID NOT NULL
);

-- Индексы для производительности
CREATE INDEX idx_training_user_id ON "training"(user_id);
CREATE INDEX idx_training_planned_date ON "training"(planned_date);
CREATE INDEX idx_training_is_done ON "training"(is_done);
CREATE INDEX idx_trained_exercise_training_id ON "trained_exercise"(training_id);
CREATE INDEX idx_trained_exercise_exercise_id ON "trained_exercise"(exercise_id);
CREATE INDEX idx_exercise_to_tag_exercise_id ON "exercise_to_tag"(exercise_id);
CREATE INDEX idx_exercise_to_tag_tag_id ON "exercise_to_tag"(tag_id);
CREATE INDEX idx_user_info_user_id ON "user_info"(user_id);
CREATE INDEX idx_user_info_user_id_date ON "user_info"(user_id, date DESC);
CREATE INDEX idx_global_training_exercise_training_id ON "global_training_exercise"(global_training_id);
CREATE INDEX idx_global_training_exercise_exercise_id ON "global_training_exercise"(exercise_id);

-- Внешние ключи
ALTER TABLE "training"
    ADD CONSTRAINT "training_user_id_foreign" 
    FOREIGN KEY("user_id") REFERENCES "user"("id") ON DELETE CASCADE;

ALTER TABLE "trained_exercise"
    ADD CONSTRAINT "trained_exercise_training_id_foreign" 
    FOREIGN KEY("training_id") REFERENCES "training"("id") ON DELETE CASCADE,
    ADD CONSTRAINT "trained_exercise_exercise_id_foreign" 
    FOREIGN KEY("exercise_id") REFERENCES "exercise"("id") ON DELETE CASCADE;

ALTER TABLE "global_training_exercise"
    ADD CONSTRAINT "global_training_exercise_training_id_foreign" 
    FOREIGN KEY("global_training_id") REFERENCES "global_training"("id") ON DELETE CASCADE,
    ADD CONSTRAINT "global_training_exercise_exercise_id_foreign" 
    FOREIGN KEY("exercise_id") REFERENCES "exercise"("id") ON DELETE CASCADE;

ALTER TABLE "recommendation"
    ADD CONSTRAINT "recommendation_training_id_foreign" 
    FOREIGN KEY("training_id") REFERENCES "training"("id") ON DELETE CASCADE;

ALTER TABLE "user_info"
    ADD CONSTRAINT "user_info_user_id_foreign" 
    FOREIGN KEY("user_id") REFERENCES "user"("id") ON DELETE CASCADE;

ALTER TABLE "exercise_to_tag"
    ADD CONSTRAINT "exercise_to_tag_tag_id_foreign" 
    FOREIGN KEY("tag_id") REFERENCES "tag"("id") ON DELETE CASCADE,
    ADD CONSTRAINT "exercise_to_tag_exercise_id_foreign" 
    FOREIGN KEY("exercise_id") REFERENCES "exercise"("id") ON DELETE CASCADE;