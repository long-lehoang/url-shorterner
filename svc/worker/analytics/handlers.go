// Package analytics provides event handlers for analytics events.
package analytics

import (
	"context"
	"log"
	"url-shorterner/svc/analytics/app"
	"url-shorterner/svc/analytics/events"
)

type EventHandlers struct {
	service app.Service
}

func NewEventHandlers(service app.Service) *EventHandlers {
	return &EventHandlers{
		service: service,
	}
}

func (h *EventHandlers) HandleClickEvent(ctx context.Context, event events.ClickEvent) error {
	err := h.service.RecordClick(
		ctx,
		event.ShortCode,
		event.IPAddress,
		event.UserAgent,
		event.Referer,
	)
	if err != nil {
		log.Printf("Failed to record click event: %v", err)
		return err
	}
	return nil
}
