package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/EnduranNSU/trainings/internal/adapter/out/postgres/gen"
	"github.com/EnduranNSU/trainings/internal/domain"
	"github.com/EnduranNSU/trainings/internal/logging"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/guregu/null/v6"
)

type TrainingRepositoryImpl struct {
	q  *gen.Queries
	db *sql.DB
}

func NewTrainingRepository(db *sql.DB) domain.TrainingRepository {
	return &TrainingRepositoryImpl{
		q:  gen.New(db),
		db: db,
	}
}

func (r *TrainingRepositoryImpl) GetTrainingsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Training, error) {
	trainings, err := r.q.GetTrainingsByUser(ctx, userID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetTrainingsByUser", jsonData, "failed to get trainings by user")
		return nil, err
	}

	result := make([]*domain.Training, len(trainings))
	for i, t := range trainings {
		result[i] = r.toDomainTraining(t)
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":         userID.String(),
		"trainings_count": len(result),
	})
	logging.Debug("GetTrainingsByUser", jsonData, "successfully retrieved user trainings")

	return result, nil
}

func (r *TrainingRepositoryImpl) GetTrainingWithExercises(ctx context.Context, trainingID int64) (*domain.Training, error) {
	training, err := r.q.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
		})
		logging.Error(err, "GetTrainingWithExercises", jsonData, "failed to get training with exercises")
		return nil, err
	}
	jsonData := logging.MarshalLogData(map[string]interface{}{
		"result": training,
	})
	logging.Debug("GetTrainingWithExercises", jsonData, "successfully retrieved training with exercises")


	domainTraining := r.toDomainTrainingFromJoined(training)

	jsonData = logging.MarshalLogData(map[string]interface{}{
		"training_id":     trainingID,
		"exercises_count": len(domainTraining.Exercises),
		"result": domainTraining,
	})
	logging.Debug("GetTrainingWithExercises", jsonData, "successfully retrieved training with exercises")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) CreateTraining(ctx context.Context, training *domain.Training) (*domain.Training, error) {
	params := gen.CreateTrainingParams{
		Title:             training.Title,
		UserID:            training.UserID,
		IsDone:            training.IsDone,
		PlannedDate:       training.PlannedDate,
		ActualDate:        null.TimeFromPtr(training.ActualDate).NullTime,
		StartedAt:         null.TimeFromPtr(training.StartedAt).NullTime,
		FinishedAt:        null.TimeFromPtr(training.FinishedAt).NullTime,
		TotalDuration:     durationToNullInt64(training.TotalDuration),
		TotalRestTime:     durationToNullInt64(training.TotalRestTime),
		TotalExerciseTime: durationToNullInt64(training.TotalExerciseTime),
		Rating:            null.Int32FromPtr(training.Rating).NullInt32,
	}

	created, err := r.q.CreateTraining(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": training.UserID.String(),
			"planned": training.PlannedDate,
			"is_done": training.IsDone,
		})
		logging.Error(err, "CreateTraining", jsonData, "failed to create training")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                created.ID,
		UserID:            created.UserID,
		IsDone:            created.IsDone,
		PlannedDate:       created.PlannedDate,
		ActualDate:        created.ActualDate,
		StartedAt:         created.StartedAt,
		FinishedAt:        created.FinishedAt,
		TotalDuration:     created.TotalDuration,
		TotalRestTime:     created.TotalRestTime,
		TotalExerciseTime: created.TotalExerciseTime,
		Rating:            created.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": domainTraining.ID,
		"user_id":     domainTraining.UserID.String(),
	})
	logging.Debug("CreateTraining", jsonData, "successfully created training")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) UpdateTraining(ctx context.Context, training *domain.Training) (*domain.Training, error) {
	params := gen.UpdateTrainingParams{
		IsDone:            training.IsDone,
		PlannedDate:       training.PlannedDate,
		ActualDate:        null.TimeFromPtr(training.ActualDate).NullTime,
		StartedAt:         null.TimeFromPtr(training.StartedAt).NullTime,
		FinishedAt:        null.TimeFromPtr(training.FinishedAt).NullTime,
		TotalDuration:     durationToNullInt64(training.TotalDuration),
		TotalRestTime:     durationToNullInt64(training.TotalRestTime),
		TotalExerciseTime: durationToNullInt64(training.TotalExerciseTime),
		Rating:            null.Int32FromPtr(training.Rating).NullInt32,
		ID:                training.ID,
		Title:             training.Title,
	}

	updated, err := r.q.UpdateTraining(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": training.ID,
			"is_done":     training.IsDone,
			"rating":      training.Rating,
		})
		logging.Error(err, "UpdateTraining", jsonData, "failed to update training")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                updated.ID,
		UserID:            updated.UserID,
		IsDone:            updated.IsDone,
		PlannedDate:       updated.PlannedDate,
		ActualDate:        updated.ActualDate,
		StartedAt:         updated.StartedAt,
		FinishedAt:        updated.FinishedAt,
		TotalDuration:     updated.TotalDuration,
		TotalRestTime:     updated.TotalRestTime,
		TotalExerciseTime: updated.TotalExerciseTime,
		Rating:            updated.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": domainTraining.ID,
	})
	logging.Debug("UpdateTraining", jsonData, "successfully updated training")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) DeleteTrainingAndExercises(ctx context.Context, trainingID int64) error {
	err := r.q.DeleteTrainingAndExercises(ctx, trainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
		})
		logging.Error(err, "DeleteTrainingAndExercises", jsonData, "failed to delete training and exercises")
		return err
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": trainingID,
	})
	logging.Debug("DeleteTrainingAndExercises", jsonData, "successfully deleted training and exercises")

	return nil
}

