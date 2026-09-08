package shortener

import (
	"regexp"
	"testing"
)

func TestGenerate_Length(t *testing.T) {
	realResult, err := Generate()
	if len(realResult) != CodeLength {
		t.Errorf("Длина искомого %d не равна ожидаемоей длине %d", len(realResult), CodeLength)
	}
	if err != nil {
		t.Errorf("Generate вернула ошибку %v", err)
	}
}
func TestGenerate_Alphabet(t *testing.T) {
	re := regexp.MustCompile(`^[A-Za-z0-9_]{10}$`)
	result, err := Generate()
	if err != nil {
		t.Fatalf("Generate вернула ошибку %v", err)
	}
	ok := re.MatchString(result)
	if !ok {
		t.Errorf("Алфавит для генерации не был соблюден!")
	}
}
func TestGenerate_NoPanicManyTimes(t *testing.T) {
	for i := 0; i < 10000; i++ {
		_, err := Generate()
		if err != nil {
			t.Errorf("Generate вернула ошибку %v", err)
		}
	}
}
func TestGenerate_Uniqueness(t *testing.T) {
	storage := make(map[string]bool)
	countToGenerate := 100000
	var temp string
	var err error
	for i := 0; i < countToGenerate; i++ {
		temp, err = Generate()
		if err != nil {
			t.Fatalf("Generate вернула ошибку %v", err)
		}
		storage[temp] = true
	}
	if len(storage) != countToGenerate {
		t.Errorf("Кол-во сгенерированных данных %d не равна ожидаемому кол-ву  %d", len(storage), countToGenerate)

	}
}
