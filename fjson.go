package gsjson

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	json "github.com/json-iterator/go"
	"github.com/sonnt85/gosutils/endec"
	"github.com/sonnt85/gosutils/lockedfile"

	// "github.com/sonnt85/gosutils/slogrus"
	"github.com/sonnt85/gosystem"
)

type Fjson struct {
	path       string
	RmdirFlag  bool
	ForceInit  bool
	passphrase []byte
	w          chan struct{}
	*Ejson
}

//Remove pid file
func (f *Fjson) Remove() (err error) {
	// var ul func()
	// if ul, err = lockedfile.LockTimeout(file.path, 0); err != nil {
	// 	return
	// }
	// defer ul()
	if f.RmdirFlag {
		return os.RemoveAll(filepath.Dir(f.path))
	} else {
		return os.Remove(f.path)
	}
}

func (f *Fjson) GetJsonFilePath() string {
	return f.path
}

func (f *Fjson) ReadFile() (jsonByte []byte, err error) {
	var ul func()
	ul, err = lockedfile.RLockTimeout(f.path, time.Second*30, time.Millisecond*100)
	if err != nil {
		return
	}
	if ul != nil {
		defer ul()
	}
	return f._readFile()
}

func (f *Fjson) GetPassPhrase() []byte {
	return f.passphrase
}

func (f *Fjson) _readFile() (jsonbyte []byte, err error) {
	if len(f.passphrase) == 0 {
		jsonbyte, err = ioutil.ReadFile(f.path)
	} else {
		jsonbyte, err = endec.DecryptFileToBytes(f.path, f.passphrase)
		if f.ForceInit && err != nil {
			// if err.Error() == "cipher: message authentication failed" {
			// 	return err
			// }{
			jsonbyte = []byte("{}")
			err = nil
		}
	}
	if err == nil {
		if !json.Valid(jsonbyte) {
			if !f.ForceInit {
				err = errors.New("efile is invalid")
			} else {
				jsonbyte = []byte("{}")
			}
		}
	}
	return
}

// New creates a Fjson file using the specified path.
func NewFjson(path string, passphrase []byte, removeIfFileInvalid bool, datas ...map[string]interface{}) (f *Fjson, err error) {
	if len(path) == 0 {
		return nil, errors.New("path is empty")
	}
	f = new(Fjson)
	// f.buffer = make([]byte, 0)

	f.path = path
	f.ForceInit = removeIfFileInvalid
	f.w = make(chan struct{}, 1)

	if len(passphrase) == 1 && passphrase[0] == 0 {
		passphrase = []byte(f.path)
	}
	f.passphrase = passphrase
	var jsonByte []byte

	if len(datas) != 0 {
		for k, v := range datas[0] {
			f.Set(k, v)
		}
	}

	if gosystem.FileIsExist(f.path) {
		if jsonByte, err = f.ReadFile(); err != nil {
			return nil, err
		}
	} else {
		if err = touchFilePublic(f.path); err != nil {
			return
		}
	}
	if len(jsonByte) == 0 {
		jsonByte = []byte("{}")
	}
	f.Ejson, err = NewGsjson(jsonByte, func() {
		f.w <- struct{}{}
	}, removeIfFileInvalid)
	if err != nil {
		return
	}
	f.w <- struct{}{}
	go func() {
		if watcher, err := fsnotify.NewWatcher(); err == nil {
			defer watcher.Close()
			if err = watcher.Add(f.path); err != nil {
				return
			}
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						if jsonByte, err = f.ReadFile(); err == nil {
							f.LoadNewData(jsonByte)
						}
						continue
					} else if event.Op&fsnotify.Remove != fsnotify.Remove {
						continue
					}

				case _, ok := <-f.w:
					if !ok {
						return
					}
				}
				watcher.Remove(f.path)

				for {
					if len(f.w) != 0 {
						<-f.w
						time.Sleep(time.Millisecond * 10)
					} else {
						break
					}
				}
				f.WriteTo(f.path, f.passphrase)
				watcher.Add(f.path)

			}
		}
	}()
	// runonce.NewRunOnceFuncPort(0)
	return
}
