/*Practitioner: Rihad Variawa
Authors: Sarah Hsu and Caryn Willis
Description: This program takes in roof and house size and creates
a visual of the United States with color coded recommendations for
each city.*/

//This file is written by Ryder Malastare

package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//This section asks the user for their house and roof size
func DisplayHouseSize(w http.ResponseWriter, r *http.Request) {
	PageTitle := "Heat Map"

	var MyRoof float64

	MyHouse := []House{
		House{"housesizeinput", 0, "Size"},
	}

	PageVars := PageVariables{
		PageTitle:     PageTitle,
		PageHouseSize: MyHouse,
		PageRoofSize:  MyRoof,
	}

	t, err := template.ParseFiles("housesizemap.html") //Parse the html file housesizemap.html
	if err != nil {
		log.Print("template parsing error:", err)
	}

	err = t.Execute(w, PageVars) //execute the template and pass it the PageVars
	if err != nil {
		log.Print("template executing error:", err)
	}

}

//This section is where the user can look at their results based on their input
//It will display a map based on recommendation (based on their % of energy
//covered) and also give the cities which fall into each recommendation
//category.
func UserInteracts(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse the page for the variables needed
	red := "red"
	yellow := "yellow"
	green := "green"
	cityData := MakeCityMap("energy.csv")
	houseSize, err1 := strconv.ParseFloat(r.Form.Get("housesizeinput"), 64)
	ErrorMessage(err1, "house size", houseSize)
	roofSize, err2 := strconv.ParseFloat(r.Form.Get("roofsize"), 64)
	ErrorMessage(err2, "roof size", roofSize)
	heatMap := MakeColorMarkers(cityData, houseSize, roofSize)
	mapColors := MakeColors("energy.csv", cityData, houseSize, roofSize)
	redList := MakeList(heatMap, red)
	yellowList := MakeList(heatMap, yellow)
	greenList := MakeList(heatMap, green)
	redPercentage := ColorPercent(heatMap, "red")
	yellowPercentage := ColorPercent(heatMap, "yellow")
	greenPercentage := ColorPercent(heatMap, "green")
	Title := "House Size Map"

	PageVars := PageVariables{
		PageTitle:     Title,
		Map:           mapColors,
		RedList:       redList,
		YellowList:    yellowList,
		GreenList:     greenList,
		RedPercent:    redPercentage,
		YellowPercent: yellowPercentage,
		GreenPercent:  greenPercentage,
	}

	t, err := template.ParseFiles("housesizemap.html")
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	err = t.Execute(w, PageVars)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

//Makes a map of color markers for each city based on chosen house size and difference in output
func MakeColorMarkers(cityData map[string]City, houseSize, roofSize float64) map[string]string {
	var output, avgEnergy float64
	var mapColor string
	colors := make(map[string]string)
	for cityName, _ := range cityData {
		output = SolarOutput(cityName, cityData, "horizontal", 15, roofSize)
		output = float64(int(output*100)) / 100
		avgEnergy = AverageEnergy(cityData, cityName) * houseSize
		avgEnergy = float64(int(avgEnergy*100)) / 100
		mapColor = MapColor(avgEnergy, output)
		colors[cityName] = mapColor
	}
	return colors
}

//Computes difference in energy and chooses color
func MapColor(avgEnergy, energyOutput float64) string {
	percentage := energyOutput / avgEnergy
	var color string
	if percentage >= .8 {
		color = "green"
	} else if percentage > 0.6 && percentage < 0.8 {
		color = "yellow"
	} else {
		color = "red"
	}
	return color
}

//Makes list of cities that have a certain color code
func MakeList(colors map[string]string, color string) []string {
	cityList := make([]string, 0)
	for cityName, mapColor := range colors {
		if mapColor == color {
			cityList = append(cityList, cityName)
		}
	}
	return cityList
}

//Computes the percentage of the cities that are defined as a certain color
func ColorPercent(colors map[string]string, color string) float64 {
	var colorCount int
	for _, mapColor := range colors {
		if mapColor == color {
			colorCount++
		}
	}
	return float64(int((float64(colorCount)/98)*1000)) / 10
	//since there are 98 cities
}

//Make an array of the city names
func MakeCityArray(filename string) []string {
	lines := ReadFile(filename)
	cityArray := make([]string, 0)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		cityArray = append(cityArray, items[0])
	}
	return cityArray
}

//Make an array of colors based alphabetically
func MakeColors(filename string, cityData map[string]City, houseSize, roofSize float64) []string {
	var output, avgEnergy float64
	var mapColor string
	cityNames := MakeCityArray(filename)
	colors := make([]string, 0)
	for _, cityName := range cityNames {
		output = SolarOutput(cityName, cityData, "horizontal", 15, roofSize)
		avgEnergy = AverageEnergy(cityData, cityName) * houseSize
		mapColor = MapColor(avgEnergy, output)
		if mapColor == "red" {
			mapColor = "#FF0000"
		} else if mapColor == "yellow" {
			mapColor = "#FFFF00"
		} else {
			mapColor = "#008000"
		}
		colors = append(colors, mapColor)
	}
	return colors
}
