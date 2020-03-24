/*
 * Copyright (C) 2020. Genome Research Ltd. All rights reserved.
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
	"strings"

	ex "github.com/kjsanger/extendo/v2"
	"github.com/pkg/errors"
)

const dutyTimeField = "Duty Time"
const trackingIDField = "Tracking ID"

type MinKNOWReport struct {
	DeviceID            string `json:"device_id"`
	DeviceType          string `json:"device_type"`
	DistributionVersion string `json:"distribution_version"`
	FlowcellID          string `json:"flow_cell_id"`
	GuppyVersion        string `json:"guppy_version"`
	Hostname            string `json:"hostname"`
	ProtocolGroupID     string `json:"protocol_group_id"`
	RunID               string `json:"run_id"`
	SampleID            string `json:"sample_id"`
}

func ParseReport(path string) (MinKNOWReport, error) {
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

	return report, nil
}

func (report MinKNOWReport) AsMetadata() []ex.AVU {
	return []ex.AVU{
		{Attr:"device_id", Value: report.DeviceID},
		{Attr: "device_type", Value: report.DeviceType},
		{Attr: "distribution_version", Value: report.DistributionVersion},
		{Attr: "flowcell_id", Value: report.FlowcellID},
		{Attr: "guppy_version", Value: report.GuppyVersion},
		{Attr: "hostname", Value: report.Hostname},
		{Attr: "protocol_group_id", Value: report.ProtocolGroupID},
		{Attr: "run_id", Value: report.RunID},
		{Attr: "sample_id", Value: report.SampleID},
	}
}
