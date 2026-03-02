package pipeline

import "strings"

// buildAccountHeaderPrompt constructs a prompt for extracting account metadata
// from the bank statement header (not individual transactions).
func buildAccountHeaderPrompt() string {
	return "You are a financial statement parser for bank statements.\n\n" +
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
		"- \"institution_name\": string or null (bank name, e.g., \"Barclays\")\n\n" +
		"- \"statement_start_date\": string or null (ISO date YYYY-MM-DD, if present)\n" +
		"- \"statement_end_date\": string or null (ISO date YYYY-MM-DD, if present)\n\n" +
		"Rules:\n" +
		"- Set a field to null if the information is not present in the statement header.\n" +
		"- Focus ONLY on the top section/header of the statement, not transaction details.\n" +
		"- For sort_code, preserve the hyphen format if shown (e.g., \"20-00-00\").\n" +
		"- For currency, use the 3-letter ISO code (GBP, USD, EUR, etc.).\n" +
		"- For account_type, use one of these uppercase values:\n" +
		"    CURRENT    — standard current/checking account\n" +
		"    SAVINGS    — savings account\n" +
		"    CREDIT_CARD — any credit card (American Express, Amex, Visa, Mastercard credit, etc.)\n" +
		"  If the statement is from American Express or any other credit card provider, set account_type to CREDIT_CARD.\n" +
		"- For American Express or Amex statements: extract the membership number (account_number) without leading asterisks (***). Use the full digits only.\n\n" +
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
		"- \"date\": string, ISO format \"YYYY-MM-DD\" (date from description if present; otherwise the statement date)\n" +
		"- \"description\": string (full description exactly as written in the statement)\n" +
		"- \"merchant_name\": string (merchant name exactly as written in the statement; if unsure, use description)\n" +
		"- \"statement_date\": string, ISO format \"YYYY-MM-DD\" (date when the transaction appears on the statement)\n" +
		"- \"transaction_type\": string or null (e.g., \"CARD PAYMENT\", \"CARD PURCHASE\", \"DIRECT DEBIT\", \"PAYMENT RECEIVED\")\n" +
		"- \"amount\": number\n" +
		"  SIGN CONVENTION (applies to ALL account types — normalise before outputting):\n" +
		"  * Positive (+): money IN — increases your balance or reduces what you owe.\n" +
		"    Examples: salary received, refund, credit card payment made by you.\n" +
		"  * Negative (-): money OUT — decreases your balance or increases what you owe.\n" +
		"    Examples: purchase, subscription, direct debit, cash withdrawal.\n" +
		"  STATEMENTS WITH SEPARATE COLUMNS (Revolut, etc.):\n" +
		"  * \"Money out\" / \"Paid out\" / \"Out\" column: ALWAYS output as NEGATIVE.\n" +
		"  * \"Money in\" / \"Paid in\" / \"In\" column: output as positive.\n" +
		"  * Do NOT copy the raw number from the money-out column — you MUST negate it.\n" +
		"  CREDIT CARD STATEMENTS (e.g. American Express, Amex):\n" +
		"  * Spend/purchases appear as positive numbers on the statement — NEGATE them (output as negative).\n" +
		"  * Payments/credits are marked CR or CREDIT on the statement — output as positive.\n" +
		"  * Do NOT output spend as positive even if the statement shows it that way.\n" +
		"- \"currency\": string (e.g. \"GBP\")\n" +
		"- \"balance_after\": number or null\n" +
		"  * For current/savings: the running account balance after this transaction.\n" +
		"  * For credit cards: the outstanding balance owed after this transaction (always a positive figure).\n" +
		"  * If no per-transaction balance is shown, set to null.\n\n"
}

// buildMerchantCategorizationPrompt constructs a prompt for categorizing a merchant
// into one of the predefined categories.
func buildMerchantCategorizationPrompt(merchantNames []string, categories []string) string {
	return "You are a merchant categorizer.\n\n" +
		"Task:\n" +
		"- Given a list of merchant names, select the single best category_id for each.\n" +
		"- If none match, set category_id to null.\n\n" +
		"Merchant names (deduplicated):\n" +
		"- " + strings.Join(merchantNames, "\n- ") + "\n\n" +
		"Allowed categories (category_id | category_name > subcategory_name):\n" +
		"- " + strings.Join(categories, "\n- ") + "\n\n" +
		"Output STRICT JSON only (no comments, no trailing commas):\n" +
		"{\"merchants\": [{\"merchant_name\": \"Starbucks\", \"category_id\": \"cat_food_dining_coffee_shops\"}, {\"merchant_name\": \"Unknown\", \"category_id\": null}]}\n\n" +
		"CRITICAL OUTPUT REQUIREMENTS:\n" +
		"- Return ONLY valid, parseable JSON that follows RFC 8259 standard.\n" +
		"- Do NOT wrap the response in code fences.\n" +
		"- Do NOT use ```json or any Markdown.\n" +
		"- Do NOT include any comments or explanatory text.\n" +
		"- Output must be a single JSON object: {...}\n"
}
