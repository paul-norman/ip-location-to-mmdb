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
	"compress/gzip"
	"io"
)

type stringSlice []string

func (i *stringSlice) String() string {
    return fmt.Sprintf("%v", *i)
}

func (i *stringSlice) Set(value string) error {
    *i = append(*i, value)
    return nil
}

type UserInput struct {
	inputPath	[]string
	outputPath	string
	fileType	string
	ipVersion	int
	recordSize	int
	compress	int
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
	mmDbWriter := initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	for _, inputFile := range userInput.inputPath {
		csvFile, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer csvFile.Close()

		csvFileReader := csv.NewReader(csvFile)

		fmt.Printf("Loading MMDB Country data from: %s\n", inputFile)

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
	}

	saveMmdbData(mmDbWriter, userInput.outputPath, userInput.compress)
}

func loadASNs(userInput UserInput) {
	mmDbWriter := initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	for _, inputFile := range userInput.inputPath {
		csvFile, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer csvFile.Close()

		csvFileReader := csv.NewReader(csvFile)

		fmt.Printf("Loading MMDB ASN data from: %s\n", inputFile)

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
	}

	saveMmdbData(mmDbWriter, userInput.outputPath, userInput.compress)
}

func loadCities(userInput UserInput) {
	mmDbWriter := initMmdbWriter(userInput.fileType, userInput.ipVersion, userInput.recordSize)

	for _, inputFile := range userInput.inputPath {
		csvFile, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer csvFile.Close()

		csvFileReader := csv.NewReader(csvFile)

		fmt.Printf("Loading MMDB City data from: %s\n", inputFile)

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
				"state1":		mmdbtype.String(record[3]),
				"state2":		mmdbtype.String(record[4]),
			}
			if record[7] != "" && record[8] != "" {
				lat, _ := strconv.ParseFloat(record[7], 32)
				lon, _ := strconv.ParseFloat(record[8], 32)
				mmdbRowData["latitude"] = mmdbtype.Float32(lat)
				mmdbRowData["longitude"] = mmdbtype.Float32(lon)
			}

			ipRanges := findIPRanges(record[0], record[1])
			for _, ipRange := range ipRanges {
				err := mmDbWriter.Insert(ipRange, mmdbRowData)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	saveMmdbData(mmDbWriter, userInput.outputPath, userInput.compress)
}

func initMmdbWriter(fileType string, ipVersion int, recordSize int) *mmdbwriter.Tree {
	var options mmdbwriter.Options

	if ipVersion > 0 {
		options = mmdbwriter.Options{
			DatabaseType:				fileType + " ipv" + strconv.Itoa(ipVersion),
			RecordSize:					recordSize,
			IPVersion:					ipVersion,
			IncludeReservedNetworks:	true,
			DisableIPv4Aliasing:		true,
		}
	} else {
		options = mmdbwriter.Options{
			DatabaseType:				fileType + " ipvAll",
			RecordSize:					recordSize,
			IncludeReservedNetworks:	true,
			DisableIPv4Aliasing:		true,
		}
	}

	mmDbWriter, err := mmdbwriter.New(options)
	if err != nil {
		panic(err)
	}

	return mmDbWriter
}

func saveMmdbData(mmDbWriter *mmdbwriter.Tree, filePath string, compress int) {
	fileHandle, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing MMDB file: " + filePath)
	_, err = mmDbWriter.WriteTo(fileHandle)
	if err != nil {
		panic(err)
	}

	fileHandle.Close()

	if compress == 1 {
		err = compressFile(filePath)
		if err != nil {
			panic(err)
		}

		fmt.Println("Removing uncompressed file: " + filePath)
		err = os.Remove(filePath)
		if err != nil {
			panic(err)
		}
	}
}

func compressFile(filePath string) error {
	fmt.Println("Compressing MMDB file: " + filePath + " to: " + filePath + ".gz")

	reader, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer reader.Close()

    writer, err := os.Create(filePath + ".gz")
    if err != nil {
		return err
	}
	defer writer.Close()

	gzWriter := gzip.NewWriter(writer)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, reader)
    if err != nil {
        return err
    }

	return nil
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
	var input []string;
	var input1 stringSlice;
	var input2 stringSlice;

	flag.Var(&input1, "i", "The input CSV file path")
	flag.Var(&input2, "input", "The input CSV file path")

	output1		:= flag.String("o", "", "The output MMDB file path")
	output2		:= flag.String("output", "", "The output MMDB file path")
	type1		:= flag.String("t", "", "The type of file to process (country, asn or city)")
	type2		:= flag.String("type", "", "The type of file to process (country, asn or city)")
	ipv			:= flag.Int("ipv", 0, "The IP Version of the data file")
	recordSize1 := flag.Int("record_size", 0, "The record size of the MMDB file")
	recordSize2 := flag.Int("r", 0, "The record size of the MMDB file")
	compress1	:= flag.Int("compress", 0, "Should the MMDB file be compressed?")
	compress2	:= flag.Int("z", 0, "Should the MMDB file be compressed?")
	flag.Parse()

	if len(input1) > 0 {
		input = append(input, input1...)
	}
	if len(input2) > 0 {
		input = append(input, input2...)
	}

	if len(input) == 0 {
		panic("an input CSV file is required")
	}

	test := strings.ToLower(input[0])

	var output string
	if len(*output1) > 0 {
		output = *output1
	} else if len(*output2) > 0 {
		output = *output2
	} else {
		output = strings.Replace(input[0], ".csv", ".mmdb", 1)

		if len(input) > 1 {
			output = strings.Replace(output, "-ipv4", "-ipvAll", 1)
			output = strings.Replace(output, "-ipv6", "-ipvAll", 1)
		}
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
		panic("a valid file type is required: `country`, `asn` or `city`")
	}

	var ipVersion int
	if *ipv > 0 {
		ipVersion = *ipv
	} else {
		if len(input) > 1 {
			ipVersion = 0
		} else if strings.Contains(test, "ipv4") {
			ipVersion = 4
		} else if strings.Contains(test, "ipv6") {
			ipVersion = 6
		}
	}

	if len(input) == 1 && ipVersion != 4 && ipVersion != 6 {
		panic("a valid IP Version is required: `4` or `6`")
	}

	var recordSize int
	if *recordSize1 > 0 {
		recordSize = *recordSize1
	} else if *recordSize2 > 0 {
		recordSize = *recordSize2
	} else {
		switch fileType {
			case "country":	recordSize = 24
			case "asn":		recordSize = 24
			case "city":	recordSize = 28
		}
	}

	if recordSize != 24 && recordSize != 28 && recordSize != 32 {
		panic("a valid MMDB record size is required: `24`, `28` or `32`")
	}

	var compress int
	if *compress1 > 0 {
		compress = *compress1
	} else if *compress2 > 0 {
		compress = *compress2
	}

	if compress != 0 && compress != 1 {
		panic("a valid compression flag is either: `0` or `1`")
	}

	return UserInput{ input, output, fileType, ipVersion, recordSize, compress }
}