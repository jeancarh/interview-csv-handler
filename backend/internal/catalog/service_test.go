package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestServiceCreateAndUpdate(t *testing.T) {
	svc := NewService(newMemRepo(), fixedClock)
	ctx := context.Background()

	p, err := svc.Create(ctx, validDraft())
	if err != nil {
		t.Fatal(err)
	}

	d := validDraft()
	d.Name = "Trail Shoes"
	updated, err := svc.Update(ctx, p.ID, d)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Trail Shoes" {
		t.Fatalf("got %q", updated.Name)
	}
}

func TestServiceRejectsDuplicateSKU(t *testing.T) {
	svc := NewService(newMemRepo(), fixedClock)
	ctx := context.Background()
	_, _ = svc.Create(ctx, validDraft())

	if _, err := svc.Create(ctx, validDraft()); !errors.Is(err, ErrDuplicateSKU) {
		t.Fatalf("got %v", err)
	}
}

func TestServiceUpdateMissingProduct(t *testing.T) {
	svc := NewService(newMemRepo(), fixedClock)
	if _, err := svc.Update(context.Background(), 99, validDraft()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestServiceSearchClampsPaging(t *testing.T) {
	svc := NewService(newMemRepo(), fixedClock)
	page, err := svc.Search(context.Background(), SearchQuery{Page: -1, PageSize: 5000})
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 1 || page.PageSize != maxPageSize {
		t.Fatalf("got page %d size %d", page.Page, page.PageSize)
	}
}
