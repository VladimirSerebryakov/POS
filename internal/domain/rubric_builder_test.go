package domain_test

import (
	"testing"

	"pos/internal/domain"
)

func TestRubricBuilderBuild(t *testing.T) {
	builder := domain.NewRubricBuilder()
	criteria := []*domain.Criterion{
		newRubricTestCriterion(t, 7, "Ходил на пары"),
		newRubricTestCriterion(t, 2, "Ходил на пары"),
	}
	for _, criterion := range criteria {
		if err := builder.AddCriterion(criterion); err != nil {
			t.Fatal(err)
		}
	}
	rubric, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := rubric.Criteria()
	if len(got) != len(criteria) {
		t.Fatalf("got %d criteria, want %d", len(got), len(criteria))
	}
	for i, want := range criteria {
		if got[i] != *want {
			t.Errorf("criterion at index %d = %v, want %v", i, got[i], *want)
		}
	}
}

func TestRubricBuilderRejectsEmptyBuild(t *testing.T) {
	tests := []struct {
		name    string
		builder *domain.RubricBuilder
	}{
		{name: "new builder", builder: domain.NewRubricBuilder()},
		{name: "zero-value builder", builder: &domain.RubricBuilder{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rubric, err := tt.builder.Build()
			if err == nil {
				t.Fatal("Build must reject an empty rubric")
			}
			if rubric != nil {
				t.Error("Build must not return a rubric on error")
			}
		})
	}
}

func TestRubricBuilderRejectsInvalidInputWithoutChangingState(t *testing.T) {
	tests := []struct {
		name      string
		criterion *domain.Criterion
	}{
		{name: "nil criterion", criterion: nil},
		{name: "zero-value criterion", criterion: &domain.Criterion{}},
		{name: "duplicate ID", criterion: newRubricTestCriterion(t, 1, "Задавал вопросы")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := domain.NewRubricBuilder()
			original := newRubricTestCriterion(t, 1, "Ходил на пары")
			if err := builder.AddCriterion(original); err != nil {
				t.Fatal(err)
			}
			if err := builder.AddCriterion(tt.criterion); err == nil {
				t.Fatal("AddCriterion must reject invalid input")
			}
			rubric, err := builder.Build()
			if err != nil {
				t.Fatalf("a failed addition must not prevent building: %v", err)
			}
			got := rubric.Criteria()
			if len(got) != 1 || got[0] != *original {
				t.Errorf("failed addition changed the builder: got %v, want [%v]", got, *original)
			}
		})
	}
}

func TestRubricBuilderCopiesInput(t *testing.T) {
	builder := domain.NewRubricBuilder()
	criterion := newRubricTestCriterion(t, 1, "Ходил на пары")
	want := *criterion
	if err := builder.AddCriterion(criterion); err != nil {
		t.Fatal(err)
	}
	*criterion = *newRubricTestCriterion(t, 2, "Задавал вопросы")
	rubric, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	got := rubric.Criteria()
	if len(got) != 1 || got[0] != want {
		t.Errorf("replacing the input changed the builder: got %v, want [%v]", got, want)
	}
}

func TestRubricBuilderBuildReturnsIndependentRubrics(t *testing.T) {
	builder := domain.NewRubricBuilder()
	originals := []*domain.Criterion{
		newRubricTestCriterion(t, 1, "Ходил на пары"),
		newRubricTestCriterion(t, 2, "Задавал вопросы"),
		newRubricTestCriterion(t, 3, "Не ныл"),
	}
	for _, criterion := range originals {
		if err := builder.AddCriterion(criterion); err != nil {
			t.Fatal(err)
		}
	}
	first, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	second, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	firstExtra := newRubricTestCriterion(t, 4, "Помогал одногруппникам")
	secondExtra := newRubricTestCriterion(t, 5, "Участвовал в обсуждениях")
	builderExtra := newRubricTestCriterion(t, 6, "Приходил вовремя")
	if err := first.AddCriterion(firstExtra); err != nil {
		t.Fatal(err)
	}
	if err := second.AddCriterion(secondExtra); err != nil {
		t.Fatal(err)
	}
	if err := builder.AddCriterion(builderExtra); err != nil {
		t.Fatal(err)
	}
	third, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		rubric *domain.Rubric
		extra  *domain.Criterion
	}{
		{name: "first rubric", rubric: first, extra: firstExtra},
		{name: "second rubric", rubric: second, extra: secondExtra},
		{name: "updated builder", rubric: third, extra: builderExtra},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rubric.Criteria()
			if len(got) != 4 {
				t.Fatalf("got %d criteria, want 4", len(got))
			}
			for i, want := range originals {
				if got[i] != *want {
					t.Errorf("original criterion at index %d changed: got %v, want %v", i, got[i], *want)
				}
			}
			if got[3] != *tt.extra {
				t.Errorf("rubrics share their storage: got extra criterion %v, want %v", got[3], *tt.extra)
			}
		})
	}
}
