package main 

import("fmt"
"github.com/gorilla/mux"
)

// Different functions and data types

// 1) struct of the diff data type for both requests and response from frontend

//2) function for the mail notification mechanism

/*3) function to do lookup in the python script and additionally retrieve or
 send requests to it too and if generates it, it adds it to the current script
 also need to end up checking for simulated population if place of population
 less than 10k (our std size)  then we will */

//4) Database function script and its modifications 

//5) Some sorta anti botting mechanism or rate limiting 

//6) Implementing some middleware (Cors) if required 

//7) func to handle requests from the frontend 

//8) Create a key map 


struct request {
	District string `json:"District"`// will probably try rearranging it by state 
	Population_number `json: Population_No` // (district should show the number via javascript ig should be dynamic? need a key value pair or smth)
	Population_Simulated `json:"Population_Simulated"` uint
	Email string `json: "email"` //regex added in the frontend itself 

}

struct Response {
	Status string `json:"status"`
	Estimated_Time string `json:"Estimated Time "`
}