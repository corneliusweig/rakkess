// +build windows

/*
Copyright 2020 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// NOTICE: This implementation comes from logrus, unfortunately logrus
// does not expose a public interface we can use to call it.
//   https://github.com/sirupsen/logrus/blob/master/terminal_check_notappengine.go
//   https://github.com/sirupsen/logrus/blob/master/terminal_windows.go

package printer

import (
	"io"
	"os"
	"syscall"

	sequences "github.com/konsorten/go-windows-terminal-sequences"
)

// initTerminal enables ANSI color escape on windows. Usually, this is done by logrus, but
// since we don't log anything before printing, we need to take care of this ourselves.
func initTerminal(w io.Writer) {
	if f, ok := w.(*os.File); ok {
		sequences.EnableVirtualTerminalProcessing(syscall.Handle(f.Fd()), true)
	}
}

func isTerminalImpl(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		var mode uint32
		err := syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode)
		return err == nil
	}
	return false
}
