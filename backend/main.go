package main 

import(
	"fmt"
	"os"
	"github.com/gorilla/mux"
	"net/smtp"
	"database/sql"
	"os/exec"
	"github.com/lib/pq" //Postgres lib
	"gopkg.in/gomail.v2" //simple way to send mails in golang
	"net/http"
	"encoding/json"
	"sync"
	"time"
	"errors"
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
	State 					string 	`json:"State"`
	District 				string 	`json:"District"`// will probably try rearranging it by state 
	Population_number 		int 	`json: Population_No` // (district should show the number via javascript ig should be dynamic? need a key value pair or smth)
	Population_Simulated 	uint 	`json:"Population_Simulated"` 
	Email 			        string 	`json: "email"` //regex added in the frontend itself 

}

type Response struct {
	Status 			string `json:"status"`
	Estimated_Time 	string `json:"Estimated Time"`
}

//Global variables 
var(
	requestQueue = make(chan Request, 100) //Buffered Channel for Request Queue
	db             *sql.DB                 // Global Database connection
 	wg             sync.WaitGroup		   // WaitGroup for goroutines
)

//Initialise database connection 
func initDB() error {
	var err error 
	db, err = s
	//postgressql server connect 
	db, err := sql.Open("posgres", "postgresql://username:password@localhost/database_name?sslmode=disable")
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v, err")
	}
	return db.Ping() //verify connection 
}

//Func to send mail to the users with CSV attachments using  gomail
func emailUserCsv(email string, csvFilePath string) error {
	//implement smtp connection and email sending
	m := gomail.NewMessage()
	m.SetHeader("From", "<My_Email_Credentials")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Requested Synthetic Population CSV")
	m.SetBody("text/plain", "Please find attached your requested synthetic population data below")
	m.Attach(csvFilePath)

	d := gomail.NewDialer("smtp.example.com", 587, "<My_Email_Credentials>","<Password>")
	
	if err:= d.DialAndSend(m); err!=nil{
		return fmt.Errorf("Failed to send email: %v", err)
	}
	return nil
}	

//Function to add data to database
func DatabaseAdd(state, district, csvFilePath string) error  {
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




func distrixtEsitsInDB(state, district string) (string, error) {
	var filePath string 
	err := db.QueryROw("SELECT file_path FROM synthetic populations WHERE state = $1 AND district = $2", state,district).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows{
			return "", nil //no error, but no file found
		}
		return "", fmt.Errorf("database query failed %v," err)
	}
	return filePath,nil
}


func processRequest (req Request) {
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
}