package cli

import (
	"encoding/json"
	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatusSeparateReportTimes(t *testing.T) {
	t.Parallel()
	v := &api.VehicleStatusResponse{AlertInfos: []api.AlertInfo{{OccurrenceDate: "20261027001037"}}, RemoteInfos: []api.RemoteInfo{{OccurrenceDate: "20261026024735"}}}
	e := &api.EVVehicleStatusResponse{ResultData: []api.EVResultData{{OccurrenceDate: "20261028183125"}}}
	s, err := displayAllStatusJSON(v, e, VehicleInfo{})
	require.NoError(t, err)
	var d struct {
		Reported map[string]string `json:"reported_at"`
	}
	require.NoError(t, json.Unmarshal([]byte(s), &d))
	assert.Equal(t, map[string]string{"doors": "20261027001037", "vehicle": "20261026024735", "ev": "20261028183125"}, d.Reported)
}
