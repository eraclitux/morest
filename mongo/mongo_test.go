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
		RequestURI: "/database.collection.find({\"name\":\"pippo\", \"number\":5}).sort().limit(5)",
	}
	case4Glob := make(map[string]interface{})
	case4Glob["name"] = "pippo"
	case4Glob["number"] = float64(5)
	case4.expectedResult = &mongoRequest{"database", "collection", "find", case4Glob, "sort", "", 5}
	cases = append(cases, &case4)
	return cases

}
func compareMapInterfaces(o, p *map[string]interface{}) bool {
	//No, really, where is my Python?
	for k, v := range *o {
		switch vv := v.(type) {
		case string:
			fmt.Println(k, "is string", vv)
			if vv != (*p)[k].(string) {
				return false
			}
		case float64:
			fmt.Println(k, "is int", vv)
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
		fmt.Println("Checking case", i+1)
		fmt.Println("===============")
		testStruct := mongoRequest{}
		err := testStruct.Decode(singleCase.Req)
		if singleCase.Err != nil && err != nil {
			fmt.Println("[OK] got the expected error")
			fmt.Println(err)
		} else if err != nil {
			fmt.Println("[FAIL] unexpected error")
			fmt.Println(err)
			t.Fail()
		} else if !(testStruct.Database == singleCase.expectedResult.Database &&
			testStruct.Collection == singleCase.expectedResult.Collection &&
			testStruct.Action == singleCase.expectedResult.Action &&
			testStruct.SubAction == singleCase.expectedResult.SubAction &&
			compareMapInterfaces(&testStruct.Args, &singleCase.expectedResult.Args)) {
			fmt.Printf("[FAIL] %s\n", singleCase.Req.RequestURI)
			fmt.Printf("Expected: %+v\n", *singleCase.expectedResult)
			fmt.Printf("Got: %+v\n", testStruct)
			t.Fail()
		}
	}
}
