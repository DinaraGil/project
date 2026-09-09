package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"projectOzonBank/internal/domain"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping postgres integration tests")
	}

	ctx := context.Background()
	storage, err := New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		_, err := storage.pool.Exec(context.Background(), "TRUNCATE links")
		if err != nil {
			t.Logf("cleanup failed: %v", err)
		}
		storage.Close()
	})

	return storage
}

func TestStorage_SaveAndGet(t *testing.T) {
	storage := newTestStorage(t)
	ctx := context.Background()

	code := "1234567890"
	originalURL := "https://google.com"

	if err := storage.Save(ctx, code, originalURL); err != nil {
		t.Fatalf("Save вернул ошибку: %v", err)
	}

	got, err := storage.Get(ctx, code)
	if err != nil {
		t.Fatalf("Get вернул ошибку: %v", err)
	}
	if got != originalURL {
		t.Errorf("originalURL = %s, ожидалось %s", got, originalURL)
	}
}

func TestStorage_Get_NotFound(t *testing.T) {
	storage := newTestStorage(t)
	ctx := context.Background()

	_, err := storage.Get(ctx, "0000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ожидалась ErrNotFound, получена: %v", err)
	}
}

func TestStorage_Save_DuplicateURL(t *testing.T) {
	storage := newTestStorage(t)
	ctx := context.Background()

	originalURL := "https://google.com"
	firstCode := "1111111111"
	secondCode := "2222222222"

	if err := storage.Save(ctx, firstCode, originalURL); err != nil {
		t.Fatalf("первый Save вернул ошибку: %v", err)
	}

	err := storage.Save(ctx, secondCode, originalURL)

	var existsErr *domain.AlreadyExistsError
	if !errors.As(err, &existsErr) {
		t.Fatalf("ожидалась AlreadyExistsError, получена: %v", err)
	}
	if existsErr.ExistingCode != firstCode {
		t.Errorf("ExistingCode = %s, ожидалось %s", existsErr.ExistingCode, firstCode)
	}
}

func TestStorage_Save_CodeTaken(t *testing.T) {
	storage := newTestStorage(t)
	ctx := context.Background()

	code := "3333333333"
	firstURL := "https://google.com"
	secondURL := "https://github.com"

	if err := storage.Save(ctx, code, firstURL); err != nil {
		t.Fatalf("первый Save вернул ошибку: %v", err)
	}

	err := storage.Save(ctx, code, secondURL)
	if !errors.Is(err, domain.ErrCodeAlreadyTaken) {
		t.Errorf("ожидалась ErrCodeAlreadyTaken, получена: %v", err)
	}
}
