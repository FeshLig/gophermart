package flags_test

import (
	"testing"

	"github.com/FeshLig/gophermart/internal/flags"
	"github.com/stretchr/testify/require"
)

func TestRunAddress_Set_Success(t *testing.T) {
	var addr flags.RunAddress

	err := addr.Set("127.0.0.1:8080")
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1", addr.Host)
	require.Equal(t, 8080, addr.Port)
}
func TestRunAddress_Set_DefaultHost(t *testing.T) {
	var addr flags.RunAddress

	err := addr.Set(":8080")
	require.NoError(t, err)

	require.Equal(t, "localhost", addr.Host)
	require.Equal(t, 8080, addr.Port)
}
func TestRunAddress_Set_InvalidFormat(t *testing.T) {
	var addr flags.RunAddress

	err := addr.Set("127.0.0.1")
	require.Error(t, err)
}
func TestRunAddress_Set_InvalidPort(t *testing.T) {
	var addr flags.RunAddress

	err := addr.Set("127.0.0.1:abc")
	require.Error(t, err)
}

func TestRunAddress_String(t *testing.T) {
	addr := flags.RunAddress{
		Host: "localhost",
		Port: 8080,
	}

	require.Equal(t, "localhost:8080", addr.String())
}
