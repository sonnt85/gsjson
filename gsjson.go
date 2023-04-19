// Package pid provides structure and helper functions to create and remove
package gsjson

import (
	// json "github.com/json-iterator/go"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sonnt85/gosutils/endec"
	"github.com/sonnt85/gosutils/lockedfile"
	"github.com/sonnt85/gosystem"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// var default_passphrase = []byte("default_passphrase")
// Ejson is a file used to store the process ID of a running process.
type Ejson struct {
	buffer []byte
	sync.RWMutex
	hookChange func()
	hookLoad   func(buffer *[]byte)
}

func (gs *Ejson) _delete(path string) (err error) {
	var tmpbytes []byte
	tmpbytes, err = sjson.DeleteBytes(gs.buffer, path)
	if err == nil {
		gs.buffer = tmpbytes
		if gs.hookChange != nil {
			gs.hookChange()
		}
	}
	return
}

func (gs *Ejson) String() string {
	return string(gs.buffer)
}

// Delete deletes the value for a key path.
func (gs *Ejson) Delete(path string) (err error) {
	gs.Lock()
	defer gs.Unlock()
	return gs._delete(path)
}

func (gs *Ejson) _set(path string, value any) (err error) {
	if retbytes, err := sjson.SetBytes(gs.buffer, path, value); err == nil && !bytes.Equal(retbytes, gs.buffer) {
		gs.buffer = retbytes
		if gs.hookChange != nil {
			gs.hookChange()
		}
	}
	return err
}

func (gs *Ejson) _setraw(path string, value []byte) (err error) {
	if retbytes, err := sjson.SetRawBytes(gs.buffer, path, value); err == nil && !bytes.Equal(retbytes, gs.buffer) {
		gs.buffer = retbytes
		if gs.hookChange != nil {
			gs.hookChange()
		}
	}
	return err
}

// set the value for a key.
func (gs *Ejson) Set(path string, value interface{}) (err error) {
	gs.Lock()
	defer gs.Unlock()
	return gs._set(path, value)
}

// set the value for a key.
func (gs *Ejson) SetRaw(path string, value []byte) (err error) {
	gs.Lock()
	defer gs.Unlock()
	return gs._setraw(path, value)
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (gs *Ejson) LoadAndDelete(path string) (value gjson.Result) {
	value = gs.Get(path)
	if value.Exists() {
		gs.Delete(path)
	}
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (gs *Ejson) LoadOrStore(path string, value any) (actual gjson.Result) {
	actual = gs.Get(path)
	if !actual.Exists() {
		if err := gs.Set(path, value); err != nil {
			actual = gjson.Result{}
		}
	}
	return
}

// get the value for a key.
func (gs *Ejson) Get(path string) (result gjson.Result) {
	gs.RLock()
	result = gjson.GetBytes(gs.buffer, path)
	gs.RUnlock()
	return
}

func (gs *Ejson) Loads() (result gjson.Result) {
	gs.RLock()
	result = gjson.ParseBytes(gs.buffer)
	gs.RUnlock()
	return
}

func (gs *Ejson) WriteTo(path string, password []byte) (err error) {
	var ul func()
	if !gosystem.FileIsExist(path) {
		if err = touchFilePublic(path); err != nil {
			return
		}
	}
	var fl *lockedfile.File

	// gosystem.WriteToFileWithLockSFL(path, gs.buffer)
	fl, ul, err = lockedfile.LockTimeout(path, time.Second*30, time.Millisecond*100)
	if err != nil {
		return
	}
	if len(password) == 0 {
		fl.Truncate(0)
		_, err = fl.Write(gs.buffer)
		// err = ioutil.WriteFile(path, gs.buffer, 0766)
	} else {
		err = endec.EncryptBytesToFile(fl, gs.buffer, password)
	}
	gosystem.Chmod(path, 0766)
	if ul != nil {
		ul()
	}
	return
}

func (gs *Ejson) LoadNewData(buffer []byte) (err error) {
	if !gjson.ValidBytes(buffer) {
		err = errors.New("JSON is not valid")
	} else {
		gs.Lock()
		gs.buffer = buffer
		if gs.hookLoad != nil {
			gs.hookLoad(&gs.buffer)
		}
		gs.Unlock()
	}
	return
}
func (gs *Ejson) SetHookLoad(f func(buffer *[]byte)) {
	gs.Lock()
	gs.hookLoad = f
	gs.Unlock()
}

func touchFilePublic(path string) (err error) {
	if !gosystem.DirIsExist(filepath.Dir(path)) {
		os.MkdirAll(filepath.Dir(path), 0755)
	}
	if err = gosystem.TouchFile(path); err == nil {
		gosystem.Chmod(path, 0766)
	}
	return
}

// NewGsjson creates a PID file using the specified path.
func NewGsjson(datas []byte, hookChange func(), forceInit ...bool) (f *Ejson, err error) {
	f = new(Ejson)
	f.buffer = datas
	f.hookChange = hookChange
	if !gjson.ValidBytes(datas) {
		if len(forceInit) != 0 && forceInit[0] {
			f.buffer = []byte("{}")
		} else {
			err = errors.New("JSON is not valid")
		}
	}
	// runonce.NewRunOnceFuncPort(0)
	return
}
