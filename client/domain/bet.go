package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Bet struct {
	Name      string
	LastName  string
	Document  string
	BirthDate string
	Number    int
}

// SerializeToCSV convierte Bet a bytes CSV "nombre;apellido;documento;nacimiento;numero"
func (b Bet) SerializeToCSV() []byte {
	return []byte(fmt.Sprintf("%s;%s;%s;%s;%d",
		b.Name, b.LastName, b.Document, b.BirthDate, b.Number))
}

// DeserializeBetCSV convierte payload CSV a Bet.
func DeserializeBetCSV(data []byte) (Bet, error) {
	fields := strings.SplitN(string(data), ";", 5)
	if len(fields) != 5 {
		return Bet{}, errors.New("invalid bet csv format")
	}

	number, err := strconv.Atoi(fields[4])
	if err != nil {
		return Bet{}, err
	}

	return Bet{
		Name:      fields[0],
		LastName:  fields[1],
		Document:  fields[2],
		BirthDate: fields[3],
		Number:    number,
	}, nil
}
