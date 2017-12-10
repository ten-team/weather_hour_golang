package main

import (
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
	"log"
	"math/rand"
	"os"
	"time"
)

func color(red uint8, green uint8, blue uint8) uint32 {
	return uint32(red)<<8 + uint32(green)<<16 + uint32(blue)
}

func weather2color() uint32 {
	var (
		clear  = color(255, 170, 0)
		clouds = color(170, 170, 170)
		rain   = color(0, 65, 255)
	)
	switch rand.Intn(5) {
	case 0, 1, 2:
		return clear
	case 3:
		return clouds
	case 4:
		return rain
	}
	return color(255, 0, 0)
}

func setLeds(leds []int, color uint32) error {
	for _, led := range leds {
		ws2811.SetLed(led, color)
	}
	return nil
}

func main() {
	log.Printf("Booting...\n")
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

	for {
		setLeds(led_matrix[0], weather2color())
		setLeds(led_matrix[1], weather2color())
		setLeds(led_matrix[2], weather2color())
		setLeds(led_matrix[3], weather2color())
		if err := ws2811.Render(); err != nil {
			log.Println(err)
		}

		time.Sleep(time.Second * 10)
	}
}
