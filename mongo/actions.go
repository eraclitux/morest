package mongo

import (
	"encoding/json"
	"fmt"
	"github.com/eraclitux/morest/external/mgo"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const DEBUG = true

//Model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	Action     string
	Args       map[string]interface{}
	SubAction  string
	SubArgs    string
	Limit      int
}

//Check if decoded action is sopported and coherent
func (s *mongoRequest) Check(r *http.Request) error {
	supportedActions := []string{"find", "insert"}
	isSupported := false
	for _, v := range supportedActions {
		if s.Action == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.Action)
	}
	switch r.Method {
	case "POST":
		if s.Action == "find" {
			return fmt.Errorf("Action requested not coherent with http method")
		}
	}
	return nil
}

//Help decode mongodb desired function (find, insert etc) and its arguments
func getActionArgs(s string) (action, args string) {
	actiond := strings.Split(s, "(")
	action = actiond[0]
	args = strings.Trim(actiond[1], ")")
	return
}

func (s *mongoRequest) Decode(r *http.Request) error {
	mongoQuery := strings.Split(r.RequestURI, "/")[1]
	for i, v := range strings.Split(mongoQuery, ".") {
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		case 2:
			//mongodb function to permform on data (find, insert etc)
			var argsString string
			s.Action, argsString = getActionArgs(v)
			//At this point you should already miss python very much
			err := json.Unmarshal([]byte(argsString), &s.Args)
			//s.Args = s.Args.(map[string]interface{})
			if err != nil {
				return err
			}
			fmt.Println(s.Args)
		case 3:
			//sub action (es sort, limit)
			s.SubAction, s.SubArgs = getActionArgs(v)
		case 4:
			action := strings.Split(v, "(")
			s.Limit, _ = strconv.Atoi(strings.Trim(action[1], ")"))
		}
	}
	if DEBUG {
		fmt.Printf("%+v\n", s)
	}
	return s.Check(r)
}

//Perform decoded action on mongodb
func (s *mongoRequest) Execute(msession *mgo.Session) ([]byte, error) {
	session := msession.Copy()
	defer session.Close()
	coll := session.DB(s.Database).C(s.Collection)
	gdata := new([]interface{})
	err := coll.Find(s.Args).Limit(6).All(gdata)
	if err != nil {
		return nil, err
	}
	jdata, err := json.Marshal(gdata)
	if err != nil {
		return nil, err
	}
	return jdata, nil
}

func MainHandler(msession *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEBUG {
			log.Printf("[DEBUG] Request.URL struct: %+v\n", r.URL)
			log.Printf("[DEBUG] Request struct: %+v\n", r)
		}
		mReq := mongoRequest{}
		mReq.Decode(r)
		jdata, _ := mReq.Execute(msession)
		fmt.Fprintf(w, "Data: %v\n", string(jdata))
	}
}
