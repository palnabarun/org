package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/org"
)

func loadOrgs(o Options, mergeTeams bool) (map[string]org.Config, error) {
	config := make(map[string]org.Config)
	for _, org := range o.Orgs {
		relativeConfigPath := fmt.Sprintf(orgConfigPathFormat, org)
		configPath := filepath.Join(o.RepoRoot, relativeConfigPath)

		cfg, err := readConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("error in %s: %v", configPath, err)
		}

		if mergeTeams {
			if cfg.Teams == nil {
				cfg.Teams = make(map[string]org.Team)
			}
			prefix := filepath.Dir(configPath)
			err := filepath.Walk(prefix, func(path string, info os.FileInfo, err error) error {
				switch {
				case path == prefix:
					return nil // Skip base dir
				case info.IsDir() && filepath.Dir(path) != prefix:
					logrus.Infof("Skipping %s and its children", path)
					return filepath.SkipDir // Skip prefix/foo/bar/ dirs
				case !info.IsDir() && filepath.Dir(path) == prefix:
					return nil // Ignore prefix/foo files
				case filepath.Base(path) == "teams.yaml":
					teamCfg, err := readConfig(path)
					if err != nil {
						return fmt.Errorf("error in %s: %v", path, err)
					}

					for name, team := range teamCfg.Teams {
						cfg.Teams[name] = team
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("merge teams %s: %v", configPath, err)
			}
		} else {
			cfg.Teams = nil
		}
		config[org] = *cfg
	}
	return config, nil
}
