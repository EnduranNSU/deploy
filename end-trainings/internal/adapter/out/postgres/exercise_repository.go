package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/EnduranNSU/trainings/internal/adapter/out/postgres/gen"
	"github.com/EnduranNSU/trainings/internal/domain"
	"github.com/EnduranNSU/trainings/internal/logging"
)

type ExerciseRepositoryImpl struct {
	q  *gen.Queries
	db *sql.DB
}

func NewExerciseRepository(db *sql.DB) domain.ExerciseRepository {
	return &ExerciseRepositoryImpl{
		q:  gen.New(db),
		db: db,
	}
}

func (r *ExerciseRepositoryImpl) GetExercisesWithTags(ctx context.Context) ([]*domain.Exercise, error) {
	exercises, err := r.q.GetExercisesWithTags(ctx)
	if err != nil {
		logging.Error(err, "GetExercisesWithTags", nil, "failed to get exercises with tags")
		return nil, err
	}
	result := make([]*domain.Exercise, len(exercises))
	for i, e := range exercises {
		result[i] = r.toDomainExercise(e)
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"exercises_count": len(result),
		"exercises":       result,
	})
	logging.Debug("GetExercisesWithTags", jsonData, "successfully retrieved exercises with tags")
	return result, nil
}

func (r *ExerciseRepositoryImpl) GetExerciseByID(ctx context.Context, id int64) (*domain.Exercise, error) {
	exercise, err := r.q.GetExerciseByID(ctx, id)
	if err == sql.ErrNoRows {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"exercise_id": id,
		})
		logging.Warn("GetExerciseByID", jsonData, "exercise not found")
		return nil, fmt.Errorf("exercise not found with id: %d", id)
	}

	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"exercise_id": id,
		})
		logging.Error(err, "GetExerciseByID", jsonData, "failed to get exercise by id")
		return nil, err
	}

	domainExercise := r.toDomainExerciseFromJoined(exercise)

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"exercise_id": id,
		"tags_count":  len(domainExercise.Tags),
	})
	logging.Debug("GetExerciseByID", jsonData, "successfully retrieved exercise by id")

	return domainExercise, nil
}

func (r *ExerciseRepositoryImpl) GetExercisesByTag(ctx context.Context, tagID int64) ([]*domain.Exercise, error) {
	exercises, err := r.q.GetExercisesByTag(ctx, tagID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"tag_id": tagID,
		})
		logging.Error(err, "GetExercisesByTag", jsonData, "failed to get exercises by tag")
		return nil, err
	}

	result := make([]*domain.Exercise, len(exercises))
	for i, e := range exercises {
		result[i] = &domain.Exercise{
			ID:          e.ID,
			Description: e.Description,
			VideoUrl:    e.VideoUrl,
			ImageUrl:    e.ImageUrl,
		}
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"tag_id":          tagID,
		"exercises_count": len(result),
	})
	logging.Debug("GetExercisesByTag", jsonData, "successfully retrieved exercises by tag")

	return result, nil
}

func (r *ExerciseRepositoryImpl) SearchExercises(ctx context.Context, filter domain.ExerciseFilter) ([]*domain.Exercise, error) {
	// Если есть фильтр по тегу, используем GetExercisesByTag
	if filter.TagID != nil {
		return r.GetExercisesByTag(ctx, *filter.TagID)
	}

	// Иначе получаем все упражнения и фильтруем по поиску
	exercises, err := r.GetExercisesWithTags(ctx)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"search": filter.Search,
			"tag_id": filter.TagID,
		})
		logging.Error(err, "SearchExercises", jsonData, "failed to search exercises")
		return nil, err
	}

	if filter.Search != nil && *filter.Search != "" {
		var filtered []*domain.Exercise
		for _, exercise := range exercises {
			if strings.Contains(strings.ToLower(exercise.Title), strings.ToLower(*filter.Search)) {
				filtered = append(filtered, exercise)
			}
		}

		jsonData := logging.MarshalLogData(map[string]interface{}{
			"search":          *filter.Search,
			"exercises_count": len(filtered),
		})
		logging.Debug("SearchExercises", jsonData, "successfully searched exercises")

		return filtered, nil
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"exercises_count": len(exercises),
	})
	logging.Debug("SearchExercises", jsonData, "successfully retrieved all exercises for search")

	return exercises, nil
}

func (r *ExerciseRepositoryImpl) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	tags, err := r.q.GetAllTags(ctx)
	if err != nil {
		logging.Error(err, "GetAllTags", nil, "failed to get all tags")
		return nil, err
	}

	result := make([]*domain.Tag, len(tags))
	for i, t := range tags {
		result[i] = &domain.Tag{
			ID:   t.ID,
			Type: t.Type,
		}
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"tags_count": len(result),
	})
	logging.Debug("GetAllTags", jsonData, "successfully retrieved all tags")

	return result, nil
}

func (r *ExerciseRepositoryImpl) GetTagByID(ctx context.Context, id int64) (*domain.Tag, error) {
	tags, err := r.q.GetAllTags(ctx)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"tag_id": id,
		})
		logging.Error(err, "GetTagByID", jsonData, "failed to get tag by id")
		return nil, err
	}

	for _, tag := range tags {
		if tag.ID == id {
			domainTag := &domain.Tag{
				ID:   tag.ID,
				Type: tag.Type,
			}

			jsonData := logging.MarshalLogData(map[string]interface{}{
				"tag_id": id,
				"type":   tag.Type,
			})
			logging.Debug("GetTagByID", jsonData, "successfully retrieved tag by id")

			return domainTag, nil
		}
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"tag_id": id,
	})
	logging.Warn("GetTagByID", jsonData, "tag not found")
	return nil, sql.ErrNoRows
}

func (r *ExerciseRepositoryImpl) GetExerciseTags(ctx context.Context, exerciseID int64) ([]*domain.Tag, error) {
	exercise, err := r.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"exercise_id": exerciseID,
		})
		logging.Error(err, "GetExerciseTags", jsonData, "failed to get exercise tags")
		return nil, err
	}

	// Преобразуем []domain.Tag в []*domain.Tag
	tags := make([]*domain.Tag, len(exercise.Tags))
	for i := range exercise.Tags {
		tags[i] = &exercise.Tags[i]
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"exercise_id": exerciseID,
		"tags_count":  len(tags),
	})
	logging.Debug("GetExerciseTags", jsonData, "successfully retrieved exercise tags")

	return tags, nil
}

func (r *ExerciseRepositoryImpl) toDomainExercise(e gen.GetExercisesWithTagsRow) *domain.Exercise {
	return &domain.Exercise{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		VideoUrl:    e.VideoUrl,
		ImageUrl:    e.ImageUrl,
		Tags:        toDomainTags(e.Tags),
	}
}

func (r *ExerciseRepositoryImpl) toDomainExerciseFromJoined(e gen.GetExerciseByIDRow) *domain.Exercise {
	return &domain.Exercise{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		VideoUrl:    e.VideoUrl,
		ImageUrl:    e.ImageUrl,
		Tags:        toDomainTags(e.Tags),
	}
}
