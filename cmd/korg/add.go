/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//
//var (
//	addHelpText = `
//Adds users to GitHub orgs and/or teams
//
//Add user to specified orgs:
//
//	korg add <github username> --org kubernetes --org kubernetes-sigs
//	korg add <github username> --org kubernetes,kubernetes-sigs
//
//Note: Adding to teams is currently unsupported.
//	`
//)
//
//func AddMemberToOrgs(username string, options Options) error {
//	if invalidOrgs := findInvalidOrgs(options.Orgs); len(invalidOrgs) > 0 {
//		return fmt.Errorf("specified invalid orgs: %s", strings.Join(invalidOrgs, ", "))
//	}
//
//	if !options.Confirm {
//		fmt.Println("!!! running in dry-run mode. pass --confirm to persist changes.")
//	}
//
//	configsModified := []string{}
//	for _, org := range options.Orgs {
//		fmt.Printf("adding %s to %s org\n", username, org)
//
//		relativeConfigPath := fmt.Sprintf(orgConfigPathFormat, org)
//		configPath := filepath.Join(options.RepoRoot, relativeConfigPath)
//
//		config, err := readConfig(configPath)
//		if err != nil {
//			return fmt.Errorf("reading config: %s", err)
//		}
//
//		if stringInSliceCaseAgnostic(config.Members, username) || stringInSliceCaseAgnostic(config.Admins, username) {
//			return fmt.Errorf("user %s already exists in org %s", username, org)
//		}
//
//		newMembers := append(config.Members, username)
//		config.Members = newMembers
//		caseAgnosticSort(config.Members)
//
//		if options.Confirm {
//			fmt.Printf("saving config for %s org\n", org)
//			if err := saveConfig(configPath, config); err != nil {
//				return fmt.Errorf("saving config: %s", err)
//			}
//		}
//
//		configsModified = append(configsModified, relativeConfigPath)
//	}
//
//	if options.Confirm {
//		fmt.Println("committing changes")
//
//		message := fmt.Sprintf("add %s to %s", username, strings.Join(options.Orgs, ", "))
//		if err := commitChanges(options.RepoRoot, configsModified, message); err != nil {
//			return fmt.Errorf("committing changes: %s", err)
//		}
//	}
//	return nil
//}
