package main

import (
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
	"log"
	"os"
	"time"
)

func color(red uint8, green uint8, blue uint8) uint32 {
	return uint32(red)<<8 + uint32(green)<<16 + uint32(blue)
}

func main() {
	log.Printf("Booting...\n")
	const (
		GPIO_PIN           = 18
		LED_COUNT          = 3
		DEFAULT_BRIGHTNESS = 255
	)
	if err := ws2811.Init(GPIO_PIN, LED_COUNT, DEFAULT_BRIGHTNESS); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	for {
		ws2811.SetLed(0, color(0, 0, 0xFF))
		ws2811.SetBrightness(1, 255)
		if err := ws2811.Render(); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second)

		ws2811.SetLed(0, color(0, 0, 0xFF))
		ws2811.SetBrightness(1, 128)
		if err := ws2811.Render(); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second)

		ws2811.SetLed(0, 0)
		ws2811.SetBrightness(1, 255)
		if err := ws2811.Render(); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second)
	}
}
