package main

import (
	"github.com/FactomProject/factom"
	"fmt"
	"time"
	"github.com/FactomProject/factomd/common/primitives/random"
	"math/rand"
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
	factom.SetWalletServer("localhost:8089")
}

func main() {
	branchNamePtr := flag.String("branchname", "develop", "Name of current branch being tested")
	blkTimePtr := flag.Int("blktime", 600, "Seconds per block of the running factomd instance")
	throttleIntervalPtr := flag.Int("throttleinterval", 3, "Number of minutes between tps throttling decisions")
	initialTPSPtr := flag.Int("initialtps", 0, "Number of transactions per second to send on the first load (default 0)")
	ecPubPtr := flag.String("ecpub", "", "Public Key used to fetch the EC Address from the running factom-walletd instance (no default, must be provided)")
	flag.Usage = func() {
		fmt.Printf("Usage: throughput-tester -ecpub=<EC-PUBLIC-KEY> [optional flags]\n\n")
		flag.PrintDefaults()
	}
	flag.CommandLine.Parse(os.Args[1:])

	// Process & check param values
	branchName := *branchNamePtr
	blkTime := *blkTimePtr
	if blkTime < 1 {
		blkTime = 1
	}
	idealMinuteTime := time.Duration(blkTime) * time.Second / 10
	throttleInterval := *throttleIntervalPtr
	if throttleInterval < 1 {
		throttleInterval = 1
	}
	initialTPS := *initialTPSPtr
	if initialTPS < 0 {
		initialTPS = 0
	}
	ecAddress , err := factom.FetchECAddress(*ecPubPtr)
	check(err, "Failed to fetch EC Address")

	testName := fmt.Sprintf("%s_%s", branchName, time.Now().Format("2006-01-02_15:04"))

	// Create output spreadsheet
	file, err := os.Create(fmt.Sprintf("%s.csv", testName))
	check(err, "Failed to create output file")
	defer file.Close()
	testContext := fmt.Sprintf(
		"Branch:,%s\nBlockTime:,%d\nExpectedMinuteTime:,%f\nNotes:\n\n",
		branchName,
		blkTime,
		idealMinuteTime.Seconds(),
	)
	_, err = file.WriteString(testContext)
	check(err, "Failed to write test context to file")
	_, err = file.WriteString("Block Number,Minute Number,TPS,Minute Length (s),Minute Average (s)\n")
	check(err, "Failed to write headers to file")

	// Wait for a new minute to start
	previous, err := getCurrentMinute()
	check(err, "Failed to request current minute")
	current := previous
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		current, err = getCurrentMinute()
		check(err, "Failed to request current minute")
		if current.Minute != previous.Minute {
			previous = current
			break
		}
	}

	// Start generating load and check for block minute length slippage (in both directions)
	tps := initialTPS
	minutesToNextThrottle := throttleInterval
	var recentMinutes []time.Duration
	throttle := make(chan int)
	go generateLoad(tps, ecAddress, throttle)
	ticker = time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		current, err = getCurrentMinute()
		check(err, "Failed to request current minute")
		if previous.Minute != current.Minute {
			// New minute started. Calculate summary metrics on previous minute
			previousMinuteTime := time.Duration(current.MinuteStartTime - previous.MinuteStartTime)
			row := fmt.Sprintf("%d,%d,%d,%f", previous.BlockHeight, previous.Minute, tps, previousMinuteTime.Seconds())
			recentMinutes = append(recentMinutes, previousMinuteTime)

			// Check if it's time to make a throttling decision
			if minutesToNextThrottle--; minutesToNextThrottle == 0 {
				avgMinuteLength := time.Duration(0)
				for _, minuteLength := range recentMinutes {
					avgMinuteLength += minuteLength
				}
				avgMinuteLength = avgMinuteLength / time.Duration(throttleInterval)
				row = fmt.Sprintf("%s,%f\n", row, avgMinuteLength.Seconds())

				// TODO: adjust the applied load tps accordingly, not just blindly incrementing
				throttle <- 1
				tps++
				minutesToNextThrottle = throttleInterval
				recentMinutes = nil
			} else {
				row = fmt.Sprintf("%s\n", row)
			}

			// Write summary to output spreadsheet
			go func() {
				_, err = file.WriteString(row)
				check(err, "Failed to write minute to file")
			}()
			previous = current

			fmt.Printf("Started minute %d at %d tps\n-----------------------------------\n", current.Minute, tps)
		}
	}
}

// generateLoad starts sending a load of txs at the specified tps.
// An integer sent through the throttle channel raises/lowers the load tps by that amount.
func generateLoad(tps int, ecAddress *factom.ECAddress, throttle chan int) {
	ticker := time.NewTicker(time.Second)
	for ; true; <- ticker.C {
		select {
		case change := <-throttle:
			tps = tps + change
		default:

		}
		go sendNTransactions(tps, ecAddress)
	}
}

func sendNTransactions(n int, ecAddress *factom.ECAddress) {
	var chain *factom.Chain = nil
	for i := 0; i < n; i++ {
		if chain == nil {
			// New random chain for this set of txs
			e := factom.Entry{
				ExtIDs: make([][]byte, rand.Intn(4) + 1),
			}
			for i := range e.ExtIDs {
				e.ExtIDs[i] = random.RandByteSliceOfLen(rand.Intn(300))
			}
			chain = factom.NewChain(&e)
			_, err := factom.CommitChain(chain, ecAddress)
			check(err, "Failed to commit chain")
			_, err = factom.RevealChain(chain)
			check(err, "Failed to reveal chain")
			continue
		}
		// Entry with random ExtIDs and Content
		e := factom.Entry{
			ChainID: chain.ChainID,
			ExtIDs: make([][]byte, rand.Intn(4) + 1),
			Content: random.RandByteSliceOfLen(rand.Intn(4000)),
		}
		for i := range e.ExtIDs {
			e.ExtIDs[i] = random.RandByteSliceOfLen(rand.Intn(300))
		}

		_, err := factom.CommitEntry(&e, ecAddress)
		check(err, "Failed to commit entry")
		_, err = factom.RevealEntry(&e)
		check(err, "Failed to reveal entry")
		//fmt.Printf("CommitTxID: %s\n", txID)
	}
	fmt.Printf("Load of %d tx sent\n", n)
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