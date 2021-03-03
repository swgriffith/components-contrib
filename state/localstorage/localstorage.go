// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation and Dapr Contributors.
// Licensed under the MIT License.
// ------------------------------------------------------------

/*
Local Storage State Store.

Sample configuration in yaml:

	apiVersion: dapr.io/v1alpha1
	kind: Component
	metadata:
	name: statestore
	spec:
	type: state.localstorage
	version: v0
	metadata:
	- name: hostPath
		value: "<local path>"

*/

package localstorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dapr/components-contrib/state"
	"github.com/dapr/dapr/pkg/logger"
	jsoniter "github.com/json-iterator/go"
)

const (
	keyDelimiter = "||"
	hostPath     = "hostPath"
)

// StateStore Type
type StateStore struct {
	state.DefaultBulkStore
	json     jsoniter.API
	hostPath string

	logger logger.Logger
}

type blobStorageMetadata struct {
	hostPath string
}

// Init the connection to blob storage, optionally creates a blob container if it doesn't exist.
func (r *StateStore) Init(metadata state.Metadata) error {
	meta, err := getStorageMetadata(metadata.Properties)
	if err != nil {
		return err
	}

	r.hostPath = meta.hostPath
	r.logger.Debugf("using host path '%s'", meta.hostPath)

	return nil
}

// Delete the state
func (r *StateStore) Delete(req *state.DeleteRequest) error {
	r.logger.Debugf("delete %s", req.Key)

	return r.deleteFile(req)
}

// Get the state
func (r *StateStore) Get(req *state.GetRequest) (*state.GetResponse, error) {
	r.logger.Debugf("fetching %s", req.Key)
	data, err := r.readFile(req)
	if err != nil {
		r.logger.Debugf("error %s", err)

		return &state.GetResponse{}, err
	}

	return &state.GetResponse{
		Data: data,
	}, err
}

// Set the state
func (r *StateStore) Set(req *state.SetRequest) error {
	r.logger.Debugf("saving %s", req.Key)

	return r.writeFile(req)
}

// NewLocalStorageStore instance
func NewLocalStorageStore(logger logger.Logger) *StateStore {
	s := &StateStore{
		json:   jsoniter.ConfigFastest,
		logger: logger,
	}
	s.DefaultBulkStore = state.NewDefaultBulkStore(s)

	return s
}

func getStorageMetadata(metadata map[string]string) (*blobStorageMetadata, error) {
	meta := blobStorageMetadata{}

	if val, ok := metadata[hostPath]; ok && val != "" {
		meta.hostPath = val
	} else {
		return nil, fmt.Errorf("missing or empty %s field from metadata", hostPath)
	}

	return &meta, nil
}

func (r *StateStore) readFile(req *state.GetRequest) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(r.hostPath, getFileName(req.Key)))
	if err != nil {
		r.logger.Debugf("read file %s, err %s", req.Key, err)

		return nil, err
	}

	return data, nil
}

func (r *StateStore) writeFile(req *state.SetRequest) error {

	f, err := os.Create(filepath.Join(r.hostPath, getFileName(req.Key)))
	if err != nil {
		r.logger.Debugf("write file %s, err %s", req.Key, err)
	}

	defer f.Close()

	_, err = f.Write(r.marshal(req))
	if err != nil {
		r.logger.Debugf("write file %s, err %s", req.Key, err)
	}

	return nil
}

func (r *StateStore) deleteFile(req *state.DeleteRequest) error {

	err := os.Remove(filepath.Join(r.hostPath, getFileName(req.Key)))
	if err != nil {
		r.logger.Debugf("delete file %s, err %s", req.Key, err)
	}

	return nil
}

func getFileName(key string) string {
	pr := strings.Split(key, keyDelimiter)
	if len(pr) != 2 {
		return pr[0]
	}

	return pr[1]
}

func (r *StateStore) marshal(req *state.SetRequest) []byte {
	var v string
	b, ok := req.Value.([]byte)
	if ok {
		v = string(b)
	} else {
		v, _ = jsoniter.MarshalToString(req.Value)
	}

	return []byte(v)
}
