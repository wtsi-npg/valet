/*
 * Copyright (C) 2020, 2021. Genome Research Ltd. All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @file report.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	ex "github.com/wtsi-npg/extendo/v2"
)

const dutyTimeField = "Duty Time"
const trackingIDField = "Tracking ID"

type MinKNOWReport struct {
	Path                string // The path of the report
	DeviceID            string `json:"device_id"`            // The device ID (flowcell position)
	DeviceType          string `json:"device_type"`          // The device type e.g. promethion
	DistributionVersion string `json:"distribution_version"` // The MinKNOW version
	FlowcellID          string `json:"flow_cell_id"`         // The flowcell ID
	GuppyVersion        string `json:"guppy_version"`        // The Guppy basecaller version
	Hostname            string `json:"hostname"`             // The sequencing instrument hostname
	ProtocolGroupID     string `json:"protocol_group_id"`    // The user-supplied experiment name
	RunID               string `json:"run_id"`               // The automatically generated run ID
	SampleID            string `json:"sample_id"`            // The user-supplied sample ID
}

var gridionDeviceIDRegex = regexp.MustCompile(`^(?:GA|X)(\d)`)

var promethion24DeviceIDMap = map[string]int{
	"1A": 1,
	"1B": 2,
	"1C": 3,
	"1D": 4,
	"1E": 5,
	"1F": 6,
	"1G": 7,
	"1H": 8,
	"2A": 9,
	"2B": 10,
	"2C": 11,
	"2D": 12,
	"2E": 13,
	"2F": 14,
	"2G": 15,
	"2H": 16,
	"3A": 17,
	"3B": 18,
	"3C": 19,
	"3D": 20,
	"3E": 21,
	"3F": 22,
	"3G": 23,
	"3H": 24,
}

// ParseMinKNOWReport parses a file at path and extracts MinKNOW run metadata
// from it.
func ParseMinKNOWReport(path string) (MinKNOWReport, error) {
	var report MinKNOWReport

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return report, err
	}

	text := string(bytes)
	ti := strings.Index(text, trackingIDField)
	if ti < 0 {
		return report, errors.Errorf("failed to find %s in report file %s",
			trackingIDField, path)
	}

	di := strings.Index(text, dutyTimeField)
	if di < 0 {
		return report, errors.Errorf("failed to find %s in report file %s",
			dutyTimeField, path)
	}

	targetRegion := text[ti+len(trackingIDField) : di]
	targetRegion = strings.ReplaceAll(targetRegion, "=", "")

	if err = json.Unmarshal([]byte(targetRegion), &report); err != nil {
		return MinKNOWReport{}, err
	}
	report.Path = path

	return report, nil
}

// AsMetadata returns the report content as iRODS AVUs.
func (report MinKNOWReport) AsMetadata() []ex.AVU {
	avus := []ex.AVU{
		{Attr: "device_id", Value: report.DeviceID},
		{Attr: "device_type", Value: report.DeviceType},
		{Attr: "distribution_version", Value: report.DistributionVersion},
		{Attr: "flowcell_id", Value: report.FlowcellID},
		{Attr: "guppy_version", Value: report.GuppyVersion},
		{Attr: "hostname", Value: report.Hostname},
		{Attr: "protocol_group_id", Value: report.ProtocolGroupID},
		{Attr: "run_id", Value: report.RunID},
		{Attr: "sample_id", Value: report.SampleID},
	}

	for i := range avus {
		avus[i] = avus[i].WithNamespace(OxfordNanoporeNamespace)
	}

	return avus
}

// AsEnhancedMetadata returns the report as iRODS AVUs. It returns all the AVUs
// of AsMetadata with some additional members:
//
// The value of 'protocol_group_id' is duplicated under the attribute
// 'experiment_name'.
//
// The value of 'device_id' is normalized to a position (in the range 1-5 for
// GridION, representing slot position on the instrument). The device ID may
// be of the form "GAn0000" or "Xn" (for GridION), where n is the position.
// The value is added under the attribute 'instrument_slot'
//
// Slot positions are more complex for the PromethION as they are arranged in a
// grid and therefore have an X and Y position. The PromethION beta and
// PromethION-24 have different nomenclature.
//
// For the PromethION-24 we are following the column-major order used by ONT's
// MinKNOW API i.e. 1A - 1H, 2A - 2H, 3A - 3H.
//
func (report MinKNOWReport) AsEnhancedMetadata() ([]ex.AVU, error) {
	avus := report.AsMetadata()

	if report.DeviceType == "gridion" {
		deviceID := report.DeviceID
		idMatch := gridionDeviceIDRegex.FindStringSubmatch(deviceID)
		if idMatch == nil {
			return avus, errors.Errorf("Failed to parse device ID '%s'",
				deviceID)
		}

		slot := ex.AVU{Attr: "instrument_slot", Value: idMatch[1]}.
			WithNamespace(OxfordNanoporeNamespace)
		avus = append(avus, slot)
	}

	if report.DeviceType == "promethion" {
		deviceID := report.DeviceID
		id, ok := promethion24DeviceIDMap[deviceID]
		if !ok {
			return avus, errors.Errorf("Failed to parse device ID '%s'",
				deviceID)
		}

		slot := ex.AVU{Attr: "instrument_slot", Value: strconv.Itoa(id)}.
			WithNamespace(OxfordNanoporeNamespace)
		avus = append(avus, slot)
	}

	expt := ex.AVU{Attr: "experiment_name", Value: report.ProtocolGroupID}.
		WithNamespace(OxfordNanoporeNamespace)

	avus = append(avus, expt)

	return avus, nil
}
