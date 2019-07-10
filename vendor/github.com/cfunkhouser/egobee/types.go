package egobee

import (
	"fmt"
	"strconv"
)

// This file contains types for the ecobee v1 API as defined in the ecobee
// developer documentation.

// Action to take when a SensorState is triggered.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Action.shtml
type Action struct {
	Type              string `json:"type"`
	SendAlert         bool   `json:"sendAlert"`
	SendUpdate        bool   `json:"sendUpdate"`
	ActivationDelay   int    `json:"activationDelay"`
	DeactivationDelay int    `json:"deactivationDelay"`
	MinActionDuration int    `json:"minActionDuration"`
	HeatAdjustTemp    int    `json:"heatAdjustTemp"`
	CoolAdjustTemp    int    `json:"coolAdjustTemp"`
	ActivateRelay     string `json:"activateRelay"`
	ActivateRelayOpen bool   `json:"activateRelayOpen"`
}

// Alert generated either by a thermostat or user which requires user attention.
// It may be an error, or a reminder for a filter change. Alerts may not be
// modified directly but rather they must be acknowledged using the Acknowledge
// function.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Alert.shtml
type Alert struct {
	AcknowledgeRef       string `json:"acknowledgeRef"`
	Date                 string `json:"date"`
	Time                 string `json:"time"`
	Severity             string `json:"severity"`
	Text                 string `json:"text"`
	AlertNumber          int    `json:"alertNumber"`
	AlertType            string `json:"alertType"`
	IsOperatorAlert      bool   `json:"isOperatorAlert"`
	Reminder             string `json:"reminder"`
	ShowIDT              bool   `json:"showIdt"`
	ShowWeb              bool   `json:"showWeb"`
	SendEmail            bool   `json:"sendEmail"`
	Acknowledgement      string `json:"acknowledgement"`
	RemindMeLater        bool   `json:"remindMeLater"`
	ThermostatIdentifier string `json:"thermostatIdentifier"`
	NotificationType     string `json:"notificationType"`
}

// Audio properties of a thermostat. Only applicable to ecobee4.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Audio.shtml
type Audio struct {
	PlaybackVolume    int           `json:"playbackVolume"`
	MicrophoneEnabled bool          `json:"microphoneEnabled"`
	SoundAlertVolume  int           `json:"soundAlertVolume"`
	SoundTickVolume   int           `json:"soundTickVolume"`
	VoiceEngines      []VoiceEngine `json:"voiceEngines"`
}

// Climate defines a thermostat settings template which is then applied to
// individual period cells of the Schedule. The result is that if you modify the
// Climate, all schedule cells which reference that Climate will automatically
// be changed.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Climate.shtml
type Climate struct {
	Name                string         `json:"name"`
	ClimateRef          string         `json:"climateRef"`
	IsOccupied          bool           `json:"isOccupied"`
	IsOptimized         bool           `json:"isOptimized"`
	CoolFan             string         `json:"coolFan"`
	HeatFan             string         `json:"heatFan"`
	Vent                string         `json:"vent"`
	VentilatorMinOnTime int            `json:"ventilatorMinOnTime"`
	Owner               string         `json:"owner"`
	Type                string         `json:"type"`
	Colour              int            `json:"colour"`
	CoolTemp            int            `json:"coolTemp"`
	HeatTemp            int            `json:"heatTemp"`
	Sensors             []RemoteSensor `json:"sensors"`
}

// Device attached to a thermostat. Devices may not be modified remotely; all
// changes must occur on the thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Device.shtml
type Device struct {
	DeviceID int      `json:"deviceId"`
	Name     string   `json:"name"`
	Sensors  []Sensor `json:"sensors"`
	Outputs  []Output `json:"outputs"`
}

// Electricity contains the last collected electricity usage measurements for
// the thermostat. An electricity object is composed of Electricity Devices,
// each of which contains readings from an Electricity Tier.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Electricity.shtml
type Electricity struct {
	Devices []ElectricityDevice `json:"devices"`
}

// ElectricityDevice represents an energy recording device. At this time, only
// meters are supported by the API.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/ElectricityDevice.shtml
type ElectricityDevice struct {
	Name        string            `json:"name"`
	Tiers       []ElectricityTier `json:"tiers"`
	LastUpdate  string            `json:"lastUpdate"`
	Cost        string            `json:"cost"`
	Consumption string            `json:"consumption"`
}

