-- Включаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица тегов упражнений
CREATE TABLE tag (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    type VARCHAR(255) NOT NULL
);

-- Таблица упражнений
CREATE TABLE exercise (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    description TEXT NOT NULL,
    href TEXT NOT NULL
);

-- Связующая таблица упражнений и тегов
CREATE TABLE exercise_to_tag (
    exercise_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (exercise_id, tag_id)
);

-- Таблица тренировок (обновленная с таймером)
CREATE TABLE training (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    is_done BOOLEAN NOT NULL DEFAULT FALSE,
    planned_date DATE NOT NULL,
    actual_date DATE NULL,
    started_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL,
    total_duration INTERVAL NULL,
    total_rest_time INTERVAL NULL,
    total_exercise_time INTERVAL NULL,
    rating INTEGER CHECK(rating >= 1 AND rating <= 5) NULL
);

-- Таблица выполненных упражнений в тренировке (обновленная с таймером)
CREATE TABLE trained_exercise (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    training_id BIGINT NOT NULL,
    exercise_id BIGINT NOT NULL,
    weight DECIMAL(5,2) NULL,
    approaches INTEGER NULL,
    reps INTEGER NULL,
    time INTERVAL NULL,
    doing INTERVAL NULL,
    rest INTERVAL NULL,
    notes TEXT NULL
);

-- Таблица глобальных тренировок
CREATE TABLE global_training (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    level VARCHAR(50) NOT NULL CHECK(level IN('beginner', 'intermediate', 'advanced'))
);

-- Связующая таблица глобальных тренировок и упражнений
CREATE TABLE global_training_exercise (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    global_training_id BIGINT NOT NULL,
    exercise_id BIGINT NOT NULL
);

-- Индексы для производительности
CREATE INDEX idx_training_user_id ON training(user_id);
CREATE INDEX idx_training_planned_date ON training(planned_date);
CREATE INDEX idx_training_is_done ON training(is_done);
CREATE INDEX idx_trained_exercise_training_id ON trained_exercise(training_id);
CREATE INDEX idx_trained_exercise_exercise_id ON trained_exercise(exercise_id);
CREATE INDEX idx_exercise_to_tag_exercise_id ON exercise_to_tag(exercise_id);
CREATE INDEX idx_exercise_to_tag_tag_id ON exercise_to_tag(tag_id);
CREATE INDEX idx_global_training_exercise_training_id ON global_training_exercise(global_training_id);
CREATE INDEX idx_global_training_exercise_exercise_id ON global_training_exercise(exercise_id);

-- Внешние ключи
ALTER TABLE trained_exercise
    ADD CONSTRAINT trained_exercise_training_id_foreign 
    FOREIGN KEY (training_id) REFERENCES training(id) ON DELETE CASCADE,
    ADD CONSTRAINT trained_exercise_exercise_id_foreign 
    FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE;

ALTER TABLE global_training_exercise
    ADD CONSTRAINT global_training_exercise_training_id_foreign 
    FOREIGN KEY (global_training_id) REFERENCES global_training(id) ON DELETE CASCADE,
    ADD CONSTRAINT global_training_exercise_exercise_id_foreign 
    FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE;

ALTER TABLE exercise_to_tag
    ADD CONSTRAINT exercise_to_tag_tag_id_foreign 
    FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE,
    ADD CONSTRAINT exercise_to_tag_exercise_id_foreign 
    FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE;