package provider

import (
	"encoding/json"

	"github.com/devblin/tuskira/internal/model"
)

// Provider is the interface each notification channel (email, slack, inapp) implements.
type Provider interface {
	Send(notification *model.Notification, rawCfg json.RawMessage) error
	Channel() model.Channel
}

// Registry maps channel types to their provider implementations.
type Registry struct {
	providers map[model.Channel]Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[model.Channel]Provider),
	}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Channel()] = p
}

func (r *Registry) Get(channel model.Channel) (Provider, bool) {
	p, ok := r.providers[channel]
	return p, ok
}
