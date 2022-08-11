package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const filename string = "./data.json"
const timeLayout string = "15:04:05"
const resultCounter int = 3
const sortPrice string = "price"
const sortDepTime string = "departure-time"
const sortArrTime string = "arrival-time"

var emptyDepStationErr = errors.New("empty departure station")
var emptyArrStationErr = errors.New("empty arrival station")
var badDepStationErr = errors.New("bad departure station input")
var badArrStationErr = errors.New("bad arrival station input")
var unsuportCriteriaErr = errors.New("unsupported criteria")

type Trains []Train

func (t Trains) PriceAsc(i, j int) bool {
	return t[i].Price < t[j].Price
}
func (t Trains) ArrivalTimeAsc(i, j int) bool {
	return t[i].ArrivalTime.Before(t[j].ArrivalTime)
}
func (t Trains) DepartureTimeAsc(i, j int) bool {
	return t[i].DepartureTime.Before(t[j].DepartureTime)
}

type Train struct {
	TrainID            int
	DepartureStationID int
	ArrivalStationID   int
	Price              float32
	ArrivalTime        time.Time
	DepartureTime      time.Time
}

type queryParam struct {
	departureStation, arrivalStation int
	sorting                          string
}

func (q queryParam) IsEmpty() bool {
	return reflect.DeepEqual(q, queryParam{departureStation: 0, arrivalStation: 0, sorting: ""})
}

func validateInput(depStation, arrStation, criteria string, qp queryParam) (queryParam, error) {

	if depStation == "" {
		return qp, emptyDepStationErr
	}

	depID, err := strconv.Atoi(depStation)
	if err != nil {
		return qp, badDepStationErr
	}
	qp.departureStation = depID

	if arrStation == "" {
		return qp, emptyArrStationErr
	}

	arrID, err := strconv.Atoi(arrStation)
	if err != nil {
		return qp, badArrStationErr
	}
	qp.arrivalStation = arrID

	switch criteria {
	case sortPrice, sortArrTime, sortDepTime:
		qp.sorting = criteria
	default:
		return qp, unsuportCriteriaErr
	}

	return qp, nil
}

func parseJSON() (Trains, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s' file in OS: %w", filename, err)
	}

	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
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
		if err != nil {
			return nil, err
		}

		for i := range sliceAny {
			jsVal, ok := sliceAny[i].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("failed casting json file '%s' file in OS: %w", filename, err)
			}

			trainID, ok := jsVal["trainId"].(float64)
			if !ok {
				return nil, fmt.Errorf("failed casting value trainID: %w", err)
			}
			train.TrainID = int(trainID)

			depStationID, ok := jsVal["departureStationId"].(float64)
			if !ok {
				return nil, fmt.Errorf("failed casting value depStationID: %w", err)
			}
			train.DepartureStationID = int(depStationID)

			arrStationID, ok := jsVal["arrivalStationId"].(float64)
			if !ok {
				return nil, fmt.Errorf("failed casting value arrStationID: %w", err)
			}
			train.ArrivalStationID = int(arrStationID)

			price, ok := jsVal["price"].(float64)
			if !ok {
				return nil, fmt.Errorf("failed casting value price: %w", err)
			}
			train.Price = float32(price)

			timeArr, ok := jsVal["arrivalTime"].(string)
			if !ok {
				return nil, fmt.Errorf("failed casting value arrivalTime: %w", err)
			}
			tArr, err := time.Parse(timeLayout, timeArr)
			if err != nil {
				return nil, fmt.Errorf("failed parsing value arrivalTime: %w", err)
			}
			train.ArrivalTime = tArr

			timeDep, ok := jsVal["departureTime"].(string)
			if !ok {
				return nil, fmt.Errorf("failed casting value departureTime: %w", err)
			}
			tDep, err := time.Parse(timeLayout, timeDep)
			if err != nil {
				return nil, fmt.Errorf("failed parsing value departureTime: %w", err)
			}
			train.DepartureTime = tDep

			trains = append(trains, train)

		}
	}

	return trains, nil
}

func main() {
	var depStation string
	var arrStation string
	var criteria string
	// ... запит даних від користувача
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter the number of the departure station: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("input error: %s\n", err)
	}
	depStation = strings.Replace(input, "\r\n", "", -1)

	fmt.Println("Enter the arrival station number: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("input error: %s\n", err)
	}
	arrStation = strings.Replace(input, "\r\n", "", -1)

	fmt.Println("Specify the search criteria (price,arrival-time,departure-time): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("input error: %s\n", err)
	}
	criteria = strings.Replace(input, "\r\n", "", -1)

	result, err := FindTrains(depStation, arrStation, criteria)

	//	... обробка помилки
	if err != nil {
		fmt.Printf("%v\n", err)
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
	var searchedTrains Trains
	var result Trains

	qp, err := validateInput(depStation, arrStation, criteria, queryParam{0, 0, ""})
	if err != nil {
		return nil, err //для тестів залишено; fmt.Errorf("failed validation input value: %w", err)
	}
	if qp.IsEmpty() {
		return nil, errors.New("not valid input parameters")
	}

	trains, err := parseJSON()
	if err != nil {
		return nil, fmt.Errorf("failed parsing json: %w", err)
	}

	for _, train := range trains {
		if (train.DepartureStationID == qp.departureStation) && (train.ArrivalStationID == qp.arrivalStation) {
			searchedTrains = append(searchedTrains, train)
		}
	}

	switch qp.sorting {
	case sortPrice:
		sort.SliceStable(searchedTrains, searchedTrains.PriceAsc)
	case sortArrTime:
		sort.SliceStable(searchedTrains, searchedTrains.ArrivalTimeAsc)
	case sortDepTime:
		sort.SliceStable(searchedTrains, searchedTrains.DepartureTimeAsc)
	}

	if len(searchedTrains) == 0 {
		return nil, nil
	}
	if len(searchedTrains) >= resultCounter {
		return searchedTrains[:resultCounter], nil
	}

	return result, nil // маєте повернути правильні значення
}
