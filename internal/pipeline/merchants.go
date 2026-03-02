package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/google/uuid"
)

func normalizeMerchantName(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func resolveMerchantsForTransactions(
	ctx context.Context,
	txs []*Transaction,
	merchantRepo bigquery.MerchantRepository,
	categoryRepo bigquery.CategoryRepository,
	documentRepo bigquery.DocumentRepository,
	aiParser AIParser,
	parsingRunID string,
	documentID string,
) error {
	if merchantRepo == nil {
		return fmt.Errorf("resolveMerchantsForTransactions: merchant repository is nil")
	}
	if categoryRepo == nil {
		return fmt.Errorf("resolveMerchantsForTransactions: category repository is nil")
	}
	if documentRepo == nil {
		return fmt.Errorf("resolveMerchantsForTransactions: document repository is nil")
	}
	if aiParser == nil {
		return fmt.Errorf("resolveMerchantsForTransactions: AI parser is nil")
	}

	categories, err := categoryRepo.ListActiveCategories(ctx)
	if err != nil {
		return fmt.Errorf("resolveMerchantsForTransactions: list categories: %w", err)
	}
	categoryIDs := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		categoryIDs[c.CategoryID] = struct{}{}
	}

	cache := make(map[string]*bigquery.MerchantRow)
	unknownByNormalized := make(map[string]string)

	for _, tx := range txs {
		merchantName := strings.TrimSpace(tx.MerchantName)
		if merchantName == "" {
			merchantName = strings.TrimSpace(tx.Description)
		}
		if merchantName == "" {
			merchantName = "Unknown"
		}
		normalized := normalizeMerchantName(merchantName)
		if normalized == "" {
			normalized = "unknown"
		}

		if cached, ok := cache[normalized]; ok {
			tx.MerchantID = cached.MerchantID
			tx.MerchantName = cached.MerchantName
			tx.CategoryID = cached.CategoryID
			continue
		}

		merchant, err := merchantRepo.FindMerchantByNormalizedName(ctx, normalized)
		if err != nil {
			return fmt.Errorf("resolveMerchantsForTransactions: find merchant %q: %w", normalized, err)
		}

		if merchant != nil {
			cache[normalized] = merchant
			tx.MerchantID = merchant.MerchantID
			tx.MerchantName = merchant.MerchantName
			tx.CategoryID = merchant.CategoryID
			continue
		}

		unknownByNormalized[normalized] = merchantName
	}

	if len(unknownByNormalized) > 0 {
		merchantNames := make([]string, 0, len(unknownByNormalized))
		for _, name := range unknownByNormalized {
			merchantNames = append(merchantNames, name)
		}

		categorized, categorizePrompt, err := aiParser.CategorizeMerchants(ctx, merchantNames, categories)
		if err != nil {
			return fmt.Errorf("resolveMerchantsForTransactions: categorize merchants: %w", err)
		}

		// Store categorize_merchants output in model_outputs for audit
		merchantEntries := make([]map[string]interface{}, 0, len(categorized))
		for name, catID := range categorized {
			merchantEntries = append(merchantEntries, map[string]interface{}{
				"merchant_name": name,
				"category_id":   catID,
			})
		}
		categorizeOutput := map[string]interface{}{"merchants": merchantEntries}
		if _, err := storeModelOutputWithRepo(ctx, parsingRunID, documentID,
			"categorize_merchants", categorizePrompt,
			categorizeOutput, documentRepo); err != nil {
			return fmt.Errorf("resolveMerchantsForTransactions: store categorize output: %w", err)
		}

		nameToCategory := make(map[string]string, len(categorized))
		for name, categoryID := range categorized {
			nameToCategory[normalizeMerchantName(name)] = categoryID
		}

		for normalized, originalName := range unknownByNormalized {
			categoryID := nameToCategory[normalized]
			if _, ok := categoryIDs[categoryID]; !ok {
				categoryID = DefaultCategoryID
			}

			newMerchant := &bigquery.MerchantRow{
				MerchantID:     uuid.NewString(),
				MerchantName:   originalName,
				NormalizedName: normalized,
				CategoryID:     categoryID,
				CreatedTS:      time.Now(),
			}

			merchantID, err := merchantRepo.InsertMerchant(ctx, newMerchant)
			if err != nil {
				return fmt.Errorf("resolveMerchantsForTransactions: insert merchant %q: %w", originalName, err)
			}
			newMerchant.MerchantID = merchantID
			cache[normalized] = newMerchant
		}
	}

	for _, tx := range txs {
		merchantName := strings.TrimSpace(tx.MerchantName)
		if merchantName == "" {
			merchantName = strings.TrimSpace(tx.Description)
		}
		if merchantName == "" {
			merchantName = "Unknown"
		}
		normalized := normalizeMerchantName(merchantName)
		if normalized == "" {
			normalized = "unknown"
		}
		merchant, ok := cache[normalized]
		if !ok {
			return fmt.Errorf("resolveMerchantsForTransactions: missing merchant for %q", merchantName)
		}
		tx.MerchantID = merchant.MerchantID
		tx.MerchantName = merchant.MerchantName
		tx.CategoryID = merchant.CategoryID
	}

	return nil
}
