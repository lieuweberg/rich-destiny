package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/service"
)

//go:embed steam_api64.dll
var steamAPIFile []byte

// My epic hacked-together install prompt that totally works 100% of the time 99% of the time.
func installProgram() {
	defer func() {
		fmt.Println("\n Press ENTER to close this window.")
		fmt.Scanln()
	}()

	fmt.Print("         _      _              _           _   _\n        (_)    | |            | |         | | (_)\n    _ __ _  ___| |__ ______ __| | ___  ___| |_ _ _ __  _   _\n   | '__| |/ __| '_ \\______/ _` |/ _ \\/ __| __| | '_ \\| | | |\n   | |  | | (__| | | |    | (_| |  __/\\__ \\ |_| | | | | |_| |\n   |_|  |_|\\___|_| |_|     \\__,_|\\___||___/\\__|_|_| |_|\\__, |\n                                                        __/ |\n                                                       |___/    ",
		version, "\n\n\n")
	log.SetFlags(log.Lshortfile)

	// This works because it deletes/starts/detects by name. Path does not matter, except when installing.
	createService()
	status, err := s.Status()
	if err != nil && !errors.Is(err, service.ErrNotInstalled) {
		log.Printf("Error trying to detect service status: %s", err)
	}

	if status == service.StatusRunning {
		fmt.Println(" rich-destiny is already installed and running. Open the control panel at:  https://richdestiny.app/cp")
		return
	} else if status == service.StatusStopped {
		fmt.Println(" rich-destiny is already installed but not running. Attempting to start it...")
		started, err := successfullyStartService()
		if err != nil {
			if err.Error() != "The system cannot find the file specified." {
				fmt.Printf(" An error occured, but rich-destiny doesn't recognise it: %s\n\n"+
					"Is this similar or does this translate to \"The system cannot find the file specified.\"?", err)
				for {
					fmt.Print("\n Choose: [Yes/No]: ")
					r, err := readUserInput()
					if err != nil {
						log.Printf(" Unable to read your input...: %s", err)
						return
					}
					if strings.Contains(r, "y") {
						break
					} else if strings.Contains(r, "n") {
						fmt.Println(" Okay. If you need help, please join the support server!  https://discord.gg/UNU4UXp")
						return
					} else {
						fmt.Println(" Invalid response.")
					}
				}
			}
			fmt.Println(" Windows can't find the file where you installed rich-destiny previously.\n\n Do you want to uninstall the original location so you can reinstall here?")
			for {
				fmt.Print("\n Choose: [Yes/No]: ")
				r, err := readUserInput()
				if err != nil {
					log.Printf(" Unable to read your input...: %s", err)
					return
				}
				if strings.Contains(r, "y") {
					err = s.Uninstall()
					if err != nil {
						log.Printf("Error trying to uninstall the service: %s", err)
					}
					fmt.Print("\n Uninstalled. Starting new installation now.\n\n")
					break
				} else if strings.Contains(r, "n") {
					fmt.Println(" Okay. If you need help, please join the support server!  https://discord.gg/UNU4UXp")
					return
				} else {
					fmt.Println(" Invalid response.")
				}
			}
		} else if started {
			fmt.Println(" rich-destiny was successfully started! Find the control panel at:  https://richdestiny.app/cp")
			return
		} else { // it didn't start, but that message was already printed so just return here
			return
		}
	}

	fmt.Println(" Welcome to the rich-destiny setup!")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not get home directory...: %s", err)
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

			currentDirectory = filepath.Base(exe)
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

	createService()
	err = s.Install()
	if err != nil {
		log.Printf("Error adding rich-destiny to the service manager: %s", err)
		return
	}

	fmt.Println(" Done! Copying Steam API files...")

	err = os.WriteFile(makePath("steam_api64.dll"), steamAPIFile, os.ModePerm)
	if err != nil {
		log.Printf("Error copying steam_api64.dll from embedded file: %s", err)
		return
	}
	err = os.WriteFile(makePath("steam_appid.txt"), []byte("1085660"), os.ModePerm)
	if err != nil {
		log.Printf("Error writing steam_appid.txt file: %s", err)
		return
	}

	fmt.Println(" Done! Waiting for rich-destiny to start...")

	started, err := successfullyStartService()
	if err != nil {
		log.Printf("Error trying to start rich-destiny: %s", err)
		return
	} else if !started {
		return
	}

	fmt.Println(" Done! Opening a browser tab to log in with Bungie.net. Setup is now complete and you can close this window.")
	openOauthTab()
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
