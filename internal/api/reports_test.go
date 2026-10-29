package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealthReport(t *testing.T) {
	t.Parallel()
	responseData := map[string]any{
		"resultCode":  "200S00",
		"remoteInfos": []map[string]any{{"OccurrenceDate": "20261028120000"}},
	}
	server := createSuccessServer(t, "/"+EndpointGetHealthReport, responseData)
	defer server.Close()

	result, err := createTestClient(t, server.URL).GetHealthReport(t.Context(), "INTERNAL123")
	require.NoError(t, err)
	assert.Equal(t, ResultCodeSuccess, result["resultCode"])
	assert.Len(t, result["remoteInfos"], 1)
}

func TestGetHVACSetting(t *testing.T) {
	t.Parallel()
	responseData := map[string]any{"resultCode": "200S00", "hvacSettings": map[string]any{"Temperature": 21}}
	server := createSuccessServer(t, "/"+EndpointGetHVACSetting, responseData)
	defer server.Close()

	result, err := createTestClient(t, server.URL).GetHVACSetting(t.Context(), "INTERNAL123")
	require.NoError(t, err)
	assert.Contains(t, result, "hvacSettings")
}
