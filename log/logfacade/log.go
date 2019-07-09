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
 * @file log.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package logfacade

import (
	"sync"
	"time"
)

// Level indicates a logging level. The smaller its numeric value, the
// more severe is the event logged.
type Level uint8

const (
	// ErrorLevel sets the logging level to Error.
	ErrorLevel Level = iota + 2
	// WarnLevel sets the logging level to Warn.
	WarnLevel
	// NoticeLevel sets the logging level to Notice. This level is present to
	// so that the logging level values match the normal Syslog severity
	// levels. In practice the logging backend may substitute Notice with Info
	// level.
	NoticeLevel
	// InfoLevel sets the logging level to Info.
	InfoLevel
	// DebugLevel sets the logging level to Debug.
	DebugLevel
)

// Levels returns a slice containing all logging Levels, in ascending numeric
// order.
func Levels() []Level {
	return []Level{ErrorLevel, WarnLevel, NoticeLevel, InfoLevel, DebugLevel}
}

// A Message is a unit of log information that is built by successive method
// calls and completed by calling the Msg method.
type Message interface {
	// Err adds an error field to the Message.
	Err(err error) Message
	// Bool adds a named bool field to the Message.
	Bool(key string, val bool) Message
	// Int adds a named int field to the Message.
	Int(key string, val int) Message
	// Int64 adds a named int64 field to the Message.
	Int64(key string, val int64) Message
	// Uint64 adds a named uint64 field to the Message.
	Uint64(key string, val uint64) Message
	// Item adds a named string field to the Message.
	Str(key string, val string) Message
	// Time adds a named Time field to the Message.
	Time(key string, val time.Time) Message
	// Msg completes and sends the Message.
	Msg(format string)
	// Msgf completes and sends the Message with printf support.
	Msgf(format string, a ...interface{})
}

// A Logger is an object that generates new log Messages.
type Logger interface {
	// Impl returns the underlying Logger implementation (such as the Golang
	// standard log.Logger or ZeroLog zerolog.Logger). This allows full access
	// for methods not otherwise exposed.
	// Impl() interface{}
	Name() string
	// Err creates a new Error level Message and adds the error.
	Err(err error) Message
	// Error creates a new Error level Message.
	Error() Message
	// Warn creates a new Warn level Message.
	Warn() Message
	// Notice creates a new Notice level Message.
	Notice() Message
	// Info creates a new Info level Message.
	Info() Message
	// Debug creates a new Debug level Message.
	Debug() Message
}

var logger Logger
var installOnce sync.Once

func InstallLogger(lg Logger) Logger {
	installOnce.Do(func() {
		lg.Info().Msgf("installing %s as global logger", lg.Name())
		logger = lg
	})
	return GetLogger()
}

func GetLogger() Logger {
	if logger == nil {
		panic("Cannot get a global logger instance because none has been " +
			"installed. You should call InstallLogger before calling GetLogger")
	}

	return logger
}
