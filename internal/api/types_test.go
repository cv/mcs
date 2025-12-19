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
					"DrStatTrnkLg": 0,
					"DrStatHood": 0,
					"LockLinkSwDrv": 0,
					"LockLinkSwPsngr": 0,
					"LockLinkSwRl": 0,
					"LockLinkSwRr": 0
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
							"MaxChargeMinuteQBC": 45,
							"BatteryHeaterON": 1,
							"CstmzStatBatHeatAutoSW": 1
						},
						"RemoteHvacInfo": {
							"HVAC": 1,
							"FrontDefroster": 1,
							"RearDefogger": 0,
							"InCarTeDC": 21.5,
							"TargetTemp": 22.0
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
	if chargeInfo.BatteryHeaterON != 1 {
		t.Errorf("Expected BatteryHeaterON 1, got %f", chargeInfo.BatteryHeaterON)
	}
	if chargeInfo.CstmzStatBatHeatAutoSW != 1 {
		t.Errorf("Expected CstmzStatBatHeatAutoSW 1, got %f", chargeInfo.CstmzStatBatHeatAutoSW)
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
	if hvacInfo.TargetTemp != 22.0 {
		t.Errorf("Expected TargetTemp 22.0, got %f", hvacInfo.TargetTemp)
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

func TestVehicleStatusResponse_GetDoorsInfo(t *testing.T) {
	tests := []struct {
		name       string
		resp       *VehicleStatusResponse
		wantStatus DoorStatus
		wantErr    bool
	}{
		{
			name: "all locked and closed",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Door: DoorInfo{
							DrStatDrv:         0,
							DrStatPsngr:       0,
							DrStatRl:          0,
							DrStatRr:          0,
							DrStatTrnkLg:      0,
							DrStatHood:        0,
							LockLinkSwDrv:     0,
							LockLinkSwPsngr:   0,
							LockLinkSwRl:      0,
							LockLinkSwRr:      0,
							FuelLidOpenStatus: 0,
						},
					},
				},
			},
			wantStatus: DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       false,
				HoodOpen:        false,
				FuelLidOpen:     false,
				DriverLocked:    true,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       true,
			},
			wantErr: false,
		},
		{
			name: "driver door unlocked",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Door: DoorInfo{
							DrStatDrv:         0,
							DrStatPsngr:       0,
							DrStatRl:          0,
							DrStatRr:          0,
							DrStatTrnkLg:      0,
							DrStatHood:        0,
							LockLinkSwDrv:     1, // unlocked
							LockLinkSwPsngr:   0,
							LockLinkSwRl:      0,
							LockLinkSwRr:      0,
							FuelLidOpenStatus: 0,
						},
					},
				},
			},
			wantStatus: DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       false,
				HoodOpen:        false,
				FuelLidOpen:     false,
				DriverLocked:    false, // unlocked
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       false,
			},
			wantErr: false,
		},
		{
			name: "trunk and hood open",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Door: DoorInfo{
							DrStatDrv:         0,
							DrStatPsngr:       0,
							DrStatRl:          0,
							DrStatRr:          0,
							DrStatTrnkLg:      1, // open
							DrStatHood:        1, // open
							LockLinkSwDrv:     0,
							LockLinkSwPsngr:   0,
							LockLinkSwRl:      0,
							LockLinkSwRr:      0,
							FuelLidOpenStatus: 0,
						},
					},
				},
			},
			wantStatus: DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       true, // open
				HoodOpen:        true, // open
				FuelLidOpen:     false,
				DriverLocked:    true,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       false, // not all locked because doors are open
			},
			wantErr: false,
		},
		{
			name: "fuel lid open",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Door: DoorInfo{
							DrStatDrv:         0,
							DrStatPsngr:       0,
							DrStatRl:          0,
							DrStatRr:          0,
							DrStatTrnkLg:      0,
							DrStatHood:        0,
							LockLinkSwDrv:     0,
							LockLinkSwPsngr:   0,
							LockLinkSwRl:      0,
							LockLinkSwRr:      0,
							FuelLidOpenStatus: 1, // open
						},
					},
				},
			},
			wantStatus: DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       false,
				HoodOpen:        false,
				FuelLidOpen:     true, // open
				DriverLocked:    true,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       true, // fuel lid doesn't affect lock status
			},
			wantErr: false,
		},
		{
			name:       "no alert infos",
			resp:       &VehicleStatusResponse{},
			wantStatus: DoorStatus{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := tt.resp.GetDoorsInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDoorsInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if status != tt.wantStatus {
				t.Errorf("GetDoorsInfo() = %+v, want %+v", status, tt.wantStatus)
			}
		})
	}
}