// ElectricityTier epresents the last reading from a given pricing tier if the
// utility provides such information. If there are no pricing tiers defined,
// than an unnamed tier will represent the total reading. The values represented
// here are a daily cumulative total in kWh. The cost is likewise a cumulative
// total in cents.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/ElectricityTier.shtml
type ElectricityTier struct {
	Name        string `json:"name"`
	Consumption string `json:"consumption"`
	Cost        string `json:"cost"`
}

// Energy is undocumented on the ecobee website.
// TODO(cfunkhouser): reverse engineer this struct.
type Energy struct{}

// EquipmentSetting represents the alert/reminder type which is associated with
// and dependent upon specific equipment controlled by the Thermostat. It is
// used when getting/setting the Thermostat NotificationSettings object.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/EquipmentSetting.shtml
type EquipmentSetting struct {
	FilterLastChanged string `json:"filterLastChanged"`
	FilterLife        int    `json:"filterLife"`
	FilterLifeUnits   string `json:"filterLifeUnits"`
	RemindMeDate      string `json:"remindMeDate"`
	Enabled           bool   `json:"enabled"`
	Type              string `json:"type"`
	RemindTechnician  bool   `json:"remindTechnician"`
}

// Event is a scheduled thermostat program change. All events have a start and
// end time during which the thermostat runtime settings will be modified.
// Events may not be directly modified, various Functions provide the capability
// to modify the calendar events and to modify the program. The event list is
// sorted with events ordered by whether they are currently running and the
// internal priority of each event. It is safe to take the first event which is
// running and show it as the currently running event. When the resume function
// is used, events are removed in the order they are listed here.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Event.shtml
type Event struct {
	Type                   string `json:"type"`
	Name                   string `json:"name"`
	Running                bool   `json:"running"`
	StartDate              string `json:"startDate"`
	StartTime              string `json:"startTime"`
	EndDate                string `json:"endDate"`
	EndTime                string `json:"endTime"`
	IsOccupied             bool   `json:"isOccupied"`
	IsCoolOff              bool   `json:"isCoolOff"`
	IsHeatOff              bool   `json:"isHeatOff"`
	CoolHoldTemp           int    `json:"coolHoldTemp"`
	HeatHoldTemp           int    `json:"heatHoldTemp"`
	Fan                    string `json:"fan"`
	Vent                   string `json:"vent"`
	VentilatorMinOnTime    int    `json:"ventilatorMinOnTime"`
	IsOptional             bool   `json:"isOptional"`
	IsTemperatureRelative  bool   `json:"isTemperatureRelative"`
	CoolRelativeTemp       int    `json:"coolRelativeTemp"`
	HeatRelativeTemp       int    `json:"heatRelativeTemp"`
	IsTemperatureAbsolute  bool   `json:"isTemperatureAbsolute"`
	DutyCyclePercentage    int    `json:"dutyCyclePercentage"`
	FanMinOnTime           int    `json:"fanMinOnTime"`
	OccupiedSensorActive   bool   `json:"occupiedSensorActive"`
	UnoccupiedSensorActive bool   `json:"unoccupiedSensorActive"`
	DRRampUpTemp           int    `json:"drRampUpTemp"`
	DRRampUpTime           int    `json:"drRampUpTime"`
	LinkRef                string `json:"linkRef"`
	HoldClimateRef         string `json:"holdClimateRef"`
}

