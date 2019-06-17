// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package log

import (
	"fmt"
	"os"

	logging "github.com/op/go-logging"
)

// Message logs a message in a fixed format so it can be analyzed by log handlers
// e.g. "METHOD CODE descriptive reason"
func Message(method, code, reason string) {
	msg := fmt.Sprintf("%s: method=%s, code=%s", reason, method, code)
	Error(msg)
}

var l = logging.MustGetLogger("serialvault")

// InitLogger initializes logger for backend with the specified level
// format = '%(asctime)s.%(msecs)03dZ %(levelname)s %(name)s "%(message)s"'
// datefmt = "%Y-%m-%d %H:%M:%S"
// 2016-07-14 01:02:03.456Z INFO app "hello"
func InitLogger(level logging.Level) {
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05Z} %{level} %{module}%{color:reset} %{message:q}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(level, "")
	logging.SetBackend(backendLeveled)
}

// Fatalf calls logger in fatal level with format
func Fatalf(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

// Fatal calls logger in fatal level
func Fatal(args ...interface{}) {
	l.Fatal(args...)
}

// Errorf calls logger in eror level with format
func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

// Error calls logger in error level
func Error(args ...interface{}) {
	l.Error(args...)
}

// Warningf calls logger in warning level with format
func Warningf(format string, args ...interface{}) {
	l.Warningf(format, args...)
}

// Warning calls logger in warning level
func Warning(args ...interface{}) {
	l.Warning(args...)
}

// Infof calls logger in info level with format
func Infof(format string, args ...interface{}) {
	l.Infof(format, args...)
}

// Info calls logger in info level
func Info(args ...interface{}) {
	l.Info(args...)
}

// Debugf calls logger in debug level with format
func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

// Debug calls logger in debug level
func Debug(args ...interface{}) {
	l.Debug(args...)
}

// Printf calls logger in info level with format
func Printf(format string, args ...interface{}) {
	Errorf(format, args...)
}

// Println calls logger in info level
func Println(args ...interface{}) {
	Error(args...)
}
