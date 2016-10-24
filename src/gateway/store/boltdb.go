package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gateway/config"
	apsql "gateway/sql"

	"github.com/boltdb/bolt"
)

const (
	boltdbCurrentVersion = 1
	metaBucket           = "meta"
	collectionSequence   = "collectionSequence"
	objectSequence       = "objectSequence"
	versionKey           = "version"
)

type BoltDBStore struct {
	conf           config.Store
	boltdb         *bolt.DB
	listeners      []apsql.Listener
	listenersMutex sync.RWMutex
}

func (s *BoltDBStore) Migrate() error {
	currentVersion := uint64(0)
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	meta, err := tx.CreateBucketIfNotExists([]byte(metaBucket))
	if err != nil {
		return err
	}

	v, migrate := meta.Get([]byte(versionKey)), s.conf.Migrate
	if v == nil {
		err := meta.Put([]byte(versionKey), itob(currentVersion))
		if err != nil {
			return err
		}
		migrate = true
	} else {
		currentVersion = btoi(v)
	}

	if currentVersion == boltdbCurrentVersion {
		return nil
	}

	if !migrate {
		return errors.New("The store is not up to date. Please migrate by invoking with the -store-migrate flag.")
	}

	if currentVersion < 1 {
		err := meta.Put([]byte(versionKey), itob(1))
		if err != nil {
			return err
		}
		_, err = meta.CreateBucketIfNotExists([]byte(collectionSequence))
		if err != nil {
			return err
		}
		_, err = meta.CreateBucketIfNotExists([]byte(objectSequence))
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *BoltDBStore) Clear() error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cursor := tx.Cursor()
	key, _ := cursor.First()
	for key != nil {
		err := tx.DeleteBucket(key)
		if err != nil {
			return err
		}
		key, _ = cursor.Next()
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *BoltDBStore) RegisterListener(l apsql.Listener) {
	defer s.listenersMutex.Unlock()
	s.listenersMutex.Lock()
	s.listeners = append(s.listeners, l)
}

func (s *BoltDBStore) notify(table string, accountID, userID, apiID, proxyEndpointID, id int64,
	event apsql.NotificationEventType, messages ...interface{}) error {
	n := apsql.Notification{
		Table:           table,
		AccountID:       accountID,
		UserID:          userID,
		APIID:           apiID,
		ProxyEndpointID: proxyEndpointID,
		ID:              id,
		Event:           event,
		Tag:             apsql.NotificationTagDefault,
		Messages:        messages,
	}

	defer s.listenersMutex.Unlock()
	s.listenersMutex.Lock()

	for _, listener := range s.listeners {
		listener.Notify(&n)
	}

	return nil
}

func (s *BoltDBStore) ListCollection(collection *Collection, collections *[]*Collection) error {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account := tx.Bucket(itob(uint64(collection.AccountID)))
	if account == nil {
		return nil
	}

	_collections := account.Bucket([]byte("$collections"))
	if _collections == nil {
		return nil
	}

	cursor := _collections.Cursor()
	key, value := cursor.First()
	for key != nil {
		_collection := &Collection{}
		err := json.Unmarshal(value, _collection)
		if err != nil {
			return err
		}
		*collections = append(*collections, _collection)
		key, value = cursor.Next()
	}

	return nil
}

func (s *BoltDBStore) CreateCollection(collection *Collection) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	meta := tx.Bucket([]byte(metaBucket))
	if meta == nil {
		return errors.New("bucket for meta doesn't exist")
	}

	account, err := tx.CreateBucketIfNotExists(itob(uint64(collection.AccountID)))
	if err != nil {
		return err
	}

	collections, err := account.CreateBucketIfNotExists([]byte("$collections"))
	if err != nil {
		return err
	}

	cursor := collections.Cursor()
	key, value := cursor.First()
	for key != nil {
		var c Collection
		err := json.Unmarshal(value, &c)
		if err != nil {
			return err
		}
		if c.Name == collection.Name {
			return ErrCollectionExists
		}
		key, value = cursor.Next()
	}

	{
		sequence := meta.Bucket([]byte(collectionSequence))
		if sequence == nil {
			return errors.New("bucket for store collection sequence doesn't exist")
		}

		key, err := sequence.NextSequence()
		if err != nil {
			return err
		}
		collection.ID = int64(key)

		value, err := json.Marshal(collection)
		if err != nil {
			return err
		}

		err = collections.Put(itob(key), value)
		if err != nil {
			return err
		}

		_, err = account.CreateBucketIfNotExists(itob(key))
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("collections", collection.AccountID, collection.UserID, 0, 0, collection.ID, apsql.Insert)
}

func (s *BoltDBStore) ShowCollection(collection *Collection) error {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account := tx.Bucket(itob(uint64(collection.AccountID)))
	if account == nil {
		return ErrCollectionDoesntExist
	}

	collections := account.Bucket([]byte("$collections"))
	if collections == nil {
		return ErrCollectionDoesntExist
	}

	value := collections.Get(itob(uint64(collection.ID)))
	if value == nil {
		return ErrCollectionDoesntExist
	}

	err = json.Unmarshal(value, collection)
	if err != nil {
		return err
	}

	return nil
}

func (s *BoltDBStore) UpdateCollection(collection *Collection) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := tx.CreateBucketIfNotExists(itob(uint64(collection.AccountID)))
	if err != nil {
		return err
	}

	collections, err := account.CreateBucketIfNotExists([]byte("$collections"))
	if err != nil {
		return err
	}

	value := collections.Get(itob(uint64(collection.ID)))
	if value == nil {
		return ErrCollectionDoesntExist
	}

	value, err = json.Marshal(collection)
	if err != nil {
		return err
	}

	err = collections.Put(itob(uint64(collection.ID)), value)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("collections", collection.AccountID, collection.UserID, 0, 0, collection.ID, apsql.Update)
}