// ExtendedRuntime contains the last three 5 minute interval values sent by the
// thermostat for the past 15 minutes of runtime. The interval values are
// valuable when you are interested in analyzing the runtime data in a more
// granular fashion, at 5 minute increments rather than the more general 15
// minute value from the Runtime Object.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/ExtendedRuntime.shtml
type ExtendedRuntime struct {
	LastReadingTimestamp     string `json:"lastReadingTimestamp"`
	RuntimeDate              string `json:"runtimeDate"`
	RuntimeInterval          int    `json:"runtimeInterval"`
	ActualTemperature        int    `json:"actualTemperature"`
	ActualHumidity           int    `json:"actualHumidity"`
	DesiredHeat              int    `json:"desiredHeat"`
	DesiredCool              int    `json:"desiredCool"`
	DesiredHumidity          int    `json:"desiredHumidity"`
	DesiredDehumidity        int    `json:"desiredDehumidity"`
	DMOffset                 int    `json:"dmOffset"`
	HVACMode                 string `json:"hvacMode"`
	HeatPump1                int    `json:"heatPump1"`
	HeatPump2                int    `json:"heatPump2"`
	AuxHeat1                 int    `json:"auxHeat1"`
	AuxHeat2                 int    `json:"auxHeat2"`
	AuxHeat3                 int    `json:"auxHeat3"`
	Cool1                    int    `json:"cool1"`
	Cool2                    int    `json:"cool2"`
	Fan                      int    `json:"fan"`
	Humidifier               int    `json:"humidifier"`
	Dehumidifier             int    `json:"dehumidifier"`
	Economizer               int    `json:"economizer"`
	Ventilator               int    `json:"ventilator"`
	CurrentElectricityBill   int    `json:"currentElectricityBill"`
	ProjectedElectricityBill int    `json:"projectedElectricityBill"`
}

// GeneralSetting represent the General alert/reminder type. It is used when
// getting/setting the Thermostat NotificationSettings object.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/GeneralSetting.shtml
type GeneralSetting struct {
	Enabled          bool   `json:"enabled"`
	Type             string `json:"type"`
	RemindTechnician bool   `json:"remindTechnician"`
}

// HouseDetails contains contains the information about the house the thermostat
// is installed in.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/HouseDetails.shtml
type HouseDetails struct {
	Style             string `json:"style"`
	Size              int    `json:"size"`
	NumberOfFloors    int    `json:"numberOfFloors"`
	NumberOfRooms     int    `json:"numberOfRooms"`
	NumberOfOccupants int    `json:"numberOfOccupants"`
	Age               int    `json:"age"`
	WindowEfficiency  int    `json:"windowEfficiency"`
}

// LimitSetting represents the alert/reminder type which is associated specific
// values, such as highHeat or lowHumidity. It is used when getting/setting the
// Thermostat NotificationSettings object.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/LimitSetting.shtml
type LimitSetting struct {
	Limit            int    `json:"limit"`
	Enabled          bool   `json:"enabled"`
	Type             string `json:"type"`
	RemindTechnician bool   `json:"remindTechnician"`
}

// Location describes the physical location and coordinates of the thermostat as entered by the thermostat owner. The
// address information is used in a geocode look up to obtain the thermostat
// coordinates. The coordinates are used to obtain accurate weather information.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Location.shtml
type Location struct {
	TimeZoneOffsetMinutes int    `json:"timeZoneOffsetMinutes"`
	TimeZone              string `json:"timeZone"`
	IsDaylightSaving      bool   `json:"isDaylightSaving"`
	StreetAddress         string `json:"streetAddress"`
	City                  string `json:"city"`
	ProvinceState         string `json:"provinceState"`
	Country               string `json:"country"`
	PostalCode            string `json:"postalCode"`
	PhoneNumber           string `json:"phoneNumber"`
	MapCoordinates        string `json:"mapCoordinates"`
}

// Management contains information about the management company the thermostat
// belongs to. The Management object is read-only, it may be modified in the web
// portal.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Management.shtml
type Management struct {
	AdministrativeContact string `json:"administrativeContact"`
	BillingContact        string `json:"billingContact"`
	Name                  string `json:"name"`
	Phone                 string `json:"phone"`
	Email                 string `json:"email"`
	Web                   string `json:"web"`
	ShowAlertIdt          bool   `json:"showAlertIdt"`
	ShowAlertWeb          bool   `json:"showAlertWeb"`
}

// NotificationSettings is a container for the configuration of the possible
// alerts and reminders which can be generated by the Thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/NotificationSettings.shtml
type NotificationSettings struct {
	EmailAddresses            string             `json:"emailAddresses"`
	EmailNotificationsEnabled bool               `json:"emailNotificationsEnabled"`
	Equipment                 []EquipmentSetting `json:"equipment"`
	General                   []GeneralSetting   `json:"general"`
	Limit                     []LimitSetting     `json:"limit"`
}

// Output is a relay connected to the thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Output.shtml
type Output struct {
	Name             string `json:"name"`
	Zone             int    `json:"zone"`
	OutputID         int    `json:"outputId"`
	Type             string `json:"type"`
	SendUpdate       bool   `json:"sendUpdate"`
	ActiveClosed     bool   `json:"activeClosed"`
	ActivationTime   int    `json:"activationTime"`
	DeactivationTime int    `json:"deactivationTime"`
}

