package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows"
)

var steamDLL *windows.DLL
var errNoConnectString = errors.New("connection string is empty")
var steamInitialised bool

var joinLink string
var joinLinkProcess = false

func spawnJoinLinkProcess() {
	cmd := exec.Command(exe, "-joinlink", strconv.Itoa(os.Getpid()))
	stdout, err := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		logErrorIfNoErrorSpam(errorOriginSteam, "Failed to start JoinLink process: "+err.Error())
		return
	}
	joinLinkProcess = true
	go func() {
		outScanner := bufio.NewScanner(stdout)
		for outScanner.Scan() {
			s := outScanner.Text()
			if !strings.HasPrefix(s, "steam://rungame/1085660/") && s != "" {
				logErrorIfNoErrorSpam(errorOriginSteam, "Unknown (error?) output from JoinLink process: "+s)
			} else {
				joinLink = s
			}
		}
		err = outScanner.Err()
		if err != nil {
			logErrorIfNoErrorSpam(errorOriginSteam, "Failed to read from JoinLink stdout: "+err.Error())
		}
		err = cmd.Process.Kill()
		if err != nil {
			logErrorIfNoErrorSpam(errorOriginSteam, "Failed to kill JoinLink process: "+err.Error())
		}
		err = cmd.Wait()
		if err != nil {
			logErrorIfNoErrorSpam(errorOriginSteam, "JoinLink process exited with error: "+err.Error())
		}
		joinLinkProcess = false
		joinLink = ""
	}()
}

func startJoinLinkOutput() {
	ticker := time.NewTicker(10 * time.Second)
	parentPID, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Println("Could not convert PID argument to int" + err.Error())
		os.Exit(1)
		return
	}
	for {
		select {
		case <-ticker.C:
			exeFound := false
			parentFound := false
			pl, _ := ps.Processes()
			for _, p := range pl {
				if p.Pid() == parentPID {
					parentFound = true
				}
				if p.Executable() == "destiny2.exe" {
					exeFound = true

					l, err := getJoinLink()
					if err != nil && !errors.Is(err, errNoConnectString) {
						fmt.Println(err)
						os.Exit(1)
						return
					}
					if l != "" {
						fmt.Println(l)
					}
				}
			}
			if !exeFound || !parentFound {
				os.Exit(0)
			}
		}
	}
}

func getJoinLink() (string, error) {
	if !steamInitialised {
		if _, err := os.Stat(makePath("steam_api64.dll")); os.IsNotExist(err) {
			err = copyEmbeddedDLL()
			if err != nil {
				return "", fmt.Errorf("Error copying files: %s", err)
			}
		}

		if _, err := os.Stat(makePath("steam_appid.txt")); os.IsNotExist(err) {
			err = copyEmbeddedDLL()
			if err != nil {
				return "", fmt.Errorf("Error copying files: %s", err)
			}
		}

		var err error
		steamDLL, err = windows.LoadDLL(makePath("steam_api64.dll"))
		if err != nil {
			return "", fmt.Errorf("Error loading steam_api64.dll: %s", err)
		}

		init, err := callProc("SteamAPI_Init")
		if err != nil {
			return "", fmt.Errorf("Error initialising steamapi: %s", err)
		}
		if init == 0 {
			return "", errors.New("Failed to initialise steamapi")
		}
		steamInitialised = true
	}

	steamUser, err := callProc("SteamAPI_SteamUser_v023")
	if err != nil {
		return "", fmt.Errorf("Failed to call SteamUser proc: %s", err)
	}

	steamID, err := callProc("SteamAPI_ISteamUser_GetSteamID", uintptr(steamUser))
	if err != nil {
		return "", fmt.Errorf("Failed to call GetSteamID proc: %s", err)
	}

	steamFriends, err := callProc("SteamAPI_SteamFriends_v017")
	if err != nil {
		return "", fmt.Errorf("Failed to call SteamFriends proc: %s", err)
	}

	arg := "connect"
	cstring := unsafe.Pointer(C.CString(arg))

	r, err := callProc("SteamAPI_ISteamFriends_GetFriendRichPresence", uintptr(steamFriends), steamID, uintptr(cstring))
	if err != nil {
		return "", fmt.Errorf("Failed to call GetFriendRichPresence proc: %s", err)
	}

	C.free(cstring)
	connect := C.GoString((*C.char)(unsafe.Pointer(r)))
	if connect == "" {
		return "", errNoConnectString
	}

	return fmt.Sprintf("steam://rungame/1085660/%d/%s", steamID, url.PathEscape(connect)), nil
}

func callProc(name string, args ...uintptr) (uintptr, error) {
	proc, err := steamDLL.FindProc(name)
	if err != nil {
		return 0, fmt.Errorf("error getting proc: %s", err)
	}

	r, _, err := proc.Call(args...)

	errno, ok := err.(windows.Errno)
	if !ok {
		return r, err
	}
	if errno != 0 {
		return r, err
	}

	return r, nil
}
