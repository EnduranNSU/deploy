package dto

// ExerciseResponse представляет ответ с информацией об упражнении
type ExerciseResponse struct {
	ID          int64         `json:"id" example:"1" description:"ID упражнения"`
	Title       string        `json:"title" example:"Жим жопой" description:"Название упражнения"`
	Description string        `json:"description" example:"Базовое упражнение для развития грудных мышц" description:"Описание упражнения"`
	VideoURL    *string       `json:"video_url,omitempty" example:"https://example.com/video.mp4" description:"Ссылка на видео с техникой выполнения"`
	ImageURL    *string       `json:"image_url,omitempty" example:"https://example.com/video.mp4" description:"Ссылка на картинку"`
	Tags        []TagResponse `json:"tags,omitempty" description:"Теги упражнения"`
}

// TagResponse представляет ответ с информацией о теге
type TagResponse struct {
	ID   int64  `json:"id" example:"1" description:"ID тега"`
	Type string `json:"type" example:"силовое" description:"Название тега"`
}

// SearchExercisesRequest представляет запрос на поиск упражнений
type SearchExercisesRequest struct {
	Query string `json:"query" form:"query" example:"жим" description:"Поисковый запрос"`
	TagID *int64 `json:"tag_id,omitempty" form:"tag_id" example:"1" description:"ID тега для фильтрации (опционально)"`
}

// GetExercisesByMultipleTagsRequest представляет запрос на получение упражнений по нескольким тегам
type GetExercisesByMultipleTagsRequest struct {
	TagIDs []int64 `json:"tag_ids" binding:"required,min=1" example:"[1,2,3]" description:"Массив ID тегов"`
}

// GetPopularTagsRequest представляет запрос на получение популярных тегов
type GetPopularTagsRequest struct {
	Limit int `json:"limit" form:"limit" binding:"min=1,max=50" example:"10" description:"Лимит тегов (1-50)"`
}