// Program is a container for the Schedule and its Climates.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Program.shtml
type Program struct {
	Schedule          [][]string `json:"schedule"`
	Climates          []Climate  `json:"climates"`
	CurrentClimateRef string     `json:"currentClimateRef"`
}

// Common remote sensor capability IDs.
const (
	CapabilityTypeOccupancy   = "occupancy"
	CapabilityTypeTemperature = "temperature"
)

// RemoteSensor represents a sensor connected to the thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/RemoteSensor.shtml
type RemoteSensor struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Type       string                   `json:"type"`
	Code       string                   `json:"code"`
	InUse      bool                     `json:"inUse"`
	Capability []RemoteSensorCapability `json:"capability"`
}

// Temperature gets the temperature for the sensor if it exists.
func (s *RemoteSensor) Temperature() (float64, error) {
	if s != nil && len(s.Capability) > 0 {
		for _, c := range s.Capability {
			if c.Type == CapabilityTypeTemperature {
				v, err := strconv.ParseFloat(c.Value, 64)
				if err != nil {
					return 0.0, err
				}
				return float64(v / 10), nil
			}
		}
	}
	return 0.0, fmt.Errorf("remote sensor %v does not have a temperature capability", s.Name)
}

// RemoteSensorCapability represents the specific capability of a sensor
// connected to the thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/RemoteSensorCapability.shtml
type RemoteSensorCapability struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Runtime epresents the last known thermostat running state. This state is
// composed from the last interval status message received from a thermostat.
// It is also updated each time the thermostat posts configuration changes to
// the server.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Runtime.shtml
type Runtime struct {
	RuntimeRev         string `json:"runtimeRev"`
	Connected          bool   `json:"connected"`
	FirstConnected     string `json:"firstConnected"`
	ConnectDateTime    string `json:"connectDateTime"`
	DisconnectDateTime string `json:"disconnectDateTime"`
	LastModified       string `json:"lastModified"`
	LastStatusModified string `json:"lastStatusModified"`
	RuntimeDate        string `json:"runtimeDate"`
	RuntimeInterval    int    `json:"runtimeInterval"`
	ActualTemperature  int    `json:"actualTemperature"`
	ActualHumidity     int    `json:"actualHumidity"`
	DesiredHeat        int    `json:"desiredHeat"`
	DesiredCool        int    `json:"desiredCool"`
	DesiredHumidity    int    `json:"desiredHumidity"`
	DesiredDehumidity  int    `json:"desiredDehumidity"`
	DesiredFanMode     string `json:"desiredFanMode"`
	DesiredHeatRange   []int  `json:"desiredHeatRange"`
	DesiredCoolRange   []int  `json:"desiredCoolRange"`
}

// SecuritySettings defines the security settings which a thermostat may have.
// Currently this object stores data specific to access control.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/SecuritySettings.shtml
type SecuritySettings struct {
	UserAccessCode  string `json:"userAccessCode"`
	AllUserAccess   bool   `json:"allUserAccess"`
	ProgramAccess   bool   `json:"programAccess"`
	DetailsAccess   bool   `json:"detailsAccess"`
	QuickSaveAccess bool   `json:"quickSaveAccess"`
	VacationAccess  bool   `json:"vacationAccess"`
}

// SelectionType defines the type of selection to perform.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Selection.shtml.
type SelectionType string

var (
	// SelectionTypeRegistered returns Thermostats registered to the current user
	// Only usable with Smart thermostats, does not work on EMS thermostats and
	// may not be used by a Utility who is not the owner of thermostats.
	SelectionTypeRegistered SelectionType = "registered"

	// SelectionTypeThermostats selects only those thermostats listed in the CSV
	// SelectionMatch. No spaces in the CSV string. There is a limit of 25
	// identifiers per request.
	SelectionTypeThermostats SelectionType = "thermostats"

	// SelectionTypeManagementSet selects all thermostats for a given management
	// set defined by the Management/Utility account. This is only available to
	// Management/Utility accounts.
	SelectionTypeManagementSet SelectionType = "managementSet"
)

