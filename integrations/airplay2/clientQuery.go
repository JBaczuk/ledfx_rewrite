package airplay2

import (
	"fmt"
	"github.com/grantmd/go-airplay" //nolint:typecheck
	log "ledfx/logger"
	"net"
	"regexp"
)

func queryDevice(params ClientDiscoveryParameters) (*airplay.AirplayDevice, error) {
	switch {
	case params.DeviceIP != "":
		ip := net.ParseIP(params.DeviceIP)
		if ip == nil {
			return nil, fmt.Errorf("could not parse IP address '%s'", params.DeviceIP)
		}
		return queryDeviceByIP(ip, params.Verbose)
	case params.DeviceNameRegex != "":
		return queryDeviceByName(params.DeviceNameRegex, params.Verbose)
	default:
		return nil, fmt.Errorf("either DeviceNameRegex or DeviceIP must be populated in the client discovery parameters")
	}
}

func queryDeviceByIP(ip net.IP, verbose bool) (device *airplay.AirplayDevice, err error) {
	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			if verbose {
				printDevice(&dev)
			}
			if dev.IP.Equal(ip) {
				return &dev, nil
			}
		}
	}
}

func queryDeviceByName(name string, verbose bool) (device *airplay.AirplayDevice, err error) {
	rxp, err := regexp.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression: %w", err)
	}

	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			if verbose {
				printDevice(&dev)
			}
			if dev.IP == nil || dev.Type != "airplay" {
				continue
			}

			// Did we find it?
			if rxp.MatchString(dev.Name) {
				// Yes, we did.
				return &dev, nil
			}
		}
	}
}

func printDevice(device *airplay.AirplayDevice) {
	log.Logger.WithField("category", "AirPlay Discovery").Infof(
		`NAME="%s" SERVER="%s:%d" HOSTNAME="%s" AUDIO="%dch/%dhz/%d-bit" PCM="%v" ALAC="%v"`,
		device.Name,
		device.IP,
		device.Port,
		device.Hostname,
		device.AudioChannels(),
		device.AudioSampleRate(),
		device.AudioSampleSize(),
		determinePCM(device),
		determineALAC(device),
	)
}

func determinePCM(device *airplay.AirplayDevice) bool {
	for _, c := range device.AudioCodecs() {
		if c == 0 {
			return true
		}
	}
	return false
}

func determineALAC(device *airplay.AirplayDevice) bool {
	for _, c := range device.AudioCodecs() {
		if c == 1 {
			return true
		}
	}
	return false
}
