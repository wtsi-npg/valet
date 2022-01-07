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
 * @file report_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	ex "github.com/wtsi-npg/extendo/v2"
)

func TestParsePromethIONBetaReport(t *testing.T) {
	path := "./testdata/valet/report_PAE48813_20200130_0940_16917585.md"
	report, err := ParseMinKNOWReport(path)
	if assert.NoError(t, err) {
		assert.Equal(t, "2-E1-H1", report.DeviceID)
		assert.Equal(t, "promethion", report.DeviceType)
		assert.Equal(t, "19.12.5", report.DistributionVersion)
		assert.Equal(t, "PAE48813", report.FlowcellID)
		assert.Equal(t, "3.2.8+bd67289", report.GuppyVersion)
		assert.Equal(t, "PCT0016", report.Hostname)
		assert.Equal(t, "mMelMel3", report.ProtocolGroupID)
		assert.Equal(t, "52a0d863bccd1d78530c425e8077150d5391fc34", report.RunID)
		assert.Equal(t, "mMelMel3", report.SampleID)
	}
}

func TestParsePromethION24Report(t *testing.T) {
	path := "./testdata/valet/report_PAH48449_20211215_1420_227842f4.md"
	report, err := ParseMinKNOWReport(path)
	if assert.NoError(t, err) {
		assert.Equal(t, "1A", report.DeviceID)
		assert.Equal(t, "promethion", report.DeviceType)
		assert.Equal(t, "21.10.8", report.DistributionVersion)
		assert.Equal(t, "5.0.17+99baa5b27", report.GuppyVersion)
		assert.Equal(t, "PC24B148", report.Hostname)
		assert.Equal(t, "lambda_p24_all_positions", report.ProtocolGroupID)
		assert.Equal(t, "39ae988ba41499479e0dc1fcae29fa4059701d08", report.RunID)
		assert.Equal(t, "lambda_151221_1", report.SampleID)
	}
}

func TestParseGridIONReport(t *testing.T) {
	path := "./testdata/valet/report_ABQ808_20200204_1257_e2e93dd1.md"
	report, err := ParseMinKNOWReport(path)
	if assert.NoError(t, err) {
		assert.Equal(t, "X2", report.DeviceID)
		assert.Equal(t, "gridion", report.DeviceType)
		assert.Equal(t, "19.12.2", report.DistributionVersion)
		assert.Equal(t, "ABQ808", report.FlowcellID)
		assert.Equal(t, "3.2.8+bd67289", report.GuppyVersion)
		assert.Equal(t, "GXB02004", report.Hostname)
		assert.Equal(t, "85", report.ProtocolGroupID)
		assert.Equal(t, "5531cbcf622d2d98dbff00af0261c6f19f91340f", report.RunID)
		assert.Equal(t, "DN615089W_B1", report.SampleID)
	}
}

func TestEnhancedPromethIONBetaMetadata(t *testing.T) {
	path := "./testdata/valet/report_PAE48813_20200130_0940_16917585.md"
	report, _ := ParseMinKNOWReport(path)
	metadata, err := report.AsEnhancedMetadata()
	if assert.NoError(t, err) {
		expected := []ex.AVU{
			{Attr: "ont:device_id", Value: "2-E1-H1"},
			{Attr: "ont:device_type", Value: "promethion"},
			{Attr: "ont:distribution_version", Value: "19.12.5"},
			{Attr: "ont:flowcell_id", Value: "PAE48813"},
			{Attr: "ont:guppy_version", Value: "3.2.8+bd67289"},
			{Attr: "ont:hostname", Value: "PCT0016"},
			{Attr: "ont:protocol_group_id", Value: "mMelMel3"},
			{Attr: "ont:run_id", Value: "52a0d863bccd1d78530c425e8077150d5391fc34"},
			{Attr: "ont:sample_id", Value: "mMelMel3"},
			// {Attr:"ont:instrument_slot", Value:"2"} TODO: slot not yet supported
			{Attr: "ont:experiment_name", Value: "mMelMel3"}}

		assert.ElementsMatch(t, expected, metadata)
	}
}

func TestEnhancedPromethION24Report(t *testing.T) {
	path := "./testdata/valet/report_PAH48449_20211215_1420_227842f4.md"
	report, err := ParseMinKNOWReport(path)
	metadata, err := report.AsEnhancedMetadata()
	if assert.NoError(t, err) {
		expected := []ex.AVU{
			{Attr: "ont:device_id", Value: "1A"},
			{Attr: "ont:device_type", Value: "promethion"},
			{Attr: "ont:distribution_version", Value: "21.10.8"},
			{Attr: "ont:flowcell_id", Value: "PAH48449"},
			{Attr: "ont:guppy_version", Value: "5.0.17+99baa5b27"},
			{Attr: "ont:hostname", Value: "PC24B148"},
			{Attr: "ont:protocol_group_id", Value: "lambda_p24_all_positions"},
			{Attr: "ont:run_id", Value: "39ae988ba41499479e0dc1fcae29fa4059701d08"},
			{Attr: "ont:sample_id", Value: "lambda_151221_1"},
			{Attr: "ont:instrument_slot", Value: "1"},
			{Attr: "ont:experiment_name", Value: "lambda_p24_all_positions"}}

		assert.ElementsMatch(t, expected, metadata)
	}
}

func TestEnhancedGridIONMetadata(t *testing.T) {
	path := "./testdata/valet/report_ABQ808_20200204_1257_e2e93dd1.md"
	report, _ := ParseMinKNOWReport(path)
	metadata, err := report.AsEnhancedMetadata()
	if assert.NoError(t, err) {
		expected := []ex.AVU{
			{Attr: "ont:device_id", Value: "X2"},
			{Attr: "ont:device_type", Value: "gridion"},
			{Attr: "ont:distribution_version", Value: "19.12.2"},
			{Attr: "ont:flowcell_id", Value: "ABQ808"},
			{Attr: "ont:guppy_version", Value: "3.2.8+bd67289"},
			{Attr: "ont:hostname", Value: "GXB02004"},
			{Attr: "ont:protocol_group_id", Value: "85"},
			{Attr: "ont:run_id", Value: "5531cbcf622d2d98dbff00af0261c6f19f91340f"},
			{Attr: "ont:sample_id", Value: "DN615089W_B1"},
			{Attr: "ont:instrument_slot", Value: "2"},
			{Attr: "ont:experiment_name", Value: "85"}}

		assert.ElementsMatch(t, expected, metadata)
	}
}

func TestPromethion24InstrumentSlots(t *testing.T) {
	paths, err := filepath.Glob("./testdata/valet/report_PAH48449_20211215*.md")
	if assert.NoError(t, err) {
		// The globbed paths are one report file for each slot in flowcell bank
		// "A", being "1A" to "1F"
		assert.Equal(t, 6, len(paths))

		// These should have slot values of 1 to 6
		for i, path := range paths {
			report, _ := ParseMinKNOWReport(path)
			metadata, err := report.AsEnhancedMetadata()
			if assert.NoError(t, err) {
				found := false
				for _, avu := range metadata {
					if avu.Attr == "ont:instrument_slot" {
						// Slots are indexed from 1
						assert.Equal(t, avu.Value, strconv.Itoa(i+1))
						found = true
						break
					}
				}
				assert.True(t, found, "Failed to parse a DeviceID",
					"ont:instrument_slot attribute missing from "+
						"metadata for %s", path)
			}
		}
	}
}
