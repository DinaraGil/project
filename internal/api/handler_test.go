package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"projectOzonBank/internal/domain"
)

type fakeService struct {
	shortenFunc func(ctx context.Context, url string) (string, error)
	resolveFunc func(ctx context.Context, code string) (string, error)
}

func (f *fakeService) Shorten(
	ctx context.Context,
	url string,
) (string, error) {
	return f.shortenFunc(ctx, url)
}

func (f *fakeService) Resolve(
	ctx context.Context,
	code string,
) (string, error) {
	return f.resolveFunc(ctx, code)
}

func TestHandler_Shorten_Success(t *testing.T) {
	service := &fakeService{
		shortenFunc: func(ctx context.Context, url string) (string, error) {
			return "AbCd123456", nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":"https://google.com"}`),
	)

	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("ожидаемый статус  %d, не совпал с %d", http.StatusCreated, w.Code)
	}

	expected := `{"code":"AbCd123456"}`

	if !strings.Contains(w.Body.String(), expected) {
		t.Errorf("тело должно было содержать %q, получили %q", expected, w.Body.String())
	}
}

func TestHandler_Shorten_InvalidJSON(t *testing.T) {
	service := &fakeService{
		shortenFunc: func(ctx context.Context, url string) (string, error) {
			t.Fatal("сервис не может быть вызван")
			return "", nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":`),
	)

	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидаемый статус %d, не совпал с результатом %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_Shorten_InvalidURL(t *testing.T) {
	service := &fakeService{
		shortenFunc: func(ctx context.Context, url string) (string, error) {
			return "", domain.ErrInvalidURL
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":"invalid"}`),
	)

	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидаемый статус %d не совпал с результатом %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_Shorten_ServiceError(t *testing.T) {
	service := &fakeService{
		shortenFunc: func(ctx context.Context, url string) (string, error) {
			return "", errors.New("какая ошибка на сервере")
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(`{"url":"https://google.com"}`),
	)

	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf(
			"ожидался статус %d, получили %d",
			http.StatusInternalServerError,
			w.Code,
		)
	}
}

func TestHandler_Resolve_Success(t *testing.T) {
	service := &fakeService{
		resolveFunc: func(ctx context.Context, code string) (string, error) {
			return "https://google.com", nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/AbCd123456",
		nil,
	)

	req.SetPathValue("code", "AbCd123456")

	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("\"ожидался статус %d, получили %d\"", http.StatusOK, w.Code)
	}

	expected := `{"original_url":"https://google.com"}`

	if !strings.Contains(w.Body.String(), expected) {
		t.Errorf("ожидалось содержание %q, получили %q", expected, w.Body.String())
	}
}

func TestHandler_Resolve_NotFound(t *testing.T) {
	service := &fakeService{
		resolveFunc: func(ctx context.Context, code string) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/AbCd123456",
		nil,
	)

	req.SetPathValue("code", "AbCd123456")

	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf(
			"ожидался статус %d, получили %d\"",
			http.StatusNotFound,
			w.Code,
		)
	}
}
