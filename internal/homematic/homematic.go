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
	metaGroups        map[string]hmip.Group      // key: group id
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
		metaGroups:  make(map[string]hmip.Group),
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

	// Register event handler
	newClient.hmipClient.RegisterEventHandler(func(baseEvent hmip.Event, _ hmip.Origin) {
		switch event := baseEvent.(type) {
		case hmip.DeviceChangedEvent:
			newClient.updateMetric(event.GetDevice())
		case hmip.GroupChangedEvent:
			newClient.updateMetaGroup(event.GetGroup())
		}
	}, hmip.EVENT_TYPE_DEVICE_CHANGED, hmip.EVENT_TYPE_GROUP_CHANGED)

	return newClient, nil
}

func (h *homemeticClient) Start() error {
	// Load initial data
	state, err := h.hmipClient.LoadCurrentState()
	if err == nil {
		fmt.Println("Loading initial state succeeded")
		for _, each := range state.GetGroups() {
			h.updateMetaGroup(each)
		}
		for _, each := range state.GetDevices() {
			h.updateMetric(each)
		}
	} else {
		h.processingError = fmt.Errorf("could not load initial state: %s", err)
		return h.processingError
	}

	// Start the event listening
	err = h.hmipClient.ListenForEvents()
	if err != nil {
		h.processingError = fmt.Errorf("could not start event listening: %s", err)
		return h.processingError
	}

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

var metricLabelNames = []string{"device_id", "device_name", "device_type", "device_model", "room_id", "room_name"}

func (h *homemeticClient) createLabels(device hmip.Device) prometheus.Labels {
	var roomID = ""
	var roomName = ""
	for _, base := range device.GetFunctionalChannelsByType(hmip.CHANNEL_TYPE_DEVICE_BASE) {
		switch channel := base.(type) {
		case hmip.BaseDeviceChannel:
			metaGroup := h.getMetaGroupFromChannel(channel)
			if metaGroup != nil {
				roomID = metaGroup.GetID()
				roomName = metaGroup.GetName()
			}
		}
	}

	return prometheus.Labels{
		"device_id":    device.GetID(),
		"device_name":  device.GetName(),
		"device_type":  device.GetType(),
		"device_model": device.GetModel(),
		"room_id":      roomID,
		"room_name":    roomName,
	}
}

func (h *homemeticClient) getMetaGroupFromChannel(channel hmip.BaseDeviceChannel) hmip.Group {
	for _, groupID := range channel.GetGroups() {
		if group, hasGroup := h.metaGroups[groupID]; hasGroup {
			return group
		}
	}

	return nil
}

func (h *homemeticClient) updateMetaGroup(group hmip.Group) {
	if group.GetType() == hmip.GROUP_TYPE_META {
		h.metaGroups[group.GetID()] = group
	}
}
