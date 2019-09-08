/*Practitioner: Rihad Variawa
Description: This program takes in roof, house size, and coordinates
to give to the user some useful information about
installing solar panels in their home.*/

/*This file is written by Sarah Hsu, except functions MakeCity(),
MakeCityMap(), and ReadFile() were written by Ryder Malastare.*/

package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*This is a city struct which stores all of the data for each city.
It stores the coordinates, temperature, solar radiation (at flat angle),
optimal angle, optimal radiation (at optimal angle), average energy usage,
installation cost, and a slice of 3 company names for each city.*/
type City struct {
	coordN    float64
	coordW    float64
	temp      float64
	solarRad  float64
	optAng    float64
	optRad    float64
	avgEnergy float64
	instCost  float64
	companies []string
}

/* This is a panel struct which stores the information for each type of solar
panel. We have chosen 6 panels for the user to choose from here, each with
information on its efficiency (percentage), watts, panel area, and price.*/
type Panel struct {
	efficiency float64
	watts      float64
	area       float64
	price      float64
}

/*This is a coordinates struct which has a identifying name (for the web
portion) , a value, and text (west or north) sections*/
type Coordinates struct {
	Name  string
	Value float64
	Text  string
}

/* This is a house struct which stores the identifying name (for the web
portion), a size value, and a text (name).*/
type House struct {
	Name  string
	Value float64
	Text  string
}

/*This is the struct storing all of the variables needed to be displayed
on the web app.*/
type PageVariables struct {
	PageTitle       string        //Title of the page
	PageCoordinates []Coordinates //Coordinates of the user
	PageHouseSize   []House       //House size of the user
	PageRoofSize    float64       //Roof size of the user
	MyCity          string        //City name that is closest to the user
	Output          float64       //Expected solar energy output
	OptAngle        float64       //Optimal angle for panels
	OptOutput       float64       //Optimal solar energy output
	Usage           float64       //Average energy usage
	Optimal         string        //Is it optimal to install solar power? Gives recommendation.
	InstCost        float64       //Installation cost
	Companies       []string      //3 company names
	NumPanels       []int         //Number of panels needed for each brand
	PanelCost       []int         //Cost of panels for each brand
	Recommendation  []string      //Recommendation for each of the user preferences (efficiency, cost, production)
	Percentage      int           //Percentage that their energy is covered by solar
	Map             []string      //Map colors for each city (red, yellow, green)
	RedList         []string      //List of cities in red
	YellowList      []string      //List of cities in yellow
	GreenList       []string      //List of cities in green
	RedPercent      float64       //Percentage of cities in red
	YellowPercent   float64       //Percentage of cities in yellow
	GreenPercent    float64       //Percentage of cities in green
}

func main() {
	http.HandleFunc("/", DisplayCoordinates)          //DisplayCoordinates() loads when called with / at the end of the URL
	http.HandleFunc("/selected", UserSelected)        //UserSelected() will load after the form with / is submitted
	http.HandleFunc("/heatmap", DisplayHouseSize)     //DisplayHouseSize() will load when URL is called with /heatmap, or click tab
	http.HandleFunc("/displayheatmap", UserInteracts) //UserInteracts() will load after form with /heatmap is submitted
	log.Fatal(http.ListenAndServe(getPort(), nil))
}

/*This function is to deployment to the web, or you can run it with localhost*/
func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return ":" + p
	}
	return ":8080"
}

//This is the function where it initially displays the page asking for the
//user's info such as house size and coordinates and then loads
//to the next page after this information is submitted.
func DisplayCoordinates(w http.ResponseWriter, r *http.Request) {

	Title := "Solar Energy"
	MyCoordinates := []Coordinates{
		Coordinates{"coordinaten", 0, "North"},
		Coordinates{"coordinatew", 0, "West"},
	}
	MyHouse := []House{
		House{"housesize", 0, "Size"},
	}

	var MyRoof float64

	MyPageVariables := PageVariables{
		PageTitle:       Title,
		PageCoordinates: MyCoordinates,
		PageHouseSize:   MyHouse,
		PageRoofSize:    MyRoof,
	}

	t, err := template.ParseFiles("solarenergy.html") //parse the html file solarenergy.html
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	err = t.Execute(w, MyPageVariables) //execute the template and pass it the MyPageVariables
	if err != nil {
		log.Print("template executing error: ", err)
	}

}

