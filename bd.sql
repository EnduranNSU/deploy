-- 1. Таблицы пользователей и профиля
CREATE TABLE "user"(
    "id" UUID NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "tg_id" UUID NOT NULL
);
ALTER TABLE "user" ADD PRIMARY KEY("id");

CREATE TABLE "user_info"(
    "id" BIGSERIAL NOT NULL,
    "weight" FLOAT(53) NOT NULL,
    "height" BIGINT NOT NULL,
    "date" DATE NOT NULL,
    "age" BIGINT NOT NULL,
    "user_id" UUID NOT NULL
);
ALTER TABLE "user_info" ADD PRIMARY KEY("id");

-- 2. Таблицы упражнений и тегов
CREATE TABLE "tag"(
    "id" BIGSERIAL NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "category" VARCHAR(50) CHECK (category IN ('MUSCLE_GROUP', 'EQUIPMENT', 'TYPE', 'DIFFICULTY'))
);
ALTER TABLE "tag" ADD PRIMARY KEY("id");

CREATE TABLE "exercise"(
    "id" BIGSERIAL NOT NULL,
    "description" TEXT NOT NULL,
    "href" TEXT NOT NULL,
    "exercise_type" VARCHAR(20) CHECK (exercise_type IN ('STRENGTH', 'CARDIO', 'FLEXIBILITY', 'BALANCE')) DEFAULT 'STRENGTH',
    "media_url" TEXT,
    "technique_description" TEXT
);
ALTER TABLE "exercise" ADD PRIMARY KEY("id");

CREATE TABLE "exercise_to_tag"(
    "execrise_id" BIGINT NOT NULL,
    "tag_id" BIGINT NOT NULL
);

-- 3. Таблицы тренировок
CREATE TABLE "training"(
    "id" BIGSERIAL NOT NULL,
    "user_id" UUID NOT NULL,
    "isDone" BOOLEAN NOT NULL,
    "planned" DATE NOT NULL,
    "done" DATE NULL,
    "total_time" INTERVAL NULL,
    "rating" INTEGER NULL
);
ALTER TABLE "training" ADD PRIMARY KEY("id");

CREATE TABLE "trained_exercise"(
    "id" BIGSERIAL NOT NULL,
    "training_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "weight" FLOAT(53) NULL,
    "approaches" BIGINT NULL,
    "reps" BIGINT NULL,
    "time" TIME(0) WITHOUT TIME ZONE NULL,
    "notes" TEXT NULL,
    "exercise_duration" INTERVAL,
    "rest_duration" INTERVAL
);
ALTER TABLE "trained_exercise" ADD PRIMARY KEY("id");

