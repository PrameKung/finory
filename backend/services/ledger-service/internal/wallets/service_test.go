package wallets

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	testUserID   = "60000000-0000-4000-8000-000000000001"
	testWalletID = "80000000-0000-4000-8000-000000000001"
)

type fakeRepository struct {
	createDefaultCashFn func(context.Context, string) (int64, error)
	listFn              func(context.Context, string) ([]Wallet, error)
	createFn            func(context.Context, string, CreateParams) (Wallet, error)
	updateFn            func(context.Context, string, string, UpdateParams) (Wallet, error)
	deleteFn            func(context.Context, string, string) (bool, error)
}

func (r *fakeRepository) CreateDefaultCash(ctx context.Context, userID string) (int64, error) {
	return r.createDefaultCashFn(ctx, userID)
}

func (r *fakeRepository) List(ctx context.Context, userID string) ([]Wallet, error) {
	return r.listFn(ctx, userID)
}

func (r *fakeRepository) Create(ctx context.Context, userID string, params CreateParams) (Wallet, error) {
	return r.createFn(ctx, userID, params)
}

func (r *fakeRepository) Update(ctx context.Context, userID, walletID string, params UpdateParams) (Wallet, error) {
	return r.updateFn(ctx, userID, walletID, params)
}

func (r *fakeRepository) Delete(ctx context.Context, userID, walletID string) (bool, error) {
	return r.deleteFn(ctx, userID, walletID)
}

