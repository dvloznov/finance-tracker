package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub/v2"
	"github.com/dvloznov/finance-tracker/internal/events"
	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Parse CLI flags
	pubsubProject := flag.String("pubsub-project", os.Getenv("PUBSUB_PROJECT"), "Pub/Sub project ID (or PUBSUB_PROJECT env)")
	pubsubSubscription := flag.String("pubsub-subscription", os.Getenv("PUBSUB_SUBSCRIPTION"), "Pub/Sub subscription ID or full name (or PUBSUB_SUBSCRIPTION env)")
	flag.Parse()

	if *pubsubProject == "" {
		*pubsubProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	if *pubsubProject == "" || *pubsubSubscription == "" {
		log.Fatal().Msg("PUBSUB_PROJECT (or GOOGLE_CLOUD_PROJECT) and PUBSUB_SUBSCRIPTION are required")
	}

	subscriptionID := events.NormalizeResourceID(*pubsubSubscription)

	log.Info().
		Str("project", *pubsubProject).
		Str("subscription", subscriptionID).
		Msg("Starting worker service")

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := pubsub.NewClient(ctx, *pubsubProject)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pub/Sub client")
	}
	defer client.Close()

	sub := client.Subscriber(subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 5
	sub.ReceiveSettings.MaxOutstandingBytes = 10 << 20

	// Handle interrupts to stop subscriber
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down worker service...")
		cancel()
	}()

	// Build shared infrastructure dependencies once at startup
	docRepo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create document repository")
	}
	defer docRepo.Close()

	accountRepo, err := infraBQ.NewBigQueryAccountRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create account repository")
	}
	defer accountRepo.Close()

	institutionRepo, err := infraBQ.NewBigQueryInstitutionRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create institution repository")
	}
	defer institutionRepo.Close()

	merchantRepo, err := infraBQ.NewBigQueryMerchantRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create merchant repository")
	}
	defer merchantRepo.Close()

	storageService := &gcsuploader.GCSStorageService{}
	aiParser := pipeline.NewGeminiAIParser(docRepo)

	log.Info().Msg("Worker service started, waiting for document events...")

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var event events.DocumentUploadedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Error().Err(err).Msg("Failed to decode document event")
			msg.Nack()
			return
		}

		if event.Type != events.DocumentUploadedEventType {
			log.Warn().Str("event_type", event.Type).Msg("Ignoring unsupported event type")
			msg.Ack()
			return
		}

		if event.DocumentID == "" || event.GCSURI == "" {
			log.Error().Msg("Invalid document event payload")
			msg.Nack()
			return
		}

		log.Info().
			Str("event_id", event.EventID).
			Str("document_id", event.DocumentID).
			Str("gcs_uri", event.GCSURI).
			Msg("Processing document upload event")

		if err := docRepo.UpdateDocumentParsingStatus(ctx, event.DocumentID, "PROCESSING"); err != nil {
			log.Warn().Err(err).Str("document_id", event.DocumentID).Msg("Failed to update document status")
		}

		err := pipeline.IngestStatementFromGCSWithDeps(
			ctx,
			event.GCSURI,
			event.DocumentID,
			docRepo,
			accountRepo,
			institutionRepo,
			merchantRepo,
			storageService,
			aiParser,
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("event_id", event.EventID).
				Str("document_id", event.DocumentID).
				Msg("Pipeline execution failed")

			if updateErr := docRepo.UpdateDocumentParsingStatus(ctx, event.DocumentID, "FAILED"); updateErr != nil {
				log.Error().Err(updateErr).Str("document_id", event.DocumentID).Msg("Failed to update document status to FAILED")
			}

			msg.Nack()
			return
		}

		log.Info().
			Str("event_id", event.EventID).
			Str("document_id", event.DocumentID).
			Msg("Pipeline execution completed successfully")

		msg.Ack()
	})

	if err != nil && ctx.Err() == nil {
		log.Fatal().Err(err).Msg("Pub/Sub receive loop ended with error")
	}

	log.Info().Msg("Worker service exited")
}
