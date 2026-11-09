package ai

import (
	"context"
	"log"
	"sync"
	"time"

	"cloudprep-backend/internal/models"
)

// ReplenishmentWorker manages non-blocking background question generation to grow the bank
// whenever a learner's unseen pool for a topic drops below threshold.
type ReplenishmentWorker struct {
	Router   *MultiProviderAiRouter
	inFlight sync.Map // map[string]time.Time to prevent concurrent duplicate generation per topic
}

func NewReplenishmentWorker(router *MultiProviderAiRouter) *ReplenishmentWorker {
	return &ReplenishmentWorker{
		Router: router,
	}
}

// Hook returns a ReplenishmentHook closure suitable for wiring into SessionHandler.
func (w *ReplenishmentWorker) Hook() func(topicID, questionStyle, difficulty string) {
	return func(topicID, questionStyle, difficulty string) {
		w.TriggerReplenishment(topicID, questionStyle, difficulty)
	}
}

// TriggerReplenishment schedules background question generation with deduplication cooldown.
func (w *ReplenishmentWorker) TriggerReplenishment(topicID, style, diff string) {
	if w.Router == nil {
		return
	}

	key := topicID
	now := time.Now()

	// Rate-limit replenishment to once every 5 minutes per topic
	if val, ok := w.inFlight.Load(key); ok {
		lastTime, _ := val.(time.Time)
		if now.Sub(lastTime) < 5*time.Minute {
			return
		}
	}
	w.inFlight.Store(key, now)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[replenishment] Recovered from panic in background generator: %v", r)
			}
		}()

		// Allow generous timeout for primary (30s) + secondary failover (30s) + DB persist (15s)
		ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
		defer cancel()

		reqStyle := NormalizeStyle(style, models.StyleScenario)
		reqDiff := NormalizeDifficulty(diff)

		req := GenerationRequest{
			TopicID:       topicID,
			QuestionStyle: reqStyle,
			Difficulty:    reqDiff,
			Count:         3,
		}

		log.Printf("[replenishment] Initiating background AI question replenishment for topic %s (%s, %s)...", topicID, reqStyle, reqDiff)
		accepted, record, err := w.Router.GenerateAndPersist(ctx, req)
		if err != nil {
			log.Printf("[replenishment] Background replenishment failed for topic %s: %v", topicID, err)
			return
		}

		log.Printf("[replenishment] Completed background generation for topic %s: %d questions accepted into bank via %s (failoverUsed=%v, duration=%dms).",
			topicID, len(accepted), record.Provider, record.FailoverUsed, record.DurationMs)
	}()
}

