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
    "title" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "video_url" TEXT NOT NULL,
    "image_url" TEXT NOT NULL
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
    "title" TEXT NOT NULL,
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
    "title" TEXT NOT NULL,
    "description" TEXT NOT NULL,
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
    FOREIGN KEY("user_id") REFERENCES "users"("id") ON DELETE CASCADE;

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
    FOREIGN KEY("user_id") REFERENCES "users"("id") ON DELETE CASCADE;

ALTER TABLE "exercise_to_tag"
    ADD CONSTRAINT "exercise_to_tag_tag_id_foreign" 
    FOREIGN KEY("tag_id") REFERENCES "tag"("id") ON DELETE CASCADE,
    ADD CONSTRAINT "exercise_to_tag_exercise_id_foreign" 
    FOREIGN KEY("exercise_id") REFERENCES "exercise"("id") ON DELETE CASCADE;


-- 1. Сначала создаем ВСЕХ пользователей, включая тех, которые упоминаются в тренировках
INSERT INTO users (id, email, password_hash, is_blocked, last_login_at, created_at, updated_at) VALUES
-- Основные пользователи для тестирования
('11111111-1111-1111-1111-111111111111', 'alex.ivanov@gmail.com', crypt('password123', gen_salt('bf')), false, '2024-01-15 14:30:00+03', '2024-01-01 10:00:00+03', '2024-01-15 14:30:00+03'),
('22222222-2222-2222-2222-222222222222', 'maria.petrova@yandex.ru', crypt('qwerty123', gen_salt('bf')), false, '2024-01-14 09:15:00+03', '2024-01-05 12:00:00+03', '2024-01-14 09:15:00+03'),
('33333333-3333-3333-3333-333333333333', 'sidorov_s@mail.ru', crypt('sidorov123', gen_salt('bf')), true, '2024-01-10 16:45:00+03', '2023-12-20 08:00:00+03', '2024-01-12 11:20:00+03'),
('44444444-4444-4444-4444-444444444444', 'olga.fitness@proton.me', crypt('fitness2024', gen_salt('bf')), false, '2024-01-18 22:45:00+03', '2024-01-15 15:30:00+03', '2024-01-18 22:45:00+03'),
('55555555-5555-5555-5555-555555555555', 'bodybuilder_max@iron.com', crypt('maxpower', gen_salt('bf')), false, '2024-01-17 11:20:00+03', '2023-11-01 09:00:00+03', '2024-01-17 11:20:00+03'),
-- Дополнительные пользователи для тестирования разных сценариев
('66666666-6666-6666-6666-666666666666', 'test.today@example.com', crypt('today123', gen_salt('bf')), false, CURRENT_TIMESTAMP, CURRENT_DATE - 5, CURRENT_TIMESTAMP),
('77777777-7777-7777-7777-777777777777', 'test.planned@example.com', crypt('planned123', gen_salt('bf')), false, CURRENT_TIMESTAMP, CURRENT_DATE - 10, CURRENT_TIMESTAMP),
('88888888-8888-8888-8888-888888888888', 'test.started@example.com', crypt('started123', gen_salt('bf')), false, CURRENT_TIMESTAMP, CURRENT_DATE - 7, CURRENT_TIMESTAMP),
('99999999-9999-9999-9999-999999999999', 'test.done@example.com', crypt('done123', gen_salt('bf')), false, CURRENT_TIMESTAMP, CURRENT_DATE - 3, CURRENT_TIMESTAMP),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'test.multiple@example.com', crypt('multi123', gen_salt('bf')), false, CURRENT_TIMESTAMP, CURRENT_DATE - 15, CURRENT_TIMESTAMP);

-- 2. Создаем теги
INSERT INTO tag (type) VALUES
('Грудь'), ('Спина'), ('Ноги'), ('Плечи'), ('Бицепс'), ('Трицепс'),
('Кардио'), ('Силовая'), ('Выносливость'), ('Гибкость'), ('Базовая'),
('Изолированная'), ('С весом'), ('Без веса'), ('Домашняя'), ('Тренажерный зал'),
('Разминка'), ('Заминка'), ('Пресс'), ('Ягодицы'), ('Бедра'),
('Круговая'), ('Супперсет'), ('Дроп-сет'), ('Разминка суставов');

