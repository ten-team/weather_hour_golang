package main

import (
	"encoding/json"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	CURRENT_WEATHER_URL   = "http://api.openweathermap.org/data/2.5/weather"
	FORECAST5_WEATHER_URL = "http://api.openweathermap.org/data/2.5/forecast"
	APPID                 = "your appid"
	LAT                   = "your lat"
	LON                   = "your lon"
)

type CurrentWeatherApiResult struct {
	Weathers []Weather `json:"weather"`
}

type Forecast5WeatherApiResult struct {
	List []ForecastWeather `json:"list"`
}

type ForecastWeather struct {
	Weathers []Weather `json:"weather"`
	DtText   string    `json:"dt_txt"`
}

// see https://openweathermap.org/weather-conditions
type Weather struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func getWeatherImpl(weather_url string) (body []byte, err error) {
	values := url.Values{}
	values.Add("APPID", APPID)
	values.Add("lat", LAT)
	values.Add("lon", LON)

	resp, err := http.Get(weather_url + "?" + values.Encode())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return b, nil
}

// see https://openweathermap.org/current
func getCurrentWeather() (weather CurrentWeatherApiResult, err error) {
	body, err := getWeatherImpl(CURRENT_WEATHER_URL)
	if err != nil {
		return CurrentWeatherApiResult{}, err
	}
	var result CurrentWeatherApiResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println(err)
		return CurrentWeatherApiResult{}, err
	}
	log.Printf("current weather: %s\n", result.Weathers[0].Main)
	return result, nil
}

// see https://openweathermap.org/forecast5
func getForecast5Weather() (weather Forecast5WeatherApiResult, err error) {
	body, err := getWeatherImpl(FORECAST5_WEATHER_URL)
	if err != nil {
		return Forecast5WeatherApiResult{}, err
	}
	var result Forecast5WeatherApiResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println(err)
		return Forecast5WeatherApiResult{}, err
	}
	log.Printf("weather: %s, %s\n", result.List[1].Weathers[0].Main, result.List[1].DtText)
	log.Printf("weather: %s, %s\n", result.List[2].Weathers[0].Main, result.List[2].DtText)
	log.Printf("weather: %s, %s\n", result.List[3].Weathers[0].Main, result.List[3].DtText)
	return result, nil
}

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

func weather2darkcolor(weather string) uint32 {
	c := weather2color(weather)
	red := uint8((c & 0x00FF00) >> 8)
	green := uint8((c & 0xFF0000) >> 16)
	blue := uint8(c & 0x0000FF)
	return color(red/8, green/8, blue/8)
}

func setLeds(leds []int, color uint32) error {
	for _, led := range leds {
		ws2811.SetLed(led, color)
	}
	return nil
}

func setErrorLedImpl(error_color uint32) error {
	const (
		ERROR_LED = 0
	)
	ws2811.SetLed(ERROR_LED, error_color)
	if err := ws2811.Render(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func setErrorLed() error {
	return setErrorLedImpl(color(0xFF, 0, 0))
}

func unsetErrorLed() error {
	return setErrorLedImpl(color(0x00, 0, 0))
}

func main() {
	log.Printf("Booting...\n")
	const (
		GPIO_PIN           = 18
		LED_COUNT          = 26
		DEFAULT_BRIGHTNESS = 255
	)
	var sleep_time = time.Second * 30
	led_matrix := [][]int{
		{1, 2, 3, 4, 5, 6, 7},
		{8, 9, 10, 11, 12, 13},
		{14, 15, 16, 17, 18, 19},
		{20, 21, 22, 23, 24, 25}}

	if err := ws2811.Init(GPIO_PIN, LED_COUNT, DEFAULT_BRIGHTNESS); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Printf("Booted\n")

	for {
		unsetErrorLed()
		current, err := getCurrentWeather()
		if err != nil {
			log.Println(err)
			setErrorLed()
			time.Sleep(sleep_time)
			continue
		}

		setLeds(led_matrix[0], weather2color(current.Weathers[0].Main))

		forecast, err := getForecast5Weather()
		if err != nil {
			log.Println(err)
			setErrorLed()
			time.Sleep(sleep_time)
			continue
		}

		setLeds(led_matrix[1], weather2darkcolor(forecast.List[1].Weathers[0].Main))
		setLeds(led_matrix[2], weather2darkcolor(forecast.List[2].Weathers[0].Main))
		setLeds(led_matrix[3], weather2darkcolor(forecast.List[3].Weathers[0].Main))
		if err := ws2811.Render(); err != nil {
			log.Println(err)
		}

		time.Sleep(sleep_time)
	}
}
