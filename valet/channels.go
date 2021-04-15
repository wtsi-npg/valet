/*
 * Copyright (C) 2019, 2021. Genome Research Ltd. All rights reserved.
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
 * @file channels.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import "github.com/wtsi-npg/logshim"

// MergeFileChannels merges values from its two input channels x and y for as
// long as at least one of them is open. One both x and y have been closed, the
// channel returned will be closed by this function. The caller should not
// close the returned channel themselves.
func MergeFileChannels(
	x <-chan FilePath,
	y <-chan FilePath) chan FilePath {
	merged := make(chan FilePath)

	log := logshim.GetLogger()

	go func() {
		defer close(merged)

		for x != nil || y != nil {
			select {
			case p, ok := <-x:
				if ok {
					log.Debug().Msg("merging an x path")
					merged <- p
				} else {
					log.Debug().Msg("x was closed")
					x = nil
				}

			case p, ok := <-y:
				if ok {
					log.Debug().Msg("merging a y path")
					merged <- p
				} else {
					log.Debug().Msg("y was closed")
					y = nil
				}
			}
		}
	}()

	return merged
}

// MergeErrorChannels merges values from its two input channels x and y for as
// long as at least one of them is open. One both x and y have been closed, the
// channel returned will be closed by this function. The caller should not
// close the returned channel themselves.
func MergeErrorChannels(x <-chan error, y <-chan error) chan error {
	merged := make(chan error)

	go func() {
		defer close(merged)

		for x != nil || y != nil {
			select {
			case p, ok := <-x:
				if ok {
					merged <- p
				} else {
					x = nil
				}

			case p, ok := <-y:
				if ok {
					merged <- p
				} else {
					y = nil
				}
			}
		}
	}()

	return merged
}
