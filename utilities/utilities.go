/*
 * Copyright (C) 2019. Genome Research Ltd. All rights reserved.
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
 * @file utilities.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package utilities

import (
	"io"
	"os"
	"strings"
)

type combinedError struct {
	errors []error
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func CopyFile(from string, to string, perm os.FileMode) (err error) { // NRV
	var src, dst *os.File

	src, err = os.Open(from)
	if err != nil {
		return
	}

	defer func() {
		err = CombineErrors(err, src.Close())
	}()

	dst, err = os.OpenFile(to, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return
	}

	defer func() {
		err = CombineErrors(err, dst.Close())
	}()

	_, err = io.Copy(dst, src)

	return
}

func CombineErrors(errors ...error) error {
	var errs []error
	for _, e := range errors {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	return &combinedError{errs}
}

func (err *combinedError) Error() string {
	var bld strings.Builder
	bld.WriteString("combined errors: [")

	last := len(err.errors) - 1
	for i, e := range err.errors {
		bld.WriteString(e.Error())
		if i < last {
			bld.WriteString(", ")
		}
	}
	bld.WriteString("]")

	return bld.String()
}

func (err *combinedError) Errors() []error {
	return make([]error, len(err.errors))
}
