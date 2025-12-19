package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/EnduranNSU/trainings/internal/domain"
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
			ID   int64  `json:"id"`
			Description string `json:"description"`
			Href string `json:"href"`
			Tags interface{} `json:"tags"`
		}
		if err := json.Unmarshal(jsonBytes, &rawExercises); err == nil {
			tags = make([]domain.Exercise, len(rawExercises))
			for i, ex := range rawExercises {
				tags[i] = domain.Exercise{
					ID:   ex.ID,
					Description: ex.Description,
					Href: ex.Href,
					Tags: toDomainTags(ex.Tags),
				}
			}
		}
	}
	return tags
}
