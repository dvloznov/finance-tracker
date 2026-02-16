package pipeline

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
		"- \"description\": string (merchant/description exactly as written in the statement)\n" +
		"- \"amount\": number (positive for money IN, negative for money OUT)\n" +
		"- \"currency\": string (e.g. \"GBP\")\n" +
		"- \"balance_after\": number or null\n\n"
}
