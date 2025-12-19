package api

import (
	"encoding/json"
	"fmt"
)

// InternalVIN is a custom type that handles the API returning internalVin as either string or number
type InternalVIN string

// UnmarshalJSON handles unmarshaling internalVin from either string or number JSON values
func (v *InternalVIN) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*v = InternalVIN(s)
		return nil
	}

	// Try number
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*v = InternalVIN(fmt.Sprintf("%.0f", f))
		return nil
	}

	return fmt.Errorf("internalVin must be string or number, got: %s", string(data))
}

// String returns the string representation of InternalVIN
func (v InternalVIN) String() string {
	return string(v)
}

// VecBaseInfosResponse represents the response from GetVecBaseInfos API
type VecBaseInfosResponse struct {
	ResultCode   string        `json:"resultCode"`
	VecBaseInfos []VecBaseInfo `json:"vecBaseInfos"`
}

// VecBaseInfo represents a single vehicle's base information
type VecBaseInfo struct {
	VIN          string  `json:"vin"`
	Nickname     string  `json:"nickname"`
	EconnectType int     `json:"econnectType"`
	Vehicle      Vehicle `json:"Vehicle"`
}

// UnmarshalJSON implements custom unmarshaling to parse the nested vehicleInformation JSON string
func (v *VecBaseInfo) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type VecBaseInfoAlias VecBaseInfo
	var alias VecBaseInfoAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*v = VecBaseInfo(alias)

	// Parse the vehicleInformation JSON string if present
	if v.Vehicle.VehicleInformationJSON != "" {
		var parsed VehicleInformationParsed
		if err := json.Unmarshal([]byte(v.Vehicle.VehicleInformationJSON), &parsed); err == nil {
			v.Vehicle.VehicleInformation = parsed
		}
	}

	return nil
}

// Vehicle represents vehicle information
type Vehicle struct {
	CvInformation          CvInformation    `json:"CvInformation"`
	OtherInformation       OtherInformation `json:"OtherInformation"`
	VehicleInformationJSON string           `json:"vehicleInformation"` // JSON-encoded string
	VehicleInformation     VehicleInformationParsed
}

// CvInformation represents connected vehicle information
type CvInformation struct {
	InternalVIN InternalVIN `json:"internalVin"`
}

// OtherInformation contains additional vehicle details (from direct JSON field, often empty)
type OtherInformation struct {
	CarlineName       string  `json:"carlineName"`
	ModelYear         string  `json:"modelYear"`
	ModelName         string  `json:"modelName"`
	ExteriorColorName string  `json:"exteriorColorName"`
	IsElectric        float64 `json:"isElectric"`
}

// VehicleInformationParsed contains parsed vehicle details from the vehicleInformation JSON string
type VehicleInformationParsed struct {
	OtherInformation OtherInformationParsed `json:"OtherInformation"`
}

// OtherInformationParsed contains model info from the vehicleInformation JSON string
type OtherInformationParsed struct {
	CarlineName       string `json:"carlineName"`
	CarlineCode       string `json:"carlineCode"`
	ModelYear         string `json:"modelYear"`
	ModelCode         string `json:"modelCode"`
	ModelName         string `json:"modelName"`
	TransmissionType  string `json:"transmissionType"`
	ExteriorColorCode string `json:"exteriorColorCode"`
	ExteriorColorName string `json:"exteriorColorName"`
	InteriorColorCode string `json:"interiorColorCode"`
	InteriorColorName string `json:"interiorColorName"`
}

// VehicleStatusResponse represents the response from GetVehicleStatus API
type VehicleStatusResponse struct {
	ResultCode  string       `json:"resultCode"`
	RemoteInfos []RemoteInfo `json:"remoteInfos"`
	AlertInfos  []AlertInfo  `json:"alertInfos"`
}

// RemoteInfo contains remote vehicle information
type RemoteInfo struct {
	ResidualFuel     ResidualFuel     `json:"ResidualFuel"`
	DriveInformation DriveInformation `json:"DriveInformation"`
	TPMSInformation  TPMSInformation  `json:"TPMSInformation"`
}

// ResidualFuel contains fuel information
type ResidualFuel struct {
	FuelSegmentDActl  float64 `json:"FuelSegementDActl"`
	RemDrvDistDActlKm float64 `json:"RemDrvDistDActlKm"`
}

// DriveInformation contains drive-related information
type DriveInformation struct {
	OdoDispValue float64 `json:"OdoDispValue"`
}

// TPMSInformation contains tire pressure information
type TPMSInformation struct {
	FLTPrsDispPsi float64 `json:"FLTPrsDispPsi"`
	FRTPrsDispPsi float64 `json:"FRTPrsDispPsi"`
	RLTPrsDispPsi float64 `json:"RLTPrsDispPsi"`
	RRTPrsDispPsi float64 `json:"RRTPrsDispPsi"`
}

