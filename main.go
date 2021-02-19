package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/manifoldco/promptui"
)

const (
	//APIUrl : URL of the API
	APIUrl = "https://upmasked.com"

	//RegisterRoute : Can register new phone-numbers with this route
	RegisterRoute = "/api/sms/numbers"

	//QueryRoute : Can query SMS given the message
	QueryRoute = "/api/sms/messages/"
)

//Number A struct that represents a new number to be addeded
type Number struct {
	Country   string `json:"country"`
	Number    string `json:"number"`
	CreatedAt string `json:"created_at"`
}

//Message a struct which represents the message
type Message struct {
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	Originator string `json:"originator"`
}

//Numbers A list of Number type
type Numbers []Number

//Messages A list of Message type
type Messages []Message

func exitFatal(err error) {
	log.Fatal(err)
}

//DB The database functions group
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

func numbersToList(numbers *Numbers) *[]string {
	listOfNumbers := make([]string, len(*numbers))
	for idx, number := range *numbers {
		listOfNumbers[idx] = fmt.Sprintf("%s (%s)", number.Number, number.Country)
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

func getAvailNumbers() *Numbers {
	uri := APIUrl + RegisterRoute
	response, err := http.Get(uri)
	if err != nil {
		exitFatal(err)
	}

	defer response.Body.Close()

	//read the body and marshall
	respBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		exitFatal(err)
	}

	numbers := Numbers{}

	err = json.Unmarshal(respBytes, &numbers)
	if err != nil {
		exitFatal(err)
	}

	return &numbers
}

func registerNumber() {
	numbers := getAvailNumbers()

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

func listNumbers() {
	db := DB{}
	numbers := db.getFromDB()

	fmt.Println("Country\t\t\tNumber\t\t\tCreated At")
	fmt.Println("=======================================================================")
	for _, number := range *numbers {
		fmt.Printf(
			"%s\t\t\t%s\t\t%s\n",
			number.Country, number.Number, number.CreatedAt,
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
		isMatch := r.Match([]byte(message.Body))
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
		fmt.Printf("Selected %s, fetching messages\n", selectedNumber)

		//check message
		url := APIUrl + QueryRoute + selectedNumber.Number
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to make GET request %s\n", url)
		}

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to decode response")
		}

		messages := Messages{}
		err = json.Unmarshal(data, &messages)
		if err != nil {
			log.Fatalf("Failed to de-serialize response body")
		}

		//run filter if enabled:
		if enableFilter {
			fmt.Println("Enter the filter regular expression:")
			filterString := `*`
			fmt.Scanln(&filterString)

			//run the filter
			messages = messagePatternCheck(&filterString, &messages)
		}

		fmt.Println("===========================================")
		for _, message := range messages {
			fmt.Printf("Sender : %s, at : %s\n", message.Originator, message.CreatedAt)
			fmt.Printf("Body : %s\n", message.Body)
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
