package onlineconf

type LazyMod struct {
	StringParams map[string]string
	IntParams    map[string]int

	RawJSONParams            map[string]string // Here will be all JSON params (not parsed)
	MapStringInterfaceParams map[string]map[string]interface{}
	MapIntIntParams          map[string]map[int]int
	MapIntStringParams       map[string]map[int]string
	MapStringIntParams       map[string]map[string]int
	MapStringStringParams    map[string]map[string]string
}
