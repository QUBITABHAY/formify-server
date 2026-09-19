package response

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const nullJSON = "null"

type QuizFieldResult struct {
	FieldID       string `json:"fieldId"`
	Title         string `json:"title"`
	Correct       bool   `json:"correct"`
	EarnedPoints  int    `json:"earnedPoints"`
	MaxPoints     int    `json:"maxPoints"`
	UserAnswer    any    `json:"userAnswer"`
	CorrectAnswer any    `json:"correctAnswer"`
}

type QuizResult struct {
	Score        int               `json:"score"`
	MaxScore     int               `json:"maxScore"`
	FieldResults []QuizFieldResult `json:"fieldResults"`
}

type rawQuizOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type rawQuizField struct {
	ID                 string          `json:"id"`
	Type               string          `json:"type"`
	Title              string          `json:"title"`
	Name               string          `json:"name"`
	CorrectAnswer      json.RawMessage `json:"correctAnswer"`
	CorrectAnswerSnake json.RawMessage `json:"correct_answer"`
	Points             *float64        `json:"points"`
	Options            []rawQuizOption `json:"options"`
}

func hasCorrectAnswer(f *rawQuizField) bool {
	raw := f.CorrectAnswer
	if len(raw) == 0 || string(raw) == nullJSON {
		raw = f.CorrectAnswerSnake
	}
	if len(raw) == 0 || string(raw) == nullJSON || string(raw) == `""` || string(raw) == `[]` {
		return false
	}
	return true
}

func hasAnyCorrectAnswer(fields []rawQuizField) bool {
	for i := range fields {
		if hasCorrectAnswer(&fields[i]) {
			return true
		}
	}
	return false
}

func isQuizSchema(schemaIsQuiz, schemaIsQuizSnake *bool, fields []rawQuizField) bool {
	if (schemaIsQuiz != nil && *schemaIsQuiz) || (schemaIsQuizSnake != nil && *schemaIsQuizSnake) {
		return true
	}
	if schemaIsQuiz == nil && schemaIsQuizSnake == nil {
		return hasAnyCorrectAnswer(fields)
	}
	return false
}

func parseQuizFields(schemaBytes []byte) (bool, []rawQuizField) {
	if len(schemaBytes) == 0 {
		return false, nil
	}

	var schemaObj struct {
		IsQuiz      *bool          `json:"isQuiz"`
		IsQuizSnake *bool          `json:"is_quiz"`
		Fields      []rawQuizField `json:"fields"`
	}

	if err := json.Unmarshal(schemaBytes, &schemaObj); err != nil {
		return false, nil
	}

	if !isQuizSchema(schemaObj.IsQuiz, schemaObj.IsQuizSnake, schemaObj.Fields) {
		return false, nil
	}

	return true, schemaObj.Fields
}

func findUserAnswer(data map[string]any, f *rawQuizField) (any, bool) {
	if val, ok := data[f.ID]; ok {
		return val, true
	}
	if f.Title != "" {
		if val, ok := data[f.Title]; ok {
			return val, true
		}
	}
	if f.Name != "" {
		if val, ok := data[f.Name]; ok {
			return val, true
		}
	}
	return nil, false
}

func extractStringSlice(val any) ([]string, bool) {
	if val == nil {
		return nil, false
	}
	switch v := val.(type) {
	case []any:
		res := make([]string, len(v))
		for i, item := range v {
			res[i] = strings.TrimSpace(fmt.Sprintf("%v", item))
		}
		return res, true
	case []string:
		res := make([]string, len(v))
		for i, item := range v {
			res[i] = strings.TrimSpace(item)
		}
		return res, true
	default:
		return nil, false
	}
}

func normalizeOptionValue(val string, options []rawQuizOption) string {
	trimmed := strings.TrimSpace(val)
	if len(options) == 0 {
		return strings.ToLower(trimmed)
	}
	for _, opt := range options {
		if strings.EqualFold(opt.Label, trimmed) || strings.EqualFold(opt.Value, trimmed) {
			return strings.ToLower(opt.Value)
		}
	}
	return strings.ToLower(trimmed)
}

func normalizeStringSlice(slice []string, options []rawQuizOption) []string {
	res := make([]string, len(slice))
	for i, s := range slice {
		res[i] = normalizeOptionValue(s, options)
	}
	return res
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := append([]string(nil), a...)
	sortedB := append([]string(nil), b...)
	sort.Strings(sortedA)
	sort.Strings(sortedB)
	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}
	return true
}

