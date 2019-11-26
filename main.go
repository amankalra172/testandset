package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func main() {
	lockCommand := flag.NewFlagSet("lock, l", flag.ExitOnError)
	getCommand := flag.NewFlagSet("get, g", flag.ExitOnError)
	refreshCommand := flag.NewFlagSet("refresh, r", flag.ExitOnError)
	unlockCommand := flag.NewFlagSet("unlock, u", flag.ExitOnError)

	lockNamePtr := lockCommand.String("name", "", "Name of the mutex")
	lockOutputPtr := lockCommand.String("output", "json", "Formats the output {json|token}")

	getNamePtr := getCommand.String("name", "", "Name of the mutex")

	refreshNamePtr := refreshCommand.String("name", "", "Name of the mutex")
	refreshTokenPtr := refreshCommand.String("token", "", "Token for manipulating an existing mutex")

	unlockNamePtr := unlockCommand.String("name", "", "Name of the mutex")
	unlockTokenPtr := unlockCommand.String("token", "", "Token for manipulating an existing mutex")

	if len(os.Args) < 3 || os.Args[1] != "mutex" {
		exitWithMessage("Wrong arguments")
	}

	switch os.Args[2] {
	case "lock", "l":
		lockCommand.Parse(os.Args[3:])
	case "get", "g":
		getCommand.Parse(os.Args[3:])
	case "refresh", "r":
		refreshCommand.Parse(os.Args[3:])
	case "unlock", "u":
		unlockCommand.Parse(os.Args[3:])
	default:
		fmt.Println("mutex lock")
		lockCommand.PrintDefaults()
		fmt.Println("\nmutex get")
		getCommand.PrintDefaults()
		fmt.Println("\nmutex refresh")
		refreshCommand.PrintDefaults()
		fmt.Println("\nmutex unlock")
		unlockCommand.PrintDefaults()
		os.Exit(1)
	}

	if lockCommand.Parsed() {
		response, err := http.Get("http://localhost:3002/v1/mutex/" + *lockNamePtr + "/lock")
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		}

		data, _ := ioutil.ReadAll(response.Body)

		if response.StatusCode != 200 {
			exitWithMessage("Could not lock mutex!")
		}

		switch *lockOutputPtr {
		case "json":
			fmt.Println(string(data))
		case "token":
			var answer LockAnswer
			err = json.Unmarshal([]byte(data), &answer)
			if err != nil || answer.Token == "" {
				exitWithMessage("Could not lock mutex!")
			}
			fmt.Println(answer.Token)
		default:
			fmt.Println(string(data))
		}
	}

	if getCommand.Parsed() {
		response, err := http.Get("http://localhost:3002/v1/mutex/" + *getNamePtr)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(data))
		}
	}

	if refreshCommand.Parsed() {
		response, err := http.Get("http://localhost:3002/v1/mutex/" + *refreshNamePtr + "/refresh/" + *refreshTokenPtr)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(data))
		}
	}

	if unlockCommand.Parsed() {
		response, err := http.Get("http://localhost:3002/v1/mutex/" + *unlockNamePtr + "/unlock/" + *unlockTokenPtr)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(data))
		}
	}
}