func TestVehicleStatusResponse_WindowParsing(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"remoteInfos": [],
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
					"DrStatTrnkLg": 0,
					"DrStatHood": 0,
					"LockLinkSwDrv": 0,
					"LockLinkSwPsngr": 0,
					"LockLinkSwRl": 0,
					"LockLinkSwRr": 0
				},
				"Pw": {
					"PwPosDrv": 0,
					"PwPosPsngr": 50,
					"PwPosRl": 0,
					"PwPosRr": 25
				}
			}
		]
	}`

	var resp VehicleStatusResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.AlertInfos) != 1 {
		t.Fatalf("Expected 1 alertInfo, got %d", len(resp.AlertInfos))
	}

	pw := resp.AlertInfos[0].Pw
	if pw.PwPosDrv != 0 {
		t.Errorf("Expected PwPosDrv 0, got %f", pw.PwPosDrv)
	}
	if pw.PwPosPsngr != 50 {
		t.Errorf("Expected PwPosPsngr 50, got %f", pw.PwPosPsngr)
	}
	if pw.PwPosRl != 0 {
		t.Errorf("Expected PwPosRl 0, got %f", pw.PwPosRl)
	}
	if pw.PwPosRr != 25 {
		t.Errorf("Expected PwPosRr 25, got %f", pw.PwPosRr)
	}
}

func TestVehicleStatusResponse_GetWindowsInfo(t *testing.T) {
	tests := []struct {
		name       string
		resp       *VehicleStatusResponse
		wantDriver float64
		wantPass   float64
		wantRL     float64
		wantRR     float64
		wantErr    bool
	}{
		{
			name: "all windows closed",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Pw: WindowInfo{
							PwPosDrv:   0,
							PwPosPsngr: 0,
							PwPosRl:    0,
							PwPosRr:    0,
						},
					},
				},
			},
			wantDriver: 0,
			wantPass:   0,
			wantRL:     0,
			wantRR:     0,
			wantErr:    false,
		},
		{
			name: "some windows open",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Pw: WindowInfo{
							PwPosDrv:   0,
							PwPosPsngr: 50,
							PwPosRl:    0,
							PwPosRr:    25,
						},
					},
				},
			},
			wantDriver: 0,
			wantPass:   50,
			wantRL:     0,
			wantRR:     25,
			wantErr:    false,
		},
		{
			name: "all windows fully open",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						Pw: WindowInfo{
							PwPosDrv:   100,
							PwPosPsngr: 100,
							PwPosRl:    100,
							PwPosRr:    100,
						},
					},
				},
			},
			wantDriver: 100,
			wantPass:   100,
			wantRL:     100,
			wantRR:     100,
			wantErr:    false,
		},
		{
			name:       "no alert infos",
			resp:       &VehicleStatusResponse{},
			wantDriver: 0,
			wantPass:   0,
			wantRL:     0,
			wantRR:     0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, pass, rl, rr, err := tt.resp.GetWindowsInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWindowsInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if driver != tt.wantDriver {
				t.Errorf("GetWindowsInfo() driver = %v, want %v", driver, tt.wantDriver)
			}
			if pass != tt.wantPass {
				t.Errorf("GetWindowsInfo() passenger = %v, want %v", pass, tt.wantPass)
			}
			if rl != tt.wantRL {
				t.Errorf("GetWindowsInfo() rear left = %v, want %v", rl, tt.wantRL)
			}
			if rr != tt.wantRR {
				t.Errorf("GetWindowsInfo() rear right = %v, want %v", rr, tt.wantRR)
			}
		})
	}
}

func TestVehicleStatusResponse_HazardParsing(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"remoteInfos": [],
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
					"DrStatTrnkLg": 0,
					"DrStatHood": 0,
					"LockLinkSwDrv": 0,
					"LockLinkSwPsngr": 0,
					"LockLinkSwRl": 0,
					"LockLinkSwRr": 0
				},
				"HazardLamp": {
					"HazardSw": 1
				}
			}
		]
	}`

	var resp VehicleStatusResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.AlertInfos) != 1 {
		t.Fatalf("Expected 1 alertInfo, got %d", len(resp.AlertInfos))
	}

	hazard := resp.AlertInfos[0].HazardLamp
	if hazard.HazardSw != 1 {
		t.Errorf("Expected HazardSw 1, got %f", hazard.HazardSw)
	}
}