-- 3. Создаем упражнения (исправлено: добавляем все поля)
INSERT INTO exercise (title, description, video_url, image_url) VALUES
-- Базовые упражнения
('Жим лежа', 'Базовое упражнение для грудных мышц', 'https://www.youtube.com/watch?v=rT7DgCr-3pg', 'https://example.com/bench_press.jpg'),
('Приседания со штангой', 'Базовое упражнение для ног', 'https://www.youtube.com/watch?v=SW_C1A-rejs', 'https://example.com/squat.jpg'),
('Становая тяга', 'Базовое упражнение для спины', 'https://www.youtube.com/watch?v=1ZXobu7JvvE', 'https://example.com/deadlift.jpg'),
('Тяга штанги в наклоне', 'Упражнение для мышц спины', 'https://www.youtube.com/watch?v=G8l_8chR5BE', 'https://example.com/barbell_row.jpg'),
('Армейский жим', 'Упражнение для плеч', 'https://www.youtube.com/watch?v=2yjwXTZQDDI', 'https://example.com/overhead_press.jpg'),
-- Изолированные упражнения
('Подъем гантелей на бицепс', 'Изолированное упражнение для бицепса', 'https://www.youtube.com/watch?v=sAq_ocpRh_I', 'https://example.com/dumbbell_curl.jpg'),
('Французский жим', 'Упражнение для трицепса', 'https://www.youtube.com/watch?v=_gsUck-7M74', 'https://example.com/french_press.jpg'),
('Разведения гантелей в стороны', 'Упражнение для средних дельт', 'https://www.youtube.com/watch?v=3VcKaXpzqRo', 'https://example.com/lateral_raise.jpg'),
('Сгибания ног лежа', 'Упражнение для бицепса бедра', 'https://www.youtube.com/watch?v=1Tq3QdYUuHs', 'https://example.com/leg_curl.jpg'),
('Разгибания ног сидя', 'Упражнение для квадрицепса', 'https://www.youtube.com/watch?v=YyvSfVjQeL0', 'https://example.com/leg_extension.jpg'),
-- Кардио
('Бег на беговой дорожке', 'Кардио упражнение', 'https://www.youtube.com/watch?v=32CM2TQ6fes', 'https://example.com/treadmill.jpg'),
('Велотренажер', 'Кардио упражнение', 'https://www.youtube.com/watch?v=6N7dN6fUJmg', 'https://example.com/bike.jpg'),
('Скакалка', 'Кардио упражнение', 'https://www.youtube.com/watch?v=1BZM2Vre5oc', 'https://example.com/jump_rope.jpg'),
-- Домашние упражнения
('Отжимания', 'Базовое упражнение без оборудования', 'https://www.youtube.com/watch?v=IODxDxX7oi4', 'https://example.com/pushups.jpg'),
('Приседания без веса', 'Упражнение для ног без веса', 'https://www.youtube.com/watch?v=aclHkVaku9U', 'https://example.com/bodyweight_squat.jpg'),
('Планка', 'Упражнение на пресс и кор', 'https://www.youtube.com/watch?v=pSHjTRCQxIw', 'https://example.com/plank.jpg'),
('Выпады', 'Упражнение для ног', 'https://www.youtube.com/watch?v=QF0BQS8YQi8', 'https://example.com/lunges.jpg'),
('Подтягивания', 'Упражнение для спины', 'https://www.youtube.com/watch?v=eGo4IYlbE5g', 'https://example.com/pullups.jpg'),
('Берпи', 'Функциональное упражнение', 'https://www.youtube.com/watch?v=auBLPXO8Fww', 'https://example.com/burpee.jpg'),
-- Упражнения на гибкость
('Наклоны вперед', 'Упражнение на растяжку', 'https://www.youtube.com/watch?v=PhZPTI1wNNo', 'https://example.com/forward_bend.jpg'),
('Мостик', 'Упражнение на гибкость спины', 'https://www.youtube.com/watch?v=nziA8NCrDCI', 'https://example.com/bridge.jpg'),
-- Тренажеры
('Тяга верхнего блока', 'Упражнение на тренажере', 'https://www.youtube.com/watch?v=CAwf7n6Luuc', 'https://example.com/lat_pulldown.jpg'),
('Жим ногами', 'Упражнение на тренажере', 'https://www.youtube.com/watch?v=IZxyjW7MPJQ', 'https://example.com/leg_press.jpg'),
('Сведение рук в бабочке', 'Упражнение на грудные', 'https://www.youtube.com/watch?v=Z57CtFmRMxA', 'https://example.com/pec_deck.jpg'),
-- Дополнительные упражнения
('Тяга гантели в наклоне', 'Упражнение для спины', 'https://www.youtube.com/watch?v=roCP6wCXPqo', 'https://example.com/dumbbell_row.jpg'),
('Подъем на носки стоя', 'Упражнение для икр', 'https://www.youtube.com/watch?v=y-wV4Venusw', 'https://example.com/calf_raise.jpg'),
('Махи гантелями перед собой', 'Упражнение для передних дельт', 'https://www.youtube.com/watch?v=-t7fuZ0KhDA', 'https://example.com/front_raise.jpg'),
('Скручивания на пресс', 'Упражнение для пресса', 'https://www.youtube.com/watch?v=1we5jnyIGE4', 'https://example.com/crunches.jpg'),
('Боковая планка', 'Упражнение на боковые мышцы пресса', 'https://www.youtube.com/watch?v=k3j7F4m5rEo', 'https://example.com/side_plank.jpg'),
('Гиперэкстензия', 'Упражнение для поясницы', 'https://www.youtube.com/watch?v=PhZPTI1wNNo', 'https://example.com/hyperextension.jpg'),
-- Дополнительные для глобальных тренировок
('Подтягивания широким хватом', 'Упражнение для широкой спины', 'https://www.youtube.com/watch?v=eGo4IYlbE5g', 'https://example.com/wide_pullups.jpg'),
('Отжимания на брусьях', 'Упражнение для груди и трицепса', 'https://www.youtube.com/watch?v=2z8JmcrW-As', 'https://example.com/dips.jpg'),
('Мертвая тяга', 'Упражнение для бицепса бедра', 'https://www.youtube.com/watch?v=1zxR1Xhqo1Y', 'https://example.com/stiff_deadlift.jpg'),
('Жим гантелей лежа', 'Альтернатива жиму штанги', 'https://www.youtube.com/watch?v=VmB1G1K7v94', 'https://example.com/dumbbell_bench.jpg'),
('Приседания с гантелями', 'Приседания с дополнительным весом', 'https://www.youtube.com/watch?v=ca_PmUumI0E', 'https://example.com/goblet_squat.jpg'),
('Тяга Т-грифа', 'Упражнение для толщины спины', 'https://www.youtube.com/watch?v=j3Igk5nyZE4', 'https://example.com/tbar_row.jpg'),
('Подъем штанги на бицепс', 'Базовое упражнение для бицепса', 'https://www.youtube.com/watch?v=kwG2ipFRgfo', 'https://example.com/barbell_curl.jpg'),
('Пуловер', 'Упражнение для груди и спины', 'https://www.youtube.com/watch?v=6yMqYJp3E6Y', 'https://example.com/pullover.jpg'),
('Шраги', 'Упражнение для трапеций', 'https://www.youtube.com/watch?v=2J3-CrR50w8', 'https://example.com/shrugs.jpg'),
('Подъем ног в висе', 'Упражнение для нижнего пресса', 'https://www.youtube.com/watch?v=JB2oyawG9KI', 'https://example.com/hanging_leg_raise.jpg');

