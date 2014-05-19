/*
MoREST - Simplistic, universal mongodb driver
Copyright (c) 2014 Andrea Masi
*/
package main

import (
	"flag"
	"fmt"
	"github.com/eraclitux/morest/external/mgo"
	"github.com/eraclitux/morest/morest"
	"log"
	"net/http"
)

var mongoAddressFlag = flag.String("a", "localhost", "Mongodb address. Can be a list of server in cluster like xxx")
var portFlag = flag.Int("p", 9002, "Port to listen for requests")

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("An error occurred: %s", r)
		}
	}()
	flag.Parse()
	msession, err := mgo.Dial(*mongoAddressFlag)
	if err != nil {
		//Deferred function are not run becuse os.Exit(1) is called in the end
		log.Fatalf("Unable to connect to Mongodb: %s", err)
	}
	defer msession.Close()
	http.HandleFunc("/", morest.MakeMainHandler(msession))
	http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil)
}
