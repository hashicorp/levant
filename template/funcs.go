// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Masterminds/sprig/v3"
	spewLib "github.com/davecgh/go-spew/spew"
	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

// funcMap builds the template functions and passes the consulClient where this
// is required.
func funcMap(consulClient *consul.Client) template.FuncMap {
	r := template.FuncMap{
		"consulKey":          consulKeyFunc(consulClient),
		"consulKeyExists":    consulKeyExistsFunc(consulClient),
		"consulKeyOrDefault": consulKeyOrDefaultFunc(consulClient),
		"env":                envFunc(),
		"fileContents":       fileContents(),
		"loop":               loop,
		"parseBool":          parseBool,
		"parseFloat":         parseFloat,
		"parseInt":           parseInt,
		"parseJSON":          parseJSON,
		"parseUint":          parseUint,
		"replace":            replace,
		"timeNow":            timeNowFunc,
		"timeNowUTC":         timeNowUTCFunc,
		"timeNowTimezone":    timeNowTimezoneFunc(),
		"toLower":            toLower,
		"toUpper":            toUpper,

		// Maths.
		"add":      add,
		"subtract": subtract,
		"multiply": multiply,
		"divide":   divide,
		"modulo":   modulo,

		// Case Helpers
		"firstRuneToUpper": firstRuneToUpper,
		"firstRuneToLower": firstRuneToLower,
		"runeToUpper":      runeToUpper,
		"runeToLower":      runeToLower,

		//debug.
		"spewDump":   spewDump,
		"spewPrintf": spewPrintf,
	}
	// Add the Sprig functions to the funcmap
	for k, v := range sprig.FuncMap() {
		// if there is a name conflict, favor sprig and rename original version
		if origFun, ok := r[k]; ok {
			if name, err := firstRuneToUpper(k); err == nil {
				name = "levant" + name
				log.Debug().Msgf("template/funcs: renaming \"%v\" function to \"%v\"", k, name)
				r[name] = origFun
			} else {
				log.Error().Msgf("template/funcs: could not add \"%v\" function. error:%v", k, err)
			}
		}
		r[k] = v
	}
	r["sprigVersion"] = sprigVersionFunc

	return r
}

// SprigVersion contains the semver of the included sprig library
// it is used in command/version and provided in the sprig_version
// template function
const SprigVersion = "3.1.0"

func sprigVersionFunc() func(string) (string, error) {
	return func(s string) (string, error) {
		return SprigVersion, nil
	}
}

func consulKeyFunc(consulClient *consul.Client) func(string) (string, error) {
	return func(s string) (string, error) {

		if len(s) == 0 {
			return "", nil
		}

		kv, _, err := consulClient.KV().Get(s, nil)
		if err != nil {
			return "", err
		}

		if kv == nil {
			return "", errors.New("Consul KV not found")
		}

		v := string(kv.Value[:])
		log.Info().Msgf("template/funcs: using Consul KV variable with key %s and value %s",
			s, v)

		return v, nil
	}
}

func consulKeyExistsFunc(consulClient *consul.Client) func(string) (bool, error) {
	return func(s string) (bool, error) {

		if len(s) == 0 {
			return false, nil
		}

		kv, _, err := consulClient.KV().Get(s, nil)
		if err != nil {
			return false, err
		}

		if kv == nil {
			return false, nil
		}

		log.Info().Msgf("template/funcs: found Consul KV variable with key %s", s)

		return true, nil
	}
}

func consulKeyOrDefaultFunc(consulClient *consul.Client) func(string, string) (string, error) {
	return func(s, d string) (string, error) {

		if len(s) == 0 {
			log.Info().Msgf("template/funcs: using default Consul KV variable with value %s", d)
			return d, nil
		}

		kv, _, err := consulClient.KV().Get(s, nil)
		if err != nil {
			return "", err
		}

		if kv == nil {
			log.Info().Msgf("template/funcs: using default Consul KV variable with value %s", d)
			return d, nil
		}

		v := string(kv.Value[:])
		log.Info().Msgf("template/funcs: using Consul KV variable with key %s and value %s",
			s, v)

		return v, nil
	}
}

