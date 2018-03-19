# proci - Go Lang Process Information Library

[![Documentation](https://godoc.org/github.com/midstar/proci?status.svg)](https://godoc.org/github.com/midstar/proci)
[![Go Report Card](https://goreportcard.com/badge/github.com/midstar/proci)](https://goreportcard.com/report/github.com/midstar/proci)
[![AppVeyor](https://ci.appveyor.com/api/projects/status/github/midstar/proci?svg=true)](https://ci.appveyor.com/api/projects/status/github/midstar/proci)
[![Coverage Status](https://coveralls.io/repos/github/midstar/proci/badge.svg?branch=master)](https://coveralls.io/github/midstar/proci?branch=master)

**proci** is a Go library that provides a set of functions for listing processes and get following information for each process:

* The process path and name
* The process command line (that was executed when the process was started)
* The process RAM memory usage 

Additionaly this library has functions for get:

* The total physical RAM memory installed on the computer
* The available physical RAM memory on the computer

Supported platforms:

* Windows 64 bit

More platforms might be added in future.

To see the full list of **proci** functions, check out the [documentation on godoc.org](https://godoc.org/github.com/midstar/proci)

## Install

```bash
go get github.com/midstar/proci
```

## Import

```go
import (
	"github.com/midstar/proci"
)
```

## Example Usage

```go
package main

import (
	"fmt"
	"github.com/midstar/proci"
)

func main() {
  pids := proci.GetProcessPids()
  for i:=0 ; i < len(pids) ; i++ {
    pid := pids[i]
    if pid == 0 {
      // This is the idle process. No operations can be performed on it.
      continue
    }
    path, patherr := proci.GetProcessPath(pid)
    if patherr != nil {
      fmt.Println("  GetProcessPath for PID", pid, "returned error:", patherr)
    }
    fmt.Println("  Path:", path)
    commandLine, cmderr := proci.GetProcessCommandLine(pid)
    if cmderr != nil {
      // Expected for some Windows system processes.
      fmt.Println("  Unable to read command line for PID", pid," error: ", cmderr)
    } else {
      fmt.Println("  Command line:", commandLine)
    }
    memoryUsage, memerr := proci.GetProcessMemoryUsage(pid)
    if memerr != nil {
      fmt.Println("  GetProcessMemoryUsage for PID", pid, "returned error:", memerr)
    }
    fmt.Println("  Memory usage:", memoryUsage, "B (", memoryUsage / 1024 / 1024, "MB )")
  }
}
```

## Author and license

This library is written by Joel Midstjärna and is licensed under the MIT License.