// Selection defines the resources and information to return as part of a
// response. It is required in all requests, but some selection fields are only
// meaningful in certain request types.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Selection.shtml
type Selection struct {
	SelectionType               SelectionType `json:"selectionType,omitempty"`
	SelectionMatch              string        `json:"selectionMatch"`
	IncludeRuntime              bool          `json:"includeRuntime,omitempty"`
	IncludeExtendedRuntime      bool          `json:"includeExtendedRuntime,omitempty"`
	IncludeElectricity          bool          `json:"includeElectricity,omitempty"`
	IncludeSettings             bool          `json:"includeSettings,omitempty"`
	IncludeLocation             bool          `json:"includeLocation,omitempty"`
	IncludeProgram              bool          `json:"includeProgram,omitempty"`
	IncludeEvents               bool          `json:"includeEvents,omitempty"`
	IncludeDevice               bool          `json:"includeDevice,omitempty"`
	IncludeTechnician           bool          `json:"includeTechnician,omitempty"`
	IncludeUtility              bool          `json:"includeUtility,omitempty"`
	IncludeManagement           bool          `json:"includeManagement,omitempty"`
	IncludeAlerts               bool          `json:"includeAlerts,omitempty"`
	IncludeReminders            bool          `json:"includeReminders,omitempty"`
	IncludeWeather              bool          `json:"includeWeather,omitempty"`
	IncludeHouseDetails         bool          `json:"includeHouseDetails,omitempty"`
	IncludeOemCfg               bool          `json:"includeOemCfg,omitempty"`
	IncludeEquipmentStatus      bool          `json:"includeEquipmentStatus,omitempty"`
	IncludeNotificationSettings bool          `json:"includeNotificationSettings,omitempty"`
	IncludePrivacy              bool          `json:"includePrivacy,omitempty"`
	IncludeVersion              bool          `json:"includeVersion,omitempty"`
	IncludeSecuritySettings     bool          `json:"includeSecuritySettings,omitempty"`
	IncludeSensors              bool          `json:"includeSensors,omitempty"`
	IncludeAudio                bool          `json:"includeAudio,omitempty"`
	IncludeEnergy               bool          `json:"includeEnergy,omitempty"`
}

// Sensor represents a sensor connected to the thermostat. Sensors may not be
// modified using the API, however some configuration may occur through the web
// portal.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Sensor.shtml
type Sensor struct {
	Name           string  `json:"name"`
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	Zone           int     `json:"zone"`
	SensorID       int     `json:"sensorId"`
	Type           string  `json:"type"`
	Usage          string  `json:"usage"`
	NumberOfBits   int     `json:"numberOfBits"`
	BConstant      int     `json:"bconstant"`
	ThermistorSize int     `json:"thermistorSize"`
	TempCorrection int     `json:"tempCorrection"`
	Gain           int     `json:"gain"`
	MaxVoltage     int     `json:"maxVoltage"`
	Multiplier     int     `json:"multiplier"`
	States         []State `json:"states"`
}

