package flags_test

import (
	"testing"

	"github.com/FeshLig/gophermart/internal/flags"
	"github.com/stretchr/testify/require"
)

func TestDatabaseURI_Set_Success(t *testing.T) {
	var uri flags.DatabaseURI

	err := uri.Set("postgres://user:pass@localhost:5432/db")
	require.NoError(t, err)

	require.Equal(t, "postgres://user:pass@localhost:5432/db", string(uri))
}

func TestDatabaseURI_Set_Empty(t *testing.T) {
	var uri flags.DatabaseURI

	err := uri.Set("")
	require.Error(t, err)
	require.EqualError(t, err, "database URI is empty")
}

func TestDatabaseURI_String(t *testing.T) {
	uri := flags.DatabaseURI("postgres://localhost:5432/testdb")

	require.Equal(t, "postgres://localhost:5432/testdb", uri.String())
}
