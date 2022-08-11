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
	"time"
)

const filename string = "./data.json"
const timeLayout string = "15:04:05"

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

func (t Train) IsEmpty() bool {
	return reflect.DeepEqual(t, Train{})
}

type queryParam struct {
	dep, arr int
	sort     string
}

func (qp *queryParam) validateInput(depStation, arrStation, criteria string) error {
	if depStation == "" {
		return emptyDepStationErr
	}
	depID, err := strconv.Atoi(depStation)
	if err != nil {
		return badDepStationErr
	}
	qp.dep = depID

	if arrStation == "" {
		return emptyArrStationErr
	}
	arrID, err := strconv.Atoi(arrStation)
	if err != nil {
		return badArrStationErr
	}
	qp.arr = arrID

	switch criteria {
	case "price", "arrival-time", "departure-time":
		qp.sort = criteria
	default:
		return unsuportCriteriaErr
	}

	return nil
}

func parseJSON() (Trains, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) (err error) {
		err = jsonFile.Close()
		if err != nil {
			return err
		}
		return nil
	}(jsonFile)

	decoder := json.NewDecoder(jsonFile)

	var trains Trains

	for {
		var sliceAny []interface{}
		var train Train

		if err := decoder.Decode(&sliceAny); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		for i := range sliceAny {
			jsVal, ok := sliceAny[i].(map[string]interface{})
			if ok {
				trainID, ok := jsVal["trainId"].(float64)
				if ok {
					train.TrainID = int(trainID)
				}
				depStationID, ok := jsVal["departureStationId"].(float64)
				if ok {
					train.DepartureStationID = int(depStationID)
				}
				arrStationID, ok := jsVal["arrivalStationId"].(float64)
				if ok {
					train.ArrivalStationID = int(arrStationID)
				}
				price, ok := jsVal["price"].(float64)
				if ok {
					train.Price = float32(price)
				}
				timeArr, ok := jsVal["arrivalTime"].(string)
				if ok {
					tArr, err := time.Parse(timeLayout, timeArr)
					if err != nil {
						return nil, err
					}
					train.ArrivalTime = tArr
				}
				timeDep, ok := jsVal["departureTime"].(string)
				if ok {
					tDep, err := time.Parse(timeLayout, timeDep)
					if err != nil {
						return nil, err
					}
					train.DepartureTime = tDep
				}

				trains = append(trains, train)

			}

		}
	}
	return trains, nil
}

func main() {

	var depStation string
	var arrStation string
	var criteria string

	// ... запит даних від користувача
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter the number of the departure station: ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}
	depStation = scanner.Text()

	fmt.Println("Enter the arrival station number: ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}
	arrStation = scanner.Text()

	fmt.Println("Specify the search criteria (price,arrival-time,departure-time): ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}
	criteria = scanner.Text()

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
	// ... код
	var qp = queryParam{0, 0, ""}
	var searchedTrains Trains
	var result Trains

	if err := qp.validateInput(depStation, arrStation, criteria); err != nil {
		return nil, err
	}

	trains, err := parseJSON()
	if err != nil {
		return nil, err
	}

	for _, train := range trains {
		if (train.DepartureStationID == qp.dep) && (train.ArrivalStationID == qp.arr) {
			searchedTrains = append(searchedTrains, train)
		}
	}

	if len(searchedTrains) == 0 {
		return nil, nil
	}

	switch qp.sort {
	case "price":
		sort.SliceStable(searchedTrains, searchedTrains.PriceAsc)
	case "arrival-time":
		sort.SliceStable(searchedTrains, searchedTrains.ArrivalTimeAsc)
	case "departure-time":
		sort.SliceStable(searchedTrains, searchedTrains.DepartureTimeAsc)
	}

	for i := 0; i < 3; i++ {
		train := searchedTrains[i]
		if train.IsEmpty() {
			continue
		}
		result = append(result, train)
	}

	return result, nil // маєте повернути правильні значення
}
