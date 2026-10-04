package budgets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	testUserID     = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	testBudgetID   = "23876640-e8aa-4d90-95d7-3c9f61ecf0e3"
	testCategoryID = "828b1e18-f55e-45c4-b758-3580b38d94eb"
)

type fakeRepository struct {
	createFn func(context.Context, string, CreateParams) (Budget, error)
	listFn   func(context.Context, string, ListParams) ([]Budget, error)
	getFn    func(context.Context, string, string) (Budget, error)
	updateFn func(context.Context, string, string, UpdateParams) (Budget, error)
	deleteFn func(context.Context, string, string) (bool, error)
}

func (r *fakeRepository) Create(ctx context.Context, userID string, params CreateParams) (Budget, error) {
	return r.createFn(ctx, userID, params)
}

func (r *fakeRepository) List(ctx context.Context, userID string, params ListParams) ([]Budget, error) {
	return r.listFn(ctx, userID, params)
}

func (r *fakeRepository) Get(ctx context.Context, userID, budgetID string) (Budget, error) {
	return r.getFn(ctx, userID, budgetID)
}

func (r *fakeRepository) Update(ctx context.Context, userID, budgetID string, params UpdateParams) (Budget, error) {
	return r.updateFn(ctx, userID, budgetID, params)
}

func (r *fakeRepository) Delete(ctx context.Context, userID, budgetID string) (bool, error) {
	return r.deleteFn(ctx, userID, budgetID)
}

func TestCreateNormalizesValidatedInput(t *testing.T) {
	var got CreateParams
	repository := &fakeRepository{createFn: func(_ context.Context, userID string, params CreateParams) (Budget, error) {
		if userID != testUserID {
			t.Fatalf("user ID = %q, want %q", userID, testUserID)
		}
		got = params
		return Budget{ID: testBudgetID}, nil
	}}
	service := NewService(repository)

	budget, err := service.Create(context.Background(), testUserID, CreateInput{
		CategoryID: " " + testCategoryID + " ", Amount: " 1250.5000 ", Month: " 2026-09 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if budget.ID != testBudgetID || got.CategoryID != testCategoryID || got.Amount != "1250.5000" {
		t.Fatalf("unexpected budget or params: budget=%+v params=%+v", budget, got)
	}
	if got.MonthStart.Format("2006-01") != "2026-09" || got.MonthStart.Day() != 1 {
		t.Fatalf("month start = %s", got.MonthStart.Format(time.DateOnly))
	}
}

func TestCreateRejectsInvalidValues(t *testing.T) {
	valid := CreateInput{CategoryID: testCategoryID, Amount: "10.25", Month: "2026-09"}
	for _, tc := range []struct {
		name   string
		change func(*CreateInput)
	}{
		{name: "category ID", change: func(input *CreateInput) { input.CategoryID = "other-user-category" }},
		{name: "zero amount", change: func(input *CreateInput) { input.Amount = "0.0000" }},
		{name: "negative amount", change: func(input *CreateInput) { input.Amount = "-1" }},
		{name: "floating precision", change: func(input *CreateInput) { input.Amount = "10.12345" }},
		{name: "month", change: func(input *CreateInput) { input.Month = "2026-13" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			service := NewService(&fakeRepository{createFn: func(context.Context, string, CreateParams) (Budget, error) {
				t.Fatal("repository called for invalid input")
				return Budget{}, nil
			}})
			if _, err := service.Create(context.Background(), testUserID, input); !errors.Is(err, ErrInvalidBudget) {
				t.Fatalf("error = %v, want ErrInvalidBudget", err)
			}
		})
	}
}

func TestCreateMapsReferenceAndConflictErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "invalid category", repoErr: pgx.ErrNoRows, wantErr: ErrInvalidBudgetReference},
		{name: "duplicate month", repoErr: &pgconn.PgError{Code: "23505"}, wantErr: ErrBudgetConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(&fakeRepository{createFn: func(context.Context, string, CreateParams) (Budget, error) {
				return Budget{}, tc.repoErr
			}})
			_, err := service.Create(context.Background(), testUserID, CreateInput{
				CategoryID: testCategoryID, Amount: "100", Month: "2026-09",
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestListParsesMonthFilter(t *testing.T) {
	var got ListParams
	service := NewService(&fakeRepository{listFn: func(_ context.Context, userID string, params ListParams) ([]Budget, error) {
		if userID != testUserID {
			t.Fatalf("user ID = %q, want %q", userID, testUserID)
		}
		got = params
		return []Budget{}, nil
	}})
	if _, err := service.List(context.Background(), testUserID, ListFilter{Month: " 2026-09 "}); err != nil {
		t.Fatal(err)
	}
	if got.MonthStart == nil || got.MonthStart.Format("2006-01") != "2026-09" {
		t.Fatalf("filter = %+v", got)
	}
	if _, err := service.List(context.Background(), testUserID, ListFilter{Month: "09-2026"}); !errors.Is(err, ErrInvalidBudget) {
		t.Fatalf("error = %v, want ErrInvalidBudget", err)
	}
}

func TestUpdateDistinguishesInvalidReferenceFromNotFound(t *testing.T) {
	categoryID := testCategoryID
	for _, tc := range []struct {
		name    string
		getErr  error
		wantErr error
	}{
		{name: "invalid reference", wantErr: ErrInvalidBudgetReference},
		{name: "not found", getErr: pgx.ErrNoRows, wantErr: ErrBudgetNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(&fakeRepository{
				updateFn: func(context.Context, string, string, UpdateParams) (Budget, error) {
					return Budget{}, pgx.ErrNoRows
				},
				getFn: func(context.Context, string, string) (Budget, error) {
					return Budget{ID: testBudgetID}, tc.getErr
				},
			})
			if _, err := service.Update(context.Background(), testUserID, testBudgetID, UpdateInput{CategoryID: &categoryID}); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteMapsMissingBudget(t *testing.T) {
	service := NewService(&fakeRepository{deleteFn: func(_ context.Context, userID, budgetID string) (bool, error) {
		if userID != testUserID || budgetID != testBudgetID {
			t.Fatalf("unexpected delete: user=%q budget=%q", userID, budgetID)
		}
		return false, nil
	}})
	if err := service.Delete(context.Background(), testUserID, testBudgetID); !errors.Is(err, ErrBudgetNotFound) {
		t.Fatalf("error = %v, want ErrBudgetNotFound", err)
	}
}
