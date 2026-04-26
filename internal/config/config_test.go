package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/FeshLig/gophermart/internal/config"
	"github.com/stretchr/testify/require"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestGetConfig_Defaults(t *testing.T) {
	resetFlags()

	// чистим env
	os.Unsetenv("RUN_ADDRESS")
	os.Unsetenv("DATABASE_URI")
	os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

	// чистим args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd"} // важно: без флагов

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	require.Equal(t, "localhost", cfg.Address.Host)
	require.Equal(t, 8080, cfg.Address.Port)

	require.Equal(t, "", string(cfg.DatabaseURI))

	require.Equal(t, "", cfg.AccuralAddress.Host)
	require.Equal(t, 0, cfg.AccuralAddress.Port)
}

func TestGetConfig_FromEnv(t *testing.T) {
	resetFlags()

	t.Setenv("RUN_ADDRESS", "127.0.0.1:9000")
	t.Setenv("DATABASE_URI", "postgres://user:pass@localhost:5432/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "acc:7000")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1", cfg.Address.Host)
	require.Equal(t, 9000, cfg.Address.Port)

	require.Equal(t, "postgres://user:pass@localhost:5432/db", string(cfg.DatabaseURI))

	require.Equal(t, "acc", cfg.AccuralAddress.Host)
	require.Equal(t, 7000, cfg.AccuralAddress.Port)
}

func TestGetConfig_FromFlags(t *testing.T) {
	resetFlags()

	os.Unsetenv("RUN_ADDRESS")
	os.Unsetenv("DATABASE_URI")
	os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"cmd",
		"-a=0.0.0.0:8081",
		"-d=postgres://u:p@localhost:5432/db",
		"-r=localhost:7001",
	}

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0", cfg.Address.Host)
	require.Equal(t, 8081, cfg.Address.Port)

	require.Equal(t, "postgres://u:p@localhost:5432/db", string(cfg.DatabaseURI))

	require.Equal(t, "localhost", cfg.AccuralAddress.Host)
	require.Equal(t, 7001, cfg.AccuralAddress.Port)
}

func TestGetConfig_InvalidEnv(t *testing.T) {
	resetFlags()

	t.Setenv("RUN_ADDRESS", "invalid-address")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	_, err := config.GetConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse env")
}
