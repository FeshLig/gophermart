package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/FeshLig/gophermart/internal/flags"
)

type Config struct {
	Address        flags.RunAddress
	DatabaseURI    flags.DatabaseURI
	AccuralAddress flags.AccuralSystemAddress
}

func GetConfig() Config {
	config := Config{
		Address: flags.RunAddress{
			Host: "localhost",
			Port: 8080,
		},
		AccuralAddress: flags.AccuralSystemAddress{
			Host: "",
			Port: 0,
		},
		DatabaseURI: "",
	}

	parseFlags(&config)
	fmt.Printf("ADDRESS: %s", config.Address.String())
	parseEnv(&config)
	fmt.Printf("ADDRESS: %s", config.Address.String())

	return config
}

func parseFlags(config *Config) {
	flag.Var(&config.Address, "a", "net address host:port")
	flag.Var(&config.DatabaseURI, "d", "postgres dsn (format: postgres://user:password@host:port/dbname)")
	flag.Var(&config.AccuralAddress, "r", "accural system address host:port")

	flag.Parse()
}

func parseEnv(config *Config) error {
	addrStr, _ := os.LookupEnv("RUN_ADDRESS")
	fmt.Printf("ADDRESS: %s", addrStr)

	if addrStr, ok := os.LookupEnv("RUN_ADDRESS"); ok && addrStr != "" {
		if err := config.Address.Set(addrStr); err != nil {
			return fmt.Errorf("invalid value of RUN_ADDRESS: %w", err)
		}
	}

	if databaseURIStr, ok := os.LookupEnv("DATABASE_URI"); ok && databaseURIStr != "" {
		if err := config.DatabaseURI.Set(databaseURIStr); err != nil {
			return fmt.Errorf("wrong value of DATABASE_URI: %w", err)
		}
	}

	if accuralAddressStr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok && accuralAddressStr != "" {
		if err := config.AccuralAddress.Set(accuralAddressStr); err != nil {
			return fmt.Errorf("wrong value of ACCRUAL_SYSTEM_ADDRESS: %w", err)
		}
	}

	return nil
}
