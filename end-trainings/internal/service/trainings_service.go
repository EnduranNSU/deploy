package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/EnduranNSU/trainings/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidTrainingID = errors.New("invalid training id")
	ErrTrainingNotFound  = errors.New("training not found")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrTrainingNotActive = errors.New("training is not active")
	ErrInvalidGlobalTrainingID = errors.New("invalid global training id")
    ErrGlobalTrainingNotFound  = errors.New("global training not found")
)

func NewTrainingService(repo domain.TrainingRepository) domain.TrainingService {
	return &trainingService{repo: repo}
}

type trainingService struct {
	repo domain.TrainingRepository
}

func (s *trainingService) GetUserTrainingStats(ctx context.Context, userID uuid.UUID) (*domain.TrainingStats, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	return s.repo.GetUserTrainingStats(ctx, userID)
}

func (s *trainingService) GetTrainingsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Training, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	return s.repo.GetTrainingsByUser(ctx, userID)
}

func (s *trainingService) GetTrainingWithExercises(ctx context.Context, trainingID int64) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	return s.repo.GetTrainingWithExercises(ctx, trainingID)
}

func (s *trainingService) CreateTraining(ctx context.Context, cmd domain.CreateTrainingCmd) (*domain.Training, error) {
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	if cmd.PlannedDate.IsZero() {
		return nil, errors.New("planned date is required")
	}

	training := &domain.Training{
		UserID:            cmd.UserID,
		IsDone:            cmd.IsDone,
		PlannedDate:       cmd.PlannedDate,
		ActualDate:        cmd.ActualDate,
		StartedAt:         cmd.StartedAt,
		FinishedAt:        cmd.FinishedAt,
		TotalDuration:     cmd.TotalDuration,
		TotalRestTime:     cmd.TotalRestTime,
		TotalExerciseTime: cmd.TotalExerciseTime,
		Rating:            cmd.Rating,
	}

	return s.repo.CreateTraining(ctx, training)
}

func (s *trainingService) UpdateTraining(ctx context.Context, cmd domain.UpdateTrainingCmd) (*domain.Training, error) {
	if cmd.ID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	// Проверяем существование тренировки
	existing, err := s.repo.GetTrainingWithExercises(ctx, cmd.ID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	// Обновляем только переданные поля
	if cmd.IsDone != nil {
		existing.IsDone = *cmd.IsDone
	}
	if !cmd.PlannedDate.IsZero() {
		existing.PlannedDate = cmd.PlannedDate
	}
	if cmd.ActualDate != nil {
		existing.ActualDate = cmd.ActualDate
	}
	if cmd.StartedAt != nil {
		existing.StartedAt = cmd.StartedAt
	}
	if cmd.FinishedAt != nil {
		existing.FinishedAt = cmd.FinishedAt
	}
	if cmd.TotalDuration != nil {
		existing.TotalDuration = cmd.TotalDuration
	}
	if cmd.TotalRestTime != nil {
		existing.TotalRestTime = cmd.TotalRestTime
	}
	if cmd.TotalExerciseTime != nil {
		existing.TotalExerciseTime = cmd.TotalExerciseTime
	}
	if cmd.Rating != nil {
		existing.Rating = cmd.Rating
	}

	return s.repo.UpdateTraining(ctx, existing)
}

func (s *trainingService) DeleteTraining(ctx context.Context, trainingID int64) error {
	if trainingID <= 0 {
		return ErrInvalidTrainingID
	}

	return s.repo.DeleteTrainingAndExercises(ctx, trainingID)
}

func (s *trainingService) AddExerciseToTraining(ctx context.Context, cmd domain.AddExerciseToTrainingCmd) (*domain.TrainedExercise, error) {
	if cmd.TrainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}
	if cmd.ExerciseID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	// Проверяем существование тренировки
	_, err := s.repo.GetTrainingWithExercises(ctx, cmd.TrainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	exercise := &domain.TrainedExercise{
		TrainingID: cmd.TrainingID,
		ExerciseID: cmd.ExerciseID,
		Weight:     cmd.Weight,
		Approaches: cmd.Approaches,
		Reps:       cmd.Reps,
		Time:       cmd.Time,
		Doing:      cmd.Doing,
		Rest:       cmd.Rest,
		Notes:      cmd.Notes,
	}

	return s.repo.AddExerciseToTraining(ctx, exercise)
}

func (s *trainingService) UpdateTrainedExercise(ctx context.Context, cmd domain.UpdateTrainedExerciseCmd) (*domain.TrainedExercise, error) {
	if cmd.ID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	exercise := &domain.TrainedExercise{
		ID:         cmd.ID,
		Weight:     cmd.Weight,
		Approaches: cmd.Approaches,
		Reps:       cmd.Reps,
		Time:       cmd.Time,
		Doing:      cmd.Doing,
		Rest:       cmd.Rest,
		Notes:      cmd.Notes,
	}

	return s.repo.UpdateTrainedExercise(ctx, exercise)
}

func (s *trainingService) RemoveExerciseFromTraining(ctx context.Context, trainingID, exerciseID int64) error {
	if trainingID <= 0 {
		return ErrInvalidTrainingID
	}
	if exerciseID <= 0 {
		return ErrInvalidExerciseID
	}

	return s.repo.DeleteExerciseFromTraining(ctx, exerciseID, trainingID)
}

func (s *trainingService) CompleteTraining(ctx context.Context, trainingID int64, rating *int32) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	training.IsDone = true
	now := time.Now().UTC()
	training.ActualDate = &now
	training.FinishedAt = &now
	training.Rating = rating

	// Если не было начато, устанавливаем время начала как текущее минус предполагаемая длительность
	if training.StartedAt == nil {
		startedAt := now.Add(-*training.TotalDuration)
		training.StartedAt = &startedAt
	}

	return s.repo.UpdateTraining(ctx, training)
}