// Settings contains all the configuration properties of a Thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Settings.shtml
type Settings struct {
	HVACMode                            string `json:"hvacMode"`
	LastServiceDate                     string `json:"lastServiceDate"`
	ServiceRemindMe                     bool   `json:"serviceRemindMe"`
	MonthsBetweenService                int    `json:"monthsBetweenService"`
	RemindMeDate                        string `json:"remindMeDate"`
	Vent                                string `json:"vent"`
	VentilatorMinOnTime                 int    `json:"ventilatorMinOnTime"`
	ServiceRemindTechnician             bool   `json:"serviceRemindTechnician"`
	EILocation                          string `json:"eiLocation"`
	ColdTempAlert                       int    `json:"coldTempAlert"`
	ColdTempAlertEnabled                bool   `json:"coldTempAlertEnabled"`
	HotTempAlert                        int    `json:"hotTempAlert"`
	HotTempAlertEnabled                 bool   `json:"hotTempAlertEnabled"`
	CoolStages                          int    `json:"coolStages"`
	HeatStages                          int    `json:"heatStages"`
	MaxSetBack                          int    `json:"maxSetBack"`
	MaxSetForward                       int    `json:"maxSetForward"`
	QuickSaveSetBack                    int    `json:"quickSaveSetBack"`
	QuickSaveSetForward                 int    `json:"quickSaveSetForward"`
	HasHeatPump                         bool   `json:"hasHeatPump"`
	HasForcedAir                        bool   `json:"hasForcedAir"`
	HasBoiler                           bool   `json:"hasBoiler"`
	HasHumidifier                       bool   `json:"hasHumidifier"`
	HasERV                              bool   `json:"hasErv"`
	HasHRV                              bool   `json:"hasHrv"`
	CondensationAvoid                   bool   `json:"condensationAvoid"`
	UseCelsius                          bool   `json:"useCelsius"`
	UseTimeFormat12                     bool   `json:"useTimeFormat12"`
	Locale                              string `json:"locale"`
	Humidity                            string `json:"humidity"`
	HumidifierMode                      string `json:"humidifierMode"`
	BacklightOnIntensity                int    `json:"backlightOnIntensity"`
	BacklightSleepIntensity             int    `json:"backlightSleepIntensity"`
	BacklightOffTime                    int    `json:"backlightOffTime"`
	SoundTickVolume                     int    `json:"soundTickVolume"`
	SoundAlertVolume                    int    `json:"soundAlertVolume"`
	CompressorProtectionMinTime         int    `json:"compressorProtectionMinTime"`
	CompressorProtectionMinTemp         int    `json:"compressorProtectionMinTemp"`
	Stage1HeatingDifferentialTemp       int    `json:"stage1HeatingDifferentialTemp"`
	Stage1CoolingDifferentialTemp       int    `json:"stage1CoolingDifferentialTemp"`
	Stage1HeatingDissipationTime        int    `json:"stage1HeatingDissipationTime"`
	Stage1CoolingDissipationTime        int    `json:"stage1CoolingDissipationTime"`
	HeatPumpReversalOnCool              bool   `json:"heatPumpReversalOnCool"`
	FanControlRequired                  bool   `json:"fanControlRequired"`
	FanMinOnTime                        int    `json:"fanMinOnTime"`
	HeatCoolMinDelta                    int    `json:"heatCoolMinDelta"`
	TempCorrection                      int    `json:"tempCorrection"`
	HoldAction                          string `json:"holdAction"`
	HeatPumpGroundWater                 bool   `json:"heatPumpGroundWater"`
	HasElectric                         bool   `json:"hasElectric"`
	HasDehumidifier                     bool   `json:"hasDehumidifier"`
	DehumidifierMode                    string `json:"dehumidifierMode"`
	DehumidifierLevel                   int    `json:"dehumidifierLevel"`
	DehumidifyWithAC                    bool   `json:"dehumidifyWithAC"`
	DehumidifyOvercoolOffset            int    `json:"dehumidifyOvercoolOffset"`
	AutoHeatCoolFeatureEnabled          bool   `json:"autoHeatCoolFeatureEnabled"`
	WifiOfflineAlert                    bool   `json:"wifiOfflineAlert"`
	HeatMinTemp                         int    `json:"heatMinTemp"`
	HeatMaxTemp                         int    `json:"heatMaxTemp"`
	CoolMinTemp                         int    `json:"coolMinTemp"`
	CoolMaxTemp                         int    `json:"coolMaxTemp"`
	HeatRangeHigh                       int    `json:"heatRangeHigh"`
	HeatRangeLow                        int    `json:"heatRangeLow"`
	CoolRangeHigh                       int    `json:"coolRangeHigh"`
	CoolRangeLow                        int    `json:"coolRangeLow"`
	UserAccessCode                      string `json:"userAccessCode"`
	UserAccessSetting                   int    `json:"userAccessSetting"`
	AuxRuntimeAlert                     int    `json:"auxRuntimeAlert"`
	AuxOutdoorTempAlert                 int    `json:"auxOutdoorTempAlert"`
	AuxMaxOutdoorTemp                   int    `json:"auxMaxOutdoorTemp"`
	AuxRuntimeAlertNotify               bool   `json:"auxRuntimeAlertNotify"`
	AuxOutdoorTempAlertNotify           bool   `json:"auxOutdoorTempAlertNotify"`
	AuxRuntimeAlertNotifyTechnician     bool   `json:"auxRuntimeAlertNotifyTechnician"`
	AuxOutdoorTempAlertNotifyTechnician bool   `json:"auxOutdoorTempAlertNotifyTechnician"`
	DisablePreHeating                   bool   `json:"disablePreHeating"`
	DisablePreCooling                   bool   `json:"disablePreCooling"`
	InstallerCodeRequired               bool   `json:"installerCodeRequired"`
	DRAccept                            string `json:"drAccept"`
	IsRentalProperty                    bool   `json:"isRentalProperty"`
	UseZoneController                   bool   `json:"useZoneController"`
	RandomStartDelayCool                int    `json:"randomStartDelayCool"`
	RandomStartDelayHeat                int    `json:"randomStartDelayHeat"`
	HumidityHighAlert                   int    `json:"humidityHighAlert"`
	HumidityLowAlert                    int    `json:"humidityLowAlert"`
	DisableHeatPumpAlerts               bool   `json:"disableHeatPumpAlerts"`
	DisableAlertsOnIdt                  bool   `json:"disableAlertsOnIdt"`
	HumidityAlertNotify                 bool   `json:"humidityAlertNotify"`
	HumidityAlertNotifyTechnician       bool   `json:"humidityAlertNotifyTechnician"`
	TempAlertNotify                     bool   `json:"tempAlertNotify"`
	TempAlertNotifyTechnician           bool   `json:"tempAlertNotifyTechnician"`
	MonthlyElectricityBillLimit         int    `json:"monthlyElectricityBillLimit"`
	EnableElectricityBillAlert          bool   `json:"enableElectricityBillAlert"`
	EnableProjectedElectricityBillAlert bool   `json:"enableProjectedElectricityBillAlert"`
	ElectricityBillingDayOfMonth        int    `json:"electricityBillingDayOfMonth"`
	ElectricityBillCycleMonths          int    `json:"electricityBillCycleMonths"`
	ElectricityBillStartMonth           int    `json:"electricityBillStartMonth"`
	VentilatorMinOnTimeHome             int    `json:"ventilatorMinOnTimeHome"`
	VentilatorMinOnTimeAway             int    `json:"ventilatorMinOnTimeAway"`
	BacklightOffDuringSleep             bool   `json:"backlightOffDuringSleep"`
	AutoAway                            bool   `json:"autoAway"`
	SmartCirculation                    bool   `json:"smartCirculation"`
	FollowMeComfort                     bool   `json:"followMeComfort"`
	VentilatorType                      string `json:"ventilatorType"`
	IsVentilatorTimerOn                 bool   `json:"isVentilatorTimerOn"`
	VentilatorOffDateTime               string `json:"ventilatorOffDateTime"`
	HasUVFilter                         bool   `json:"hasUVFilter"`
	CoolingLockout                      bool   `json:"coolingLockout"`
	VentilatorFreeCooling               bool   `json:"ventilatorFreeCooling"`
	DehumidifyWhenHeating               bool   `json:"dehumidifyWhenHeating"`
	VentilatorDehumidify                bool   `json:"ventilatorDehumidify"`
	GroupRef                            string `json:"groupRef"`
	GroupName                           string `json:"groupName"`
	GroupSetting                        int    `json:"groupSetting"`
}