//This is the main function where the user interacts with the web app.
//There are several different variables in use here to be able to interact
//with.
func UserSelected(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse the page for the variables needed
	cityData := MakeCityMap("energy.csv")
	solarPanels := MakeSolarMap("solar.csv")
	northcoord, err1 := strconv.ParseFloat(r.Form.Get("coordinaten"), 64)
	ErrorMessage(err1, "north coordinate", northcoord)
	westcoord, err2 := strconv.ParseFloat(r.Form.Get("coordinatew"), 64)
	ErrorMessage(err2, "west coordinate", westcoord)
	houseSize, err3 := strconv.ParseFloat(r.Form.Get("housesize"), 64)
	ErrorMessage(err3, "house size", houseSize)
	roofSize, err4 := strconv.ParseFloat(r.Form.Get("roofsize"), 64)
	ErrorMessage(err4, "roof size", roofSize)
	closestcity := ClosestCity(cityData, northcoord, westcoord)
	solarOutput := SolarOutput(closestcity, cityData, "horizontal", 15, roofSize)
	solarOutput = float64(int(solarOutput*100)) / 100
	optAngle := OptAngle(cityData, closestcity)
	optEnergy := OptEnergy(cityData, closestcity, 15, roofSize)
	optEnergy = float64(int(optEnergy*100)) / 100
	avgUsage := AverageEnergy(cityData, closestcity) * houseSize
	avgUsage = float64(int(avgUsage*100)) / 100
	percent, recommendation := IsItOptimal(avgUsage, solarOutput)
	percentage := int(percent * 100)
	companylist := Companies(closestcity, cityData)
	instCost := InstallationCost(cityData, closestcity)
	instCost = float64(int(instCost*100)) / 100
	numPanels, panelCost := CalcCostBrand(solarOutput, roofSize, cityData, closestcity, solarPanels)
	preferences := Preferences(panelCost, solarPanels, closestcity, cityData, houseSize)

	Title := "Your Home"
	MyPageVariables := PageVariables{
		PageTitle:      Title,
		MyCity:         closestcity,
		Output:         solarOutput,
		OptAngle:       optAngle,
		OptOutput:      optEnergy,
		Usage:          avgUsage,
		Optimal:        recommendation,
		InstCost:       instCost,
		Companies:      companylist,
		NumPanels:      numPanels,
		PanelCost:      panelCost,
		Recommendation: preferences,
		Percentage:     percentage,
	}

	t, err := template.ParseFiles("solarenergy.html") //parse the html file solarenergy.html
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	err = t.Execute(w, MyPageVariables) //execute the template and pass it the MyPageVariables
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

//Read in the file.
func ReadFile(filename string) []string {
	//reads file and makes a line for each file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error: couldn't open the file")
		os.Exit(1)
	}
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Println("Error: there was some kind of error during the file reading")
		os.Exit(1)
	}
	file.Close()
	return lines
}

//Makes a map data structure of all of the City objects.
func MakeCityMap(filename string) map[string]City {
	lines := ReadFile(filename)
	cityData := make(map[string]City)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		cityName := items[0]
		cityData[cityName] = MakeCity(cityData, items)
	}
	delete(cityData, "")
	return cityData
}

//Creates a City object with its characteristics using City struct.
func MakeCity(cityData map[string]City, items []string) City {
	var city City
	city.coordN, _ = strconv.ParseFloat(items[1], 64)
	city.coordW, _ = strconv.ParseFloat(items[2], 64)
	city.temp, _ = strconv.ParseFloat(items[3], 64)
	city.solarRad, _ = strconv.ParseFloat(items[4], 64)
	city.optAng, _ = strconv.ParseFloat(items[5], 64)
	city.optRad, _ = strconv.ParseFloat(items[6], 64)
	city.avgEnergy, _ = strconv.ParseFloat(items[7], 64)
	city.instCost, _ = strconv.ParseFloat(items[8], 64)
	var companyNames []string = strings.Split(items[9], ";")
	for i := range companyNames {
		city.companies = append(city.companies, companyNames[i])
	}
	return city
}

//Make the map data structure of all of the different solar panel brands.
func MakeSolarMap(filename string) map[string]Panel {
	lines := ReadFile(filename)
	solarPanels := make(map[string]Panel)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		panelName := items[0]
		solarPanels[panelName] = MakePanel(solarPanels, items)
	}
	keys := make([]string, 0)
	for key, _ := range solarPanels {
		keys = append(keys, key)
	}
	return solarPanels
}

//Make a Solar Panel object using Panel struct.
func MakePanel(solarPanels map[string]Panel, items []string) Panel {
	var panel Panel
	panel.efficiency, _ = strconv.ParseFloat(items[1], 64)
	panel.watts, _ = strconv.ParseFloat(items[2], 64)
	panel.area, _ = strconv.ParseFloat(items[3], 64)
	panel.price, _ = strconv.ParseFloat(items[4], 64)
	return panel
}

