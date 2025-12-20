package pipeline

import (
	"context"
	"fmt"
	"strings"
)

// buildCategoriesPromptWithRepo constructs a prompt string containing all active categories
// and subcategories from BigQuery, formatted for LLM consumption.
func buildCategoriesPromptWithRepo(ctx context.Context, repo CategoryRepository) (string, error) {
	rows, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return "", fmt.Errorf("buildCategoriesPrompt: list categories: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("buildCategoriesPrompt: no active categories found")
	}

	// Group by category name
	categoryMap := make(map[string][]string)
	for _, row := range rows {
		cat := row.CategoryName
		if row.SubcategoryName.Valid && row.SubcategoryName.StringVal != "" {
			categoryMap[cat] = append(categoryMap[cat], row.SubcategoryName.StringVal)
		} else {
			// Ensure category exists even with no subcategory
			if _, exists := categoryMap[cat]; !exists {
				categoryMap[cat] = []string{}
			}
		}
	}

	var b strings.Builder
	b.WriteString("Use ONLY the following Categories and Subcategories:\n\n")

	for cat, subs := range categoryMap {
		b.WriteString(cat + ":\n")
		if len(subs) == 0 {
			b.WriteString("  (no subcategories - use empty string \"\")\n\n")
			continue
		}
		for _, s := range subs {
			b.WriteString("  - " + s + "\n")
		}
		b.WriteString("\n")
	}

	// Additionally, constrain what the model is allowed to output.
	b.WriteString("CATEGORY ASSIGNMENT RULES:\n")
	b.WriteString("1. Category must be EXACTLY one of the category names shown above (case-sensitive).\n")
	b.WriteString("2. If a category has subcategories listed, you MUST choose one of them - never use empty string.\n")
	b.WriteString("3. If a category shows \"(no subcategories)\", use empty string \"\" for subcategory.\n")
	b.WriteString("4. If you are unsure, use category \"Uncategorized\" with subcategory \"\".\n")
	b.WriteString("5. For Uber/taxi rides, use: category \"Transportation\", subcategory \"Public Transit\".\n")
	b.WriteString("6. Never leave subcategory empty when the category has available subcategories.\n")

	return b.String(), nil
}

// buildAccountHeaderPrompt constructs a prompt for extracting account metadata
// from the bank statement header (not individual transactions).
func buildAccountHeaderPrompt() string {
	return "You are a financial statement parser for Barclays UK PDF bank statements.\n\n" +
		"Task:\n" +
		"- Extract ONLY the account metadata from the statement header/top section.\n" +
		"- DO NOT parse transactions - only account information.\n" +
		"- Output STRICT JSON only (no comments, no trailing commas, no extra text).\n\n" +
		"Output a single JSON object with these fields:\n" +
		"- \"account_number\": string or null (last 4 digits or full account number)\n" +
		"- \"iban\": string or null (International Bank Account Number)\n" +
		"- \"sort_code\": string or null (UK bank sort code, format XX-XX-XX)\n" +
		"- \"account_name\": string or null (e.g., \"Current Account\", \"Savings Account\")\n" +
		"- \"account_type\": string or null (e.g., \"CURRENT\", \"SAVINGS\", \"CREDIT_CARD\")\n" +
		"- \"currency\": string or null (e.g., \"GBP\", \"USD\", \"EUR\")\n" +
		"- \"institution_id\": string or null (bank name, e.g., \"BARCLAYS\")\n" +
		"- \"opened_date\": string or null (ISO format \"YYYY-MM-DD\" if shown on statement)\n\n" +
		"Rules:\n" +
		"- Set a field to null if the information is not present in the statement header.\n" +
		"- Focus ONLY on the top section/header of the statement, not transaction details.\n" +
		"- For sort_code, preserve the hyphen format if shown (e.g., \"20-00-00\").\n" +
		"- For currency, use the 3-letter ISO code (GBP, USD, EUR, etc.).\n" +
		"- For account_type, use uppercase: CURRENT, SAVINGS, CREDIT_CARD, etc.\n\n" +
		"CRITICAL OUTPUT REQUIREMENTS:\n" +
		"- Return ONLY valid, parseable JSON that follows RFC 8259 standard.\n" +
		"- Do NOT wrap the response in code fences.\n" +
		"- Do NOT use ```json or any Markdown.\n" +
		"- Do NOT include any comments or explanatory text.\n" +
		"- Output must be a single JSON object: {...}\n" +
		"- Example format: {\"account_number\": \"1234\", \"iban\": null, ...}\n"
}

// buildTransactionSchema returns the transaction schema portion of the prompt.
// Account fields (account_name, account_number) are removed since accounts are
// extracted separately via buildAccountHeaderPrompt.
func buildTransactionSchema() string {
	return "Each transaction object must have these fields:\n" +
		"- \"date\": string, ISO format \"YYYY-MM-DD\"\n" +
		"- \"description\": string\n" +
		"- \"amount\": number (positive for money IN, negative for money OUT)\n" +
		"- \"currency\": string (e.g. \"GBP\")\n" +
		"- \"balance_after\": number or null\n" +
		"- \"category\": string (MUST be one of the predefined categories below)\n" +
		"- \"subcategory\": string (MUST be one of the valid subcategories for that category, or empty string if category has no subcategories)\n\n"
}
