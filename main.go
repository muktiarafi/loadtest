package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Order struct {
	Products    []*Product `json:"products"`
	PaymentType string     `json:"paymentType"`
}

type Product struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

func main() {
	url := os.Getenv("URL")
	orderEndpoint := fmt.Sprintf("http://%s/api/orders", url)

	users, err := readFile("./users.txt")
	if err != nil {
		log.Fatal(err)
	}

	producst, err := readFile("./products.txt")
	if err != nil {
		log.Fatal(err)
	}

	rate := vegeta.Rate{Freq: 3000, Per: time.Second}
	duration := 10 * time.Minute

	targeter := func() vegeta.Targeter {
		return func(target *vegeta.Target) error {
			target.Method = http.MethodPost
			target.URL = orderEndpoint
			target.Header = http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": {"Basic " + basicAuth(users[rand.Intn(len(users))], "")},
			}

			body1 := Order{
				Products: []*Product{
					{
						ID:       producst[rand.Intn(len(producst))],
						Quantity: 2,
					},
				},
				PaymentType: "PAYMENT_TYPE_BALANCE",
			}

			target.Body, err = json.Marshal(body1)

			return err
		}
	}()
	attacker := vegeta.NewAttacker()

	fmt.Println("Running loadtest....")
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}

func readFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var users []string
	for scanner.Scan() {
		users = append(users, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
