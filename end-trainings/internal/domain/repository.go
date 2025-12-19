package domain

import (
	"context"

	"github.com/google/uuid"
)

type TrainingRepository interface {
	// Тренировки
	GetTrainingsByUser(ctx context.Context, userID uuid.UUID) ([]*Training, error)
	GetTrainingWithExercises(ctx context.Context, trainingID int64) (*Training, error)
	CreateTraining(ctx context.Context, training *Training) (*Training, error)
	UpdateTraining(ctx context.Context, training *Training) (*Training, error)
	DeleteTrainingAndExercises(ctx context.Context, trainingID int64) error
	
	// Упражнения в тренировках
	AddExerciseToTraining(ctx context.Context, exercise *TrainedExercise) (*TrainedExercise, error)
	UpdateTrainedExercise(ctx context.Context, exercise *TrainedExercise) (*TrainedExercise, error)
	DeleteExerciseFromTraining(ctx context.Context, exerciseID, trainingID int64) error
	
	// Статистика
	GetUserTrainingStats(ctx context.Context, userID uuid.UUID) (*TrainingStats, error)

	// Таймер
	UpdateExerciseTime(ctx context.Context, exercise *TrainedExercise) (*TrainedExercise, error)
	UpdateTrainingTimers(ctx context.Context, training *Training) (*Training, error)
	CalculateTrainingTotalTime(ctx context.Context, trainingID int64) (*TrainingTime, error)
	
	// Актуальные тренировки
	GetCurrentTraining(ctx context.Context, userID uuid.UUID) (*Training, error)
	GetTodaysTraining(ctx context.Context, userID uuid.UUID) ([]*Training, error)
	
	// Популярные/известные
	GetGlobalTrainings(ctx context.Context) ([]*GlobalTraining, error)
	GetGlobalTrainingByLevel(ctx context.Context, level string) ([]*GlobalTraining, error)
	GetGlobalTrainingById(ctx context.Context, trainingID int64) (*GlobalTraining, error)
	
	//Прогресс тренировки
	MarkTrainingAsDone(ctx context.Context, trainingID int64, userID uuid.UUID) (*Training, error)
	GetTrainingStats(ctx context.Context, trainingID int64) (*TrainingStats, error)
	StartTraining(ctx context.Context, trainingID int64, userID uuid.UUID) (*Training, error)

	AssignGlobalTrainingToUser(ctx context.Context, cmd AssignGlobalTrainingCmd) (*Training, error)
}




type ExerciseRepository interface {
	// Упражнения
	GetExercisesWithTags(ctx context.Context) ([]*Exercise, error)
	GetExerciseByID(ctx context.Context, id int64) (*Exercise, error)
	GetExercisesByTag(ctx context.Context, tagID int64) ([]*Exercise, error)
	SearchExercises(ctx context.Context, filter ExerciseFilter) ([]*Exercise, error)
	
	// Теги
	GetAllTags(ctx context.Context) ([]*Tag, error)
	GetTagByID(ctx context.Context, id int64) (*Tag, error)
	
	// Связи упражнений с тегами
	GetExerciseTags(ctx context.Context, exerciseID int64) ([]*Tag, error)
}