package handler

import "fmt"

type Airline struct {
	ID   int32
	name string
}

func (a *Airline) Reard(id int32) (string, error) {
	var airlineModel []Airline
	for _, airline := range airlineModel {
		if airline.ID == id {
			return airline.name, nil
		}
	}
	return "", fmt.Errorf("no airline with %d ID", id)
}
