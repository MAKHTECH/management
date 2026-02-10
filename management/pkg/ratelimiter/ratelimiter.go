package ratelimiter

import (
	"sync"
	"time"
)

// TokenBucket реализует Rate Limiter с использованием Token Bucket алгоритма
// для ограничения запросов по access token
type TokenBucket struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	rate     int           // количество токенов в секунду
	capacity int           // максимальная ёмкость bucket
	cleanup  time.Duration // интервал очистки неиспользуемых buckets
}

type bucket struct {
	tokens     float64
	lastUpdate time.Time
}

// Config конфигурация Rate Limiter
type Config struct {
	// Rate количество запросов в секунду на один access token
	Rate int `json:"rate"`
	// Capacity максимальное количество накопленных запросов (burst)
	Capacity int `json:"capacity"`
	// CleanupInterval интервал очистки неактивных buckets
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
// 10 запросов в секунду с burst до 20 запросов
func DefaultConfig() Config {
	return Config{
		Rate:            10,
		Capacity:        20,
		CleanupInterval: 5 * time.Minute,
	}
}

// New создаёт новый Rate Limiter
func New(cfg Config) *TokenBucket {
	if cfg.Rate <= 0 {
		cfg.Rate = 10
	}
	if cfg.Capacity <= 0 {
		cfg.Capacity = cfg.Rate * 2
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}

	tb := &TokenBucket{
		buckets:  make(map[string]*bucket),
		rate:     cfg.Rate,
		capacity: cfg.Capacity,
		cleanup:  cfg.CleanupInterval,
	}

	// Запускаем фоновую очистку неактивных buckets
	go tb.startCleanup()

	return tb
}

// Allow проверяет, разрешён ли запрос для указанного токена
// Возвращает true, если запрос разрешён, и количество оставшихся токенов
func (tb *TokenBucket) Allow(accessToken string) (allowed bool, remaining int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, exists := tb.buckets[accessToken]
	now := time.Now()

	if !exists {
		// Создаём новый bucket для токена
		tb.buckets[accessToken] = &bucket{
			tokens:     float64(tb.capacity - 1), // вычитаем 1 за текущий запрос
			lastUpdate: now,
		}
		return true, tb.capacity - 1
	}

	// Вычисляем, сколько токенов накопилось с момента последнего обновления
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * float64(tb.rate)

	// Ограничиваем максимумом
	if b.tokens > float64(tb.capacity) {
		b.tokens = float64(tb.capacity)
	}

	b.lastUpdate = now

	// Проверяем, есть ли доступные токены
	if b.tokens >= 1 {
		b.tokens--
		return true, int(b.tokens)
	}

	return false, 0
}

// Reset сбрасывает лимит для указанного токена
func (tb *TokenBucket) Reset(accessToken string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	delete(tb.buckets, accessToken)
}

// startCleanup запускает периодическую очистку неактивных buckets
func (tb *TokenBucket) startCleanup() {
	ticker := time.NewTicker(tb.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		tb.cleanup_old_buckets()
	}
}

// cleanup_old_buckets удаляет buckets, которые не использовались дольше cleanup интервала
func (tb *TokenBucket) cleanup_old_buckets() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	threshold := time.Now().Add(-tb.cleanup)

	for token, b := range tb.buckets {
		if b.lastUpdate.Before(threshold) {
			delete(tb.buckets, token)
		}
	}
}

// Stats возвращает количество активных buckets (для мониторинга)
func (tb *TokenBucket) Stats() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return len(tb.buckets)
}