func (r *TrainingRepositoryImpl) AddExerciseToTraining(ctx context.Context, exercise *domain.TrainedExercise) (*domain.TrainedExercise, error) {
	weight := exercise.Weight.String()
	params := gen.AddExerciseToTrainingParams{
		TrainingID: exercise.TrainingID,
		ExerciseID: exercise.ExerciseID,
		Weight:     null.StringFromPtr(&weight).NullString,
		Approaches: null.Int32FromPtr(exercise.Approaches).NullInt32,
		Reps:       null.Int32FromPtr(exercise.Reps).NullInt32,
		Time:       durationToNullInt64(exercise.Time),
		Doing:      durationToNullInt64(exercise.Doing),
		Rest:       durationToNullInt64(exercise.Rest),
		Notes:      null.StringFromPtr(exercise.Notes).NullString,
	}
	created, err := r.q.AddExerciseToTraining(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": exercise.TrainingID,
			"exercise_id": exercise.ExerciseID,
		})
		logging.Error(err, "AddExerciseToTraining", jsonData, "failed to add exercise to training")
		return nil, err
	}

	domainExercise := r.toDomainTrainedExercise(created)

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"trained_exercise_id": domainExercise.ID,
		"training_id":         domainExercise.TrainingID,
		"exercise_id":         domainExercise.ExerciseID,
	})
	logging.Debug("AddExerciseToTraining", jsonData, "successfully added exercise to training")

	return domainExercise, nil
}

func (r *TrainingRepositoryImpl) UpdateTrainedExercise(ctx context.Context, exercise *domain.TrainedExercise) (*domain.TrainedExercise, error) {
	weight := exercise.Weight.String()
	params := gen.UpdateTrainedExerciseParams{
		Weight:     null.StringFromPtr(&weight).NullString,
		Approaches: null.Int32FromPtr(exercise.Approaches).NullInt32,
		Reps:       null.Int32FromPtr(exercise.Reps).NullInt32,
		Time:       durationToNullInt64(exercise.Time),
		Doing:      durationToNullInt64(exercise.Doing),
		Rest:       durationToNullInt64(exercise.Rest),
		Notes:      null.StringFromPtr(exercise.Notes).NullString,
		ID:         exercise.ID,
	}

	updated, err := r.q.UpdateTrainedExercise(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"trained_exercise_id": exercise.ID,
			"training_id":         exercise.TrainingID,
		})
		logging.Error(err, "UpdateTrainedExercise", jsonData, "failed to update trained exercise")
		return nil, err
	}

	domainExercise := r.toDomainTrainedExercise(gen.AddExerciseToTrainingRow{
		ID:         updated.ID,
		TrainingID: updated.TrainingID,
		ExerciseID: updated.ExerciseID,
		Weight:     updated.Weight,
		Approaches: updated.Approaches,
		Reps:       updated.Reps,
		Time:       updated.Time,
		Doing:      updated.Doing,
		Rest:       updated.Rest,
		Notes:      updated.Notes,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"trained_exercise_id": domainExercise.ID,
	})
	logging.Debug("UpdateTrainedExercise", jsonData, "successfully updated trained exercise")

	return domainExercise, nil
}

