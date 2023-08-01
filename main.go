package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/manifoldco/promptui"
)

// Country A struct that represents a new country to be addeded
type Country struct {
	Country int    `json:"country"`
	Name    string `json:"name"`
	Online  bool   `json:"online"`
	Locale  string `json:"locale"`
}

type Response struct {
	Response  string    `json:"response"`
	Countries []Country `json:"counties"`
}

// Number A struct that represents a new number to be addeded
type Number struct {
	Country    int    `json:"country"`
	DataHumans string `json:"data_humans"`
	FullNumber string `json:"full_number"`
	Number     string `json:"number"`
	Code       string `json:"code"`
	IsArchive  bool   `json:"is_archive"`
}

type ResponseNumbers struct {
	Response string   `json:"response"`
	Code     int      `json:"code"`
	Numbers  []Number `json:"numbers"`
}

// Message a struct which represents the message
type Message struct {
	Text       string `json:"text"`
	InNumber   string `json:"in_number"`
	MyNumber   int    `json:"my_number"`
	CreatedAt  string `json:"created_at"`
	DataHumans string `json:"data_humans"`
	Code       string `json:"code"`
}

type MessagesResponse struct {
	Response     int         `json:"response"`
	Messages     Msg         `json:"messages"`
	CurrentPage  int         `json:"current_page"`
	FirstPageURL string      `json:"first_page_url"`
	From         int         `json:"from"`
	LastPage     int         `json:"last_page"`
	LastPageURL  string      `json:"last_page_url"`
	Links        []Link      `json:"links"`
	NextPageURL  string      `json:"next_page_url"`
	Path         string      `json:"path"`
	PerPage      int         `json:"per_page"`
	PrevPageURL  interface{} `json:"prev_page_url"`
	To           int         `json:"to"`
	Total        int         `json:"total"`
}

type Msg struct {
	CurrentPage int       `json:"current_page"`
	Data        []Message `json:"data"`
}