-- 4. Таблицы рекомендаций
CREATE TABLE "recommendation"(
    "id" BIGSERIAL NOT NULL,
    "training_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "approach" BIGINT NULL,
    "weight" FLOAT(53) NULL,
    "time" TIME(0) WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE "recommendation" ADD PRIMARY KEY("id");

-- 5. Таблицы расписания тренировок
CREATE TABLE "training_schedule"(
    "id" BIGSERIAL NOT NULL,
    "user_id" UUID NOT NULL,
    "day_of_week" INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE "training_schedule" ADD PRIMARY KEY("id");

CREATE TABLE "schedule_exercise"(
    "id" BIGSERIAL NOT NULL,
    "schedule_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "order_index" INTEGER NOT NULL,
    "sets" INTEGER NOT NULL,
    "reps_min" INTEGER,
    "reps_max" INTEGER,
    "rest_time" INTERVAL,
    "duration" INTERVAL
);
ALTER TABLE "schedule_exercise" ADD PRIMARY KEY("id");

-- 6. Таблицы таймеров упражнений
CREATE TABLE "exercise_timer"(
    "id" BIGSERIAL NOT NULL,
    "trained_exercise_id" BIGINT NOT NULL,
    "start_time" TIMESTAMP NOT NULL,
    "end_time" TIMESTAMP,
    "type" VARCHAR(20) CHECK (type IN ('EXERCISE', 'REST')) NOT NULL
);
ALTER TABLE "exercise_timer" ADD PRIMARY KEY("id");

-- 7. Таблицы шаблонов тренировок
CREATE TABLE "workout_template"(
    "id" BIGSERIAL NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "difficulty" VARCHAR(20) CHECK (difficulty IN ('BEGINNER', 'INTERMEDIATE', 'ADVANCED')),
    "estimated_duration" INTERVAL,
    "is_public" BOOLEAN NOT NULL DEFAULT true,
    "created_by" UUID,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE "workout_template" ADD PRIMARY KEY("id");

CREATE TABLE "template_exercise"(
    "id" BIGSERIAL NOT NULL,
    "template_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "order_index" INTEGER NOT NULL,
    "sets" INTEGER NOT NULL,
    "reps_min" INTEGER,
    "reps_max" INTEGER,
    "rest_time" INTERVAL
);
ALTER TABLE "template_exercise" ADD PRIMARY KEY("id");

-- 8. Таблица избранных тренировок
CREATE TABLE "user_favorite_workouts"(
    "user_id" UUID NOT NULL,
    "template_id" BIGINT NOT NULL,
    "added_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, template_id)
);

-- Внешние ключи
ALTER TABLE "user_info" ADD CONSTRAINT "user_info_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");

ALTER TABLE "exercise_to_tag" ADD CONSTRAINT "exercise_to_tag_tag_id_foreign" FOREIGN KEY("tag_id") REFERENCES "tag"("id");
ALTER TABLE "exercise_to_tag" ADD CONSTRAINT "exercise_to_tag_execrise_id_foreign" FOREIGN KEY("execrise_id") REFERENCES "exercise"("id");

ALTER TABLE "training" ADD CONSTRAINT "training_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");

ALTER TABLE "trained_exercise" ADD CONSTRAINT "trained_exercise_training_id_foreign" FOREIGN KEY("training_id") REFERENCES "training"("id");
ALTER TABLE "trained_exercise" ADD CONSTRAINT "trained_exercise_exercise_id_foreign" FOREIGN KEY("exercise_id") REFERENCES "exercise"("id");

ALTER TABLE "recommendation" ADD CONSTRAINT "recommendation_training_id_foreign" FOREIGN KEY("training_id") REFERENCES "training"("id");
ALTER TABLE "recommendation" ADD CONSTRAINT "recommendation_exercise_id_foreign" FOREIGN KEY("exercise_id") REFERENCES "exercise"("id");

ALTER TABLE "training_schedule" ADD CONSTRAINT "training_schedule_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");

ALTER TABLE "schedule_exercise" ADD CONSTRAINT "schedule_exercise_schedule_id_foreign" FOREIGN KEY("schedule_id") REFERENCES "training_schedule"("id");
ALTER TABLE "schedule_exercise" ADD CONSTRAINT "schedule_exercise_exercise_id_foreign" FOREIGN KEY("exercise_id") REFERENCES "exercise"("id");

ALTER TABLE "exercise_timer" ADD CONSTRAINT "exercise_timer_trained_exercise_id_foreign" FOREIGN KEY("trained_exercise_id") REFERENCES "trained_exercise"("id");

ALTER TABLE "workout_template" ADD CONSTRAINT "workout_template_created_by_foreign" FOREIGN KEY("created_by") REFERENCES "user"("id");

ALTER TABLE "template_exercise" ADD CONSTRAINT "template_exercise_template_id_foreign" FOREIGN KEY("template_id") REFERENCES "workout_template"("id");
ALTER TABLE "template_exercise" ADD CONSTRAINT "template_exercise_exercise_id_foreign" FOREIGN KEY("exercise_id") REFERENCES "exercise"("id");

ALTER TABLE "user_favorite_workouts" ADD CONSTRAINT "user_favorite_workouts_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");
ALTER TABLE "user_favorite_workouts" ADD CONSTRAINT "user_favorite_workouts_template_id_foreign" FOREIGN KEY("template_id") REFERENCES "workout_template"("id");

-- Индексы
CREATE INDEX idx_training_user_id ON training(user_id);
CREATE INDEX idx_training_planned ON training(planned);
CREATE INDEX idx_trained_exercise_training_id ON trained_exercise(training_id);
CREATE INDEX idx_schedule_exercise_schedule_id ON schedule_exercise(schedule_id);
CREATE INDEX idx_template_exercise_template_id ON template_exercise(template_id);
CREATE INDEX idx_workout_template_is_public ON workout_template(is_public);
CREATE INDEX idx_exercise_timer_trained_exercise_id ON exercise_timer(trained_exercise_id);
CREATE INDEX idx_exercise_to_tag_exercise_id ON exercise_to_tag(execrise_id);
CREATE INDEX idx_exercise_to_tag_tag_id ON exercise_to_tag(tag_id);
CREATE INDEX idx_user_info_user_id ON user_info(user_id);
CREATE INDEX idx_recommendation_training_id ON recommendation(training_id);
CREATE INDEX idx_training_schedule_user_id ON training_schedule(user_id);
CREATE INDEX idx_user_favorite_workouts_user_id ON user_favorite_workouts(user_id);