func (s *BoltDBStore) DeleteCollection(collection *Collection) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := tx.CreateBucketIfNotExists(itob(uint64(collection.AccountID)))
	if err != nil {
		return err
	}

	collections, err := account.CreateBucketIfNotExists([]byte("$collections"))
	if err != nil {
		return err
	}

	value := collections.Get(itob(uint64(collection.ID)))
	if value == nil {
		return ErrCollectionDoesntExist
	}
	err = json.Unmarshal(value, collection)
	if err != nil {
		return err
	}

	_, err = s.DeleteTx(tx, false, collection.AccountID, collection.Name, "true")
	if err != nil {
		return err
	}

	err = collections.Delete(itob(uint64(collection.ID)))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("collections", collection.AccountID, collection.UserID, 0, 0, collection.ID, apsql.Delete)
}

func findCollection(collections *bolt.Bucket, collection *Collection) (bool, error) {
	cursor := collections.Cursor()
	key, value := cursor.First()
	if collection.ID != 0 {
		for key != nil {
			var c Collection
			err := json.Unmarshal(value, &c)
			if err != nil {
				return false, err
			}
			if c.ID == collection.ID {
				*collection = c
				return true, nil
			}
			key, value = cursor.Next()
		}
	} else {
		for key != nil {
			var c Collection
			err := json.Unmarshal(value, &c)
			if err != nil {
				return false, err
			}
			if c.Name == collection.Name {
				*collection = c
				return true, nil
			}
			key, value = cursor.Next()
		}
	}
	return false, nil
}

func (s *BoltDBStore) getBucket(tx *bolt.Tx, collection *Collection) (*bolt.Bucket, *bolt.Bucket, error) {
	meta := tx.Bucket([]byte(metaBucket))
	if meta == nil {
		return nil, nil, errors.New("bucket for meta doesn't exist")
	}

	sequence := meta.Bucket([]byte(objectSequence))
	if sequence == nil {
		return nil, nil, errors.New("bucket for store object sequence doesn't exist")
	}

	account := tx.Bucket(itob(uint64(collection.AccountID)))
	if account == nil {
		return nil, nil, errors.New("bucket for account doesn't exist")
	}

	collections := account.Bucket([]byte("$collections"))
	if collections == nil {
		return nil, nil, errors.New("bucket for $collections doesn't exist")
	}

	found, err := findCollection(collections, collection)
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, ErrCollectionDoesntExist
	}

	bucket := account.Bucket(itob(uint64(collection.ID)))
	if bucket == nil {
		return nil, nil, ErrCollectionDoesntExist
	}

	return bucket, sequence, nil
}

