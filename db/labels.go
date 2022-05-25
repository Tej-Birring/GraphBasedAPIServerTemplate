package db

type Label = string
type Labels []string

func (ll Labels) ToString(varPrefix string) string {
	ret := ""
	if len(ll) < 1 {
		// nothing
	} else if len(ll) == 1 {
		ret = ":" + ll[0]
	} else {
		ret = ":" + ll[0]
		for _, l := range ll {
			ret += ":" + l
		}
	}
	if len(varPrefix) > 0 {
		return varPrefix + ret
	} else {
		return ret
	}
}
