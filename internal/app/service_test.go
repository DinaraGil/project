package app

import (
	"context"
	"errors"
	"testing"

	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/shortener"
	"projectOzonBank/internal/storage/memory"
)

func TestLinkService_Shorten_Success(t *testing.T) {
	ctx := context.Background()
	storage := memory.New()
	service := New(storage)

	code, err := service.Shorten(ctx, "https://google.com")
	if err != nil {
		t.Fatalf("Shorten вернул ошибку: %v", err)
	}

	if len(code) != shortener.CodeLength {
		t.Errorf(
			"длина code = %d, ожидалось %d",
			len(code),
			shortener.CodeLength,
		)
	}

	originalURL, err := storage.Get(ctx, code)
	if err != nil {
		t.Fatalf("Get вернул ошибку: %v", err)
	}

	if originalURL != "https://google.com" {
		t.Errorf(
			"originalURL = %s, ожидалось https://google.com",
			originalURL,
		)
	}
}

func TestLinkService_Shorten_InvalidURL(t *testing.T) {
	ctx := context.Background()

	storage := memory.New()
	service := New(storage)

	invalidURLs := []string{
		"google.com",
		"ftp://google.com",
		"https://",
		"not-a-url",
	}

	for _, originalURL := range invalidURLs {
		t.Run(originalURL, func(t *testing.T) {
			_, err := service.Shorten(ctx, originalURL)
			if !errors.Is(err, domain.ErrInvalidURL) {
				t.Errorf(
					"ожидалась ErrInvalidURL, получена: %v",
					err,
				)
			}
		})
	}
}

func TestLinkService_Shorten_DuplicateURL_ReturnsSameCode(t *testing.T) {
	ctx := context.Background()

	storage := memory.New()
	service := New(storage)

	originalURL := "https://google.com"

	firstCode, err := service.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("первый Shorten вернул ошибку: %v", err)
	}

	secondCode, err := service.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("второй Shorten вернул ошибку: %v", err)
	}

	if firstCode != secondCode {
		t.Errorf(
			"коды отличаются: первый=%s, второй=%s",
			firstCode,
			secondCode,
		)
	}
}

func TestLinkService_Resolve_Success(t *testing.T) {
	ctx := context.Background()

	storage := memory.New()
	service := New(storage)

	originalURL := "https://google.com"

	code, err := service.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("Shorten вернул ошибку: %v", err)
	}

	result, err := service.Resolve(ctx, code)
	if err != nil {
		t.Fatalf("Resolve вернул ошибку: %v", err)
	}

	if result != originalURL {
		t.Errorf(
			"Resolve вернул %s, ожидалось %s",
			result,
			originalURL,
		)
	}
}

func TestLinkService_Resolve_NotFound(t *testing.T) {
	ctx := context.Background()

	storage := memory.New()
	service := New(storage)

	_, err := service.Resolve(ctx, "1234567890")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf(
			"ожидалась ErrNotFound, получена: %v",
			err,
		)
	}
}
