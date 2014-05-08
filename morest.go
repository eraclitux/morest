/*
MoREST - an ingenuous Mongodb RESTful proxy
Copyright (c) 2014 Andrea Masi
*/
package main

import (
	"flag"
	"fmt"
	"github.com/eraclitux/morest/external/mgo"
	"github.com/eraclitux/morest/mongo"
	"log"
	"net/http"
)

var mongoAddressFlag = flag.String("a", "localhost", "Mongodb address. Can be a list of server in cluster like xxx")
var portFlag = flag.Int("p", 9002, "Port to listen for requests")

func main() {
	flag.Parse()
	msession, err := mgo.Dial(*mongoAddressFlag)
	if err != nil {
		log.Panic("Unable to connect to Mongodb ", err)
	}
	defer msession.Close()
	http.HandleFunc("/", mongo.MainHandler(msession))
	http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil)
}
