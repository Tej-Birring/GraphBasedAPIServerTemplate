package utils

import (
	"github.com/dchest/uniuri"
	"reflect"
)

func RandomString(length int) string {
	return uniuri.NewLen(length)
}

//func SumMaps(maps ...map[interface{}]interface{}) map[interface{}]interface{} {
//	ret := map[interface{}]interface{}{} // create new map to return
//	for _, m := range maps {             // for each map
//		for k, v := range m { // add key-val pair to return value
//			ret[k] = v
//		}
//	}
//	return ret // return the result
//}

func SumMaps(maps ...map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{} // create new map to return
	for _, m := range maps {        // for each map
		for k, v := range m { // add key-val pair to return value
			ret[k] = v
		}
	}
	return ret // return the result
}

func StructToGenericMap(in interface{}) map[string]interface{} {
	retVal := map[string]interface{}{}

	v := reflect.ValueOf(in)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		key := (typeOfS.Field(i)).Name
		val := v.Field(i).Interface()

		// for debug:
		//fmt.Println(key, reflect.TypeOf(key).Kind())
		//fmt.Println(val, reflect.TypeOf(val).Kind())

		// uncomment if needed:
		// special case: always omit if empty (zero val)
		//if v.Field(i).IsZero() {
		//	continue
		//}
		
		retVal[key] = val
	}

	return retVal
}
