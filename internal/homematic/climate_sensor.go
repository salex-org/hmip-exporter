package homematic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

type climateSensorMetric struct {
	temperatureMetric *prometheus.GaugeVec
	humidityMetric    *prometheus.GaugeVec
}

func newClimateSensorMetric() homematicMetric {
	metric := &climateSensorMetric{
		temperatureMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "environment_sensor",
			Name:      "current_temperature",
			Help:      "Current temperature measured by an environment sensor (degree celsius)",
		}, metricLabelNames),
		humidityMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "hmip",
			Subsystem: "environment_sensor",
			Name:      "current_humidity",
			Help:      "Current relative humidity measured by an environment sensor (percent)",
		}, metricLabelNames),
	}
	prometheus.MustRegister(metric.temperatureMetric)
	prometheus.MustRegister(metric.humidityMetric)

	return metric
}

func (m *climateSensorMetric) update(device hmip.Device, labels prometheus.Labels) {
	for _, base := range device.GetFunctionalChannels() {
		switch channel := base.(type) {
		case hmip.ClimateSensorChannel:
			m.temperatureMetric.With(labels).Set(channel.GetActualTemperature())
			m.humidityMetric.With(labels).Set(float64(channel.GetHumidity()))
		}
	}
}
