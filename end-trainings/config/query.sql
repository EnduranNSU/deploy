-- name: GetExercisesWithTags :many
SELECT 
    e.id,
    e.description,
    e.href,
    COALESCE(
        json_agg(
            json_build_object(
                'id', t.id,
                'type', t.type
            )
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'
    ) as tags
FROM exercise e
LEFT JOIN exercise_to_tag et ON e.id = et.exercise_id
LEFT JOIN tag t ON et.tag_id = t.id
GROUP BY e.id, e.description, e.href
ORDER BY e.id;

-- name: GetExerciseByID :one
SELECT 
    e.id,
    e.description,
    e.href,
    COALESCE(
        json_agg(
            json_build_object(
                'id', t.id,
                'type', t.type
            )
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'
    ) as tags
FROM exercise e
LEFT JOIN exercise_to_tag et ON e.id = et.exercise_id
LEFT JOIN tag t ON et.tag_id = t.id
WHERE e.id = $1
GROUP BY e.id, e.description, e.href;

-- name: GetAllTags :many
SELECT id, type FROM tag ORDER BY id;

-- name: GetExercisesByTag :many
SELECT 
    e.id,
    e.description,
    e.href
FROM exercise e
INNER JOIN exercise_to_tag et ON e.id = et.exercise_id
WHERE et.tag_id = $1
ORDER BY e.id;

-- name: GetTrainingsByUser :many
SELECT 
    t.id,
    t.user_id,
    t.is_done,
    t.planned_date,
    t.actual_date,
    t.started_at,
    t.finished_at,
    t.total_duration,
    t.total_rest_time,
    t.total_exercise_time,
    t.rating,
    COALESCE(
        json_agg(
            json_build_object(
                'id', te.id,
                'exercise_id', te.exercise_id,
                'weight', te.weight,
                'approaches', te.approaches,
                'reps', te.reps,
                'time', te.time,
                'doing', te.doing,
                'rest', te.rest,
                'notes', te.notes
            )
        ) FILTER (WHERE te.id IS NOT NULL),
        '[]'
    ) as exercises
FROM training t
LEFT JOIN trained_exercise te ON t.id = te.training_id
WHERE t.user_id = $1
GROUP BY t.id
ORDER BY t.planned_date DESC;

-- name: CreateTraining :one
INSERT INTO training (
    user_id,
    is_done,
    planned_date,
    actual_date,
    started_at,
    finished_at,
    total_duration,
    total_rest_time,
    total_exercise_time,
    rating
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: AddExerciseToTraining :one
INSERT INTO trained_exercise (
    training_id,
    exercise_id,
    weight,
    approaches,
    reps,
    time,
    doing,
    rest,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateTrainedExercise :one
UPDATE trained_exercise
SET 
    weight = COALESCE($1, weight),
    approaches = COALESCE($2, approaches),
    reps = COALESCE($3, reps),
    time = COALESCE($4, time),
    doing = COALESCE($5, doing),
    rest = COALESCE($6, rest),
    notes = COALESCE($7, notes)
WHERE id = $8
RETURNING *;

-- name: UpdateTraining :one
UPDATE training
SET 
    is_done = COALESCE($1, is_done),
    planned_date = COALESCE($2, planned_date),
    actual_date = COALESCE($3, actual_date),
    started_at = COALESCE($4, started_at),
    finished_at = COALESCE($5, finished_at),
    total_duration = COALESCE($6, total_duration),
    total_rest_time = COALESCE($7, total_rest_time),
    total_exercise_time = COALESCE($8, total_exercise_time),
    rating = COALESCE($9, rating)
WHERE id = $10
RETURNING *;

-- name: GetTrainingWithExercises :one
SELECT 
    t.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', te.id,
                'exercise_id', te.exercise_id,
                'weight', te.weight,
                'approaches', te.approaches,
                'reps', te.reps,
                'time', te.time,
                'doing', te.doing,
                'rest', te.rest,
                'notes', te.notes
            )
        ) FILTER (WHERE te.id IS NOT NULL),
        '[]'
    ) as exercises
FROM training t
LEFT JOIN trained_exercise te ON t.id = te.training_id
WHERE t.id = $1
GROUP BY t.id;

-- name: DeleteExerciseFromTraining :exec
DELETE FROM trained_exercise 
WHERE id = $1 AND training_id = $2;

-- name: DeleteTrainingAndExercises :exec
WITH deleted_exercises AS (
    DELETE FROM trained_exercise WHERE training_id = $1
)
DELETE FROM training WHERE training.id = $1;


-- name: UpdateExerciseTime :one
-- Обновление времени выполнения упражнения (doing) и времени отдыха (rest)
UPDATE trained_exercise
SET 
    doing = COALESCE($1, doing),
    rest = COALESCE($2, rest),
    time = COALESCE($3, time)  -- Общее время упражнения (doing + rest)
WHERE id = $4 AND training_id = $5
RETURNING *;

-- name: UpdateTrainingTimers :one
-- Обновление времени тренировки (старт, финиш, общая продолжительность)
UPDATE training
SET 
    started_at = COALESCE($1, started_at),
    finished_at = COALESCE($2, finished_at),
    total_duration = COALESCE($3, total_duration),
    total_rest_time = COALESCE($4, total_rest_time),
    total_exercise_time = COALESCE($5, total_exercise_time)
WHERE id = $6
RETURNING *;

-- name: CalculateTrainingTotalTime :one
-- Расчет общего времени тренировки на основе всех упражнений
SELECT 
    COALESCE(SUM(EXTRACT(EPOCH FROM te.doing)), 0) as total_exercise_seconds,
    COALESCE(SUM(EXTRACT(EPOCH FROM te.rest)), 0) as total_rest_seconds,
    COALESCE(SUM(EXTRACT(EPOCH FROM te.doing)) + SUM(EXTRACT(EPOCH FROM te.rest)), 0) as total_seconds
FROM trained_exercise te
WHERE te.training_id = $1;

-- name: GetCurrentTraining :one
-- Получение тренировки на сегодня для пользователя
SELECT 
    t.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', te.id,
                'exercise_id', te.exercise_id,
                'weight', te.weight,
                'approaches', te.approaches,
                'reps', te.reps,
                'time', te.time,
                'doing', te.doing,
                'rest', te.rest,
                'notes', te.notes
            )
        ) FILTER (WHERE te.id IS NOT NULL),
        '[]'
    ) as exercises
