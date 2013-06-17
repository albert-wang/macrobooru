macrobooru Network API Specification 
==================================================
This document describes the macrobooru Network API, which is used by the client ORM to communicate with the macrobooru server, implemented
in node.js. It is provided for third-party developers to write additional client/server functionality, and but is not necessary
for any other purpose.


A Note on FIPS Compliance
----------------------------------------------------
While the API specified is purely HTTP, any deployed servers will access the API over an encrypted HTTPS channel. This allows us
to make the implementation of the API simple, while still conforming to the cryptographic requirements of FIPs for data-in-motion.

We do not currently have plans to integrate data-at-rest encryption, since this is already provided by the Android kernel
developed for the Transformative Apps program. This may be amended at a later date.

API Specification
----------------------------------------------------
The API should abandon all hope of having a semblance to a RESTful API; a distributed object graph does not map well to an
endpoint per object. Instead, we should have a single endpoint which receives three types of requests: 
	* Fetching a subgraph of the graph
	* Performing a mutation on the global graph
	* Acquiring an authentication token
	* Downloading, uploading files and checking to see if a file exists.

The endpoint can receive input data in two forms -- using a normal POST request with an application/json Content-Type, or using a
multipart/mixed POST request.  

### Normal POST Request ###
~~~
POST /v2/api HTTP/1.1 
Host: <hostname> 
Content-Type: application/json 
Content-Length: <length>

{ "operation": "<query|modify|authenticate|download|exists>" 
, "data": <operation-specific payload> 
, "token" : <optional authentication token> 
}
~~~

In the later case, the primary payload should be in the part marked with the `data` Content-ID, e.g.  Multipart POST Request

~~~
POST /v2/api HTTP/1.1 
Host: <hostname> 
Content-Type: multipart/mixed; boundary=asd 
Content-Length: <length>

--asd Content-ID: data Content-Type: application/json Content-Length: <length>

{ "operation" : "<query|modify|authenticate|download> "
, "data" : <operation-specific payload> 
, "token" : <optional authentication token> 
}

--asd Content-ID: <unique id referenced in payload> Content-Type: application/octet-stream Content-Length: <length>

<operation-specific data> 

--asd--
~~~

In either case, the server will always respond with HTTP 200, returning a JSON blob which wraps the response payload in an object
which contains the status of the request in a machine readable format (statusCode) and a human-readable format (statusMsg). If
the request is successful (statusCode == 0) the `data` field will contain an operation-dependent object with the payload:

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 , "statusMsg" : "ok", "data" : <operation-dependent data> }
~~~

Or, for example:

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 1 , "statusMsg" : "You do not have permission to modify 'How to Field Clean an M9'" , "data" : null }
~~~

### Asynchronous Requests ###

The Network API provides for a single client to make multiple asynchronous requests to the server, provided that they operate on
distinct sets of objects. Due to the nature of HTTP, each request will need a separate connection to the server.  In practice,
the client built into the ORM uses a single connection for sending object modifications to the server, and a connection pool to
handle queries. While it is possible to use a pool for modifications, the additional logic to ensure that all in-flight requests
operate on mutually distinct subgraphs is non-trivial and may introduce errors. For queries which simply read data from the
server (and thus operate on the empty set), using a connection pool may increase throughput in high-latency environments.
Authentication Operation Payloads

The client authenticates by sending an `authenticate` operation with a payload containing `username` and `password` fields:

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "authenticate" , "data" : { "username" : <username> , "password" : <password> } }
~~~

A successful response will contain a `token` field, e.g.

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 , "statusMsg" : "ok", "data" : { "token" : <authentication token> } }
~~~

The returned token field should be stuffed into the root object of every subsequent request.  Modify Operation Payloads

### Request Payload Format ###

The payload consists of a list of objects to insert/update. Whether an object should be inserted/updated on the server is
inferred from the server's datastore -- if the ID already exists, the respective object will be updated; if not it will be
created.

Object relations are always unidirectional with respect to ownership, e.g., though the object graph is cyclic, each object has a
single, distinct owner with the exception of the universal group object, which is the root object. Object relations are either
one-to-one or one-to-many; the former is represented as a nullable object reference in the returned JSON in the form of a GUID
reference to the object, the later is a perhaps-empty array of object references in the form of GUIDs. Therefore, to delete an
object, you simply null the reference to it in the owning object, or remove it from the owning array.

Thus, the payload consists solely of a JSON-encoded array of objects to insert/modify. Any references to objects should be by-ID
rather than the full object (e.g., do not send nested objects as the previous API wanted). For brevity, any unchanged fields may
be omitted. The string "#primary" may be used to refer to the primary key field of a model, instead of the actual field name.

Each modify request is performed within a separate transaction -- if there is an error while modifying a model, the request will
fail as a whole and no changes will be applied to the datastore.  

