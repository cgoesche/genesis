package main

import (
	"os"
	"fmt"
  "github.com/godbus/dbus"
  "github.com/pborman/options"
  "strconv"
  "strings"
  "math"
)

const(
  Version string = "Genesis version 0.9.1, build for GNU/Linux on ARM64"
)

var(
  CurrentBrightness           int
  CurrentBrightnessPercentage int
  MaxBrightness               int
  NewBrightness               int

  opts = struct {
		Help              options.Help  `getopt:"--help                       Display this help page"`
    Animation         bool          `getopt:"--animate -a                 Start a smooth keyboard light show"`
    NewBrightness     string        `getopt:"--brightness -b=[-+]INT[%]   Define the new keyboard brightness"`
    ShowBrightness    bool          `getopt:"--current -c                 Show current keyboard brightness in decimal"`
    ShowBrightnessPc  bool          `getopt:"--percentage -p              Show current keyboard brightness in %"`
    TurnBrightnessOn  bool          `getopt:"--on                         Set maximum keyboard brightness"`
    TurnBrightnessOff bool          `getopt:"--off                        Set keyboard brightness to 0"`
    AutoAdjust        bool          `getopt:"--auto                       Automatically adjust the keyboard brightness"`
    ShowMaxBrightness bool          `getopt:"--maximum -m                 Show maximum keyboard brightness"`
    ShowVersion       bool          `getopt:"--version -v                 Show version information"`
  }{
    Animation:false,
    ShowBrightness:false,
    ShowBrightnessPc:false,
    TurnBrightnessOn:false,
    TurnBrightnessOff:false,
    AutoAdjust:false,
    ShowMaxBrightness:false,
    ShowVersion:false,
  }

)

type BrightnessArgComputeData struct {
  NewBrightnessInt    int
  Operator            string
  Percentage          bool
}

func GetCurrentBrightness(Bus dbus.BusObject) (err error) {
  err = Bus.Call("org.freedesktop.UPower.KbdBacklight.GetBrightness", 0).Store(&CurrentBrightness)
  if err != nil {
    return fmt.Errorf("Failed to get current brightness\nDetails: %s", err)
  }
  CurrentBrightnessPercentage = int(math.Round( ( float64(CurrentBrightness) / float64(MaxBrightness)) * 100.0))

  return nil
}

func GetMaxBrightness(Bus dbus.BusObject) (err error) {
  err = Bus.Call("org.freedesktop.UPower.KbdBacklight.GetMaxBrightness", 0).Store(&MaxBrightness)
  if err != nil {
    return fmt.Errorf("Failed to get maximum brightness\nDetails: %s", err)
  }

  return nil
}

func SetBrightness(Bus dbus.BusObject, NewBrightness int) (error) {
  _ = Bus.Call("org.freedesktop.UPower.KbdBacklight.SetBrightness", 0, NewBrightness )

  return nil
}

// Further parsing --brightness
func ParseBrightnessArg(BrightnessArgStr string) (bdata *BrightnessArgComputeData, err error) {
  BrightnessArgComputeData := BrightnessArgComputeData{}
  var newString = make([]string, 0)

  for i := 0; i < len(BrightnessArgStr); i++{
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
    } else if i == len(BrightnessArgStr) - 1 && stringValue == "%" {
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
  var percentageValue = int( math.Round(( float64(bCompData.NewBrightnessInt) / 100.0 ) * float64(MaxBrightness) ))

  switch bCompData.Percentage {
    case true:
      if bCompData.Operator == "+" {
        newBrightness = CurrentBrightness + percentageValue
      }else if bCompData.Operator == "-" {
        newBrightness = CurrentBrightness - percentageValue
      }else {
        newBrightness = percentageValue
      }
    case false:
      if bCompData.Operator == "+" {
        newBrightness = CurrentBrightness + bCompData.NewBrightnessInt
      }else if bCompData.Operator == "-" {
        newBrightness = CurrentBrightness - bCompData.NewBrightnessInt
      }else {
        newBrightness = bCompData.NewBrightnessInt
      }
    default:
      return 0, fmt.Errorf("Could not compute new keyboard brightness!")
  }

  if newBrightness < 0 {
    newBrightness = 0
  }else if newBrightness > 255 {
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

  // Get maximum keyboard brightness
  if err = GetMaxBrightness(Bus); err != nil {
    fmt.Printf("%s\n", err)
    os.Exit(2)
  }

  // Get current keyboard brightness
  if err = GetCurrentBrightness(Bus); err != nil {
    fmt.Printf("%s\n", err)
    os.Exit(2)
  }

  // Testing options that will exit program before performing
  if opts.ShowVersion == true {
    fmt.Printf("%s\n", Version)
    os.Exit(0)
  } else if opts.ShowBrightness == true {
    fmt.Printf("%d\n", CurrentBrightness)
    os.Exit(0)
  } else if opts.ShowBrightnessPc == true {
    fmt.Printf("%d%%\n", CurrentBrightnessPercentage)
    os.Exit(0)
  } else if opts.ShowMaxBrightness == true {
    fmt.Printf("%d\n", MaxBrightness)
    os.Exit(0)
  } else if opts.TurnBrightnessOn == true {
     if err = SetBrightness(Bus, MaxBrightness); err != nil {
      fmt.Printf("genesis: Failed to set keyboard brightness\nDetails: %s\n", err)
      os.Exit(2)
    }
    os.Exit(0)
  } else if opts.TurnBrightnessOff == true {
     if err = SetBrightness(Bus, 0); err != nil {
      fmt.Printf("genesis: Failed to set keyboard brightness\nDetails: %s\n", err)
      os.Exit(2)
    }
    os.Exit(0)
  } else if opts.AutoAdjust == true {
    switch  {
      case CurrentBrightness <= MaxBrightness && CurrentBrightness > 0:
        NewBrightness = 0
      case CurrentBrightness >= 0:
        NewBrightness = MaxBrightness
      default:
        NewBrightness = CurrentBrightness
    }

     if err = SetBrightness(Bus, NewBrightness); err != nil {
      fmt.Printf("genesis: Failed to set keyboard brightness\nDetails: %s\n", err)
      os.Exit(2)
    }
    os.Exit(0)
  }

  // Setting the new keyboard brightness
  if opts.NewBrightness != "" {
    // Parsing and validating the --brightness option argument to control the arithmetic operations
    bArgData, err := ParseBrightnessArg(opts.NewBrightness)
    if err != nil {
      fmt.Printf("genesis: %s\n", err)
      os.Exit(2)
    }

    // Computing the new brightness level from the returned data generated by parsing
    // the --brightness option argument
    NewBrightness, err = ComputeNewBrightnessValue(bArgData)
    if err != nil {
      fmt.Printf("genesis: %s\n", err)
      os.Exit(2)
    }

    // Setting new brightness level
    if err = SetBrightness(Bus, NewBrightness); err != nil {
      fmt.Printf("genesis: Failed to set keyboard brightness\nDetails: %s\n", err)
      os.Exit(2)
    }
  }
}
