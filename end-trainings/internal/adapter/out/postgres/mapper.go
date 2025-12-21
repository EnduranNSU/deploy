package postgres

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

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
			Title       string      `json:"title"`
			Description string      `json:"description"`
			VideoUrl    string      `json:"video_url"`
			ImageUrl    string      `json:"image_url"`
			Tags        interface{} `json:"tags"`
		}
		if err := json.Unmarshal(jsonBytes, &rawExercises); err == nil {
			tags = make([]domain.Exercise, len(rawExercises))
			for i, ex := range rawExercises {
				tags[i] = domain.Exercise{
					ID:          ex.ID,
					Title:       ex.Title,
					Description: ex.Description,
					VideoUrl:    ex.VideoUrl,
					ImageUrl:    ex.ImageUrl,
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
		// Try to decode Base64 first
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			jsonBytes = decoded
		} else {
			// If not Base64, treat as regular string
			jsonBytes = []byte(v)
		}
	case json.RawMessage:
		jsonBytes = []byte(v)
	case sql.NullString:
		if v.Valid {
			// Try to decode Base64 first
			if decoded, err := base64.StdEncoding.DecodeString(v.String); err == nil {
				jsonBytes = decoded
			} else {
				// If not Base64, treat as regular string
				jsonBytes = []byte(v.String)
			}
		}
	default:
		if b, err := json.Marshal(v); err == nil {
			jsonBytes = b
		}
	}

	if len(jsonBytes) > 0 && string(jsonBytes) != "[]" && string(jsonBytes) != "null" {
		var rawExercises []struct {
			ID         int64       `json:"id"`
			TrainingID int64       `json:"training_id"`
			ExerciseID int64       `json:"exercise_id"`
			Weight     interface{} `json:"weight"` // Use interface{} to handle both string and number
			Approaches int32       `json:"approaches"`
			Reps       int32       `json:"reps"`
			Time       int64       `json:"time"`
			Doing      int64       `json:"doing"`
			Rest       int64       `json:"rest"`
			Notes      string      `json:"notes"`
		}
		if err := json.Unmarshal(jsonBytes, &rawExercises); err == nil {
			tags = make([]domain.TrainedExercise, len(rawExercises))
			for i, ex := range rawExercises {
				var weightPtr *decimal.Decimal

				// Handle weight field which can be string, number, or null
				if ex.Weight != nil {
					switch w := ex.Weight.(type) {
					case string:
						if w != "" {
							weight, err := decimal.NewFromString(w)
							if err == nil {
								weightPtr = &weight
							}
						}
					case float64:
						weight := decimal.NewFromFloat(w)
						weightPtr = &weight
					case int64:
						weight := decimal.NewFromInt(w)
						weightPtr = &weight
					case int:
						weight := decimal.NewFromInt(int64(w))
						weightPtr = &weight
					case float32:
						weight := decimal.NewFromFloat32(w)
						weightPtr = &weight
					}
				}

				tags[i] = domain.TrainedExercise{
					ID:         ex.ID,
					TrainingID: ex.TrainingID,
					ExerciseID: ex.ExerciseID,
					Weight:     weightPtr,
					Approaches: &ex.Approaches,
					Reps:       &ex.Reps,
					Time:       toDuration(ex.Time),
					Doing:      toDuration(ex.Doing),
					Rest:       toDuration(ex.Rest),
					Notes:      &ex.Notes,
				}
			}
		} else {
			// Log the error for debugging
			fmt.Printf("Failed to unmarshal exercises JSON: %v\n", err)
			fmt.Printf("JSON data: %s\n", string(jsonBytes))
		}
	}
	return tags
}
