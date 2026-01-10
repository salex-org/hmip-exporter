package homematic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

type baseHomematicMetric struct {
	reachableMetric         *prometheus.GaugeVec
	connectionQualityMetric *prometheus.GaugeVec
	lastSeenMetric          *prometheus.GaugeVec
	batteryLowMetric        *prometheus.GaugeVec
	underVoltageMetric      *prometheus.GaugeVec
	overheatedMetric        *prometheus.GaugeVec
}

func newBaseHomematicMetric() homematicMetric {
	metric := &baseHomematicMetric{
		reachableMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "reachable",
			Help:      "Reachability of a device (0 = unreachable, 1 = reachable)",
		}, metricLabelNames),
		connectionQualityMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "connection_quality_rssi",
			Help:      "Current connection quality",
		}, metricLabelNames),
		lastSeenMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "last_seen_timestamp",
			Help:      "Last time the device was seen (Unix timestamp in seconds)",
		}, metricLabelNames),
		batteryLowMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "battery_low",
			Help:      "Current battery status of a device (0 = ok, 1 = low)",
		}, metricLabelNames),
		underVoltageMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "under_voltage",
			Help:      "Current undervoltage status of a device (0 = ok, 1 = undervoltage)",
		}, metricLabelNames),
		overheatedMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "device",
			Name:      "overheated",
			Help:      "Current temperature status of a device (0 = ok, 1 = overheated)",
		}, metricLabelNames),
	}
	prometheus.MustRegister(metric.reachableMetric)
	prometheus.MustRegister(metric.connectionQualityMetric)
	prometheus.MustRegister(metric.lastSeenMetric)
	prometheus.MustRegister(metric.batteryLowMetric)
	prometheus.MustRegister(metric.underVoltageMetric)
	prometheus.MustRegister(metric.overheatedMetric)

	return metric
}

func (m *baseHomematicMetric) update(device hmip.Device, labels prometheus.Labels) {
	m.lastSeenMetric.With(labels).Set(float64(device.GetLastUpdated().Unix()))
	for _, base := range device.GetFunctionalChannelsByType(hmip.CHANNEL_TYPE_DEVICE_BASE) {
		switch channel := base.(type) {
		case hmip.BaseDeviceChannel:
			var value float64 = 1
			if channel.IsUnreached() {
				value = 0
			}
			m.reachableMetric.With(labels).Set(value)

			value = 0
			if channel.IsOverheated() {
				value = 1
			}
			m.overheatedMetric.With(labels).Set(value)

			value = 0
			if channel.HasLowBattery() {
				value = 1
			}
			m.batteryLowMetric.With(labels).Set(value)

			value = 0
			if channel.HasUnderVoltage() {
				value = 1
			}
			m.underVoltageMetric.With(labels).Set(value)

			m.connectionQualityMetric.With(labels).Set(float64(channel.GetRSSIValue()))
		}
	}
}
