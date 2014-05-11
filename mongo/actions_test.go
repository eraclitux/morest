package mongo

import (
	"fmt"
	"net/http"
	"testing"
)

type testCase struct {
	Req            *http.Request
	expectedResult *mongoRequest
	Err            error
}

func buildCases() []*testCase {
	cases := []*testCase{}
	case1 := testCase{}
	case1.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.find({name:pippo}).sort().limit(5)",
	}
	case1.Err = fmt.Errorf("We expect an error for invalid json formatting")
	//case1.expectedResult = &mongoRequest{"database", "collection", "find", "{name:pippo}", "sort", "", 5}
	cases = append(cases, &case1)
	case2 := testCase{}
	case2.Req = &http.Request{
		Method:     "POST",
		RequestURI: "/database.collection.find({\"name\":\"pippo\"}).sort().limit(5)",
	}
	case2.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case2)
	case3 := testCase{}
	case3.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.fund({\"name\":\"pippo\"}).sort().limit(5)",
	}
	case3.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case3)
	case4 := testCase{}
	case4.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.find({\"name\":\"pippo\", \"number\":42}).sort().limit(5)",
	}
	case4Glob := make(map[string]interface{})
	case4Glob["name"] = "pippo"
	case4Glob["number"] = float64(42)
	case4.expectedResult = &mongoRequest{"database", "collection", "find", case4Glob, "sort", "", "limit", "5"}
	cases = append(cases, &case4)
	case5 := testCase{}
	case5.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection",
	}
	case5.Err = fmt.Errorf("We expect an error")
	cases = append(cases, &case5)
	case6 := testCase{}
	case6.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.find().sort().limit(5)",
	}
	case6Glob := make(map[string]interface{})
	case6.expectedResult = &mongoRequest{"database", "collection", "find", case6Glob, "sort", "", "limit", "5"}
	cases = append(cases, &case6)
	case7 := testCase{}
	case7.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.find().limit(5).sort(\"name\", \"surname\")",
	}
	case7Glob := make(map[string]interface{})
	case7.expectedResult = &mongoRequest{"database", "collection", "find", case7Glob, "limit", "5", "sort", "\"name\", \"surname\""}
	cases = append(cases, &case7)

	case8 := testCase{}
	case8.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/database.collection.insert({\"name\":\"mario\",\"num\":42})",
	}
	case8Glob := make(map[string]interface{})
	case8Glob["name"] = "mario"
	case8Glob["num"] = float64(42)
	case8.expectedResult = &mongoRequest{"database", "collection", "insert", case8Glob, "", "", "", ""}
	cases = append(cases, &case8)
	return cases

}
func compareMapInterfaces(o, p *map[string]interface{}) bool {
	// TODO add more types
	//No, really, where is my Python?
	if len(*o) != len(*p) {
		return false
	}
	for k, v := range *o {
		switch vv := v.(type) {
		case string:
			fmt.Println(k, "is string", vv)
			if vv != (*p)[k].(string) {
				return false
			}
		case float64:
			fmt.Println(k, "is float64", vv)
			if vv != (*p)[k].(float64) {
				return false
			}
		case []interface{}:
			fmt.Println(k, "is an array:")
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
	for i, singleCase := range buildCases() {
		fmt.Println("===============")
		fmt.Println("Checking case", i+1)
		fmt.Println(singleCase.Req.RequestURI)
		testStruct := mongoRequest{}
		err := testStruct.Decode(singleCase.Req)
		if singleCase.Err != nil && err != nil {
			fmt.Println("[TEST][OK] got the expected error")
			fmt.Println(err)
		} else if err != nil {
			fmt.Println("[FAIL] unexpected error")
			fmt.Println(err)
			t.Fail()
		} else if !(testStruct.Database == singleCase.expectedResult.Database &&
			testStruct.Collection == singleCase.expectedResult.Collection &&
			testStruct.Action == singleCase.expectedResult.Action &&
			testStruct.SubAction1 == singleCase.expectedResult.SubAction1 &&
			testStruct.SubArgs1 == singleCase.expectedResult.SubArgs1 &&
			testStruct.SubAction2 == singleCase.expectedResult.SubAction2 &&
			testStruct.SubArgs2 == singleCase.expectedResult.SubArgs2 &&
			compareMapInterfaces(&testStruct.Args, &singleCase.expectedResult.Args)) {
			fmt.Printf("[FAIL] %s\n", singleCase.Req.RequestURI)
			fmt.Printf("Expected: %+v\n", *singleCase.expectedResult)
			fmt.Printf("Got: %+v\n", testStruct)
			t.Fail()
		}
	}
}
