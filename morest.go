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

var mongoAddressFlag = flag.String("mongodb-address", "localhost", "Mongodb address. Can be a list of server in a cluster.")
var tlsCertFlag = flag.String("ssl-cert", "", "Path to certificate file")
var tlsKeyFlag = flag.String("ssl-key", "", "Path to key file")
var portFlag = flag.Int("port", 9002, "Port to listen for requests.")
var safeFlag = flag.Bool("safe-mode", true, "When false, MongoDB does not acknowledge the receipt of write operations. Faster but may lead to data loss.")

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("An error occurred: %s", r)
		}
	}()
	flag.Parse()
	msession, err := mgo.Dial(*mongoAddressFlag)
	if !*safeFlag {
		//msession.SetSafe(&mgo.Safe{WTimeout:100})
		msession.SetSafe(nil)
	}
	if err != nil {
		//Deferred functions are not run becuse os.Exit(1) is called in the end
		log.Fatalf("Unable to connect to Mongodb: %s", err)
	}
	defer msession.Close()
	http.HandleFunc("/", morest.MakeMainHandler(msession))
	if *tlsKeyFlag == "" && *tlsCertFlag == "" {
		http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil)
	} else if *tlsKeyFlag != "" && *tlsCertFlag != "" {
		http.ListenAndServeTLS(fmt.Sprintf(":%d", *portFlag), *tlsCertFlag, *tlsKeyFlag, nil)
	} else {
		log.Fatalf("Invalid ssl options")
	}
}
