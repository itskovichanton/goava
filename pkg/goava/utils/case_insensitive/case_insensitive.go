package case_insensitive

import "strings"

func Set(m map[string]interface{}, key string, value interface{}) {
	klower := strings.ToLower(key)
	m[klower] = value
	if klower != key {
		delete(m, key)
	}
}

func Get(m map[string]interface{}, key string) interface{} {
	r, ok := m[key]
	if ok {
		return r
	}
	r, ok = m[strings.ToLower(key)]
	if ok {
		return r
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return nil
}
