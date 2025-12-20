package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TrainingService interface {
	GetTrainingsByUser(ctx context.Context, userID uuid.UUID) ([]*Training, error)
	GetTrainingWithExercises(ctx context.Context, trainingID int64) (*Training, error)
	CreateTraining(ctx context.Context, cmd CreateTrainingCmd) (*Training, error)
	UpdateTraining(ctx context.Context, cmd UpdateTrainingCmd) (*Training, error)
	DeleteTraining(ctx context.Context, trainingID int64) error
	AddExerciseToTraining(ctx context.Context, cmd AddExerciseToTrainingCmd) (*TrainedExercise, error)
	UpdateTrainedExercise(ctx context.Context, cmd UpdateTrainedExerciseCmd) (*TrainedExercise, error)
	RemoveExerciseFromTraining(ctx context.Context, trainingID, exerciseID int64) error
	GetUserTrainingStats(ctx context.Context, userID uuid.UUID) (*TrainingStats, error)
	CompleteTraining(ctx context.Context, trainingID int64, rating *int32) (*Training, error)

	UpdateExerciseTime(ctx context.Context, exerciseID int64, weight *decimal.Decimal, approaches *int32, reps *int32, time *time.Duration, doing *time.Duration, rest *time.Duration) (*TrainedExercise, error)
	UpdateTrainingTimers(ctx context.Context, trainingID int64, totalDuration *time.Duration, totalRestTime *time.Duration, totalExerciseTime *time.Duration) (*Training, error)
	CalculateTrainingTotalTime(ctx context.Context, trainingID int64) (*TrainingTime, error)
	GetCurrentTraining(ctx context.Context, userID uuid.UUID) (*Training, error)
	GetTodaysTraining(ctx context.Context, userID uuid.UUID) ([]*Training, error)

	GetGlobalTrainings(ctx context.Context) ([]*GlobalTraining, error)
	GetGlobalTrainingByLevel(ctx context.Context, level string) ([]*GlobalTraining, error)
	GetGlobalTrainingById(ctx context.Context, trainingID int64) (*GlobalTraining, error)
	AssignGlobalTraining(ctx context.Context, cmd AssignGlobalTrainingCmd) (*Training, error)

	MarkTrainingAsDone(ctx context.Context, trainingID int64, userID uuid.UUID) (*Training, error)
	GetTrainingStats(ctx context.Context, trainingID int64) (*TrainingStats, error)
	StartTraining(ctx context.Context, trainingID int64, userID uuid.UUID) (*Training, error)
	UpdateExerciseRestTime(ctx context.Context, exerciseID int64, restTime time.Duration) (*TrainedExercise, error)
	UpdateExerciseDoingTime(ctx context.Context, exerciseID int64, doingTime time.Duration) (*TrainedExercise, error)
	PauseTraining(ctx context.Context, trainingID int64) (*Training, error)
	ResumeTraining(ctx context.Context, trainingID int64) (*Training, error)
}

type CreateTrainingCmd struct {
	UserID            uuid.UUID
	Title             string
	IsDone            bool
	PlannedDate       time.Time
	ActualDate        *time.Time
	StartedAt         *time.Time
	FinishedAt        *time.Time
	TotalDuration     *time.Duration
	TotalRestTime     *time.Duration
	TotalExerciseTime *time.Duration
	Rating            *int32
}

type UpdateTrainingCmd struct {
	ID                int64
	Title             string
	IsDone            *bool
	PlannedDate       time.Time
	ActualDate        *time.Time
	StartedAt         *time.Time
	FinishedAt        *time.Time
	TotalDuration     *time.Duration
	TotalRestTime     *time.Duration
	TotalExerciseTime *time.Duration
	Rating            *int32
}

type AddExerciseToTrainingCmd struct {
	TrainingID int64
	ExerciseID int64
	Weight     *decimal.Decimal
	Approaches *int32
	Reps       *int32
	Time       *time.Duration
	Doing      *time.Duration
	Rest       *time.Duration
	Notes      *string
}

type UpdateTrainedExerciseCmd struct {
	ID         int64
	Weight     *decimal.Decimal
	Approaches *int32
	Reps       *int32
	Time       *time.Duration
	Doing      *time.Duration
	Rest       *time.Duration
	Notes      *string
}

type AssignGlobalTrainingCmd struct {
	UserID           uuid.UUID
	GlobalTrainingID int64
	PlannedDate      time.Time // Дата, на которую назначается тренировка
}

type ExerciseService interface {
	GetAllExercises(ctx context.Context) ([]*Exercise, error)
	GetExerciseByID(ctx context.Context, id int64) (*Exercise, error)
	GetExercisesByTag(ctx context.Context, tagID int64) ([]*Exercise, error)
	SearchExercises(ctx context.Context, query string, tagID *int64) ([]*Exercise, error)
	GetAllTags(ctx context.Context) ([]*Tag, error)
	GetTagByID(ctx context.Context, id int64) (*Tag, error)
	GetExerciseTags(ctx context.Context, exerciseID int64) ([]*Tag, error)
	GetExercisesByMultipleTags(ctx context.Context, tagIDs []int64) ([]*Exercise, error)
	GetPopularTags(ctx context.Context, limit int) ([]*Tag, error)
}
