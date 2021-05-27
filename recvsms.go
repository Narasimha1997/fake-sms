package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

const (
	pageURL     = "https://receive-smss.com/"
	cookieName  = "__cfduid"
	smsEndpoint = "sms/"
)

//ScrapeAvailableNumbers Extracts the list of phone-numbers from the page
func ScrapeAvailableNumbers() []Number {
	response, err := soup.Get(pageURL)
	if err != nil {
		log.Fatalf("Failed to make HTTP request to %s\n", pageURL)
	}

	numbers := make([]Number, 0)

	//scrape the page
	document := soup.HTMLParse(response)
	numberBoxes := document.Find("div", "class", "number-boxes").FindAllStrict(
		"div", "class", "number-boxes-item d-flex flex-column ",
	)

	for _, numberBox := range numberBoxes {
		numberElement := numberBox.FindStrict("div", "class", "row")
		if numberElement.Error == nil {
			numberContainer := numberElement.FindStrict("h4")
			countryContainer := numberElement.FindStrict("h5")
			if numberContainer.Error == nil && countryContainer.Error == nil {
				number := Number{
					CreatedAt: time.Now().Format("2006-01-02 15:04:05 Monday"),
					Number:    numberContainer.Text(),
					Country:   countryContainer.Text(),
				}

				numbers = append(numbers, number)
			}
		}
	}

	return numbers
}

//ScrapeMessagesForNumber GET SMS from number
func ScrapeMessagesForNumber(number string) []Message {
	//Get cookie first
	resp, err := http.Get(pageURL)
	if err != nil {
		log.Fatalln("Failed to make GET request")
	}

	cookies := resp.Cookies()
	cookieValue := ""
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			cookieValue = cookie.Value
		}
	}

	//now use that value to set the cookie in soup
	soup.Cookie(cookieName, cookieValue)
	requestURL := pageURL + smsEndpoint + strings.ReplaceAll(number, "+", "") + "/"

	//make GET with soup:
	response, err := soup.Get(requestURL)

	document := soup.HTMLParse(response)

	table := document.Find("table")
	if table.Error != nil {
		log.Fatalln("Failed to load messages")
	}

	tbody := table.Find("tbody")
	if tbody.Error != nil {
		log.Fatalln("Failed to load messages")
	}

	tableRows := tbody.FindAll("tr")

	messages := make([]Message, 0)

	for _, row := range tableRows {
		cols := row.FindAll("td")
		if len(cols) < 6 {
			continue
		}

		message := Message{
			Originator: cols[1].FullText(),
			Body:       cols[4].FullText(),
			CreatedAt:  cols[3].FullText(),
		}

		messages = append(messages, message)
	}

	return messages
}