// AlertInfo contains alert and position information
type AlertInfo struct {
	PositionInfo PositionInfo `json:"PositionInfo"`
	Door         DoorInfo     `json:"Door"`
	Pw           WindowInfo   `json:"Pw"`
	HazardLamp   HazardLamp   `json:"HazardLamp"`
}

// PositionInfo contains GPS location information
type PositionInfo struct {
	Latitude            float64 `json:"Latitude"`
	Longitude           float64 `json:"Longitude"`
	AcquisitionDatetime string  `json:"AcquisitionDatetime"`
}

// DoorInfo contains door lock status
type DoorInfo struct {
	DrStatDrv         float64 `json:"DrStatDrv"`
	DrStatPsngr       float64 `json:"DrStatPsngr"`
	DrStatRl          float64 `json:"DrStatRl"`
	DrStatRr          float64 `json:"DrStatRr"`
	DrStatTrnkLg      float64 `json:"DrStatTrnkLg"`
	DrStatHood        float64 `json:"DrStatHood"`
	LockLinkSwDrv     float64 `json:"LockLinkSwDrv"`
	LockLinkSwPsngr   float64 `json:"LockLinkSwPsngr"`
	LockLinkSwRl      float64 `json:"LockLinkSwRl"`
	LockLinkSwRr      float64 `json:"LockLinkSwRr"`
	FuelLidOpenStatus float64 `json:"FuelLidOpenStatus"`
}

// WindowInfo contains window position information
type WindowInfo struct {
	PwPosDrv   float64 `json:"PwPosDrv"`
	PwPosPsngr float64 `json:"PwPosPsngr"`
	PwPosRl    float64 `json:"PwPosRl"`
	PwPosRr    float64 `json:"PwPosRr"`
}

// HazardLamp contains hazard lights information
type HazardLamp struct {
	HazardSw float64 `json:"HazardSw"`
}

// EVVehicleStatusResponse represents the response from GetEVVehicleStatus API
type EVVehicleStatusResponse struct {
	ResultCode string       `json:"resultCode"`
	ResultData []EVResultData `json:"resultData"`
}

// EVResultData contains EV-specific vehicle data
type EVResultData struct {
	OccurrenceDate   string           `json:"OccurrenceDate"`
	PlusBInformation PlusBInformation `json:"PlusBInformation"`
}

// PlusBInformation contains Plus-B (PHEV/EV) information
type PlusBInformation struct {
	VehicleInfo EVVehicleInfo `json:"VehicleInfo"`
}

// EVVehicleInfo contains EV vehicle information
type EVVehicleInfo struct {
	ChargeInfo     ChargeInfo      `json:"ChargeInfo"`
	RemoteHvacInfo *RemoteHvacInfo `json:"RemoteHvacInfo,omitempty"`
}

// ChargeInfo contains battery and charging information
type ChargeInfo struct {
	SmaphSOC                float64 `json:"SmaphSOC"`
	SmaphRemDrvDistKm       float64 `json:"SmaphRemDrvDistKm"`
	ChargerConnectorFitting float64 `json:"ChargerConnectorFitting"`
	ChargeStatusSub         float64 `json:"ChargeStatusSub"`
	MaxChargeMinuteAC       float64 `json:"MaxChargeMinuteAC"`
	MaxChargeMinuteQBC      float64 `json:"MaxChargeMinuteQBC"`
	BatteryHeaterON         float64 `json:"BatteryHeaterON"`
	CstmzStatBatHeatAutoSW  float64 `json:"CstmzStatBatHeatAutoSW"`
}

// RemoteHvacInfo contains HVAC system information
type RemoteHvacInfo struct {
	HVAC           float64 `json:"HVAC"`
	FrontDefroster float64 `json:"FrontDefroster"`
	RearDefogger   float64 `json:"RearDefogger"`
	InCarTeDC      float64 `json:"InCarTeDC"`
	TargetTemp     float64 `json:"TargetTemp"`
}

// Helper methods for extracting data

// GetInternalVIN extracts the internal VIN from the first vehicle in the response
func (r *VecBaseInfosResponse) GetInternalVIN() (string, error) {
	if len(r.VecBaseInfos) == 0 {
		return "", fmt.Errorf("no vehicles found")
	}
	return string(r.VecBaseInfos[0].Vehicle.CvInformation.InternalVIN), nil
}

// GetVehicleInfo extracts vehicle identification info from the response
func (r *VecBaseInfosResponse) GetVehicleInfo() (vin, nickname, modelName, modelYear string, err error) {
	if len(r.VecBaseInfos) == 0 {
		err = fmt.Errorf("no vehicles found")
		return
	}
	info := r.VecBaseInfos[0]
	vin = info.VIN
	nickname = info.Nickname
	// Use the parsed vehicleInformation (JSON string) which has the actual model data
	modelName = info.Vehicle.VehicleInformation.OtherInformation.ModelName
	modelYear = info.Vehicle.VehicleInformation.OtherInformation.ModelYear
	return
}

