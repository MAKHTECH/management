package repository

import ()

type AuthRepository interface {
}

// PostgresRepository объединяет все PostgreSQL репозитории
type PostgresRepository interface {
	Close() error
}

// RedisRepository объединяет все Redis репозитории
type RedisRepository interface {
}