func (r *TrainingRepositoryImpl) DeleteExerciseFromTraining(ctx context.Context, exerciseID, trainingID int64) error {
	err := r.q.DeleteExerciseFromTraining(ctx, gen.DeleteExerciseFromTrainingParams{
		ID:         exerciseID,
		TrainingID: trainingID,
	})
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"trained_exercise_id": exerciseID,
			"training_id":         trainingID,
		})
		logging.Error(err, "DeleteExerciseFromTraining", jsonData, "failed to delete exercise from training")
		return err
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"trained_exercise_id": exerciseID,
		"training_id":         trainingID,
	})
	logging.Debug("DeleteExerciseFromTraining", jsonData, "successfully deleted exercise from training")

	return nil
}

func (r *TrainingRepositoryImpl) GetUserTrainingStats(ctx context.Context, userID uuid.UUID) (*domain.TrainingStats, error) {
	trainings, err := r.GetTrainingsByUser(ctx, userID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetUserTrainingStats", jsonData, "failed to get user training stats")
		return nil, err
	}

	stats := &domain.TrainingStats{}
	var totalRating int64
	var completedCount int

	for _, t := range trainings {
		stats.TotalTrainings++
		if t.IsDone {
			completedCount++
			if t.Rating != nil {
				totalRating += int64(*t.Rating)
			}
		}
	}

	stats.CompletedTrainings = int64(completedCount)
	if completedCount > 0 {
		stats.AverageRating = float64(totalRating) / float64(completedCount)
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":             userID.String(),
		"total_trainings":     stats.TotalTrainings,
		"completed_trainings": stats.CompletedTrainings,
		"average_rating":      stats.AverageRating,
	})
	logging.Debug("GetUserTrainingStats", jsonData, "successfully calculated user training stats")

	return stats, nil
}

func (r *TrainingRepositoryImpl) toDomainTraining(t gen.GetTrainingsByUserRow) *domain.Training {
	return &domain.Training{
		ID:                t.ID,
		Title:             t.Title,
		UserID:            t.UserID,
		IsDone:            t.IsDone,
		PlannedDate:       t.PlannedDate,
		ActualDate:        nullTimeFromSQL(t.ActualDate),
		StartedAt:         nullTimeFromSQL(t.StartedAt),
		FinishedAt:        nullTimeFromSQL(t.FinishedAt),
		TotalDuration:     toDuration(t.TotalDuration),
		TotalRestTime:     toDuration(t.TotalRestTime),
		TotalExerciseTime: toDuration(t.TotalExerciseTime),
		Rating:            nullIntFromSQL32(t.Rating),
	}
}

func (r *TrainingRepositoryImpl) toDomainTrainingFromJoined(t gen.GetTrainingWithExercisesRow) *domain.Training {
	training := &domain.Training{
		ID:                t.ID,
		Title:             t.Title,
		UserID:            t.UserID,
		IsDone:            t.IsDone,
		PlannedDate:       t.PlannedDate,
		ActualDate:        nullTimeFromSQL(t.ActualDate),
		StartedAt:         nullTimeFromSQL(t.StartedAt),
		FinishedAt:        nullTimeFromSQL(t.FinishedAt),
		TotalDuration:     toDuration(t.TotalDuration),
		TotalRestTime:     toDuration(t.TotalRestTime),
		TotalExerciseTime: toDuration(t.TotalExerciseTime),
		Rating:            nullIntFromSQL32(t.Rating),
		Exercises:         toDomainTrainedExercise(t.Exercises),
	}
	return training
}

