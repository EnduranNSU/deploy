package service

import (
	"context"
	"errors"
	"strings"

	"github.com/EnduranNSU/trainings/internal/domain"
)

var (
	ErrInvalidExerciseID = errors.New("invalid exercise id")
	ErrInvalidTagID      = errors.New("invalid tag id")
	ErrExerciseNotFound  = errors.New("exercise not found")
	ErrTagNotFound       = errors.New("tag not found")
	ErrEmptySearchQuery  = errors.New("search query cannot be empty")
)

func NewExerciseService(repo domain.ExerciseRepository) domain.ExerciseService {
	return &exerciseService{repo: repo}
}

type exerciseService struct {
	repo domain.ExerciseRepository
}

func (s *exerciseService) GetAllExercises(ctx context.Context) ([]*domain.Exercise, error) {
	return s.repo.GetExercisesWithTags(ctx)
}

func (s *exerciseService) GetExerciseByID(ctx context.Context, id int64) (*domain.Exercise, error) {
	if id <= 0 {
		return nil, ErrInvalidExerciseID
	}

	return s.repo.GetExerciseByID(ctx, id)
}

func (s *exerciseService) GetExercisesByTag(ctx context.Context, tagID int64) ([]*domain.Exercise, error) {
	if tagID <= 0 {
		return nil, ErrInvalidTagID
	}

	// Проверяем существование тега
	_, err := s.repo.GetTagByID(ctx, tagID)
	if err != nil {
		return nil, ErrTagNotFound
	}

	return s.repo.GetExercisesByTag(ctx, tagID)
}

func (s *exerciseService) SearchExercises(ctx context.Context, query string, tagID *int64) ([]*domain.Exercise, error) {
	query = strings.TrimSpace(query)
	
	filter := domain.ExerciseFilter{
		Search: &query,
		TagID:  tagID,
	}

	// Если передан пустой поисковый запрос и нет тега, возвращаем все упражнения
	if query == "" && tagID == nil {
		return s.GetAllExercises(ctx)
	}

	// Если передан пустой поисковый запрос, но есть тег, возвращаем упражнения по тегу
	if query == "" && tagID != nil {
		return s.GetExercisesByTag(ctx, *tagID)
	}

	// Если поисковый запрос слишком короткий
	if len(query) < 2 {
		return nil, errors.New("search query must be at least 2 characters long")
	}

	return s.repo.SearchExercises(ctx, filter)
}

func (s *exerciseService) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	return s.repo.GetAllTags(ctx)
}

func (s *exerciseService) GetTagByID(ctx context.Context, id int64) (*domain.Tag, error) {
	if id <= 0 {
		return nil, ErrInvalidTagID
	}

	return s.repo.GetTagByID(ctx, id)
}

func (s *exerciseService) GetExerciseTags(ctx context.Context, exerciseID int64) ([]*domain.Tag, error) {
	if exerciseID <= 0 {
		return nil, ErrInvalidExerciseID
	}

	// Проверяем существование упражнения
	_, err := s.repo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, ErrExerciseNotFound
	}

	return s.repo.GetExerciseTags(ctx, exerciseID)
}

func (s *exerciseService) GetExercisesByMultipleTags(ctx context.Context, tagIDs []int64) ([]*domain.Exercise, error) {
	if len(tagIDs) == 0 {
		return nil, errors.New("at least one tag id is required")
	}

	// Проверяем существование всех тегов
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			return nil, ErrInvalidTagID
		}
		_, err := s.repo.GetTagByID(ctx, tagID)
		if err != nil {
			return nil, ErrTagNotFound
		}
	}

	// Получаем все упражнения и фильтруем по нескольким тегам
	allExercises, err := s.repo.GetExercisesWithTags(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Exercise
	for _, exercise := range allExercises {
		if s.exerciseHasAllTags(exercise, tagIDs) {
			filtered = append(filtered, exercise)
		}
	}

	return filtered, nil
}

func (s *exerciseService) GetPopularTags(ctx context.Context, limit int) ([]*domain.Tag, error) {
	if limit <= 0 {
		limit = 10 // значение по умолчанию
	}

	allExercises, err := s.repo.GetExercisesWithTags(ctx)
	if err != nil {
		return nil, err
	}

	// Считаем популярность тегов
	tagCount := make(map[int64]int)
	for _, exercise := range allExercises {
		for _, tag := range exercise.Tags {
			tagCount[tag.ID]++
		}
	}

	// Получаем все теги для получения полной информации
	allTags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		return nil, err
	}

	// Сортируем теги по популярности
	tagMap := make(map[int64]*domain.Tag)
	for _, tag := range allTags {
		tagMap[tag.ID] = tag
	}

	// Создаем слайс тегов с количеством использований
	type tagWithCount struct {
		tag   *domain.Tag
		count int
	}

	var tagsWithCount []tagWithCount
	for tagID, count := range tagCount {
		if tag, exists := tagMap[tagID]; exists {
			tagsWithCount = append(tagsWithCount, tagWithCount{tag: tag, count: count})
		}
	}

	// Сортируем по убыванию популярности (простая пузырьковая сортировка)
	for i := 0; i < len(tagsWithCount)-1; i++ {
		for j := i + 1; j < len(tagsWithCount); j++ {
			if tagsWithCount[i].count < tagsWithCount[j].count {
				tagsWithCount[i], tagsWithCount[j] = tagsWithCount[j], tagsWithCount[i]
			}
		}
	}

	// Возвращаем топ-N тегов
	resultCount := limit
	if len(tagsWithCount) < limit {
		resultCount = len(tagsWithCount)
	}

	result := make([]*domain.Tag, resultCount)
	for i := 0; i < resultCount; i++ {
		result[i] = tagsWithCount[i].tag
	}

	return result, nil
}

// Вспомогательный метод для проверки наличия всех тегов у упражнения
func (s *exerciseService) exerciseHasAllTags(exercise *domain.Exercise, tagIDs []int64) bool {
	exerciseTagMap := make(map[int64]bool)
	for _, tag := range exercise.Tags {
		exerciseTagMap[tag.ID] = true
	}

	for _, tagID := range tagIDs {
		if !exerciseTagMap[tagID] {
			return false
		}
	}

	return true
}