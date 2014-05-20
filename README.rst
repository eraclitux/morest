MoREST
======
Sometimes adding mongodb driver code to our application can be overkill. Maybe we are dealing with an embedded device, maybe we are coding an app (html5, native), maybe platform we are using doesn't have yet a mongodb driver, or we are making a simple script, or maybe we are just lazy. In all these cases (and even in others) you can try MoREST, the simplistic, universal mongodb driver.

It sits in front your mongodb server (or replica set!) and exposes, via a `RESTful-like`_ interface, a **subset** of mongodb commands. 

MoREST *mimics* mongodb syntax so you dont have to learn some other rules. 

Plans are not to implemet all mongodb functions but just a subset ot the most useful (es no administrative task will be exposed with this driver).

RESTful-like
------------
MoREST doesn't really apply the RESTful paradigm, it just follows some of its patterns. 

Supported actions
=================
- ``find``, 
- ``insert``, 
- ``remove``, 
- ``sort``, 
- ``limit``, 
- ``count`` (on collections)

Examples of usage
=================

curl
----
Get numbers of all documents in a collection::

        $ curl -X GET 'localhost:9002/my-db.my-coll.count()'

Find documents with a given pattern::

        $ curl -g -X GET 'localhost:9002/my-db.my-coll.find({"name":"pippo"}).limit(5)'

Delete a single document::

        $ curl -g -X DELETE 'localhost:9002/my-db.my-coll.remove({"name":"pippo"})'

Find documents, sort them and limit results::

        $ curl -g -X GET 'localhost:9002/my-db.my-coll.find({"number":42}).sort({"name":-1}).limit(5)'

Insert a sigle document::

        $ curl -g -X POST 'localhost:9002/my-db.my-coll.insert({"name":"pippo"})'
Note
~~~~
- Do not use any whitespace in query passed with url.
- You have to escape double quotes.

Important notices
=================
Many RFCs were hurt developing this (poor) code.

This code is not even alfa quality. It is a work in progress and should not be used in production environments.