func TestListInitializesDefaultCashWalletBeforeListing(t *testing.T) {
	calls := make([]string, 0, 2)
	repository := &fakeRepository{
		createDefaultCashFn: func(_ context.Context, userID string) (int64, error) {
			calls = append(calls, "default")
			if userID != testUserID {
				t.Fatalf("default user ID = %q", userID)
			}
			return 1, nil
		},
		listFn: func(_ context.Context, userID string) ([]Wallet, error) {
			calls = append(calls, "list")
			if userID != testUserID {
				t.Fatalf("list user ID = %q", userID)
			}
			return []Wallet{{ID: testWalletID, IsDefault: true}}, nil
		},
	}

	items, err := NewService(repository).List(context.Background(), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != testWalletID || strings.Join(calls, ",") != "default,list" {
		t.Fatalf("items=%+v calls=%v", items, calls)
	}
}

func TestCreateWalletNormalizesInputAndPreservesDecimalString(t *testing.T) {
	calls := make([]string, 0, 2)
	var got CreateParams
	repository := &fakeRepository{
		createDefaultCashFn: func(context.Context, string) (int64, error) {
			calls = append(calls, "default")
			return 0, nil
		},
		createFn: func(_ context.Context, userID string, params CreateParams) (Wallet, error) {
			calls = append(calls, "create")
			if userID != testUserID {
				t.Fatalf("user ID = %q", userID)
			}
			got = params
			return Wallet{ID: testWalletID}, nil
		},
	}

	wallet, err := NewService(repository).Create(context.Background(), testUserID, CreateInput{
		Name: " Checking ", Type: " BANK ", Balance: " -1250.5000 ", CurrencyCode: " thb ",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := CreateParams{Name: "Checking", Type: "bank", Balance: "-1250.5000", CurrencyCode: "THB"}
	if wallet.ID != testWalletID || got != want || strings.Join(calls, ",") != "default,create" {
		t.Fatalf("wallet=%+v params=%+v calls=%v", wallet, got, calls)
	}
}

func TestCreateWalletDefaultsEmptyBalanceToZero(t *testing.T) {
	var gotBalance string
	repository := &fakeRepository{
		createDefaultCashFn: func(context.Context, string) (int64, error) { return 0, nil },
		createFn: func(_ context.Context, _ string, params CreateParams) (Wallet, error) {
			gotBalance = params.Balance
			return Wallet{ID: testWalletID}, nil
		},
	}
	if _, err := NewService(repository).Create(context.Background(), testUserID, CreateInput{Name: "Cash 2", Type: "cash", CurrencyCode: "THB"}); err != nil {
		t.Fatal(err)
	}
	if gotBalance != "0" {
		t.Fatalf("balance = %q, want zero", gotBalance)
	}
}

func TestCreateWalletRejectsInvalidValues(t *testing.T) {
	valid := CreateInput{Name: "Checking", Type: "bank", Balance: "100.00", CurrencyCode: "THB"}
	for _, tc := range []struct {
		name   string
		change func(*CreateInput)
	}{
		{name: "blank name", change: func(input *CreateInput) { input.Name = "  " }},
		{name: "long name", change: func(input *CreateInput) { input.Name = strings.Repeat("a", 101) }},
		{name: "invalid type", change: func(input *CreateInput) { input.Type = "credit" }},
		{name: "leading zero balance", change: func(input *CreateInput) { input.Balance = "01.00" }},
		{name: "excess precision", change: func(input *CreateInput) { input.Balance = "1.00001" }},
		{name: "too many integer digits", change: func(input *CreateInput) { input.Balance = "1234567890123456" }},
		{name: "invalid currency", change: func(input *CreateInput) { input.CurrencyCode = "BAHT" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			repository := &fakeRepository{
				createDefaultCashFn: func(context.Context, string) (int64, error) {
					t.Fatal("default wallet created for invalid input")
					return 0, nil
				},
				createFn: func(context.Context, string, CreateParams) (Wallet, error) {
					t.Fatal("repository called for invalid wallet")
					return Wallet{}, nil
				},
			}
			if _, err := NewService(repository).Create(context.Background(), testUserID, input); !errors.Is(err, ErrInvalidWallet) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidWallet)
			}
		})
	}
}

func TestUpdateWalletNormalizesInput(t *testing.T) {
	name, walletType, balance, currency := " Savings ", " BANK ", " -20.2500 ", " usd "
	var got UpdateParams
	repository := &fakeRepository{updateFn: func(_ context.Context, userID, walletID string, params UpdateParams) (Wallet, error) {
		if userID != testUserID || walletID != testWalletID {
			t.Fatalf("update target = %q/%q", userID, walletID)
		}
		got = params
		return Wallet{ID: walletID}, nil
	}}

	_, err := NewService(repository).Update(context.Background(), testUserID, testWalletID, UpdateInput{
		Name: &name, Type: &walletType, Balance: &balance, CurrencyCode: &currency,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name == nil || *got.Name != "Savings" || got.Type == nil || *got.Type != "bank" ||
		got.Balance == nil || *got.Balance != "-20.2500" || got.CurrencyCode == nil || *got.CurrencyCode != "USD" {
		t.Fatalf("update params = %+v", got)
	}
}

func TestWalletUniquenessConflictsAreStable(t *testing.T) {
	duplicate := &pgconn.PgError{Code: "23505"}
	for _, tc := range []struct {
		name string
		run  func(*Service) error
	}{
		{name: "create", run: func(service *Service) error {
			_, err := service.Create(context.Background(), testUserID, CreateInput{Name: "Checking", Type: "bank", CurrencyCode: "THB"})
			return err
		}},
		{name: "update", run: func(service *Service) error {
			name := "Checking"
			_, err := service.Update(context.Background(), testUserID, testWalletID, UpdateInput{Name: &name})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := &fakeRepository{
				createDefaultCashFn: func(context.Context, string) (int64, error) { return 0, nil },
				createFn:            func(context.Context, string, CreateParams) (Wallet, error) { return Wallet{}, duplicate },
				updateFn:            func(context.Context, string, string, UpdateParams) (Wallet, error) { return Wallet{}, duplicate },
			}
			if err := tc.run(NewService(repository)); !errors.Is(err, ErrWalletConflict) {
				t.Fatalf("error = %v, want %v", err, ErrWalletConflict)
			}
		})
	}
}

func TestDefaultWalletCannotBeModified(t *testing.T) {
	repository := &fakeRepository{
		updateFn: func(context.Context, string, string, UpdateParams) (Wallet, error) {
			return Wallet{}, pgx.ErrNoRows
		},
		deleteFn: func(context.Context, string, string) (bool, error) { return false, nil },
	}
	service := NewService(repository)
	name := "Changed"
	if _, err := service.Update(context.Background(), testUserID, testWalletID, UpdateInput{Name: &name}); !errors.Is(err, ErrWalletNotFound) {
		t.Fatalf("update error = %v, want %v", err, ErrWalletNotFound)
	}
	if err := service.Delete(context.Background(), testUserID, testWalletID); !errors.Is(err, ErrWalletNotFound) {
		t.Fatalf("delete error = %v, want %v", err, ErrWalletNotFound)
	}
}

func TestUpdateWalletRequiresAValidChange(t *testing.T) {
	invalidBalance := "1.00001"
	for _, input := range []UpdateInput{{}, {Balance: &invalidBalance}} {
		repository := &fakeRepository{updateFn: func(context.Context, string, string, UpdateParams) (Wallet, error) {
			t.Fatal("repository called for invalid update")
			return Wallet{}, nil
		}}
		if _, err := NewService(repository).Update(context.Background(), testUserID, testWalletID, input); !errors.Is(err, ErrInvalidWallet) {
			t.Fatalf("input=%+v error=%v, want %v", input, err, ErrInvalidWallet)
		}
	}
}
