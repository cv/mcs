package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, ResultCodeSuccess, resp.ResultCode, "Expected resultCode 'ResultCodeSuccess', got '%s'", resp.ResultCode)

	require.Lenf(t, resp.VecBaseInfos, 1, "Expected 1 vehicle, got %d", len(resp.VecBaseInfos))

	vin := resp.VecBaseInfos[0].Vehicle.CvInformation.InternalVIN
	assert.EqualValuesf(t, "12345678901234567", vin, "Expected internalVin '12345678901234567', got '%v'", vin)
}

func TestVecBaseInfosResponse_VehicleInfo(t *testing.T) {
	jsonData := `{
		"resultCode": "200S00",
		"vecBaseInfos": [
			{
				"vin": "JM3KKEHC1R0123456",
				"nickname": "My CX-90",
				"econnectType": 1,
				"Vehicle": {
					"CvInformation": {
						"internalVin": "12345678901234567"
					}
				}
			}
		]
	}`

	var resp VecBaseInfosResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	require.Lenf(t, resp.VecBaseInfos, 1, "Expected 1 vehicle, got %d", len(resp.VecBaseInfos))

	info := resp.VecBaseInfos[0]
	assert.Equalf(t, "JM3KKEHC1R0123456", info.VIN, "Expected VIN 'JM3KKEHC1R0123456', got '%s'", info.VIN)
	assert.Equalf(t, "My CX-90", info.Nickname, "Expected Nickname 'My CX-90', got '%s'", info.Nickname)
	assert.Equalf(t, 1, info.EconnectType, "Expected EconnectType 1, got %d", info.EconnectType)
}

func TestVecBaseInfosResponse_GetVehicleInfo(t *testing.T) {
	// Test with JSON parsing to verify vehicleInformation string is properly parsed
	jsonData := `{
		"resultCode": "200S00",
		"vecBaseInfos": [
			{
				"vin": "JM3KKEHC1R0123456",
				"nickname": "My Car",
				"Vehicle": {
					"CvInformation": {"internalVin": "12345"},
					"vehicleInformation": "{\"OtherInformation\":{\"modelName\":\"CX-90 PHEV\",\"modelYear\":\"2024\"}}"
				}
			}
		]
	}`

	var resp VecBaseInfosResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	vin, nickname, modelName, modelYear, err := resp.GetVehicleInfo()
	require.NoError(t, err, "Unexpected error: %v")
	assert.Equalf(t, "JM3KKEHC1R0123456", vin, "Expected VIN 'JM3KKEHC1R0123456', got '%s'", vin)
	assert.Equalf(t, "My Car", nickname, "Expected nickname 'My Car', got '%s'", nickname)
	assert.Equalf(t, "CX-90 PHEV", modelName, "Expected modelName 'CX-90 PHEV', got '%s'", modelName)
	assert.Equalf(t, "2024", modelYear, "Expected modelYear '2024', got '%s'", modelYear)
}

