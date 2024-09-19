package main

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/pborman/options"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	Version string = "Genesis 0.5v compiled for GNU/Linux on ARM64"
)

var (
	CurrentBrightness int
	MaxBrightness     int
	NewBrightness     int

	opts = struct {
		Help              options.Help `getopt:"--help                       Display this help page"`
		NewBrightness     string       `getopt:"--brightness -b=[-+]INT[%]   Define the new keyboard brightness"`
		ShowBrightness    bool         `getopt:"--current -c                 Show current keyboard brightness in %"`
		ShowMaxBrightness bool         `getopt:"--maximum -m                 Show maximum keyboard brightness"`
		ShowVersion       bool         `getopt:"--version -v                 Show version information"`
	}{ShowVersion: false}
)

type BrightnessArgComputeData struct {
	NewBrightnessInt int
	Operator         string
	Percentage       bool
}

func GetCurrentBrightness(Bus dbus.BusObject) (err error) {
	err = Bus.Call("org.freedesktop.UPower.KbdBacklight.GetBrightness", 0).Store(&CurrentBrightness)
	if err != nil {
		return fmt.Errorf("Failed to get current brightness\nDetails: %s", err)
	}

	return nil
}

func GetMaxBrightness(Bus dbus.BusObject) (err error) {
	err = Bus.Call("org.freedesktop.UPower.KbdBacklight.GetMaxBrightness", 0).Store(&MaxBrightness)
	if err != nil {
		return fmt.Errorf("Failed to get maximum brightness\nDetails: %s", err)
	}

	return nil
}

func SetBrightness(Bus dbus.BusObject, NewBrightness int) error {
	_ = Bus.Call("org.freedesktop.UPower.KbdBacklight.SetBrightness", 0, NewBrightness)

	return nil
}

// Further parsing --brightness
func ParseBrightnessArg(BrightnessArgStr string) (bdata *BrightnessArgComputeData, err error) {
	BrightnessArgComputeData := BrightnessArgComputeData{}
	var newString = make([]string, 0)

	for i := 0; i < len(BrightnessArgStr); i++ {
		stringValue := string(BrightnessArgStr[i])

		if i == 0 {
			switch stringValue {
			case "-":
				BrightnessArgComputeData.Operator = "-"
				continue
			case "+":
				BrightnessArgComputeData.Operator = "+"
				continue
			}
		} else if i == len(BrightnessArgStr)-1 && stringValue == "%" {
			BrightnessArgComputeData.Percentage = true
			continue
		}

		newString = append(newString, stringValue)
	}

	finalIntValue := strings.Join(newString, "")
	BrightnessArgComputeData.NewBrightnessInt, err = strconv.Atoi(finalIntValue)
	if err != nil {
		return nil, fmt.Errorf("Invalid input type")
	}

	return &BrightnessArgComputeData, nil
}

// Compute the brightness value for SetBrightness()
func ComputeNewBrightnessValue(bCompData *BrightnessArgComputeData) (newBrightness int, err error) {
	// Round result and convert float64 to int
	var percentageValue = int(math.Round((float64(bCompData.NewBrightnessInt) / 100.0) * float64(MaxBrightness)))

	switch bCompData.Percentage {
	case true:
		if bCompData.Operator == "+" {
			newBrightness = CurrentBrightness + percentageValue
		} else if bCompData.Operator == "-" {
			newBrightness = CurrentBrightness - percentageValue
		} else {
			newBrightness = percentageValue
		}
	case false:
		if bCompData.Operator == "+" {
			newBrightness = CurrentBrightness + bCompData.NewBrightnessInt
		} else if bCompData.Operator == "-" {
			newBrightness = CurrentBrightness - bCompData.NewBrightnessInt
		} else {
			newBrightness = bCompData.NewBrightnessInt
		}
	default:
		return 0, fmt.Errorf("Could not compute new keyboard brightness!")
	}

	if newBrightness < 0 {
		newBrightness = 0
	} else if newBrightness > 255 {
		newBrightness = MaxBrightness
	}

	return newBrightness, nil
}

// Parse command-line arguments
func parseArgs() (err error) {
	// Command line arguments parsing
	options.Register(&opts)
	options.Parse()
	err = options.Validate(&opts)
	if err != nil {
		return fmt.Errorf("Failed to parse arguments!\nError: %s\n", err)
	}

	return nil
}

func main() {
	// Parse command-line arguments
	err := parseArgs()
	if err != nil {
		fmt.Printf("genesis: Error: %s\n", err)
		os.Exit(1)
	}

	// Connect to system dbus
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("genesis: Failed to connect to system dbus\nDetails: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Get the D-Bus Object that handles the keyboard light levels
	Bus := conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower/KbdBacklight")

	// Get current keyboard brightness
	if err = GetCurrentBrightness(Bus); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(2)
	}

	// Get maximum keyboard brightness
	if err = GetMaxBrightness(Bus); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(2)
	}

	// Setting the new keyboard brightness
	if opts.NewBrightness != "" {
		// Parsing and validating the --brightness option argument to control the arithmetic operations
		bdata, err := ParseBrightnessArg(opts.NewBrightness)
		if err != nil {
			fmt.Printf("genesis: %s\n", err)
			os.Exit(2)
		}

		// Computing the new brightness level from the returned data generated by parsing
		// the --brightness option argument
		NewBrightness, err = ComputeNewBrightnessValue(bdata)
		if err != nil {
			fmt.Printf("genesis: %s\n", err)
			os.Exit(2)
		}
		fmt.Printf("Brightness Level: %d\n", NewBrightness)

		// Setting new brightness level
		if err = SetBrightness(Bus, NewBrightness); err != nil {
			fmt.Printf("genesis: Failed to set keyboard brightness\nDetails: %s\n", err)
			os.Exit(2)
		}
	}
}
