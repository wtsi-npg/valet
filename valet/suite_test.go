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
 * @file suite_test.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet_test

import (
	"os"
	"testing"
	"time"

	"github.com/kjsanger/logshim-zerolog/zlog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"

	logs "github.com/kjsanger/logshim"
)

func TestValet(t *testing.T) {
	loggerImpl := zlog.New(os.Stderr, logs.ErrorLevel)

	writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	consoleLogger := loggerImpl.Logger.Output(zerolog.SyncWriter(writer))
	loggerImpl.Logger = &consoleLogger

	logs.InstallLogger(loggerImpl)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Valet Suite")
}


