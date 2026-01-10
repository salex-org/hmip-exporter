package homematic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

type smokeDetectorMetric struct {
	smokeChamberDegradedMetric *prometheus.GaugeVec
}

func newSmokeDetectorMetric() homematicMetric {
	metric := &smokeDetectorMetric{
		smokeChamberDegradedMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "smoke_detector",
			Name:      "smoke_chamber_degraded",
			Help:      "Current smoke chamber status of a smoke detector (0 = ok, 1 = degraded)",
		}, metricLabelNames),
	}
	prometheus.MustRegister(metric.smokeChamberDegradedMetric)

	return metric
}

func (m *smokeDetectorMetric) update(device hmip.Device, labels prometheus.Labels) {
	for _, base := range device.GetFunctionalChannels() {
		switch channel := base.(type) {
		case hmip.SmokeDetectorChannel:
			var value float64 = 0
			if channel.IsChamberDegraded() {
				value = 1
			}
			m.smokeChamberDegradedMetric.With(labels).Set(value)
		}
	}
}
