package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/grokify/gotilla/fmt/fmtutil"
)

var (
	IMFFileXML  = "API_BN.GSR.GNFS.CD_DS2_en_xml_v2.xml"
	IMFFileJSON = "API_BN.GSR.GNFS.CD_DS2_en_xml_v2.json"
)

type IMFXml struct {
	Data Data `xml:"data"`
}

type Data struct {
	Records []Record `xml:"record"`
}

type Record struct {
	Fields []Field `xml:"field"`
}

type Field struct {
	Name  string `xml:"name,attr"`
	Key   string `xml:"key,attr"`
	Value string `xml:",innerxml"`
}

type IMFData struct {
	Records []CountryYearBOP
}

type CountryYearBOP struct {
	CountryOrAreaName      string
	CountryISOAlpha3       string
	Year                   int16
	BalanceOfPayments      float64
	BalanceOfPaymentsKnown bool
	IMFReference           string `json:"-"`
}

func ParseXMLRecord(xrec Record) (CountryYearBOP, error) {
	bop := CountryYearBOP{}
	for _, field := range xrec.Fields {
		switch field.Name {
		case "Item":
			if field.Value != "Net trade in goods and services (BoP, current US$)" {
				return bop, fmt.Errorf("Item is not Net Trade")
			}
			bop.IMFReference = field.Key
		case "Country or Area":
			bop.CountryOrAreaName = field.Value
			bop.CountryISOAlpha3 = field.Key
		case "Year":
			year, err := strconv.Atoi(field.Value)
			if err != nil {
				return bop, err
			}
			bop.Year = int16(year)
		case "Value":
			if len(field.Value) > 0 {
				f, err := strconv.ParseFloat(field.Value, 64)
				if err != nil {
					return bop, err
				}
				bop.BalanceOfPayments = f
				bop.BalanceOfPaymentsKnown = true
			} else {
				bop.BalanceOfPaymentsKnown = false
			}
		}
	}

	return bop, nil
}

func WriteJSONIndent(filepath string, data interface{}, perm os.FileMode, prefix, indent string) error {
	bytes, err := json.MarshalIndent(data, prefix, indent)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, bytes, perm)
}

func main() {
	var imfxml = IMFXml{}
	bytes, err := ioutil.ReadFile(IMFFileXML)
	if err != nil {
		panic(err)
	}
	err = xml.Unmarshal(bytes, &imfxml)
	if err != nil {
		panic(err)
	}
	fmtutil.PrintJSON(imfxml)

	imfdata := IMFData{Records: []CountryYearBOP{}}

	for _, rec := range imfxml.Data.Records {
		bop, err := ParseXMLRecord(rec)
		if err != nil {
			panic(err)
		}
		imfdata.Records = append(imfdata.Records, bop)
	}

	fmtutil.PrintJSON(imfdata)

	err = WriteJSONIndent(IMFFileJSON, imfdata, 644, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("DONE")
}