func TestVecBaseInfosResponse_GetVehicleInfo_Empty(t *testing.T) {
	resp := &VecBaseInfosResponse{}
	_, _, _, _, err := resp.GetVehicleInfo()
	require.Error(t, err, "Expected error for empty response")
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	vin := resp.VecBaseInfos[0].Vehicle.CvInformation.InternalVIN
	// When parsed as float64, large numbers lose precision
	assert.NotEmpty(t, vin, "Expected internalVin to be set")
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, ResultCodeSuccess, resp.ResultCode, "Expected resultCode 'ResultCodeSuccess', got '%s'", resp.ResultCode)

	require.Lenf(t, resp.RemoteInfos, 1, "Expected 1 remoteInfo, got %d", len(resp.RemoteInfos))

	fuel := resp.RemoteInfos[0].ResidualFuel
	assert.InDelta(t, 92.0, fuel.FuelSegmentDActl, 0.0001)
	assert.InDelta(t, 630.5, fuel.RemDrvDistDActlKm, 0.0001)

	driveInfo := resp.RemoteInfos[0].DriveInformation
	assert.InDelta(t, 12345.6, driveInfo.OdoDispValue, 0.0001)

	tpms := resp.RemoteInfos[0].TPMSInformation
	assert.InDelta(t, 32.5, tpms.FLTPrsDispPsi, 0.0001)

	alert := resp.AlertInfos[0]
	assert.InDelta(t, 37.7749, alert.PositionInfo.Latitude, 0.0001)
	assert.InDelta(t, -122.4194, alert.PositionInfo.Longitude, 0.0001)

	door := alert.Door
	assert.Zero(t, door.DrStatDrv)
	assert.Zero(t, door.DrStatPsngr)
	assert.Zero(t, door.DrStatRl)
	assert.Zero(t, door.DrStatRr)
	assert.Zero(t, door.DrStatTrnkLg)
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, ResultCodeSuccess, resp.ResultCode, "Expected resultCode 'ResultCodeSuccess', got '%s'", resp.ResultCode)

	require.Lenf(t, resp.ResultData, 1, "Expected 1 resultData, got %d", len(resp.ResultData))

	result := resp.ResultData[0]
	assert.Equalf(t, "20231201120000", result.OccurrenceDate, "Expected OccurrenceDate '20231201120000', got '%s'", result.OccurrenceDate)

	chargeInfo := result.PlusBInformation.VehicleInfo.ChargeInfo
	assert.InDelta(t, 66.0, chargeInfo.SmaphSOC, 0.0001)
	assert.InDelta(t, 245.5, chargeInfo.SmaphRemDrvDistKm, 0.0001)
	assert.InDelta(t, 1, chargeInfo.ChargerConnectorFitting, 0.0001)
	assert.InDelta(t, 6, chargeInfo.ChargeStatusSub, 0.0001)
	assert.InDelta(t, 180, chargeInfo.MaxChargeMinuteAC, 0.0001)
	assert.InDelta(t, 45, chargeInfo.MaxChargeMinuteQBC, 0.0001)
	assert.InDelta(t, 1, chargeInfo.BatteryHeaterON, 0.0001)
	assert.InDelta(t, 1, chargeInfo.CstmzStatBatHeatAutoSW, 0.0001)

	hvacInfo := result.PlusBInformation.VehicleInfo.RemoteHvacInfo
	require.NotNil(t, hvacInfo, "Expected RemoteHvacInfo to be set")
	assert.InDelta(t, 1, hvacInfo.HVAC, 0.0001)
	assert.InDelta(t, 1, hvacInfo.FrontDefroster, 0.0001)
	assert.InDelta(t, 21.5, hvacInfo.InCarTeDC, 0.0001)
	assert.InDelta(t, 22.0, hvacInfo.TargetTemp, 0.0001)
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	hvacInfo := resp.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo
	assert.Nil(t, hvacInfo)
}

func TestInternalVIN_UnmarshalString(t *testing.T) {
	var vin InternalVIN
	err := json.Unmarshal([]byte(`"ABC123"`), &vin)
	require.NoError(t, err, "Failed to unmarshal string: %v")

	assert.Equalf(t, "ABC123", string(vin), "Expected 'ABC123', got '%s'", string(vin))
}

func TestInternalVIN_UnmarshalNumber(t *testing.T) {
	var vin InternalVIN
	err := json.Unmarshal([]byte(`12345`), &vin)
	require.NoError(t, err, "Failed to unmarshal number: %v")

	assert.Equalf(t, "12345", string(vin), "Expected '12345', got '%s'", string(vin))
}

func TestInternalVIN_UnmarshalLargeNumber(t *testing.T) {
	// Test with a large number that would lose precision as float64
	var vin InternalVIN
	err := json.Unmarshal([]byte(`12345678901234567`), &vin)
	require.NoError(t, err, "Failed to unmarshal large number: %v")

	// The exact value may be affected by float64 precision
	assert.NotEmpty(t, string(vin), "Expected non-empty VIN")
}

// Auth response struct tests

func TestCheckVersionResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"encKey": "test-encryption-key-1234",
		"signKey": "test-signing-key-5678"
	}`

	var resp CheckVersionResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, "test-encryption-key-1234", resp.EncKey, "Expected encKey 'test-encryption-key-1234', got '%s'", resp.EncKey)
	assert.Equalf(t, "test-signing-key-5678", resp.SignKey, "Expected signKey 'test-signing-key-5678', got '%s'", resp.SignKey)
}

func TestUsherEncryptionKeyResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"data": {
			"publicKey": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...",
			"versionPrefix": "v1:"
		}
	}`

	var resp UsherEncryptionKeyResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...", resp.Data.PublicKey, "Expected publicKey 'MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...', got '%s'", resp.Data.PublicKey)
	assert.Equalf(t, "v1:", resp.Data.VersionPrefix, "Expected versionPrefix 'v1:', got '%s'", resp.Data.VersionPrefix)
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, "OK", resp.Status, "Expected status 'OK', got '%s'", resp.Status)
	assert.Equalf(t, "test-access-token-abc123", resp.Data.AccessToken, "Expected accessToken 'test-access-token-abc123', got '%s'", resp.Data.AccessToken)
	assert.EqualValuesf(t, 1701446400, resp.Data.AccessTokenExpirationTs, "Expected accessTokenExpirationTs 1701446400, got %d", resp.Data.AccessTokenExpirationTs)
}

