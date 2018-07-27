package main

import (
"github.com/FactomProject/factom"
"fmt"
"time"
"os"
"flag"
"encoding/json"
"net/http"
"io/ioutil"
"crypto/x509"
"crypto/tls"
"bytes"
"strings"
"log"
)

func init() {
	factom.SetFactomdServer("localhost:8088")
}

func main() {
	tagPtr := flag.String("tag", "mainnet-follower", "Name to tag this test with")
	blkTimePtr := flag.Int("blktime", 600, "Seconds per block of the running factomd instance")
	flag.Usage = func() {
		fmt.Printf("Usage: minute-tracker [optional flags]\n\n")
		flag.PrintDefaults()
	}
	flag.CommandLine.Parse(os.Args[1:])

	// Process & check param values
	tag := *tagPtr
	blkTime := *blkTimePtr
	if blkTime < 1 {
		blkTime = 1
	}
	idealMinuteTime := time.Duration(blkTime) * time.Second / 10

	// Create output spreadsheet
	filename := fmt.Sprintf("%s_%s.csv", tag, time.Now().Format("2006-01-02_15:04:05"))
	file, err := os.Create(filename)
	check(err, "Failed to create output file")
	defer file.Close()
	testContext := fmt.Sprintf(
		"Tag:,%s\nBlockTime:,%d\nExpectedMinuteTime:,%f\nNotes:\n\n",
		tag,
		blkTime,
		idealMinuteTime.Seconds(),
	)
	_, err = file.WriteString(testContext)
	check(err, "Failed to write test context to file")
	_, err = file.WriteString("Block Number,Minute Number,Minute Length (s)\n")
	check(err, "Failed to write headers to file")

	// Start checking block minute length
	previous, err := getCurrentMinute()
	check(err, "Failed to request current minute")
	current := previous
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		current, err = getCurrentMinute()
		check(err, "Failed to request current minute")
		if previous.Minute != current.Minute {
			// New minute started. Calculate summary of previous minute, then write it to the spreadsheet
			previousMinuteTime := time.Duration(current.MinuteStartTime - previous.MinuteStartTime)
			row := fmt.Sprintf("%d,%d,%f\n", previous.BlockHeight, previous.Minute, previousMinuteTime.Seconds())
			go func() {
				_, err = file.WriteString(row)
				check(err, "Failed to write minute to file")
			}()
			previous = current

			fmt.Printf("Started minute %d in block %d\n", current.Minute, current.BlockHeight)
		}
	}
}

type CurrentMinuteResponse struct {
	Minute int `json:"Minute"`
	MinuteStartTime int64 `json:"MinuteStartTime"`
	BlockHeight uint32 `json:"BlockHeight"`
}

// getCurrentMinute requests information about the current minute of the block being built
func getCurrentMinute() (*CurrentMinuteResponse, error) {
	req := factom.NewJSON2Request("current-minute", factom.APICounter(), nil)
	resp, err := sendFactomdDebugRequest(req)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	currentMinute := new(CurrentMinuteResponse)
	if err := json.Unmarshal(resp.JSONResult(), currentMinute); err != nil {
		return nil, err
	}
	return currentMinute, nil
}

// sendFactomdDebugRequest sends a json object to factomd's debug api rather than v2.
// It's adapted from func factomdRequest(...) in FactomProject/factom/jsonrpc.go
func sendFactomdDebugRequest(req *factom.JSON2Request) (*factom.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	factomdTls, factomdCertPath := factom.GetFactomdEncryption()

	var client *http.Client
	var httpx string

	if factomdTls == true {
		caCert, err := ioutil.ReadFile(factomdCertPath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}

		client = &http.Client{Transport: tr}
		httpx = "https"

	} else {
		client = &http.Client{}
		httpx = "http"
	}
	re, err := http.NewRequest("POST",
		fmt.Sprintf("%s://%s/debug", httpx, factom.FactomdServer()),
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	user, pass := factom.GetFactomdRpcConfig()
	re.SetBasicAuth(user, pass)
	re.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(re)
	if err != nil {
		errs := fmt.Sprintf("%s", err)
		if strings.Contains(errs, "\\x15\\x03\\x01\\x00\\x02\\x02\\x16") {
			err = fmt.Errorf("Factomd API connection is encrypted. Please specify -factomdtls=true and -factomdcert=factomdAPIpub.cert (%v)", err.Error())
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Factomd username/password incorrect.  Edit factomd.conf or\ncall factom-cli with -factomduser=<user> -factomdpassword=<pass>")
	}
	r := factom.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}

func check(err error, message string) {
	if err != nil {
		log.Fatal(message, err)
	}
}