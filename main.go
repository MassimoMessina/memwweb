package main

import (
		"flag"
		"fmt"
		"errors"
		"log"
 		"bufio"
 		"strconv"
 		"strings"
  		"os"
  		"encoding/json"
	    //"io/ioutil"
	    "net/http"
	    "github.com/gorilla/mux"
)


//Global Variable definition
var mapNumbers map[int]string
var aLines []string
var aConsonanti []string
var aDiscard []string

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
  
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  	}
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    aLines = append(aLines, scanner.Text())
 	}
  return aLines, scanner.Err()
}

// It will scan the words matching the rules for the specified number
func FindWords (number int) ([5000]string, int) {

	inxWords := 0
	var words [5000]string

	//log.Printf("memwords - FindWords number=%d\n", number)
	 
	val := mapNumbers[number]
	first, _ := strconv.Atoi (string(val[0]))
	second, _ := strconv.Atoi (string(val[1]))

	//init the discard box
	copy (aDiscard, aConsonanti)
	aDiscard[first]=" "
	aDiscard[second]=" "
	//fmt.Println(aDiscard)

	bWord_is_good:=false
	bFirst_is_good:=false
	bFirst_is_not_burned:=true


	var lastChar string
	//log.Printf("memwords first=%d second=%d, consonant=%s,%s\n", first, second, aConsonanti[first] ,aConsonanti[second])

	 
	//For all the words in the dictionary
	scan_aLines: for wi, word := range aLines {

	 	wordUpper := strings.ToUpper(word)
	 	//log.Print(wordUpper)
	 	bWord_is_good=false
	    bFirst_is_good=false
	    bFirst_is_not_burned=true

	 	// For all the characters in the word
	 	for _, c := range wordUpper {

	 		st := string(c)

	 		if st == string(aConsonanti[first]) && bFirst_is_not_burned {
	 			if bFirst_is_good {
	 				if lastChar==st {
	 					//Verify if is the case such as anno with a word valid for double consonant
	 					if st == string(aConsonanti[second]) {
	 						bWord_is_good=true
	 						continue 
	 				    }
	 				    // it is a case such as "annoiare"
	 				    continue // it means that it is a double consonant

	 				} else {
	 					//continue scan_aLines  // if you have another occurrence of first then discard the word
	 					bWord_is_good=true
	 					lastChar=st
	 					continue
	 				    }
	 				
	 			}
	 			bFirst_is_good=true
	 			lastChar=st
	 			continue
	 		}

	 		if st == string(aConsonanti[second] ) {
	 			if  bWord_is_good {
	 				if lastChar==st {
	 					continue // it means that it is a double consonant
	 				} else {
	 					continue scan_aLines  // if you have another occurrence of first then discard the word
	 				    }
	 			}
	 			if bFirst_is_good {
	 				bWord_is_good=true
	 				lastChar=st
	 				continue
	 			}
	 			
	 			continue scan_aLines
	 		}

	 		lastChar=st


	 		// following test is required for cases such as "fifa" 
	 		if bFirst_is_good {
	 			bFirst_is_not_burned=false
	 			}
	 		
	 		for _ ,d := range aDiscard {
	 			//log.Printf("st=%s d=%s ; ", st, string(d))
	 			if (st==string(d)) {
	 				//log.Printf("scan interrotto per %s", wordUpper)
	 				continue scan_aLines
	 			}
	 		}

	 	} // end scan wordUpper

	 	if bWord_is_good {
	 		//log.Printf("memwords trovata %s\n", wordUpper)
	 		words[inxWords]=aLines[wi]
	 		inxWords++
	 	}
	 		

	}  // End scan aLines
	 
	 return words, inxWords

}


// Creates the map containing the association between Numbers 
// and its string representation with leading zeros
func CreateNumbersMap () (error) {

	 mapNumbers = make (map [int] string)

	 for i := 0 ; i < 100 ; i++ {

	 	n := fmt.Sprintf("%02d",i)
	 	mapNumbers[i]=n
	 	//log.Printf("%d (%s)\n", i, mapNumbers[i] )
	 } 
	 
	 log.Printf("memwords - Created mapNumbers with %d elements\n", len(mapNumbers))
	 return  errors.New("Completed SUccessfully")

}