//Finds the city closest to their coordinates.
func ClosestCity(cityData map[string]City, userCoordN, userCoordW float64) string {
	distance := 10000.00
	closestCityName := ""
	var cityDistance float64
	if userCoordN < 0 {
		userCoordN *= -1
	}
	if userCoordW < 0 {
		userCoordW *= -1
	}
	for city, data := range cityData {
		cityDistance = math.Sqrt(math.Pow((userCoordN-data.coordN), 2) + math.Pow((userCoordW-data.coordW), 2))
		if cityDistance < distance {
			distance = cityDistance
			closestCityName = city
		}
	}
	return closestCityName
}

//Calculates expected generated energy from solar panels. (in kwh per month)
func SolarOutput(cityName string, cityData map[string]City, angleType string, efficiency, houseSize float64) float64 {
	var radiation float64
	//241.5479 meters squared as solar panel area (average house size)
	//assume standard 15% efficiency
	//0.75 default performance ratio
	// E = Solar Panel Area * solar panel efficiency * radiation * performance ratio
	if angleType == "horizontal" {
		radiation = cityData[cityName].solarRad
	} else if angleType == "optimal" {
		radiation = cityData[cityName].optRad
	}
	houseSize *= 0.092903 //convert square feet to square meters
	energyOutput := houseSize * efficiency * radiation * 0.75
	return energyOutput / 12
}

//Gives the potential optimal energy output for their home.
func OptEnergy(cityData map[string]City, cityName string, efficiency, houseSize float64) float64 {
	optOutput := SolarOutput(cityName, cityData, "optimal", efficiency, houseSize)
	return optOutput
}

//Output the optimal angle that they should use to get the optimal output.
func OptAngle(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	optAngle := data.optAng
	return optAngle
}

//Calculates average energy required at a house in their area. (kwh per month)
func AverageEnergy(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	averageEnergy := data.avgEnergy
	return averageEnergy / 2600
}

//Gives a recommendation based on energy produced from solar panels and energy requirement.
func IsItOptimal(avgUsage float64, solarOutput float64) (float64, string) {
	percentage := solarOutput / avgUsage
	if percentage <= 0.60 {
		return percentage, "is not recommended"
	} else if percentage > .60 && percentage < .8 {
		return percentage, "is recommended"
	} else {
		return percentage, "is highly recommended"
	}
}

//Calculates the installation cost.
func InstallationCost(cityData map[string]City, cityName string) float64 {
	return cityData[cityName].instCost * 5000
}

//Calculates the number of solar panels needed on their house.
func NumSolarPanels(energyOutput, houseSize float64, cityData map[string]City, panelName, cityName string, solarPanels map[string]Panel) int {
	houseSize *= 0.092903 //convert square feet to square meters
	oneSolarPanelOutput := (energyOutput * 12 / houseSize) * solarPanels[panelName].area
	numPanels := cityData[cityName].avgEnergy / oneSolarPanelOutput
	return int(numPanels)
}

//Calculates how much it would cost for user to get that brand of solar panels on their house.
func SolarPanelCost(energyOutput, houseSize float64, cityData map[string]City, panelName, cityName string, solarPanels map[string]Panel, numPanels int) float64 {
	houseSize *= 0.092903 //convert square feet to square meters
	cost := solarPanels[panelName].price * float64(numPanels)
	return cost + InstallationCost(cityData, cityName)
}

//Puts companies close to their city into a slice of strings.
func Companies(cityName string, cityData map[string]City) []string {
	data := cityData[cityName]
	companyNames := data.companies
	return companyNames
}

