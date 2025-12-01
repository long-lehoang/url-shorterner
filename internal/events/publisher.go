// Package events provides event publishing interfaces for asynchronous event handling.
package events

import (
	"context"
	"url-shorterner/svc/analytics/events"
)

// Publisher defines the interface for publishing events.
type Publisher interface {
	PublishClickEvent(ctx context.Context, event events.ClickEvent) error
}
