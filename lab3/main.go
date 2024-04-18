package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

const geocodeApiKey = "81c6fcc2-ee21-4cff-a88c-d5b4784f7b82"
const weatherApiKey = "053887bb3aef32d634171be47dc9ab82"
const opentreemapApiKey = "5ae2e3f221c38a28845f05b6b08295007e96d512b23eabe1cf8e2755"
const radius = 500

type point struct {
	text string
	lat  float64
	lng  float64
}
type weather struct {
	temp         float64
	minTemp      float64
	maxTemp      float64
	descrWeather string
}
type place struct {
	name string
	desc string
}

/* type WeatherHitsResponse struct {
	Hits WeatherHit `json:"hits"`
}

type WeatherHit struct {
	Name    string          `json:"name"`
	Country string          `json:"country"`
	Point   WeatherHitPoint `json:"point"`
}

type WeatherHitPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
} */

func castToMap(val []interface{}) []map[string]interface{} {
	var resMap []map[string]interface{}
	for _, v := range val {
		resMap = append(resMap, v.(map[string]interface{}))
	}
	return resMap
}

func geocodingApi(addres string, ch chan []point) {
	reqUrl := "https://graphhopper.com/api/1/geocode"
	req, err := http.NewRequest("", reqUrl, nil)
	if nil != err {
		panic(err)
	}

	query := req.URL.Query()
	//fmt.Println(query)

	query.Add("q", addres)
	query.Add("locale", "en")
	query.Add("limit", "5")
	query.Add("reverse", "false")
	query.Add("debug", "false")
	query.Add("provider", "default")
	query.Add("key", geocodeApiKey) //past key
	req.URL.RawQuery = query.Encode()
	//fmt.Println(query)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	dc := json.NewDecoder(res.Body)
	var result map[string]interface{}

	err = dc.Decode(&result)
	if nil != err {
		fmt.Printf("%s\n\n", err.Error())
	}

	hits := castToMap(result["hits"].([]interface{}))

	var points []point

	for _, v := range hits {
		text := fmt.Sprintf("Name: %s\nCountry: %s\n", v["name"], v["country"])
		lat := (v["point"].(map[string]interface{}))["lat"]
		lng := (v["point"].(map[string]interface{}))["lng"]
		points = append(points, point{text, lat.(float64), lng.(float64)})
	}

	ch <- points
}

func weatherApi(lat, lng float64, ch chan weather) {
	reqUrl := "https://api.openweathermap.org/data/2.5/weather"
	req, err := http.NewRequest("", reqUrl, nil)
	if nil != err {
		panic(err)
	}

	query := req.URL.Query()
	query.Add("lat", fmt.Sprintf("%g", lat))
	query.Add("lon", fmt.Sprintf("%g", lng))
	query.Add("locale", "ru")
	query.Add("appid", weatherApiKey)
	query.Add("units", "metric")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()

	dc := json.NewDecoder(res.Body)
	var result map[string]interface{}
	err = dc.Decode(&result)
	if nil != err {
		panic(err)
	}

	resWeather := castToMap(result["weather"].([]interface{}))
	descr := resWeather[0]["description"].(string)

	resMain := result["main"].(map[string]interface{})
	tmp := resMain["temp"].(float64)
	minTmp := resMain["temp_min"].(float64)
	maxTmp := resMain["temp_max"].(float64)

	ch <- weather{tmp, minTmp, maxTmp, descr}
	close(ch)
}

func openTreeMapRadiusApi(lat, lng float64, radius int, ch chan place) {
	var reqUrl = "http://api.opentripmap.com/0.1/ru/places/radius"

	req, err := http.NewRequest("", reqUrl, nil)
	if nil != err {
		panic(err)
	}

	query := req.URL.Query()
	query.Add("lat", fmt.Sprint(lat))
	query.Add("lon", fmt.Sprint(lng))
	query.Add("radius", fmt.Sprint(radius))
	query.Add("lang", "ru")
	query.Add("apikey", opentreemapApiKey)
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()

	dc := json.NewDecoder(res.Body)
	var result map[string]interface{}
	err = dc.Decode(&result)
	if nil != err {
		panic(err)
	}

	tmp := castToMap(result["features"].([]interface{}))
	var wg sync.WaitGroup

	wg.Add(len(tmp))
	for _, v := range tmp {
		go func(m map[string]interface{}) {
			defer wg.Done()
			openTreeMapXidApi(m["properties"].(map[string]interface{})["xid"].(string), ch)
		}(v)
	}

	wg.Wait()
	close(ch)
}

func openTreeMapXidApi(xid string, ch chan place) {
	var reqUrl = "http://api.opentripmap.com/0.1/ru/places/xid"

	req, err := http.NewRequest("", reqUrl, nil)
	if nil != err {
		panic(err)
	}

	//fmt.Println(req.URL.Path)
	req.URL = req.URL.JoinPath(xid)
	//fmt.Println(req.URL.Path)

	query := req.URL.Query()
	query.Add("apikey", opentreemapApiKey)
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()

	dc := json.NewDecoder(res.Body)
	var result map[string]interface{}
	err = dc.Decode(&result)
	if nil != err {
		panic(err)
	}

	if tmp := result["error"]; nil != tmp {
		fmt.Println(tmp)
		return
	}

	var name, desc string
	name = result["name"].(string)
	if name == "" {
		name = "No name"
	}

	info := result["info"]
	if nil != info {
		if tmp := info.(map[string]interface{})["descr"]; nil != tmp {
			desc = tmp.(string)
		}
	}
	if desc == "" {
		desc = "No description"
	}

	ch <- place{name, desc}
}

func main() {
	fmt.Printf("Введите место: ")

	r := bufio.NewReader(os.Stdin)
	addres, err := r.ReadString('\n')
	if nil != err {
		panic(err)
	}

	var pointsCh = make(chan []point)
	go geocodingApi(addres, pointsCh)
	///
	///
	///
	points := <-pointsCh
	for i, v := range points {
		fmt.Printf("%d)%sШирота: %g\nДолгота: %g\n\n", i+1, v.text, v.lat, v.lng)
	}

	fmt.Printf("Выберите точку: ")
	var point int
	fmt.Scan(&point)
	point--

	weatherCh := make(chan weather)
	placesCh := make(chan place)
	go weatherApi(points[point].lat, points[point].lng, weatherCh)
	go openTreeMapRadiusApi(points[point].lat, points[point].lng, radius, placesCh)
	///
	///
	///
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Printf("%sШирота: %g\nДолгота: %g\n\n", points[point].text, points[point].lat, points[point].lng)
	resWeather := <-weatherCh
	fmt.Printf("Температура:%f\n Диапазон:%f-%f\n Погода:%s\n", resWeather.temp, resWeather.minTemp, resWeather.maxTemp, resWeather.descrWeather)
	fmt.Printf("\nИнтересные места в радиусе %d метров:\n", radius)
	for {
		placeInfo, status := <-placesCh
		if !status {
			break
		}
		fmt.Printf("Название: %s\nИнформация: %s\n", placeInfo.name, placeInfo.desc)
	}

}