// GetBatteryInfo extracts battery information from the EV status response
func (r *EVVehicleStatusResponse) GetBatteryInfo() (batteryLevel, rangeKm, chargeTimeACMin, chargeTimeQBCMin float64, pluggedIn, charging, heaterOn, heaterAuto bool, err error) {
	if len(r.ResultData) == 0 {
		err = fmt.Errorf("no EV status data available")
		return
	}
	chargeInfo := r.ResultData[0].PlusBInformation.VehicleInfo.ChargeInfo
	batteryLevel = chargeInfo.SmaphSOC
	rangeKm = chargeInfo.SmaphRemDrvDistKm
	chargeTimeACMin = chargeInfo.MaxChargeMinuteAC
	chargeTimeQBCMin = chargeInfo.MaxChargeMinuteQBC
	pluggedIn = int(chargeInfo.ChargerConnectorFitting) == 1
	charging = int(chargeInfo.ChargeStatusSub) == 6
	heaterOn = int(chargeInfo.BatteryHeaterON) == 1
	heaterAuto = int(chargeInfo.CstmzStatBatHeatAutoSW) == 1
	return
}

// GetHvacInfo extracts HVAC information from the EV status response
func (r *EVVehicleStatusResponse) GetHvacInfo() (hvacOn, frontDefroster, rearDefroster bool, interiorTempC, targetTempC float64) {
	if len(r.ResultData) == 0 {
		return false, false, false, 0, 0
	}
	hvacInfo := r.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo
	if hvacInfo == nil {
		return false, false, false, 0, 0
	}
	hvacOn = int(hvacInfo.HVAC) == 1
	frontDefroster = int(hvacInfo.FrontDefroster) == 1
	rearDefroster = int(hvacInfo.RearDefogger) == 1
	interiorTempC = hvacInfo.InCarTeDC
	targetTempC = hvacInfo.TargetTemp
	return
}

// GetOccurrenceDate returns the occurrence date from the first result
func (r *EVVehicleStatusResponse) GetOccurrenceDate() string {
	if len(r.ResultData) == 0 {
		return ""
	}
	return r.ResultData[0].OccurrenceDate
}

// GetFuelInfo extracts fuel information from the vehicle status response
func (r *VehicleStatusResponse) GetFuelInfo() (fuelLevel, rangeKm float64, err error) {
	if len(r.RemoteInfos) == 0 {
		err = fmt.Errorf("no vehicle status data available")
		return
	}
	fuel := r.RemoteInfos[0].ResidualFuel
	fuelLevel = fuel.FuelSegmentDActl
	rangeKm = fuel.RemDrvDistDActlKm
	return
}

// GetTiresInfo extracts tire pressure information from the vehicle status response
func (r *VehicleStatusResponse) GetTiresInfo() (fl, fr, rl, rr float64, err error) {
	if len(r.RemoteInfos) == 0 {
		err = fmt.Errorf("no vehicle status data available")
		return
	}
	tpms := r.RemoteInfos[0].TPMSInformation
	fl = tpms.FLTPrsDispPsi
	fr = tpms.FRTPrsDispPsi
	rl = tpms.RLTPrsDispPsi
	rr = tpms.RRTPrsDispPsi
	return
}

// GetLocationInfo extracts location information from the vehicle status response
func (r *VehicleStatusResponse) GetLocationInfo() (lat, lon float64, timestamp string, err error) {
	if len(r.AlertInfos) == 0 {
		err = fmt.Errorf("no alert info available")
		return
	}
	pos := r.AlertInfos[0].PositionInfo
	lat = pos.Latitude
	lon = pos.Longitude
	timestamp = pos.AcquisitionDatetime
	return
}

// DoorStatus represents the detailed status of all doors
type DoorStatus struct {
	DriverOpen     bool
	PassengerOpen  bool
	RearLeftOpen   bool
	RearRightOpen  bool
	TrunkOpen      bool
	HoodOpen       bool
	FuelLidOpen    bool
	DriverLocked   bool
	PassengerLocked bool
	RearLeftLocked  bool
	RearRightLocked bool
	AllLocked      bool
}

