var AP = AP || {};

/**
 * HTTP holds helper classes for HTTP related tasks, such as making
 * requests with Gateway, and parsing their responses.
 *
 * @namespace
 */
AP.HTTP = AP.HTTP || {};

/**
 * Creates a new HTTP request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the method, query, headers, and body.
 */
AP.HTTP.Request = function() {
  /**
   * The HTTP method to use.
   * @type {string}
   */
  this.method = "GET";

  /**
   * The body of the request.
   * @type {string}
   */
  this.body = null;

  /**
   * The request's query parameters.
   * @type {Object.<string,string>}
   */
  this.query = {};

  /**
   * The request's headers.
   * @type {Object.<string,string|string[]>}
   */
  this.headers = {};

  if (arguments.length == 1) {
    var request = arguments[0];
    this.method = _.clone(request.method);
    this.query = _.clone(request.query);
    this.headers = _.clone(request.headers);
    this.body = _.clone(request.body);
  }
}

/**
 * Sets the object as the request's body, formatted as an HTTP encoded form.
 *
 * This method automatically sets the 'Content-Type' header
 * to 'application/json' to match.
 *
 * @param {object} object The object to serialize and use as the request body
 */
AP.HTTP.Request.prototype.setFormBody = function(object) {
  this.method = "POST";
  this.headers["Content-Type"] = "application/x-www-form-urlencoded";
  this.body = AP.HTTP.Request.encodeForm(object);
}

AP.HTTP.Request.encodeForm = function(object) {
  var encodeValue = function(encoded, key, value) {
    if (_.isArray(value)) {
      return encodeArray(encoded, key, value);
    }
    if (_.isObject(value)) {
      return encodeObject(encoded, key, value);
    }
    return key + "=" + encodeURI(value);
  };
  var encodeArray = function(encoded, key, arr) {
    _.map(arr, function(value) {
      if (encoded != "") {
        encoded += "&";
      }
      encoded += encodeValue(encoded, key + "[]", value);
    });
    return encoded;
  };
  var encodeObject = function(encoded, key, obj) {
    _.map(obj, function(value, objKey) {
      if (encoded != "") {
        encoded += "&";
      }
      encoded += encodeValue(encoded, key + "[" + objKey + "]", value);
    });
    return encoded;
  };
  var encoded = "";
  _.map(object, function(v, k) {
    if (encoded != "") {
      encoded += "&";
    }
    encoded += encodeValue("", k, v);
  });
  return encoded;
}

/**
 * SQLServer holds helper classes for SQLServer related tasks
 *
 * @namespace
 */
AP.SQLServer = AP.SQLServer || {};


/**
 * Creates a new SQLServer request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the statement and parameters
 */
AP.SQLServer.Request = function() {

  /**
   * The request's SQL statement to be executed.  Must be a query that does
   * not modify data
   * @type {string}
   */
  this.queryStatement = null;

  /**
   * The request's SQL statement to be executed.  Must be an update that modifies
   * data.
   * @type {string}
   */
  this.executeStatement = null;

  /**
   * The result types expected by the query.  The keys represent the column
   * names of the result set, and the values represent a conversion object
   * such as Int or Float.
   * @type {object}
   */
   this.resultTypes = null;

  /**
   * The request's parameters to the SQL statement.
   * @type {Array.<object>}
   */
  this.parameters = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.query = _.clone(request.queryStatement);
    this.execute = _.clone(request.executeStatement);
    this.parameters = _.clone(request.parameters);
    this.resultTypes = _.clone(request.resultTypes);
  }
}

AP.SQLServer.Request.prototype.execute = function(stmt, params) {
  this.executeStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }
}

AP.SQLServer.Request.prototype.query = function(stmt, params, resultTypes) {
  this.queryStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }

  var newResultTypes = {};
  _.each(_.pairs(this.resultTypes), function(pair) {
    var k = pair[0];
    var v = pair[1];
    newResultTypes[k] = v(null);
  });
  this.resultTypes = newResultTypes;
}

/**
 * Postgres holds helper classes for Postgres related tasks
 *
 * @namespace
 */
