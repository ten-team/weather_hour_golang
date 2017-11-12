package main

/*
// Modify jgarff/rpi_ws281x
// Add SetBrightness
diff --git a/golang/ws2811/ws2811.go b/golang/ws2811/ws2811.go
index 483b668..613cc8a 100644
--- a/golang/ws2811/ws2811.go
+++ b/golang/ws2811/ws2811.go
@@ -82,6 +82,10 @@ func SetLed(index int, value uint32) {
        C.ws2811_set_led(&C.ledstring, C.int(index), C.uint32_t(value))
 }

+func SetBrightness(index int, value uint8) {
+       C.ws2811_set_brightness(&C.ledstring, C.int(index), C.uint8_t(value))
+}
+
 func Clear() {
        C.ws2811_clear(&C.ledstring)
 }
diff --git a/golang/ws2811/ws2811.go.h b/golang/ws2811/ws2811.go.h
index 48c12f5..0dba647 100644
--- a/golang/ws2811/ws2811.go.h
+++ b/golang/ws2811/ws2811.go.h
@@ -52,6 +52,10 @@ void ws2811_set_led(ws2811_t *ws2811, int index, uint32_t value) {
        ws2811->channel[0].leds[index] = value;
 }

+void ws2811_set_brightness(ws2811_t *ws2811, int index, uint8_t value) {
+       ws2811->channel[0].brightness = value;
+}
+
 void ws2811_clear(ws2811_t *ws2811) {
        for (int chan = 0; chan < RPI_PWM_CHANNELS; chan++) {
                ws2811_channel_t *channel = &ws2811->channel[chan];
*/

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
	APPID                 = ""
	LAT                   = ""
	LON                   = ""
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

func main() {
	log.Printf("Booting...\n")
	const (
		ERROR_LED             = iota
		CURRENT_WEATHER_LED   = iota
		FORECAST1_WEATHER_LED = iota
		FORECAST2_WEATHER_LED = iota
		FORECAST3_WEATHER_LED = iota

		GPIO_PIN                    = 18
		LED_COUNT                   = 5
		DEFAULT_BRIGHTNESS          = 255
		ERROR_BRIGHTNESS            = 255
		CURRENT_WEATHER_BRIGHTNESS  = 255
		FORECAST_WEATHER_BRIGHTNESS = 128
	)
	var error_color = color(0xFF, 0, 0)
	var sleep_time = time.Second * 30

	if err := ws2811.Init(GPIO_PIN, LED_COUNT, DEFAULT_BRIGHTNESS); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Printf("Booted\n")

	for {
		current, err := getCurrentWeather()
		if err != nil {
			log.Println(err)
			ws2811.SetBrightness(ERROR_LED, ERROR_BRIGHTNESS)
			ws2811.SetLed(ERROR_LED, error_color)
			time.Sleep(sleep_time)
			continue
		}

		ws2811.SetBrightness(CURRENT_WEATHER_LED, CURRENT_WEATHER_BRIGHTNESS)
		ws2811.SetLed(CURRENT_WEATHER_LED, weather2color(current.Weathers[0].Main))

		forecast, err := getForecast5Weather()
		if err != nil {
			log.Println(err)
			ws2811.SetBrightness(ERROR_LED, ERROR_BRIGHTNESS)
			ws2811.SetLed(ERROR_LED, error_color)
			time.Sleep(sleep_time)
			continue
		}

		ws2811.SetBrightness(FORECAST1_WEATHER_LED, FORECAST_WEATHER_BRIGHTNESS)
		ws2811.SetLed(FORECAST1_WEATHER_LED, weather2color(forecast.List[1].Weathers[0].Main))

		ws2811.SetBrightness(FORECAST2_WEATHER_LED, FORECAST_WEATHER_BRIGHTNESS)
		ws2811.SetLed(FORECAST2_WEATHER_LED, weather2color(forecast.List[2].Weathers[0].Main))

		ws2811.SetBrightness(FORECAST3_WEATHER_LED, FORECAST_WEATHER_BRIGHTNESS)
		ws2811.SetLed(FORECAST3_WEATHER_LED, weather2color(forecast.List[3].Weathers[0].Main))

		if err := ws2811.Render(); err != nil {
			log.Println(err)
			time.Sleep(sleep_time)
			continue
		}

		time.Sleep(sleep_time)
	}
}
