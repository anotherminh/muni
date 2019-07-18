/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

var listroutesCmd = &cobra.Command{
	Use:   "listroutes",
	Short: "List all the muni routes",
	Long:  `muni listroutes`,
	Run: func(cmd *cobra.Command, args []string) {
		routes, err := listRoutes()
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			for _, route := range routes {
				fmt.Printf("%s - %s\n", route.Tag, route.Title)
			}
		}
	},
}

const listRoutesPath = "http://webservices.nextbus.com/service/publicXMLFeed?command=routeList&a=sf-muni"

type Route struct {
	Tag   string `xml:"tag,attr"`
	Title string `xml:"title,attr"`
}

func listRoutes() ([]Route, error) {
	resp, err := http.Get(listRoutesPath)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	type Body struct {
		Routes []Route `xml:"route"`
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := new(Body)
	xmlErr := xml.Unmarshal([]byte(data), &body)
	return body.Routes, xmlErr
}

func init() {
	rootCmd.AddCommand(listroutesCmd)
}