AP.Postgres = AP.Postgres || {};


/**
 * Creates a new Postgres request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the statement and parameters
 */
AP.Postgres.Request = function() {

  /**
   * The request's SQL statement to be executed.  Must be a query that does
   * not modify data
   * @type {string}
   */
  this.queryStatement = null;

  /**
   * The request's SQL statement to be executed.  Must be an update that modifies
   * data.
   * @type {string}
   */
  this.executeStatement = null;

  /**
   * The result types expected by the query.  The keys represent the column
   * names of the result set, and the values represent a conversion object
   * such as Int or Float.
   * @type {object}
   */
   this.resultTypes = null;

  /**
   * The request's parameters to the SQL statement.
   * @type {Array.<object>}
   */
  this.parameters = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.query = _.clone(request.queryStatement);
    this.execute = _.clone(request.executeStatement);
    this.parameters = _.clone(request.parameters);
    this.resultTypes = _.clone(request.resultTypes);
  }
}

AP.Postgres.Request.prototype.execute = function(stmt, params) {
  this.executeStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }
}

AP.Postgres.Request.prototype.query = function(stmt, params, resultTypes) {
  this.queryStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }

  var newResultTypes = {};
  _.each(_.pairs(this.resultTypes), function(pair) {
    var k = pair[0];
    var v = pair[1];
    newResultTypes[k] = v(null);
  });
  this.resultTypes = newResultTypes;
}

/**
 * MySQL holds helper classes for MySQL related tasks
 *
 * @namespace
 */
AP.MySQL = AP.MySQL || {};


/**
 * Creates a new MySQL request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the statement and parameters
 */
AP.MySQL.Request = function() {

  /**
   * The request's SQL statement to be executed.  Must be a query that does
   * not modify data
   * @type {string}
   */
  this.queryStatement = null;

  /**
   * The request's SQL statement to be executed.  Must be an update that modifies
   * data.
   * @type {string}
   */
  this.executeStatement = null;

  /**
   * The result types expected by the query.  The keys represent the column
   * names of the result set, and the values represent a conversion object
   * such as Int or Float.
   * @type {object}
   */
   this.resultTypes = null;

  /**
   * The request's parameters to the SQL statement.
   * @type {Array.<object>}
   */
  this.parameters = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.query = _.clone(request.queryStatement);
    this.execute = _.clone(request.executeStatement);
    this.parameters = _.clone(request.parameters);
    this.resultTypes = _.clone(request.resultTypes);
  }
}

AP.MySQL.Request.prototype.execute = function(stmt, params) {
  this.executeStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }
}

AP.MySQL.Request.prototype.query = function(stmt, params, resultTypes) {
  this.queryStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }

  var newResultTypes = {};
  _.each(_.pairs(this.resultTypes), function(pair) {
    var k = pair[0];
    var v = pair[1];
    newResultTypes[k] = v(null);
  });
  this.resultTypes = newResultTypes;
}

/**
 * SQL holds helper classes for SQL related tasks
 *
 * @namespace
 */
AP.SQL = AP.SQL || {};

AP.SQL.Converter = function(value, convertTo) {
  this._type = "Converter";
  this.value = value;
  this.convertTo = convertTo;
}

AP.SQL.Float = function(a) {
  return new AP.SQL.Converter(a, 'float64');
}

AP.SQL.Int = function(a) {
  return new AP.SQL.Converter(a, 'int64');
};

AP.SQL.Bool = function(a) {
  return new AP.SQL.Converter(a, 'bool');
}

AP.SQL.String = function(a) {
  return new AP.SQL.Converter(a, 'string');
}

/**
 * Mongo holds helper classes for Mongo related tasks
 *
 * @namespace
 */
AP.Mongo = AP.Mongo || {};

AP.Mongo.convertFunctions = function(args) {
  for (var i = 0; i < args.length; i++) {
    if (_.isFunction(args[i])) {
      args[i] = new String(args[i]);
    }
  }
}

AP.Mongo.ObjectId = function(_id) {
  if (_id.length != 24) {
    throw "ObjectId must be 12 bytes long";
  }
  this._id = _id;
  this.type = "id";
}

