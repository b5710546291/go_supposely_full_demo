package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
	stomp "github.com/go-stomp/stomp"
)

type MyCustomHandler struct {
	conn *stomp.Conn
}

func (handler *MyCustomHandler) rootEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("ROOT HIT")
	fmt.Fprintf(w, `
	HOW TO
		1. POST TO /checkNumber WITH DATA "command"="csr", "number"="XXXXXXXXXX" * X is any number
		2. GET TO /getLog
	`)
}

func (handler *MyCustomHandler) checkNumberRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	number := r.FormValue("number")
	command := r.FormValue("command")
	if number != "" && len(number) == 10 && regexp.MustCompile(`^[0-9]+$`).MatchString(number) && command != "" && len(command) == 3 {
		sub, err := handler.conn.Subscribe("/topic/response/"+command+number, stomp.AckAuto)
		if err != nil {
			panic(err.Error())
		}
		go func() {
			err := handler.conn.Send(
				"/topic/request", // destination
				"text/plain",     // content-type
				[]byte(fmt.Sprintf("%s%s", command, number))) // body
			if err != nil {
				panic(err.Error())
			}
		}()
		resp := <-sub.C
		var sresp string = string(resp.Body)
		log.Println(sresp)
		fmt.Fprintf(w, sresp)

		err = sub.Unsubscribe()
		if err != nil {
			panic(err.Error())
		}
	} else {
		fmt.Fprintf(w, "Invalid data")
	}
}

func (handler *MyCustomHandler) getLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sub, err := handler.conn.Subscribe("/topic/logres/", stomp.AckAuto)
	if err != nil {
		panic(err.Error())
	}
	go func() {
		err := handler.conn.Send(
			"/topic/logreq", // destination
			"text/plain",    // content-type
			[]byte("log"))   // body
		if err != nil {
			panic(err.Error())
		}
	}()
	resp := <-sub.C
	var sresp string = string(resp.Body)
	log.Println("log request done")
	fmt.Fprintf(w, sresp)

	err = sub.Unsubscribe()
	if err != nil {
		panic(err.Error())
	}
}

func handleRequests(handler *MyCustomHandler) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", handler.rootEndpoint)
	myRouter.HandleFunc("/checkNumber", handler.checkNumberRequest).Methods("POST")
	myRouter.HandleFunc("/getLog", handler.getLog).Methods("GET")
	log.Fatal(http.ListenAndServe(":9001", myRouter))
}

func main() {
	fmt.Println("Demo redis sql activemq")
	defer func() {
		log.Println("Exist")
	}()
	conn, err := stomp.Dial("tcp", "localhost:61613", stomp.ConnOpt.HeartBeat(0, 0))
	if err != nil {
		fmt.Println(err)
	}

	myhandler := &MyCustomHandler{conn: conn}
	handleRequests(myhandler)

}
