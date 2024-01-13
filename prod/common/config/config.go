/*
 * Utility library for retrieving configuration.
 *     Thin wrapper around the great env variable
 *     config library from Carlos Becker.
 */

package config

import (
	"github.com/caarlos0/env/v10"

	"log"
)

// attempt to load config. panic in case of failure
func LoadConfig(cfg interface{}, prefix string) {
	opts := env.Options{
		Prefix: prefix,
	}
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatalf("could not parse environment config: %v", err)
	}
	log.Printf("Starting process with config %+v", cfg)
}