func (r *TrainingRepositoryImpl) UpdateExerciseTime(ctx context.Context, exercise *domain.TrainedExercise) (*domain.TrainedExercise, error) {
	params := gen.UpdateExerciseTimeParams{
		Doing:      durationToNullInt64(exercise.Doing),
		Rest:       durationToNullInt64(exercise.Rest),
		Time:       durationToNullInt64(exercise.Time),
		ID:         exercise.ID,
	}

	updated, err := r.q.UpdateExerciseTime(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"trained_exercise_id": exercise.ID,
		})
		logging.Error(err, "UpdateExerciseTime", jsonData, "failed to update exercise time")
		return nil, err
	}

	domainExercise := r.toDomainTrainedExercise(gen.AddExerciseToTrainingRow{
		ID:         updated.ID,
		TrainingID: updated.TrainingID,
		ExerciseID: updated.ExerciseID,
		Weight:     updated.Weight,
		Approaches: updated.Approaches,
		Reps:       updated.Reps,
		Time:       updated.Time,
		Doing:      updated.Doing,
		Rest:       updated.Rest,
		Notes:      updated.Notes,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"trained_exercise_id": domainExercise.ID,
		"training_id":         domainExercise.TrainingID,
	})
	logging.Debug("UpdateExerciseTime", jsonData, "successfully updated exercise time")

	return domainExercise, nil
}

func (r *TrainingRepositoryImpl) UpdateTrainingTimers(ctx context.Context, training *domain.Training) (*domain.Training, error) {
	params := gen.UpdateTrainingTimersParams{
		StartedAt:         null.TimeFromPtr(training.StartedAt).NullTime,
		FinishedAt:        null.TimeFromPtr(training.FinishedAt).NullTime,
		TotalDuration:     durationToNullInt64(training.TotalDuration),
		TotalRestTime:     durationToNullInt64(training.TotalRestTime),
		TotalExerciseTime: durationToNullInt64(training.TotalExerciseTime),
		ID:                training.ID,
	}

	updated, err := r.q.UpdateTrainingTimers(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": training.ID,
		})
		logging.Error(err, "UpdateTrainingTimers", jsonData, "failed to update training timers")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                updated.ID,
		UserID:            updated.UserID,
		IsDone:            updated.IsDone,
		PlannedDate:       updated.PlannedDate,
		ActualDate:        updated.ActualDate,
		StartedAt:         updated.StartedAt,
		FinishedAt:        updated.FinishedAt,
		TotalDuration:     updated.TotalDuration,
		TotalRestTime:     updated.TotalRestTime,
		TotalExerciseTime: updated.TotalExerciseTime,
		Rating:            updated.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": domainTraining.ID,
	})
	logging.Debug("UpdateTrainingTimers", jsonData, "successfully updated training timers")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) GetCurrentTraining(ctx context.Context, userID uuid.UUID) (*domain.Training, error) {
	t, err := r.q.GetCurrentTraining(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetCurrentTraining", jsonData, "failed to get current training")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                t.ID,
		Title:             t.Title,
		UserID:            t.UserID,
		IsDone:            t.IsDone,
		PlannedDate:       t.PlannedDate,
		ActualDate:        t.ActualDate,
		StartedAt:         t.StartedAt,
		FinishedAt:        t.FinishedAt,
		TotalDuration:     t.TotalDuration,
		TotalRestTime:     t.TotalRestTime,
		TotalExerciseTime: t.TotalExerciseTime,
		Rating:            t.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":         userID.String(),
		"training_id":     domainTraining.ID,
		"exercises_count": len(domainTraining.Exercises),
	})
	logging.Debug("GetCurrentTraining", jsonData, "successfully retrieved current training")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) GetTodaysTraining(ctx context.Context, userID uuid.UUID) ([]*domain.Training, error) {
	trainingRows, err := r.q.GetTodaysTraining(ctx, userID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetTodaysTraining", jsonData, "failed to get today's training")
		return nil, err
	}

	trainings := make([]*domain.Training, len(trainingRows))
	for i, t := range trainingRows {
		trainings[i] = r.toDomainTraining(gen.GetTrainingsByUserRow{
			ID:                t.ID,
			Title:             t.Title,
			UserID:            t.UserID,
			IsDone:            t.IsDone,
			PlannedDate:       t.PlannedDate,
			ActualDate:        t.ActualDate,
			StartedAt:         t.StartedAt,
			FinishedAt:        t.FinishedAt,
			TotalDuration:     t.TotalDuration,
			TotalRestTime:     t.TotalRestTime,
			TotalExerciseTime: t.TotalExerciseTime,
			Rating:            t.Rating,
		})
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":         userID.String(),
		"trainings_count": len(trainings),
	})
	logging.Debug("GetTodaysTraining", jsonData, "successfully retrieved today's training")

	return trainings, nil
}