function ObjectId(_id) {
  return new AP.Mongo.ObjectId(_id);
}

AP.Mongo.unnormalizeObjectId = function(hash) {
  for (var i in hash) {
    var item = hash[i];
    if (item !== null && typeof item === 'object') {
      if (item._id !== null && item.type === 'id' && _.size(item) == 2) {
        hash[i] = ObjectId(item._id);
      } else {
        AP.Mongo.unnormalizeObjectId(item);
      }
    }
  }
}

/**
 * Creates a new Mongo request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the parameters
 */
AP.Mongo.Request = function() {
  this.arguments = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.arguments = _.clone(request.arguments);
    AP.Mongo.convertFunctions(this.arguments);
  }
}

AP.Mongo.Request.prototype.query = function() {
  this.arguments = arguments;
  AP.Mongo.convertFunctions(this.arguments);
}

/**
 * Performs a query operation on a collection.
 *
 * @param {string} collection The collection to perform the query on.
 * @param {Object} [query] Defaults to {}. See: http://docs.mongodb.org/master/reference/operator/query/
 * @param {Number} [limit] Limit the number of results. Set to 0 for no limit.
 * @param {Number} [skip] Skips some number of results.
 */
AP.Mongo.Request.prototype.find = function(collection, query, limit, skip) {
  this.query(collection, "find", query, limit, skip);
}

/**
 * Inserts a document into a collection.
 *
 * @param {string} collection The collection to insert the document into.
 * @param {Object} document A document or an array of documents to insert.
 */
AP.Mongo.Request.prototype.insert = function(collection, document) {
  this.query(collection, "insert", document);
}

/**
 * Updates the document(s) selected by query in collection.
 *
 * @param {string} collection The collection to perform the update on.
 * @param {Object} query See: http://docs.mongodb.org/master/reference/operator/query/
 * @param {Object} update A document or an update operator. See: http://docs.mongodb.org/master/reference/operator/update/
 * @param {Object} [options] Options for the update.
 * @param {Boolean} [options.upsert] Insert a document if the query doesn't match any documents.
 * @param {Boolean} [options.multi] Updates multiple documents matched by the query.
 */
AP.Mongo.Request.prototype.update = function(collection, query, update, options) {
  this.query(collection, "update", query, update, options);
}

/**
 * Performs an upsert if the document has an id, or performs an insert if the
 * document doesn't have an id.
 *
 * @param {string} collection The collection to save the document in.
 * @param {Object} document The document to upsert or insert.
 */
AP.Mongo.Request.prototype.save = function(collection, document) {
  this.query(collection, "save", document);
}

/**
 * Removes the document(s) selected by query from the collection.
 *
 * @param {string} collection The collection to remove the document(s) from.
 * @param {Object} [query] Defaults to {}. See: http://docs.mongodb.org/master/reference/operator/query/
 * @param {Boolean} [justOne] Remove just one document from the collection.
 */
AP.Mongo.Request.prototype.remove = function(collection, query, justOne) {
  this.query(collection, "remove", query, justOne);
}

/**
 * Delete the entire collection.
 *
 * @param {string} collection The collection to drop.
 */
AP.Mongo.Request.prototype.drop = function(collection) {
  this.query(collection, "drop");
}

/**
 * Process documents with an aggregation pipeline.
 *
 * @param {string} collection The collection to perform the aggregation on.
 * @param {Object} stages An array of aggregation pipeline stages. See: http://docs.mongodb.org/master/reference/operator/aggregation-pipeline/
 */
AP.Mongo.Request.prototype.aggregate = function(collection, stages) {
  var _stages = [];
  if (stages instanceof Array) {
    _stages = stages
  } else {
    for (var i = 1; i < arguments.length; i++) {
      _stages.push(arguments[i]);
    }
  }
  this.query(collection, "aggregate", _stages);
}

/**
 * Count the documents matched by query in collection.
 *
 * @param {string} collection The collection to perform the count on.
 * @param {Object} [query] Defaults to {}. See: http://docs.mongodb.org/master/reference/operator/query/
 */
