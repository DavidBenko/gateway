# Remote Endpoints

    {
        "remote_endpoint": {
            "name": "hello",
            "description": "world",
            "type": "http",
            "environment_data": [
                {
                    "environment_id": 1,
                    "data": {
                        "url": "http://demo.ap.com"
                    }
                }
            ]
        }
    }

* In the future there will be multiple types of remote endpoints, differentiated by the `type` key. For the moment, `type` should always be "http".
* `name` is used in JavaScript code and should be a valid as a JavaScript identifier
* `environment_data` is an array of objects that configure the endpoint per environment. The `environment_id` should only ever point to an environment in the same API as this endpoint; pointint to an environment of another API is an error. 
* `data` within `environment_data` is arbitrary JSON. The data necessary will depend on the endpoint type. 

  For the "http" type, the data may include:

    - `url`: string of URL to hit
    - `method`: HTTP method to use
    - `headers`: object indicating headers to use, i.e.
      `{ "X-Header": "Foo" }`
  

#### Change Warnings
  
The `data` format isn't fully specced out. Validation has not yet been decided or implemented and the keys in http `data` may change in the future depending on implementation needs.

`environment_data` array items are stored in a separate table. I am considering refactoring this implementation to use a single value primary key `id` like the proxy endpoints use for submodels. 

# Proxy Endpoints

## Routes

    "routes": [
      {
        "path": "/v10/something.json",
        "methods": ["POST", "PUT"]        
      }
    ]
    

* `routes` is an array of routes the proxy endpoint is accessible on. Each object should have a single `path` string, and an array of 1 or more HTTP verbs in its `method` key.
  
## Components

    "components": [
      {
        "type": "single",
        "conditional": "",
        "conditional_positive": true,
        ...
      },
      {
        "type": "multi",
        ...
      },
      {
        "type": "js",
        ...
      }
    ]
    
* `components` is an array of component objects
* `type` indicates the component type, which may be one of: `single`, `multi`, or `js`; the details of each are discussed below.
* `conditional` is an optional string of JavaScript that the proxy will evaluate to determine if the component should be executed. Semantically the statement should evaluate to a boolean.
* `conditional_positive` is a boolean value that indicates whether the component should be executed by the proxy when the statement in `conditional` evaluates to true. 
  
  *Examples:*
  
  - `{"conditional": "true;", "conditional_positive": true }`: the component will be executed
  - `{"conditional": "true;", "conditional_positive": false }`: the component will not be executed
  - `{"conditional": "false;", "conditional_positive": false }`: the component will be executed
      
* Components are executed by the proxy in order, so the position in the array is meaningful and is preserved by the database.
* Components are stored in the database in a separate table, keyed by primary key `id`. For creation, no `id` is necessary, however, for updating a component its id should be included. 
  
  *Examples:*
  
  - On endpoint creation, a component: `{ "type: "js", ... }` will result in a new component being added with
    a new endpoint.
  - On endpoint update, a component: `{ "type: "js", ... }` will result in a new component being added to
    the existing endpoint.
  - On endpoint update, a component: `{ "id": 1, "type: "js", ... }` will result in the existing component being
    updated on the existing endpoint.
  - On endpoint update, omitting the component will result in the existing component being deleted from
    the existing endpoint.

### Single Proxy Components

    {
      "type": "single",
      "conditional": "",
      "conditional_positive": true,
      "before": [
        {  
          "type": "js",
          "data": "before single"
        }
      ],
      "call": {
        "endpoint_name_override": "singleCall",
        "remote_endpoint_id": 1
      },
      "after": [
        {  
          "type": "js",
          "data": "after single"
        }
      ]
    }

Single proxy components execute exactly one remote endpoint call.

* Both `before` and `after` contain arrays of transformation objects. Presently the `type` is always `js`, and the `data` (which may be arbitrary JSON), should always be a JavaScript string. The order is meaningful and should be preserved.
* `call` is a call to a remote endpoint. Like components, calls are stored separately in the database, and the `id` key should be included in updates.
  
  - `endpoint_name_override` is a key that identifies the remote endpoint if its name is unsuitable for some reason
  - `remote_endpoint_id` should only point to a remote endpoint in the same API as this proxy endpoint
  
### Multi Proxy Components

    {
      "type": "multi",
      "conditional": "",
      "conditional_positive": true,
      "before": [...],
      "calls": [
        {
          "conditional": "",
          "conditional_positive": true,
          "before": [
            {  
              "type": "js",
              "data": "before multi call a"
            }
          ],
          "endpoint_name_override": "multiA",
          "remote_endpoint_id": 1,
          "after": [
            {  
              "type": "js",
              "data": "after multi call a"
            }
          ]
        },
        ...
      ],
      "after": [...]
    }
    
Multi proxy components execute zero or more remote endpoint calls in parallel.

* `before` and `after` are available to multi proxy components and are defined the same as for single proxy components.
* Instead of a single `call`, there is the plural `calls` key which contains an array of call objects.
* Each `call` of a multi proxy component may also include `conditional`, `conditional_positive`, `before`, and `after` keys. All of these are structured the same as the component versions and perform analogous functions.

### JavaScript Components

    {
      "type": "js",
      "conditional": "",
      "conditional_positive": true,
      "data": "code string"
    }
    
A JavaScript component executes some JavaScript.

* `data` is a string of JavaScript to execute.
* There are no remote endpoint calls.
* `before` and `after` are valid, but omitted in examples since we presently only support JavaScript transformations, which gain us nothing in this case. They may be added once we support more types of transformations.





