package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"google.golang.org/genai"
)

var llmCallCounter atomic.Uint64

// loggedGenerateContent wraps client.Models.GenerateContent with structured
// before/after log events, recording prompt text, latency, and raw output.
// PDF bytes in contents are never logged; only the text prompt string is captured.
func loggedGenerateContent(ctx context.Context, client *genai.Client, operation, prompt string, contents []*genai.Content) (*genai.GenerateContentResponse, error) {
	log := logger.FromContext(ctx)
	callID := llmCallCounter.Add(1)

	log.Info().
		Uint64("llm_call_id", callID).
		Str("llm_operation", operation).
		Str("llm_model", DefaultModelName).
		Str("llm_prompt", prompt).
		Msg("LLM call started")

	start := time.Now()
	resp, err := client.Models.GenerateContent(ctx, DefaultModelName, contents, nil)
	latency := time.Since(start)

	if err != nil {
		log.Error().
			Err(err).
			Uint64("llm_call_id", callID).
			Str("llm_operation", operation).
			Dur("llm_latency", latency).
			Bool("llm_success", false).
			Msg("LLM call finished")
		return nil, err
	}

	log.Info().
		Uint64("llm_call_id", callID).
		Str("llm_operation", operation).
		Dur("llm_latency", latency).
		Bool("llm_success", true).
		Str("llm_raw_output", resp.Text()).
		Msg("LLM call finished")

	return resp, nil
}

// parseStatementWithModel sends the PDF to Gemini and returns the parsed JSON output.
// It expects the model to return a STRICT JSON array of transactions.
func parseStatementWithModel(ctx context.Context, pdfBytes []byte, repo CategoryRepository, bankName string) (map[string]interface{}, string, error) {
	basePrompt :=
		"You are a financial statement parser for " + bankName + " PDF bank statements.\n\n" +
			"Task:\n" +
			"- Parse ALL transactions in the attached statement.\n" +
			"- Extract structured fields only;\n" +
			"- Output STRICT JSON only (no comments, no trailing commas, no extra text).\n" +
			"- Output a JSON array of objects.\n\n"

	// Transaction schema (account fields removed - handled separately).
	txSchema := buildTransactionSchema()

	rulesPrompt :=
		"Rules:\n" +
			"- Use the description exactly as written in the statement.\n" +
			"- If the statement has separate \"Money out\" / \"Money in\" or \"Paid out\" / \"Paid in\" columns:\n" +
			"  * Values in the OUT column must be output as NEGATIVE amounts.\n" +
			"  * Values in the IN column must be output as positive amounts.\n" +
			"  * Never output a positive amount for money that left the account.\n" +
			"- If the running balance is missing, set \"balance_after\" to null.\n\n" +
			"CRITICAL OUTPUT REQUIREMENTS:\n" +
			"- Return ONLY valid, parseable JSON that follows RFC 8259 standard.\n" +
			"- Separate array elements with COMMAS (,) - never use words or other separators.\n" +
			"- Do NOT wrap the response in code fences.\n" +
			"- Do NOT use ```json or any Markdown.\n" +
			"- Do NOT include any comments or explanatory text.\n" +
			"- Output must begin with \"[\" and end with \"]\".\n" +
			"- Example format: [{...}, {...}, {...}]\n"

	fullPrompt := basePrompt + txSchema + "\n" + rulesPrompt

	// 3) Create GenAI client (same style as your test program).
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, "", fmt.Errorf("parseStatementWithModel: create genai client: %w", err)
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: fullPrompt},
				{
					InlineData: &genai.Blob{
						MIMEType: "application/pdf",
						Data:     pdfBytes,
					},
				},
			},
		},
	}

	resp, err := loggedGenerateContent(ctx, client, "parse_statement", fullPrompt, contents)
	if err != nil {
		return nil, "", fmt.Errorf("parseStatementWithModel: generate content: %w", err)
	}

	rawText := resp.Text()
	if rawText == "" {
		return nil, "", fmt.Errorf("parseStatementWithModel: empty response from model")
	}

	// Clean up Markdown fences / extra text if the model ignored instructions.
	clean := cleanModelJSON(rawText)

	// 4) Parse JSON into a generic value.
	var parsed interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, "", fmt.Errorf("parseStatementWithModel: unmarshal JSON: %w\nraw response: %s", err, rawText)
	}

	// Expect top-level array; for flexibility we just wrap it under "transactions".
	return map[string]interface{}{
		"transactions": parsed,
	}, fullPrompt, nil
}

