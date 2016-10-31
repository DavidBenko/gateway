package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
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
	constraints := Constraints{
		context: context{buffer},
		param:   params,
	}
	constraints.process(ast)
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
			selector := Selector{
				context: context{buffer},
				json:    _json,
				param:   params,
			}
			if selector.process(ast).b {
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
			selector := Selector{
				context: context{buffer},
				json:    _json,
				param:   params,
			}
			if selector.process(ast).b {
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

	aggregations := Aggregations{
		context:      context{buffer},
		Aggregations: make(map[string]Aggregation),
	}
	aggregations.Process(ast)
	if len(aggregations.errors) > 0 {
		return nil, aggregations.errors[0]
	}
	if len(aggregations.Aggregations) > 0 {
		for _, result := range results {
			var _json interface{}
			err := json.Unmarshal(result.Data, &_json)
			if err != nil {
				return nil, err
			}
			for _, aggregation := range aggregations.Aggregations {
				aggregation.Accumulate(_json)
			}
		}
		result := make(map[string]interface{})
		for name, aggregation := range aggregations.Aggregations {
			result[name] = aggregation.Compute()
		}
		data, err := json.Marshal(&result)
		if err != nil {
			return nil, err
		}
		results = []*Object{&Object{
			AccountID:    collection.AccountID,
			CollectionID: collection.ID,
			Data:         data,
		}}
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
