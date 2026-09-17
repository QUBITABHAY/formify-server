package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

func TestPgtypeTextToString(t *testing.T) {
	tests := []struct {
		name     string
		input    pgtype.Text
		wantNil  bool
		expected string
	}{
		{
			name:    "invalid text returns nil",
			input:   pgtype.Text{Valid: false},
			wantNil: true,
		},
		{
			name:     "valid text returns pointer to string",
			input:    pgtype.Text{String: "hello", Valid: true},
			wantNil:  false,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PgtypeTextToString(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", *got)
				}
			} else {
				if got == nil || *got != tt.expected {
					t.Fatalf("expected %q, got %v", tt.expected, got)
				}
			}
		})
	}
}

func TestStringToPgtypeText(t *testing.T) {
	t.Run("nil string returns invalid pgtype.Text", func(t *testing.T) {
		got := StringToPgtypeText(nil)
		if got.Valid {
			t.Fatal("expected Valid to be false")
		}
	})

	t.Run("non-nil string returns valid pgtype.Text", func(t *testing.T) {
		s := "test string"
		got := StringToPgtypeText(&s)
		if !got.Valid {
			t.Fatal("expected Valid to be true")
		}
		if got.String != s {
			t.Fatalf("expected %q, got %q", s, got.String)
		}
	})
}

func TestPgtypeTimestamptzToTime(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)

	t.Run("invalid timestamptz returns zero time", func(t *testing.T) {
		got := PgtypeTimestamptzToTime(pgtype.Timestamptz{Valid: false})
		if !got.IsZero() {
			t.Fatalf("expected zero time, got %v", got)
		}
	})

	t.Run("valid timestamptz returns time", func(t *testing.T) {
		got := PgtypeTimestamptzToTime(pgtype.Timestamptz{Time: now, Valid: true})
		if !got.Equal(now) {
			t.Fatalf("expected %v, got %v", now, got)
		}
	})
}

func TestPgtypeTimestamptzToTimePtr(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)

	t.Run("invalid timestamptz returns nil", func(t *testing.T) {
		got := PgtypeTimestamptzToTimePtr(pgtype.Timestamptz{Valid: false})
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("valid timestamptz returns pointer to time", func(t *testing.T) {
		got := PgtypeTimestamptzToTimePtr(pgtype.Timestamptz{Time: now, Valid: true})
		if got == nil || !got.Equal(now) {
			t.Fatalf("expected %v, got %v", now, got)
		}
	})
}

func TestPgtypeBoolToBool(t *testing.T) {
	t.Run("invalid bool returns false", func(t *testing.T) {
		if got := PgtypeBoolToBool(pgtype.Bool{Valid: false}); got {
			t.Fatalf("expected false, got %v", got)
		}
	})

	t.Run("valid true returns true", func(t *testing.T) {
		if got := PgtypeBoolToBool(pgtype.Bool{Bool: true, Valid: true}); !got {
			t.Fatalf("expected true, got %v", got)
		}
	})

	t.Run("valid false returns false", func(t *testing.T) {
		if got := PgtypeBoolToBool(pgtype.Bool{Bool: false, Valid: true}); got {
			t.Fatalf("expected false, got %v", got)
		}
	})
}

func TestBoolToPgtypeBool(t *testing.T) {
	gotTrue := BoolToPgtypeBool(true)
	if !gotTrue.Valid || !gotTrue.Bool {
		t.Fatalf("expected Valid: true, Bool: true, got %+v", gotTrue)
	}

	gotFalse := BoolToPgtypeBool(false)
	if !gotFalse.Valid || gotFalse.Bool {
		t.Fatalf("expected Valid: true, Bool: false, got %+v", gotFalse)
	}
}

func TestTimeToPgtypeTimestamptz(t *testing.T) {
	t.Run("nil time returns invalid timestamptz", func(t *testing.T) {
		got := TimeToPgtypeTimestamptz(nil)
		if got.Valid {
			t.Fatal("expected Valid to be false")
		}
	})

	t.Run("non-nil time returns valid timestamptz", func(t *testing.T) {
		now := time.Now()
		got := TimeToPgtypeTimestamptz(&now)
		if !got.Valid {
			t.Fatal("expected Valid to be true")
		}
		if !got.Time.Equal(now) {
			t.Fatalf("expected %v, got %v", now, got.Time)
		}
	})
}

func TestGetAuthUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Run("returns false when user_id missing", func(t *testing.T) {
		id, ok := GetAuthUserID(c)
		if ok || id != 0 {
			t.Fatalf("expected (0, false), got (%d, %v)", id, ok)
		}
	})

	t.Run("returns false when user_id is not float64", func(t *testing.T) {
		c.Set("user_id", "123")
		id, ok := GetAuthUserID(c)
		if ok || id != 0 {
			t.Fatalf("expected (0, false), got (%d, %v)", id, ok)
		}
	})

	t.Run("returns userID when user_id is float64", func(t *testing.T) {
		c.Set("user_id", float64(42))
		id, ok := GetAuthUserID(c)
		if !ok || id != 42 {
			t.Fatalf("expected (42, true), got (%d, %v)", id, ok)
		}
	})
}

func TestGenerateShareURL(t *testing.T) {
	token, err := GenerateShareURL(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	token2, err := GenerateShareURL(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == token2 {
		t.Fatal("expected distinct tokens to be generated")
	}
}
