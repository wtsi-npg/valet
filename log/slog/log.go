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

package slog

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	logf "valet/log/logfacade"
)

type levelName string

const (
	errorLevel  levelName = "ERROR"
	warnLevel   levelName = "WARN"
	// noticeLevel levelName = "NOTICE"
	infoLevel   levelName = "INFO"
	debugLevel  levelName = "DEBUG"
)

func translateLevel(level logf.Level) (levelName, error) {
	var (
		lvn levelName
		err error
	)

	switch level {
	case logf.ErrorLevel:
		lvn = errorLevel
	case logf.WarnLevel:
		lvn = warnLevel
	case logf.NoticeLevel:
		fallthrough
	case logf.InfoLevel:
		lvn = infoLevel
	case logf.DebugLevel:
		lvn = debugLevel
	default:
		lvn = warnLevel
		err = fmt.Errorf("invalid log level %d, defaulting to "+
			"WARN level", level)
	}

	return lvn, err
}

type StdLogger struct {
	name  string
	Level logf.Level
	*log.Logger
}

func New(writer io.Writer, level logf.Level) *StdLogger {
	lg := log.New(writer, "", log.LstdFlags|log.Lshortfile)

	_, err := translateLevel(level)
	if err != nil {
		log.Print(errorLevel, "log configuration error", err)
		level = logf.WarnLevel
	}

	return &StdLogger{"StdLog", level, lg}
}

func (log *StdLogger) Name() string {
	return log.name
}

func (log *StdLogger) Err(err error) logf.Message {
	effectiveLevel := logf.InfoLevel
	if err != nil {
		effectiveLevel = logf.ErrorLevel
	}

	active := log.Level >= effectiveLevel
	msg := &stdMessage{active, effectiveLevel, &strings.Builder{}}
	msg.Err(err)
	return msg
}

func (log *StdLogger) Error() logf.Message {
	active := log.Level >= logf.ErrorLevel
	msg := &stdMessage{active, logf.ErrorLevel, &strings.Builder{}}
	return msg
}

func (log *StdLogger) Warn() logf.Message {
	active := log.Level >= logf.WarnLevel
	msg := &stdMessage{active, logf.WarnLevel, &strings.Builder{}}
	return msg
}

func (log *StdLogger) Notice() logf.Message {
	active := log.Level >= logf.NoticeLevel
	msg := &stdMessage{active, logf.InfoLevel, &strings.Builder{}}
	return msg
}

func (log *StdLogger) Info() logf.Message {
	active := log.Level >= logf.InfoLevel
	msg := &stdMessage{active, logf.InfoLevel, &strings.Builder{}}
	return msg
}

func (log *StdLogger) Debug() logf.Message {
	active := log.Level >= logf.DebugLevel
	msg := &stdMessage{active, logf.DebugLevel, &strings.Builder{}}
	return msg
}

type stdMessage struct {
	active  bool
	level   logf.Level
	builder *strings.Builder
}

func (msg *stdMessage) Err(err error) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" error: %v", err))
	}
	return msg
}

func (msg *stdMessage) Bool(key string, val bool) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %v", key, val))
	}
	return msg
}

func (msg *stdMessage) Int(key string, val int) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %d", key, val))
	}
	return msg
}

func (msg *stdMessage) Int64(key string, val int64) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %d", key, val))
	}
	return msg
}

func (msg *stdMessage) Uint64(key string, val uint64) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %d", key, val))
	}
	return msg
}

func (msg *stdMessage) Str(key string, val string) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %s", key, val))
	}
	return msg
}

func (msg *stdMessage) Time(key string, val time.Time) logf.Message {
	if msg.active {
		msg.builder.WriteString(fmt.Sprintf(" %s: %v", key, val))
	}
	return msg
}

func (msg *stdMessage) Msg(val string) {
	if msg.active {
		lvn, err := translateLevel(msg.level)
		if err != nil {
			// This should never happen because the Logger constructor corrects
			// invalid level values.
			log.Print(errorLevel, "log configuration error", err)
		}

		msg.builder.WriteString(" ")
		msg.builder.WriteString(val)
		log.Print(lvn, msg.builder.String())
		// Once this method is called, deactivate for all future calls
		msg.active = false
	}
}

func (msg *stdMessage) Msgf(format string, a ...interface{}) {
	msg.Msg(fmt.Sprintf(format, a...))
}
