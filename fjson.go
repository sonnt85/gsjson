package gsjson

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	json "github.com/json-iterator/go"
	"github.com/sonnt85/gosutils/endec"
	"github.com/sonnt85/gosutils/lockedfile"

	"github.com/sonnt85/gosutils/slogrus"
	"github.com/sonnt85/gosystem"
)

type Fjson struct {
	path       string
	RmdirFlag  bool `json:"RmdirFlag"`
	ForceInit  bool
	passphrase []byte
	w          chan struct{}
	cmd        chan int
	*Ejson
}

// Remove pid file
func (f *Fjson) Remove() (err error) {
	// var ul func()
	// if ul, err = lockedfile.LockTimeout(file.path, 0); err != nil {
	// 	return
	// }
	// defer ul()
	if f.RmdirFlag {
		err = os.RemoveAll(filepath.Dir(f.path))
	} else {
		err = os.Remove(f.path)
	}
	if err == nil {
		f.SendCmd(0)
	}
	return
}

func (f *Fjson) SendCmd(cmd int) {
	f.cmd <- cmd
}

func (f *Fjson) GetJsonFilePath() string {
	return f.path
}

func (f *Fjson) ReadFile(paths ...string) (jsonByte []byte, err error) {
	var ul func()
	path := f.path
	if len(paths) != 0 {
		path = paths[0]
	}
	var fl *lockedfile.File
	fl, ul, err = lockedfile.RLockTimeout(path, time.Second*30, time.Millisecond*100)

	if err != nil {
		return
	}
	if ul != nil {
		defer ul()
	}
	return f._readFile(fl)
}

func (f *Fjson) GetPassPhrase() []byte {
	return f.passphrase
}

//paths is pathfile or io.Reader interface

func (f *Fjson) _readFile(paths ...interface{}) (jsonbyte []byte, err error) {
	var r io.ReadSeeker
	// path := f.path
	if len(paths) != 0 {
		switch v := paths[0].(type) {
		case io.ReadSeeker:
			r = v
			r.Seek(0, 0)
		case string:
			jsonbyte, err = ioutil.ReadFile(v)
		default:
			return jsonbyte, errors.New("param is of unknown type")
		}
	} else {
		if f, e := os.Open(f.path); e != nil {
			return jsonbyte, e
		} else {
			defer f.Close()
			r = io.ReadSeeker(f)
		}
	}

	if len(f.passphrase) == 0 {
		jsonbyte, err = io.ReadAll(r) // ioutil.ReadFile(path)
	} else {
		jsonbyte, err = endec.DecryptFileToBytes(r, f.passphrase)
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

func DecodeJsonFileNoErr(path string, passphrase []byte) (retstr string) {
	retstr, _ = DecodeJsonFile(path, passphrase)
	return
}

func DecodeJsonFile(path string, passphrase []byte) (retstr string, err error) {
	if len(path) == 0 {
		return "", errors.New("path is empty")
	}
	f := new(Fjson)
	// f.buffer = make([]byte, 0)

	f.path = path
	if len(passphrase) == 1 && passphrase[0] == 0 {
		passphrase = []byte(f.path)
	}
	f.passphrase = passphrase
	var jsonByte []byte

	if gosystem.FileIsExist(f.path) {
		if jsonByte, err = f.ReadFile(); err != nil {
			return
		} else {
			return string(jsonByte), nil
		}
	} else {
		return "", errors.New("file is not exits")
	}
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
	f.cmd = make(chan int, 1)

	if len(passphrase) == 1 && passphrase[0] == 0 {
		passphrase = []byte(f.path)
	}
	f.passphrase = passphrase
	var jsonByte []byte

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
						var jsonByte []byte
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
				case i, ok := <-f.cmd:
					if ok {
						if i == 0 {
							return
						} else if i == 1 {
							f.WriteTo(f.path+".dec", []byte{})
						} else if i == 2 {
							if gosystem.FileIsExist(f.path + ".load") {
								if jsonbyte, err := ioutil.ReadFile(f.path + ".load"); err == nil {
									f.LoadNewData(jsonbyte)
								}
							}
						} else if i == 3 {
							break
						}
					}
					continue
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
				// slogrus.Info(f.String())
				if err = f.WriteTo(f.path, f.passphrase); err != nil {
					slogrus.TraceStack(err)
				}
				if gosystem.FileIsExist(f.path + ".dec") {
					f.WriteTo(f.path+".dec", []byte{})
				} else if gosystem.FileIsExist(f.path + ".load") {
					if jsonbyte, err := ioutil.ReadFile(f.path + ".load"); err == nil {
						f.LoadNewData(jsonbyte)
					}
				}
				watcher.Add(f.path)

			}
		}
	}()
	if len(datas) != 0 {
		for k, v := range datas[0] {
			f.Set(k, v)
		}
	} else {
		f.w <- struct{}{}
	}
	// runonce.NewRunOnceFuncPort(0)
	return
}
