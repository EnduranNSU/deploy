package httpin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EnduranNSU/trainings/internal/adapter/in/http/dto"
	svcexercise "github.com/EnduranNSU/trainings/internal/domain"
)

type ExerciseHandler struct {
	svc svcexercise.ExerciseService
}

func NewExerciseHandler(svc svcexercise.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: svc}
}

// GetAllExercises получает все упражнения
// @Summary      Получить все упражнения
// @Description  Возвращает список всех упражнений
// @Tags         exercises
// @Produce      json
// @Success      200  {array}   dto.ExerciseResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /exercises [get]
func (h *ExerciseHandler) GetAllExercises(c *gin.Context) {
	exercises, err := h.svc.GetAllExercises(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get exercises"})
		return
	}

	resp := make([]dto.ExerciseResponse, 0, len(exercises))
	for _, exercise := range exercises {
		resp = append(resp, h.exerciseToResponse(exercise))
	}

	c.JSON(http.StatusOK, resp)
}

// GetExerciseByID получает упражнение по ID
// @Summary      Получить упражнение по ID
// @Description  Возвращает информацию об упражнении по его ID
// @Tags         exercises
// @Produce      json
// @Param        id path int64 true "Exercise ID"
// @Success      200  {object}  dto.ExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /exercises/{id} [get]
func (h *ExerciseHandler) GetExerciseByID(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	exercise, err := h.svc.GetExerciseByID(c.Request.Context(), exerciseID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "exercise not found"})
		return
	}

	c.JSON(http.StatusOK, h.exerciseToResponse(exercise))
}

// SearchExercises ищет упражнения
// @Summary      Поиск упражнений
// @Description  Возвращает список упражнений, соответствующих поисковому запросу и фильтрам
// @Tags         exercises
// @Produce      json
// @Param        query query string true "Поисковый запрос - ищет по описанию"
// @Param        tag_id query int64 false "ID тега для фильтрации"
// @Success      200  {array}   dto.ExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /exercises/search [get]
func (h *ExerciseHandler) SearchExercises(c *gin.Context) {
	var req dto.SearchExercisesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid query parameters"})
		return
	}

	if req.Query == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "query parameter is required"})
		return
	}

	exercises, err := h.svc.SearchExercises(c.Request.Context(), req.Query, req.TagID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to search exercises"})
		return
	}

	resp := make([]dto.ExerciseResponse, 0, len(exercises))
	for _, exercise := range exercises {
		resp = append(resp, h.exerciseToResponse(exercise))
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllTags получает все теги
// @Summary      Получить все теги
// @Description  Возвращает список всех тегов
// @Tags         tags
// @Produce      json
// @Success      200  {array}   dto.TagResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /tags [get]
func (h *ExerciseHandler) GetAllTags(c *gin.Context) {
	tags, err := h.svc.GetAllTags(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get tags"})
		return
	}

	resp := make([]dto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		resp = append(resp, h.tagToResponse(tag))
	}

	c.JSON(http.StatusOK, resp)
}

// GetTagByID получает тег по ID
// @Summary      Получить тег по ID
// @Description  Возвращает информацию о теге по его ID
// @Tags         tags
// @Produce      json
// @Param        id path int64 true "Tag ID"
// @Success      200  {object}  dto.TagResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /tags/{id} [get]
func (h *ExerciseHandler) GetTagByID(c *gin.Context) {
	tagID, err := parseInt64Param(c, "id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid tag id"})
		return
	}

	tag, err := h.svc.GetTagByID(c.Request.Context(), tagID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "tag not found"})
		return
	}

	c.JSON(http.StatusOK, h.tagToResponse(tag))
}

// GetExerciseTags получает теги упражнения
// @Summary      Получить теги упражнения
// @Description  Возвращает список тегов, связанных с указанным упражнением
// @Tags         exercises
// @Produce      json
// @Param        id path int64 true "Exercise ID"
// @Success      200  {array}   dto.TagResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /exercises/{id}/tags [get]
func (h *ExerciseHandler) GetExerciseTags(c *gin.Context) {
	exerciseID, err := parseInt64Param(c, "id") // Используем "exercise_id"
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid exercise id"})
		return
	}

	tags, err := h.svc.GetExerciseTags(c.Request.Context(), exerciseID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get exercise tags"})
		return
	}

	resp := make([]dto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		resp = append(resp, h.tagToResponse(tag))
	}

	c.JSON(http.StatusOK, resp)
}

// GetExercisesByMultipleTags получает упражнения по нескольким тегам
// @Summary      Получить упражнения по нескольким тегам
// @Description  Возвращает список упражнений, связанных со всеми указанными тегами
// @Tags         exercises
// @Accept       json
// @Produce      json
// @Param        request body dto.GetExercisesByMultipleTagsRequest true "Массив ID тегов"
// @Success      200  {array}   dto.ExerciseResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /exercises/by-tags [post]
func (h *ExerciseHandler) GetExercisesByMultipleTags(c *gin.Context) {
	var req dto.GetExercisesByMultipleTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	if len(req.TagIDs) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "at least one tag id is required"})
		return
	}

	exercises, err := h.svc.GetExercisesByMultipleTags(c.Request.Context(), req.TagIDs)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get exercises by tags"})
		return
	}

	resp := make([]dto.ExerciseResponse, 0, len(exercises))
	for _, exercise := range exercises {
		resp = append(resp, h.exerciseToResponse(exercise))
	}

	c.JSON(http.StatusOK, resp)
}

// GetPopularTags получает популярные теги
// @Summary      Получить популярные теги
// @Description  Возвращает список самых популярных тегов
// @Tags         tags
// @Produce      json
// @Param        limit query int false "Лимит тегов" default(10) minimum(1) maximum(50)
// @Success      200  {array}   dto.TagResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /tags/popular [get]
func (h *ExerciseHandler) GetPopularTags(c *gin.Context) {
	var req dto.GetPopularTagsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid query parameters"})
		return
	}

	// Устанавливаем значение по умолчанию
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Проверяем лимит
	if req.Limit < 1 || req.Limit > 50 {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "limit must be between 1 and 50"})
		return
	}

	tags, err := h.svc.GetPopularTags(c.Request.Context(), req.Limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get popular tags"})
		return
	}

	resp := make([]dto.TagResponse, 0, len(tags))
	for _, tag := range tags {
		resp = append(resp, h.tagToResponse(tag))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExerciseHandler) exerciseToResponse(exercise *svcexercise.Exercise) dto.ExerciseResponse {
	var tags []dto.TagResponse
	if exercise.Tags != nil {
		tags = make([]dto.TagResponse, 0, len(exercise.Tags))
		for _, tag := range exercise.Tags {
			tags = append(tags, h.tagToResponse(&tag))
		}
	}

	return dto.ExerciseResponse{
		ID:          exercise.ID,
		Title:       exercise.Title,
		Description: exercise.Description,
		VideoURL:    &exercise.VideoUrl,
		ImageURL:    &exercise.ImageUrl,
		Tags:        tags,
	}
}

func (h *ExerciseHandler) tagToResponse(tag *svcexercise.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:   tag.ID,
		Type: tag.Type,
	}
}
