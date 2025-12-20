package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/EnduranNSU/trainings/internal/domain"
	"github.com/shopspring/decimal"
)

func toDomainTags(genTags interface{}) []domain.Tag {
	var tags []domain.Tag = nil
	var jsonBytes []byte

	switch v := genTags.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	case json.RawMessage:
		jsonBytes = []byte(v)
	case sql.NullString:
		if v.Valid {
			jsonBytes = []byte(v.String)
		}
	default:
		if b, err := json.Marshal(v); err == nil {
			jsonBytes = b
		}
	}

	if len(jsonBytes) > 0 && string(jsonBytes) != "[]" && string(jsonBytes) != "null" {
		var rawTags []struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(jsonBytes, &rawTags); err == nil {
			tags = make([]domain.Tag, len(rawTags))
			for i, tag := range rawTags {
				tags[i] = domain.Tag{
					ID:   tag.ID,
					Type: tag.Type,
				}
			}
		}
	}
	return tags
}

func toDomainExercise(genExercises interface{}) []domain.Exercise {
	var tags []domain.Exercise = nil
	var jsonBytes []byte

	switch v := genExercises.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	case json.RawMessage:
		jsonBytes = []byte(v)
	case sql.NullString:
		if v.Valid {
			jsonBytes = []byte(v.String)
		}
	default:
		if b, err := json.Marshal(v); err == nil {
			jsonBytes = b
		}
	}

	if len(jsonBytes) > 0 && string(jsonBytes) != "[]" && string(jsonBytes) != "null" {
		var rawExercises []struct {
			ID          int64       `json:"id"`
			Description string      `json:"description"`
			Href        string      `json:"href"`
			Tags        interface{} `json:"tags"`
		}
		if err := json.Unmarshal(jsonBytes, &rawExercises); err == nil {
			tags = make([]domain.Exercise, len(rawExercises))
			for i, ex := range rawExercises {
				tags[i] = domain.Exercise{
					ID:          ex.ID,
					Description: ex.Description,
					Href:        ex.Href,
					Tags:        toDomainTags(ex.Tags),
				}
			}
		}
	}
	return tags
}

func toDomainTrainedExercise(genExercises interface{}) []domain.TrainedExercise {
	var tags []domain.TrainedExercise = nil
	var jsonBytes []byte

	switch v := genExercises.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	case json.RawMessage:
		jsonBytes = []byte(v)
	case sql.NullString:
		if v.Valid {
			jsonBytes = []byte(v.String)
		}
	default:
		if b, err := json.Marshal(v); err == nil {
			jsonBytes = b
		}
	}

	if len(jsonBytes) > 0 && string(jsonBytes) != "[]" && string(jsonBytes) != "null" {
		var rawExercises []struct {
			ID         int64          `json:"id"`
			TrainingID int64          `json:"training_id"`
			ExerciseID int64          `json:"exercise_id"`
			Weight     sql.NullString `json:"weight"`
			Approaches sql.NullInt32  `json:"approaches"`
			Reps       sql.NullInt32  `json:"reps"`
			Time       int64          `json:"time"`
			Doing      int64          `json:"doing"`
			Rest       int64          `json:"rest"`
			Notes      sql.NullString `json:"notes"`
		}
		if err := json.Unmarshal(jsonBytes, &rawExercises); err == nil {
			tags = make([]domain.TrainedExercise, len(rawExercises))
			for i, ex := range rawExercises {
				weight, _ := decimal.NewFromString(ex.Weight.String)
				tags[i] = domain.TrainedExercise{
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
		}
	}
	return tags
}
