package morest

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/eraclitux/morest/external/mgo"
	"log"
)


type testCase struct {
	Req            *http.Request
	//To test Decode method
	expectedResult *mongoRequest
	Err            error
	//To test MakeMainHandler 
	ExpectedJson map[string]interface{}
}

var testCases []*testCase

func buildTestCases() []*testCase {
	cases := []*testCase{}
	case1 := testCase{}
	case1.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find({name:pippo}).sort().limit(5)",
	}
	case1.Err = fmt.Errorf("Error for invalid json formatting")
	cases = append(cases, &case1)

	case2 := testCase{}
	case2.Req = &http.Request{
		Method:     "POST",
		RequestURI: "/testing-db.testing-collection.find({\"name\":\"pippo\"}).sort().limit(5)",
	}
	case2.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case2)

	case3 := testCase{}
	case3.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.fund({\"name\":\"pippo\"}).sort().limit(5)",
	}
	case3.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case3)

	case4 := testCase{}
	case4.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find({\"name\":\"Pippo-0\", \"num\": 0}).sort().limit(5)",
	}
	case4Glob := make(map[string]interface{})
	case4Glob["name"] = "Pippo-0"
	case4Glob["num"] = float64(0)
	case4.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:case4Glob,
		SubAction1:"sort", SubArgs1:"", SubAction2:"limit", SubArgs2:"5",
	}
	cases = append(cases, &case4)

	case5 := testCase{}
	case5.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection",
	}
	case5.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case5)

	case6 := testCase{}
	case6.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find().sort().limit(5)",
	}
	case6Glob := make(map[string]interface{})
	case6.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:case6Glob,
		SubAction1:"sort", SubArgs1:"", SubAction2:"limit", SubArgs2:"5",
	}
	cases = append(cases, &case6)
	case7 := testCase{}
	case7.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find().limit(5).sort(\"name\", \"surname\")",
	}
	case7Glob := make(map[string]interface{})
	case7.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:case7Glob,
		SubAction1:"limit", SubArgs1:"5", SubAction2:"sort", SubArgs2:"\"name\", \"surname\"",
	}
	cases = append(cases, &case7)

	case8 := testCase{}
	case8.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.insert({\"name\":\"mario\",\"num\":42})",
	}
	case8Glob := make(map[string]interface{})
	case8Glob["name"] = "mario"
	case8Glob["num"] = float64(42)
	case8.expectedResult = &mongoRequest{Database:"testing-db", Collection:"testing-collection", Action:"insert", Args1:case8Glob}
	case9 := testCase{}
	case9.Req = &http.Request{
		Method:     "DELETE",
		RequestURI: "/testing-db.testing-collection.remove({\"name\":\"mario\",\"num\":42},{\"param\":1})",
	}
	case9Glob := make(map[string]interface{})
	case9Glob["name"] = "mario"
	case9Glob["num"] = float64(42)
	case9.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"remove", Args1:case8Glob,
	}
	cases = append(cases, &case9)

	case10 := testCase{}
	case10.Req = &http.Request{
		Method:     "DELETE",
		RequestURI: "/testing-db.testing-collection.update({\"name\":\"mario\",\"num\":42},{\"param\":1},{\"param\":3})",
	}
	case10Glob := make(map[string]interface{})
	case10Glob["name"] = "mario"
	case10Glob["num"] = float64(42)
	case10.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"update", Args1:case8Glob,
	}
	cases = append(cases, &case10)

	return cases
}

type dummyMongoData struct {
	Name string
	Num int
}
//Fill test db with dummy data
func arrangeDB(session *mgo.Session) {
	session.DB("testing-db").DropDatabase()
	for i := 0; i < 100; i++ {
		dummy := dummyMongoData{}
		dummy.Name = fmt.Sprintf("Pippo-%d", i)
		dummy.Num = i
		session.DB("testing-db").C("testing-collection").Insert(dummy)
	}
}

