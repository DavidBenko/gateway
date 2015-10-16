//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package bleve

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const metaFilename = "index_meta.json"

type indexMeta struct {
	Storage string                 `json:"storage"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

func newIndexMeta(storage string, config map[string]interface{}) *indexMeta {
	return &indexMeta{
		Storage: storage,
		Config:  config,
	}
}

func openIndexMeta(path string) (*indexMeta, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrorIndexPathDoesNotExist
	}
	indexMetaPath := indexMetaPath(path)
	metaBytes, err := ioutil.ReadFile(indexMetaPath)
	if err != nil {
		return nil, ErrorIndexMetaMissing
	}
	var im indexMeta
	err = json.Unmarshal(metaBytes, &im)
	if err != nil {
		return nil, ErrorIndexMetaCorrupt
	}
	return &im, nil
}

func (i *indexMeta) Save(path string) (err error) {
	indexMetaPath := indexMetaPath(path)
	// ensure any necessary parent directories exist
	err = os.Mkdir(path, 0700)
	if err != nil {
		return ErrorIndexPathExists
	}
	metaBytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	indexMetaFile, err := os.OpenFile(indexMetaPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			return ErrorIndexPathExists
		}
		return err
	}
	defer func() {
		if ierr := indexMetaFile.Close(); err == nil && ierr != nil {
			err = ierr
		}
	}()
	_, err = indexMetaFile.Write(metaBytes)
	if err != nil {
		return err
	}
	return nil
}

func indexMetaPath(path string) string {
	return path + string(os.PathSeparator) + metaFilename
}
