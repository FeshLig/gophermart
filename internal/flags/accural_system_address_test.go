package flags_test

import (
	"testing"

	"github.com/FeshLig/gophermart/internal/flags"
	"github.com/stretchr/testify/require"
)

func TestAccuralSystemAddress_Set_HostPort(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("localhost:8080")
	require.NoError(t, err)

	require.Equal(t, "localhost", addr.Host)
	require.Equal(t, 8080, addr.Port)
}

func TestAccuralSystemAddress_Set_HTTP(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("http://example.com:9000")
	require.NoError(t, err)

	require.Equal(t, "example.com", addr.Host)
	require.Equal(t, 9000, addr.Port)
}

func TestAccuralSystemAddress_Set_HTTPS(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("https://example.com:443")
	require.NoError(t, err)

	require.Equal(t, "example.com", addr.Host)
	require.Equal(t, 443, addr.Port)
}

func TestAccuralSystemAddress_Set_URLMissingPort(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("http://example.com")
	require.Error(t, err)
	require.EqualError(t, err, "port is required")
}

func TestAccuralSystemAddress_Set_InvalidHostPort(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("invalid-format")
	require.Error(t, err)
}

func TestAccuralSystemAddress_Set_InvalidPort(t *testing.T) {
	var addr flags.AccuralSystemAddress

	err := addr.Set("localhost:abc")
	require.Error(t, err)
}

func TestAccuralSystemAddress_String(t *testing.T) {
	addr := flags.AccuralSystemAddress{
		Host: "localhost",
		Port: 8080,
	}

	require.Equal(t, "localhost:8080", addr.String())
}

func TestAccuralSystemAddress_String_Empty(t *testing.T) {
	addr := flags.AccuralSystemAddress{}

	require.Equal(t, "", addr.String())
}

func TestAccuralSystemAddress_URL(t *testing.T) {
	addr := flags.AccuralSystemAddress{
		Host: "localhost",
		Port: 8080,
	}

	require.Equal(t, "http://localhost:8080", addr.URL())
}

func TestAccuralSystemAddress_URL_Empty(t *testing.T) {
	addr := flags.AccuralSystemAddress{}

	require.Equal(t, "", addr.URL())
}
