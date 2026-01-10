package homematic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

// remoteControlMetric contains specific metrics for remote controls
// Attention: Currently there are no specific metrics for remote controls, just the baseHomematicMetric
type remoteControlMetric struct{}

func newRemoteControlMetric() homematicMetric {
	return &remoteControlMetric{}
}

func (m *remoteControlMetric) update(device hmip.Device, labels prometheus.Labels) {}