AP.Mongo.Request.prototype.count = function(collection, query) {
  this.query(collection, "count", query);
}

/**
 * Perform a map reduce operation on collection.
 *
 * @param {string} collection The collection to map reduce.
 * @param {Object} query Selects documents from collection to map reduce.
 * @param {Object} scope Parametrizes the map, reduce, and finalize stages.
 * @param {Function} map The map stage.
 * @param {Function} reduce The reduce stage.
 * @param {Function} finalize The finalize stage.
 * @param {Object} [out] See: http://docs.mongodb.org/master/reference/method/db.collection.mapReduce/#mapreduce-out-mtd
 */
AP.Mongo.Request.prototype.mapReduce = function(collection, query, scope, map,
  reduce, finalize, out) {
  this.query(collection, "mapReduce", query, scope, map, reduce, finalize, out);
}

AP.SOAP = AP.SOAP || {};

/**
 * Creates a new SOAP request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the parameters
 */
AP.SOAP.Request = function() {
  /**
   * The parameters to pass into the SOAP operation.
   * @type {object}
   */
  this.params = {};
  /**
   * The SOAP service's name, as specified in the WSDL.
   * @type {string}
   */
  this.serviceName = null;
  /**
   * The endpoint name, as specified in the WSDL.
   * @type {string}
   */
  this.endpointName = null;
  /**
   * The operation name to invoke, as specified in the WSDL.
   * @type {string}
   */
  this.operationName = null;
  /**
   * The action name to invoke, as specified in the WSDL.  Will be ignored if
   * operationName is present.
   * @type {string}
   */
  this.actionName = null;
  /**
   * The URL at which to invoke the service
   * @type {string}
   */
  this.url = null;
  /**
   * The WSSE credentials to use.  Valid value is a hash including a 'username'
   * and 'password' field.
   * @type {object}
   */
  this.wssePasswordCredentials = null;

  if (arguments.length == 1) {
    var request = arguments[0];
    this.params = _.clone(request.params);
    this.serviceName = _.clone(request.serviceName);
    this.endpointName = _.clone(request.endpointName);
    this.operationName = _.clone(request.operationName);
    this.actionName = _.clone(request.actionName);
    this.url = _.clone(request.url);
    this.wssePasswordCredentials = _.clone(request.wssePasswordCredentials);
  }

}

/**
 * Script holds helper classes for Script related tasks.
 *
 * @namespace
 */
AP.Script = AP.Script || {};

/**
 * Creates a new Script request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the environment variables.
 */
AP.Script.Request = function() {
  /**
   * The request's environment variables.
   * @type {Object.<string,string>}
   */
  this.env = {};

  if (arguments.length == 1) {
    var request = arguments[0];
    this.env = _.clone(request.env);
  }
}

/**
 * Store holds helper classes for Store related tasks
 *
 * @namespace
 */
AP.Store = AP.Store || {};

/**
 * Creates a new Store request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the parameters
 */
AP.Store.Request = function() {
  this.arguments = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.arguments = _.clone(request.arguments);
  }
}

AP.Store.Request.prototype.query = function() {
  this.arguments = arguments;
}

/**
 * Inserts an object into a collection.
 *
 * @param {string} collection The collection to insert the object into.
 * @param {Object} object An object to insert.
 */
AP.Store.Request.prototype.insert = function(collection, object) {
  this.query("insert", collection, object);
}

/**
 * Selects object(s) from a collection.
 *
 * @param {string} collection The collection to select object(s) from.
 * @param {string|Number} query A query or id for selecting object(s).
 */
AP.Store.Request.prototype.select = function(collection, query) {
  var args = Array.prototype.slice.call(arguments);
  args.unshift("select");
  this.query.apply(this, args);
}

/**
 * Updates an object in a collection.
 *
 * @param {string} collection The collection to update the object in.
 * @param {Number} id The id of the object to update.
 * @param {Object} object An object to update with.
 */
AP.Store.Request.prototype.update = function(collection, id, object) {
  this.query("update", collection, id, object);
}

/**
 * Deletes an object from a collection.
 *
 * @param {string} collection The collection to delete the object from.
 * @param {string|Number} query A query or id for the object(s) to delete.
 */