func cleanModelJSON(raw string) string {
	s := strings.TrimSpace(raw)

	// Handle ```json ... ``` or ``` ... ``` wrappers.
	if strings.HasPrefix(s, "```") {
		// Drop the first line (``` or ```json).
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		} else {
			// Single-line weirdness; just return as-is.
			return s
		}
		s = strings.TrimSpace(s)
	}

	// Remove trailing ``` if present.
	if idx := strings.LastIndex(s, "```"); idx != -1 {
		s = s[:idx]
	}

	s = strings.TrimSpace(s)

	// Fix common AI model errors: replace "tapos" or similar separators with commas
	// This handles cases where the model outputs "} tapos {" instead of "}, {"
	s = strings.ReplaceAll(s, "} tapos", "},")
	s = strings.ReplaceAll(s, "}  tapos", "},")
	s = strings.ReplaceAll(s, "}\n  tapos", "},")
	s = strings.ReplaceAll(s, "}\ntapos", "},")

	// Extra safety: if there's still junk around the JSON array,
	// try to keep only from the first '[' to the last ']'.
	if start := strings.Index(s, "["); start != -1 {
		if end := strings.LastIndex(s, "]"); end != -1 && end > start {
			s = s[start : end+1]
			s = strings.TrimSpace(s)
		}
	}

	return s
}

// extractAccountHeaderWithModel sends the PDF to Gemini and returns the parsed account metadata.
// It expects the model to return a STRICT JSON object with account fields.
func extractAccountHeaderWithModel(ctx context.Context, pdfBytes []byte) (map[string]interface{}, string, error) {
	// Use the account header extraction prompt
	prompt := buildAccountHeaderPrompt()

	// Create GenAI client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, "", fmt.Errorf("extractAccountHeaderWithModel: create genai client: %w", err)
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
				{
					InlineData: &genai.Blob{
						MIMEType: "application/pdf",
						Data:     pdfBytes,
					},
				},
			},
		},
	}

	resp, err := loggedGenerateContent(ctx, client, "extract_account_header", prompt, contents)
	if err != nil {
		return nil, "", fmt.Errorf("extractAccountHeaderWithModel: generate content: %w", err)
	}

	rawText := resp.Text()
	if rawText == "" {
		return nil, "", fmt.Errorf("extractAccountHeaderWithModel: empty response from model")
	}

	// Clean up Markdown fences / extra text
	clean := cleanModelJSON(rawText)

	// Parse JSON into a generic value
	var parsed interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, "", fmt.Errorf("extractAccountHeaderWithModel: unmarshal JSON: %w\nraw response: %s", err, rawText)
	}

	// Expect a JSON object (not array)
	accountObj, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("extractAccountHeaderWithModel: expected JSON object, got %T", parsed)
	}

	return accountObj, prompt, nil
}

// categorizeMerchantWithModel sends a merchant name to Gemini and returns the selected category_id.
// Returns an empty string if no category is selected.
func categorizeMerchantsWithModel(ctx context.Context, merchantNames []string, categories []bigquery.CategoryRow) (map[string]string, string, error) {
	if len(merchantNames) == 0 {
		return map[string]string{}, "", nil
	}
	if len(categories) == 0 {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: categories list is empty")
	}

	categoryLines := make([]string, 0, len(categories))
	for _, c := range categories {
		label := fmt.Sprintf("%s | %s", c.CategoryID, c.CategoryName)
		if c.SubcategoryName.Valid {
			label += " > " + c.SubcategoryName.StringVal
		}
		categoryLines = append(categoryLines, label)
	}

	prompt := buildMerchantCategorizationPrompt(merchantNames, categoryLines)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: create genai client: %w", err)
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{{Text: prompt}},
		},
	}

	resp, err := loggedGenerateContent(ctx, client, "categorize_merchants", prompt, contents)
	if err != nil {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: generate content: %w", err)
	}

	rawText := resp.Text()
	if rawText == "" {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: empty response from model")
	}

	clean := cleanModelJSON(rawText)

	var parsed interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: unmarshal JSON: %w\nraw response: %s", err, rawText)
	}

	var items []interface{}
	if obj, ok := parsed.(map[string]interface{}); ok {
		if arr, ok := obj["merchants"].([]interface{}); ok {
			items = arr
		} else {
			return nil, "", fmt.Errorf("categorizeMerchantsWithModel: expected merchants array")
		}
	} else if arr, ok := parsed.([]interface{}); ok {
		items = arr
	} else {
		return nil, "", fmt.Errorf("categorizeMerchantsWithModel: expected JSON object or array, got %T", parsed)
	}

	results := make(map[string]string, len(items))
	for i, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("categorizeMerchantsWithModel: merchants[%d] is %T, want object", i, item)
		}
		nameVal, ok := entry["merchant_name"].(string)
		if !ok {
			return nil, "", fmt.Errorf("categorizeMerchantsWithModel: merchants[%d] missing merchant_name", i)
		}
		categoryVal := ""
		if entry["category_id"] != nil {
			if s, ok := entry["category_id"].(string); ok {
				categoryVal = strings.TrimSpace(s)
			}
		}
		results[strings.TrimSpace(nameVal)] = categoryVal
	}

	return results, prompt, nil
}
