package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv" // for env files
	_ "github.com/lib/pq"      //Postgres lib
	"github.com/rs/cors"
	"gopkg.in/gomail.v2" //simple way to send mails in golang
)

// Different functions and data types

// 1) struct of the diff data type for both requests and response from frontend

//2) function for the mail notification mechanism

/*3) function to do lookup in the database and additionally retrieve or
send requests to it too and if generates it, it adds it to the current script
also need to end up checking for simulated population if place of population
less than 10k (our std size)  then we will */

//4) Database function script and its modifications

//5) Some sorta anti botting mechanism or rate limiting

//6) Implementing some middleware (Cors) if required

//7) func to handle requests from the frontend

type Request struct {
	State                string `json:"State"`
	District             string `json:"District"`      // will probably try rearranging it by state
	Population_number    int    `json:"Population_No"` // (district should show the number via javascript ig should be dynamic? need a key value pair or smth)
	Population_Simulated uint   `json:"Population_Simulated"`
	Email                string `json:"email"` //regex added in the frontend itself

}

type Response struct {
	Status         string `json:"status"`
	Estimated_Time string `json:"Estimated Time"`
}

// Global variables
var (
	requestQueue = make(chan Request, 100) //Buffered Channel for Request Queue
	db           *sql.DB                   // Global Database connection
	wg           sync.WaitGroup            // WaitGroup for goroutines
)

// Initialise database connection
func initDB() (*sql.DB, error) {
	//Load env file
	err := godotenv.Load()
	//postgressql server connect
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}
	//Get connection detail from environment variables

	dbUser := os.Getenv("DB_USER")
	dbPassowrd := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	//Construct the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassowrd, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}
	// Test connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping databse : %v", err)
	}
	return db, nil
}

// Func to send mail to the users with CSV attachments using  gomail
func emailUserCsv(email string, csvFilePath string) error {
	//implement smtp connection and email sending
	m := gomail.NewMessage()
	m.SetHeader("From", "<My_Email_Credentials")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Requested Synthetic Population CSV")
	m.SetBody("text/plain", "Please find attached your requested synthetic population data below")
	m.Attach(csvFilePath)

	// Use environment variables for email configuration
	smtpHost := os.Getenv("SMTP_HOST")
	//smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	d := gomail.NewDialer(smtpHost, 587, smtpUser, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("Failed to send email: %v", err)
	}
	return nil
}

// Function to add data to database
func DatabaseAdd(state, district, csvFilePath string) error {
	/*Traverse directories:
	if struct.State == Database.state:
		open District
			if struct.District = Database.District && District_Synth_population_size != District_Synth.csv
				Add District_synth_population_no.csv to the database
		disconnect with database (but not sure if i should keep it running or close when not in use ) */
	_, err := db.Exec("INSERT INTO synthetic_populations (state, district, file_path) VALUES($1, $2, $3) ON CONFLICT (state,district) DO UPDATE SET file_path=$3", state, district, csvFilePath)
	if err != nil {
		return fmt.Errorf("Failed to add to databse: %v", err)
	}
	return nil
}

/*store requests in array or some data structure and keep it queued
when no running object on server
remove object from list and call func.main()*/

func districtExistsInDB(state, district string) (string, error) {
	var filePath string
	err := db.QueryRow("SELECT file_path FROM synthetic_populations WHERE state = $1 AND district = $2", state, district).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil //no error, but no file found
		}
		return "", fmt.Errorf("database query failed %v", err)
	}
	return filePath, nil
}

func processRequest(req Request) {
	//get request from html
	/*t.District = struct request
	if t.District is in Database (
		return District_Synth.csv as json
	)
	else:
		call python3 generate.py --state_name <STATE> --district <district>
		if (call == unsuccesful && District_Synth.csv != EMPTY ):
			return error
		elif succesful:
			(func email_user_csv(struct Email,District_Synth.csv))
			func Database_add(District_Synth.csv)
			District_synth.csv call python3 Report_generation.py
			Report_generation.py takes the file and returns images of statistics back in the file
	*/
	defer wg.Done() //Ensure wait group is deceremented when function completes
	csvFilePath, err := districtExistsInDB(req.State, req.District)
	if err != nil {
		fmt.Printf("Error checking databse: %v\n", err)
		return
	}
	if csvFilePath == "" {
		//Generate new synthpop
		cmd := exec.Command("python3", "generate.py", "--state_name", req.State, "--district", req.District)
		if err := cmd.Run(); err != nil {
			fmt.Printf("error generating synthetic population %v\n", err)
			return
		}
		//Assume CSV is generated in known location
		csvFilePath = fmt.Sprintf("path/to/generated/%s_%s.csv", req.State, req.District)

		//Add CSV to database
		if err := DatabaseAdd(req.State, req.District, csvFilePath); err != nil {
			fmt.Printf("Error adding to database: %v\n", err)
			return
		}
	}

	//email csv to user
	if err := emailUserCsv(req.Email, csvFilePath); err != nil {
		fmt.Printf("Error sending mail %v", err)
		return
	}
	//Generating report
	cmd := exec.Command("python3", "Report_generation.py", csvFilePath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating report: %v\n", err)

	}
}

// Handler for incoming requests
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	//Check if queue is full
	if len(requestQueue) >= cap(requestQueue) {
		http.Error(w, "Server is at capacity. Please try again later.", http.StatusServiceUnavailable)
		return
	}
	//Add request to queue
	select {
	case requestQueue <- req:
		resp := Response{
			Status:         "Request queued",
			Estimated_Time: "Processing time may vary",
		}
		json.NewEncoder(w).Encode(resp)
	case <-time.After(5 * time.Second):
		http.Error(w, "Request timed out. Server might be overloaded", http.StatusRequestTimeout)
	}
}

// Worker to process requests from Queue
func worker(id int) {
	for req := range requestQueue {
		fmt.Printf("Worker %d processing request for %s, %s \n", id, req.State, req.District)
		wg.Add(1)
		processRequest(req)
	}
}

func main() {
	//Initialise database
	var err error
	db, err = initDB()
	if err != nil {
		fmt.Printf("Failed to initialise database : %v \n", err)
		return
	}
	defer db.Close()
	// Determine no of worker goroutines
	numWorkers := 5
	fmt.Printf("starting %d workers \n", numWorkers)

	//Start worker go routines
	for i := 0; i < numWorkers; i++ {
		go worker(i)
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/request", HandleRequest).Methods("POST")

	// Serve static files from the frontend directory
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend")))

	// Creating a CORs wrapper

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, //allows all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})
	//Wrapping router with CORs middleware

	handler := c.Handler(r)

	fmt.Println("server is running on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("server failed to start %v\n", err)
	}

	//Close queue and wait for all go routines to finish
	close(requestQueue)
	wg.Wait()
}