func (s *trainingService) UpdateExerciseTime(ctx context.Context, exerciseID int64, weight *decimal.Decimal, approaches *int32, reps *int32, time *time.Duration, doing *time.Duration, rest *time.Duration) (*domain.TrainedExercise, error) {
	if exerciseID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	// Создаем объект упражнения с обновленными временными параметрами
	exercise := &domain.TrainedExercise{
		ID:         exerciseID,
		Weight:     weight,
		Approaches: approaches,
		Reps:       reps,
		Time:       time,
		Doing:      doing,
		Rest:       rest,
	}

	return s.repo.UpdateExerciseTime(ctx, exercise)
}

func (s *trainingService) UpdateTrainingTimers(ctx context.Context, trainingID int64, totalDuration *time.Duration, totalRestTime *time.Duration, totalExerciseTime *time.Duration) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	// Получаем существующую тренировку
	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	// Обновляем таймеры
	if totalDuration != nil {
		training.TotalDuration = totalDuration
	}
	if totalRestTime != nil {
		training.TotalRestTime = totalRestTime
	}
	if totalExerciseTime != nil {
		training.TotalExerciseTime = totalExerciseTime
	}

	return s.repo.UpdateTrainingTimers(ctx, training)
}

func (s *trainingService) CalculateTrainingTotalTime(ctx context.Context, trainingID int64) (*domain.TrainingTime, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	// Проверяем существование тренировки
	_, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	return s.repo.CalculateTrainingTotalTime(ctx, trainingID)
}

func (s *trainingService) GetCurrentTraining(ctx context.Context, userID uuid.UUID) (*domain.Training, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	return s.repo.GetCurrentTraining(ctx, userID)
}

func (s *trainingService) GetTodaysTraining(ctx context.Context, userID uuid.UUID) ([]*domain.Training, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	trainings, err := s.repo.GetTodaysTraining(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Если нет тренировок на сегодня, можно создать новую или вернуть пустой список
	// В зависимости от бизнес-логики:
	// 1. Создать новую тренировку на сегодня
	// 2. Вернуть тренировки, запланированные на сегодня
	// 3. Вернуть активную тренировку (если есть)

	return trainings, nil
}

func (s *trainingService) GetGlobalTrainings(ctx context.Context) ([]*domain.GlobalTraining, error) {
	return s.repo.GetGlobalTrainings(ctx)
}

func (s *trainingService) GetGlobalTrainingByLevel(ctx context.Context, level string) ([]*domain.GlobalTraining, error) {
	if level == "" {
		return nil, errors.New("level is required")
	}
	return s.repo.GetGlobalTrainingByLevel(ctx, level)
}

func (s *trainingService) GetGlobalTrainingById(ctx context.Context, trainingID int64) (*domain.GlobalTraining, error) {
	globalTraining, err := s.repo.GetGlobalTrainingById(ctx, trainingID)
	if err != nil {
		return nil, err
	}
	return globalTraining, nil
}

func (s *trainingService) MarkTrainingAsDone(ctx context.Context, trainingID int64, userID uuid.UUID) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	// Проверяем, что тренировка принадлежит пользователю
	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	if training.UserID != userID {
		return nil, errors.New("training does not belong to user")
	}

	// Проверяем, что тренировка не завершена
	if training.IsDone {
		return training, nil // Уже завершена
	}

	// Завершаем тренировку
	return s.repo.MarkTrainingAsDone(ctx, trainingID, userID)
}