func TestLoginResponse_InvalidCredential(t *testing.T) {
	jsonData := `{
		"status": "INVALID_CREDENTIAL"
	}`

	var resp LoginResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	assert.Equalf(t, "INVALID_CREDENTIAL", resp.Status, "Expected status 'INVALID_CREDENTIAL', got '%s'", resp.Status)
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
			err := json.Unmarshal([]byte(tt.jsonData), &resp)
			require.NoError(t, err, "Failed to unmarshal: %v")

			assert.Equal(t, tt.wantState, resp.State)
			assert.InDelta(t, tt.wantErr, resp.ErrorCode, 0.0001)
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
			assert.Equal(t, tt.want, tt.unit.String())
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
		{"CELSIUS", Celsius, false},
		{"CeLsIuS", Celsius, false},
		{"f", Fahrenheit, false},
		{"F", Fahrenheit, false},
		{"fahrenheit", Fahrenheit, false},
		{"Fahrenheit", Fahrenheit, false},
		{"FAHRENHEIT", Fahrenheit, false},
		{"FaHrEnHeIt", Fahrenheit, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"kelvin", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTemperatureUnit(tt.input)
			assert.Equalf(t, tt.wantErr, (err != nil), "ParseTemperatureUnit(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "ParseTemperatureUnit(%q) = %v, want %v", tt.input, got, tt.want)
		})
	}
}

