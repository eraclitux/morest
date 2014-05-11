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

//Mongodb supported actions. 
//Declared global to save some memory
var supportedActions = []string{}

//Model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	Action     string
	Args       map[string]interface{}
	SubAction1  string
	SubArgs1    string
	SubAction2 string
	SubArgs2    string
}

//Check if decoded action is sopported and coherent with http method
func (s *mongoRequest) Check(r *http.Request) error {
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

//Help decode mongodb functions (find, insert etc) and its arguments
func getActionArgs(s string) (action, args string) {
	actiond := strings.Split(s, "(")
	action = actiond[0]
	args = strings.Trim(actiond[1], ")")
	return
}

func (s *mongoRequest) Decode(r *http.Request) error {
	mongoQuery := strings.Split(r.RequestURI, "/")[1]
	parameters := strings.Split(mongoQuery, ".")
	if len(parameters) < 3 {
		return fmt.Errorf("Too few arguments")
	}
	for i, v := range parameters {
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		//mongodb main function (find, insert etc)
		case 2:
			var argsString string
			s.Action, argsString = getActionArgs(v)
			if len(argsString) != 0 {
				//At this point you should already miss python very much
				err := json.Unmarshal([]byte(argsString), &s.Args)
				if err != nil {
					return err
				}
			}
		//sub action (es sort, limit)
		case 3:
			s.SubAction1, s.SubArgs1 = getActionArgs(v)
		//limit function
		case 4:
			s.SubAction2, s.SubArgs2 = getActionArgs(v)
		}
	}
	if DEBUG {
		log.Printf("[DEBUG] %+v\n", s)
	}
	return s.Check(r)
}

//Decode json sort argumets to be passed to mgo' Sort()
func decodeSortArgs(s string) []string {
	//FIXME decode json 
	s = strings.Replace(s, "\"", "",-1)
	s = strings.Replace(s, " ", "",-1)
	return strings.Split(s, ",")
}
//Prepares the query to exucute primary action
func bakeAction(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	switch s.Action {
	case "find":
		*queryP = coll.Find(s.Args)
		return nil
	case "insert":
		return nil
	default:
		return fmt.Errorf("Unable to execute %s", s.Action)
	}
}
//Prepares the query to exucute secondary actions
func bakeSubActions(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	if s.SubAction1 != "" {
		switch s.SubAction1 {
		case "sort":
			*queryP = queryP.Sort(decodeSortArgs(s.SubArgs1)...)
			return nil
		case "limit":
			num, err := strconv.Atoi(s.SubArgs1)
			if err != nil {
				return fmt.Errorf("Unable to convert limit argument")
			} else {
				*queryP = queryP.Limit(num)
				return nil
			}
		default:
			return fmt.Errorf("Unable to execute %s", s.SubAction1)
		}
	}
	return nil
}
func executeQuery(query *mgo.Query, s *mongoRequest, coll *mgo.Collection) ([]byte, error) {
	gdata := new([]interface{})
	switch s.Action {
	case "find":
		err := query.All(gdata)
		if err != nil {
			return []byte{}, err
		}
		return json.Marshal(gdata)
	case "insert":
		return []byte{}, coll.Insert(s.Args)
	default:
		return []byte{}, fmt.Errorf("Unable to execute %s", s.Action)
	}
}
//Perform decoded action on mongodb
func (s *mongoRequest) Execute(msession *mgo.Session) ([]byte, error) {
	//TODO add session to mongoRequest struct?
	session := msession.Copy()
	defer session.Close()
	coll := session.DB(s.Database).C(s.Collection)
	query := new(mgo.Query)
	bakeAction(&query, s, coll)
	bakeSubActions(&query, s, coll)
	jdata, err := executeQuery(query, s, coll)
	if err != nil {
		return nil, err
	}
	return jdata, nil
}

func MakeMainHandler(msession *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEBUG {
			log.Printf("[DEBUG] Request.URL struct: %+v\n", r.URL)
			log.Printf("[DEBUG] Request struct: %+v\n", r)
		}
		mReq := mongoRequest{}
		err := mReq.Decode(r)
		if err != nil {
		        http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jdata, err := mReq.Execute(msession)
		if err != nil {
		        http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Data: %v\n", string(jdata))
	}
}

func init() {
	supportedActions = []string{"find", "insert"}
	//supportedSubActions := []string{"sort", "limit"}
}
