package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

const owmApiUrl = "https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s"

var errorCnt int32

type OwmData struct {
	Main Main   `json:"main"`
	Name string `json:"name"`
}

type Main struct {
	Temp     float32 `json:"temp"`
	Humidity int8    `json:"humidity"`
	Pressure int16   `json:"pressure"`
}

func main() {
	http.HandleFunc("/metrics", Export)

	err := http.ListenAndServe(":8091", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Export(w http.ResponseWriter, _ *http.Request) {
	owmData, err := ReadTemperature("Thun,CH")

	if err != nil {
		log.Print(err)
		_, _ = fmt.Fprintf(w, "owm_error %d\n", atomic.AddInt32(&errorCnt, 1))
		return
	}

	_, _ = fmt.Fprintf(w, "owm_temperature{location=\"%s\"} %f\n", owmData.Name, owmData.Main.Temp)
	_, _ = fmt.Fprintf(w, "owm_humidity{location=\"%s\"} %d\n", owmData.Name, owmData.Main.Humidity)
	_, _ = fmt.Fprintf(w, "owm_pressure{location=\"%s\"} %d\n", owmData.Name, owmData.Main.Pressure)
	w.WriteHeader(200)
}

func ReadTemperature(location string) (OwmData, error) {

	owm := OwmData{}
	res, err := http.Get(fmt.Sprintf(owmApiUrl, location, os.Getenv("OWM_API_KEY")))

	if err != nil {
		return owm, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return owm, err
	}

	if res.StatusCode >= 400 {
		return owm, errors.New(string(body))
	}

	err = json.Unmarshal(body, &owm)
	if err != nil {
		return OwmData{}, err
	}

	return owm, nil
}