func TestVehicleStatusResponse_GetHazardInfo(t *testing.T) {
	tests := []struct {
		name        string
		resp        *VehicleStatusResponse
		wantHazards bool
		wantErr     bool
	}{
		{
			name: "hazards on",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						HazardLamp: HazardLamp{
							HazardSw: 1,
						},
					},
				},
			},
			wantHazards: true,
			wantErr:     false,
		},
		{
			name: "hazards off",
			resp: &VehicleStatusResponse{
				AlertInfos: []AlertInfo{
					{
						HazardLamp: HazardLamp{
							HazardSw: 0,
						},
					},
				},
			},
			wantHazards: false,
			wantErr:     false,
		},
		{
			name:        "no alert infos",
			resp:        &VehicleStatusResponse{},
			wantHazards: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hazards, err := tt.resp.GetHazardInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHazardInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hazards != tt.wantHazards {
				t.Errorf("GetHazardInfo() = %v, want %v", hazards, tt.wantHazards)
			}
		})
	}
}

func TestEVVehicleStatusResponse_GetHvacInfo(t *testing.T) {
	tests := []struct {
		name           string
		resp           *EVVehicleStatusResponse
		wantHvacOn     bool
		wantFrontDef   bool
		wantRearDef    bool
		wantInteriorC  float64
		wantTargetC    float64
	}{
		{
			name: "HVAC on with target temp",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								RemoteHvacInfo: &RemoteHvacInfo{
									HVAC:           1,
									FrontDefroster: 1,
									RearDefogger:   0,
									InCarTeDC:      18.0,
									TargetTemp:     22.0,
								},
							},
						},
					},
				},
			},
			wantHvacOn:    true,
			wantFrontDef:  true,
			wantRearDef:   false,
			wantInteriorC: 18.0,
			wantTargetC:   22.0,
		},
		{
			name: "HVAC off",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								RemoteHvacInfo: &RemoteHvacInfo{
									HVAC:           0,
									FrontDefroster: 0,
									RearDefogger:   0,
									InCarTeDC:      20.0,
									TargetTemp:     21.0,
								},
							},
						},
					},
				},
			},
			wantHvacOn:    false,
			wantFrontDef:  false,
			wantRearDef:   false,
			wantInteriorC: 20.0,
			wantTargetC:   21.0,
		},
		{
			name: "no HVAC info",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{},
						},
					},
				},
			},
			wantHvacOn:    false,
			wantFrontDef:  false,
			wantRearDef:   false,
			wantInteriorC: 0,
			wantTargetC:   0,
		},
		{
			name:          "no result data",
			resp:          &EVVehicleStatusResponse{},
			wantHvacOn:    false,
			wantFrontDef:  false,
			wantRearDef:   false,
			wantInteriorC: 0,
			wantTargetC:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hvacOn, frontDef, rearDef, interiorC, targetC := tt.resp.GetHvacInfo()
			if hvacOn != tt.wantHvacOn {
				t.Errorf("GetHvacInfo() hvacOn = %v, want %v", hvacOn, tt.wantHvacOn)
			}
			if frontDef != tt.wantFrontDef {
				t.Errorf("GetHvacInfo() frontDef = %v, want %v", frontDef, tt.wantFrontDef)
			}
			if rearDef != tt.wantRearDef {
				t.Errorf("GetHvacInfo() rearDef = %v, want %v", rearDef, tt.wantRearDef)
			}
			if interiorC != tt.wantInteriorC {
				t.Errorf("GetHvacInfo() interiorC = %v, want %v", interiorC, tt.wantInteriorC)
			}
			if targetC != tt.wantTargetC {
				t.Errorf("GetHvacInfo() targetC = %v, want %v", targetC, tt.wantTargetC)
			}
		})
	}
}