// State is a configurable trigger for a number of SensorActions.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/State.shtml
type State struct {
	MaxValue int      `json:"maxValue"`
	MinValue int      `json:"minValue"`
	Type     string   `json:"type"`
	Actions  []Action `json:"actions"`
}

// Technician associated with a thermostat. may not be modified through the API.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Technician.shtml
type Technician struct{}

// Thermostat is the central piece of the ecobee API. All objects relate in one
// way or another to a real thermostat. The thermostat object and its component
// objects define the real thermostat device.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Thermostat.shtml
type Thermostat struct {
	Alerts              []Alert               `json:"alerts"`
	Audio               Audio                 `json:"audio"`
	Brand               string                `json:"brand"`
	Devices             []Device              `json:"devices"`
	Electricity         Electricity           `json:"electricity"`
	Energy              Energy                `json:"energy"`
	EquipmentStatus     string                `json:"equipmentStatus"`
	Events              []Event               `json:"events"`
	ExtendedRuntime     ExtendedRuntime       `json:"extendedRuntime"`
	Features            string                `json:"features"`
	HouseDetails        HouseDetails          `json:"houseDetails"`
	Identifier          string                `json:"identifier"`
	IsRegistered        bool                  `json:"isRegistered"`
	LastModified        string                `json:"lastModified"`
	Location            Location              `json:"location"`
	Management          Management            `json:"management"`
	ModelNumber         string                `json:"modelNumber"`
	Name                string                `json:"name"`
	NotifictionSettings NotificationSettings  `json:"notificationSettings"`
	Program             Program               `json:"program"`
	Reminders           []ThermostatReminder2 `json:"reminders"`
	RemoteSensors       []RemoteSensor        `json:"remoteSensors"`
	Runtime             Runtime               `json:"runtime"`
	SecuritySettings    SecuritySettings      `json:"securitySettings"`
	Settings            Settings              `json:"settings"`
	Technician          Technician            `json:"technician"`
	ThermostatRev       string                `json:"thermostatRev"`
	ThermostatTime      string                `json:"thermostatTime"`
	UTCTime             string                `json:"utcTime"`
	Utility             Utility               `json:"utility"`
	Version             Version               `json:"version"`
	Weather             Weather               `json:"weather"`

	// Privacy ... `json:"privacy"`
	// OEMCfg ... `json:"oemCfg"`
}