func (s *trainingService) GetTrainingStats(ctx context.Context, trainingID int64) (*domain.TrainingStats, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	return s.repo.GetTrainingStats(ctx, trainingID)
}

func (s *trainingService) StartTraining(ctx context.Context, trainingID int64, userID uuid.UUID) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	// Проверяем, что тренировка принадлежит пользователю
	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	if training.UserID != userID {
		return nil, errors.New("training does not belong to user")
	}

	// Проверяем, что тренировка еще не начата
	if training.StartedAt != nil {
		return training, nil // Уже начата
	}

	// Начинаем тренировку
	return s.repo.StartTraining(ctx, trainingID, userID)
}

// Дополнительные методы для управления временем тренировки

func (s *trainingService) UpdateExerciseRestTime(ctx context.Context, exerciseID int64, restTime time.Duration) (*domain.TrainedExercise, error) {
	if exerciseID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	// Получаем упражнение
	// В реальной реализации нужно получить упражнение из репозитория
	// Для примера создаем новый объект
	exercise := &domain.TrainedExercise{
		ID:   exerciseID,
		Rest: &restTime,
	}

	return s.repo.UpdateExerciseTime(ctx, exercise)
}

func (s *trainingService) UpdateExerciseDoingTime(ctx context.Context, exerciseID int64, doingTime time.Duration) (*domain.TrainedExercise, error) {
	if exerciseID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	exercise := &domain.TrainedExercise{
		ID:    exerciseID,
		Doing: &doingTime,
	}

	return s.repo.UpdateExerciseTime(ctx, exercise)
}

func (s *trainingService) PauseTraining(ctx context.Context, trainingID int64) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	// Получаем тренировку
	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	// Проверяем, что тренировка начата
	if training.StartedAt == nil {
		return nil, ErrTrainingNotActive
	}

	// В реальной реализации здесь можно:
	// 1. Сохранить время паузы
	// 2. Обновить общее время тренировки
	// 3. Запустить таймер паузы

	return training, nil
}

func (s *trainingService) ResumeTraining(ctx context.Context, trainingID int64) (*domain.Training, error) {
	if trainingID <= 0 {
		return nil, ErrInvalidTrainingID
	}

	// Получаем тренировку
	training, err := s.repo.GetTrainingWithExercises(ctx, trainingID)
	if err != nil {
		return nil, ErrTrainingNotFound
	}

	// Проверяем, что тренировка начата
	if training.StartedAt == nil {
		return nil, ErrTrainingNotActive
	}

	// В реальной реализации здесь можно:
	// 1. Обновить время тренировки с учетом паузы
	// 2. Запустить таймер заново

	return training, nil
}

// Реализация метода в trainingService структуре
func (s *trainingService) AssignGlobalTraining(ctx context.Context, cmd domain.AssignGlobalTrainingCmd) (*domain.Training, error) {
    // Валидация входных данных
    if cmd.UserID == uuid.Nil {
        return nil, ErrInvalidUserID
    }
    if cmd.GlobalTrainingID <= 0 {
        return nil, ErrInvalidGlobalTrainingID
    }

    // Вызываем метод репозитория для назначения глобальной тренировки
    training, err := s.repo.AssignGlobalTrainingToUser(ctx, cmd)
    if err != nil {
        // Обрабатываем возможные ошибки
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrGlobalTrainingNotFound
        }
        return nil, err
    }

    // Можно добавить дополнительную бизнес-логику:
    // 1. Отправка уведомления пользователю
    // 2. Создание напоминаний
    // 3. Логирование события
    // 4. Обновление статистики пользователя

    return training, nil
}