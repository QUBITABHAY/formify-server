package response

import (
	"encoding/json"
	"testing"
)

func TestEvaluateQuiz_NonQuiz(t *testing.T) {
	schema := []byte(`{"type":"single","isQuiz":false,"fields":[{"id":"q1","title":"Q1","correctAnswer":"A"}]}`)
	result := evaluateQuiz(schema, []byte(`{"q1":"A"}`))
	if result != nil {
		t.Fatalf("expected nil for non-quiz, got: %+v", result)
	}

	if res := evaluateQuiz(nil, []byte(`{}`)); res != nil {
		t.Fatalf("expected nil for nil schema")
	}
	if res := evaluateQuiz([]byte(`invalid json`), []byte(`{}`)); res != nil {
		t.Fatalf("expected nil for invalid json")
	}
}

func assertFieldResult(t *testing.T, res *QuizFieldResult, expectedCorrect bool, expectedPoints int) {
	t.Helper()
	if res.Correct != expectedCorrect || res.EarnedPoints != expectedPoints {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestEvaluateQuiz_Scoring(t *testing.T) {
	schema := []byte(`{
		"type": "single",
		"isQuiz": true,
		"fields": [
			{
				"id": "q1",
				"title": "What is 2+2?",
				"correctAnswer": "4",
				"points": 2
			},
			{
				"id": "q2",
				"title": "Select prime numbers",
				"correctAnswer": ["2", "3"],
				"points": 3
			},
			{
				"id": "q3",
				"title": "Capital of Spain",
				"correctAnswer": "Madrid"
			},
			{
				"id": "q4",
				"title": "Wrong answer question",
				"correctAnswer": "Yes",
				"points": 1
			},
			{
				"id": "q5",
				"title": "Unanswered question",
				"correctAnswer": "Blue",
				"points": 2
			}
		]
	}`)

	answers := []byte(`{
		"q1": "4",
		"Select prime numbers": ["3", "2"],
		"Capital of Spain": "Madrid",
		"q4": "No"
	}`)

	result := evaluateQuiz(schema, answers)
	if result == nil {
		t.Fatalf("expected non-nil QuizResult")
	}

	if result.Score != 6 || result.MaxScore != 9 || len(result.FieldResults) != 5 {
		t.Fatalf("expected score 6/9 and 5 field results, got: %+v", result)
	}

	assertFieldResult(t, &result.FieldResults[0], true, 2)
	assertFieldResult(t, &result.FieldResults[1], true, 3)
	assertFieldResult(t, &result.FieldResults[3], false, 0)
	assertFieldResult(t, &result.FieldResults[4], false, 0)
}

func TestEvaluateQuiz_SnakeCase(t *testing.T) {
	schema := []byte(`{
		"is_quiz": true,
		"fields": [
			{
				"id": "q1",
				"title": "Q1",
				"correct_answer": "Option A",
				"points": 5
			}
		]
	}`)
	answers := []byte(`{"q1": "Option A"}`)

	result := evaluateQuiz(schema, answers)
	if result == nil || result.Score != 5 || result.MaxScore != 5 {
		t.Fatalf("expected 5/5 score, got: %+v", result)
	}
}

func TestAttachAndExtractQuizResultMeta(t *testing.T) {
	qr := &QuizResult{
		Score:    10,
		MaxScore: 10,
		FieldResults: []QuizFieldResult{
			{FieldID: "q1", Title: "Q1", Correct: true, EarnedPoints: 10, MaxPoints: 10, UserAnswer: "A", CorrectAnswer: "A"},
		},
	}

	attached := attachQuizResultToMeta(nil, qr)
	extracted := extractQuizResultFromMeta(attached)
	if extracted == nil || extracted.Score != 10 {
		t.Fatalf("expected extracted score 10, got: %+v", extracted)
	}

	existing := json.RawMessage(`{"source":"embed"}`)
	attachedExisting := attachQuizResultToMeta(existing, qr)
	extractedExisting := extractQuizResultFromMeta(attachedExisting)
	if extractedExisting == nil || extractedExisting.Score != 10 {
		t.Fatalf("expected extracted score 10, got: %+v", extractedExisting)
	}

	if res := extractQuizResultFromMeta(nil); res != nil {
		t.Fatalf("expected nil for nil meta")
	}
	if res := extractQuizResultFromMeta([]byte(`{}`)); res != nil {
		t.Fatalf("expected nil for empty meta")
	}
}

func TestEvaluateQuiz_OptionsAndAutoDetect(t *testing.T) {
	schema := []byte(`{
		"fields": [
			{
				"id": "q1",
				"title": "Select country",
				"type": "radio",
				"options": [
					{"label": "France", "value": "fr"},
					{"label": "Germany", "value": "de"}
				],
				"correctAnswer": "fr",
				"points": 5
			},
			{
				"id": "q2",
				"title": "Multi select",
				"type": "radio",
				"options": [
					{"label": "Apple", "value": "app"},
					{"label": "Banana", "value": "ban"}
				],
				"correctAnswer": ["app", "ban"],
				"points": 5
			}
		]
	}`)

	answers := []byte(`{
		"q1": "France",
		"q2": ["Banana", "Apple"]
	}`)

	result := evaluateQuiz(schema, answers)
	if result == nil {
		t.Fatalf("expected auto-detected quiz result")
	}
	if result.Score != 10 || result.MaxScore != 10 {
		t.Fatalf("expected 10/10 score, got: %+v", result)
	}
}
