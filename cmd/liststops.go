/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

type DirectionFlag struct {
	inbound  bool
	outbound bool
}

func (f *DirectionFlag) String() string {
	return fmt.Sprintf("inbound: %t, outbound: %t", f.inbound, f.outbound)
}

// can probably clean all this up
func (f *DirectionFlag) Set(value string) error {
	lowercase := strings.ToLower(value)
	if lowercase == "inbound" || lowercase == "i" {
		f.inbound = true
		f.outbound = false
	} else if lowercase == "outbound" || lowercase == "o" {
		f.inbound = false
		f.outbound = true
	} else if lowercase == "" {
		f.inbound = true
		f.outbound = true
	} else {
		errMsg := fmt.Sprintf("Please enter one of the following: inbound, i, outbound, o")
		return errors.New(errMsg)
	}
	return nil
}

func (f *DirectionFlag) Type() string {
	return "DirectionFlag"
}

// listStopsCmd represents the listStops command
var listStopsCmd = &cobra.Command{
	Use:   "liststops [routeTag]",
	Short: "List stops for a given MUNI route",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stops, err := listStops(args[0])
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			stops.printStops(directionFlag)
		}
	},
}

const listStopsPath = "http://webservices.nextbus.com/service/publicXMLFeed?command=routeConfig&a=sf-muni&r="

type Stop struct {
	Title  string `xml:"title,attr"`
	StopId string `xml:"stopId,attr"`
	Tag    string `xml:"tag,attr"`
}

type StopTag struct {
	Tag string `xml:"tag,attr"`
}

type Direction struct {
	Name     string    `xml:"name,attr"`
	StopTags []StopTag `xml:"stop"`
}

type StopsByDirection struct {
	inboundStops  []Stop
	outboundStops []Stop
}

func (stopsByDirection StopsByDirection) printStops(direction DirectionFlag) {
	if direction.inbound {
		fmt.Println("----------INBOUND STOPS----------")
		for _, stop := range stopsByDirection.inboundStops {
			fmt.Printf("%s - %s\n", stop.StopId, stop.Title)
		}
	}

	if direction.outbound {
		fmt.Println("----------OUTBOUND STOPS----------")
		for _, stop := range stopsByDirection.outboundStops {
			fmt.Printf("%s - %s\n", stop.StopId, stop.Title)
		}
	}
}

func listStops(routeTag string) (*StopsByDirection, error) {
	type Body struct {
		Stops      []Stop      `xml:"route>stop"`
		Directions []Direction `xml:"route>direction"`
	}

	// new(s) returns a pointer to the struct
	// go automatically de-references for us when accessing fields of the struct through the pointer
	stopsByDirection := new(StopsByDirection)

	resp, err := http.Get(listStopsPath + routeTag)
	defer resp.Body.Close()

	if err != nil {
		return stopsByDirection, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stopsByDirection, err
	}

	body := new(Body)
	xmlErr := xml.Unmarshal([]byte(data), &body)

	if xmlErr != nil {
		return stopsByDirection, xmlErr
	}

	stopsByTag := map[string]Stop{}
	for _, stop := range body.Stops {
		stopsByTag[stop.Tag] = stop
	}

	for _, direction := range body.Directions {
		stops := []Stop{}
		for _, stopTag := range direction.StopTags {
			stops = append(stops, stopsByTag[stopTag.Tag])
		}

		if direction.Name == "Outbound" {
			stopsByDirection.outboundStops = stops
		} else {
			stopsByDirection.inboundStops = stops
		}
	}

	return stopsByDirection, xmlErr
}

var directionFlag DirectionFlag

func init() {
	rootCmd.AddCommand(listStopsCmd)
	listStopsCmd.Flags().VarP(&directionFlag, "direction", "d", "Show only inbound or outbound stops")
}
