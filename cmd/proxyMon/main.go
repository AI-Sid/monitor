package main

import (
	"AI-Sid/monitor/internal/tools"
	"flag"
	"fmt"
)

import _ "AI-Sid/monitor/cmd/proxyMon/resources"

// used by LDFLAGS on build
var Build = "false"

var startFlag, stopFlag, quitFlag bool

func usage() {
	flag.PrintDefaults()
}

func init() {
	if !tools.SetBuildMode(Build) {
		tools.SetResourceModule("proxyMon")
	}
	flag.Usage = usage
	flag.BoolVar(&startFlag, "start", false, "Option for start Proxy Settings monitoring")
	flag.BoolVar(&stopFlag, "stop", false, "Option for stop Proxy Settings monitoring")
	flag.BoolVar(&quitFlag, "quit", false, "Option for quit Proxy Settings monitor")
	flag.Parse()
}

const welcome = "Proxy Settings Monitor v.1.0"

func main() {
	fmt.Printf("Build mode: %v\n", Build)
	action := tools.ACTION_NONE
	if quitFlag {
		action = tools.ACTION_QUIT
	} else if stopFlag {
		action = tools.ACTION_STOP
	} else if startFlag {
		action = tools.ACTION_START
	}
	if tools.InitializeControl(action) { // primary instance
		if action != tools.ACTION_QUIT {
			tools.RunTray()
		} else {
            fmt.Printf("%v\nInstance closed by -quit flag is set.\n", welcome)
        }
	} else if tools.IsNormalState() { // secondary instance
        fmt.Printf("%v (secondary)\n", welcome)
        if action == tools.ACTION_NONE {
            fmt.Println("Can't change main instance state because no flags is set")
        } else {
            if tools.SendAction(action) {
                fmt.Printf("Action %v sent successfully", tools.ActionsDisplay[action])
            } else {
                fmt.Printf("Action %v sent with error", tools.ActionsDisplay[action])
            }
        }
	}
	tools.DoExitProgram()
}
