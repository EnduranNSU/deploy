// @title Training API
// @version 1.0
// @description Сервис информации о тренировках и упражнения
// @BasePath /api/v1
package httpin

import (
	"github.com/gin-gonic/gin"

	_ "github.com/EnduranNSU/trainings/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewGinRouter создает новый Gin router
// @title Enduran Training API
// @version 1.0
// @description Сервис информации о тренировках и упражнения
// @BasePath /api/v1
func NewGinRouter(training *TrainingHandler, exercise *ExerciseHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.StaticFile("/openapi.yaml", "docs/swagger.yaml")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// Training routes
		trainings := api.Group("/trainings")
		{
			// Основные операции с тренировками
			trainings.GET("", training.GetTrainingsByUser)
			trainings.POST("", training.CreateTraining)
			trainings.GET("/stats", training.GetUserTrainingStats)
			trainings.GET("/current", training.GetCurrentTraining)
			trainings.GET("/today", training.GetTodaysTraining)
			
			// Операции с конкретной тренировкой
			trainings.GET("/:id", training.GetTrainingWithExercises)
			trainings.PUT("/:id", training.UpdateTraining)
			trainings.DELETE("/:id", training.DeleteTraining)
			trainings.GET("/:id/stats", training.GetTrainingStats)
			trainings.GET("/:id/calculate-time", training.CalculateTrainingTotalTime)
			
			// Действия с тренировкой
			trainings.PATCH("/:id/complete", training.CompleteTraining)
			trainings.PATCH("/:id/mark-done", training.MarkTrainingAsDone)
			trainings.PATCH("/:id/start", training.StartTraining)
			trainings.PATCH("/:id/pause", training.PauseTraining)
			trainings.PATCH("/:id/resume", training.ResumeTraining)
			
			// Таймеры тренировки
			trainings.PATCH("/:id/timers", training.UpdateTrainingTimers)
		}

		// Training exercises routes
		trainingExercises := api.Group("/training-exercises")
		{
			trainingExercises.POST("", training.AddExerciseToTraining)
			trainingExercises.PUT("/:id", training.UpdateTrainedExercise)
			trainingExercises.DELETE("", training.RemoveExerciseFromTraining)
			
			// Действия с упражнениями
			trainingExercises.PATCH("/:id/time", training.UpdateExerciseTime)
			trainingExercises.PATCH("/:id/rest-time", training.UpdateExerciseRestTime)
			trainingExercises.PATCH("/:id/doing-time", training.UpdateExerciseDoingTime)
		}

		// Global trainings routes
		globalTrainings := api.Group("/global-trainings")
		{
			globalTrainings.GET("", training.GetGlobalTrainings)
			globalTrainings.POST("/assign", training.AssignGlobalTraining)
			
			// Операции с глобальной тренировкой по уровню
			globalTrainings.GET("/level/:level", training.GetGlobalTrainingByLevel)
			globalTrainings.GET("/:id", training.GetGlobalTrainingById)
		}

		// Exercise routes
		exercises := api.Group("/exercises")
		{
			exercises.GET("", exercise.GetAllExercises)
			exercises.GET("/search", exercise.SearchExercises)
			exercises.POST("/by-tags", exercise.GetExercisesByMultipleTags)
			exercises.GET("/:id/tags", exercise.GetExerciseTags)
			exercises.GET("/:id", exercise.GetExerciseByID)
		}

		// Tag routes
		tags := api.Group("/tags")
		{
			tags.GET("", exercise.GetAllTags)
		}
	}

	return r
}