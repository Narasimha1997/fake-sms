package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	pageURL = "https://onlinesim.io/"
)

func ScrapeAvailableCountries() []Country {
	apiURL := fmt.Sprintf("%sapi/v1/free_numbers_content/countries", pageURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return data.Countries
}

// ScrapeAvailableNumbers Extracts the list of phone-numbers from the page
func ScrapeAvailableNumbers(code string) []Number {
	apiURL := fmt.Sprintf("%sapi/v1/free_numbers_content/countries/%s", pageURL, code)

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var data ResponseNumbers
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return data.Numbers
}

// ScrapeMessagesForNumber GET SMS from number
func ScrapeMessagesForNumber(countryCode int, number string) []Message {
	apiURL := fmt.Sprintf("%sapi/getFreeMessageList?&phone=%s&country=%d", pageURL, number, countryCode)

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var data MessagesResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return data.Messages.Data
}
