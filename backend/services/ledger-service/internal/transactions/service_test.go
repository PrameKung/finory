package transactions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	testUserID        = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	testTransactionID = "23876640-e8aa-4d90-95d7-3c9f61ecf0e3"
	testCategoryID    = "828b1e18-f55e-45c4-b758-3580b38d94eb"
	testWalletID      = "b46ac984-6b0a-4b2f-b08a-3660b97a0f29"
)

type fakeRepository struct {
	createFn func(context.Context, string, CreateParams) (Transaction, error)
	listFn   func(context.Context, string, ListParams) ([]Transaction, error)
	getFn    func(context.Context, string, string) (Transaction, error)
	updateFn func(context.Context, string, string, UpdateParams) (Transaction, error)
	deleteFn func(context.Context, string, string) (bool, error)
}

func (r *fakeRepository) Create(ctx context.Context, userID string, params CreateParams) (Transaction, error) {
	return r.createFn(ctx, userID, params)
}

func (r *fakeRepository) List(ctx context.Context, userID string, params ListParams) ([]Transaction, error) {
	return r.listFn(ctx, userID, params)
}

func (r *fakeRepository) Get(ctx context.Context, userID, transactionID string) (Transaction, error) {
	return r.getFn(ctx, userID, transactionID)
}

func (r *fakeRepository) Update(ctx context.Context, userID, transactionID string, params UpdateParams) (Transaction, error) {
	return r.updateFn(ctx, userID, transactionID, params)
}

func (r *fakeRepository) Delete(ctx context.Context, userID, transactionID string) (bool, error) {
	return r.deleteFn(ctx, userID, transactionID)
}

