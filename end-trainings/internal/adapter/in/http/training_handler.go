package httpin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/EnduranNSU/trainings/internal/adapter/in/http/dto"
	svctraining "github.com/EnduranNSU/trainings/internal/domain"
)

type TrainingHandler struct {
	svc svctraining.TrainingService
}

func NewTrainingHandler(svc svctraining.TrainingService) *TrainingHandler {
	return &TrainingHandler{svc: svc}
}

// GetTrainingsByUser получает все тренировки пользователя
// @Summary      Получить тренировки пользователя
// @Description  Возвращает все тренировки указанного пользователя
// @Tags         trainings
// @Produce      json
// @Param        user_id query string true "User ID"
// @Success      200  {array}   dto.UserTrainingsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings [get]
func (h *TrainingHandler) GetTrainingsByUser(c *gin.Context) {
	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	trainings, err := h.svc.GetTrainingsByUser(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get trainings"})
		return
	}

	if len(trainings) == 0 {
		c.JSON(http.StatusOK, []dto.UserTrainingsResponse{})
		return
	}

	resp := make([]dto.UserTrainingsResponse, 0, len(trainings))
	for _, training := range trainings {
		resp = append(resp, h.userTrainingToResponse(training))
	}

	c.JSON(http.StatusOK, resp)
}

// GetTrainingWithExercises получает тренировку с упражнениями
// @Summary      Получить тренировку с упражнениями
// @Description  Возвращает информацию о тренировке со списком упражнений
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id} [get]
func (h *TrainingHandler) GetTrainingWithExercises(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	training, err := h.svc.GetTrainingWithExercises(c.Request.Context(), trainingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "training not found"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// CreateTraining создает новую тренировку
// @Summary      Создать тренировку
// @Description  Создает новую тренировку для пользователя
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTrainingRequest true "Данные тренировки"
// @Success      201  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings [post]
func (h *TrainingHandler) CreateTraining(c *gin.Context) {
	var req dto.CreateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	plannedDate, err := time.Parse(time.RFC3339, req.PlannedDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid planned_date time format"})
		return
	}

	var actualDate, startedAt, finishedAt *time.Time
	var totalDuration, totalRestTime, totalExerciseTime *time.Duration

	// Parse optional time fields
	if req.ActualDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ActualDate)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid actual_date time format"})
			return
		}
		actualDate = &t
	}
	if req.StartedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid started_at time format"})
			return
		}
		startedAt = &t
	}
	if req.FinishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.FinishedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid finished_at time format"})
			return
		}
		finishedAt = &t
	}

	// Parse optional duration fields
	if req.TotalDuration != nil {
		duration, err := time.ParseDuration(*req.TotalDuration)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_duration format"})
			return
		}
		totalDuration = &duration
	}
	if req.TotalRestTime != nil {
		duration, err := time.ParseDuration(*req.TotalRestTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_rest_time format"})
			return
		}
		totalRestTime = &duration
	}
	if req.TotalExerciseTime != nil {
		duration, err := time.ParseDuration(*req.TotalExerciseTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_exercise_time format"})
			return
		}
		totalExerciseTime = &duration
	}

	cmd := svctraining.CreateTrainingCmd{
		UserID:            uid,
		Title:             req.Title,
		IsDone:            req.IsDone,
		PlannedDate:       plannedDate,
		ActualDate:        actualDate,
		StartedAt:         startedAt,
		FinishedAt:        finishedAt,
		TotalDuration:     totalDuration,
		TotalRestTime:     totalRestTime,
		TotalExerciseTime: totalExerciseTime,
		Rating:            req.Rating,
	}

	training, err := h.svc.CreateTraining(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create training"})
		return
	}

	c.JSON(http.StatusCreated, h.trainingToResponse(training))
}

