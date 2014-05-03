/*
MoREST - an ingenuous Mongodb REST proxy
Copyright (c) 2014 Andrea Masi
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/eraclitux/morest/external/mgo"
	"github.com/eraclitux/morest/external/mgo/bson"
	"log"
	"net/http"
	"strings"
	"flag"
)

var mongoAddressFlag = flag.String("a", "localhost", "Mongodb address. Can be a list of server in cluster like xxx")
var portFlag = flag.Int("p", 9002, "Port to listen for request")
var debugFlag = flag.Bool("d", false, "Enable debug")

//model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	Action     string
	Args       string
	SubAction  string //action' params
}

func (s *mongoRequest) Decode(r *http.Request) {
	r.ParseForm()
	mongoQuery := strings.Split(r.RequestURI, "/")[1]
	for i, v := range strings.Split(mongoQuery, ".") {
		fmt.Println(i, v)
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		case 2:
			//mongodb function to permform on data (find, insert etc)
			action := strings.Split(v, "(")
			s.Action = action[0]
			s.Args = strings.Trim(action[1], ")")
		case 3:
			//sub action (es sort, limit)
			s.SubAction = v
		}
	}
	fmt.Println(s)
}

func (s *mongoRequest) Execute(msession *mgo.Session) ([]byte, error) {
	session := msession.Copy()
	defer session.Close()
	coll := session.DB(s.Database).C(s.Collection)
	gdata := new([]interface{})
	coll.Find(bson.M{"my-key": "my_value"}).Limit(6).All(gdata)
	jdata, err := json.Marshal(gdata)
	if err != nil {
		return nil, err
	}
	return jdata, nil
}

func mainHandler(msession *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if *debugFlag {
			log.Println("[DEBUG] Request struct:", r)
		}
		mReq := mongoRequest{}
		mReq.Decode(r)
		jdata, _ := mReq.Execute(msession)
		fmt.Fprintf(w, "Data: %v\n", jdata, string(jdata))
	}
}

func main() {
	flag.Parse()
	msession, err := mgo.Dial(*mongoAddressFlag)
	if err != nil {
		log.Panic("Unable to connect to Mongodb ", err)
	}
	defer msession.Close()
	http.HandleFunc("/", mainHandler(msession))
	http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil)
}