func TestVehicleStatusResponse_GetOdometerInfo(t *testing.T) {
	tests := []struct {
		name    string
		resp    *VehicleStatusResponse
		want    OdometerInfo
		wantErr bool
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
			want:    OdometerInfo{OdometerKm: 12345.6},
			wantErr: false,
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
			want:    OdometerInfo{OdometerKm: 99999.9},
			wantErr: false,
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
			want:    OdometerInfo{OdometerKm: 0},
			wantErr: false,
		},
		{
			name:    "no remote infos",
			resp:    &VehicleStatusResponse{},
			want:    OdometerInfo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resp.GetOdometerInfo()
			if tt.wantErr {
				require.Error(t, err, "GetOdometerInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetOdometerInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.want, got)
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
			if tt.wantErr {
				require.Error(t, err, "GetDoorsInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetDoorsInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.wantStatus, status)
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	require.Lenf(t, resp.AlertInfos, 1, "Expected 1 alertInfo, got %d", len(resp.AlertInfos))

	pw := resp.AlertInfos[0].Pw
	assert.InDelta(t, 0, pw.PwPosDrv, 0.0001)
	assert.InDelta(t, 50, pw.PwPosPsngr, 0.0001)
	assert.InDelta(t, 0, pw.PwPosRl, 0.0001)
	assert.InDelta(t, 25, pw.PwPosRr, 0.0001)
}

func TestVehicleStatusResponse_GetWindowsInfo(t *testing.T) {
	tests := []struct {
		name    string
		resp    *VehicleStatusResponse
		want    WindowStatus
		wantErr bool
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
			want: WindowStatus{
				DriverPosition:    0,
				PassengerPosition: 0,
				RearLeftPosition:  0,
				RearRightPosition: 0,
			},
			wantErr: false,
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
			want: WindowStatus{
				DriverPosition:    0,
				PassengerPosition: 50,
				RearLeftPosition:  0,
				RearRightPosition: 25,
			},
			wantErr: false,
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
			want: WindowStatus{
				DriverPosition:    100,
				PassengerPosition: 100,
				RearLeftPosition:  100,
				RearRightPosition: 100,
			},
			wantErr: false,
		},
		{
			name:    "no alert infos",
			resp:    &VehicleStatusResponse{},
			want:    WindowStatus{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resp.GetWindowsInfo()
			if tt.wantErr {
				require.Error(t, err, "GetWindowsInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetWindowsInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.want, got)
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
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err, "Failed to unmarshal: %v")

	require.Lenf(t, resp.AlertInfos, 1, "Expected 1 alertInfo, got %d", len(resp.AlertInfos))

	hazard := resp.AlertInfos[0].HazardLamp
	assert.InDelta(t, 1, hazard.HazardSw, 0.0001)
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
			if tt.wantErr {
				require.Error(t, err, "GetHazardInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetHazardInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.wantHazards, hazards)
		})
	}
}

func TestEVVehicleStatusResponse_GetHvacInfo(t *testing.T) {
	tests := []struct {
		name    string
		resp    *EVVehicleStatusResponse
		want    HVACInfo
		wantErr bool
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
			want: HVACInfo{
				HVACOn:         true,
				FrontDefroster: true,
				RearDefroster:  false,
				InteriorTempC:  18.0,
				TargetTempC:    22.0,
			},
			wantErr: false,
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
			want: HVACInfo{
				HVACOn:         false,
				FrontDefroster: false,
				RearDefroster:  false,
				InteriorTempC:  20.0,
				TargetTempC:    21.0,
			},
			wantErr: false,
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
			want:    HVACInfo{},
			wantErr: true,
		},
		{
			name:    "no result data",
			resp:    &EVVehicleStatusResponse{},
			want:    HVACInfo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resp.GetHvacInfo()
			if tt.wantErr {
				require.Error(t, err, "GetHvacInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetHvacInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEVVehicleStatusResponse_GetBatteryInfo(t *testing.T) {
	tests := []struct {
		name    string
		resp    *EVVehicleStatusResponse
		want    BatteryInfo
		wantErr bool
	}{
		{
			name: "charging with heater on",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:                66.0,
									SmaphRemDrvDistKm:       245.5,
									ChargerConnectorFitting: 1,
									ChargeStatusSub:         6,
									MaxChargeMinuteAC:       180,
									MaxChargeMinuteQBC:      45,
									BatteryHeaterON:         1,
									CstmzStatBatHeatAutoSW:  1,
								},
							},
						},
					},
				},
			},
			want: BatteryInfo{
				BatteryLevel:     66.0,
				RangeKm:          245.5,
				ChargeTimeACMin:  180,
				ChargeTimeQBCMin: 45,
				PluggedIn:        true,
				Charging:         true,
				HeaterOn:         true,
				HeaterAuto:       true,
			},
			wantErr: false,
		},
		{
			name: "not charging, heater off",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:                80.0,
									SmaphRemDrvDistKm:       300.0,
									ChargerConnectorFitting: 0,
									ChargeStatusSub:         0,
									BatteryHeaterON:         0,
									CstmzStatBatHeatAutoSW:  0,
								},
							},
						},
					},
				},
			},
			want: BatteryInfo{
				BatteryLevel:     80.0,
				RangeKm:          300.0,
				ChargeTimeACMin:  0,
				ChargeTimeQBCMin: 0,
				PluggedIn:        false,
				Charging:         false,
				HeaterOn:         false,
				HeaterAuto:       false,
			},
			wantErr: false,
		},
		{
			name: "heater on auto but not running",
			resp: &EVVehicleStatusResponse{
				ResultData: []EVResultData{
					{
						PlusBInformation: PlusBInformation{
							VehicleInfo: EVVehicleInfo{
								ChargeInfo: ChargeInfo{
									SmaphSOC:                75.0,
									SmaphRemDrvDistKm:       280.0,
									ChargerConnectorFitting: 0,
									ChargeStatusSub:         0,
									BatteryHeaterON:         0,
									CstmzStatBatHeatAutoSW:  1,
								},
							},
						},
					},
				},
			},
			want: BatteryInfo{
				BatteryLevel:     75.0,
				RangeKm:          280.0,
				ChargeTimeACMin:  0,
				ChargeTimeQBCMin: 0,
				PluggedIn:        false,
				Charging:         false,
				HeaterOn:         false,
				HeaterAuto:       true,
			},
			wantErr: false,
		},
		{
			name:    "no result data",
			resp:    &EVVehicleStatusResponse{},
			want:    BatteryInfo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resp.GetBatteryInfo()
			if tt.wantErr {
				require.Error(t, err, "GetBatteryInfo() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "GetBatteryInfo() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