// UpdateTraining обновляет тренировку
// @Summary      Обновить тренировку
// @Description  Обновляет информацию о тренировке
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Param        request body dto.UpdateTrainingRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id} [put]
func (h *TrainingHandler) UpdateTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	var req dto.UpdateTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	plannedDate, err := time.Parse(time.RFC3339, req.PlannedDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid planned_date time format"})
		return
	}

	var actualDate, startedAt, finishedAt *time.Time
	var totalDuration, totalRestTime, totalExerciseTime *time.Duration

	if req.ActualDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ActualDate)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid actual_date time format"})
			return
		}
		actualDate = &t
	}
	if req.StartedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid started_at time format"})
			return
		}
		startedAt = &t
	}
	if req.FinishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.FinishedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid finished_at time format"})
			return
		}
		finishedAt = &t
	}

	if req.TotalDuration != nil {
		duration, err := time.ParseDuration(*req.TotalDuration)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_duration format"})
			return
		}
		totalDuration = &duration
	}
	if req.TotalRestTime != nil {
		duration, err := time.ParseDuration(*req.TotalRestTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_rest_time format"})
			return
		}
		totalRestTime = &duration
	}
	if req.TotalExerciseTime != nil {
		duration, err := time.ParseDuration(*req.TotalExerciseTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_exercise_time format"})
			return
		}
		totalExerciseTime = &duration
	}

	cmd := svctraining.UpdateTrainingCmd{
		ID:                trainingID,
		Title:             req.Title,
		IsDone:            req.IsDone,
		PlannedDate:       plannedDate,
		ActualDate:        actualDate,
		StartedAt:         startedAt,
		FinishedAt:        finishedAt,
		TotalDuration:     totalDuration,
		TotalRestTime:     totalRestTime,
		TotalExerciseTime: totalExerciseTime,
		Rating:            req.Rating,
	}

	training, err := h.svc.UpdateTraining(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update training"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// DeleteTraining удаляет тренировку
// @Summary      Удалить тренировку
// @Description  Удаляет тренировку по ID
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id} [delete]
func (h *TrainingHandler) DeleteTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	err = h.svc.DeleteTraining(c.Request.Context(), trainingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete training"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddExerciseToTraining добавляет упражнение к тренировке
// @Summary      Добавить упражнение к тренировке
// @Description  Добавляет упражнение к существующей тренировке
// @Tags         training-exercises
// @Accept       json
// @Produce      json
// @Param        request body dto.AddExerciseToTrainingRequest true "Данные упражнения"
// @Success      201  {object}  dto.TrainedExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises [post]
func (h *TrainingHandler) AddExerciseToTraining(c *gin.Context) {
	var req dto.AddExerciseToTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	var weight *decimal.Decimal
	if req.Weight != nil {
		w := decimal.NewFromFloat(*req.Weight)
		weight = &w
	}

	var timeVal, doing, rest *time.Duration
	if req.Time != nil {
		duration, err := time.ParseDuration(*req.Time)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid time format, use duration format like '1h30m'"})
			return
		}
		timeVal = &duration
	}
	if req.Doing != nil {
		duration, err := time.ParseDuration(*req.Doing)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid doing format, use duration format like '1h30m'"})
			return
		}
		doing = &duration
	}
	if req.Rest != nil {
		duration, err := time.ParseDuration(*req.Rest)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rest format, use duration format like '1h30m'"})
			return
		}
		rest = &duration
	}

	// Convert approaches and reps from *int64 to *int32
	var approaches, reps *int32
	if req.Approaches != nil {
		a := int32(*req.Approaches)
		approaches = &a
	}
	if req.Reps != nil {
		r := int32(*req.Reps)
		reps = &r
	}

	cmd := svctraining.AddExerciseToTrainingCmd{
		TrainingID: req.TrainingID,
		ExerciseID: req.ExerciseID,
		Weight:     weight,
		Approaches: approaches,
		Reps:       reps,
		Time:       timeVal,
		Doing:      doing,
		Rest:       rest,
		Notes:      req.Notes,
	}

	exercise, err := h.svc.AddExerciseToTraining(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to add exercise to training"})
		return
	}

	c.JSON(http.StatusCreated, h.trainedExerciseToResponse(exercise))
}

