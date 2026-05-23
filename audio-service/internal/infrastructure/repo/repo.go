package repo

import (
	"context"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/notify"
)

type AudioRepo interface {
	GetAudio(ctx context.Context, id string) (audio.Audio, error)
	CreateAudio(ctx context.Context, audio audio.Audio) error
	UpdateAudio(ctx context.Context, audio audio.UpdateAudio, version int64) error
	DeleteAudio(ctx context.Context, id string) error
}

type TransactionalOutboxRepo interface {
	CreateTransactionalOutboxRecord(ctx context.Context, transactionalOutboxRecrod notify.TransactionalOutboxRecord) error
}

type Repository interface {
	AudioRepo() AudioRepo
	TransactionalOutboxRepo() TransactionalOutboxRepo
}