### Response Format ###

As with any request, the server will wrap the response in the response template (which contains a statusCode, statusString and an
optional response payload). Each `modify` operation is done as a single transaction -- it will either succeed as a whole, or fail
as a whole. While there may be multiple errors in a single request, the server will abort upon encountering the first and will
return only the first encountered error.  The response payload will be a JSON-encoded array of warnings that were encountered
while updating the data. The warnings might include things like potentially overwriting conflicting revisions, etc.

#### Example 1: Creating a new Module Object ####

To create a Module object, we simply post the correct request type with a payload containing an array with one object: the
JSON-encoded Module we'd like to create:

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "modify"
, "data" : [ 
	{ "#model" : "Module"
	, "guid" : <the GUID of the module> 
	, "createdby" : <User's GUID> 
	, "name" : <module name>
	} ] 
}
~~~

Note that the `#model` field must be included with all updates to annotate the model type -- while the type may be inferred from
the GUID, this is not always the case and the type should be explicitly annotated.

In this case, we expect the server to return with an all-clear.

~~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 , "statusMsg" : "ok" , "data" : null }
~~~~

#### Example 2: Adding Steps to our new Module object ####

Next, we want to add steps (and potentially other data) to our Module. This is done in a similar fashion -- we simply send a
modification operation. In the object array, we send the new Step and the part of the Module object that weve changed (the steps
array):

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "modify"
, "data" : [ 
	{ "#model" : "Step"
	, "guid" : 1234 
	, "name" : <name of step> 
	, "text" : <description of step>
	},
	{ "#model" : "Module"
	, "guid" : <GUID of module in Example 1> 
	, "steps" : [1234] 
	} ]
}
~~~

Note that the Step's GUID was put into Module.steps, not the Step object itself. As far as the network API is concerned, object
should never be sent nested (nesting large objects creates additional context for a JSON parser to maintain, which will result in
larger memory overhead).  

#### Example 3: Deleting a Step object

Deleting an object from the server simply means removing it from it's owning object. Steps are owned by the Module they're
contained in, so we can delete a Step by removing it from the Module.steps array. Following from the example above, we can simply
send

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "modify"
, "data" : [
	{ "#model" : "Module"
	, "guid" : <GUID of module in Examples 1 & 2> 
	, "steps" : []
	} ] 
}
~~~

Note that we don’t have to send anything relevant to the Step, or any fields of the Module we haven’t modified. We simply remove
the owning reference to the Step and it becomes orphaned from the ownership tree of the object graph and is purged.  

### Uploading Files ###

Any changeset that modifies a `file` type field (e.g., a Static object) must use the multipart form of the API request. Each file
field should contain the string identifier which references the corresponding Content-ID of the part of the multipart request
containing the uploaded file.

Due to the nature of uploaded files to be large, it is recommended that each file is uploaded as a separate request. In the case
there is an error in the upload (or the connection is reset due to connectivity issues), uploading them piecemeal will reduce the
throughput loss due to error.

Before uploading a file, the client should compute the SHA-1 hash of the file, and send an `exists` query to check if the file
already exists. An `exists` query must contain an authorization token, and the `data` member must be a string with the SHA-1 
digest encoded as hex in ASCII characters with no spaces. Hex letters can be either lowercase or uppercase.

#### Example: Checking if a file exists on the server ###
~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "exists"
, "data" : "2fd4e1c67a2d28fced849ee1bb76e7391b93eb12"
}
~~~

The response will be contained in the data field as a boolean. The server may return true if the file exists
on the server, in which case the file should not be uploaded. 

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 , "statusMsg" : "ok", "data" : true }
~~~

#### Example 4: Add a Step with an Static object containing an image and thumbnail ####

This is pretty straightforward; given we have a module, we simply create the appropriate objects and link them onto that existing
module (which owns steps, which owns static). The Static.image is of `file` type, so their values are set
to the Content-ID of the multipart part containing the data of the upload. The `Static.thumb` property should be a JSON object
that has x and y coordinates. These coordinates are used as a suggestion for the thumbnailer to use as a point of focus for
thumbnailing. If this property is not set, then the center of the image is used instead.

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: multipart/mixed; boundary=asd 
Content-Length: <length>

--asd Content-ID: data Content-Type: application/json Content-Length: <length>

{ "operation" : "modify"
, "data" : [	
	{ "#model" : "Module"
	, "guid" : <Module GUID> 
	, "steps" : [<Step GUID>] 
	},
	{ "#model" : "Step"
	, "guid" : <Step GUID> 
	, "name" : <Step name> 
	, "image" : <Static GUID> 
	},
	{ "#model" : "Static"
	, "guid" : <Static GUID> 
	, "image" : "content_id_image"
	, "thumb" : 
		{
			x: 0,
			y: 0
		}
	} ] 
}

