package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type responseData struct {
	Campsites map[int]interface{} `json:"campsites"`
	Count     int                 `json:"count"`
}

type chanMessage struct {
	response []byte
	id       int
}

func main() {
	locationIds := []int{232447, 232449, 232451, 232450}
	tw := ""
	ch := make(chan chanMessage)

	for _, id := range locationIds {
		go makeRequest(id, ch)
	}

	for i := 0; i < len(locationIds); i++ {
		m := <-ch
		var countAvailable int
		var result responseData
		json.Unmarshal(m.response, &result)
		for _, value := range result.Campsites {
			cs, ok := value.(map[string]interface{})

			if ok == false {
				fmt.Println("Error occured::: changing campsite types")
			}
			av := cs["availabilities"].(map[string]interface{})

			campsiteAvailabilities := false
			for _, a := range av {
				if a == "Available" {
					campsiteAvailabilities = true
				}
			}

			if campsiteAvailabilities {
				countAvailable++
			}
		}

		s := fmt.Sprintf("Area id: %v, %v campsites out of %v have availabilities on July 3\n", m.id, countAvailable, result.Count)
		fmt.Println(s)

		// TODO : tweet at user
		if countAvailable > 0 {
			tw = tw + s
		}
	}

	if len(tw) > 0 {
		fmt.Println("About to tweet")
		tweet(tw)
	} else {
		fmt.Println("Not tweeting")
	}

}

func makeRequest(id int, ch chan chanMessage) {
	sID := strconv.Itoa(id)
	url := "http://www.recreation.gov/api/camps/availability/campground/" + sID + "?start_date=2020-07-03T00%3A00%3A00.000Z&end_date=2020-07-04T00%3A00%3A00.000Z"
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Golang")
	resp, _ := client.Do(req)
	bs, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("An Error has occurred")
	}

	m := chanMessage{bs, id}

	ch <- m
}

func tweet(tw string) {
	C := setConfig()
	config := oauth1.NewConfig(C.consumerKey, C.consumerSecret)
	token := oauth1.NewToken(C.accessToken, C.accessSecret)
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// twitter client
	client := twitter.NewClient(httpClient)
	_, _, err := client.Statuses.Update(tw, nil)

	fmt.Println(err, ":::err")
}