// UpdateTrainedExercise обновляет выполненное упражнение
// @Summary      Обновить выполненное упражнение
// @Description  Обновляет информацию о выполненном упражнении в тренировке
// @Tags         training-exercises
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Trained Exercise ID"
// @Param        request body dto.UpdateTrainedExerciseRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainedExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises/{id} [put]
func (h *TrainingHandler) UpdateTrainedExercise(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	var req dto.UpdateTrainedExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	var weight *decimal.Decimal
	if req.Weight != nil {
		w := decimal.NewFromFloat(*req.Weight)
		weight = &w
	}

	var timeVal, doing, rest *time.Duration
	if req.Time != nil {
		duration, err := time.ParseDuration(*req.Time)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid time format, use duration format like '1h30m'"})
			return
		}
		timeVal = &duration
	}
	if req.Doing != nil {
		duration, err := time.ParseDuration(*req.Doing)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid doing format, use duration format like '1h30m'"})
			return
		}
		doing = &duration
	}
	if req.Rest != nil {
		duration, err := time.ParseDuration(*req.Rest)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rest format, use duration format like '1h30m'"})
			return
		}
		rest = &duration
	}

	// Convert approaches and reps from *int64 to *int32
	var approaches, reps *int32
	if req.Approaches != nil {
		a := int32(*req.Approaches)
		approaches = &a
	}
	if req.Reps != nil {
		r := int32(*req.Reps)
		reps = &r
	}

	cmd := svctraining.UpdateTrainedExerciseCmd{
		ID:         exerciseID,
		Weight:     weight,
		Approaches: approaches,
		Reps:       reps,
		Time:       timeVal,
		Doing:      doing,
		Rest:       rest,
		Notes:      req.Notes,
	}

	exercise, err := h.svc.UpdateTrainedExercise(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update exercise"})
		return
	}

	c.JSON(http.StatusOK, h.trainedExerciseToResponse(exercise))
}