--asd Content-ID: content_id_image Content-Type: image/png Content-Length: <length>

<image binary> 

--asd-- 
~~~

### Download Operation Payloads ###

To download a file attachment, send a payload with a download operation.  Request Payload Format

The request for a download takes two fields --  the GUID of the specific model instance, and a 
mime-type requesting what format you'd like the data in. Downloads should have a `resolution` field
which describes the maximum screen resolution of the device in question. This field should be a string, 
with the width of the resolution, followed by the character 'x', and the height of the resolution. 
This field is used to detemrine the size of the download. Downloads should also have a `connection` 
field which describes the quality of the devices internet connection. Valid values are '2G', '3G', '4G', 'Wi-Fi', 'LAN'.

#### Example 1: Download an uploaded image

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "download"
, "data" : 
	{ "guid" : <guid to download> 
	, "mime" : "image/png" 
	, "resolution" : "1920x1080"
	, "connection" : "LAN"
	} 
} 
~~~

#### Response Payload ####

Unlike every other API call, the `download` payload does not return a JSON response. Instead, it returns a normal HTTP response,
which may not be 200. The content served will be the binary content in the requested format.  

## Query Operation Payloads ##

### General Query Format, Named Subgraphs ###

The general aim of the query language is to provide a means to operate and return select subgraphs of the overall object graph.
This design uses named subgraphs to facilitate returning multiple subgraphs from a single query, or for making drill-down
subqueries.

For example, if you wanted to fetch a single module by GUID from the server, you could make a single query and bind it to the
`module` subgraph, which would be returned in the `data` response, as follows: 

#### Example 1: Fetch metadata for a Module ####

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "module" : 
		{ "model" : "Module"
		, "where" : 
			{ "guid" : <Module GUID> } 
		} 
	} 
}
~~~

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 
, "statusMsg" : "ok"
, "data" : 
	{ "module" : 
		{ "total" : <number of modules total, if limit used> 
		, "model" : <echo back the passed model field>
		, "slice" :	[
			{ "guid" : <Module GUID> 
			, "name" : <Module Name> 
			, "steps" : [<Step 1 GUID>, <Step 2 GUID>, ...]  ...
			} ] 
		} 
	}
}
~~~

As the example shows, the GUID parameter was bound into the `where` clause object. The where clause consists of sets of
(key, value) pairs to match vertex attributes against. The key portion must refer to a non-relational field. If the field
refers to a primary key field, then value may be an array of GUIDs to match against. If the field is not a primary key field, 
it may be a query string in the format "operator test", where operator is one of `>, >=, <, <=, ==, =, !=, NULL, NOTNULL`, and 
test is the value to test against. If the field is a string, the test portion of the query string must be enclosed in single quotes.
If the string does not start with an operator, then it is equivilent to the query string `= '<input string>'`, where the single
quotes are ommited if the input string already had single quotes, or if the field is not of string type. The operator `==` is 
an alias for the operator `=`. The value may also be an array of query strings, as described above. If the value, or an 
element in the value does not conform to the query string format, it will be ignored. If the value is not
one of the above options, then the query is malformed. Only vertices that satisfy all constraints are returned 
(e.g., constraints are AND'd together).

As a convenience feature, a client can use the string '#primary' to refer to a model's primary key.

#### Example: Where with predicates ####

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "module" : 
		{ "model" : "Module"
		, "where" : 
			{ "guid" : <Module GUID> 
			, "averageRating" : ">= 2.5"
			, "title" : "Using an iPad"
			} 
		} 
	} 
}
~~~
The above query will return all modules with an average rating greater than or equal to 2.5, and is titled 'Using an iPad'.

The response payload contains a key-value array mapping each query to a query response object. The query response object contains
three fields -- “total”, "model" and “slice”. “total” is the integer number of objects that matched the constraints, in case not 
all of them were returned (either due to a “limit” clause, or because the server didn’t want to). The “slice” field contains an 
array of all matching vertexes, up to the user provided limit or 200. The "model" field contains the name of the model 
returned.

The returned vertexes contain all the attributes, but any relations are strictly GUIDs only.  Explicit Model Annotation

It is important to note that the `model` clause is required. We cannot generate a non-homogeneous subgraph: a subgraph must
contain exactly one type of object. Unfortunately, we cannot infer the object type, so it must be explicitly annotated. The
`model` clause specifies what model there `where` clause should search over -- it does not necessary indicate what kind of object
will be returned by the query, as shown later.  Fetching Relations instead of GUIDs

Instead of returning a raw list of GUIDs for a relational field, we can instead choose to return all the relations in all objects
in a subgraph, using a `relation` clause. It is important to note that, if a `relation` clause is present, the type of objects
referenced by that relation are returned instead of the type indicated in the `model` clause.

