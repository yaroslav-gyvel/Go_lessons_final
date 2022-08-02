package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
)

type Trains []Train

type Train struct {
	TrainID            int
	DepartureStationID int
	ArrivalStationID   int
	Price              float32
	ArrivalTime        time.Time
	DepartureTime      time.Time
}

var emptyDepStationErr = errors.New("empty departure station")
var emptyArrStationErr = errors.New("empty arrival station")
var badDepStationErr = errors.New("bad departure station input")
var badArrStationErr = errors.New("bad arrival station input")
var unsuportCriteriaErr = errors.New("unsupported criteria")

var criterias = []string{"price", "arrival-time", "departure-time"} // unsorted

func parsingJSON() (Trains, error) {

	filename := "./data.json"
	jsonFile, _ := os.Open(filename)
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			fmt.Printf("json file is not closed, error: %v", err)
			return
		}
	}(jsonFile)

	decoder := json.NewDecoder(jsonFile)

	var trains Trains

	for {
		var sliceAny []interface{}
		var train Train

		err := decoder.Decode(&sliceAny)

		if err == io.EOF {
			break
		}

		for i := range sliceAny {
			jsVal := sliceAny[i].(map[string]interface{})

			train.TrainID = int(jsVal["trainId"].(float64))
			train.DepartureStationID = int(jsVal["departureStationId"].(float64))
			train.ArrivalStationID = int(jsVal["arrivalStationId"].(float64))
			train.Price = float32(jsVal["price"].(float64))

			tArr, _ := time.Parse("15:04:05", jsVal["arrivalTime"].(string))
			tDep, _ := time.Parse("15:04:05", jsVal["departureTime"].(string))
			train.ArrivalTime = tArr
			train.DepartureTime = tDep

			trains = append(trains, train)
		}
	}

	return trains, nil
}

func validateDepStation(depStation string) error {

	if depStation == "" {
		return emptyDepStationErr
	}

	_, err := strconv.Atoi(depStation)
	if err != nil {
		return badDepStationErr
	}

	return nil
}
func validateArrStation(arrStation string) error {

	if arrStation == "" {
		return emptyArrStationErr
	}
	_, err := strconv.Atoi(arrStation)
	if err != nil {
		return badArrStationErr
	}

	return nil
}
func validateCriteria(criteria string) error {
	if criteria == "" {
		return unsuportCriteriaErr
	}
	sort.Strings(criterias)
	if sort.StringsAreSorted(criterias) {
		i := sort.SearchStrings(criterias, criteria)
		if i < len(criterias) && (criterias)[i] == criteria {
			return nil
		}
	}

	return unsuportCriteriaErr
}

func main() {

	var depStation string
	var arrStation string
	var criteria string

	// ... запит даних від користувача

	fmt.Println("Enter the number of the departure station: ")
	_, _ = fmt.Scanln(&depStation)

	fmt.Println("Enter the arrival station number: ")
	_, _ = fmt.Scanln(&arrStation)

	fmt.Println("Specify the search criteria (price,arrival-time,departure-time): ")
	_, _ = fmt.Scanln(&criteria)

	result, err := FindTrains(depStation, arrStation, criteria)

	//	... обробка помилки

	if errors.Is(err, emptyDepStationErr) {
		fmt.Printf("%v\n", emptyDepStationErr)
	}
	if errors.Is(err, emptyArrStationErr) {
		fmt.Printf("%v\n", emptyArrStationErr)
	}
	if errors.Is(err, unsuportCriteriaErr) {
		fmt.Printf("%v\n", unsuportCriteriaErr)
	}
	if errors.Is(err, badDepStationErr) {
		fmt.Printf("%v\n", badDepStationErr)
	}
	if errors.Is(err, badArrStationErr) {
		fmt.Printf("%v\n", badArrStationErr)
	}

	//	... друк result

	for _, train := range result {
		fmt.Printf("{TrainID: %v,"+
			" DepartureStationID: %v,"+
			" ArrivalStationID: %v,"+
			" Price: %v,"+
			" ArrivalTime: time.Date(%v, time.%v, %v, %v, %v, %v, %v, time.%v),"+
			" DepartureTime: time.Date(%v, time.%v, %v, %v, %v, %v, %v, time.%v)}\n",
			train.TrainID,
			train.DepartureStationID,
			train.ArrivalStationID,
			train.Price,
			train.ArrivalTime.Year(),
			train.ArrivalTime.Month().String(),
			train.ArrivalTime.Day(),
			train.ArrivalTime.Hour(),
			train.ArrivalTime.Minute(),
			train.ArrivalTime.Second(),
			0,
			time.UTC,
			train.DepartureTime.Year(),
			train.DepartureTime.Month().String(),
			train.DepartureTime.Day(),
			train.DepartureTime.Hour(),
			train.DepartureTime.Minute(),
			train.DepartureTime.Second(),
			0,
			time.UTC)
	}
}

func FindTrains(depStation, arrStation, criteria string) (Trains, error) {
	// ... код

	var trains Trains
	var searchedTrains Trains

	trains, _ = parsingJSON()

	if err := validateDepStation(depStation); err != nil {
		return nil, err
	}
	if err := validateArrStation(arrStation); err != nil {
		return nil, err
	}
	if err := validateCriteria(criteria); err != nil {
		return nil, err
	}

	depID, _ := strconv.Atoi(depStation)
	arrID, _ := strconv.Atoi(arrStation)

	for _, train := range trains {
		if (train.DepartureStationID == depID) && (train.ArrivalStationID == arrID) {
			searchedTrains = append(searchedTrains, train)
		}
	}

	if len(searchedTrains) == 0 {
		return nil, nil
	}

	switch criteria {
	case "price":
		sort.SliceStable(searchedTrains, func(i, j int) bool { return searchedTrains[i].Price < searchedTrains[j].Price })
	case "arrival-time":
		sort.SliceStable(searchedTrains, func(i, j int) bool { return searchedTrains[i].ArrivalTime.Before(searchedTrains[j].ArrivalTime) })
	case "departure-time":
		sort.SliceStable(searchedTrains, func(i, j int) bool { return searchedTrains[i].DepartureTime.Before(searchedTrains[j].DepartureTime) })
	}

	result := Trains{searchedTrains[0], searchedTrains[1], searchedTrains[2]}

	return result, nil // маєте повернути правильні значення
}
