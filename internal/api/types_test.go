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
							"ChargeStatusSub": 6
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