FROM training t
LEFT JOIN trained_exercise te ON t.id = te.training_id
WHERE t.user_id = $1 
    AND t.planned_date = CURRENT_DATE
    AND t.is_done = false
GROUP BY t.id
ORDER BY t.planned_date DESC
LIMIT 1;

-- name: GetTodaysTraining :many
-- Получение всех тренировок на сегодня для пользователя
SELECT 
    t.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', te.id,
                'exercise_id', te.exercise_id,
                'weight', te.weight,
                'approaches', te.approaches,
                'reps', te.reps,
                'time', te.time,
                'doing', te.doing,
                'rest', te.rest,
                'notes', te.notes
            )
        ) FILTER (WHERE te.id IS NOT NULL),
        '[]'
    ) as exercises
FROM training t
LEFT JOIN trained_exercise te ON t.id = te.training_id
WHERE t.user_id = $1 
    AND t.planned_date = CURRENT_DATE
GROUP BY t.id
ORDER BY t.planned_date DESC;

-- name: GetGlobalTrainings :many
-- Получение всех глобальных тренировок с упражнениями и их тегами
SELECT 
    gt.id,
    gt.level,
    COALESCE(
        json_agg(
            json_build_object(
                'id', e.id,
                'description', e.description,
                'href', e.href,
                'tags', COALESCE(
                    (
                        SELECT json_agg(
                            json_build_object(
                                'id', t.id,
                                'type', t.type
                            )
                        )
                        FROM exercise_to_tag et2
                        JOIN tag t ON et2.tag_id = t.id
                        WHERE et2.exercise_id = e.id
                    ),
                    '[]'
                )
            )
        ) FILTER (WHERE e.id IS NOT NULL),
        '[]'
    ) as exercises
