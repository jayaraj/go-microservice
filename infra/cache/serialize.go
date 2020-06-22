package cache

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func serialize(value interface{}) ([]byte, error) {
	if data, ok := value.([]byte); ok {
		return data, nil
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		log.WithFields(log.Fields{
			"value": value,
			"error": err,
		}).Error("Serialize: gob encoding failed")
		return nil, err
	}
	return b.Bytes(), nil
}

func deserialize(byt []byte, ptr interface{}) (err error) {
	if data, ok := ptr.(*[]byte); ok {
		*data = byt
		return
	}

	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		switch p := v.Elem(); p.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var i int64
			i, err = strconv.ParseInt(string(byt), 10, 64)
			if err != nil {
				log.WithFields(log.Fields{
					"value": string(byt),
					"error": err,
				}).Error("Deserialize: failed to parse int")
			} else {
				p.SetInt(i)
			}
			return

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var i uint64
			i, err = strconv.ParseUint(string(byt), 10, 64)
			if err != nil {
				log.WithFields(log.Fields{
					"value": string(byt),
					"error": err,
				}).Error("Deserialize: failed to parse uint")
			} else {
				p.SetUint(i)
			}
			return
		}
	}

	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)
	if err = decoder.Decode(ptr); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Deserialize: glob decoding failed")
		return
	}
	return
}