// Start of new code

// error response contains everything we need to use http.Error
type handlerError struct {
	Error   error
	Message string
	Code    int
}

// word model
type worditem struct {
	Word  		string `json:"word"`
	Memonum     string  `json:"memonum"`
}

// list of all of the books
var wordlist = make([]worditem, 0)

// a custom type that we can use for handling errors and formatting responses
type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// here we could do some prep work before calling the handler if we wanted to

	// call the actual handler
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.Message)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		//http.Error(w, "Internal server error. Check the logs.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

func listwords(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	return wordlist, nil
}

func getWords(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {

	log.Printf("entrato in get words")
	// mux.Vars grabs variables from the path

	var payload worditem


	param := mux.Vars(r)["id"]
	id, e := strconv.Atoi(param)
	if e != nil {
		return nil, &handlerError{ e, "Id should be an integer", http.StatusBadRequest}
	}

	if id > 99 {
		return nil, &handlerError{nil, "Could not find mnemonic word for number " + param+ ". Try again entering a number between 0 and 99.", http.StatusBadRequest}
	}

	log.Printf("id is %d", id)

	words, numWords := FindWords(id)
		  	
	//fmt.Println("words=",words)
	//fmt.Printf("%s = ",mapNumbers[iii])

	var b string

	for  i:=0; i< numWords; i++ {
		    	b= b + fmt.Sprintf("%s,", words[i])
	} 	

	payload.Word = b
	payload.Memonum = fmt.Sprintf("%d",id)

	log.Println(payload)
	return payload, nil


}


// end of new code

// Initialize the arrays with constant data
func init() {
	
    log.Printf("MEMWWEB- (C) Massimo Messina \n")

	aConsonanti = append (aConsonanti, "Z" , "L", "N", "M", "R" , "F", "B", "T", "G", "P", "C", "D", "H", "Q", "S", "V", "K", "Y", "W", "X", "J")
	aDiscard = make ([]string, len(aConsonanti))
	fmt.Println(aConsonanti)

}


func main() {

	var srvItalian string
    var srvIPaddr string

	// Initialize arrays with constants
	//InitializeArrays()
	

	// command line flags
    flag.StringVar(&srvItalian, "italian", "Italian.txt", "File name containing Italian Dictionary")
    flag.StringVar(&srvIPaddr, "ip", "127.0.0.1", "IP address on which to listen")
	port := flag.Int("port", 8080, "port to serve on")
	dir := flag.String("directory", "web/", "directory of web files")
	flag.Parse()
	

	log.Printf("Server starting with italian:%s\n", srvItalian)

	aLines, err := readLines(srvItalian)
  	
  	if err != nil {
    	log.Fatalf("readLines: %s", err)
  	}
  	
  	//for i, line := range aLines {
    //	fmt.Println(i, line)
  	//}

  	log.Printf("memwords - read %d lines\n", len(aLines))

  	err = CreateNumbersMap()

    
  	// Look for all numbers between 0 and 99
  	for iii:=0; iii< 100; iii++ {

		    words, numWords := FindWords(iii)
		  	
		    //fmt.Println("words=",words)
		  	fmt.Printf("%s = ",mapNumbers[iii])

		  	for  i:=0; i< numWords; i++ {
		    	fmt.Printf("%s,", words[i])
		    	// bootstrap some data
				wordlist = append(wordlist, worditem{words[i], mapNumbers[iii]})
		  	}
		  	fmt.Println(" ")  	
	}

	log.Println(wordlist)

  	log.Printf("memwords - Initialization terminated without errors\n")


  	// handle all requests by serving a file of the same name
	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)

	// setup routes
	router := mux.NewRouter()
	router.Handle("/", http.RedirectHandler("/static/", 302))
	router.Handle("/words", handler(listwords)).Methods("GET")
	router.Handle("/words/{id}", handler(getWords)).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	http.Handle("/", router)

	//addr := fmt.Sprintf("127.0.0.1:%d", *port)
	addr := fmt.Sprintf("%s:%d", srvIPaddr, *port)
    log.Printf("Activated listening on %s\n", addr)

	// this call blocks -- the progam runs here forever
	err = http.ListenAndServe(addr, nil)
	fmt.Println(err.Error())
}


