package domain

import (
	"errors"
	"fmt"
)
// TODO: зачем по-русски??
var (
	ErrNotFound         = errors.New("Сокращенная ссылка не найдена в хранилище")
	ErrInvalidURL       = errors.New("Ориг ссылка неправильного формата")
	ErrCodeAlreadyTaken = errors.New("Сокращенная ссылка уже зарегистрирована")
	ErrGenerationFailed = errors.New("не удалось сгенерировать уникальный код")
)
// TODO: Код alreadyExists вроде существует? МБ 409?
type AlreadyExistsError struct {
	ExistingCode string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("URL уже сокращен, его сокращенная версия: %s", e.ExistingCode)
}