// ThermostatReminder2 is not documented.
// TODO(cfunkhouser): reverse engineer this struct.
type ThermostatReminder2 struct{}

// ThermostatSummary describes Thermostats and their status according to the
// API.
// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostat-summary.shtml
type ThermostatSummary struct {
	RevisionList    []string `json:"revisionList,omitempty"`
	ThermostatCount int      `json:"thermostatCount,omitempty"`
	StatusList      []string `json:"statusList,omitempty"`
	Status          struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"status,omitempty"`
}

// Utility the Thermostat belongs to. May not be modified.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Utility.shtml
type Utility struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Web   string `json:"web"`
}

// Version of a Thermostat.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Version.shtml
type Version struct {
	ThermostatFirmwareVersion string `json:"thermostatFirmwareVersion"`
}

// VoiceEngine contains information about the voice assistant that the selected
// thermostat supports.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/VoiceEngine.shtml
type VoiceEngine struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Weather and forecast information for a Thermostat's location.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/Weather.shtml
type Weather struct {
	Timestamp      string            `json:"timestamp"`
	WeatherStation string            `json:"weatherStation"`
	Forecasts      []WeatherForecast `json:"forecasts"`
}

// WeatherSymbol for use with WeatherForcast
type WeatherSymbol int

// WeatherSymbol constants
const (
	WeatherSymbolNone  WeatherSymbol = -2
	WeatherSymbolSunny WeatherSymbol = iota
	WeatherSymbolFewClouds
	WeatherSymbolPartlyCloudy
	WeatherSymbolMostlyCloudy
	WeatherSymbolOvercast
	WeatherSymbolDrizzle
	WeatherSymbolRain
	WeatherSymbolFreezingRain
	WeatherSymbolShowers
	WeatherSymbolHail
	WeatherSymbolSnow
	WeatherSymbolFlurries
	WeatherSymbolFreeingSnow
	WeatherSymbolBlizzard
	WeatherSymbolPellets
	WeatherSymbolThunderstorm
	WeatherSymbolWindy
	WeatherSymbolTornado
	WeatherSymbolFog
	WeatherSymbolHaze
	WeatherSymbolSmoke
	WeatherSymbolDust
)

// WeatherForecast information for a Thermostat. The first forecast is the most
// accurate, later forecasts become less accurate in distance and time.
// See https://www.ecobee.com/home/developer/api/documentation/v1/objects/WeatherForecast.shtml
type WeatherForecast struct {
	WeatherSymbol    WeatherSymbol `json:"weatherSymbol"`
	DateTime         string        `json:"dateTime"`
	Condition        string        `json:"condition"`
	Temperature      int           `json:"temperature"`
	Pressure         int           `json:"pressure"`
	RelativeHumidity int           `json:"relativeHumidity"`
	Dewpoint         int           `json:"dewpoint"`
	Visibility       int           `json:"visibility"`
	WindSpeed        int           `json:"windSpeed"`
	WindGust         int           `json:"windGust"`
	WindDirection    string        `json:"windDirection"`
	WindBearing      int           `json:"windBearing"`
	Pop              int           `json:"pop"`
	TempHigh         int           `json:"tempHigh"`
	TempLow          int           `json:"tempLow"`
	Sky              int           `json:"sky"`
}
