package template

import (
	"errors"
	"fmt"
	"text/template"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

// funcMap builds the template functions and passes the consulClient where this
// is required.
func funcMap(consulClient *consul.Client) template.FuncMap {
	return template.FuncMap{
		"consulKey":          consulKeyFunc(consulClient),
		"consulKeyExists":    consulKeyExistsFunc(consulClient),
		"consulKeyOrDefault": consulKeyOrDefaultFunc(consulClient),
		"loop":               loop,
		"timeNow":            timeNowFunc,
		"timeNowUTC":         timeNowUTCFunc,
		"timeNowTimezone":    timeNowTimezoneFunc(),
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