func (s *BoltDBStore) createBucket(tx *bolt.Tx, collection *Collection) (*bolt.Bucket, *bolt.Bucket, error) {
	if !tx.Writable() {
		return nil, nil, errors.New("transaction isn't writable")
	}

	meta := tx.Bucket([]byte(metaBucket))
	if meta == nil {
		return nil, nil, errors.New("bucket for meta doesn't exist")
	}

	sequence := meta.Bucket([]byte(objectSequence))
	if sequence == nil {
		return nil, nil, errors.New("bucket for store object sequence doesn't exist")
	}

	account, err := tx.CreateBucketIfNotExists(itob(uint64(collection.AccountID)))
	if err != nil {
		return nil, nil, err
	}

	collections, err := account.CreateBucketIfNotExists([]byte("$collections"))
	if err != nil {
		return nil, nil, err
	}

	found, err := findCollection(collections, collection)
	if err != nil {
		return nil, nil, err
	}

	if !found {
		if collection.Name == "" {
			return nil, nil, errors.New("store collection doesn't have a name")
		}

		sequence := meta.Bucket([]byte(collectionSequence))
		if sequence == nil {
			return nil, nil, errors.New("bucket for store collection sequence doesn't exist")
		}

		key, err := sequence.NextSequence()
		if err != nil {
			return nil, nil, err
		}
		collection.ID = int64(key)

		value, err := json.Marshal(collection)
		if err != nil {
			return nil, nil, err
		}

		err = collections.Put(itob(key), value)
		if err != nil {
			return nil, nil, err
		}

		err = s.notify("collections", collection.AccountID, collection.UserID, 0, 0, collection.ID, apsql.Insert)
		if err != nil {
			return nil, nil, err
		}
	}

	bucket, err := account.CreateBucketIfNotExists(itob(uint64(collection.ID)))
	if err != nil {
		return nil, nil, err
	}

	return bucket, sequence, nil
}

func (s *BoltDBStore) ListObject(object *Object, objects *[]*Object) error {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	collect := &Collection{ID: object.CollectionID, AccountID: object.AccountID}
	bucket, _, err := s.getBucket(tx, collect)
	if err != nil {
		return err
	}

	objs, err := s._Select(tx, bucket, collect, "true")
	if err != nil {
		return err
	}
	*objects = objs

	return nil
}

func (s *BoltDBStore) CreateObject(object *Object) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket, sequence, err := s.getBucket(tx, &Collection{ID: object.CollectionID, AccountID: object.AccountID})
	if err != nil {
		return err
	}

	key, err := sequence.NextSequence()
	if err != nil {
		return err
	}

	err = bucket.Put(itob(key), object.Data)
	if err != nil {
		return err
	}
	object.ID = int64(key)

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("objects", object.AccountID, object.UserID, 0, 0, object.ID, apsql.Insert)
}

func (s *BoltDBStore) ShowObject(object *Object) error {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket, _, err := s.getBucket(tx, &Collection{ID: object.CollectionID, AccountID: object.AccountID})
	if err != nil {
		return err
	}

	object.Data = bucket.Get(itob(uint64(object.ID)))
	if object.Data == nil {
		return errors.New("id doesn't exist")
	}

	return nil
}

func (s *BoltDBStore) UpdateObject(object *Object) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket, _, err := s.getBucket(tx, &Collection{ID: object.CollectionID, AccountID: object.AccountID})
	if err != nil {
		return err
	}

	err = bucket.Put(itob(uint64(object.ID)), object.Data)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("objects", object.AccountID, object.UserID, 0, 0, object.ID, apsql.Update)
}

func (s *BoltDBStore) DeleteObject(object *Object) error {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket, _, err := s.getBucket(tx, &Collection{ID: object.CollectionID, AccountID: object.AccountID})
	if err != nil {
		return err
	}

	err = bucket.Delete(itob(uint64(object.ID)))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return s.notify("objects", object.AccountID, object.UserID, 0, 0, object.ID, apsql.Delete)
}

