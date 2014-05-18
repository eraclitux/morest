package morest

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
var supportedSubActions = []string{}

//Model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	//mydb.mycoll.action(args1, args2, args3)
	Action     string
	//REF convert into one slice
	Args1       map[string]interface{}
	Args2       map[string]interface{}
	Args3       map[string]interface{}
	//REF convert into slices
	SubAction1 string
	SubArgs1   string
	SubAction2 string
	SubArgs2   string
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
	isSupported = false
	for _, v := range supportedSubActions {
		if s.SubAction1 == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.SubAction1)
	}
	isSupported = false
	for _, v := range supportedSubActions {
		if s.SubAction2 == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.SubAction2)
	}
	switch r.Method {
	case "POST":
		if s.Action == "find" {
			return fmt.Errorf("Action requested not coherent with http method")
		}
	}
	return nil
}

//Help decode mongodb functions and arguments
//This is used for find, insert, update
func getActionArgs(s string) (action string, args1, args2, args3 map[string]interface{}, er error) {
	argsPointerSlice := []*map[string]interface{}{&args1, &args2, &args3}
	actiond := strings.Split(s, "(")
	action = actiond[0]
	args := strings.Trim(actiond[1], ")")
	if len(args) == 0 {
		return
	}
	argsSlice := strings.SplitAfter(args, "},")
	for i, v := range argsSlice {
		v = strings.TrimRight(v, ",")
		//Here we dont need to args1.assert(map[string]interface{})
		//becuse we dont pass interface{} to Unmarshal but the right type
		err := json.Unmarshal([]byte(v), argsPointerSlice[i])
		if err != nil {
			er = err
			return
		}
	}
	return
}

//Help decode mongodb functions and its arguments
//This is used for sort, limit 
func getSubActionArgs(s string) (action, args string) {
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
	} else if len(parameters) > 5 {
		return fmt.Errorf("Too much arguments")
	}
	for i, v := range parameters {
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		//mongodb main function (find, insert etc)
		case 2:
			var err error
			s.Action, s.Args1, s.Args2, s.Args3, err = getActionArgs(v)
			if err != nil {
				return err
			}
		//sub action (es sort, limit)
		case 3:
			s.SubAction1, s.SubArgs1 = getSubActionArgs(v)
		//sub action (es sort, limit)
		case 4:
			s.SubAction2, s.SubArgs2 = getSubActionArgs(v)
		}
	}
	if DEBUG {
		log.Printf("[DEBUG] %+v\n", s)
	}
	return s.Check(r)
}

//Decode json sort argumets to be passed to mgo Sort()
func decodeSortArgs(s string) []string {
	//FIXME parse $natural key
	returnValue := []string{}
	glob := new(map[string]interface{})
	if len(s) != 0 {
		err := json.Unmarshal([]byte(s), glob)
		if err != nil {
			return []string{""}
		}
		for k, v := range *glob {
			i := v.(float64)
			if int(i) < 0 {
				k := "-" + k
				returnValue = append(returnValue, k)
			} else {
				returnValue = append(returnValue, k)
			}
		}
	}
	return returnValue
}

//Setup the query to exucute primary action
func bakeAction(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	switch s.Action {
	case "find":
		*queryP = coll.Find(s.Args1)
		return nil
	case "count":
		return nil
	case "insert":
		return nil
	default:
		return fmt.Errorf("Unable to execute %s", s.Action)
	}
}

//Prepares the query to exucute secondary actions
func bakeSubActions(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	//TODO parse SubAction2
	//No subactions on these
	if s.Action == "count" ||
	   s.Action == "insert" ||
	   s.Action == "remove" {
		return nil
	}
	if s.SubAction1 != "" {
		switch s.SubAction1 {
		case "sort":
			*queryP = queryP.Sort(decodeSortArgs(s.SubArgs1)...)
		case "limit":
			num, err := strconv.Atoi(s.SubArgs1)
			if err != nil {
				return fmt.Errorf("Unable to convert limit argument")
			} else {
				*queryP = queryP.Limit(num)
			}
		default:
			return fmt.Errorf("Unable to execute %s", s.SubAction1)
		}
	}
	if s.SubAction2 != "" {
		switch s.SubAction2 {
		case "sort":
			*queryP = queryP.Sort(decodeSortArgs(s.SubArgs2)...)
			return nil
		case "limit":
			num, err := strconv.Atoi(s.SubArgs2)
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
//Exectute query on mongodb
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
		err := coll.Insert(s.Args1)
		if err != nil {
			return []byte{}, err
		}
		return []byte("{\"nInserted\":1}"), nil
	case "remove":
		//This removes a single document
		err := coll.Remove(s.Args1)
		if err != nil {
			return []byte{}, err
		}
		return []byte("{\"nRemoved\":1}"), nil
	case "count":
		n, err := coll.Count()
		if err != nil {
			return []byte{}, err
		}
		number := strconv.Itoa(n)
		return []byte(number), nil
	default:
		return []byte{}, fmt.Errorf("Unable to execute %s", s.Action)
	}
}

//Performs decoded action on mongodb
func (s *mongoRequest) Execute(msession *mgo.Session) ([]byte, error) {
	//FIXME add session to mongoRequest struct?
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
		fmt.Fprintf(w, "%s\n", string(jdata))
	}
}

func init() {
	//To check against user requests
	supportedActions = []string{"find", "insert", "remove", "count"}
	supportedSubActions = []string{"sort", "limit", ""}
}