AP.Store.Request.prototype.delete = function(collection, query) {
  var args = Array.prototype.slice.call(arguments);
  args.unshift("delete");
  this.query.apply(this, args);
}

/**
* LDAP holds helper classes for LDAP related tasks.
*
* @namespace
*/
AP.LDAP = AP.LDAP || {};

/**
* Creates a new LDAP request
*
* @class
* @constructor
* @param [request] - An incoming request to copy the parameters
*/
AP.LDAP.Request = function() {

  /**
   * The host where the target LDAP service is hosted
   * @type {string}
   */
  this.host = null;

  /**
   * The port on which the LDAP service is listening.  Defaults to 389
   * @type {Number}
   */
  this.port = 389;

  /**
   * The username to use for authentication to the LDAP service
   * @type {string}
   */
  this.username = null;

  /**
   * The password to use for authentication to the LDAP service
   * @type {string}
   */
  this.password = null;

  /**
   * The operation name of the LDAP function to invoke
   * @type {string}
   */
   this.operationName = null;

  /**
   * The arguments that will be passed to the LDAP function call
   * @type {Object}
   */
  this.arguments = {};

  /**
   * The additional options that are applied to the LDAP operation
   * @type {Object}
   */
  this.options = null;


  if (arguments.length == 1) {
    var request = arguments[0];
    this.host = _.clone(request.host);
    this.port = _.clone(request.port);
    this.username = _.clone(request.username);
    this.password = _.clone(request.password);
    this.operationName = _.clone(request.operationName);
    this.arguments = _.clone(request.arguments);
    this.options = _.clone(request.options);
  }
}


/**
 * An enumeration containing possible values for search scope
 */
AP.LDAP.Scope = {
  base:    "base",
  single:  "single",
  subtree: "subtree"
}

/**
 * An enumeration containing possible values for dereferencing aliases
 * during a search
 */
AP.LDAP.DereferenceAliases = {
  never:  "never",
  search: "search",
  find:   "find",
  always: "always"
}

/**
 * Execute an LDAP request.
 */
AP.LDAP.Request.prototype._execute = function(arguments, operationName, opts) {
  this.arguments = arguments;
  this.operationName = operationName;
  this.options = opts;
}

/**
 * Execute a search request.
 */
AP.LDAP.Request.prototype.search = function(baseDistinguishedName, scope,
  dereferenceAliases, sizeLimit, timeLimit,
  typesOnly, filter, attributes, controls,
  opts
) {
  var searchParams = {
    "baseDistinguishedName": baseDistinguishedName,
    "scope": scope,
    "dereferenceAliases": dereferenceAliases,
    "sizeLimit": sizeLimit,
    "timeLimit": timeLimit,
    "typesOnly": typesOnly,
    "filter": filter,
    "attributes": attributes,
    "controls": controls
  };
  this._execute(searchParams, "search", opts);
}

/**
 * Execute a bind request.
 */
AP.LDAP.Request.prototype.bind = function(username, password) {
  var bindParams = {
    "username": username,
    "password": password
  };
  this._execute(bindParams, "bind");
}

/**
 * Execute an add request
 */
AP.LDAP.Request.prototype.add = function(distinguishedName, attributes) {
  var addParams = {
    "distinguishedName": distinguishedName,
    "attributes": attributes
  };

  this._execute(addParams, "add");
}

/**
 * Execute a delete request
 */
AP.LDAP.Request.prototype.delete = function(distinguishedName) {
  var deleteParams = {
    "distinguishedName": distinguishedName
  };

  this._execute(deleteParams, "delete");
}

/**
 * Execute a modify request
 */
AP.LDAP.Request.prototype.modify = function(dinstinguishedName, addAttributes, deleteAttributes, replaceAttributes) {
  var modifyParams = {
    "distinguishedName": dinstinguishedName,
    "addAttributes": addAttributes,
    "deleteAttributes": deleteAttributes,
    "replaceAttributes": replaceAttributes
  };

  this._execute(modifyParams, "modify");
}

/**
 * Execute a compare request
 */