FROM global_training gt
LEFT JOIN global_training_exercise gte ON gt.id = gte.global_training_id
LEFT JOIN exercise e ON gte.exercise_id = e.id
GROUP BY gt.id, gt.level
ORDER BY 
    CASE gt.level 
        WHEN 'beginner' THEN 1
        WHEN 'intermediate' THEN 2
        WHEN 'advanced' THEN 3
    END;

-- name: GetGlobalTrainingByID :one
-- Получение глобальной тренировки по ID с упражнениями и их тегами
SELECT 
    gt.id,
    gt.level,
    COALESCE(
        json_agg(
            json_build_object(
                'id', e.id,
                'description', e.description,
                'href', e.href,
                'tags', COALESCE(
                    (
                        SELECT json_agg(
                            json_build_object(
                                'id', t.id,
                                'type', t.type
                            )
                        )
                        FROM exercise_to_tag et2
                        JOIN tag t ON et2.tag_id = t.id
                        WHERE et2.exercise_id = e.id
                    ),
                    '[]'
                )
            )
        ) FILTER (WHERE e.id IS NOT NULL),
        '[]'
    ) as exercises
FROM global_training gt
LEFT JOIN global_training_exercise gte ON gt.id = gte.global_training_id
LEFT JOIN exercise e ON gte.exercise_id = e.id
WHERE gt.id = $1
GROUP BY gt.id, gt.level;

-- name: GetGlobalTrainingByLevel :many
-- Получение глобальных тренировок по уровню с упражнениями и их тегами
SELECT 
    gt.id,
    gt.level,
    COALESCE(
        json_agg(
            json_build_object(
                'id', e.id,
                'description', e.description,
                'href', e.href,
                'tags', COALESCE(
                    (
                        SELECT json_agg(
                            json_build_object(
                                'id', t.id,
                                'type', t.type
                            )
                        )
                        FROM exercise_to_tag et2
                        JOIN tag t ON et2.tag_id = t.id
                        WHERE et2.exercise_id = e.id
                    ),
                    '[]'
                )
            )
        ) FILTER (WHERE e.id IS NOT NULL),
        '[]'
    ) as exercises
FROM global_training gt
LEFT JOIN global_training_exercise gte ON gt.id = gte.global_training_id
LEFT JOIN exercise e ON gte.exercise_id = e.id
WHERE gt.level = $1
GROUP BY gt.id, gt.level
ORDER BY gt.id;

-- name: MarkTrainingAsDone :one
-- Отметить тренировку как выполненную
UPDATE training
SET 
    is_done = true,
    actual_date = CURRENT_DATE,
    finished_at = COALESCE($1, CURRENT_TIMESTAMP)
WHERE id = $2 AND user_id = $3
RETURNING *;

-- name: GetTrainingStats :one
-- Получение статистики по тренировке (общее время выполнения и отдыха)
SELECT 
    t.id,
    t.total_duration,
    t.total_rest_time,
    t.total_exercise_time,
    COUNT(te.id) as exercise_count,
    COALESCE(SUM(te.approaches), 0) as total_approaches,
    COALESCE(SUM(te.reps), 0) as total_reps
FROM training t
LEFT JOIN trained_exercise te ON t.id = te.training_id
WHERE t.id = $1
GROUP BY t.id;

-- name: StartTraining :one
-- Начать тренировку (установить время начала)
UPDATE training
SET 
    started_at = COALESCE($1, CURRENT_TIMESTAMP),
    is_done = false
WHERE id = $2 AND user_id = $3
RETURNING *;

-- name: GetGlobalTrainingById :one
SELECT id, level
FROM global_training
WHERE id = $1;

-- name: GetGlobalTrainingExercises :many
SELECT gte.id, gte.global_training_id, gte.exercise_id
FROM global_training_exercise gte
WHERE gte.global_training_id = $1;