package rpcbus

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCBus(t *testing.T) {
	t.Parallel()

	// example methods:
	// 		ret.unitList
	// 		alerts.get_dictionary
	// 		agent.get_rru_info
	// 		rpcbus.registerClient
	// если result, error - это response
	// если id - это request
	// если id нет - это notif

	rpcBusClient, err := NewClient("10.12.0.14:8080")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, rpcBusClient.Close())
	})

	resp, err := rpcBusClient.Call(t.Context(), "alerts.get_dictionary", nil)
	require.NoError(t, err)
	require.NotEmpty(t, resp)
}

func BenchmarkGenUUID(b *testing.B) { // 499.0 ns/op; 486.8 ns/op; 497.0 ns/op
	for i := 0; i < b.N; i++ {
		_ = uuid.NewString()
	}
}

func BenchmarkGenTimeNano(b *testing.B) { // 188.9 ns/op; 190.3 ns/op; 223.6 ns/op
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format(time.RFC3339Nano)
	}
}