func loop(ints ...int64) (<-chan int64, error) {
	var start, stop int64
	switch len(ints) {
	case 1:
		start, stop = 0, ints[0]
	case 2:
		start, stop = ints[0], ints[1]
	default:
		return nil, fmt.Errorf("loop: wrong number of arguments, expected 1 or 2"+
			", but got %d", len(ints))
	}

	ch := make(chan int64)

	go func() {
		for i := start; i < stop; i++ {
			ch <- i
		}
		close(ch)
	}()

	return ch, nil
}

func parseBool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}

	result, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return result, nil
}

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0.0, nil
	}

	result, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func parseInt(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	result, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func parseJSON(s string) (interface{}, error) {
	if s == "" {
		return map[string]interface{}{}, nil
	}

	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseUint(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}

	result, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func timeNowFunc() string {
	return time.Now().Format("2006-01-02T15:04:05Z07:00")
}

func timeNowUTCFunc() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
}

func timeNowTimezoneFunc() func(string) (string, error) {
	return func(t string) (string, error) {

		if t == "" {
			return "", nil
		}

		loc, err := time.LoadLocation(t)
		if err != nil {
			return "", err
		}

		return time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), nil
	}
}

func toLower(s string) (string, error) {
	return strings.ToLower(s), nil
}

func toUpper(s string) (string, error) {
	return strings.ToUpper(s), nil
}

func envFunc() func(string) (string, error) {
	return func(s string) (string, error) {
		if s == "" {
			return "", nil
		}
		return os.Getenv(s), nil
	}
}

func fileContents() func(string) (string, error) {
	return func(s string) (string, error) {
		if s == "" {
			return "", nil
		}
		contents, err := ioutil.ReadFile(s)
		if err != nil {
			return "", err
		}
		return string(contents), nil
	}
}

func add(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float(), nil
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("add: unknown type for %q (%T)", av, a)
	}
}

func subtract(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() - int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() - float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float(), nil
		default:
			return nil, fmt.Errorf("subtract: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("subtract: unknown type for %q (%T)", av, a)
	}
}

func multiply(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() * int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() * float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float(), nil
		default:
			return nil, fmt.Errorf("multiply: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("multiply: unknown type for %q (%T)", av, a)
	}
}

func divide(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() / int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() / float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() / float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() / bv.Float(), nil
		default:
			return nil, fmt.Errorf("divide: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("divide: unknown type for %q (%T)", av, a)
	}
}

func modulo(b, a interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() % int64(bv.Uint()), nil
		default:
			return nil, fmt.Errorf("modulo: unknown type for %q (%T)", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() % bv.Uint(), nil
		default:
			return nil, fmt.Errorf("modulo: unknown type for %q (%T)", bv, b)
		}
	default:
		return nil, fmt.Errorf("modulo: unknown type for %q (%T)", av, a)
	}
}

func firstRuneToUpper(s string) (string, error) {
	return runeToUpper(s, 0)
}

func runeToUpper(inString string, runeIndex int) (string, error) {
	return funcOnRune(unicode.ToUpper, inString, runeIndex)
}

func firstRuneToLower(s string) (string, error) {
	return runeToLower(s, 0)
}

func runeToLower(inString string, runeIndex int) (string, error) {
	return funcOnRune(unicode.ToLower, inString, runeIndex)
}

func funcOnRune(inFunc func(rune) rune, inString string, runeIndex int) (string, error) {
	if !utf8.ValidString(inString) {
		return "", errors.New("funcOnRune: not a valid UTF-8 string")
	}

	runeCount := utf8.RuneCountInString(inString)

	if runeIndex > runeCount-1 || runeIndex < 0 {
		return "", fmt.Errorf("funcOnRune: runeIndex out of range (max:%v, provided:%v)", runeCount-1, runeIndex)
	}
	runes := []rune(inString)
	transformedRune := inFunc(runes[runeIndex])

	if runes[runeIndex] == transformedRune {
		return inString, nil
	}
	runes[runeIndex] = transformedRune
	return string(runes), nil
}

func spewDump(a interface{}) (string, error) {
	return spewLib.Sdump(a), nil
}

func spewPrintf(format string, args ...interface{}) (string, error) {
	return spewLib.Sprintf(format, args), nil
}
