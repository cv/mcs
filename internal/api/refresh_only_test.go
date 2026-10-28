package api

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRefreshOnlyRefusesNewSessions(t *testing.T) {
	t.Parallel()
	c, err := NewClient("test@example.com", "unused", RegionMNAO)
	require.NoError(t, err)
	c.SetRefreshOnly(true)
	require.ErrorContains(t, c.Login(context.Background()), "password login disabled")
	require.ErrorContains(t, c.attachSession(context.Background()), "session attachment disabled")
}
