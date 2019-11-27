package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type LockAnswer struct {
	Token     string
	ExpiresAt time.Time
}

func exitWithMessage(message string) {
	fmt.Println(message)
	os.Exit(1)
}

var LockCommand *flag.FlagSet
var GetCommand *flag.FlagSet
var RefreshCommand *flag.FlagSet
var UnlockCommand *flag.FlagSet
var AutoRefreshCommand *flag.FlagSet

var LockName string
var LockOutput string
var LockTimeout int

var GetName string

var RefreshName string
var RefreshToken string

var UnlockName string
var UnlockToken string

var AutoRefreshName string
var AutoRefreshToken string

func createFlagSets() {
	LockCommand = flag.NewFlagSet("lock, l", flag.ExitOnError)
	GetCommand = flag.NewFlagSet("get, g", flag.ExitOnError)
	RefreshCommand = flag.NewFlagSet("refresh, r", flag.ExitOnError)
	UnlockCommand = flag.NewFlagSet("unlock, u", flag.ExitOnError)
	AutoRefreshCommand = flag.NewFlagSet("auto-refresh, a", flag.ExitOnError)
}

func setCommands() {
	setLockCommands()
	setGetCommands()
	setRefreshCommands()
	setUnlockCommands()
	setAutoRefreshCommands()
}

func setLockCommands() {
	LockCommand.StringVar(&LockName, "name", "", "Name of the mutex")
	LockCommand.StringVar(&LockName, "n", "", "Name of the mutex (shorthand)")
	LockCommand.StringVar(&LockOutput, "output", "json", "Formats the output {json|token}")
	LockCommand.StringVar(&LockOutput, "o", "json", "Formats the output {json|token} (shorthand)")
	LockCommand.IntVar(&LockTimeout, "timeout", 0, "Time in seconds with automatically trying to lock a mutex, when it is already lock by someone else")
	LockCommand.IntVar(&LockTimeout, "t", 0, "Time in seconds with automatically trying to lock a mutex, when it is already lock by someone else (shorthand)")
}

func setGetCommands() {
	GetCommand.StringVar(&GetName, "name", "", "Name of the mutex")
	GetCommand.StringVar(&GetName, "n", "", "Name of the mutex (shorthand)")
}

func setRefreshCommands() {
	RefreshCommand.StringVar(&RefreshName, "name", "", "Name of the mutex")
	RefreshCommand.StringVar(&RefreshName, "n", "", "Name of the mutex (shorthand)")
	RefreshCommand.StringVar(&RefreshToken, "token", "", "Token for manipulating an existing mutex")
	RefreshCommand.StringVar(&RefreshToken, "t", "", "Token for manipulating an existing mutex (shorthand)")
}

func setUnlockCommands() {
	UnlockCommand.StringVar(&UnlockName, "name", "", "Name of the mutex")
	UnlockCommand.StringVar(&UnlockName, "n", "", "Name of the mutex (shorthand)")
	UnlockCommand.StringVar(&UnlockToken, "token", "", "Token for manipulating an existing mutex")
	UnlockCommand.StringVar(&UnlockToken, "t", "", "Token for manipulating an existing mutex (shorthand)")
}

func setAutoRefreshCommands() {
	AutoRefreshCommand.StringVar(&AutoRefreshName, "name", "", "Name of the mutex")
	AutoRefreshCommand.StringVar(&AutoRefreshName, "n", "", "Name of the mutex (shorthand)")
	AutoRefreshCommand.StringVar(&AutoRefreshToken, "token", "", "Token for manipulating an existing mutex")
	AutoRefreshCommand.StringVar(&AutoRefreshToken, "t", "", "Token for manipulating an existing mutex (shorthand)")
}

