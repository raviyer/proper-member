package util

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type person struct {
	XMLName   xml.Name `xml:"person"`
	FirstName string   `xml:"firstname"`
	LastName  string   `xml:"lastname"`
	Country   string   `xml:"country"`
	Gender    string   `xml:"gender"`
}
type data struct {
	XMLName xml.Name `xml:"data"`
	Persons []person `xml:"person"`
}

func GetRandomName() (name string, err error) {
	resp, err := http.Get("http://namegenerator.juergenbouche.de/generate.php?c=-1&g=2&n=1&m=xml")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	d := new(data)
	err = decoder.Decode(d)
	name = fmt.Sprintf("%v %v", d.Persons[0].FirstName, d.Persons[0].LastName)
	return
}
