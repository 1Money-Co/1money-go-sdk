/*
 * Copyright 2025 1Money Co.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/1Money-Co/1money-go-sdk/pkg/webhook"
)

// WebhookEvent represents a single webhook event as stored in Badger and
// broadcasted on the in-memory channel.
type WebhookEvent struct {
	Metadata webhook.Metadata `json:"metadata"`
	Payload  json.RawMessage  `json:"payload"`
	StoredAt time.Time        `json:"stored_at"`
}

// eventStore is a thin wrapper around Badger used by the webhook server.
type eventStore struct {
	db *badger.DB
}

func newEventStore(path string) (*eventStore, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", path, err)
	}

	opts := badger.DefaultOptions(filepath.Clean(path))
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &eventStore{db: db}, nil
}

func (s *eventStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *eventStore) ExportJSON(path string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("event store not initialized")
	}

	var events []json.RawMessage

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			events = append(events, json.RawMessage(value))
		}
		return nil
	})
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func (s *eventStore) Save(evt WebhookEvent) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("event store not initialized")
	}

	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("event:%d:%s:%s", evt.Metadata.RawUnixTs, evt.Metadata.EventID, evt.Metadata.DeliveryID)

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// WebhookServer encapsulates the webhook HTTP server, event persistence, and
// in-memory event broadcast channel.
type WebhookServer struct {
	addr       string
	exportPath string

	store  *eventStore
	server *http.Server

	events chan WebhookEvent
}

// NewWebhookServerFromEnv constructs a WebhookServer using environment variables
// for configuration:
//   - WEBHOOK_SECRET_KEY
//   - WEBHOOK_DB_PATH (default: ./data/webhook-events)
//   - WEBHOOK_EXPORT_PATH (default: ./data/webhook-events.json)
//   - WEBHOOK_LISTEN_ADDR (default: 0.0.0.0:25556)
func NewWebhookServerFromEnv() (*WebhookServer, error) {
	secret := os.Getenv("WEBHOOK_SECRET_KEY")

	dbPath := os.Getenv("WEBHOOK_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/webhook-events"
	}

	exportPath := os.Getenv("WEBHOOK_EXPORT_PATH")
	if exportPath == "" {
		exportPath = "./data/webhook-events.json"
	}

	sandbox := os.Getenv("WEBHOOK_SANDBOX") == "1"

	addr := os.Getenv("WEBHOOK_LISTEN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:25556"
	}

	store, err := newEventStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open webhook event store at %s: %w", dbPath, err)
	}

	s := &WebhookServer{
		addr:       addr,
		exportPath: exportPath,
		store:      store,
		events:     make(chan WebhookEvent, 128),
	}

	handler := webhook.NewHandler(secret, &webhook.HandlerOptions{
		OnEvent: s.handleEvent,
		Sandbox: sandbox,
	})

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.POST("/webhook", gin.WrapH(handler))

	s.server = &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	return s, nil
}

// Events returns a read-only channel that receives a WebhookEvent for each
// verified webhook processed by the server.
func (s *WebhookServer) Events() <-chan WebhookEvent {
	return s.events
}

// handleEvent is wired into the webhook.Handler and is responsible for
// persisting and broadcasting events.
func (s *WebhookServer) handleEvent(ctx context.Context, meta webhook.Metadata, payload json.RawMessage) {
	evt := WebhookEvent{
		Metadata: meta,
		Payload:  payload,
		StoredAt: time.Now().UTC(),
	}

	if err := s.store.Save(evt); err != nil {
		logger.Error("failed to persist webhook event", zap.Error(err))
	}

	select {
	case s.events <- evt:
	default:
		logger.Warn("dropping webhook event; channel buffer full",
			zap.String("event_id", meta.EventID),
			zap.String("event_type", meta.EventType),
			zap.String("delivery_id", meta.DeliveryID),
		)
	}

	if len(payload) > 0 {
		var pretty map[string]any
		if err := json.Unmarshal(payload, &pretty); err != nil {
			logger.Warn("failed to parse webhook payload", zap.Error(err))
			return
		}

		logger.Info("webhook payload", zap.Any("payload", pretty))
	}
}

// Start begins serving HTTP requests and listens for the provided context
// cancellation to gracefully shutdown. It also exports all stored events
// to JSON and closes internal resources on shutdown.
func (s *WebhookServer) Start(ctx context.Context) {
	go func() {
		logger.Info("starting webhook server",
			zap.String("addr", s.addr),
		)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("webhook server error", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown webhook server", zap.Error(err))
		} else {
			logger.Info("webhook server shutdown complete")
		}

		if err := s.store.ExportJSON(s.exportPath); err != nil {
			logger.Error("failed to export events to json", zap.Error(err), zap.String("path", s.exportPath))
		} else {
			logger.Info("exported webhook events to json", zap.String("path", s.exportPath))
		}

		if err := s.store.Close(); err != nil {
			logger.Error("failed to close webhook event store", zap.Error(err))
		} else {
			logger.Info("webhook event store closed")
		}

		close(s.events)
	}()
}

// startWebhookServer is a convenience helper used by the example main.
// It constructs a WebhookServer from environment configuration and starts it
// with an internal signal-aware context.
func startWebhookServer() *WebhookServer {
	srv, err := NewWebhookServerFromEnv()
	if err != nil {
		logger.Error("failed to initialize webhook server", zap.Error(err))
		return nil
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	srv.Start(ctx)

	return srv
}
