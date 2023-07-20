package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"log"
	"net/url"
	"unsafe"

	"golang.org/x/sys/windows"
)

var dll *windows.DLL

func getJoinLink() string {
	dll = windows.MustLoadDLL("steam_api64.dll")

	init, err := callProc("SteamAPI_Init")
	if err != nil {
		log.Printf("Error initialising steamapi: %s", err)
		return ""
	}
	if init == 0 {
		log.Printf("Failed to initialise steamapi")
		return ""
	}

	steamUser, err := callProc("SteamAPI_SteamUser_v023")
	if err != nil {
		fmt.Printf("Failed to call SteamUser proc: %s", err)
		return ""
	}

	steamID, err := callProc("SteamAPI_ISteamUser_GetSteamID", uintptr(steamUser))
	if err != nil {
		log.Printf("Failed to call GetSteamID proc: %s", err)
		return ""
	}

	log.Printf("SteamID64: %d\n", steamID)

	steamFriends, err := callProc("SteamAPI_SteamFriends_v017")
	if err != nil {
		log.Printf("Failed to call SteamFriends proc: %s", err)
		return ""
	}

	arg := "connect"
	cstring := unsafe.Pointer(C.CString(arg))

	r, err := callProc("SteamAPI_ISteamFriends_GetFriendRichPresence", uintptr(steamFriends), steamID, uintptr(cstring))
	if err != nil {
		log.Printf("Failed to call GetFriendRichPresence proc: %s", err)
		return ""
	}

	C.free(cstring)
	connect := C.GoString((*C.char)(unsafe.Pointer(r)))

	log.Printf("Connect string: %v\n", connect)
	return fmt.Sprintf("steam://rungame/1085660/%d/%s", steamID, url.PathEscape(connect))
}

func callProc(name string, args ...uintptr) (uintptr, error) {
	proc, err := dll.FindProc(name)
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