func (s *BoltDBStore) Insert(accountID int64, collection string, object interface{}) ([]interface{}, error) {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, sequence, err := s.createBucket(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	add := func(object interface{}) error {
		delete(object.(map[string]interface{}), "$id")
		value, err := json.Marshal(object)
		if err != nil {
			return err
		}

		key, err := sequence.NextSequence()
		if err != nil {
			return err
		}

		err = bucket.Put(itob(key), value)
		if err != nil {
			return err
		}
		object.(map[string]interface{})["$id"] = key

		return nil
	}

	var results []interface{}
	if objects, valid := object.([]interface{}); valid {
		for _, object := range objects {
			err := add(object)
			if err != nil {
				return nil, err
			}
		}
		results = objects
	} else {
		err := add(object)
		if err != nil {
			return nil, err
		}
		results = []interface{}{object}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if objects, valid := object.([]interface{}); valid {
		for _, object := range objects {
			id := object.(map[string]interface{})["$id"].(uint64)
			s.notify("objects", accountID, 0, 0, 0, int64(id), apsql.Insert)
		}
		results = objects
	} else {
		id := object.(map[string]interface{})["$id"].(uint64)
		s.notify("objects", accountID, 0, 0, 0, int64(id), apsql.Insert)
	}

	return results, nil
}

func (s *BoltDBStore) SelectByID(accountID int64, collection string, id uint64) (interface{}, error) {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, _, err := s.getBucket(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	value := bucket.Get(itob(id))
	if value == nil {
		return nil, errors.New("id doesn't exist")
	}

	var _json interface{}
	err = json.Unmarshal(value, &_json)
	if err != nil {
		return nil, err
	}

	_json.(map[string]interface{})["$id"] = id

	return _json, nil
}

func (s *BoltDBStore) UpdateByID(accountID int64, collection string, id uint64, object interface{}) (interface{}, error) {
	delete(object.(map[string]interface{}), "$id")
	value, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, _, err := s.createBucket(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	err = bucket.Put(itob(id), value)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	object.(map[string]interface{})["$id"] = id

	err = s.notify("objects", accountID, 0, 0, 0, int64(id), apsql.Update)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s *BoltDBStore) DeleteByID(accountID int64, collection string, id uint64) (interface{}, error) {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, _, err := s.createBucket(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	value := bucket.Get(itob(id))
	if value == nil {
		return nil, errors.New("id doesn't exist")
	}

	err = bucket.Delete(itob(id))
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var _json interface{}
	err = json.Unmarshal(value, &_json)
	if err != nil {
		return nil, err
	}

	_json.(map[string]interface{})["$id"] = id

	err = s.notify("objects", accountID, 0, 0, 0, int64(id), apsql.Delete)
	if err != nil {
		return nil, err
	}

	return _json, nil
}

func (s *BoltDBStore) DeleteTx(tx *bolt.Tx, notify bool, accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error) {
	collect := &Collection{AccountID: accountID, Name: collection}
	bucket, _, err := s.createBucket(tx, collect)
	if err != nil {
		return nil, err
	}

	objects, err := s._Select(tx, bucket, collect, query, params...)
	if err != nil {
		return nil, err
	}
	var results []interface{}
	for _, object := range objects {
		err = bucket.Delete(itob(uint64(object.ID)))
		if err != nil {
			return nil, err
		}
		var result map[string]interface{}
		err = object.Data.Unmarshal(&result)
		if err != nil {
			return nil, err
		}
		result["$id"] = uint64(object.ID)
		results = append(results, result)
	}

	if notify {
		for _, object := range objects {
			err = s.notify("objects", object.AccountID, 0, 0, 0, object.ID, apsql.Delete)
			if err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}

func (s *BoltDBStore) Delete(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error) {
	tx, err := s.boltdb.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results, err := s.DeleteTx(tx, true, accountID, collection, query, params...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *BoltDBStore) _Select(tx *bolt.Tx, bucket *bolt.Bucket, collection *Collection, query string, params ...interface{}) ([]*Object, error) {
	jql := &JQL{Buffer: query}
	jql.Init()
	err := jql.Parse()
	if err != nil {
		return nil, err
	}

	ast, buffer := jql.AST(), []rune(jql.Buffer)
	constraints := getConstraints(ast, &Context{buffer, nil, params})
	var results []*Object
	if len(constraints.order.path) > 0 {
		cursor := bucket.Cursor()
		key, value := cursor.First()
		for key != nil {
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.UseNumber()
			var _json interface{}
			err = decoder.Decode(&_json)
			if err != nil {
				return nil, err
			}
			if process(ast, &Context{buffer, _json, params}).b {
				_value := make([]byte, len(value))
				copy(_value, value)
				object := &Object{
					ID:           int64(btoi(key)),
					AccountID:    collection.AccountID,
					CollectionID: collection.ID,
					Data:         _value,
				}
				results = append(results, object)
			}
			key, value = cursor.Next()
		}
		var sorted sort.Interface
		sorted = &Results{results, constraints.order.path, constraints.order.numeric}
		if constraints.order.dir == "desc" {
			sorted = sort.Reverse(sorted)
		}
		sort.Sort(sorted)
		if constraints.hasOffset && constraints.offset < len(results) {
			results = results[constraints.offset:]
		}
		if constraints.hasLimit && constraints.limit <= len(results) {
			results = results[:constraints.limit]
		}
	} else {
		cursor := bucket.Cursor()
		key, value := cursor.First()
		if constraints.hasOffset {
			offset := 0
			for key != nil && offset < constraints.offset {
				key, value = cursor.Next()
			}
		}
		for key != nil {
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.UseNumber()
			var _json interface{}
			err = decoder.Decode(&_json)
			if err != nil {
				return nil, err
			}
			if process(ast, &Context{buffer, _json, params}).b {
				_value := make([]byte, len(value))
				copy(_value, value)
				object := &Object{
					ID:           int64(btoi(key)),
					AccountID:    collection.AccountID,
					CollectionID: collection.ID,
					Data:         _value,
				}
				results = append(results, object)
				if constraints.hasLimit && len(results) == constraints.limit {
					break
				}
			}
			key, value = cursor.Next()
		}
	}

	return results, nil
}

func (s *BoltDBStore) Select(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error) {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	collect := &Collection{AccountID: accountID, Name: collection}
	bucket, _, err := s.getBucket(tx, collect)
	if err != nil {
		return nil, err
	}

	objects, err := s._Select(tx, bucket, collect, query, params...)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, object := range objects {
		var result map[string]interface{}
		err = object.Data.Unmarshal(&result)
		if err != nil {
			return nil, err
		}
		result["$id"] = uint64(object.ID)
		results = append(results, result)
	}

	return results, nil
}

func (s *BoltDBStore) Shutdown() {
	if s.boltdb != nil {
		s.boltdb.Close()
	}
}

func itob(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

type Value struct {
	b      bool
	errors []error
}

type Constraints struct {
	errors []error
	order  struct {
		path    []string
		dir     string
		numeric bool
	}
	hasLimit  bool
	limit     int
	hasOffset bool
	offset    int
}

type Context struct {
	buffer []rune
	json   interface{}
	param  []interface{}
}

type Results struct {
	results []*Object
	path    []string
	numeric bool
}

var _ sort.Interface = (*Results)(nil)

func (r *Results) Len() int {
	return len(r.results)
}

func (r *Results) walkPath(i int) (string, bool) {
	var result interface{}
	err := r.results[i].Data.Unmarshal(&result)
	if err != nil {
		return "", false
	}
	valid := false
	for _, path := range r.path {
		result, valid = result.(map[string]interface{})[path]
		if !valid {
			return "", false
		}
	}
	return fmt.Sprintf("%v", result), true
}

func (r *Results) Less(i, j int) bool {
	ii, ivalid := r.walkPath(i)
	jj, jvalid := r.walkPath(j)
	if !ivalid || !jvalid {
		return false
	}
	if r.numeric {
		iii, jjj := &big.Rat{}, &big.Rat{}
		_, ivalid = iii.SetString(ii)
		_, jvalid = jjj.SetString(jj)
		if !ivalid || !jvalid {
			return ii < jj
		}
		return iii.Cmp(jjj) < 0
	}
	return ii < jj
}

func (r *Results) Swap(i, j int) {
	r.results[i], r.results[j] = r.results[j], r.results[i]
}

func getConstraints(node *node32, context *Context) (c Constraints) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return getConstraints(node.up, context)
		case ruleorder:
			c.processOrder(node.up, context)
		case rulelimit:
			c.processLimit(node.up, context)
		case ruleoffset:
			c.processOffset(node.up, context)
		}
		node = node.next
	}
	return
}

func (c *Constraints) processOrder(node *node32, context *Context) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			c.processPath(node.up, context)
		case rulecast:
			c.processCast(node.up, context)
		case ruleasc:
			c.order.dir = string(context.buffer[node.begin:node.end])
		case ruledesc:
			c.order.dir = string(context.buffer[node.begin:node.end])
		}
		node = node.next
	}
}

func (c *Constraints) processCast(node *node32, context *Context) {
	c.order.numeric = true
	for node != nil {
		if node.pegRule == rulepath {
			c.processPath(node.up, context)
		}
		node = node.next
	}
}

func (c *Constraints) processPath(node *node32, context *Context) {
	for node != nil {
		if node.pegRule == ruleword {
			c.order.path = append(c.order.path, string(context.buffer[node.begin:node.end]))
		}
		node = node.next
	}
}

func (c *Constraints) processLimit(node *node32, context *Context) {
	c.hasLimit = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.limit, err = processValue1(node.up, context)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func (c *Constraints) processOffset(node *node32, context *Context) {
	c.hasOffset = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.offset, err = processValue1(node.up, context)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func processValue1(node *node32, context *Context) (int, error) {
	for node != nil {
		switch node.pegRule {
		case ruleplaceholder:
			placeholder, err := strconv.Atoi(string(context.buffer[node.begin+1 : node.end]))
			if err != nil {
				return -1, err
			}
			if placeholder > len(context.param) {
				return -1, errors.New("placholder too large")
			}
			if holder, valid := context.param[placeholder-1].(int); valid {
				return holder, nil
			} else {
				return -1, errors.New("value must be type int")
			}
		case rulewhole:
			whole, err := strconv.Atoi(string(context.buffer[node.begin:node.end]))
			if err != nil {
				return -1, err
			}
			return whole, nil
		}
		node = node.next
	}
	return -1, errors.New("no value")
}

func process(node *node32, context *Context) (v Value) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return process(node.up, context)
		case rulee1:
			v = processRulee1(node.up, context)
		}
		node = node.next
	}
	return
}

func processRulee1(node *node32, context *Context) (v Value) {
	for node != nil {
		if node.pegRule == rulee2 {
			if !v.b {
				x := processRulee2(node.up, context)
				v.b = v.b || x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func processRulee2(node *node32, context *Context) (v Value) {
	v.b = true
	for node != nil {
		if node.pegRule == rulee3 {
			if v.b {
				x := processRulee3(node.up, context)
				v.b = v.b && x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func processRulee3(node *node32, context *Context) (v Value) {
	if node.pegRule == ruleexpression {
		return processExpression(node.up, context)
	}
	return process(node.next.up, context)
}

func compareString(op, a, b string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	}
	return false
}

func compareRat(op string, a, b *big.Rat) bool {
	switch op {
	case "=":
		return a.Cmp(b) == 0
	case "!=":
		return a.Cmp(b) != 0
	case ">":
		return a.Cmp(b) > 0
	case "<":
		return a.Cmp(b) < 0
	case ">=":
		return a.Cmp(b) >= 0
	case "<=":
		return a.Cmp(b) <= 0
	}
	return false
}

func compareBool(op string, a, b bool) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	}
	return false
}

func compareNull(op string, a, b interface{}) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	}
	return false
}

func processExpression(node *node32, context *Context) (v Value) {
	if node.pegRule == ruleboolean {
		v.b = string(context.buffer[node.begin:node.end]) == "true"
		return
	}

	path, _json, valid := node.up, context.json, false
	for path != nil {
		if path.pegRule == ruleword {
			_json, valid = _json.(map[string]interface{})[string(context.buffer[path.begin:path.end])]
			if !valid {
				return
			}
		}
		path = path.next
	}
	node = node.next
	op := strings.TrimSpace(string(context.buffer[node.begin:node.end]))
	node = node.next.up
	_a := fmt.Sprintf("%v", _json)
	switch node.pegRule {
	case ruleplaceholder:
		placeholder, err := strconv.Atoi(string(context.buffer[node.begin+1 : node.end]))
		if err != nil {
			v.errors = append(v.errors, err)
			return
		}

		if placeholder > len(context.param) {
			v.errors = append(v.errors, errors.New("placholder to large"))
			return
		}
		switch _b := context.param[placeholder-1].(type) {
		case string:
			v.b = compareString(op, _a, _b)
		case float64:
			a, b := &big.Rat{}, &big.Rat{}
			a.SetString(_a)
			b.SetFloat64(_b)
			v.b = compareRat(op, a, b)
		case int:
			a, b := &big.Rat{}, &big.Rat{}
			a.SetString(_a)
			b.SetInt64(int64(_b))
			v.b = compareRat(op, a, b)
		case bool:
			v.b = compareBool(op, _a == "true", _b)
		default:
			v.b = compareNull(op, _json, _b)
		}
	case rulestring:
		b := string(context.buffer[node.begin+1 : node.end-1])
		v.b = compareString(op, _a, b)
	case rulenumber:
		_b := string(context.buffer[node.begin:node.end])
		b := &big.Rat{}
		b.SetString(_b)
		a := &big.Rat{}
		a.SetString(_a)
		v.b = compareRat(op, a, b)
	case ruleboolean:
		b := string(context.buffer[node.begin:node.end])
		v.b = compareBool(op, _a == "true", b == "true")
	case rulenull:
		v.b = compareNull(op, _json, nil)
	}
	return
}
