package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/seancfoley/ipaddress-go/ipaddr"
)

type UserInput struct {
	inputPath	string
	outputPath	string
	fileType	string
	ipVersion	int
	recordSize	int
}

func main() {
	userInput := handleUserInput()

	switch userInput.fileType {
		case "country":	loadCountries(userInput)
		case "asn":		loadASNs(userInput)
		case "city":	loadCities(userInput)
	}
}

func loadCountries(userInput UserInput) {
	csvFile, err := os.Open(userInput.inputPath)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	csvFileReader	:= csv.NewReader(csvFile)
	mmDbWriter		:= initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	fmt.Printf("Loading MMDB Country ipv%d data from: %s\n", userInput.ipVersion, userInput.inputPath)

	for {
		record, err := csvFileReader.Read()
		if err != nil {
			break
		}

		mmdbRowData := mmdbtype.Map{
			"country_code": mmdbtype.String(record[2]),
		}

		ipRanges := findIPRanges(record[0], record[1])
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, mmdbRowData)
			if err != nil {
				panic(err)
			}
		}
	}

	saveMmdbData(mmDbWriter, userInput.outputPath)
}

func loadASNs(userInput UserInput) {
	csvFile, err := os.Open(userInput.inputPath)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	csvFileReader	:= csv.NewReader(csvFile)
	mmDbWriter		:= initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	fmt.Printf("Loading MMDB ASN ipv%d data from: %s\n", userInput.ipVersion, userInput.inputPath)

	for {
		record, err := csvFileReader.Read()
		if err != nil {
			break
		}

		number, _ := strconv.Atoi(record[2])

		mmdbRowData := mmdbtype.Map{
			"autonomous_system_number":			mmdbtype.Uint32(number),
			"autonomous_system_organization":	mmdbtype.String(record[3]),
		}

		ipRanges := findIPRanges(record[0], record[1])
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, mmdbRowData)
			if err != nil {
				panic(err)
			}
		}
	}

	saveMmdbData(mmDbWriter, userInput.outputPath)
}

func loadCities(userInput UserInput) {
	csvFile, err := os.Open(userInput.inputPath)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	csvFileReader	:= csv.NewReader(csvFile)
	mmDbWriter		:= initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	fmt.Printf("Loading MMDB City ipv%d data from: %s\n", userInput.ipVersion, userInput.inputPath)

	for {
		record, err := csvFileReader.Read()
		if err != nil {
			break
		}
		//lat, _ := strconv.ParseFloat(record[7], 64)
		//lon, _ := strconv.ParseFloat(record[8], 64)

		mmdbRowData := mmdbtype.Map{
			"city":			mmdbtype.String(record[5]),
			"postcode":		mmdbtype.String(record[6]),
			"timezone":		mmdbtype.String(record[9]),
			"country_code":	mmdbtype.String(record[2]),
			"latitude":		mmdbtype.String(record[7]),
			"longitude":	mmdbtype.String(record[8]),
			"state1":		mmdbtype.String(record[3]),
			"state2":		mmdbtype.String(record[4]),
		}

		ipRanges := findIPRanges(record[0], record[1])
		for _, ipRange := range ipRanges {
			err := mmDbWriter.Insert(ipRange, mmdbRowData)
			if err != nil {
				panic(err)
			}
		}
	}

	saveMmdbData(mmDbWriter, userInput.outputPath)
}

func initMmdbWriter(fileType string, ipVersion int, recordSize int) *mmdbwriter.Tree {
	mmDbWriter, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType:				fileType + " ipv" + strconv.Itoa(ipVersion),
			RecordSize:					recordSize,
			IPVersion:					ipVersion,
			IncludeReservedNetworks:	true,
			DisableIPv4Aliasing:		true,
		},
	)
	if err != nil {
		panic(err)
	}

	return mmDbWriter
}