// RemoveExerciseFromTraining удаляет упражнение из тренировки
// @Summary      Удалить упражнение из тренировки
// @Description  Удаляет упражнение из тренировки
// @Tags         training-exercises
// @Produce      json
// @Param        training_id query int64 true "Training ID"
// @Param        exercise_id query int64 true "Exercise ID"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises [delete]
func (h *TrainingHandler) RemoveExerciseFromTraining(c *gin.Context) {
	trainingID, err := parseInt64Query(c, "training_id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training_id"})
		return
	}

	exerciseID, err := parseInt64Query(c, "exercise_id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise_id"})
		return
	}

	err = h.svc.RemoveExerciseFromTraining(c.Request.Context(), trainingID, exerciseID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to remove exercise from training"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserTrainingStats получает статистику тренировок пользователя
// @Summary      Получить статистику тренировок
// @Description  Возвращает статистику тренировок пользователя
// @Tags         trainings
// @Produce      json
// @Success      200  {object}  dto.TrainingStatsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/stats [get]
func (h *TrainingHandler) GetUserTrainingStats(c *gin.Context) {
	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	stats, err := h.svc.GetUserTrainingStats(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get training stats"})
		return
	}

	resp := dto.TrainingStatsResponse{
		TotalTrainings:     stats.TotalTrainings,
		CompletedTrainings: stats.CompletedTrainings,
		AverageRating:      stats.AverageRating,
		TotalDuration:      stats.TotalDuration.String(),
	}

	c.JSON(http.StatusOK, resp)
}

// CompleteTraining завершает тренировку
// @Summary      Завершить тренировку
// @Description  Отмечает тренировку как завершенную и устанавливает рейтинг
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Param        request body dto.CompleteTrainingRequest true "Данные для завершения"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/complete [patch]
func (h *TrainingHandler) CompleteTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	var req dto.CompleteTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	training, err := h.svc.CompleteTraining(c.Request.Context(), trainingID, req.Rating)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to complete training"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// Вспомогательные методы
func (h *TrainingHandler) userTrainingToResponse(training *svctraining.Training) dto.UserTrainingsResponse {
	var actualDate, startedAt, finishedAt *string
	if training.ActualDate != nil {
		s := training.ActualDate.Format(time.RFC3339)
		actualDate = &s
	}
	if training.StartedAt != nil {
		s := training.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if training.FinishedAt != nil {
		s := training.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	var totalDuration, totalRestTime, totalExerciseTime *string
	if training.TotalDuration != nil {
		s := formatDuration(*training.TotalDuration)
		totalDuration = &s
	}
	if training.TotalRestTime != nil {
		s := formatDuration(*training.TotalRestTime)
		totalRestTime = &s
	}
	if training.TotalExerciseTime != nil {
		s := formatDuration(*training.TotalExerciseTime)
		totalExerciseTime = &s
	}

	return dto.UserTrainingsResponse{
		ID:                training.ID,
		Title:             training.Title,
		UserID:            training.UserID.String(),
		IsDone:            training.IsDone,
		PlannedDate:       training.PlannedDate.Format(time.RFC3339),
		ActualDate:        actualDate,
		StartedAt:         startedAt,
		FinishedAt:        finishedAt,
		TotalDuration:     totalDuration,
		TotalRestTime:     totalRestTime,
		TotalExerciseTime: totalExerciseTime,
		Rating:            training.Rating,
	}
}

func (h *TrainingHandler) trainingToResponse(training *svctraining.Training) dto.TrainingResponse {
	var actualDate, startedAt, finishedAt *string
	if training.ActualDate != nil {
		s := training.ActualDate.Format(time.RFC3339)
		actualDate = &s
	}
	if training.StartedAt != nil {
		s := training.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if training.FinishedAt != nil {
		s := training.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	var totalDuration, totalRestTime, totalExerciseTime *string
	if training.TotalDuration != nil {
		s := formatDuration(*training.TotalDuration)
		totalDuration = &s
	}
	if training.TotalRestTime != nil {
		s := formatDuration(*training.TotalRestTime)
		totalRestTime = &s
	}
	if training.TotalExerciseTime != nil {
		s := formatDuration(*training.TotalExerciseTime)
		totalExerciseTime = &s
	}

	var exercises []dto.TrainedExerciseResponse
	if training.Exercises != nil {
		exercises = make([]dto.TrainedExerciseResponse, 0, len(training.Exercises))
		for _, exercise := range training.Exercises {
			exercises = append(exercises, h.trainedExerciseToResponse(&exercise))
		}
	}

	return dto.TrainingResponse{
		ID:                training.ID,
		Title:             training.Title,
		UserID:            training.UserID.String(),
		IsDone:            training.IsDone,
		PlannedDate:       training.PlannedDate.Format(time.RFC3339),
		ActualDate:        actualDate,
		StartedAt:         startedAt,
		FinishedAt:        finishedAt,
		TotalDuration:     totalDuration,
		TotalRestTime:     totalRestTime,
		TotalExerciseTime: totalExerciseTime,
		Rating:            training.Rating,
		Exercises:         exercises,
	}
}

func (h *TrainingHandler) trainedExerciseToResponse(exercise *svctraining.TrainedExercise) dto.TrainedExerciseResponse {
	var weight *float64
	if exercise.Weight != nil {
		f, _ := exercise.Weight.Float64()
		weight = &f
	}

	var timeStr, doingStr, restStr *string
	if exercise.Time != nil {
		s := formatDuration(*exercise.Time)
		timeStr = &s
	}
	if exercise.Doing != nil {
		s := formatDuration(*exercise.Doing)
		doingStr = &s
	}
	if exercise.Rest != nil {
		s := formatDuration(*exercise.Rest)
		restStr = &s
	}

	return dto.TrainedExerciseResponse{
		ID:         exercise.ID,
		TrainingID: exercise.TrainingID,
		ExerciseID: exercise.ExerciseID,
		Weight:     weight,
		Approaches: exercise.Approaches,
		Reps:       exercise.Reps,
		Time:       timeStr,
		Doing:      doingStr,
		Rest:       restStr,
		Notes:      exercise.Notes,
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// UpdateExerciseTime обновляет временные параметры упражнения
// @Summary      Обновить временные параметры упражнения
// @Description  Обновляет вес, подходы, повторения и временные параметры упражнения
// @Tags         training-exercises
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Exercise ID"
// @Param        request body dto.UpdateExerciseTimeRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainedExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises/{id}/time [patch]
func (h *TrainingHandler) UpdateExerciseTime(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	var req dto.UpdateExerciseTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	// Парсим параметры
	var weight *decimal.Decimal
	if req.Weight != nil {
		w := decimal.NewFromFloat(*req.Weight)
		weight = &w
	}

	var timeVal, doing, rest *time.Duration
	if req.Time != nil {
		duration, err := time.ParseDuration(*req.Time)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid time format, use duration format like '1h30m'"})
			return
		}
		timeVal = &duration
	}
	if req.Doing != nil {
		duration, err := time.ParseDuration(*req.Doing)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid doing format, use duration format like '1h30m'"})
			return
		}
		doing = &duration
	}
	if req.Rest != nil {
		duration, err := time.ParseDuration(*req.Rest)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rest format, use duration format like '1h30m'"})
			return
		}
		rest = &duration
	}

	var approaches, reps *int32
	if req.Approaches != nil {
		a := int32(*req.Approaches)
		approaches = &a
	}
	if req.Reps != nil {
		r := int32(*req.Reps)
		reps = &r
	}

	exercise, err := h.svc.UpdateExerciseTime(c.Request.Context(), exerciseID, weight, approaches, reps, timeVal, doing, rest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update exercise time"})
		return
	}

	c.JSON(http.StatusOK, h.trainedExerciseToResponse(exercise))
}

// UpdateTrainingTimers обновляет таймеры тренировки
// @Summary      Обновить таймеры тренировки
// @Description  Обновляет общее время, время отдыха и время выполнения упражнений
// @Tags         trainings
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Param        request body dto.UpdateTrainingTimersRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/timers [patch]
func (h *TrainingHandler) UpdateTrainingTimers(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	var req dto.UpdateTrainingTimersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	// Парсим duration поля
	var totalDuration, totalRestTime, totalExerciseTime *time.Duration
	if req.TotalDuration != nil {
		duration, err := time.ParseDuration(*req.TotalDuration)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_duration format"})
			return
		}
		totalDuration = &duration
	}
	if req.TotalRestTime != nil {
		duration, err := time.ParseDuration(*req.TotalRestTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_rest_time format"})
			return
		}
		totalRestTime = &duration
	}
	if req.TotalExerciseTime != nil {
		duration, err := time.ParseDuration(*req.TotalExerciseTime)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid total_exercise_time format"})
			return
		}
		totalExerciseTime = &duration
	}

	training, err := h.svc.UpdateTrainingTimers(c.Request.Context(), trainingID, totalDuration, totalRestTime, totalExerciseTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update training timers"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// CalculateTrainingTotalTime вычисляет общее время тренировки
// @Summary      Вычислить общее время тренировки
// @Description  Вычисляет общее время тренировки, время отдыха и время выполнения упражнений
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingTimeResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/calculate-time [get]
func (h *TrainingHandler) CalculateTrainingTotalTime(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	trainingTime, err := h.svc.CalculateTrainingTotalTime(c.Request.Context(), trainingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to calculate training time"})
		return
	}

	resp := dto.TrainingTimeResponse{
		TotalDuration:     formatDuration(time.Duration(trainingTime.TotalSeconds)),
		TotalRestTime:     formatDuration(time.Duration(trainingTime.TotalRestSeconds)),
		TotalExerciseTime: formatDuration(time.Duration(trainingTime.TotalExerciseSeconds)),
	}

	c.JSON(http.StatusOK, resp)
}

// GetCurrentTraining получает текущую активную тренировку пользователя
// @Summary      Получить текущую тренировку
// @Description  Возвращает активную тренировку пользователя (если есть)
// @Tags         trainings
// @Produce      json
// @Success      200  {object}  dto.TrainingResponse
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/current [get]
func (h *TrainingHandler) GetCurrentTraining(c *gin.Context) {
	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	training, err := h.svc.GetCurrentTraining(c.Request.Context(), uid)
	if err != nil {
		// Если тренировка не найдена, возвращаем 204
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// GetTodaysTraining получает тренировки на сегодня
// @Summary      Получить тренировки на сегодня
// @Description  Возвращает тренировки пользователя, запланированные на сегодня
// @Tags         trainings
// @Produce      json
// @Success      200  {array}   dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/today [get]
func (h *TrainingHandler) GetTodaysTraining(c *gin.Context) {
	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	trainings, err := h.svc.GetTodaysTraining(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get today's training"})
		return
	}

	if len(trainings) == 0 {
		c.JSON(http.StatusOK, []dto.TrainingResponse{})
		return
	}

	resp := make([]dto.TrainingResponse, 0, len(trainings))
	for _, training := range trainings {
		resp = append(resp, h.trainingToResponse(training))
	}

	c.JSON(http.StatusOK, resp)
}

// GetGlobalTrainings получает все глобальные тренировки
// @Summary      Получить глобальные тренировки
// @Description  Возвращает список всех глобальных тренировок
// @Tags         global-trainings
// @Produce      json
// @Success      200  {array}   dto.GlobalTrainingWithTagsResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /global-trainings [get]
func (h *TrainingHandler) GetGlobalTrainings(c *gin.Context) {
	globalTrainings, err := h.svc.GetGlobalTrainings(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get global trainings"})
		return
	}

	if len(globalTrainings) == 0 {
		c.JSON(http.StatusOK, []dto.GlobalTrainingWithTagsResponse{})
		return
	}

	resp := make([]dto.GlobalTrainingWithTagsResponse, 0, len(globalTrainings))
	for _, gt := range globalTrainings {
		resp = append(resp, h.globalTrainingWithTagsToResponse(gt))
	}

	c.JSON(http.StatusOK, resp)
}

// GetGlobalTrainingByLevel получает глобальную тренировку по уровню
// @Summary      Получить глобальную тренировку по уровню
// @Description  Возвращает глобальную тренировку по указанному уровню
// @Tags         global-trainings
// @Produce      json
// @Param        level path string true "Уровень тренировки"
// @Success      200  {object}  dto.GlobalTrainingWithTagsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /global-trainings/level/{level} [get]
func (h *TrainingHandler) GetGlobalTrainingByLevel(c *gin.Context) {
	level := c.Param("level")
	if level == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "level is required"})
		return
	}

	globalTrainings, err := h.svc.GetGlobalTrainingByLevel(c.Request.Context(), level)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "global training not found"})
		return
	}

	resp := make([]dto.GlobalTrainingWithTagsResponse, 0, len(globalTrainings))
	for _, gt := range globalTrainings {
		resp = append(resp, h.globalTrainingWithTagsToResponse(gt))
	}

	c.JSON(http.StatusOK, resp)
}

