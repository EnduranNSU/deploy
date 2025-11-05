CREATE TABLE "user"(
    "id" UUID NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "tg_id" UUID NOT NULL
);
ALTER TABLE
    "user" ADD PRIMARY KEY("id");

CREATE TABLE "exercise"(
    "id" BIGINT NOT NULL,
    "description" TEXT NOT NULL,
    "href" TEXT NOT NULL,
    "tags" BIGINT NOT NULL
);
ALTER TABLE
    "exercise" ADD PRIMARY KEY("id");

CREATE TABLE "tag"(
    "id" BIGINT NOT NULL,
    "type" VARCHAR(255) CHECK
        ("type" IN('')) NOT NULL
);
ALTER TABLE
    "tag" ADD PRIMARY KEY("id");

CREATE TABLE "exercise_to_tag"(
    "execrise_id" BIGINT NOT NULL,
    "tag_id" BIGINT NOT NULL
);

CREATE TABLE "training"(
    "id" BIGINT NOT NULL,
    "user_id" UUID NOT NULL,
    "isDone" BOOLEAN NOT NULL,
    "planned" DATE NOT NULL,
    "done" DATE NULL,
    "total_time" INTERVAL NULL,
    "rating" INTEGER NULL
);
ALTER TABLE
    "training" ADD PRIMARY KEY("id");

CREATE TABLE "trained_exercise"(
    "id" BIGINT NOT NULL,
    "training_id" BIGINT NOT NULL,
    "exercise_id" BIGINT NOT NULL,
    "weight" FLOAT(53) NULL,
    "approaches" BIGINT NULL,
    "reps" BIGINT NULL,
    "time" TIME(0) WITHOUT TIME ZONE NULL,
    "notes" TEXT NULL
);
ALTER TABLE
    "trained_exercise" ADD PRIMARY KEY("id");

CREATE TABLE "recommedantion"(
    "id" BIGINT NOT NULL,
    "training_id" BIGINT NOT NULL,
    "approach" BIGINT NULL,
    "weight" FLOAT(53) NULL,
    "time" TIME(0) WITHOUT TIME ZONE NOT NULL,
    "new_column" BIGINT NOT NULL
);
ALTER TABLE
    "recommedantion" ADD PRIMARY KEY("id");

CREATE TABLE "user_info"(
    "id" BIGINT NOT NULL,
    "weight" FLOAT(53) NOT NULL,
    "height" BIGINT NOT NULL,
    "date" DATE NOT NULL,
    "age" BIGINT NOT NULL,
    "user_id" UUID NOT NULL
);
ALTER TABLE
    "user_info" ADD PRIMARY KEY("id");

ALTER TABLE
    "training" ADD CONSTRAINT "training_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");
ALTER TABLE
    "recommedantion" ADD CONSTRAINT "recommedantion_training_id_foreign" FOREIGN KEY("training_id") REFERENCES "training"("id");
ALTER TABLE
    "user_info" ADD CONSTRAINT "user_info_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "user"("id");
ALTER TABLE
    "exercise_to_tag" ADD CONSTRAINT "exercise_to_tag_tag_id_foreign" FOREIGN KEY("tag_id") REFERENCES "tag"("id");
ALTER TABLE
    "exercise_to_tag" ADD CONSTRAINT "exercise_to_tag_execrise_id_foreign" FOREIGN KEY("execrise_id") REFERENCES "exercise"("id");
ALTER TABLE
    "trained_exercise" ADD CONSTRAINT "trained_exercise_exercise_id_foreign" FOREIGN KEY("exercise_id") REFERENCES "exercise"("id");
ALTER TABLE
    "trained_exercise" ADD CONSTRAINT "trained_exercise_training_id_foreign" FOREIGN KEY("training_id") REFERENCES "training"("id");