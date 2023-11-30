package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"net/url"
	"unsafe"

	"golang.org/x/sys/windows"
)

var steamDLL *windows.DLL
var errNoConnectString = errors.New("connection string is empty")
var steamInitialised bool

func getJoinLink() (string, error) {
	if !steamInitialised {
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
		return 0, fmt.Errorf("error getting init proc: %s", err)
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