// GetGlobalTrainingById получает глобальную тренировку по id
// @Summary      Получить глобальную тренировку по id
// @Description  Возвращает глобальную тренировку с упражнениями и их тегами
// @Tags         global-trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.GlobalTrainingWithTagsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /global-trainings/{id} [get]
func (h *TrainingHandler) GetGlobalTrainingById(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	globalTraining, err := h.svc.GetGlobalTrainingById(c.Request.Context(), trainingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "global training not found"})
		return
	}

	c.JSON(http.StatusOK, h.globalTrainingWithTagsToResponse(globalTraining))
}

// MarkTrainingAsDone отмечает тренировку как завершенную
// @Summary      Отметить тренировку как завершенную
// @Description  Отмечает тренировку как завершенную (проверяет принадлежность пользователю)
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/mark-done [patch]
func (h *TrainingHandler) MarkTrainingAsDone(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	training, err := h.svc.MarkTrainingAsDone(c.Request.Context(), trainingID, uid)
	if err != nil {
		if err.Error() == "training does not belong to user" {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to mark training as done"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// GetTrainingStats получает статистику по тренировке
// @Summary      Получить статистику тренировки
// @Description  Возвращает статистику по конкретной тренировке
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingStatsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/stats [get]
func (h *TrainingHandler) GetTrainingStats(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	stats, err := h.svc.GetTrainingStats(c.Request.Context(), trainingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get training stats"})
		return
	}

	resp := dto.TrainingStatsResponse{
		TotalTrainings:     stats.TotalTrainings,
		CompletedTrainings: stats.CompletedTrainings,
		AverageRating:      stats.AverageRating,
		TotalDuration:      stats.TotalDuration.String(),
	}

	c.JSON(http.StatusOK, resp)
}

// StartTraining начинает тренировку
// @Summary      Начать тренировку
// @Description  Начинает тренировку (устанавливает время начала)
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/start [patch]
func (h *TrainingHandler) StartTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	training, err := h.svc.StartTraining(c.Request.Context(), trainingID, uid)
	if err != nil {
		if err.Error() == "training does not belong to user" {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to start training"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// UpdateExerciseRestTime обновляет время отдыха упражнения
// @Summary      Обновить время отдыха упражнения
// @Description  Обновляет время отдыха для конкретного упражнения
// @Tags         training-exercises
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Exercise ID"
// @Param        request body dto.UpdateExerciseRestTimeRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainedExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises/{id}/rest-time [patch]
func (h *TrainingHandler) UpdateExerciseRestTime(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	var req dto.UpdateExerciseRestTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	restTime, err := time.ParseDuration(req.RestTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rest_time format, use duration format like '1h30m'"})
		return
	}

	exercise, err := h.svc.UpdateExerciseRestTime(c.Request.Context(), exerciseID, restTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update exercise rest time"})
		return
	}

	c.JSON(http.StatusOK, h.trainedExerciseToResponse(exercise))
}

// UpdateExerciseDoingTime обновляет время выполнения упражнения
// @Summary      Обновить время выполнения упражнения
// @Description  Обновляет время выполнения для конкретного упражнения
// @Tags         training-exercises
// @Accept       json
// @Produce      json
// @Param        id path int64 true "Exercise ID"
// @Param        request body dto.UpdateExerciseDoingTimeRequest true "Данные для обновления"
// @Success      200  {object}  dto.TrainedExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /training-exercises/{id}/doing-time [patch]
func (h *TrainingHandler) UpdateExerciseDoingTime(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	var req dto.UpdateExerciseDoingTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	doingTime, err := time.ParseDuration(req.DoingTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid doing_time format, use duration format like '1h30m'"})
		return
	}

	exercise, err := h.svc.UpdateExerciseDoingTime(c.Request.Context(), exerciseID, doingTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update exercise doing time"})
		return
	}

	c.JSON(http.StatusOK, h.trainedExerciseToResponse(exercise))
}

// PauseTraining приостанавливает тренировку
// @Summary      Приостановить тренировку
// @Description  Приостанавливает активную тренировку
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      409  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/pause [patch]
func (h *TrainingHandler) PauseTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	training, err := h.svc.PauseTraining(c.Request.Context(), trainingID)
	if err != nil {
		if err.Error() == "training is not active" {
			c.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to pause training"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// ResumeTraining возобновляет тренировку
// @Summary      Возобновить тренировку
// @Description  Возобновляет приостановленную тренировку
// @Tags         trainings
// @Produce      json
// @Param        id path int64 true "Training ID"
// @Success      200  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      409  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /trainings/{id}/resume [patch]
func (h *TrainingHandler) ResumeTraining(c *gin.Context) {
	trainingID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid training id"})
		return
	}

	training, err := h.svc.ResumeTraining(c.Request.Context(), trainingID)
	if err != nil {
		if err.Error() == "training is not active" {
			c.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to resume training"})
		return
	}

	c.JSON(http.StatusOK, h.trainingToResponse(training))
}

// AssignGlobalTraining назначает глобальную тренировку пользователю
// @Summary      Назначить глобальную тренировку
// @Description  Назначает глобальную тренировку пользователю на определенную дату
// @Tags         global-trainings
// @Accept       json
// @Produce      json
// @Param        request body dto.AssignGlobalTrainingRequest true "Данные для назначения"
// @Success      201  {object}  dto.TrainingResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /global-trainings/assign [post]
func (h *TrainingHandler) AssignGlobalTraining(c *gin.Context) {
	var req dto.AssignGlobalTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	uid, ok := userIDFromContext(c)
	if !ok {
		return
	}

	plannedDate, err := time.Parse(time.RFC3339, req.PlannedDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid planned_date time format"})
		return
	}

	cmd := svctraining.AssignGlobalTrainingCmd{
		UserID:           uid,
		GlobalTrainingID: req.GlobalTrainingID,
		PlannedDate:      plannedDate,
	}

	training, err := h.svc.AssignGlobalTraining(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to assign global training"})
		return
	}

	c.JSON(http.StatusCreated, h.trainingToResponse(training))
}

func (h *TrainingHandler) globalTrainingWithTagsToResponse(gt *svctraining.GlobalTraining) dto.GlobalTrainingWithTagsResponse {
	var exercises []dto.ExerciseWithTagsResponse
	if gt.Exercises != nil {
		exercises = make([]dto.ExerciseWithTagsResponse, 0, len(gt.Exercises))
		for _, exercise := range gt.Exercises {
			var tags []dto.TagResponse
			if exercise.Tags != nil {
				tags = make([]dto.TagResponse, 0, len(exercise.Tags))
				for _, tag := range exercise.Tags {
					tags = append(tags, dto.TagResponse{
						ID:   tag.ID,
						Type: tag.Type,
					})
				}
			}

			exercises = append(exercises, dto.ExerciseWithTagsResponse{
				ID:          exercise.ID,
				Title:       exercise.Title,
				Description: exercise.Description,
				VideoURL:    &exercise.VideoUrl,
				ImageURL:    &exercise.ImageUrl,
				Tags:        tags,
			})
		}
	}

	return dto.GlobalTrainingWithTagsResponse{
		ID:          gt.ID,
		Title:       gt.Title,
		Description: gt.Description,
		Level:       gt.Level,
		Exercises:   exercises,
	}
}