func (r *TrainingRepositoryImpl) GetGlobalTrainings(ctx context.Context) ([]*domain.GlobalTraining, error) {
	globalTrainingRows, err := r.q.GetGlobalTrainings(ctx)
	if err != nil {
		logging.Error(err, "GetGlobalTrainings", nil, "failed to get global trainings")
		return nil, err
	}

	globalTrainings := make([]*domain.GlobalTraining, len(globalTrainingRows))
	for i, gt := range globalTrainingRows {
		globalTrainings[i] = r.toDomainGlobalTraining(GlobalTrainingRow{
			ID:          gt.ID,
			Title:       gt.Title,
			Description: gt.Description,
			Level:       gt.Level,
			Exercises:   gt.Exercises,
		})
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"global_trainings_count": len(globalTrainings),
	})
	logging.Debug("GetGlobalTrainings", jsonData, "successfully retrieved global trainings")

	return globalTrainings, nil
}

func (r *TrainingRepositoryImpl) GetGlobalTrainingById(ctx context.Context, trainingID int64) (*domain.GlobalTraining, error) {
	gt, err := r.q.GetGlobalTrainingByID(ctx, trainingID)
	if err != nil {
		logging.Error(err, "GetGlobalTrainings", nil, "failed to get global trainings")
		return nil, err
	}

	globalTraining := r.toDomainGlobalTraining(GlobalTrainingRow{
		ID:          gt.ID,
		Title:       gt.Title,
		Description: gt.Description,
		Level:       gt.Level,
		Exercises:   gt.Exercises,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"Id":        globalTraining.ID,
		"Level":     globalTraining.Level,
		"Exercises": globalTraining.Exercises,
	})
	logging.Debug("GetGlobalTrainings", jsonData, "successfully retrieved global trainings")

	return globalTraining, nil
}

func (r *TrainingRepositoryImpl) GetGlobalTrainingByLevel(ctx context.Context, level string) ([]*domain.GlobalTraining, error) {
	globalTrainingRows, err := r.q.GetGlobalTrainingByLevel(ctx, level)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"level": level,
		})
		logging.Error(err, "GetGlobalTrainingByLevel", jsonData, "failed to get global training by level")
		return nil, err
	}

	globalTrainings := make([]*domain.GlobalTraining, len(globalTrainingRows))
	for i, gt := range globalTrainingRows {
		globalTrainings[i] = r.toDomainGlobalTraining(GlobalTrainingRow{
			ID:          gt.ID,
			Title:       gt.Title,
			Description: gt.Description,
			Level:       gt.Level,
			Exercises:   gt.Exercises,
		})
	}
	jsonData := logging.MarshalLogData(map[string]interface{}{
		"trainings_count": len(globalTrainings),
	})
	logging.Debug("GetGlobalTrainingByLevel", jsonData, "successfully retrieved global training by level")

	return globalTrainings, nil
}

func (r *TrainingRepositoryImpl) MarkTrainingAsDone(ctx context.Context, trainingID int64, userID uuid.UUID) (*domain.Training, error) {
	finishedAt := null.TimeFromPtr(nil)
	params := gen.MarkTrainingAsDoneParams{
		FinishedAt: finishedAt.NullTime,
		ID:         trainingID,
		UserID:     userID,
	}

	updated, err := r.q.MarkTrainingAsDone(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
			"user_id":     userID.String(),
		})
		logging.Error(err, "MarkTrainingAsDone", jsonData, "failed to mark training as done")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                updated.ID,
		Title:             updated.Title,
		UserID:            updated.UserID,
		IsDone:            updated.IsDone,
		PlannedDate:       updated.PlannedDate,
		ActualDate:        updated.ActualDate,
		StartedAt:         updated.StartedAt,
		FinishedAt:        updated.FinishedAt,
		TotalDuration:     updated.TotalDuration,
		TotalRestTime:     updated.TotalRestTime,
		TotalExerciseTime: updated.TotalExerciseTime,
		Rating:            updated.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": domainTraining.ID,
	})
	logging.Debug("MarkTrainingAsDone", jsonData, "successfully marked training as done")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) GetTrainingStats(ctx context.Context, trainingID int64) (*domain.TrainingStats, error) {
	statsRow, err := r.q.GetTrainingStats(ctx, trainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
		})
		logging.Error(err, "GetTrainingStats", jsonData, "failed to get training stats")
		return nil, err
	}

	training, err := r.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, err
	}

	var totalDuration time.Duration
	if training.TotalDuration != nil {
		totalDuration = *training.TotalDuration
	}

	stats := &domain.TrainingStats{
		TotalTrainings:     1,
		CompletedTrainings: 0,
		AverageRating:      float64(*training.Rating),
		TotalDuration:      totalDuration,
	}

	if training.IsDone {
		stats.CompletedTrainings = 1
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id":        trainingID,
		"exercise_count":     statsRow.ExerciseCount,
		"total_approaches":   statsRow.TotalApproaches,
		"total_reps":         statsRow.TotalReps,
		"total_duration_sec": totalDuration.Seconds(),
	})
	logging.Debug("GetTrainingStats", jsonData, "successfully retrieved training stats")

	return stats, nil
}

