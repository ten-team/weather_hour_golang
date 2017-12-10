package main

import (
	"flag"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
	"log"
	"os"
)

func color(red uint8, green uint8, blue uint8) uint32 {
	return uint32(red)<<8 + uint32(green)<<16 + uint32(blue)
}

func weather2color(weather string) uint32 {
	var (
		clear   = color(255, 170, 0)
		clouds  = color(170, 170, 170)
		rain    = color(0, 65, 255)
		snow    = color(242, 242, 255)
		unknown = color(255, 0, 0)
	)
	// see http://www.jma.go.jp/jma/kishou/info/colorguide/120524_hpcolorguide.pdf
	switch weather {
	case "Thunderstorm":
		return rain
	case "Drizzle":
		return rain
	case "Rain":
		return rain
	case "Snow":
		return snow
	case "Atmosphere":
		return clouds
	case "Clear":
		return clear
	case "Clouds":
		return clouds
	case "Extreme":
		return unknown
	case "Additional":
		return unknown
	default:
		log.Printf("unknown weather: %s\n", weather)
		return unknown
	}
}

func setLeds(leds []int, color uint32) error {
	for _, led := range leds {
		ws2811.SetLed(led, color)
	}
	return nil
}

var (
	weater0Option = flag.String("w0", "Clear", "weather (Clear, Clouds, Rain, or Snow)")
	weater1Option = flag.String("w1", "Clear", "weather (Clear, Clouds, Rain, or Snow)")
	weater2Option = flag.String("w2", "Clear", "weather (Clear, Clouds, Rain, or Snow)")
	weater3Option = flag.String("w3", "Clear", "weather (Clear, Clouds, Rain, or Snow)")
)

func main() {
	flag.Parse()

	const (
		GPIO_PIN           = 18
		LED_COUNT          = 26
		DEFAULT_BRIGHTNESS = 255
	)
	led_matrix := [][]int{
		{20, 21, 22, 23, 24, 25},
		{14, 15, 16, 17, 18, 19},
		{7, 8, 9, 10, 11, 12, 13},
		{0, 1, 2, 3, 4, 5, 6}}

	if err := ws2811.Init(GPIO_PIN, LED_COUNT, DEFAULT_BRIGHTNESS); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	setLeds(led_matrix[0], weather2color(*weater0Option))
	setLeds(led_matrix[1], weather2color(*weater1Option))
	setLeds(led_matrix[2], weather2color(*weater2Option))
	setLeds(led_matrix[3], weather2color(*weater3Option))
	if err := ws2811.Render(); err != nil {
		log.Println(err)
	}
}
