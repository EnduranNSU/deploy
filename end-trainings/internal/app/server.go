package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpin "github.com/EnduranNSU/trainings/internal/adapter/in/http"
	svc "github.com/EnduranNSU/trainings/internal/domain"
	"github.com/rs/zerolog/log"
)

type Server struct {
	TrainingSvc svc.TrainingService
	ExerciseSvc svc.ExerciseService
	Addr string
}

func SetupServer(trainingSvc svc.TrainingService,
	exerciseSvc svc.ExerciseService, addr string) *Server {
	return &Server{
		TrainingSvc: trainingSvc,
		ExerciseSvc: exerciseSvc,
		Addr: addr,
	}
}

func (s *Server) StartServer() error {
	eh := httpin.NewExerciseHandler(s.ExerciseSvc)
	th := httpin.NewTrainingHandler(s.TrainingSvc)
	engine := httpin.NewGinRouter(th, eh)

	srv := &http.Server{
		Addr:              s.Addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	// Запуск сервера в отдельной горутине
	go func() {
		log.Info().Msgf("HTTP server starting on %s", s.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).
			Str("service", "trainings").Msg("HTTP server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().
	Str("service", "trainings").Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).
		Str("service", "trainings").Msg("HTTP server forced to shutdown")
		return err
	}

	log.Info().
	Str("service", "trainings").Msg("Server stopped gracefully")
	return nil
}
