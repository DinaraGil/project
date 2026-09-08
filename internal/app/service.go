package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/shortener"
)

const defaultMaxRetries = 5

type LinkService struct {
	storage    domain.Storage
	maxRetries int
}

func New(storage domain.Storage) *LinkService {
	return &LinkService{
		storage:    storage,
		maxRetries: defaultMaxRetries,
	}
}

func (s *LinkService) Shorten(ctx context.Context, originalURL string) (string, error) {
	parsed, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", domain.ErrInvalidURL
	}

	if parsed.Host == "" {
		return "", domain.ErrInvalidURL
	}

	for i := 0; i < s.maxRetries; i++ {
		code, err := shortener.Generate()
		if err != nil {
			return "", err
		}

		err = s.storage.Save(ctx, code, originalURL)
		if err == nil {
			return code, nil
		}

		var existsErr *domain.AlreadyExistsError
		if errors.As(err, &existsErr) {
			return existsErr.ExistingCode, nil
		}

		if errors.Is(err, domain.ErrCodeAlreadyTaken) {
			continue
		}

		return "", err
	}

	return "", fmt.Errorf(
		"Не получилось сгенерировать код после попыток %d ",
		s.maxRetries,
	)
}

func (s *LinkService) Resolve(ctx context.Context, code string) (string, error) {
	if len(code) != shortener.CodeLength {
		return "", domain.ErrNotFound
	}

	return s.storage.Get(ctx, code)
}