func parseArguments() {
	if len(os.Args) < 3 || os.Args[1] != "mutex" {
		exitWithMessage("Wrong arguments")
	}

	switch os.Args[2] {
	case "lock", "l":
		LockCommand.Parse(os.Args[3:])
	case "get", "g":
		GetCommand.Parse(os.Args[3:])
	case "refresh", "r":
		RefreshCommand.Parse(os.Args[3:])
	case "unlock", "u":
		UnlockCommand.Parse(os.Args[3:])
	case "auto-refresh", "a":
		AutoRefreshCommand.Parse(os.Args[3:])
	default:
		fmt.Println("mutex lock")
		LockCommand.PrintDefaults()
		fmt.Println("\nmutex get")
		GetCommand.PrintDefaults()
		fmt.Println("\nmutex refresh")
		RefreshCommand.PrintDefaults()
		fmt.Println("\nmutex unlock")
		UnlockCommand.PrintDefaults()
		fmt.Println("\nmutex auto-refresh")
		AutoRefreshCommand.PrintDefaults()
		os.Exit(1)
	}
}

func tryLockViaPolling(tryUntil time.Time) []byte {
	pollTime := 5

	if LockTimeout < 5 {
		pollTime = LockTimeout
	}

	for {
		time.Sleep(time.Duration(pollTime) * time.Second)
		if time.Now().After(tryUntil) {
			exitWithMessage("Timeout ellapsed. Could not lock mutex!")
		}
		response, err := http.Get("http://localhost:3002/v1/mutex/" + LockName + "/lock")
		if err != nil {
			exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s\n", err))
		}
		data, _ := ioutil.ReadAll(response.Body)

		if response.StatusCode == 200 {
			return data
		}
	}
}

func setLockOutput(data []byte) {
	switch LockOutput {
	case "json":
		fmt.Println(string(data))
	case "token":
		var answer LockAnswer
		err := json.Unmarshal([]byte(data), &answer)
		if err != nil || answer.Token == "" {
			exitWithMessage("Could not lock mutex!")
		}
		fmt.Println(answer.Token)
	default:
		fmt.Println(string(data))
	}
}

func handleLockCommand() {
	response, err := http.Get("http://localhost:3002/v1/mutex/" + LockName + "/lock")
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s\n", err))
	}

	tryUntil := time.Now().Add(time.Duration(LockTimeout) * time.Second)
	data, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode != 200 {
		if LockTimeout > 0 {
			data = tryLockViaPolling(tryUntil)
		} else {
			exitWithMessage("Could not lock mutex!")
		}
	}

	setLockOutput(data)
}

func handleGetCommand() {
	response, err := http.Get("http://localhost:3002/v1/mutex/" + GetName)
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s\n", err))
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

func handleRefreshCommand() {
	response, err := http.Get("http://localhost:3002/v1/mutex/" + RefreshName + "/refresh/" + RefreshToken)
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s\n", err))
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

func handleUnlockCommand() {
	response, err := http.Get("http://localhost:3002/v1/mutex/" + UnlockName + "/unlock/" + UnlockToken)
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s\n", err))
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

func unlockWhenInterrupted(c chan os.Signal) {
	<-c
	response, err := http.Get("http://localhost:3002/v1/mutex/" + AutoRefreshName + "/unlock/" + AutoRefreshToken)
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s", err))
	}

	data, _ := ioutil.ReadAll(response.Body)
	exitWithMessage(string(data))
}

func tryAutoRefresh() {
	time.Sleep(5 * time.Second)

	response, err := http.Get("http://localhost:3002/v1/mutex/" + AutoRefreshName + "/refresh/" + AutoRefreshToken)
	if err != nil {
		exitWithMessage(fmt.Sprintf("The HTTP request failed with error %s", err))
	}

	if response.StatusCode != 200 {
		exitWithMessage("Could not refresh anymore")
	}
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data))
}

func handleAutoRefreshCommand() {
	//unlock when user aborts autorefresh
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go unlockWhenInterrupted(c)

	for {
		tryAutoRefresh()
	}
}

func main() {
	createFlagSets()
	setCommands()
	parseArguments()

	if LockCommand.Parsed() {
		handleLockCommand()
	}

	if GetCommand.Parsed() {
		handleGetCommand()
	}

	if RefreshCommand.Parsed() {
		handleRefreshCommand()
	}

	if UnlockCommand.Parsed() {
		handleUnlockCommand()
	}

	if AutoRefreshCommand.Parsed() {
		handleAutoRefreshCommand()
	}
}
