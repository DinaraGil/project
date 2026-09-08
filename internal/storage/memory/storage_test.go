package memory

import (
	"context"
	"errors"
	"fmt"
	"projectOzonBank/internal/domain"
	"sync"
	"testing"
)

const (
	code        = "1234567890"
	originalURL = "google.com"
)

func TestStorage_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	st := New()
	err := st.Save(ctx, code, originalURL)
	if err != nil {
		t.Fatalf("Ошибка %v", err)
	}
	realUrl, err := st.Get(ctx, code)
	if err != nil {
		t.Fatalf("Ошибка %v", err)
	}
	if realUrl != originalURL {
		t.Errorf("Полученная ссылка %s не равна ожидаемой %s", realUrl, originalURL)
	}
}

func TestStorage_Get_NotFound(t *testing.T) {
	st := New()
	ctx := context.Background()
	_, err := st.Get(ctx, code)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Полученная ошибка %v не равна ожидаемой %v", err, domain.ErrNotFound)
	}
}

func TestStorage_Save_DuplicateURL(t *testing.T) {
	ctx := context.Background()
	st := New()

	firstCode := "1234567890"
	secondCode := "abcdefghij"

	err := st.Save(ctx, firstCode, originalURL)
	if err != nil {
		t.Fatalf("первый Save вернул ошибку: %v", err)
	}

	err = st.Save(ctx, secondCode, originalURL)

	var existsErr *domain.AlreadyExistsError
	if !errors.As(err, &existsErr) {
		t.Fatalf("ожидалась AlreadyExistsError, получена: %v", err)
	}

	if existsErr.ExistingCode != firstCode {
		t.Errorf(
			"ожидался ExistingCode %s, получен %s",
			firstCode,
			existsErr.ExistingCode,
		)
	}
}

func TestStorage_Save_CodeTaken(t *testing.T) {
	ctx := context.Background()
	st := New()

	firstURL := "https://google.com"
	secondURL := "https://github.com"

	err := st.Save(ctx, code, firstURL)
	if err != nil {
		t.Fatalf("первый Save вернул ошибку: %v", err)
	}

	err = st.Save(ctx, code, secondURL)

	if !errors.Is(err, domain.ErrCodeAlreadyTaken) {
		t.Errorf(
			"ожидалась ErrCodeAlreadyTaken, получена: %v",
			err,
		)
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	st := New()
	ctx := context.Background()

	const count = 1000

	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			code := fmt.Sprintf("%010d", i)
			url := fmt.Sprintf("https://example.com/%d", i)

			if err := st.Save(ctx, code, url); err != nil {
				t.Errorf("Save вернул ошибку: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if len(st.codeToURL) != count {
		t.Errorf("ожидалось %d записей, получено %d", count, len(st.codeToURL))
	}

	if len(st.urlToCode) != count {
		t.Errorf("ожидалось %d записей, получено %d", count, len(st.urlToCode))
	}
}
func TestStorage_ConcurrentAccess_SameURL(t *testing.T) {
	st := New()
	ctx := context.Background()

	const count = 100
	sameURL := "https://example.com/popular"

	var wg sync.WaitGroup
	wg.Add(count)

	var mu sync.Mutex
	successCount := 0

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			code := fmt.Sprintf("code, %06d", i)
			err := st.Save(ctx, code, sameURL)

			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
				return
			}

			var existsErr *domain.AlreadyExistsError
			if !errors.As(err, &existsErr) {
				t.Errorf("неожиданная ошибка: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("ожидался 1 успешный Save, получено %d", successCount)
	}
}
