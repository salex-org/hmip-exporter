package homematic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

type switchMetric struct {
	isOnMetric               *prometheus.GaugeVec
	currentActivePowerMetric *prometheus.GaugeVec
}

func newSwitchMetric() homematicMetric {
	metric := &switchMetric{
		isOnMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "outlet",
			Name:      "current_state",
			Help:      "Current switch state of an outlet (0 = off, 1 = on)",
		}, metricLabelNames),
		currentActivePowerMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "outlet",
			Name:      "current_active_power",
			Help:      "Power currently consumed at an outlet - consumers only (watts)",
		}, metricLabelNames),
	}
	prometheus.MustRegister(metric.isOnMetric)
	prometheus.MustRegister(metric.currentActivePowerMetric)

	return metric
}

func (m *switchMetric) update(device hmip.Device, labels prometheus.Labels) {
	for _, base := range device.GetFunctionalChannels() {
		switch channel := base.(type) {
		case hmip.SwitchMeasuringChannel:
			m.setIsOnMetric(channel, labels)
			m.currentActivePowerMetric.With(labels).Set(channel.GetCurrentPowerConsumption())
		case hmip.SwitchChannel:
			m.setIsOnMetric(channel, labels)
		}
	}
}

func (m *switchMetric) setIsOnMetric(sw hmip.Switchable, labels prometheus.Labels) {
	var value float64 = 0
	if sw.IsSwitchedOn() {
		value = 1
	}
	m.isOnMetric.With(labels).Set(value)
}
