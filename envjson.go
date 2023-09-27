package gsjson

import (
	"os"
	"sync"

	"github.com/sonnt85/gosutils/endec"
	"github.com/sonnt85/gosutils/sutils"
)

type EnvJson struct {
	password string
	envname  string
	sync.RWMutex
	ej *Ejson
}

var defaultENVNAME = "ENVJSON"
var defaultPwd = "ENVJSON"

var gnevjson *EnvJson

func Init() {
	gnevjson = NewEnvJson(defaultENVNAME, defaultPwd)
}

func NewEnvJson(envkey string, pwd string) *EnvJson {
	dd := EnvJson{
		password: defaultPwd,
		envname:  defaultENVNAME,
	}
	if envkey != "" {
		dd.envname = envkey
	}
	if pwd != "" {
		dd.password = pwd
	}

	ej, _ := NewGsjson([]byte(""), nil, true)
	dd.ej = ej
	return &dd
}

func GetGenvjson() *EnvJson {
	return gnevjson
}

func (dd *EnvJson) GetDecryptedEnvValue() string {
	dd.RLock()
	bs := dd.getDecryptedEnvValue()
	dd.RUnlock()
	return bs
}
func GetDecryptedEnvValue() string {
	return gnevjson.GetDecryptedEnvValue()
}

func (dd *EnvJson) GetDecryptedEnvValueThenAddEnv(m map[string]string) (ejsondata string) {
	dd.RLock()
	ejsondata = dd.getDecryptedEnvValue()
	dd.RUnlock()
	for key, value := range m {
		ejsondata, _ = sutils.JsonSet(ejsondata, key, value)
	}
	return
}

func GetDecryptedEnvValueThenAddEnv(m map[string]string) (ejsondata string) {
	return gnevjson.GetDecryptedEnvValueThenAddEnv(m)
}

func (dd *EnvJson) GetEncryptedEnvValueThenAddEnv(m map[string]string) (ejsondata string) {
	dd.RLock()
	jsondata := dd.getDecryptedEnvValue()
	dd.RUnlock()
	// var ejsondata string
	for key, value := range m {
		jsondata, _ = sutils.JsonSet(jsondata, key, value)
	}
	ejsondata, _ = endec.EncrypBytesToString([]byte(jsondata), []byte(dd.password))
	return
}

func GetEncryptedEnvValueThenAddEnv(m map[string]string) (ejsondata string) {
	return gnevjson.GetEncryptedEnvValueThenAddEnv(m)
}

func (dd *EnvJson) SetEnvFromDecryptedValue(value string) {
	dd.RLock()
	if ejsondata, err := endec.EncrypBytesToString([]byte(value), []byte(dd.password)); err == nil {
		os.Setenv(dd.envname, ejsondata)
	}
	dd.RUnlock()
}

func SetEnvFromDecryptedValue(value string) {
	gnevjson.SetEnvFromDecryptedValue(value)
}

func (dd *EnvJson) GetEncryptedEnvValue() string {
	dd.RLock()
	bs := os.Getenv(dd.envname)
	dd.RUnlock()
	return bs
}

func GetEncryptedEnvValue() string {
	return gnevjson.GetEncryptedEnvValue()
}

func (dd *EnvJson) GetEnvName() string {
	dd.RLock()
	bs := dd.envname
	dd.RUnlock()
	return bs
}

func GetEnvName() string {
	return gnevjson.GetEnvName()
}

func (dd *EnvJson) Decode(encenv string) string {
	dd.RLock()
	bs, _ := endec.DecryptBytesFromString(encenv, []byte(dd.password))
	dd.RUnlock()
	return string(bs)
}

func Decode(encenv string) string {
	return gnevjson.Decode(encenv)
}

func (dd *EnvJson) String() string {
	return dd.GetDecryptedEnvValue()
}

func String() string {
	return gnevjson.String()
}

func (dd *EnvJson) getDecryptedEnvValue() string {
	bs, _ := endec.DecryptBytesFromString(os.Getenv(dd.envname), []byte(dd.password))
	return string(bs)
}

func (dd *EnvJson) Getenv(key string) (value string) {
	data := dd.GetDecryptedEnvValue()
	value, _ = sutils.JsonStringFindElement(&data, key)
	return
}

func Getenv(key string) (value string) {
	return gnevjson.Getenv(key)
}

func (dd *EnvJson) Unsetenv(key string) (err error) {
	var ejsondata string
	dd.Lock()
	jsondata := dd.getDecryptedEnvValue()
	if jsondata, err = sutils.JsonDelete(jsondata, key); err == nil {
		if ejsondata, err = endec.EncrypBytesToString([]byte(jsondata), []byte(dd.password)); err == nil {
			os.Setenv(dd.envname, ejsondata)
		}
	}
	dd.Unlock()
	return
}

func Unsetenv(key string) (err error) {
	return gnevjson.Unsetenv(key)
}

func (dd *EnvJson) Hasenv(key string) bool {
	data := dd.GetDecryptedEnvValue()
	_, err := sutils.JsonStringFindElement(&data, key)
	return err == nil
}

func Hasenv(key string) bool {
	return gnevjson.Hasenv(key)
}

func (dd *EnvJson) GetOrCreateEnv(key, defaultValue string) (retstr string) {
	var err error
	data := dd.GetDecryptedEnvValue()
	retstr, err = sutils.JsonStringFindElement(&data, key)
	if err != nil {
		if err = dd.Setenv(key, defaultValue); err == nil {
			retstr = defaultValue
		}
	}
	return
}

func GetOrCreateEnv(key, defaultValue string) (retstr string) {
	return gnevjson.GetOrCreateEnv(key, defaultValue)
}

func (dd *EnvJson) Setenv(key, value string) (err error) {
	var ejsondata string
	dd.Lock()
	jsondata := dd.getDecryptedEnvValue()
	if jsondata, err = sutils.JsonSet(jsondata, key, value); err == nil {
		if ejsondata, err = endec.EncrypBytesToString([]byte(jsondata), []byte(dd.password)); err == nil {
			os.Setenv(dd.envname, ejsondata)
		}
	}
	dd.Unlock()
	return
}

func Setenv(key, value string) (err error) {
	return gnevjson.Setenv(key, value)
}