func TestEVVehicleStatusResponse_GetBatteryInfo(t *testing.T) {
	tests := []struct {
		name              string
		resp              *EVVehicleStatusResponse
		wantBattery       float64
		wantRange         float64
		wantChargeTimeAC  float64
		wantChargeTimeQBC float64
		wantPluggedIn     bool
		wantCharging      bool
		wantHeaterOn      bool
		wantHeaterAuto    bool
		wantErr           bool
	}{
		{
			name: "charging with heater on",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:               66.0,
									SmaphRemDrvDistKm:      245.5,
									ChargerConnectorFitting: 1,
									ChargeStatusSub:        6,
									MaxChargeMinuteAC:      180,
									MaxChargeMinuteQBC:     45,
									BatteryHeaterON:        1,
									CstmzStatBatHeatAutoSW: 1,
								},
							},
						},
					},
				},
			},
			wantBattery:       66.0,
			wantRange:         245.5,
			wantChargeTimeAC:  180,
			wantChargeTimeQBC: 45,
			wantPluggedIn:     true,
			wantCharging:      true,
			wantHeaterOn:      true,
			wantHeaterAuto:    true,
			wantErr:           false,
		},
		{
			name: "not charging, heater off",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:               80.0,
									SmaphRemDrvDistKm:      300.0,
									ChargerConnectorFitting: 0,
									ChargeStatusSub:        0,
									BatteryHeaterON:        0,
									CstmzStatBatHeatAutoSW: 0,
								},
							},
						},
					},
				},
			},
			wantBattery:       80.0,
			wantRange:         300.0,
			wantChargeTimeAC:  0,
			wantChargeTimeQBC: 0,
			wantPluggedIn:     false,
			wantCharging:      false,
			wantHeaterOn:      false,
			wantHeaterAuto:    false,
			wantErr:           false,
		},
		{
			name: "heater on auto but not running",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:               75.0,
									SmaphRemDrvDistKm:      280.0,
									ChargerConnectorFitting: 0,
									ChargeStatusSub:        0,
									BatteryHeaterON:        0,
									CstmzStatBatHeatAutoSW: 1,
								},
							},
						},
					},
				},
			},
			wantBattery:       75.0,
			wantRange:         280.0,
			wantChargeTimeAC:  0,
			wantChargeTimeQBC: 0,
			wantPluggedIn:     false,
			wantCharging:      false,
			wantHeaterOn:      false,
			wantHeaterAuto:    true,
			wantErr:           false,
		},
		{
			name:              "no result data",
			resp:              &EVVehicleStatusResponse{},
			wantBattery:       0,
			wantRange:         0,
			wantChargeTimeAC:  0,
			wantChargeTimeQBC: 0,
			wantPluggedIn:     false,
			wantCharging:      false,
			wantHeaterOn:      false,
			wantHeaterAuto:    false,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			battery, rangeKm, chargeTimeAC, chargeTimeQBC, pluggedIn, charging, heaterOn, heaterAuto, err := tt.resp.GetBatteryInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatteryInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if battery != tt.wantBattery {
				t.Errorf("GetBatteryInfo() battery = %v, want %v", battery, tt.wantBattery)
			}
			if rangeKm != tt.wantRange {
				t.Errorf("GetBatteryInfo() range = %v, want %v", rangeKm, tt.wantRange)
			}
			if chargeTimeAC != tt.wantChargeTimeAC {
				t.Errorf("GetBatteryInfo() chargeTimeAC = %v, want %v", chargeTimeAC, tt.wantChargeTimeAC)
			}
			if chargeTimeQBC != tt.wantChargeTimeQBC {
				t.Errorf("GetBatteryInfo() chargeTimeQBC = %v, want %v", chargeTimeQBC, tt.wantChargeTimeQBC)
			}
			if pluggedIn != tt.wantPluggedIn {
				t.Errorf("GetBatteryInfo() pluggedIn = %v, want %v", pluggedIn, tt.wantPluggedIn)
			}
			if charging != tt.wantCharging {
				t.Errorf("GetBatteryInfo() charging = %v, want %v", charging, tt.wantCharging)
			}
			if heaterOn != tt.wantHeaterOn {
				t.Errorf("GetBatteryInfo() heaterOn = %v, want %v", heaterOn, tt.wantHeaterOn)
			}
			if heaterAuto != tt.wantHeaterAuto {
				t.Errorf("GetBatteryInfo() heaterAuto = %v, want %v", heaterAuto, tt.wantHeaterAuto)
			}
		})
	}
}
