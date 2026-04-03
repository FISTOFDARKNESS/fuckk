package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "github.com/lib/pq"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/app/config"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/color"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/console"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/files"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/roblox"
)

const (
	DB_URL = "postgresql://neondb_owner:npg_ZIigvq76hfeD@ep-lucky-night-ahauzl87-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"
)

var (
	cookieFile = config.Get("cookie_file")
	port       = config.Get("port")
)

func main() {
	console.ClearScreen()

	validateKey()

	fmt.Println("Authenticating cookie...")

	cookie, readErr := files.Read(cookieFile)
	cookie = strings.TrimSpace(cookie)

	c, clientErr := roblox.NewClient(cookie)
	console.ClearScreen()

	if readErr != nil || clientErr != nil {
		if readErr != nil && !os.IsNotExist(readErr) {
			color.Error.Println(readErr)
		}

		if clientErr != nil && cookie != "" {
			color.Error.Println(clientErr)
		}

		getCookie(c)
	}

	if err := files.Write(cookieFile, c.Cookie); err != nil {
		color.Error.Println("Failed to save cookie: ", err)
	}

	fmt.Println("localhost started on port " + port + ". Waiting to start reuploading.")
	if err := serve(c); err != nil {
		logFatal(err)
	}
}

func getCookie(c *roblox.Client) {
	for {
		i, err := console.LongInput("ROBLOSECURITY: ")
		console.ClearScreen()
		if err != nil {
			color.Error.Println(err)
			continue
		}

		fmt.Println("Authenticating cookie...")
		err = c.SetCookie(i)
		console.ClearScreen()
		if err != nil {
			color.Error.Println(err)
			continue
		}

		files.Write(cookieFile, i)
		break
	}
}

func validateKey() {
	localHWID := getHWID()

	db, err := sql.Open("postgres", DB_URL)
	if err != nil {
		logFatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()

	// Try loading from config first
	savedKey := config.Get("license_key")
	if savedKey != "" {
		var remoteHWID string
		err = db.QueryRow("SELECT hwid FROM keys WHERE key_id = $1", savedKey).Scan(&remoteHWID)
		if err == nil {
			if remoteHWID == "ANY" || remoteHWID == "RESETED" || remoteHWID == "" || remoteHWID == localHWID {
				// Key is valid, proceed
				if remoteHWID == "RESETED" || remoteHWID == "" {
					db.Exec("UPDATE keys SET hwid = $1 WHERE key_id = $2", localHWID, savedKey)
				}
				fmt.Println("License Validated (Saved).")
				return
			}
		}
	}

	for {
		i, err := console.LongInput("License: ")
		console.ClearScreen()
		if err != nil {
			color.Error.Println(err)
			continue
		}

		var remoteHWID string
		err = db.QueryRow("SELECT hwid FROM keys WHERE key_id = $1", i).Scan(&remoteHWID)

		if err == sql.ErrNoRows {
			color.Error.Println("Invalid License. This key does not exist.")
			continue
		} else if err != nil {
			logFatal(fmt.Sprintf("Database error: %v", err))
		}

		// Public Key logic (Anyone can use)
		if remoteHWID == "ANY" {
			fmt.Println("Access Granted (Public Key).")
			config.Set("license_key", i)
			config.Save()
			break
		}

		// Autolock logic (First use locks the key)
		if remoteHWID == "RESETED" || remoteHWID == "" {
			fmt.Println("Key is reset. Autolocking to this PC...")
			_, err = db.Exec("UPDATE keys SET hwid = $1 WHERE key_id = $2", localHWID, i)
			if err != nil {
				logFatal(fmt.Sprintf("Failed to update key in database: %v", err))
			}
			fmt.Println("Access Granted (Key Locked to this PC).")
			config.Set("license_key", i)
			config.Save()
			break
		}

		// HWID Check (Already locked)
		if remoteHWID == localHWID {
			fmt.Println("Access Granted.")
			config.Set("license_key", i)
			config.Save()
			break
		} else {
			color.Error.Println(fmt.Sprintf("Access Denied. Key is locked to another PC.\nYour HWID: %s", localHWID))
		}
	}
}

func getHWID() string {
	cmd := exec.Command("powershell", "-Command", "(Get-WmiObject -Class Win32_ComputerSystemProduct).UUID")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logFatal(fmt.Sprintf("Failed to get HWID: %v", err))
	}
	return strings.TrimSpace(out.String())
}

func logFatal(a ...any) {
	color.Error.Println(a...)
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
	os.Exit(1)
}
