package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The current API accepts every remote command with MSG200001 (the car then
// carries it out); an unlock that the car performed must not report a failure.
func TestCommandAccepted(t *testing.T) {
	t.Parallel()
	server := createSuccessServer(t, "/"+EndpointDoorUnlock, map[string]any{"resultCode": ResultCodeRefreshAccepted})
	defer server.Close()

	require.NoError(t, createTestClient(t, server.URL).DoorUnlock(t.Context(), "INTERNAL123"))
}

func TestCommandRefused(t *testing.T) {
	t.Parallel()
	server := createSuccessServer(t, "/"+EndpointDoorLock, map[string]any{"resultCode": "920E00"})
	defer server.Close()

	require.ErrorContains(t, createTestClient(t, server.URL).DoorLock(t.Context(), "INTERNAL123"), "result code 920E00")
}