-- 4. Связываем упражнения с тегами
INSERT INTO exercise_to_tag (exercise_id, tag_id)
SELECT e.id, t.id
FROM exercise e
CROSS JOIN tag t
WHERE t.type IN ('Грудь', 'Спина', 'Ноги', 'Плечи', 'Бицепс', 'Трицепс', 'Базовая', 'Изолированная')
AND (
    (e.title LIKE '%жим%' AND t.type IN ('Грудь', 'Плечи', 'Базовая')) OR
    (e.title LIKE '%тяга%' AND t.type IN ('Спина', 'Базовая')) OR
    (e.title LIKE '%присед%' AND t.type IN ('Ноги', 'Базовая')) OR
    (e.title LIKE '%бицепс%' AND t.type = 'Бицепс') OR
    (e.title LIKE '%трицепс%' AND t.type = 'Трицепс')
)
ON CONFLICT DO NOTHING;

-- 5. Добавляем глобальные тренировки (исправлено: добавлены все поля)
INSERT INTO global_training (title, description, level) VALUES
('Начальный уровень', 'Тренировка для начинающих', 'beginner'),
('Средний уровень', 'Тренировка для продолжающих', 'intermediate'),
('Продвинутый уровень', 'Тренировка для опытных', 'advanced');

-- 6. Добавляем упражнения в глобальные тренировки
INSERT INTO global_training_exercise (global_training_id, exercise_id)
SELECT gt.id, e.id
FROM global_training gt
CROSS JOIN exercise e
WHERE (gt.level = 'beginner' AND e.id IN (1, 2, 3, 4, 5, 14, 15, 16, 17)) OR
      (gt.level = 'intermediate' AND e.id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 22, 23, 31, 32)) OR
      (gt.level = 'advanced' AND e.id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 24, 25, 26, 27, 28, 29, 30, 33, 34, 35, 36, 37, 38, 39, 40));

