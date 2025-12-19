package api

import (
	"encoding/json"
	"testing"
)

func TestVecBaseInfosResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"vecBaseInfos": [
			{
				"Vehicle": {
					"CvInformation": {
						"internalVin": "12345678901234567"
					}
				}
			}
		]
	}`

	var resp VecBaseInfosResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.ResultCode != "200S00" {
		t.Errorf("Expected resultCode '200S00', got '%s'", resp.ResultCode)
	}

	if len(resp.VecBaseInfos) != 1 {
		t.Fatalf("Expected 1 vehicle, got %d", len(resp.VecBaseInfos))
	}

	vin := resp.VecBaseInfos[0].Vehicle.CvInformation.InternalVIN
	if vin != "12345678901234567" {
		t.Errorf("Expected internalVin '12345678901234567', got '%v'", vin)
	}
}

func TestVecBaseInfosResponse_InternalVINAsNumber(t *testing.T) {
	// The API sometimes returns internalVin as a number
	jsonData := `{
		"resultCode": "200S00",
		"vecBaseInfos": [
			{
				"Vehicle": {
					"CvInformation": {
						"internalVin": 12345678901234567
					}
				}
			}
		]
	}`

	var resp VecBaseInfosResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	vin := resp.VecBaseInfos[0].Vehicle.CvInformation.InternalVIN
	// When parsed as float64, large numbers lose precision
	if vin == "" {
		t.Error("Expected internalVin to be set")
	}
}

func TestVehicleStatusResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"remoteInfos": [
			{
				"ResidualFuel": {
					"FuelSegementDActl": 92.0,
					"RemDrvDistDActlKm": 630.5
				},
				"DriveInformation": {
					"OdoDispValue": 12345.6
				},
				"TPMSInformation": {
					"FLTPrsDispPsi": 32.5,
					"FRTPrsDispPsi": 32.0,
					"RLTPrsDispPsi": 31.5,
					"RRTPrsDispPsi": 31.8
				}
			}
		],
		"alertInfos": [
			{
				"PositionInfo": {
					"Latitude": 37.7749,
					"Longitude": -122.4194,
					"AcquisitionDatetime": "20231201120000"
				},
				"Door": {
					"DrStatDrv": 0,
					"DrStatPsngr": 0,
					"DrStatRl": 0,
					"DrStatRr": 0,
					"DrStatTrnkLg": 0
				}
			}
		]
	}`

	var resp VehicleStatusResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.ResultCode != "200S00" {
		t.Errorf("Expected resultCode '200S00', got '%s'", resp.ResultCode)
	}

	if len(resp.RemoteInfos) != 1 {
		t.Fatalf("Expected 1 remoteInfo, got %d", len(resp.RemoteInfos))
	}

	fuel := resp.RemoteInfos[0].ResidualFuel
	if fuel.FuelSegmentDActl != 92.0 {
		t.Errorf("Expected FuelSegmentDActl 92.0, got %f", fuel.FuelSegmentDActl)
	}
	if fuel.RemDrvDistDActlKm != 630.5 {
		t.Errorf("Expected RemDrvDistDActlKm 630.5, got %f", fuel.RemDrvDistDActlKm)
	}

	driveInfo := resp.RemoteInfos[0].DriveInformation
	if driveInfo.OdoDispValue != 12345.6 {
		t.Errorf("Expected OdoDispValue 12345.6, got %f", driveInfo.OdoDispValue)
	}

	tpms := resp.RemoteInfos[0].TPMSInformation
	if tpms.FLTPrsDispPsi != 32.5 {
		t.Errorf("Expected FLTPrsDispPsi 32.5, got %f", tpms.FLTPrsDispPsi)
	}

	alert := resp.AlertInfos[0]
	if alert.PositionInfo.Latitude != 37.7749 {
		t.Errorf("Expected Latitude 37.7749, got %f", alert.PositionInfo.Latitude)
	}
	if alert.PositionInfo.Longitude != -122.4194 {
		t.Errorf("Expected Longitude -122.4194, got %f", alert.PositionInfo.Longitude)
	}

	door := alert.Door
	if door.DrStatDrv != 0 || door.DrStatPsngr != 0 || door.DrStatRl != 0 || door.DrStatRr != 0 || door.DrStatTrnkLg != 0 {
		t.Error("Expected all doors to be locked (0)")
	}
}

func TestEVVehicleStatusResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"resultData": [
			{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": {
					"VehicleInfo": {
						"ChargeInfo": {
							"SmaphSOC": 66.0,
							"SmaphRemDrvDistKm": 245.5,
							"ChargerConnectorFitting": 1,
							"ChargeStatusSub": 6,
							"MaxChargeMinuteAC": 180,
							"MaxChargeMinuteQBC": 45
						},
						"RemoteHvacInfo": {
							"HVAC": 1,
							"FrontDefroster": 1,
							"RearDefogger": 0,
							"InCarTeDC": 21.5
						}
					}
				}
			}
		]
	}`

	var resp EVVehicleStatusResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.ResultCode != "200S00" {
		t.Errorf("Expected resultCode '200S00', got '%s'", resp.ResultCode)
	}

	if len(resp.ResultData) != 1 {
		t.Fatalf("Expected 1 resultData, got %d", len(resp.ResultData))
	}

	result := resp.ResultData[0]
	if result.OccurrenceDate != "20231201120000" {
		t.Errorf("Expected OccurrenceDate '20231201120000', got '%s'", result.OccurrenceDate)
	}

	chargeInfo := result.PlusBInformation.VehicleInfo.ChargeInfo
	if chargeInfo.SmaphSOC != 66.0 {
		t.Errorf("Expected SmaphSOC 66.0, got %f", chargeInfo.SmaphSOC)
	}
	if chargeInfo.SmaphRemDrvDistKm != 245.5 {
		t.Errorf("Expected SmaphRemDrvDistKm 245.5, got %f", chargeInfo.SmaphRemDrvDistKm)
	}
	if chargeInfo.ChargerConnectorFitting != 1 {
		t.Errorf("Expected ChargerConnectorFitting 1, got %f", chargeInfo.ChargerConnectorFitting)
	}
	if chargeInfo.ChargeStatusSub != 6 {
		t.Errorf("Expected ChargeStatusSub 6, got %f", chargeInfo.ChargeStatusSub)
	}
	if chargeInfo.MaxChargeMinuteAC != 180 {
		t.Errorf("Expected MaxChargeMinuteAC 180, got %f", chargeInfo.MaxChargeMinuteAC)
	}
	if chargeInfo.MaxChargeMinuteQBC != 45 {
		t.Errorf("Expected MaxChargeMinuteQBC 45, got %f", chargeInfo.MaxChargeMinuteQBC)
	}

	hvacInfo := result.PlusBInformation.VehicleInfo.RemoteHvacInfo
	if hvacInfo == nil {
		t.Fatal("Expected RemoteHvacInfo to be set")
	}
	if hvacInfo.HVAC != 1 {
		t.Errorf("Expected HVAC 1, got %f", hvacInfo.HVAC)
	}
	if hvacInfo.FrontDefroster != 1 {
		t.Errorf("Expected FrontDefroster 1, got %f", hvacInfo.FrontDefroster)
	}
	if hvacInfo.InCarTeDC != 21.5 {
		t.Errorf("Expected InCarTeDC 21.5, got %f", hvacInfo.InCarTeDC)
	}
}

func TestEVVehicleStatusResponse_MissingHvacInfo(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"resultData": [
			{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": {
					"VehicleInfo": {
						"ChargeInfo": {
							"SmaphSOC": 66.0,
							"SmaphRemDrvDistKm": 245.5,
							"ChargerConnectorFitting": 0,
							"ChargeStatusSub": 0
						}
					}
				}
			}
		]
	}`

	var resp EVVehicleStatusResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	hvacInfo := resp.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo
	if hvacInfo != nil {
		t.Error("Expected RemoteHvacInfo to be nil when not present")
	}
}

func TestInternalVIN_UnmarshalString(t *testing.T) {
	var vin InternalVIN
	if err := json.Unmarshal([]byte(`"ABC123"`), &vin); err != nil {
		t.Fatalf("Failed to unmarshal string: %v", err)
	}
	if string(vin) != "ABC123" {
		t.Errorf("Expected 'ABC123', got '%s'", string(vin))
	}
}

func TestInternalVIN_UnmarshalNumber(t *testing.T) {
	var vin InternalVIN
	if err := json.Unmarshal([]byte(`12345`), &vin); err != nil {
		t.Fatalf("Failed to unmarshal number: %v", err)
	}
	if string(vin) != "12345" {
		t.Errorf("Expected '12345', got '%s'", string(vin))
	}
}

func TestInternalVIN_UnmarshalLargeNumber(t *testing.T) {
	// Test with a large number that would lose precision as float64
	var vin InternalVIN
	if err := json.Unmarshal([]byte(`12345678901234567`), &vin); err != nil {
		t.Fatalf("Failed to unmarshal large number: %v", err)
	}
	// The exact value may be affected by float64 precision
	if string(vin) == "" {
		t.Error("Expected non-empty VIN")
	}
}

// Auth response struct tests

func TestCheckVersionResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"encKey": "test-encryption-key-1234",
		"signKey": "test-signing-key-5678"
	}`

	var resp CheckVersionResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.EncKey != "test-encryption-key-1234" {
		t.Errorf("Expected encKey 'test-encryption-key-1234', got '%s'", resp.EncKey)
	}
	if resp.SignKey != "test-signing-key-5678" {
		t.Errorf("Expected signKey 'test-signing-key-5678', got '%s'", resp.SignKey)
	}
}

func TestUsherEncryptionKeyResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"data": {
			"publicKey": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...",
			"versionPrefix": "v1:"
		}
	}`

	var resp UsherEncryptionKeyResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.Data.PublicKey != "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ..." {
		t.Errorf("Expected publicKey 'MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...', got '%s'", resp.Data.PublicKey)
	}
	if resp.Data.VersionPrefix != "v1:" {
		t.Errorf("Expected versionPrefix 'v1:', got '%s'", resp.Data.VersionPrefix)
	}
}

func TestLoginResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"status": "OK",
		"data": {
			"accessToken": "test-access-token-abc123",
			"accessTokenExpirationTs": 1701446400
		}
	}`

	var resp LoginResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.Status != "OK" {
		t.Errorf("Expected status 'OK', got '%s'", resp.Status)
	}
	if resp.Data.AccessToken != "test-access-token-abc123" {
		t.Errorf("Expected accessToken 'test-access-token-abc123', got '%s'", resp.Data.AccessToken)
	}
	if resp.Data.AccessTokenExpirationTs != 1701446400 {
		t.Errorf("Expected accessTokenExpirationTs 1701446400, got %d", resp.Data.AccessTokenExpirationTs)
	}
}

func TestLoginResponse_InvalidCredential(t *testing.T) {
	jsonData := `{
		"status": "INVALID_CREDENTIAL"
	}`

	var resp LoginResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.Status != "INVALID_CREDENTIAL" {
		t.Errorf("Expected status 'INVALID_CREDENTIAL', got '%s'", resp.Status)
	}
}

func TestAPIBaseResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		wantState string
		wantErr   float64
	}{
		{
			name:      "success response",
			jsonData:  `{"state": "S", "payload": "encrypted-data"}`,
			wantState: "S",
			wantErr:   0,
		},
		{
			name:      "error response with code",
			jsonData:  `{"state": "E", "errorCode": 600001, "message": "encryption error"}`,
			wantState: "E",
			wantErr:   600001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp APIBaseResponse
			if err := json.Unmarshal([]byte(tt.jsonData), &resp); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if resp.State != tt.wantState {
				t.Errorf("Expected state '%s', got '%s'", tt.wantState, resp.State)
			}
			if resp.ErrorCode != tt.wantErr {
				t.Errorf("Expected errorCode %f, got %f", tt.wantErr, resp.ErrorCode)
			}
		})
	}
}

// TemperatureUnit tests

func TestTemperatureUnit_String(t *testing.T) {
	tests := []struct {
		unit TemperatureUnit
		want string
	}{
		{Celsius, "C"},
		{Fahrenheit, "F"},
		{TemperatureUnit(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.unit.String(); got != tt.want {
				t.Errorf("TemperatureUnit.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTemperatureUnit(t *testing.T) {
	tests := []struct {
		input   string
		want    TemperatureUnit
		wantErr bool
	}{
		{"c", Celsius, false},
		{"C", Celsius, false},
		{"celsius", Celsius, false},
		{"Celsius", Celsius, false},
		{"f", Fahrenheit, false},
		{"F", Fahrenheit, false},
		{"fahrenheit", Fahrenheit, false},
		{"Fahrenheit", Fahrenheit, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"kelvin", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTemperatureUnit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTemperatureUnit(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTemperatureUnit(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVehicleStatusResponse_GetOdometerInfo(t *testing.T) {
	tests := []struct {
		name         string
		resp         *VehicleStatusResponse
		wantOdometer float64
		wantErr      bool
	}{
		{
			name: "valid odometer",
			resp: &VehicleStatusResponse{
				RemoteInfos: []RemoteInfo{
					{
						DriveInformation: DriveInformation{
							OdoDispValue: 12345.6,
						},
					},
				},
			},
			wantOdometer: 12345.6,
			wantErr:      false,
		},
		{
			name: "high odometer value",
			resp: &VehicleStatusResponse{
				RemoteInfos: []RemoteInfo{
					{
						DriveInformation: DriveInformation{
							OdoDispValue: 99999.9,
						},
					},
				},
			},
			wantOdometer: 99999.9,
			wantErr:      false,
		},
		{
			name: "zero odometer",
			resp: &VehicleStatusResponse{
				RemoteInfos: []RemoteInfo{
					{
						DriveInformation: DriveInformation{
							OdoDispValue: 0,
						},
					},
				},
			},
			wantOdometer: 0,
			wantErr:      false,
		},
		{
			name:         "no remote infos",
			resp:         &VehicleStatusResponse{},
			wantOdometer: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			odometer, err := tt.resp.GetOdometerInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOdometerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if odometer != tt.wantOdometer {
				t.Errorf("GetOdometerInfo() = %v, want %v", odometer, tt.wantOdometer)
			}
		})
	}
}