type Link struct {
	URL    string `json:"url"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type Countries []Country

// Numbers A list of Number type
type Numbers []Number

// Messages A list of Message type
type Messages []Message

func exitFatal(err error) {
	log.Fatal(err)
}

// DB The database functions group
type DB struct {
}

func (d *DB) getDBPath() string {
	/*
		Look for the path to be specified in ENV FAKE_SMS_DB_DIR,
		if not, use default $HOME as the path to create DB.
		The DB will be created at <db_dir>/.fake-sms/db.json
		If the DB does not exist, it will be created and will be
		initialized to an empty array []
	*/

	dbPath, exists := os.LookupEnv("FAKE_SMS_DB_DIR")
	if !exists {
		dbPath = os.Getenv("HOME")
		dbPath = filepath.Join(dbPath, ".fake-sms")
	}

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dbPath, 0700)
		if err != nil {
			log.Fatalf("Failed to create DB directory at %s\n", dbPath)
		}
	}

	dbPath = filepath.Join(dbPath, "db.json")
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		emptyArray := []byte("[\n]\n")
		err = ioutil.WriteFile(dbPath, emptyArray, 0700)
		if err != nil {
			log.Fatalf("Faild to create DB file at %s\n", dbPath)
		}
	}

	return dbPath
}

func (d *DB) addToDB(number *Number) {
	dbPath := d.getDBPath()
	//read and serialize it to numbers
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read DB file at %s\n", dbPath)
	}

	//unmarshall the db to Numbers type
	numbers := Numbers{}
	err = json.Unmarshal(data, &numbers)
	if err != nil {
		log.Fatalf("Failed to de-serialize DB file %s\n", dbPath)
	}

	numbers = append(numbers, *number)

	//write it back to the db
	data, err = json.Marshal(numbers)
	if err != nil {
		log.Fatalf("Failed to serialize DB file %s\n", dbPath)
	}

	err = ioutil.WriteFile(dbPath, data, 0700)
	if err != nil {
		log.Fatalf("Failed to save DB file %s\n", dbPath)
	}
}

func (d *DB) getFromDB() *Numbers {
	dbPath := d.getDBPath()
	//read and serialize it to numbers
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read DB file at %s\n", dbPath)
	}

	//unmarshall the db to Numbers type
	numbers := Numbers{}
	err = json.Unmarshal(data, &numbers)
	if err != nil {
		log.Fatalf("Failed to de-serialize DB file %s\n", dbPath)
	}

	return &numbers
}

func (d *DB) deleteFromDB(idx *int) {
	dbPath := d.getDBPath()
	//read and serialize it to numbers
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read DB file at %s\n", dbPath)
	}

	//unmarshall the db to Numbers type
	numbers := Numbers{}
	err = json.Unmarshal(data, &numbers)
	if err != nil {
		log.Fatalf("Failed to de-serialize DB file %s\n", dbPath)
	}

	//delete by index
	if *idx > len(numbers)-1 {
		log.Fatalln("Number does not exist to be deleted from DB")
	}

	numbers = append(numbers[:*idx], numbers[*idx+1:]...)
	//serialize it back
	data, err = json.Marshal(numbers)
	if err != nil {
		log.Fatalf("Failed to serialize DB file %s\n", dbPath)
	}

	err = ioutil.WriteFile(dbPath, data, 0700)
	if err != nil {
		log.Fatalf("Failed to save DB file %s\n", dbPath)
	}
}

func countriesToList(countries *Countries) *[]string {
	listOfCountries := make([]string, len(*countries))
	for idx, country := range *countries {
		listOfCountries[idx] = fmt.Sprintf("%s (%d)", country.Name, country.Country)
	}
	return &listOfCountries
}

func numbersToList(numbers *Numbers) *[]string {
	listOfNumbers := make([]string, len(*numbers))
	for idx, number := range *numbers {
		listOfNumbers[idx] = fmt.Sprintf("%s (%d)", number.Code, number.Country)
	}
	return &listOfNumbers
}

func displayInitParameters() int {
	prompt := promptui.Select{
		Label: "What you want to do?",
		Items: []string{"Add a new number", "List my numbers", "Remove a number", "Get my messages", "Exit"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		exitFatal(err)
	}

	//Return the index of the parameter selected
	return idx
}

func getAvailCountries() *Countries {
	counriesArray := ScrapeAvailableCountries()
	countries := Countries(counriesArray)
	return &countries
}

func listCountries() {
	countries := getAvailCountries()

	fmt.Println("Name\t\tCode")
	fmt.Println("=======================================================================")
	for _, country := range *countries {
		fmt.Printf(
			"%s\t\t%n\n",
			country.Name, country.Country,
		)
	}
}

func getAvailNumbers(country string) *Numbers {
	numArray := ScrapeAvailableNumbers(country)
	numbers := Numbers(numArray)
	return &numbers
}

func registerNumber() {
	countries := getAvailCountries()

	if len(*countries) == 0 {
		fmt.Println("No new countries available right now")
	} else {
		countriesList := countriesToList(countries)
		//display numbers
		prompt := promptui.Select{
			Label: "These are the available countries, choose any one of them",
			Items: *countriesList,
		}

		code, _, err := prompt.Run()
		if err != nil {
			exitFatal(err)
		}

		if code == -1 {
			fmt.Println("Nothing selected")
		} else {
			//new number selected, save it to the database file
			selectedCountry := &(*countries)[code]

			numbers := getAvailNumbers(selectedCountry.Name)

			if len(*numbers) == 0 {
				fmt.Println("No new numbers available right now")
			} else {
				numberList := numbersToList(numbers)
				//display numbers
				prompt := promptui.Select{
					Label: "These are the available numbers, choose any one of them",
					Items: *numberList,
				}

				idx, _, err := prompt.Run()
				if err != nil {
					exitFatal(err)
				}

				if idx == -1 {
					fmt.Println("Nothing selected")
				} else {
					//new number selected, save it to the database file
					selectedNumber := &(*numbers)[idx]
					fmt.Printf("Selected %s, saving to database\n", selectedNumber)
					db := DB{}
					db.addToDB(selectedNumber)
				}
			}
		}
	}

}

func listNumbers() {
	db := DB{}
	numbers := db.getFromDB()

	fmt.Println("Country\t\tNumber\t\tData Humans")
	fmt.Println("=======================================================================")
	for _, number := range *numbers {
		fmt.Printf(
			"%d\t\t%s\t\t%s\n",
			number.Country, number.Number, number.DataHumans,
		)
	}
}

func removeNumbers() {
	db := DB{}
	numbers := db.getFromDB()

	numberList := numbersToList(numbers)

	if len(*numberList) == 0 {
		log.Fatalln("No numbers saved to delete")
	}

	//display the list
	prompt := promptui.Select{
		Label: "These are the available numbers, choose any one of them",
		Items: *numberList,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		exitFatal(err)
	}

	if idx == -1 {
		fmt.Println("Nothing selected")
	} else {
		//new number selected, save it to the database file
		selectedNumber := &(*numbers)[idx]
		fmt.Printf("Selected %s, removing from database\n", selectedNumber)
		db.deleteFromDB(&idx)
	}
}

func messagePatternCheck(pattern *string, messages *Messages) Messages {
	r, err := regexp.Compile(*pattern)
	if err != nil {
		log.Fatalln("Invalid regular expression provided")
	}

	filteredMessages := make([]Message, 0)
	for _, message := range *messages {
		//check match
		isMatch := r.Match([]byte(message.Text))
		if isMatch {
			filteredMessages = append(filteredMessages, message)
		}
	}

	return Messages(filteredMessages)
}

func checkMessages(enableFilter bool) {

	db := DB{}
	numbers := db.getFromDB()

	numberList := numbersToList(numbers)

	if len(*numberList) == 0 {
		log.Fatalln("No numbers saved to delete")
	}

	//display the list
	prompt := promptui.Select{
		Label: "These are the available numbers, choose any one of them",
		Items: *numberList,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		exitFatal(err)
	}

	if idx == -1 {
		fmt.Println("Nothing selected")
	} else {
		//new number selected, save it to the database file
		selectedNumber := &(*numbers)[idx]
		fmt.Printf("Selected %s, fetching messages\n", selectedNumber.FullNumber)

		messagesArray := ScrapeMessagesForNumber(selectedNumber.Country, selectedNumber.Number)

		//check message
		messages := Messages(messagesArray)

		//run filter if enabled:
		if enableFilter {
			fmt.Println("Enter the filter regular expression:")
			userFilterInput := ""

			fmt.Scanln(&userFilterInput)
			if userFilterInput == "" {
				userFilterInput = `.*`
			}

			//run the filter
			messages = messagePatternCheck(&userFilterInput, &messages)
		}

		fmt.Println("===========================================")
		for _, message := range messages {
			fmt.Printf("Sender : %s, at : %s\n", message.InNumber, message.CreatedAt)
			fmt.Printf("Body : %s\n", message.Text)
			fmt.Println("===========================================")
		}

		indentedData, _ := json.MarshalIndent(messages, "", "\t")

		//save the body as json
		fileName := fmt.Sprintf("%s.json", selectedNumber.Number)
		err = ioutil.WriteFile(fileName, indentedData, 0700)
		if err != nil {
			log.Fatalf("Failed to save file %s\n", fileName)
		}
	}
}

func shouldIncludeFilter() bool {
	prompt := promptui.Select{
		Label: "Do you want to filter the messages?",
		Items: []string{"Yes", "No"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		log.Fatalln("Failed to render prompt")
	}

	if idx == 0 {
		return true
	}

	return false
}

func main() {
	for true {
		idx := displayInitParameters()

		switch idx {
		case 0:
			registerNumber()
			break
		case 1:
			listNumbers()
			break
		case 2:
			removeNumbers()
			break
		case 3:
			//check if filter needs to be enabled
			includeFilter := shouldIncludeFilter()
			checkMessages(includeFilter)
			break
		case 4:
			fmt.Println("Bye!")
			os.Exit(0)
		default:
			log.Fatalf("Option %d yet to be implemented\n", idx)
		}
	}
}
