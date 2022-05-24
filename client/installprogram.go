package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows"
)

// My epic hacked-together install prompt that totally works 100% of the time 99% of the time.
func installProgram() {
	defer func() {
		fmt.Println("\n Press ENTER to close this window.")
		fmt.Scanln()
	}()

	fmt.Print("         _      _              _           _   _\n        (_)    | |            | |         | | (_)\n    _ __ _  ___| |__ ______ __| | ___  ___| |_ _ _ __  _   _\n   | '__| |/ __| '_ \\______/ _` |/ _ \\/ __| __| | '_ \\| | | |\n   | |  | | (__| | | |    | (_| |  __/\\__ \\ |_| | | | | |_| |\n   |_|  |_|\\___|_| |_|     \\__,_|\\___||___/\\__|_|_| |_|\\__, |\n                                                        __/ |\n                                                       |___/    ",
		version, "\n\n\n")
	log.SetFlags(log.Lshortfile)

	var isAlreadyRunning bool
	pl, _ := ps.Processes()
	for _, p := range pl {
		if p.Executable() == "rich-destiny.exe" && p.Pid() != os.Getpid() {
			isAlreadyRunning = true
			break
		}
	}

	startupShortcutPath, err := getStartupShortcutPath()
	if err != nil {
		log.Printf("Couldn't get startup folder path: %s", err)
		return
	}

	if isAlreadyRunning {
		fmt.Println(" rich-destiny is already running. Open the control panel at:  https://richdestiny.app/cp")
		return
	} else if _, err = os.Stat(startupShortcutPath); err == nil {
		fmt.Println(" rich-destiny is already installed but not running. Attempting to start it...")

		shortcut, err := getShortcutIDispatch(startupShortcutPath)
		if err != nil {
			log.Printf("Couldn't get shortcut at path %s despite the file existing: %s", startupShortcutPath, err)
			return
		}

		variant, err := shortcut.GetProperty("TargetPath")
		if err != nil {
			log.Printf("Couldn't get TargetPath property of shortcut %s: %s", startupShortcutPath, err)
			return
		}
		targetPath := variant.ToString()
		shortcut.Release()

		if _, err = os.Stat(targetPath); err == nil {
			started, err := successfullyStartDaemon(targetPath)
			if err != nil {
				log.Printf("Error trying to start daemon at path %s: %s", targetPath, err)
			} else if started {
				fmt.Println(" rich-destiny was successfully started! Find the control panel at:  https://richdestiny.app/cp")
				return
			} else { // it didn't start, but that message was already printed so just return here
				return
			}
		} else {
			if os.IsNotExist(err) {
				// Second argument adds a fresh empty new line :)
				fmt.Println(" A rich-destiny startup shortcut exists but the target file was not found. You will now be guided through the setup again.", "")
			} else {
				log.Printf("Error checking if file %s exists: %s", targetPath, err)
				return
			}
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Error checking if file %s exists: %s", startupShortcutPath, err)
		return
	}

	fmt.Println(" Welcome to the rich-destiny setup!")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not get home directory...: %s", err)
		return
	}

	var downloadDir bool
	if currentDirectory == filepath.Join(home, "Downloads") {
		downloadDir = true
		fmt.Println("\n This program will refuse to install in the Downloads folder as it often will not work from there.")
	} else {
		fmt.Println("\n You can install the program in either the default (recommended) folder or the current folder.")
	}

	defaultDirectory := filepath.Join(home, "rich-destiny")
	fmt.Printf(" If you want to install the program in another folder: Exit, move this file to the desired folder and then run it again.\n\n → Default: %s", defaultDirectory)
	if !downloadDir {
		fmt.Printf("\n → Current: %s", currentDirectory)
	}

	for {
		fmt.Print("\n\n  > Choose a location: [Default/")
		if !downloadDir {
			fmt.Print("Current/")
		}
		fmt.Print("Exit]: ")

		r, err := readUserInput()
		if err != nil {
			log.Printf(" Unable to read your input...: %s", err)
			return
		}
		if strings.Contains(r, "d") {
			fmt.Println(" Okay, attempting to move there...")

			err = os.Mkdir(defaultDirectory, os.ModePerm)
			if err != nil && !errors.Is(err, os.ErrExist) {
				log.Printf("Error trying to create %s\\rich-destiny folder: %s", home, err)
				return
			}

			oldExe := exe
			exe = filepath.Join(defaultDirectory, "rich-destiny.exe")

			if oldExe[0] != exe[0] {
				newFile, err := os.Create(exe)
				if err != nil {
					log.Printf("Error creating file at new location: %s", err)
					return
				}
				oldFile, err := os.Open(oldExe)
				if err != nil {
					log.Printf("Error opening old file: %s", err)
					return
				}
				_, err = io.Copy(newFile, oldFile)
				if err != nil {
					log.Printf("Error copying file: %s", err)
					return
				}
				newFile.Close()
				oldFile.Close()
			} else {
				err = os.Rename(oldExe, exe)
				if err != nil {
					log.Printf("Error moving rich-destiny.exe to new location: %s", err)
					return
				}
			}

			fmt.Println(" Successfully moved.")
			break
		} else if strings.Contains(r, "c") {
			if downloadDir {
				fmt.Println(" Okay, move this program to a different directory manually and run it from there.")
				return
			}
			fmt.Println(" Okay, installing at the current location.")
			break
		} else if strings.Contains(r, "x") {
			fmt.Println(" Okay, exiting. If you intend to move this program to another folder, make sure to close this window first.")
			return
		} else {
			fmt.Println(" Invalid response. Please reply with Default, Current or Exit.")
		}
	}

	makeShortcut(startupShortcutPath)

	fmt.Println(" Done! Waiting for rich-destiny to start...")

	started, err := successfullyStartDaemon(exe)
	if err != nil {
		log.Printf("Error trying to start rich-destiny: %s", err)
		return
	} else if !started {
		return
	}

	fmt.Println(" Done! Opening a browser tab to log in with Bungie.net. Setup is now complete and you can close this window.\n\n   ♥ Thanks for using rich-destiny! ♥")
	openOauthTab()
}

