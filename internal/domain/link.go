package domain

import "context"
// TODO: зачем здесь описаны эти интерфейсы и на уровне домена ничего кроме интерфейсов?
type Link struct {
	ShortCode   string
	OriginalURL string
}
type Storage interface {
	Save(ctx context.Context, code, originalURL string) error
	Get(ctx context.Context, code string) (originalURL string, err error)
}