func compareJsonResponses(r string, expect map[string]interface{}) bool {
	resp := map[string]interface{}{}
	err := json.Unmarshal([]byte(r), &resp)
	if err != nil {
		return false
	}
	//We cannot predict _id value so drop it
	delete(resp, "_id")
	return compareMapInterfaces(resp, expect)
}

//This test needs mongodb running @ localhost
func TestMakeMainHandler (t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fail()
			log.Printf("Panic in TestMakeMainHandler: %s. Maybe mongodb unreacheble?", r)
		}
	}()
	msession, err := mgo.Dial("localhost")
	if err != nil {
		log.Print("Error connecting to Mongodb ", err)
	}
	defer msession.Close()
	arrangeDB(msession)

	handler := MakeMainHandler(msession)
	for _, singleCase := range testCases {
		if singleCase.Err != nil {
			continue
		}
		recorder := httptest.NewRecorder()
		handler(recorder, singleCase.Req)
		if !compareJsonResponses(recorder.Body.String(), singleCase.ExpectedJson) {
			t.Fail()
		}
	}
}
func compareMapInterfaces(o, p map[string]interface{}) bool {
	//TODO add more types
	//No, really, where is my Python?
	if len(o) != len(p) {
		return false
	}
	for k, v := range o {
		switch vv := v.(type) {
		case string:
			//fmt.Println(k, "is string", vv)
			if vv != p[k].(string) {
				return false
			}
		case float64:
			//fmt.Println(k, "is float64", vv)
			if vv != p[k].(float64) {
				return false
			}
		case []interface{}:
			//fmt.Println(k, "is an array:")
			for i, u := range vv {
				fmt.Println(i, u)
			}
		default:
			fmt.Printf("%v is of a type I don't know how to handle %T(%v)\n", k, v, v)
			return false
		}
	}
	return true
}
func TestDecode(t *testing.T) {
	for i, singleCase := range testCases {
		testStruct := mongoRequest{}
		err := testStruct.Decode(singleCase.Req)
		if singleCase.Err != nil && err != nil {
			continue
		} else if err != nil {
			fmt.Println("[FAIL] unexpected error")
			fmt.Println(singleCase.Req.RequestURI)
			fmt.Println(err)
			t.Fail()
		} else if !(testStruct.Database == singleCase.expectedResult.Database &&
			testStruct.Collection == singleCase.expectedResult.Collection &&
			testStruct.Action == singleCase.expectedResult.Action &&
			testStruct.SubAction1 == singleCase.expectedResult.SubAction1 &&
			testStruct.SubArgs1 == singleCase.expectedResult.SubArgs1 &&
			testStruct.SubAction2 == singleCase.expectedResult.SubAction2 &&
			testStruct.SubArgs2 == singleCase.expectedResult.SubArgs2 &&
			compareMapInterfaces(testStruct.Args1, singleCase.expectedResult.Args1)) {
			fmt.Println("===============")
			fmt.Println("Checking case", i+1)
			fmt.Println(singleCase.Req.RequestURI)
			fmt.Printf("[FAIL] %s\n", singleCase.Req.RequestURI)
			fmt.Printf("Expected: %+v\n", *singleCase.expectedResult)
			fmt.Printf("Got: %+v\n", testStruct)
			t.Fail()
		}
	}
}

type decodeSortCase struct {
	Case     string
	Expected []string
}

func compareSortSlices(o, p []string) bool {
	if len(o) != len(p) {
		return false
	}
	for i, v := range o {
		if v != p[i] {
			return false
		}
	}
	return true
}
func TestDecodeSortArgs(t *testing.T) {
	//TODO add $natural case
	cases := []decodeSortCase{}
	oneCase := decodeSortCase{"{\"name\":-1,\"age\":1}", []string{"-name", "age"}}
	cases = append(cases, oneCase)
	for _, singleCase := range cases {
		got := decodeSortArgs(singleCase.Case)
		if !compareSortSlices(got, singleCase.Expected) {
			fmt.Printf("expected: %+v\n", singleCase)
			fmt.Printf("got: %+v\n", got)
			t.Fail()
		}
	}
}
func init() {
	testCases = buildTestCases()
}