func (r *TrainingRepositoryImpl) StartTraining(ctx context.Context, trainingID int64, userID uuid.UUID) (*domain.Training, error) {
	startedAt := null.TimeFromPtr(nil)
	params := gen.StartTrainingParams{
		StartedAt: startedAt.NullTime,
		ID:        trainingID,
		UserID:    userID,
	}

	updated, err := r.q.StartTraining(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
			"user_id":     userID.String(),
		})
		logging.Error(err, "StartTraining", jsonData, "failed to start training")
		return nil, err
	}

	domainTraining := r.toDomainTraining(gen.GetTrainingsByUserRow{
		ID:                updated.ID,
		Title:             updated.Title,
		UserID:            updated.UserID,
		IsDone:            updated.IsDone,
		PlannedDate:       updated.PlannedDate,
		ActualDate:        updated.ActualDate,
		StartedAt:         updated.StartedAt,
		FinishedAt:        updated.FinishedAt,
		TotalDuration:     updated.TotalDuration,
		TotalRestTime:     updated.TotalRestTime,
		TotalExerciseTime: updated.TotalExerciseTime,
		Rating:            updated.Rating,
	})

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id": domainTraining.ID,
		"started_at":  domainTraining.StartedAt,
	})
	logging.Debug("StartTraining", jsonData, "successfully started training")

	return domainTraining, nil
}

func (r *TrainingRepositoryImpl) toDomainTrainedExercise(ex gen.AddExerciseToTrainingRow) *domain.TrainedExercise {
	weight, _ := decimal.NewFromString(ex.Weight.String)
	return &domain.TrainedExercise{
		ID:         ex.ID,
		TrainingID: ex.TrainingID,
		ExerciseID: ex.ExerciseID,
		Weight:     &weight,
		Approaches: nullIntFromSQL32(ex.Approaches),
		Reps:       nullIntFromSQL32(ex.Reps),
		Time:       toDuration(ex.Time),
		Doing:      toDuration(ex.Doing),
		Rest:       toDuration(ex.Rest),
		Notes:      nullStringFromSQL(ex.Notes),
	}
}

func (r *TrainingRepositoryImpl) CalculateTrainingTotalTime(ctx context.Context, trainingID int64) (*domain.TrainingTime, error) {
	timeStats, err := r.q.CalculateTrainingTotalTime(ctx, trainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": trainingID,
		})
		logging.Error(err, "CalculateTrainingTotalTime", jsonData, "failed to calculate training total time")
		return nil, err
	}

	// Вспомогательная функция для безопасного преобразования
	convertToInt64 := func(val interface{}) int64 {
		if val == nil {
			return 0
		}
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		default:
			return 0
		}
	}

	trainingTime := &domain.TrainingTime{
		TotalExerciseSeconds: convertToInt64(timeStats.TotalExerciseSeconds),
		TotalRestSeconds:     convertToInt64(timeStats.TotalRestSeconds),
		TotalSeconds:         convertToInt64(timeStats.TotalSeconds),
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"training_id":            trainingID,
		"total_exercise_seconds": trainingTime.TotalExerciseSeconds,
		"total_rest_seconds":     trainingTime.TotalRestSeconds,
		"total_seconds":          trainingTime.TotalSeconds,
	})
	logging.Debug("CalculateTrainingTotalTime", jsonData, "successfully calculated training total time")

	return trainingTime, nil
}

type GlobalTrainingRow struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Level       string      `json:"level"`
	Exercises   interface{} `json:"exercises"`
}

