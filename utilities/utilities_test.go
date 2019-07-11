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
 * @file utilities_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package utilities

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCombineErrors(t *testing.T) {
	err1 := errors.New("Error 1")
	err2 := errors.New("Error 2")
	err3 := errors.New("Error 3")

	cerr0 := CombineErrors(nil)
	assert.Nil(t, cerr0)

	cerr1 := CombineErrors(err1)
	if assert.Error(t, cerr1) {
		assert.Equal(t, &combinedError{[]error{err1}}, cerr1)
	}

	cerr2 := CombineErrors(nil, err1)
	if assert.Error(t, cerr2) {
		assert.Equal(t, &combinedError{[]error{err1}}, cerr2)
	}

	cerr3 := CombineErrors(err1, err2, err3)
	if assert.Error(t, cerr3) {
		assert.Equal(t, &combinedError{[]error{err1, err2, err3}}, cerr3)
	}
}