//Calculates the cost and number of panels required for each brand of solar panel.
//0: Suntech, 1: Samsung, 2: Kyocera, 3: Canadian Solar, 4: Grape Solar 390W, 5: Grape Solar 250
func CalcCostBrand(energyOutput, houseSize float64, cityData map[string]City, cityName string, solarPanels map[string]Panel) ([]int, []int) {
	NumPanels := make([]int, 6)
	PanelCosts := make([]int, 6)
	NumPanels[0] = NumSolarPanels(energyOutput, houseSize, cityData, "Suntech", cityName, solarPanels)
	NumPanels[1] = NumSolarPanels(energyOutput, houseSize, cityData, "Samsung", cityName, solarPanels)
	NumPanels[2] = NumSolarPanels(energyOutput, houseSize, cityData, "Kyocera", cityName, solarPanels)
	NumPanels[3] = NumSolarPanels(energyOutput, houseSize, cityData, "CanadianSolar", cityName, solarPanels)
	NumPanels[4] = NumSolarPanels(energyOutput, houseSize, cityData, "GrapeSolar390W", cityName, solarPanels)
	NumPanels[5] = NumSolarPanels(energyOutput, houseSize, cityData, "GrapeSolar250", cityName, solarPanels)
	PanelCosts[0] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Suntech", cityName, solarPanels, NumPanels[0]))
	PanelCosts[1] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Samsung", cityName, solarPanels, NumPanels[1]))
	PanelCosts[2] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Kyocera", cityName, solarPanels, NumPanels[2]))
	PanelCosts[3] = int(SolarPanelCost(energyOutput, houseSize, cityData, "CanadianSolar", cityName, solarPanels, NumPanels[3]))
	PanelCosts[4] = int(SolarPanelCost(energyOutput, houseSize, cityData, "GrapeSolar390W", cityName, solarPanels, NumPanels[4]))
	PanelCosts[5] = int(SolarPanelCost(energyOutput, houseSize, cityData, "GrapeSolar250", cityName, solarPanels, NumPanels[5]))
	return NumPanels, PanelCosts
}

//Preferences in a slice, with 0: min cost, 1: max output, 2: max efficiency
func Preferences(panelCost []int, solarPanels map[string]Panel, cityName string, cityData map[string]City, houseSize float64) []string {
	minCostPanel := FindMinCostPanel(panelCost)
	efficiencyarray := MakeEfficiencyArray(solarPanels)
	mostEfficientPanel := FindMostEfficient(efficiencyarray)
	maxOutput := FindMaxOutput(efficiencyarray, cityName, cityData, houseSize)
	preferences := []string{minCostPanel, maxOutput, mostEfficientPanel}
	return preferences
}

//Converts index to panel brand name.
func IdxToPanel(idx int) string {
	switch idx {
	case 0:
		return "Suntech"
	case 1:
		return "Samsung"
	case 2:
		return "Kyocera"
	case 3:
		return "CanadianSolar"
	case 4:
		return "GrapeSolar390W"
	case 5:
		return "GrapeSolar250"
	}
	return ""
}

//Puts panel brand efficiencies in an array according to the same indices as above.
func MakeEfficiencyArray(solarPanels map[string]Panel) []float64 {
	efficiencyArray := make([]float64, 6)
	efficiencyArray[0] = solarPanels["Suntech"].efficiency
	efficiencyArray[1] = solarPanels["Samsung"].efficiency
	efficiencyArray[2] = solarPanels["Kyocera"].efficiency
	efficiencyArray[3] = solarPanels["CanadianSolar"].efficiency
	efficiencyArray[4] = solarPanels["GrapeSolar390W"].efficiency
	efficiencyArray[5] = solarPanels["GrapeSolar250"].efficiency
	return efficiencyArray
}

//Gives the minimum cost panel option.
func FindMinCostPanel(panelCost []int) string {
	minCost := panelCost[0]
	minCostIDX := 0
	for idx, cost := range panelCost {
		if cost < minCost {
			minCost = cost
			minCostIDX = idx
		}
	}
	minCostPanel := IdxToPanel(minCostIDX)
	return minCostPanel
}

//Finds the brand of solar panel with the highest efficiency.
func FindMostEfficient(efficiencyarray []float64) string {
	mostEfficient := efficiencyarray[0]
	mostEfficientIDX := 0
	for idx, efficiency := range efficiencyarray {
		if efficiency > mostEfficient {
			mostEfficient = efficiency
			mostEfficientIDX = idx
		}
	}
	mostEfficientPanel := IdxToPanel(mostEfficientIDX)
	return mostEfficientPanel
}

//Finds the panel brand with the highest output of solar energy.
func FindMaxOutput(efficiencyarray []float64, cityName string, cityData map[string]City, houseSize float64) string {
	var maxOutput, output float64
	var maxIDX int
	for i := range efficiencyarray {
		output = SolarOutput(cityName, cityData, "horizontal", efficiencyarray[i], houseSize)
		if output > maxOutput {
			maxOutput = output
			maxIDX = i
		}
	}
	maxOutputPanel := IdxToPanel(maxIDX)
	return maxOutputPanel
}

//Error message, error if not nil or less than 0.
func ErrorMessage(err error, input string, variableInput float64) {
	if err != nil { //there was a problem
		fmt.Printf("Error: Number entered for %s was invalid.\n", input)
	} else if variableInput <= 0 {
		fmt.Printf("Error: Number entered for %s was less than zero.\n", input)
	} // else no errors
}