// GetDoorsInfo extracts door lock status from the vehicle status response
func (r *VehicleStatusResponse) GetDoorsInfo() (status DoorStatus, err error) {
	if len(r.AlertInfos) == 0 {
		err = fmt.Errorf("no alert info available")
		return
	}
	door := r.AlertInfos[0].Door

	// Open status (1=open, 0=closed)
	status.DriverOpen = int(door.DrStatDrv) == 1
	status.PassengerOpen = int(door.DrStatPsngr) == 1
	status.RearLeftOpen = int(door.DrStatRl) == 1
	status.RearRightOpen = int(door.DrStatRr) == 1
	status.TrunkOpen = int(door.DrStatTrnkLg) == 1
	status.HoodOpen = int(door.DrStatHood) == 1
	status.FuelLidOpen = int(door.FuelLidOpenStatus) == 1

	// Lock status (0=locked, 1=unlocked)
	status.DriverLocked = int(door.LockLinkSwDrv) == 0
	status.PassengerLocked = int(door.LockLinkSwPsngr) == 0
	status.RearLeftLocked = int(door.LockLinkSwRl) == 0
	status.RearRightLocked = int(door.LockLinkSwRr) == 0

	// All locked if no doors are open and all are locked
	status.AllLocked = !status.DriverOpen && !status.PassengerOpen &&
		!status.RearLeftOpen && !status.RearRightOpen &&
		!status.TrunkOpen && !status.HoodOpen &&
		status.DriverLocked && status.PassengerLocked &&
		status.RearLeftLocked && status.RearRightLocked

	return
}

// GetOdometerInfo extracts odometer reading from the vehicle status response
func (r *VehicleStatusResponse) GetOdometerInfo() (odometerKm float64, err error) {
	if len(r.RemoteInfos) == 0 {
		err = fmt.Errorf("no vehicle status data available")
		return
	}
	odometerKm = r.RemoteInfos[0].DriveInformation.OdoDispValue
	return
}

// GetWindowsInfo extracts window position information from the vehicle status response
func (r *VehicleStatusResponse) GetWindowsInfo() (driver, passenger, rearLeft, rearRight float64, err error) {
	if len(r.AlertInfos) == 0 {
		err = fmt.Errorf("no alert info available")
		return
	}
	pw := r.AlertInfos[0].Pw
	driver = pw.PwPosDrv
	passenger = pw.PwPosPsngr
	rearLeft = pw.PwPosRl
	rearRight = pw.PwPosRr
	return
}

// GetHazardInfo extracts hazard lights status from the vehicle status response
func (r *VehicleStatusResponse) GetHazardInfo() (hazardsOn bool, err error) {
	if len(r.AlertInfos) == 0 {
		err = fmt.Errorf("no alert info available")
		return
	}
	hazardsOn = int(r.AlertInfos[0].HazardLamp.HazardSw) == 1
	return
}

// Auth response types

// APIBaseResponse represents the common base structure for API responses
type APIBaseResponse struct {
	State     string  `json:"state"`
	Payload   string  `json:"payload"`
	ErrorCode float64 `json:"errorCode"`
	ExtraCode string  `json:"extraCode"`
	Message   string  `json:"message"`
	Error     string  `json:"error"`
}

// CheckVersionResponse represents the decrypted response from checkVersion endpoint
type CheckVersionResponse struct {
	EncKey  string `json:"encKey"`
	SignKey string `json:"signKey"`
}

// UsherEncryptionKeyData contains the encryption key data from Usher API
type UsherEncryptionKeyData struct {
	PublicKey     string `json:"publicKey"`
	VersionPrefix string `json:"versionPrefix"`
}

// UsherEncryptionKeyResponse represents the response from system/encryptionKey endpoint
type UsherEncryptionKeyResponse struct {
	Data UsherEncryptionKeyData `json:"data"`
}

// LoginData contains the login response data
type LoginData struct {
	AccessToken             string `json:"accessToken"`
	AccessTokenExpirationTs int64  `json:"accessTokenExpirationTs"`
}

// LoginResponse represents the response from user/login endpoint
type LoginResponse struct {
	Status string    `json:"status"`
	Data   LoginData `json:"data"`
}

// TemperatureUnit represents the unit for temperature values
type TemperatureUnit int

const (
	// Celsius represents temperatures in Celsius
	Celsius TemperatureUnit = 1
	// Fahrenheit represents temperatures in Fahrenheit
	Fahrenheit TemperatureUnit = 2
)

// String returns the string representation of the temperature unit
func (t TemperatureUnit) String() string {
	switch t {
	case Celsius:
		return "C"
	case Fahrenheit:
		return "F"
	default:
		return "unknown"
	}
}

// ParseTemperatureUnit converts a string to a TemperatureUnit.
// Accepts "c", "C", "celsius" for Celsius and "f", "F", "fahrenheit" for Fahrenheit.
func ParseTemperatureUnit(s string) (TemperatureUnit, error) {
	switch s {
	case "c", "C", "celsius", "Celsius":
		return Celsius, nil
	case "f", "F", "fahrenheit", "Fahrenheit":
		return Fahrenheit, nil
	default:
		return 0, fmt.Errorf("invalid temperature unit: %s (must be 'c' or 'f')", s)
	}
}
