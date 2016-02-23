package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gateway/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	postgresCurrentVersion = 1
)

type PostgresStore struct {
	conf config.Store
	db   *sqlx.DB
}

func (s *PostgresStore) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStore) Migrate() error {
	var currentVersion int64
	err := s.db.Get(&currentVersion, `SELECT version FROM schema LIMIT 1`)
	migrate := s.conf.Migrate
	if err != nil {
		tx := s.db.MustBegin()
		tx.MustExec(`
      CREATE TABLE IF NOT EXISTS schema (
        version integer
      );
    `)
		tx.MustExec(`INSERT INTO schema VALUES (0);`)
		err := tx.Commit()
		if err != nil {
			return err
		}

		migrate = true
	}

	if currentVersion == postgresCurrentVersion {
		return nil
	}

	if !migrate {
		return errors.New("The store is not up to date. Please migrate by invoking with the -store-migrate flag.")
	}

	if currentVersion < 1 {
		tx := s.db.MustBegin()
		tx.MustExec(`
      CREATE TABLE IF NOT EXISTS "collections" (
        "id" SERIAL PRIMARY KEY,
        "account_id" INTEGER NOT NULL,
        "name" TEXT NOT NULL,
				UNIQUE ("account_id", "name")
      );
    `)
		tx.MustExec(`
			CREATE INDEX idx_collections_account_id ON collections USING btree(account_id);
			CREATE INDEX idx_collections_name ON collections USING btree(name);
			ANALYZE;
		`)
		tx.MustExec(`
      CREATE TABLE IF NOT EXISTS "objects" (
        "id" SERIAL PRIMARY KEY,
        "account_id" INTEGER NOT NULL,
				"collection_id" INTEGER NOT NULL,
        "data" JSON NOT NULL,
				FOREIGN KEY("collection_id") REFERENCES "collections"("id") ON DELETE CASCADE
      );
    `)
		tx.MustExec(`
      CREATE INDEX idx_objects_account_id ON objects USING btree(account_id);
			CREATE INDEX idx_objects_collection_id ON objects USING btree(collection_id);
      ANALYZE;
    `)
		tx.MustExec(`UPDATE schema SET version = 1;`)
		err := tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) Clear() error {
	tx := s.db.MustBegin()
	tx.MustExec(`DROP TABLE IF EXISTS "schema" CASCADE;`)
	tx.MustExec(`DROP TABLE IF EXISTS "collections" CASCADE;`)
	tx.MustExec(`DROP TABLE IF EXISTS "objects" CASCADE;`)
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) ListCollection(collection *Collection, collections *[]*Collection) error {
	rows, err := s.db.Queryx("SELECT id, account_id, name FROM collections WHERE account_id = $1",
		collection.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	for rows.Next() {
		_collection := &Collection{}
		err := rows.StructScan(_collection)
		if err != nil {
			return err
		}
		*collections = append(*collections, _collection)
	}

	return nil
}

func (s *PostgresStore) CreateCollection(collection *Collection) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	row := tx.QueryRowx("SELECT id, account_id, name FROM collections WHERE account_id = $1 AND name = $2;",
		collection.AccountID, collection.Name)
	if err != nil {
		return err
	}
	err = row.StructScan(collection)
	if err != nil {
		if err == sql.ErrNoRows {
			err = tx.Get(&collection.ID, `INSERT into collections (account_id, name) VALUES ($1, $2) RETURNING "id";`,
				collection.AccountID, collection.Name)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return ErrCollectionExists
	}

	return nil
}

func (s *PostgresStore) ShowCollection(collection *Collection) error {
	err := s.db.Get(collection, "SELECT id, account_id, name FROM collections WHERE account_id = $1 AND name = $2;",
		collection.AccountID, collection.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateCollection(collection *Collection) (err error) {
	rows, err := s.db.Queryx("UPDATE collections SET name = $1 WHERE id = $2 AND account_id = $3;",
		collection.Name, collection.ID, collection.AccountID)
	if err != nil {
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteCollection(collection *Collection) error {
	err := s.db.Get(collection, "DELETE FROM collections WHERE id = $1 AND account_id = $2 RETURNING *;",
		collection.ID, collection.AccountID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) getCollection(tx *sqlx.Tx, collection *Collection) error {
	var row *sqlx.Row
	if collection.ID != 0 {
		row = tx.QueryRowx("SELECT id, account_id, name FROM collections WHERE account_id = $1 AND id = $2;",
			collection.AccountID, collection.ID)
	} else {
		row = tx.QueryRowx("SELECT id, account_id, name FROM collections WHERE account_id = $1 AND name = $2;",
			collection.AccountID, collection.Name)
	}
	return row.StructScan(collection)
}

func (s *PostgresStore) addCollection(tx *sqlx.Tx, collection *Collection) (err error) {
	err = s.getCollection(tx, collection)
	if err != nil {
		if err == sql.ErrNoRows {
			err = tx.Get(&collection.ID, `INSERT into collections (account_id, name) VALUES ($1, $2) RETURNING "id";`,
				collection.AccountID, collection.Name)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) ListObject(object *Object, objects *[]*Object) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	collection := &Collection{ID: object.CollectionID, AccountID: object.AccountID}
	err = s.getCollection(tx, collection)
	if err != nil {
		return err
	}
	objs, err := s._Select(tx, collection.AccountID, collection.ID, "true")
	if err != nil {
		return err
	}
	*objects = objs

	return nil
}

func (s *PostgresStore) CreateObject(object *Object) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	collection := &Collection{ID: object.CollectionID, AccountID: object.AccountID}
	err = s.getCollection(tx, collection)
	if err != nil {
		return err
	}

	return tx.Get(&object.ID, `INSERT into objects (account_id, collection_id, data) VALUES ($1, $2, $3) RETURNING "id";`,
		object.AccountID, object.CollectionID, object.Data)
}

func (s *PostgresStore) ShowObject(object *Object) error {
	return s.db.Get(object, "SELECT id, account_id, collection_id, data FROM objects WHERE id = $1 AND account_id = $2;",
		object.ID, object.AccountID)
}

func (s *PostgresStore) UpdateObject(object *Object) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = s.getCollection(tx, &Collection{ID: object.CollectionID, AccountID: object.AccountID})
	if err != nil {
		return err
	}

	rows, err := tx.Queryx("UPDATE objects SET data = $1 WHERE id = $2 AND account_id = $3;",
		object.Data, object.ID, object.AccountID)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (s *PostgresStore) DeleteObject(object *Object) error {
	return s.db.Get(object, "DELETE FROM objects WHERE id = $1 AND account_id = $2 RETURNING *;", object.ID, object.AccountID)
}

func (s *PostgresStore) SelectByID(accountID int64, collection string, id uint64) (interface{}, error) {
	object := Object{}
	err := s.db.Get(&object, "SELECT id, account_id, collection_id, data FROM objects WHERE id = $1 AND account_id = $2;",
		id, accountID)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = object.Data.Unmarshal(&result)
	if err != nil {
		return nil, err
	}
	result.(map[string]interface{})["$id"] = id
	return result, nil
}

func (s *PostgresStore) UpdateByID(accountID int64, collection string, id uint64, object interface{}) (result interface{}, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = s.addCollection(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	delete(object.(map[string]interface{}), "$id")
	value, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Queryx("UPDATE objects SET data = $1 WHERE id = $2 AND account_id = $3;",
		string(value), id, accountID)
	if err != nil {
		return nil, err
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	object.(map[string]interface{})["$id"] = id
	return object, nil
}

func (s *PostgresStore) DeleteByID(accountID int64, collection string, id uint64) (result interface{}, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = s.addCollection(tx, &Collection{AccountID: accountID, Name: collection})
	if err != nil {
		return nil, err
	}

	object := Object{}
	err = tx.Get(&object, "DELETE FROM objects WHERE id = $1 AND account_id = $2 RETURNING *;", id, accountID)
	if err != nil {
		return nil, err
	}

	err = object.Data.Unmarshal(&result)
	if err != nil {
		return nil, err
	}
	result.(map[string]interface{})["$id"] = id

	return result, nil
}

func (s *PostgresStore) Insert(accountID int64, collection string, object interface{}) (results []interface{}, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	collect := &Collection{AccountID: accountID, Name: collection}
	err = s.addCollection(tx, collect)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Preparex(`INSERT into objects (account_id, collection_id, data) VALUES ($1, $2, $3) RETURNING "id";`)
	if err != nil {
		return nil, err
	}
	add := func(object interface{}) error {
		delete(object.(map[string]interface{}), "$id")
		value, err := json.Marshal(object)
		if err != nil {
			return err
		}
		var id int64
		err = stmt.Get(&id, accountID, collect.ID, string(value))
		if err != nil {
			return err
		}
		object.(map[string]interface{})["$id"] = uint64(id)
		return nil
	}
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
	return results, nil
}

func (s *PostgresStore) Delete(accountID int64, collection string, query string, params ...interface{}) (results []interface{}, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	collect := &Collection{AccountID: accountID, Name: collection}
	err = s.addCollection(tx, collect)
	if err != nil {
		return nil, err
	}

	objects, err := s._Select(tx, accountID, collect.ID, query, params...)
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Preparex("DELETE FROM objects WHERE id = $1 AND account_id = $2;")
	if err != nil {
		return nil, err
	}
	for _, object := range objects {
		rows, err := stmt.Queryx(object.ID, accountID)
		if err != nil {
			return nil, err
		}
		err = rows.Close()
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
	return results, nil
}

func (s *PostgresStore) _Select(tx *sqlx.Tx, accountID int64, collectionID int64, query string, params ...interface{}) ([]*Object, error) {
	jql := &JQL{Buffer: query}
	jql.Init()
	if err := jql.Parse(); err != nil {
		return nil, err
	}
	ast, buffer := jql.tokenTree.AST(), []rune(jql.Buffer)
	query, length := pgProcess(ast, &Context{buffer, nil, params}).s, len(params)
	params = append(params, accountID, collectionID)
	query = fmt.Sprintf(`SELECT id, account_id, collection_id, data FROM objects WHERE account_id = $%v AND collection_id = $%v AND %v;`,
		length+1, length+2, query)

	rows, err := tx.Queryx(query, params...)
	if err != nil {
		return nil, err
	}

	var objects []*Object
	for rows.Next() {
		object := &Object{}
		err = rows.StructScan(object)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func (s *PostgresStore) Select(accountID int64, collection string, query string, params ...interface{}) (results []interface{}, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	collect := &Collection{AccountID: accountID, Name: collection}
	err = s.getCollection(tx, collect)
	if err != nil {
		return nil, err
	}
	objects, err := s._Select(tx, accountID, collect.ID, query, params...)
	if err != nil {
		return nil, err
	}
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

func (s *PostgresStore) Shutdown() {
	if s.db != nil {
		s.db.Close()
	}
}

type Query struct {
	s      string
	errors []error
}

func pgProcess(node *node32, context *Context) (q Query) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return pgProcess(node.up, context)
		case rulee1:
			x := pgProcessRulee1(node.up, context)
			q.s += "( " + x.s + " )"
			q.errors = append(q.errors, x.errors...)
		case ruleorder:
			x := pgProcessOrder(node.up, context)
			q.s += " " + x.s
		case rulelimit:
			x := pgProcessLimit(node.up, context)
			q.s += " " + x.s
		case ruleoffset:
			x := pgProcessOffset(node.up, context)
			q.s += " " + x.s
		}
		node = node.next
	}
	return
}

func pgProcessOrder(node *node32, context *Context) (q Query) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			path := pgProcessPath(node.up, context)
			q.s = "ORDER BY " + path.s
		case rulecast:
			cast := pgProcessCast(node.up, context)
			q.s = "ORDER BY " + cast.s
		case ruleasc:
			q.s += " ASC"
		case ruledesc:
			q.s += " DESC"
		}
		node = node.next
	}
	return
}

func pgProcessCast(node *node32, context *Context) (q Query) {
	for node != nil {
		if node.pegRule == rulepath {
			path := pgProcessPath(node.up, context)
			q.s = "CAST( " + path.s + " as numeric )"
		}
		node = node.next
	}
	return
}

func pgProcessPath(node *node32, context *Context) (q Query) {
	segments := []string{}
	for node != nil {
		if node.pegRule == ruleword {
			segments = append(segments, string(context.buffer[node.begin:node.end]))
		}
		node = node.next
	}
	last := len(segments) - 1
	q.s = "data"
	for _, segment := range segments[:last] {
		q.s += "->'" + segment + "'"
	}
	q.s += "->>'" + segments[last] + "'"
	return
}

func pgProcessLimit(node *node32, context *Context) (q Query) {
	for node != nil {
		if node.pegRule == rulevalue1 {
			q.s += "LIMIT " + strings.TrimSpace(string(context.buffer[node.begin:node.end]))
		}
		node = node.next
	}
	return
}

func pgProcessOffset(node *node32, context *Context) (q Query) {
	for node != nil {
		if node.pegRule == rulevalue1 {
			q.s += "OFFSET " + strings.TrimSpace(string(context.buffer[node.begin:node.end]))
		}
		node = node.next
	}
	return
}

func pgProcessRulee1(node *node32, context *Context) (q Query) {
	or := ""
	for node != nil {
		if node.pegRule == rulee2 {
			x := pgProcessRulee2(node.up, context)
			q.s += or + x.s
			q.errors = append(q.errors, x.errors...)
			or = " OR "
		}
		node = node.next
	}
	return
}

func pgProcessRulee2(node *node32, context *Context) (q Query) {
	and := ""
	for node != nil {
		if node.pegRule == rulee3 {
			x := pgProcessRulee3(node.up, context)
			q.s += and + x.s
			q.errors = append(q.errors, x.errors...)
			and = " AND "
		}
		node = node.next
	}
	return
}

func pgProcessRulee3(node *node32, context *Context) (q Query) {
	if node.pegRule == ruleexpression {
		return pgProcessExpression(node.up, context)
	}
	x := pgProcess(node.next.up, context)
	q.s = "(" + x.s + ")"
	q.errors = x.errors
	return
}

func pgProcessExpression(node *node32, context *Context) (q Query) {
	if node.pegRule == ruleboolean {
		q.s = string(context.buffer[node.begin:node.end])
		return
	}

	path, segments := node.up, []string{}
	for path != nil {
		if path.pegRule == ruleword {
			segments = append(segments, string(context.buffer[path.begin:path.end]))
		}
		path = path.next
	}
	q.s = "data"
	last := len(segments) - 1
	for _, segment := range segments[:last] {
		q.s += "->'" + segment + "'"
	}
	q.s += "->>'" + segments[last] + "'"

	node = node.next
	op := strings.TrimSpace(string(context.buffer[node.begin:node.end]))
	node = node.next.up
	switch node.pegRule {
	case ruleplaceholder:
		placeholder, err := strconv.Atoi(string(context.buffer[node.begin+1 : node.end]))
		if err != nil {
			q.errors = append(q.errors, err)
			return
		}

		if placeholder > len(context.param) {
			q.errors = append(q.errors, errors.New("placholder to large"))
			return
		}
		switch context.param[placeholder-1].(type) {
		case string:
			q.s = fmt.Sprintf("%v %v $%v", q.s, op, placeholder)
		case float64:
			q.s = fmt.Sprintf("CAST(%v as FLOAT) %v $%v", q.s, op, placeholder)
		case int:
			q.s = fmt.Sprintf("CAST(%v as INTEGER) %v $%v", q.s, op, placeholder)
		case bool:
			q.s = fmt.Sprintf("CAST(%v as BOOLEAN) %v $%v", q.s, op, placeholder)
		default:
			switch op {
			case "=":
				q.s = fmt.Sprintf("%v IS NULL", q.s)
			case "!=":
				q.s = fmt.Sprintf("%v IS NOT NULL", q.s)
			}
		}
	case rulestring:
		param := string(context.buffer[node.begin+1 : node.end-1])
		q.s = fmt.Sprintf("%v %v '%v'", q.s, op, param)
	case rulenumber:
		param := string(context.buffer[node.begin:node.end])
		if strings.Contains(param, ".") {
			q.s = fmt.Sprintf("CAST(%v as FLOAT) %v %v", q.s, op, param)
		} else {
			q.s = fmt.Sprintf("CAST(%v as INTEGER) %v %v", q.s, op, param)
		}
	case ruleboolean:
		param := string(context.buffer[node.begin:node.end])
		q.s = fmt.Sprintf("CAST(%v as BOOLEAN) %v %v", q.s, op, param)
	case rulenull:
		switch op {
		case "=":
			q.s = fmt.Sprintf("%v IS NULL", q.s)
		case "!=":
			q.s = fmt.Sprintf("%v IS NOT NULL", q.s)
		}
	}
	return
}
