package categories

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	testUserID     = "60000000-0000-4000-8000-000000000001"
	testCategoryID = "70000000-0000-4000-8000-000000000001"
)

type fakeRepository struct {
	createDefaultsFn func(context.Context, string) (int64, error)
	listFn           func(context.Context, string) ([]Category, error)
	createFn         func(context.Context, string, CreateParams) (Category, error)
	updateFn         func(context.Context, string, string, UpdateParams) (Category, error)
	deleteFn         func(context.Context, string, string) (bool, error)
}

func (r *fakeRepository) CreateDefaults(ctx context.Context, userID string) (int64, error) {
	return r.createDefaultsFn(ctx, userID)
}

func (r *fakeRepository) List(ctx context.Context, userID string) ([]Category, error) {
	return r.listFn(ctx, userID)
}

func (r *fakeRepository) Create(ctx context.Context, userID string, params CreateParams) (Category, error) {
	return r.createFn(ctx, userID, params)
}

func (r *fakeRepository) Update(ctx context.Context, userID, categoryID string, params UpdateParams) (Category, error) {
	return r.updateFn(ctx, userID, categoryID, params)
}

func (r *fakeRepository) Delete(ctx context.Context, userID, categoryID string) (bool, error) {
	return r.deleteFn(ctx, userID, categoryID)
}

func TestListInitializesDefaultCategoriesBeforeListing(t *testing.T) {
	calls := make([]string, 0, 2)
	repository := &fakeRepository{
		createDefaultsFn: func(_ context.Context, userID string) (int64, error) {
			calls = append(calls, "defaults")
			if userID != testUserID {
				t.Fatalf("default user ID = %q", userID)
			}
			return 12, nil
		},
		listFn: func(_ context.Context, userID string) ([]Category, error) {
			calls = append(calls, "list")
			if userID != testUserID {
				t.Fatalf("list user ID = %q", userID)
			}
			return []Category{{ID: testCategoryID, IsDefault: true}}, nil
		},
	}

	items, err := NewService(repository).List(context.Background(), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != testCategoryID || strings.Join(calls, ",") != "defaults,list" {
		t.Fatalf("items=%+v calls=%v", items, calls)
	}
}

func TestCreateCategoryNormalizesInput(t *testing.T) {
	var got CreateParams
	repository := &fakeRepository{createFn: func(_ context.Context, userID string, params CreateParams) (Category, error) {
		if userID != testUserID {
			t.Fatalf("user ID = %q", userID)
		}
		got = params
		return Category{ID: testCategoryID}, nil
	}}

	category, err := NewService(repository).Create(context.Background(), testUserID, CreateInput{
		Name: " Dining ", Type: " EXPENSE ", Icon: " utensils ", Color: " #EA580C ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if category.ID != testCategoryID || got != (CreateParams{Name: "Dining", Type: "expense", Icon: "utensils", Color: "#EA580C"}) {
		t.Fatalf("category=%+v params=%+v", category, got)
	}
}

func TestCreateCategoryRejectsInvalidValues(t *testing.T) {
	valid := CreateInput{Name: "Dining", Type: "expense"}
	for _, tc := range []struct {
		name   string
		change func(*CreateInput)
	}{
		{name: "blank name", change: func(input *CreateInput) { input.Name = "  " }},
		{name: "long name", change: func(input *CreateInput) { input.Name = strings.Repeat("a", 101) }},
		{name: "invalid type", change: func(input *CreateInput) { input.Type = "transfer" }},
		{name: "long icon", change: func(input *CreateInput) { input.Icon = strings.Repeat("a", 101) }},
		{name: "long color", change: func(input *CreateInput) { input.Color = strings.Repeat("a", 101) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			repository := &fakeRepository{createFn: func(context.Context, string, CreateParams) (Category, error) {
				t.Fatal("repository called for invalid category")
				return Category{}, nil
			}}
			if _, err := NewService(repository).Create(context.Background(), testUserID, input); !errors.Is(err, ErrInvalidCategory) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidCategory)
			}
		})
	}
}

func TestUpdateCategoryNormalizesInput(t *testing.T) {
	name, categoryType, icon, color := " Food ", " EXPENSE ", " utensils ", " #fff "
	var got UpdateParams
	repository := &fakeRepository{updateFn: func(_ context.Context, userID, categoryID string, params UpdateParams) (Category, error) {
		if userID != testUserID || categoryID != testCategoryID {
			t.Fatalf("update target = %q/%q", userID, categoryID)
		}
		got = params
		return Category{ID: categoryID}, nil
	}}

	_, err := NewService(repository).Update(context.Background(), testUserID, testCategoryID, UpdateInput{
		Name: &name, Type: &categoryType, Icon: &icon, Color: &color,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name == nil || *got.Name != "Food" || got.Type == nil || *got.Type != "expense" ||
		got.Icon == nil || *got.Icon != "utensils" || got.Color == nil || *got.Color != "#fff" {
		t.Fatalf("update params = %+v", got)
	}
}

func TestCategoryUniquenessConflictsAreStable(t *testing.T) {
	duplicate := &pgconn.PgError{Code: "23505"}
	for _, tc := range []struct {
		name string
		run  func(*Service) error
	}{
		{name: "create", run: func(service *Service) error {
			_, err := service.Create(context.Background(), testUserID, CreateInput{Name: "Dining", Type: "expense"})
			return err
		}},
		{name: "update", run: func(service *Service) error {
			name := "Dining"
			_, err := service.Update(context.Background(), testUserID, testCategoryID, UpdateInput{Name: &name})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := &fakeRepository{
				createFn: func(context.Context, string, CreateParams) (Category, error) { return Category{}, duplicate },
				updateFn: func(context.Context, string, string, UpdateParams) (Category, error) { return Category{}, duplicate },
			}
			if err := tc.run(NewService(repository)); !errors.Is(err, ErrCategoryConflict) {
				t.Fatalf("error = %v, want %v", err, ErrCategoryConflict)
			}
		})
	}
}

func TestDefaultCategoryCannotBeModified(t *testing.T) {
	repository := &fakeRepository{
		updateFn: func(context.Context, string, string, UpdateParams) (Category, error) {
			return Category{}, pgx.ErrNoRows
		},
		deleteFn: func(context.Context, string, string) (bool, error) { return false, nil },
	}
	service := NewService(repository)
	name := "Changed"
	if _, err := service.Update(context.Background(), testUserID, testCategoryID, UpdateInput{Name: &name}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("update error = %v, want %v", err, ErrCategoryNotFound)
	}
	if err := service.Delete(context.Background(), testUserID, testCategoryID); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("delete error = %v, want %v", err, ErrCategoryNotFound)
	}
}

func TestUpdateCategoryRequiresAValidChange(t *testing.T) {
	blank := " "
	for _, input := range []UpdateInput{{}, {Icon: &blank}, {Color: &blank}} {
		repository := &fakeRepository{updateFn: func(context.Context, string, string, UpdateParams) (Category, error) {
			t.Fatal("repository called for invalid update")
			return Category{}, nil
		}}
		if _, err := NewService(repository).Update(context.Background(), testUserID, testCategoryID, input); !errors.Is(err, ErrInvalidCategory) {
			t.Fatalf("input=%+v error=%v, want %v", input, err, ErrInvalidCategory)
		}
	}
}