func (r *TrainingRepositoryImpl) toDomainGlobalTraining(gt GlobalTrainingRow) *domain.GlobalTraining {
	return &domain.GlobalTraining{
		ID:          gt.ID,
		Title:       gt.Title,
		Description: gt.Description,
		Level:       gt.Level,
		Exercises:   toDomainExercise(gt.Exercises),
	}
}

func (r *TrainingRepositoryImpl) AssignGlobalTrainingToUser(ctx context.Context, cmd domain.AssignGlobalTrainingCmd) (*domain.Training, error) {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id":            cmd.UserID.String(),
			"global_training_id": cmd.GlobalTrainingID,
			"planned_date":       cmd.PlannedDate.Format("2006-01-02"),
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	q := r.q.WithTx(tx)

	// 1. Получаем информацию о глобальной тренировке
	globalTraining, err := q.GetGlobalTrainingById(ctx, cmd.GlobalTrainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id":            cmd.UserID.String(),
			"global_training_id": cmd.GlobalTrainingID,
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to get global training")
		return nil, err
	}

	// 2. Создаем тренировку для пользователя на указанную дату
	trainingParams := gen.CreateTrainingParams{
		UserID:      cmd.UserID,
		Title:       globalTraining.Title,
		IsDone:      false,
		PlannedDate: cmd.PlannedDate,
		// Если тренировка на сегодня, устанавливаем actual_date
		ActualDate: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: cmd.PlannedDate.Equal(time.Now().UTC().Truncate(24 * time.Hour)),
		},
		StartedAt:  null.TimeFromPtr(nil).NullTime,
		FinishedAt: null.TimeFromPtr(nil).NullTime,
		Rating:     null.Int32FromPtr(nil).NullInt32,
	}

	createdTraining, err := q.CreateTraining(ctx, trainingParams)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id":      cmd.UserID.String(),
			"planned_date": cmd.PlannedDate.Format("2006-01-02"),
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to create training")
		return nil, err
	}

	// 3. Получаем упражнения из глобальной тренировки
	globalExercises, err := q.GetGlobalTrainingExercises(ctx, cmd.GlobalTrainingID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id":            cmd.UserID.String(),
			"global_training_id": cmd.GlobalTrainingID,
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to get global training exercises")
		return nil, err
	}

	// 4. Добавляем упражнения в пользовательскую тренировку
	for _, globalExercise := range globalExercises {
		exerciseParams := gen.AddExerciseToTrainingParams{
			TrainingID: createdTraining.ID,
			ExerciseID: globalExercise.ExerciseID,
			Weight:     null.StringFromPtr(nil).NullString,
			Approaches: null.Int32FromPtr(nil).NullInt32,
			Reps:       null.Int32FromPtr(nil).NullInt32,
			Time:       sql.NullInt64{Valid: false},
			Doing:      sql.NullInt64{Valid: false},
			Rest:       sql.NullInt64{Valid: false},
			Notes:      null.StringFromPtr(nil).NullString,
		}

		_, err := q.AddExerciseToTraining(ctx, exerciseParams)
		if err != nil {
			jsonData := logging.MarshalLogData(map[string]interface{}{
				"training_id": createdTraining.ID,
				"exercise_id": globalExercise.ExerciseID,
			})
			logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to add exercise to training")
			return nil, err
		}
	}

	// 5. Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id":            cmd.UserID.String(),
			"global_training_id": cmd.GlobalTrainingID,
			"training_id":        createdTraining.ID,
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to commit transaction")
		return nil, err
	}

	// 6. Получаем полную информацию о созданной тренировке
	fullTraining, err := r.GetTrainingWithExercises(ctx, createdTraining.ID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"training_id": createdTraining.ID,
		})
		logging.Error(err, "AssignGlobalTrainingToUser", jsonData, "failed to get full training info")
		return nil, err
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":               cmd.UserID.String(),
		"global_training_id":    cmd.GlobalTrainingID,
		"global_training_level": globalTraining.Level,
		"training_id":           fullTraining.ID,
		"planned_date":          cmd.PlannedDate.Format("2006-01-02"),
		"exercises_count":       len(fullTraining.Exercises),
		"is_today":              cmd.PlannedDate.Equal(time.Now().UTC().Truncate(24 * time.Hour)),
	})
	logging.Debug("AssignGlobalTrainingToUser", jsonData, "successfully assigned global training to user")

	return fullTraining, nil
}