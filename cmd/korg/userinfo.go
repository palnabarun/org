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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/hound-search/hound/client"
)

func findUserDetails(username string) {
	fmt.Printf("\n=== fetching details for user %s\n", username)

	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var u github.User

	err = json.Unmarshal(body, &u)
	if err != nil {
		panic(err)
	}

	if u.Company != nil {
		fmt.Println("Company:", *u.Company)
	} else {
		fmt.Println("Company: **Not Found**")
	}

	url = fmt.Sprintf("https://cs.k8s.io/api/v1/search?stats=fosho&repos=*&rng=:20&q=%s&i=fosho&files=OWNERS&excludeFiles=vendor/", username)
	resp, err = http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var r client.Response

	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}

	fmt.Println("Owner Files:")
	for repo, matches := range r.Results {
		for _, match := range matches.Matches {
			fmt.Printf("%s:%s\n", repo, match.Filename)
		}
	}

}