func getStartupShortcutPath() (string, error) {
	startupFolderPath, err := windows.KnownFolderPath(windows.FOLDERID_Startup, 0)
	if err != nil {
		return "", err
	}
	return filepath.Join(startupFolderPath, "rich-destiny.lnk"), nil
}

// https://stackoverflow.com/a/41886180/10530600
func getShortcutIDispatch(path string) (*ole.IDispatch, error) {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return nil, err
	}
	defer oleShellObject.Release()
	wshShell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}
	defer wshShell.Release()
	shortcut, err := oleutil.CallMethod(wshShell, "CreateShortcut", path)
	if err != nil {
		return nil, err
	}
	return shortcut.ToIDispatch(), nil
}

// successfullyStartDaemon starts the rich-destiny exe in daemon mode.
// path should be the path of the executable. This function adds -daemon as an argument.
func successfullyStartDaemon(path string) (success bool, err error) {
	var cmd *exec.Cmd

	for i := 0; i <= 10; i++ {
		if i == 0 || i == 5 {
			// First kill the process that was started at i == 0 in case it exists
			if i == 5 {
				if cmd.Process != nil {
					err = cmd.Process.Kill()
					if err != nil {
						return
					}
				}
			}

			cmd = exec.Command(path, "-daemon")
			err = cmd.Start()
			if err != nil {
				return
			}

			if cmd.Process == nil {
				return false, errors.New("Failed to start rich-destiny via shortcut")
			}
		}
		_, err = http.Get("http://localhost:35893")
		if err != nil {
			time.Sleep(3 * time.Second)
		} else {
			success = true
			break
		}
	}

	if !success {
		fmt.Println(" It seems rich-destiny didn't want to start at all..." +
			"Try seeing if there is any information in the logs folder where rich-destiny was installed or head to the support server for help ( https://discord.gg/UNU4UXp ).")
		return
	}

	err = cmd.Process.Release()
	if err != nil {
		return
	}

	return
}

func makeShortcut(path string) error {
	shortcut, err := getShortcutIDispatch(path)
	if err != nil {
		return err
	}

	// Load             Method
	// Save             Method
	// Arguments        Property   string
	// Description      Property   string
	// FullName         Property   string
	// Hotkey           Property   string
	// IconLocation     Property   string
	// RelativePath     Property   string
	// TargetPath       Property   string
	// WindowStyle      Property   int    7=minimised, 3 or 0(?)=maximised, 4=normal
	// WorkingDirectory Property
	properties := map[string]interface{}{
		"Arguments":   "-daemon",
		"TargetPath":  exe,
		"WindowStyle": 7,
	}
	for p, v := range properties {
		_, err = shortcut.PutProperty(p, v)
		if err != nil {
			return fmt.Errorf("Failed putting property %s with value %v on shortcut: %s", p, v, err)
		}
	}

	_, err = shortcut.CallMethod("Save")
	if err != nil {
		return fmt.Errorf("Failed to call Save method on shortcut: %s", err)
	}

	shortcut.Release()
	return nil
}

func readUserInput() (string, error) {
	var r string
	_, err := fmt.Scanln(&r)
	// trigger invalid input instead of an error
	if err != nil && err.Error() == "unexpected newline" {
		err = nil
	}
	return strings.ToLower(r), err
}