-- 7. Добавляем информацию о пользователях
INSERT INTO user_info (weight, height, date, age, user_id)
VALUES
-- Для пользователя test.today@example.com
(75.5, 180, CURRENT_DATE - 30, 25, '66666666-6666-6666-6666-666666666666'),
(76.0, 180, CURRENT_DATE - 15, 25, '66666666-6666-6666-6666-666666666666'),
(76.5, 180, CURRENT_DATE, 25, '66666666-6666-6666-6666-666666666666'),
-- Для пользователя test.multiple@example.com
(65.0, 170, CURRENT_DATE - 60, 22, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
(66.0, 170, CURRENT_DATE - 30, 22, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
(67.0, 170, CURRENT_DATE, 22, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
-- Для основного пользователя alex.ivanov@gmail.com
(80.0, 185, CURRENT_DATE - 90, 30, '11111111-1111-1111-1111-111111111111'),
(81.5, 185, CURRENT_DATE - 60, 30, '11111111-1111-1111-1111-111111111111'),
(82.0, 185, CURRENT_DATE - 30, 30, '11111111-1111-1111-1111-111111111111'),
-- Для основного пользователя maria.petrova@yandex.ru
(60.0, 168, CURRENT_DATE - 90, 28, '22222222-2222-2222-2222-222222222222'),
(61.0, 168, CURRENT_DATE - 60, 28, '22222222-2222-2222-2222-222222222222'),
(62.0, 168, CURRENT_DATE - 30, 28, '22222222-2222-2222-2222-222222222222');

-- Остальной код вставки тренировок и упражнений остается без изменений...

-- 8. Создаем тренировки для тестирования всех сценариев
-- (Остальная часть кода с 8.1 до конца остается без изменений, так как там нет проблем)

-- 8. Создаем тренировки для тестирования всех сценариев

-- 8.1 Тренировка на сегодня (не начата)
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES (
    'Тренировка на сегодня',  -- Добавлено название
    '66666666-6666-6666-6666-666666666666', false, CURRENT_DATE, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
)
RETURNING id AS training_today_id;

-- 8.2 Тренировка на сегодня (начата, но не завершена)
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES (
    'Утренняя тренировка',  -- Добавлено название
    '88888888-8888-8888-8888-888888888888', false, CURRENT_DATE, NULL, 
    CURRENT_TIMESTAMP - interval '30 minutes', NULL,
    NULL, NULL, NULL, NULL
)
RETURNING id AS training_started_id;

-- 8.3 Тренировка на сегодня (завершена)
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES (
    'Вечерняя силовая',  -- Добавлено название
    '99999999-9999-9999-9999-999999999999', true, CURRENT_DATE, CURRENT_DATE,
    CURRENT_TIMESTAMP - interval '2 hours', CURRENT_TIMESTAMP - interval '1 hour',
    interval '1 hour', interval '20 minutes', interval '40 minutes', 5
)
RETURNING id AS training_done_id;

-- 8.4 Несколько тренировок на сегодня для одного пользователя
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES 
(
    'Утренняя кардио',  -- Добавлено название
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', false, CURRENT_DATE, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
),
(
    'Силовая тренировка',  -- Добавлено название
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', true, CURRENT_DATE, CURRENT_DATE,
    CURRENT_TIMESTAMP - interval '3 hours', CURRENT_TIMESTAMP - interval '2 hours',
    interval '1 hour', interval '15 minutes', interval '45 minutes', 4
)
RETURNING id AS training_multiple_id;

-- 8.5 Запланированные тренировки на будущее
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES 
(
    'Завтрашняя тренировка',  -- Добавлено название
    '77777777-7777-7777-7777-777777777777', false, CURRENT_DATE + 1, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
),
(
    'Тренировка на среду',  -- Добавлено название
    '77777777-7777-7777-7777-777777777777', false, CURRENT_DATE + 2, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
),
(
    'Тренировка на четверг',  -- Добавлено название
    '77777777-7777-7777-7777-777777777777', false, CURRENT_DATE + 3, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
);

-- 8.6 Прошлые тренировки с разными статусами для существующих пользователей
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
) VALUES 
(
    'Тренировка 7 дней назад',  -- Добавлено название
    '11111111-1111-1111-1111-111111111111', true, CURRENT_DATE - 7, CURRENT_DATE - 7,
    CURRENT_TIMESTAMP - interval '7 days 2 hours', CURRENT_TIMESTAMP - interval '7 days 1 hour',
    interval '1 hour', interval '18 minutes', interval '42 minutes', 5
),
(
    'Тренировка 5 дней назад',  -- Добавлено название
    '11111111-1111-1111-1111-111111111111', true, CURRENT_DATE - 5, CURRENT_DATE - 5,
    CURRENT_TIMESTAMP - interval '5 days 3 hours', CURRENT_TIMESTAMP - interval '5 days 2 hours',
    interval '1 hour 15 minutes', interval '25 minutes', interval '50 minutes', 4
),
(
    'Тренировка 3 дня назад',  -- Добавлено название
    '22222222-2222-2222-2222-222222222222', true, CURRENT_DATE - 3, CURRENT_DATE - 3,
    CURRENT_TIMESTAMP - interval '3 days 4 hours', CURRENT_TIMESTAMP - interval '3 days 3 hours',
    interval '45 minutes', interval '10 minutes', interval '35 minutes', 3
),
(
    'Вчерашняя тренировка',  -- Добавлено название
    '22222222-2222-2222-2222-222222222222', false, CURRENT_DATE - 1, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL
),
(
    'Тренировка 10 дней назад',  -- Добавлено название
    '33333333-3333-3333-3333-333333333333', true, CURRENT_DATE - 10, CURRENT_DATE - 10,
    CURRENT_TIMESTAMP - interval '10 days 3 hours', CURRENT_TIMESTAMP - interval '10 days 2 hours',
    interval '50 minutes', interval '15 minutes', interval '35 minutes', 2
),
(
    'Тренировка 4 дня назад',  -- Добавлено название
    '44444444-4444-4444-4444-444444444444', true, CURRENT_DATE - 4, CURRENT_DATE - 4,
    CURRENT_TIMESTAMP - interval '4 days 5 hours', CURRENT_TIMESTAMP - interval '4 days 4 hours',
    interval '1 hour 10 minutes', interval '20 minutes', interval '50 minutes', 4
),
(
    'Тренировка 2 дня назад',  -- Добавлено название
    '55555555-5555-5555-5555-555555555555', true, CURRENT_DATE - 2, CURRENT_DATE - 2,
    CURRENT_TIMESTAMP - interval '2 days 6 hours', CURRENT_TIMESTAMP - interval '2 days 5 hours',
    interval '1 hour 30 minutes', interval '30 minutes', interval '1 hour', 5
);

-- 9. Создаем переменные для ID тренировок (используем временные таблицы вместо \gset)
DO $$
DECLARE
    training_today_id BIGINT;
    training_started_id BIGINT;
    training_done_id BIGINT;
    training_multiple_id BIGINT;
    training_stats_id BIGINT;
    training_update_id BIGINT;
    training_delete_id BIGINT;
    exercise_update_id BIGINT;
BEGIN
    -- Получаем ID тренировок
    SELECT id INTO training_today_id FROM training WHERE user_id = '66666666-6666-6666-6666-666666666666' AND planned_date = CURRENT_DATE AND is_done = false ORDER BY id DESC LIMIT 1;
    SELECT id INTO training_started_id FROM training WHERE user_id = '88888888-8888-8888-8888-888888888888' AND planned_date = CURRENT_DATE AND started_at IS NOT NULL ORDER BY id DESC LIMIT 1;
    SELECT id INTO training_done_id FROM training WHERE user_id = '99999999-9999-9999-9999-999999999999' AND planned_date = CURRENT_DATE AND is_done = true ORDER BY id DESC LIMIT 1;
    SELECT id INTO training_multiple_id FROM training WHERE user_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AND planned_date = CURRENT_DATE AND is_done = true ORDER BY id DESC LIMIT 1;

    -- 10. Добавляем упражнения в тренировки для тестирования

    -- 10.1 В тренировку на сегодня (не начатую)
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    SELECT 
        training_today_id,
        e.id,
        40.0 + floor(random() * 40)::numeric,
        3 + floor(random() * 2)::int,
        8 + floor(random() * 7)::int,
        NULL,
        interval '30 seconds',
        interval '90 seconds',
        'Планирую сделать'
    FROM exercise e
    WHERE e.title IN ('Жим лежа', 'Приседания со штангой', 'Тяга штанги в наклоне', 'Подъем гантелей на бицепс')
    LIMIT 4;

        -- 10.2 В тренировку на сегодня (начатую)
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    VALUES 
    (
        training_started_id,
        (SELECT id FROM exercise WHERE title = 'Жим лежа'),
        80.0, 4, 10, interval '4 minutes', interval '2 minutes', interval '2 minutes', 'Сделал 4 подхода'
    ),
    (
        training_started_id,
        (SELECT id FROM exercise WHERE title = 'Тяга штанги в наклоне'),
        60.0, 3, 12, interval '3 minutes 30 seconds', interval '1 minute 45 seconds', interval '1 minute 45 seconds', 'Хорошая техника'
    );

    -- 10.3 В тренировку на сегодня (завершенную)
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    VALUES 
    (
        training_done_id,
        (SELECT id FROM exercise WHERE title = 'Жим лежа'),
        85.0, 4, 8, interval '4 minutes 30 seconds', interval '2 minutes 15 seconds', interval '2 minutes 15 seconds', 'Тяжело, но сделал'
    ),
    (
        training_done_id,
        (SELECT id FROM exercise WHERE title = 'Приседания со штангой'),
        100.0, 3, 10, interval '3 minutes 45 seconds', interval '1 minute 50 seconds', interval '1 minute 55 seconds', 'Отличная форма'
    ),
    (
        training_done_id,
        (SELECT id FROM exercise WHERE title = 'Подъем гантелей на бицепс'),
        18.0, 3, 12, interval '3 minutes', interval '1 minute 30 seconds', interval '1 minute 30 seconds', 'Легко'
    );

    -- 10.4 Добавляем упражнения с разными типами времени для тестирования
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    SELECT 
        t.id,
        e.id,
        CASE WHEN e.description LIKE '%гантел%' THEN 15.0 ELSE 50.0 END,
        3,
        10,
        interval '3 minutes',
        interval '1 minute 30 seconds',
        interval '1 minute 30 seconds',
        'Тестовое упражнение'
    FROM training t
    CROSS JOIN exercise e
    WHERE t.user_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
    AND t.planned_date = CURRENT_DATE
    AND t.is_done = true
    AND e.title IN ('Бег на беговой дорожке', 'Планка', 'Отжимания')
    LIMIT 3;

    -- 11. Обновляем время тренировок на основе добавленных упражнений
    WITH time_calc AS (
        SELECT 
            te.training_id,
            SUM(EXTRACT(EPOCH FROM te.doing)) as total_exercise_sec,
            SUM(EXTRACT(EPOCH FROM te.rest)) as total_rest_sec
        FROM trained_exercise te
        WHERE te.training_id IN (training_started_id, training_done_id)
        GROUP BY te.training_id
    )
    UPDATE training t
    SET 
        total_duration = make_interval(secs => tc.total_exercise_sec + tc.total_rest_sec),
        total_rest_time = make_interval(secs => tc.total_rest_sec),
        total_exercise_time = make_interval(secs => tc.total_exercise_sec)
    FROM time_calc tc
    WHERE t.id = tc.training_id;

    -- 12. Добавляем рекомендации для тренировок
    INSERT INTO recommendation (training_id, approach, weight, time, reason)
    VALUES 
    (
        training_done_id,
        1,
        5.0,
        interval '30 seconds',
        'Увеличить время отдыха между подходами'
    ),
    (
        training_done_id,
        NULL,
        NULL,
        interval '10 seconds',
        'Улучшить технику выполнения'
    ),
    (
        training_started_id,
        2,
        2.5,
        interval '45 seconds',
        'Добавить еще один подход'
    );

    -- 13. Создаем тренировку с большим количеством упражнений для тестирования статистики
    INSERT INTO training (
        title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
        total_duration, total_rest_time, total_exercise_time, rating
    ) VALUES (
        'Тренировка для статистики',
        '11111111-1111-1111-1111-111111111111', true, CURRENT_DATE - 2, CURRENT_DATE - 2,
        CURRENT_TIMESTAMP - interval '2 days 5 hours', CURRENT_TIMESTAMP - interval '2 days 4 hours',
        interval '1 hour 30 minutes', interval '30 minutes', interval '1 hour', 5
    )
    RETURNING id INTO training_stats_id;

    -- 14. Добавляем много упражнений в эту тренировку
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    SELECT 
        training_stats_id,
        e.id,
        CASE 
            WHEN e.title LIKE '%штан%' THEN 70.0 + (row_number() over()) * 5
            WHEN e.title LIKE '%гантел%' THEN 15.0 + (row_number() over()) * 2
            ELSE NULL
        END,
        4,
        10 + (row_number() over()),
        interval '4 minutes',
        interval '2 minutes',
        interval '2 minutes',
        'Упражнение ' || row_number() over()
    FROM exercise e
    WHERE e.title IN ('Жим лежа', 'Тяга штанги в наклоне', 'Подъем гантелей на бицепс', 'Французский жим', 'Разведения гантелей в стороны', 'Сгибания ног лежа', 'Разгибания ног сидя')
    LIMIT 7;

    -- 15. Создаем тренировку для тестирования операций обновления
    INSERT INTO training (
        title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
        total_duration, total_rest_time, total_exercise_time, rating
    ) VALUES (
        'Тренировка для обновления',
        '44444444-4444-4444-4444-444444444444', false, CURRENT_DATE, NULL, NULL, NULL,
        NULL, NULL, NULL, NULL
    )
    RETURNING id INTO training_update_id;

    -- 16. Добавляем упражнение для тестирования обновления
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    VALUES (
        training_update_id,
        (SELECT id FROM exercise WHERE title = 'Жим лежа'),
        70.0, 3, 10, interval '3 minutes', interval '1 minute 30 seconds', interval '1 minute 30 seconds', 'Начальный вариант'
    )
    RETURNING id INTO exercise_update_id;

    -- 17. Создаем тренировку для тестирования удаления
    INSERT INTO training (
        title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
        total_duration, total_rest_time, total_exercise_time, rating
    ) VALUES (
        'Тренировка для удаления',
        '55555555-5555-5555-5555-555555555555', false, CURRENT_DATE, NULL, NULL, NULL,
        NULL, NULL, NULL, NULL
    )
    RETURNING id INTO training_delete_id;

    -- 18. Добавляем несколько упражнений для тестирования удаления
    INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
    SELECT 
        training_delete_id,
        e.id,
        50.0, 3, 10, interval '3 minutes', interval '1 minute 30 seconds', interval '1 minute 30 seconds', 'Для удаления'
    FROM exercise e
    WHERE e.title IN ('Жим лежа', 'Приседания со штангой', 'Тяга штанги в наклоне')
    LIMIT 3;

    -- 19. Обновляем статистику для всех тренировок с упражнениями
    WITH stats AS (
        SELECT 
            te.training_id,
            COUNT(te.id) as exercise_count,
            COALESCE(SUM(EXTRACT(EPOCH FROM te.doing)), 0) as total_exercise_sec,
            COALESCE(SUM(EXTRACT(EPOCH FROM te.rest)), 0) as total_rest_sec,
            COALESCE(SUM(te.approaches), 0) as total_approaches,
            COALESCE(SUM(te.reps), 0) as total_reps
        FROM trained_exercise te
        GROUP BY te.training_id
    )
    UPDATE training t
    SET 
        total_duration = make_interval(secs => s.total_exercise_sec + s.total_rest_sec),
        total_rest_time = make_interval(secs => s.total_rest_sec),
        total_exercise_time = make_interval(secs => s.total_exercise_sec)
    FROM stats s
    WHERE t.id = s.training_id
    AND (t.total_duration IS NULL OR t.total_rest_time IS NULL OR t.total_exercise_time IS NULL);

END $$;

-- 20. Добавляем дополнительные связи упражнений с тегами для полноты
INSERT INTO exercise_to_tag (exercise_id, tag_id)
SELECT DISTINCT e.id, t.id
FROM exercise e
CROSS JOIN tag t
WHERE t.type IN ('Разминка', 'Заминка', 'Пресс', 'Кардио', 'Выносливость', 'Гибкость')
AND (
    (e.title LIKE '%бег%' AND t.type IN ('Кардио', 'Выносливость')) OR
    (e.title LIKE '%планк%' AND t.type IN ('Пресс', 'Без веса', 'Домашняя')) OR
    (e.title LIKE '%наклоны%' AND t.type IN ('Гибкость', 'Заминка')) OR
    (e.title LIKE '%мостик%' AND t.type IN ('Гибкость', 'Заминка'))
)
ON CONFLICT DO NOTHING;

-- 21. Создаем дополнительные тренировки для полноты тестовых данных
-- 21. Создаем дополнительные тренировки для полноты тестовых данных
INSERT INTO training (
    title, user_id, is_done, planned_date, actual_date, started_at, finished_at,
    total_duration, total_rest_time, total_exercise_time, rating
)
SELECT 
    'Дополнительная тренировка ' || row_number() OVER (),  -- Добавлено название
    u.id,
    CASE WHEN random() > 0.3 THEN true ELSE false END,
    date_trunc('day', CURRENT_DATE - (n || ' days')::interval)::date,
    CASE WHEN random() > 0.3 THEN date_trunc('day', CURRENT_DATE - (n || ' days')::interval)::date ELSE NULL END,
    CASE WHEN random() > 0.3 THEN CURRENT_TIMESTAMP - (n || ' days 2 hours')::interval ELSE NULL END,
    CASE WHEN random() > 0.3 THEN CURRENT_TIMESTAMP - (n || ' days 1 hour')::interval ELSE NULL END,
    CASE WHEN random() > 0.3 THEN interval '1 hour' ELSE NULL END,
    CASE WHEN random() > 0.3 THEN interval '20 minutes' ELSE NULL END,
    CASE WHEN random() > 0.3 THEN interval '40 minutes' ELSE NULL END,
    CASE WHEN random() > 0.7 THEN 3 + floor(random() * 3)::int ELSE NULL END
FROM users u
CROSS JOIN generate_series(1, 10) n
WHERE u.email NOT LIKE '%@test.com'
AND u.id IN ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222')
ORDER BY random()
LIMIT 10;

-- 22. Добавляем упражнения в эти тренировки
INSERT INTO trained_exercise (training_id, exercise_id, weight, approaches, reps, time, doing, rest, notes)
SELECT 
    t.id,
    e.id,
    CASE WHEN random() > 0.5 THEN 20.0 + random() * 50 ELSE NULL END,
    3 + floor(random() * 3)::int,
    8 + floor(random() * 8)::int,
    CASE WHEN random() > 0.5 THEN interval '3 minutes' ELSE NULL END,
    CASE WHEN random() > 0.5 THEN interval '1 minute 30 seconds' ELSE NULL END,
    CASE WHEN random() > 0.5 THEN interval '1 minute 30 seconds' ELSE NULL END,
    CASE WHEN random() > 0.7 THEN 'Тестовые заметки' ELSE NULL END
FROM training t
CROSS JOIN exercise e
WHERE t.user_id IN ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222')
AND t.id NOT IN (SELECT training_id FROM trained_exercise)
AND e.id IN (1, 2, 3, 4, 5, 6, 7)
ORDER BY random()
LIMIT 20;

-- Выводим информацию о созданных данных
SELECT 'Тестовые данные успешно созданы!' as message;
SELECT COUNT(*) as user_count FROM users;
SELECT COUNT(*) as exercise_count FROM exercise;
SELECT COUNT(*) as tag_count FROM tag;
SELECT COUNT(*) as training_count FROM training;
SELECT COUNT(*) as trained_exercise_count FROM trained_exercise;
SELECT COUNT(*) as recommendation_count FROM recommendation;
SELECT COUNT(*) as user_info_count FROM user_info;