func saveMmdbData(mmDbWriter *mmdbwriter.Tree, filePath string) {
	fileHandle, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer fileHandle.Close()

	fmt.Println("Writing MMDB file: " + filePath)
	_, err = mmDbWriter.WriteTo(fileHandle)
	if err != nil {
		panic(err)
	}
}

func findIPRanges(ipRangeStart string, ipRangeEnd string) []*net.IPNet {
	ipStart	:= ipaddr.NewIPAddressString(ipRangeStart)
	ipEnd	:= ipaddr.NewIPAddressString(ipRangeEnd)

	addressStart	:= ipStart.GetAddress()
	addressEnd		:= ipEnd.GetAddress()

	ipRange		:= addressStart.SpanWithRange(addressEnd)
	rangeSlice	:= ipRange.SpanWithPrefixBlocks()

	var ipNets []*net.IPNet
	for _, val := range rangeSlice {
		_, network, err := net.ParseCIDR(val.String())
		if err != nil {
			panic(err)
		}

		ipNets = append(ipNets, network)
	}

	return ipNets
}

func handleUserInput() UserInput {
	input1		:= flag.String("i", "", "The input CSV file path")
	input2		:= flag.String("input", "", "The input CSV file path")
	output1		:= flag.String("o", "", "The output MMDB file path")
	output2		:= flag.String("output", "", "The output MMDB file path")
	type1		:= flag.String("t", "", "The type of file to process (country, asn or city)")
	type2		:= flag.String("type", "", "The type of file to process (country, asn or city)")
	ipv			:= flag.Int("ipv", 0, "The IP Version of the data file")
	recordSize1 := flag.Int("record_size", 0, "The record size of the MMDB file")
	recordSize2 := flag.Int("r", 0, "The record size of the MMDB file")
	flag.Parse()

	var input string
	if len(*input1) > 0 {
		input = *input1
	} else if len(*input2) > 0 {
		input = *input2
	}

	if len(input) == 0 {
		panic("an input CSV file is required")
	}

	test := strings.ToLower(input)

	var output string
	if len(*output1) > 0 {
		output = *output1
	} else if len(*output2) > 0 {
		output = *output2
	} else {
		output = strings.Replace(input, ".csv", ".mmdb", 1)
	}

	var fileType string
	if len(*type1) > 0 {
		fileType = *type1
	} else if len(*type2) > 0 {
		fileType = *type2
	} else if len(test) > 0 {
		if strings.Contains(test, "country") || strings.Contains(test, "countries") {
			fileType = "country"
		} else if strings.Contains(test, "asn") {
			fileType = "asn"
		} else if strings.Contains(test, "city") || strings.Contains(test, "cities") {
			fileType = "city"
		}
	}
	fileType = strings.ToLower(fileType)

	allowedTypes := []string{ "country", "asn", "city" }
	if !slices.Contains(allowedTypes, fileType) {
		panic("the file type must be: `country`, `asn` or `city`")
	}

	var ipVersion int
	if *ipv > 0 {
		ipVersion = *ipv
	} else {
		if strings.Contains(test, "ipv4") {
			ipVersion = 4
		} else if strings.Contains(test, "ipv6") {
			ipVersion = 6
		}
	}

	if ipVersion != 4 && ipVersion != 6 {
		panic("an IP Version is required: `4` or `6`")
	}

	var recordSize int
	if *recordSize1 > 0 {
		recordSize = *recordSize1
	} else if *recordSize2 > 0 {
		recordSize = *recordSize2
	} else if len(test) > 0 {
		if strings.Contains(test, "country") || strings.Contains(test, "countries") {
			recordSize = 24
		} else if strings.Contains(test, "asn") {
			recordSize = 24
		} else if strings.Contains(test, "city") || strings.Contains(test, "cities") {
			recordSize = 28
		}
	}

	if recordSize != 24 && recordSize != 28 && recordSize != 32 {
		panic("an MMDB record size is required: `24`, `28` or `32`")
	}

	return UserInput{ input, output, fileType, ipVersion, recordSize }
}