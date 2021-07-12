/*
Copyright 2021 The Kubernetes Authors.

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

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	validOrgs = []string{
		"kubernetes",
		"kubernetes-client",
		"kubernetes-csi",
		"kubernetes-incubator",
		"kubernetes-retired",
		"kubernetes-sigs",
	}
)

type Options struct {
	// global options
	Confirm bool
	OrgRoot string

	// add options
	Orgs  []string
	Teams []string
}

func AddMemberToOrgs(username string, orgs []string) error {
	return errors.New("not implemented")
}

func AddMemberToTeams(username string, teams []string) error {
	return errors.New("not implemented")
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "korg",
		Short: "Manage kubernetes organizations",
	}

	o := Options{}
	rootCmd.PersistentFlags().BoolVar(&o.Confirm, "confirm", false, "confirm the changes")
	rootCmd.PersistentFlags().StringVar(&o.OrgRoot, "root", ".", "root of the k/org repo")

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add members to org or teams",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("add only adds one user at a time. specified %d", len(args))
			}

			if len(o.Orgs) == 0 && len(o.Teams) == 0 {
				return fmt.Errorf("please specify either --org or --team or both")
			}

			invalidOrgs := []string{}
			for _, org := range o.Orgs {
				if !stringInSlice(validOrgs, org) {
					invalidOrgs = append(invalidOrgs, org)
				}
			}
			if len(invalidOrgs) > 0 {
				return fmt.Errorf("specified invalid orgs: %s", strings.Join(invalidOrgs, ", "))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			user := args[1]
			if len(o.Orgs) > 0 {
				return AddMemberToOrgs(user, o.Orgs)
			}

			if len(o.Orgs) > 0 {
				return AddMemberToTeams(user, o.Teams)
			}

			return nil
		},
	}

	addCmd.Flags().StringSliceVar(&o.Orgs, "org", []string{}, "orgs to add the user to")
	addCmd.Flags().StringSliceVar(&o.Teams, "team", []string{}, "teams to add the user to")
	rootCmd.AddCommand(addCmd)
	// Note: In future, add more korg commands remove/delete

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
