package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Dictionary struct {
	file *os.File
}

func checkword(dictword string, lettersavailable string, desiredlength int, w http.ResponseWriter, hint string) {
	//fmt.Println("checkword...")

	// convert to lower case
	// PERFORMANCE OPTIMISATION
	// we ensure that the source file is all lower case
	// we convert the user input to lower case before calling checkword.
	// this allows us to avoid checking them both inside this function,
	// which increases performance.

	//dictword = strings.ToLower(dictword)
	//lettersavailable = strings.ToLower(lettersavailable)

	// TODO: strip whitespace from the hint e.g. in case it has a trailing space
	hint = strings.TrimSpace(hint)

	// TODO: if the hint given differs in length to the desired length, ignore it and display an error
	if len(hint) != desiredlength {
		// we have an invalid hint, so set it to empty
		hint = ""
	}

	// map to hold letters from the dictionary word we read in
	m := make(map[string]int)

	// map to hold letters from the letters we have
	m2 := make(map[string]int)

	for _, letter := range dictword {
		// add the letter
		m[string(letter)]++
	}

	for _, yourletter := range lettersavailable {
		m2[string(yourletter)]++
	}

	wordlength := len(dictword)
	numberoflettersfound := 0

	//fmt.Println("testing word: ", dictword)
	//fmt.Printf("it is %d letters in length\n", wordlength)

	for _, c := range dictword {
		if m2[string(c)] > 0 {
			//fmt.Println("DEBUG: we have a: ", string(c))
			numberoflettersfound++
			m2[string(c)]--
		}

	}

	//fmt.Printf("we found %d of the %d letters in the word\n", numberoflettersfound, wordlength)
	if numberoflettersfound == wordlength {
		if wordlength == desiredlength {
			// only print words of a length we specified
			//fmt.Fprintf(w, "%s<br>", dictword)

			checkHint(dictword, hint, w)

		}
	}
}

func checkHint(dictword string, hint string, w http.ResponseWriter) {
	/*
	   1. count the number of not null characters in the hint
	   2. iterate over the hint, comparing letters (skip nulls)
	   3. if the number of characters we find == the number of not null characters in the hint, we have an exact match
	   4. else it is not a match
	*/

	// convert to lower case
	dictword = strings.ToLower(dictword)

	if len(hint) > 0 {
		hint = strings.ToLower(hint)

		//wordlength := len(dictword)
		numberoflettersweneedtomatch := 0
		numberofletterswehavematched := 0

		// figure out how many non-null letters are in the hint
		for _, c := range hint {
			if c != '.' {
				numberoflettersweneedtomatch++
			}
		}

		for i, ch := range hint {
			if ch != '.' {
				if hint[i] == dictword[i] {
					numberofletterswehavematched++
				}

			}
		}

		if numberofletterswehavematched == numberoflettersweneedtomatch {
			// bold an exact match
			fmt.Fprintf(w, "<b>%s</b><br>", dictword)
		}
	}

	if len(hint) == 0 {
		// show non matches for now, eventually we will hide them or make them a different color or something
		fmt.Fprintf(w, "%s<br>", dictword)
	}

}

// func hello(w http.ResponseWriter, r *http.Request) {

func (dictionary *Dictionary) mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fmt.Println("DEBUG: Path was: ", r.URL.Path)
		//http.Error(w, "404 not found.", http.StatusNotFound)
		//return
	}

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, "form.html")
		case "/favicon.ico":
			fmt.Println("DEBUG: served: ", r.URL.Path)
			http.ServeFile(w, r, "static/favicon.ico")
		case "/manifest.webmanifest":
			fmt.Println("DEBUG: served: ", r.URL.Path)
			http.ServeFile(w, r, "static/manifest.webmanifest")
		case "/images/icon-512x512.png":
			http.ServeFile(w, r, "images/icon-512x512.png")
		case "/images/icon-192x192.png":
			http.ServeFile(w, r, "images/icon-192x192.png")
		default:
			fmt.Println("DEBUG: served: ", r.URL.Path)
			http.Error(w, "404 not found.", http.StatusNotFound)
		}

	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		letters := r.FormValue("letters")
		templength := r.FormValue("length")
		hint := r.FormValue("hint")
		intlength, _ := strconv.Atoi(templength)

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "letters", Value: letters, Expires: expiration}
		http.SetCookie(w, &cookie)

		// read in the file
		//file, err := os.Open("test.txt")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//defer file.Close()

		// render output as html
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html lang=\"en-AU\">")
		fmt.Fprintf(w, "<head>")

		headValues := `
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
		<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
		<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
		`

		fmt.Fprintf(w, headValues)
		fmt.Fprintf(w, "</head>")
		fmt.Fprintf(w, "<body>")
		fmt.Fprintf(w, "<div class=\"container\">")
		fmt.Fprintf(w, "<h1>Results</h1>")
		fmt.Fprintf(w, "<p>")

		letters = strings.ToLower(letters)

		// we need to ensure we seek back to the start of the file, since a previous execution will have left it at the EOF
		// TODO: suspicious this may not be concurrency safe?
		_, _ = dictionary.file.Seek(0, 0)
		scanner := bufio.NewScanner(dictionary.file)
		var wg sync.WaitGroup
		start := time.Now()
		for scanner.Scan() {
			word := scanner.Text()
			wg.Add(1)
			//fmt.Printf("DEBUG: word from dict: %s\n", word)
			go func() {
				defer wg.Done()
				checkword(word, letters, intlength, w, hint)
			}()

			//fmt.Fprintf(w, "<br>")
		}
		wg.Wait()
		finish := time.Now()
		duration := finish.Sub(start)
		fmt.Fprintf(w, "<br>")
		fmt.Fprintf(w, "duration: %f seconds.\n", duration.Seconds())
		fmt.Fprintf(w, "<br><br>")
		fmt.Fprintf(w, "<a href=https://www.aaash.com/>Try another word</a>")

		fmt.Fprintf(w, "</div>")
		fmt.Fprintf(w, "</body>")
		fmt.Fprintf(w, "</html>")

		// done! :)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {

	dictionary := Dictionary{}
	// read in the file (just once, to save performance)
	// TODO: embed this file in the executable
	// TODO: strip out words > 7 characters in length
	dictionary.file, _ = os.Open("test.txt")
	defer dictionary.file.Close()

	//mux := http.NewServeMux()
	http.HandleFunc("/", dictionary.mainHandler)

	fileServer := http.FileServer(http.Dir("./static/"))

	http.Handle("/static/", http.StripPrefix("/static", fileServer))

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
