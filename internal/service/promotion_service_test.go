package service

import (
	"testing"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

func TestPromotionServiceApplyPromotionGracefullyFallsBackWhenRepoNil(t *testing.T) {
	product := &models.Product{
		ID:          1,
		PriceAmount: models.NewMoneyFromDecimal(decimal.NewFromInt(99)),
	}

	svc := NewPromotionService(nil)
	promotion, price, err := svc.ApplyPromotion(product, 1)
	if err != nil {
		t.Fatalf("expected no error when promotion repo is nil, got %v", err)
	}
	if promotion != nil {
		t.Fatalf("expected nil promotion when repo is nil")
	}
	if !price.Equal(product.PriceAmount.Decimal) {
		t.Fatalf("expected fallback price %s, got %s", product.PriceAmount.Decimal.String(), price.Decimal.String())
	}
}

func TestPromotionServiceGetProductPromotionsReturnsEmptyWhenRepoNil(t *testing.T) {
	svc := NewPromotionService(nil)
	items, err := svc.GetProductPromotions(1)
	if err != nil {
		t.Fatalf("expected no error when promotion repo is nil, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty promotions when repo is nil, got %d", len(items))
	}
}
