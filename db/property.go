package db

import "fmt"

type Property struct {
	Key   string
	Value interface{}
}

func (prop Property) GetMatchQueryString() string {
	return fmt.Sprintf("%s:$MATCH%s", prop.Key, prop.Key)
}

func (prop Property) GetMatchQueryParameter() (string, interface{}) {
	return "$MATCH" + prop.Key, prop.Value
}

func (prop Property) GetQueryAssignString() string {
	return fmt.Sprintf("%s=$SET%s", prop.Key, prop.Key)
}

func (prop Property) GetQueryAssignParameter() (string, interface{}) {
	return "$SET" + prop.Key, prop.Value
}

//func (p Property) valueForQueryString() string {
//	t := reflect.TypeOf(p.Value).Kind()
//	switch t {
//	case reflect.String:
//		out, ok := p.Value.(string)
//		if ok != true {
//			panic("Something went wrong in type comparison!") // this should be a string, see encapsulating code/if statement!
//		}
//		return fmt.Sprintf("'%s'", out)
//	case reflect.Int:
//		return strconv.Itoa(p.Value.(int))
//	case reflect.Float64:
//
//		break
//	case reflect.Bool:
//		break
//	}
//	// if this is a string, we need to wrap it in quotation marks for neo4j query
//	if reflect.TypeOf(p.Value).Kind() == reflect.String {
//
//	}
//	// if not string, then we need to convert it to string anyway but NOT wrap it in quotes
//	out, ok := p.value.(string)
//	if ok != true {
//		log.Println(p.value)
//		panic("Looks like something very strange has been passed into the property value!")
//	}
//	return out
//}
//
//func (p Property) asMatchString() string {
//	return fmt.Sprintf("%s:%s", p.key, p.valueForQueryString())
//}
//
//func (p Property) asAssignmentString() string {
//	return fmt.Sprintf("%s=%s", p.key, p.valueForQueryString())
//}