We can modify Example 1 slightly to include a `relation` clause, which will cause the query to return all steps from all modules
matched by the `where` clause: Example 2: Fetch all Step data for a Module

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "steps" : 
		{ "model" : "Module"
		, "where" : { "guid" : <Module GUID> } 
		, "relation" : "steps"
		} 
	}
}
~~~

~~~
HTTP/1.1 200 OK 
Content-Type: application/json 
Content-Length: <length>

{ "statusCode" : 0 
, "statusMsg" : "ok"
, "data" : 
	{ "steps" : 
		{ "total" : <number of steps total, if limit used> 
		, "model" : <echo back the passed model field>
		, "slice" :	[
			{ "guid" : <Step 1 GUID> , "name" : <Step 1 Name> , "image" : <Static GUID> } ...
			] 
		} 
	} 
}
~~~

An important note is that the objects returned are a simple, flat array. It is possible to write a `where` clause which matches
multiple objects; in this case you'll get a list containing the concatenated list of all their relations. There may be no way to
separate them, unless you already know which Module owns which Step GUIDs.

We can mitigate this slightly by grabbing both the desired Module and all Steps with a single query.

### Referencing a named subgraph from a query ###

Effectively, we want to combine the return results of Example 1 and 2 -- we want both the Module and the list of Steps. Because
each query must be homogeneous with respect to object type, we need two separate queries.

While we could just jam the two previous queries together, we can optimize further by having one query run in the context of
another with the `subgraph` clause. The `subgraph` clause effectively restricts the search space to the result of a previously
named query, such that further `where` clauses can be used to filter it, or a `relation` clause can be used to extract child
objects.

For example, if we want to grab a specific module, then in addition, grab that module's steps, we can send the following request:

#### Example 3: Fetch Module and all Steps ####

~~~
POST /api/v2 HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "module" : 
		{ "model" : "Module"
		, "where" : {"guid" : <Module GUID>} 
		} 
	, "steps" : 
		{ "subgraph" : "module"
		, "relation" : "steps" } 
	} 
}
~~~

An interesting note is that the `subgraph` clause is mutually exclusive with the `model` clause -- because the specified subgraph
is already typed, the object type specification is superfluous. All references t

A query containing both a `subgraph` and `model` clause will return an error.

The other important note is that a subgraph can not cyclically refer to itself.

A query which cyclically refers to itself via a chain of `subgraph` clauses will return an error.

### Limiting returned results ###

Receiving results on a mobile client is expensive in terms of both bandwidth and JSON processing time. For this reason, there are
two constructs for limiting the amount of returned data: the `transient` clause, and the `order/limit/offset` clauses.

The `transient` clause, when set to true, indicates that the named query is simply used for intermediate processing and should
not be returned to the client.

The `order/limit/offset` clauses are fairly mundane. Some edge cases: If `order` is omitted, the objects will be returned in a
stable but undefined order (e.g., they'll be sorted by an implicit index like GUID) If `limit` is omitted, the server may return
whatever length of values it feels comfortable returning. If `limit` is too high, the server may issue an error.  If `offset` is
omitted, it is assumed to be 0.

The "order" clause consists of two parts, the field and the ordering type, in the format `field ordering_type`, where 
`ordering_type` may be one of `ASC` or `DESC`, for ascending and descending, respectively. If the `ordering_type` parameter
is omitted, then it defaults to `ASC`. The 'order' clause may also be an array of such strings.

#### Example 4: Fetch subset of Modules in a Group, don't get Group ####

~~~
POST /api/v2 HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "group" : 
		{ "model" : "Group"
		, "where" : {"guid" : <Group GUID>} 
		, "transient" : true 
		}
	, "acl_modules" : 
		{ "subgraph" : "group"
		, "relation" : "module_acls"
		, "transient" : true 
		} 
	, "modules" : 
		{ "subgraph" : "acl_modules"
		, "relation" : "module" 
		, "order" : "name"
		, "limit" : 50 
		, "offset" : 25 
		} 
	} 
}
~~~

### Invoking the fulltext indexer ###

The previous examples use only the `where` clause to filter search results; while this is useful, the `where` clause only does
exact-value matching. For searching, this is non-optimal.

For objects that have pre-configured fulltext indexes, you may use the `search` clause. The `search` clause functions exactly
like the `where` clause, except that the key refers to a pre-configured index on the object, and the `value` is the phrase to
feed to the search engine.

#### Example 5: Finding Modules by name ####

~~~
POST /v2/api HTTP/1.1 
Host: <host> 
Content-Type: application/json 
Content-Length: <length>

{ "operation" : "query"
, "data" : 
	{ "modules" : 
		{ "model" : "Module"
		, "search" : {"name" : <query>} 
		} 
	} 
}
~~~