func checkSliceAnswer(userAnswer any, correctSlice []string, options []rawQuizOption) bool {
	normCorrect := normalizeStringSlice(correctSlice, options)
	userSlice, isUserSlice := extractStringSlice(userAnswer)
	if isUserSlice {
		normUser := normalizeStringSlice(userSlice, options)
		return slicesEqual(normUser, normCorrect)
	}
	if userAnswer != nil {
		normUser := normalizeStringSlice([]string{fmt.Sprintf("%v", userAnswer)}, options)
		return slicesEqual(normUser, normCorrect)
	}
	return false
}

func checkScalarAnswer(userAnswer, correctAnswer any, options []rawQuizOption) bool {
	if userAnswer == nil {
		return false
	}
	normCorrect := normalizeOptionValue(fmt.Sprintf("%v", correctAnswer), options)
	userSlice, isUserSlice := extractStringSlice(userAnswer)
	if isUserSlice {
		if len(userSlice) == 1 {
			return normalizeOptionValue(userSlice[0], options) == normCorrect
		}
		return false
	}
	normUser := normalizeOptionValue(fmt.Sprintf("%v", userAnswer), options)
	return normUser == normCorrect
}

func checkAnswer(userAnswer, correctAnswer any, options []rawQuizOption) bool {
	if correctAnswer == nil {
		return false
	}
	if correctSlice, isSlice := extractStringSlice(correctAnswer); isSlice {
		return checkSliceAnswer(userAnswer, correctSlice, options)
	}
	return checkScalarAnswer(userAnswer, correctAnswer, options)
}

func parseCorrectAnswer(f *rawQuizField) (any, bool) {
	correctRaw := f.CorrectAnswer
	if len(correctRaw) == 0 || string(correctRaw) == nullJSON {
		correctRaw = f.CorrectAnswerSnake
	}
	if len(correctRaw) == 0 || string(correctRaw) == nullJSON {
		return nil, false
	}

	var correctAnswer any
	if err := json.Unmarshal(correctRaw, &correctAnswer); err != nil {
		return nil, false
	}
	return correctAnswer, true
}

func getFieldPoints(f *rawQuizField) int {
	if f.Points != nil && *f.Points > 0 {
		return int(*f.Points)
	}
	return 1
}

func evaluateSingleField(f *rawQuizField, userAnswers map[string]any) (*QuizFieldResult, int, bool) {
	correctAnswer, ok := parseCorrectAnswer(f)
	if !ok {
		return nil, 0, false
	}

	pts := getFieldPoints(f)
	userAns, found := findUserAnswer(userAnswers, f)
	isCorrect := false
	if found && userAns != nil {
		isCorrect = checkAnswer(userAns, correctAnswer, f.Options)
	}

	earned := 0
	if isCorrect {
		earned = pts
	}

	return &QuizFieldResult{
		FieldID:       f.ID,
		Title:         f.Title,
		Correct:       isCorrect,
		EarnedPoints:  earned,
		MaxPoints:     pts,
		UserAnswer:    userAns,
		CorrectAnswer: correctAnswer,
	}, pts, true
}

func evaluateQuiz(schemaBytes, reqData []byte) *QuizResult {
	isQuiz, fields := parseQuizFields(schemaBytes)
	if !isQuiz {
		return nil
	}

	var userAnswers map[string]any
	if len(reqData) > 0 {
		_ = json.Unmarshal(reqData, &userAnswers)
	}
	if userAnswers == nil {
		userAnswers = make(map[string]any)
	}

	totalScore := 0
	maxScore := 0
	fieldResults := make([]QuizFieldResult, 0, len(fields))

	for i := range fields {
		res, pts, scorable := evaluateSingleField(&fields[i], userAnswers)
		if !scorable {
			continue
		}
		maxScore += pts
		totalScore += res.EarnedPoints
		fieldResults = append(fieldResults, *res)
	}

	return &QuizResult{
		Score:        totalScore,
		MaxScore:     maxScore,
		FieldResults: fieldResults,
	}
}

func attachQuizResultToMeta(rawMeta json.RawMessage, quizResult *QuizResult) json.RawMessage {
	if quizResult == nil {
		return rawMeta
	}
	metaMap := make(map[string]any)
	if len(rawMeta) > 0 && string(rawMeta) != nullJSON {
		_ = json.Unmarshal(rawMeta, &metaMap)
	}
	metaMap["quiz_result"] = quizResult
	encoded, err := json.Marshal(metaMap)
	if err != nil {
		return rawMeta
	}
	return json.RawMessage(encoded)
}

func extractQuizResultFromMeta(rawMeta json.RawMessage) *QuizResult {
	if len(rawMeta) == 0 || string(rawMeta) == nullJSON {
		return nil
	}
	var metaMap struct {
		QuizResult *QuizResult `json:"quiz_result"`
	}
	if err := json.Unmarshal(rawMeta, &metaMap); err != nil {
		return nil
	}
	return metaMap.QuizResult
}