func TestCreateNormalizesValidatedInput(t *testing.T) {
	var got CreateParams
	repository := &fakeRepository{createFn: func(_ context.Context, userID string, params CreateParams) (Transaction, error) {
		if userID != testUserID {
			t.Fatalf("user ID = %q, want %q", userID, testUserID)
		}
		got = params
		return Transaction{ID: testTransactionID}, nil
	}}
	service := NewService(repository)

	transaction, err := service.Create(context.Background(), testUserID, CreateInput{
		CategoryID: testCategoryID, WalletID: testWalletID, Type: " Expense ",
		Amount: " 1250.5000 ", Description: " Lunch ", TransactionDate: "2026-09-26",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transaction.ID != testTransactionID || got.Type != "expense" || got.Amount != "1250.5000" || got.Description != "Lunch" {
		t.Fatalf("unexpected transaction or params: transaction=%+v params=%+v", transaction, got)
	}
	if got.TransactionDate.Format(time.DateOnly) != "2026-09-26" {
		t.Fatalf("transaction date = %s", got.TransactionDate.Format(time.DateOnly))
	}
}

func TestCreateRejectsInvalidValues(t *testing.T) {
	valid := CreateInput{
		CategoryID: testCategoryID, WalletID: testWalletID, Type: "expense",
		Amount: "10.25", TransactionDate: "2026-09-26",
	}
	for _, tc := range []struct {
		name   string
		change func(*CreateInput)
	}{
		{name: "category ID", change: func(input *CreateInput) { input.CategoryID = "other-user-category" }},
		{name: "wallet ID", change: func(input *CreateInput) { input.WalletID = "other-user-wallet" }},
		{name: "type", change: func(input *CreateInput) { input.Type = "transfer" }},
		{name: "zero amount", change: func(input *CreateInput) { input.Amount = "0.0000" }},
		{name: "floating precision", change: func(input *CreateInput) { input.Amount = "10.12345" }},
		{name: "date", change: func(input *CreateInput) { input.TransactionDate = "26/09/2026" }},
		{name: "description", change: func(input *CreateInput) { input.Description = string(make([]rune, 501)) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			service := NewService(&fakeRepository{createFn: func(context.Context, string, CreateParams) (Transaction, error) {
				t.Fatal("repository called for invalid input")
				return Transaction{}, nil
			}})
			if _, err := service.Create(context.Background(), testUserID, input); !errors.Is(err, ErrInvalidTransaction) {
				t.Fatalf("error = %v, want ErrInvalidTransaction", err)
			}
		})
	}
}

func TestCreateRejectsUnownedReferences(t *testing.T) {
	service := NewService(&fakeRepository{createFn: func(context.Context, string, CreateParams) (Transaction, error) {
		return Transaction{}, pgx.ErrNoRows
	}})
	_, err := service.Create(context.Background(), testUserID, CreateInput{
		CategoryID: testCategoryID, WalletID: testWalletID, Type: "income",
		Amount: "100", TransactionDate: "2026-09-26",
	})
	if !errors.Is(err, ErrInvalidTransactionReference) {
		t.Fatalf("error = %v, want ErrInvalidTransactionReference", err)
	}
}

func TestListParsesMonthAndTypeFilters(t *testing.T) {
	var got ListParams
	service := NewService(&fakeRepository{listFn: func(_ context.Context, userID string, params ListParams) ([]Transaction, error) {
		if userID != testUserID {
			t.Fatalf("user ID = %q, want %q", userID, testUserID)
		}
		got = params
		return []Transaction{}, nil
	}})
	if _, err := service.List(context.Background(), testUserID, ListFilter{Month: "2026-09", Type: " EXPENSE "}); err != nil {
		t.Fatal(err)
	}
	if got.MonthStart == nil || got.MonthStart.Format("2006-01") != "2026-09" || got.Type == nil || *got.Type != "expense" {
		t.Fatalf("filters = %+v", got)
	}
	for _, filter := range []ListFilter{{Month: "2026-13"}, {Type: "transfer"}} {
		if _, err := service.List(context.Background(), testUserID, filter); !errors.Is(err, ErrInvalidTransaction) {
			t.Fatalf("filter %+v error = %v, want ErrInvalidTransaction", filter, err)
		}
	}
}

func TestGetHidesOtherUsersTransactions(t *testing.T) {
	service := NewService(&fakeRepository{getFn: func(_ context.Context, userID, transactionID string) (Transaction, error) {
		if userID != testUserID || transactionID != testTransactionID {
			t.Fatalf("unexpected lookup: user=%q transaction=%q", userID, transactionID)
		}
		return Transaction{}, pgx.ErrNoRows
	}})
	if _, err := service.Get(context.Background(), testUserID, testTransactionID); !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("error = %v, want ErrTransactionNotFound", err)
	}
}

func TestUpdateDistinguishesInvalidReferenceFromNotFound(t *testing.T) {
	typeValue := "income"
	for _, tc := range []struct {
		name    string
		getErr  error
		wantErr error
	}{
		{name: "invalid reference", wantErr: ErrInvalidTransactionReference},
		{name: "not found", getErr: pgx.ErrNoRows, wantErr: ErrTransactionNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(&fakeRepository{
				updateFn: func(context.Context, string, string, UpdateParams) (Transaction, error) {
					return Transaction{}, pgx.ErrNoRows
				},
				getFn: func(context.Context, string, string) (Transaction, error) {
					return Transaction{ID: testTransactionID}, tc.getErr
				},
			})
			if _, err := service.Update(context.Background(), testUserID, testTransactionID, UpdateInput{Type: &typeValue}); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteMapsMissingTransaction(t *testing.T) {
	service := NewService(&fakeRepository{deleteFn: func(_ context.Context, userID, transactionID string) (bool, error) {
		if userID != testUserID || transactionID != testTransactionID {
			t.Fatalf("unexpected delete: user=%q transaction=%q", userID, transactionID)
		}
		return false, nil
	}})
	if err := service.Delete(context.Background(), testUserID, testTransactionID); !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("error = %v, want ErrTransactionNotFound", err)
	}
}