AP.LDAP.Request.prototype.compare = function(distinguishedName, attribute, value) {
  var compareParams = {
    "distinguishedName": distinguishedName,
    "attribute": attribute,
    "value": value
  }

  this._execute(compareParams, "compare");
}

/**
 * Hana holds helper classes for SAP Hana related tasks
 *
 * @namespace
 */
AP.Hana = AP.Hana || {};


/**
 * Creates a new SAP Hana request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the statement and parameters
 */
AP.Hana.Request = function() {

  /**
   * The request's SQL statement to be executed.  Must be a query that does
   * not modify data
   * @type {string}
   */
  this.queryStatement = null;

  /**
   * The request's SQL statement to be executed.  Must be an update that modifies
   * data.
   * @type {string}
   */
  this.executeStatement = null;

  /**
   * The result types expected by the query.  The keys represent the column
   * names of the result set, and the values represent a conversion object
   * such as Int or Float.
   * @type {object}
   */
   this.resultTypes = null;

  /**
   * The request's parameters to the SQL statement.
   * @type {Array.<object>}
   */
  this.parameters = [];

  if (arguments.length == 1) {
    var request = arguments[0];
    this.query = _.clone(request.queryStatement);
    this.execute = _.clone(request.executeStatement);
    this.parameters = _.clone(request.parameters);
    this.resultTypes = _.clone(request.resultTypes);
  }
}

AP.Hana.Request.prototype.execute = function(stmt, params) {
  this.executeStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }
}

AP.Hana.Request.prototype.query = function(stmt, params, resultTypes) {
  this.queryStatement = stmt;
  if (typeof params === 'undefined' || params === null) {
    this.parameters = [];
  } else {
    this.parameters = params;
  }

  if (typeof resultTypes === 'undefined' || resultTypes === null) {
    this.resultTypes = {};
  } else {
    this.resultTypes = resultTypes;
  }

  var newResultTypes = {};
  _.each(_.pairs(this.resultTypes), function(pair) {
    var k = pair[0];
    var v = pair[1];
    newResultTypes[k] = v(null);
  });
  this.resultTypes = newResultTypes;
}
/**
 * Push holds helper classes for Push related tasks.
 *
 * @namespace
 */
AP.Push = AP.Push || {};

/**
 * Creates a new Push request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the channel and payload.
 */
AP.Push.Request = function() {
  /**
   * The channel to send the payload to.
   * @type {string}
   */
  this.channel = null;

  /**
   * The payload to send to devices.
   * @type {Object}
   */
  this.payload = {};

  if (arguments.length == 1) {
    var request = arguments[0];
    this.channel = _.clone(request.channel);
    this.payload = _.clone(request.payload);
  }
}

/**
 * Set the channel and payload
 *
 * @param {string} channel The channel to push to
 * @param {object} payload The payload to send to the channel
 */
AP.Push.Request.prototype.push = function(channel, payload) {
  this.channel = channel;
  this.payload = payload;
1}

/**
 * Redis holds helper classes for Redis related tasks
 *
 * @namespace
 */
AP.Redis = AP.Redis || {};

/**
 * Creates a new Redis request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the statement and parameters
 */
AP.Redis.Request = function() {
  this.executeStatement = null;
  
  if (arguments.length == 1) {
    var request = arguments[0];
    this.executeStatement = _.clone(request.executeStatement)
  }
}

AP.Redis.Request.prototype.execute = function(stmt) {
  this.executeStatement = stmt;
}

/**
 * Smtp holds helper classes for SMTP related tasks
 *
 * @namespace
 */
 
AP.Smtp = AP.Smtp || {};

/**
 * Creates a new SMTP request.
 *
 * @class
 * @constructor
 * @param [request] - An incoming request to copy the parameters
 */
 AP.Smtp.Request = function() {
   this.address = null;
   this.body = null;
   
   if (arguments.length == 1) {
     var request = arguments[0];
     this.address = _.clone(request.address);
     this.body = _.clone(request.body);
   }
 }
 
 AP.Smtp.Request.prototype.send = function(address, body) {
   this.address = address;
   this.body = body;
 }
