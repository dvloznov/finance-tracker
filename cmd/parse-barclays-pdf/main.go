package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

const (
	// Adjust if you want another model; 2.5-flash is fast + good at docs.
	modelName = "gemini-2.5-flash"
	// Path to your local Barclays test statement (not committed to git).
	pdfPath = "static/Statement 14-NOV-25 AC 70745057  16043935.pdf"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Create Gen AI client.
	// Vertex vs Gemini Dev is controlled via env vars:
	//  - GOOGLE_GENAI_USE_VERTEXAI=True  -> Vertex AI
	//  - GOOGLE_CLOUD_PROJECT
	//  - GOOGLE_CLOUD_LOCATION
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		// API version v1 is what docs use for current Gemini models.
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}

	// Read local Barclays PDF into memory.
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF at %q: %w", pdfPath, err)
	}

	// Prompt: ask Gemini to parse into *strict* JSON.
	prompt :=
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
			"- \"subcategory\": string (one of the predefined subcategories below)\n\n" +

			"Use ONLY the following Categories and Subcategories:\n\n" +
			"HOUSING:\n" +
			"  - Rent\n" +
			"  - Utilities\n" +
			"  - Internet & Mobile\n" +
			"  - Home Services\n" +
			"  - Home Supplies\n\n" +

			"FOOD & DINING:\n" +
			"  - Groceries\n" +
			"  - Restaurants\n" +
			"  - Cafes\n" +
			"  - Snacks\n\n" +

			"TRANSPORT:\n" +
			"  - Public Transport\n" +
			"  - Ride-Hailing\n" +
			"  - Taxi\n" +
			"  - Fuel\n" +
			"  - Parking\n\n" +

			"SHOPPING:\n" +
			"  - Clothing\n" +
			"  - Electronics\n" +
			"  - Household Items\n" +
			"  - Personal Care\n" +
			"  - Gifts\n\n" +

			"ENTERTAINMENT:\n" +
			"  - Subscriptions\n" +
			"  - Streaming Services\n" +
			"  - Movies\n" +
			"  - Games\n" +
			"  - Events\n\n" +

			"HEALTH & FITNESS:\n" +
			"  - Gym\n" +
			"  - Health Services\n" +
			"  - Medicine\n\n" +

			"TRAVEL:\n" +
			"  - Flights\n" +
			"  - Hotels\n" +
			"  - Vacation\n" +
			"  - Travel Services\n\n" +

			"FINANCIAL:\n" +
			"  - Income\n" +
			"  - Transfers\n" +
			"  - Fees\n" +
			"  - Taxes\n\n" +

			"OTHER:\n" +
			"  - Charity\n" +
			"  - Other\n\n" +

			"Rules:\n" +
			"- Classify each transaction into the most appropriate category/subcategory.\n" +
			"- If you are unsure, default to category \"OTHER\" / subcategory \"Other\".\n" +
			"- If the statement has separate \"paid out\" / \"paid in\" columns, convert to a single signed \"amount\".\n" +
			"- If the running balance is missing, set balance_after to null.\n" +
			"- If account name or number cannot be determined, set them to null.\n" +
			"- If the PDF contains multiple accounts, attribute transactions correctly.\n\n" +

			"Return ONLY valid raw JSON.\n" +
			"Do NOT wrap the response in code fences.\n" +
			"Do NOT use ```json or any Markdown.\n" +
			"Output must begin with \"[\" and end with \"]\".\n"

	// Build multimodal content:
	//  - Text instructions
	//  - Inline PDF bytes
	// Docs show using InlineData / Blob for inline media (PDFs allowed).  [oai_citation:3‡Google AI for Developers](https://ai.google.dev/gemini-api/docs/document-processing?utm_source=chatgpt.com)
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{
					Text: prompt,
				},
				{
					InlineData: &genai.Blob{
						MIMEType: "application/pdf",
						Data:     pdfBytes,
					},
				},
			},
		},
	}

	// Call the model.
	resp, err := client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// For now, just dump the text to stdout.
	// Ideally this will be a JSON array string per our prompt.
	fmt.Println(resp.Text())

	return nil
}
