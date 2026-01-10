package homematic

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/salex-org/hmip-exporter/internal/util"
	"github.com/salex-org/hmip-go-client/pkg/hmip"
)

type HomematicClient interface {
	Start() error
	Shutdown() error
	Health() error
}

type homemeticClient struct {
	hmipClient        hmip.Homematic
	baseMetrics       homematicMetric
	additionalMetrics map[string]homematicMetric // key: device type
	processingError   error
}

type homematicMetric interface {
	update(device hmip.Device, labels prometheus.Labels)
}

func NewHomematicClient() (HomematicClient, error) {
	// Create client
	hmipConfig, err := hmip.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get HomematicIP client configuration: %s", err)
	}
	hmipConfig.AccessPointSGTIN = util.ReadEnvVar("HMIP_AP_SGTIN")
	hmipConfig.DeviceID = util.ReadEnvVar("HMIP_DEVICE_ID")
	hmipConfig.ClientID = util.ReadEnvVar("HMIP_CLIENT_ID")
	hmipConfig.ClientName = util.ReadEnvVar("HMIP_CLIENT_NAME")
	hmipConfig.ClientAuthToken = util.ReadEnvVar("HMIP_CLIENT_AUTH_TOKEN")
	hmipConfig.AuthToken = util.ReadEnvVar("HMIP_AUTH_TOKEN")
	switchMetric := newSwitchMetric()
	newClient := &homemeticClient{
		baseMetrics: newBaseHomematicMetric(),
		additionalMetrics: map[string]homematicMetric{
			"REMOTE_CONTROL_8": newRemoteControlMetric(),
			hmip.DEVICE_TYPE_TEMPERATURE_HUMIDITY_SENSOR_OUTDOOR: newClimateSensorMetric(),
			hmip.DEVICE_TYPE_PLUGABLE_SWITCH:                     switchMetric,
			hmip.DEVICE_TYPE_PLUGABLE_SWITCH_MEASURING:           switchMetric,
			hmip.DEVICE_TYPE_SMOKE_DETECTOR:                      newSmokeDetectorMetric(),
		},
	}
	newClient.hmipClient, err = hmip.GetClientWithConfig(hmipConfig)
	if err != nil {
		return nil, fmt.Errorf("could not get HomematicIP client: %s", err)
	}

	// TODO Register event handler

	// Load initial data
	state, err := newClient.hmipClient.LoadCurrentState()
	if err != nil {
		return nil, fmt.Errorf("could not load HomematicIP state: %s", err)
	}
	for _, each := range state.GetDevices() {
		newClient.updateMetric(each)
	}

	return newClient, nil
}

func (h *homemeticClient) Start() error {
	// TODO implement event listening
	return nil
}

func (h *homemeticClient) Shutdown() error {
	return h.hmipClient.StopEventListening()
}

func (h *homemeticClient) Health() error {
	if h.processingError != nil {
		return h.processingError
	}
	return h.hmipClient.GetEventLoopState()
}

func (h *homemeticClient) updateMetric(device hmip.Device) {
	if device.GetType() == hmip.DEVICE_TYPE_HOME_CONTROL_ACCESS_POINT {
		return
	}

	if metric, metricFound := h.additionalMetrics[device.GetType()]; metricFound {
		labels := h.createLabels(device)
		h.baseMetrics.update(device, labels)
		metric.update(device, labels)

		return
	}
	fmt.Printf("Warning: No metric registered for %s\n", device.GetType())
}

// TODO add room information to labels

var metricLabelNames = []string{"device_id", "device_name", "device_type", "device_model"}

func (h *homemeticClient) createLabels(device hmip.Device) prometheus.Labels {
	return prometheus.Labels{
		"device_id":    device.GetID(),
		"device_name":  device.GetName(),
		"device_type":  device.GetType(),
		"device_model": device.GetModel(),
	}
}
