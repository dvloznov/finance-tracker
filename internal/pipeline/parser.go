package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// parseStatementWithModel sends the PDF to Gemini and returns the parsed JSON output.
// It expects the model to return a STRICT JSON array of transactions.
func parseStatementWithModel(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	// 1) Build category prompt from BigQuery taxonomy.
	catPrompt, err := buildCategoriesPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: loading categories: %w", err)
	}

	// 2) Base instructions (very close to your test code).
	basePrompt :=
		"You are a financial statement parser for Barclays UK PDF bank statements.\n\n" +
			"Task:\n" +
			"- Parse ALL transactions in the attached Barclays statement.\n" +
			"- Output STRICT JSON only (no comments, no trailing commas, no extra text).\n" +
			"- Output a JSON array of objects.\n\n" +
			"Each object must have these fields:\n" +
			"- \"account_name\": string or null\n" +
			"- \"account_number\": string or null\n" +
			"- \"date\": string, ISO format \"YYYY-MM-DD\"\n" +
			"- \"description\": string\n" +
			"- \"amount\": number (positive for money IN, negative for money OUT)\n" +
			"- \"currency\": string (e.g. \"GBP\")\n" +
			"- \"balance_after\": number or null\n" +
			"- \"category\": string (one of the predefined categories)\n" +
			"- \"subcategory\": string (one of the predefined subcategories below)\n\n"

	rulesPrompt :=
		"Rules:\n" +
			"- Classify each transaction into the most appropriate category/subcategory.\n" +
			"- If the statement has separate \"paid out\" / \"paid in\" columns, convert to a single signed \"amount\".\n" +
			"- If the running balance is missing, set \"balance_after\" to null.\n" +
			"- If account name or number cannot be determined, set them to null.\n" +
			"- If the PDF contains multiple accounts, attribute transactions correctly.\n\n" +
			"Return ONLY valid raw JSON.\n" +
			"Do NOT wrap the response in code fences.\n" +
			"Do NOT use ```json or any Markdown.\n" +
			"Output must begin with \"[\" and end with \"]\".\n"

	fullPrompt := basePrompt + "\n" + catPrompt + "\n\n" + rulesPrompt

	// 3) Create GenAI client (same style as your test program).
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: create genai client: %w", err)
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

	resp, err := client.Models.GenerateContent(ctx, DefaultModelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: generate content: %w", err)
	}

	rawText := resp.Text()
	if rawText == "" {
		return nil, fmt.Errorf("parseStatementWithModel: empty response from model")
	}

	// Clean up Markdown fences / extra text if the model ignored instructions.
	clean := cleanModelJSON(rawText)

	// 4) Parse JSON into a generic value.
	var parsed interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: unmarshal JSON: %w\nraw response: %s", err, rawText)
	}

	// Expect top-level array; for flexibility we just wrap it under "transactions".
	return map[string]interface{}{
		"transactions": parsed,
	}, nil
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
