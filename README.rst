
.. contents::

MoREST
======
Sometimes adding mongodb driver code to our application can be overkill. Maybe we are dealing with an embedded device, maybe we are coding an app (html5, native), maybe platform we are using doesn't have yet a mongodb driver, or we are writing a shell script, or maybe we are just lazy. In all these cases (and even in others) you can try MoREST, the simplistic, universal mongodb driver.

Every language that has http requests capabilities (GET, POST, PUT, DELETE) can query mongodb through this driver. MoREST **mimics** mongodb syntax via a `RESTful-like`_ interface so you dont have to learn some other rules. 

Plans are to not implement all mongodb functions but just a **subset** of the most useful from the end user perspective (es. no administrative task will be exposed with this driver).

RESTful-like
------------
MoREST doesn't really apply the RESTful paradigm, it just follows some of its patterns. 

Supported actions
=================
find
----
Syntax::

        db.collection.find(<criteria>)

Projection is not supported.

insert
------
Syntax::

        db.collection.insert(<single document>)

To insert multiple documents::

        db.collection.insert()

passing an array of json data as request body. 

remove 
------
Syntax::

        db.collection.remove(<query>, {justOne: <boolean>})

Second argument is optional, default is to remove multiple documents.

update
------
Syntax::

        db.collection.update(<query>, <update>, {upsert: <boolean>, multi: <boolean>})

sort
----
Syntax::

        db.collection.find(<criteria>).sort(<number>)

limit
-----
Syntax::

        db.collection.find(<criteria>).limit(<number>)

count
-----
Actually only supported on collections::

        db.collection.count()

Note
----
- **Do not** use whitespaces in query passed as url.

Examples of usage
=================
Here some examples.

curl
----
Get numbers of all documents in a collection::

        $ curl -X GET 'localhost:9002/my-db.my-coll.count()'

Find documents with a given pattern::

        $ curl -g -X GET 'localhost:9002/my-db.my-coll.find({"name":"Zaphod"}).limit(5)'

Delete a single document::

        $ curl -g -X DELETE 'localhost:9002/my-db.my-coll.remove({"name":"Zaphod"})'

Find documents, sort them and limit results::

        $ curl -g -X GET 'localhost:9002/my-db.my-coll.find({"number":42}).sort({"name":-1}).limit(5)'

Insert a sigle document::

        $ curl -g -X POST 'localhost:9002/my-db.my-coll.insert({"name":"Zaphod"})'

Insert multiple documents::

        $ curl -X 'localhost:9002/my-db.my-coll.insert()'\
        > POST -d '{"name":"Arthur"},{"name":"Ford"},{"name":"Zaphod"}' 

Update a sigle document::

        $ curl -g -X PUT 'localhost:9002/my-db.my-coll.update({"name":"Ford"},{"name":"Arthur"})'

Update multiple documents::

	$ curl -g -X PUT 'localhost:9002/my-db.my-coll.update({"name":"Ford"},{"$set":{"num":42}},{"multi":1})`,
Note
~~~~
- **Do not** use whitespaces in url or in payloads passed with POST.
- ``$`` operators must be quoted.

.. It sits in front your mongodb server (or replica set!) and exposes, , a **subset** of mongodb commands. 
.. Being based on the amazing `mgo <http://labix.org/mgo>`_, you can configure it to act in different consistency modes in case you are using replication. From mgo's documentation:

.. - **Strong** consistency uses a unique connection with the master so that all reads and writes are as up-to-date as possible and consistent with each other.

.. Can we achieve Monotonic making Copy()/Clone() for every http request?
.. - **Monotonic** consistency will start reading from a slave if possible, so that the load is better distributed, and once the first write happens the connection is switched to the master. This offers consistent reads and writes, but may not show the most up-to-date data on reads which precede the first write.

.. - **Eventual** consistency offers the best resource usage, distributing reads across multiple slaves and writes across multiple connections to the master, but consistency isn't guaranteed.

Important notices
=================
- Some RFCs were hurt developing this (poor) code.
- This code is alfa quality, it is a work in progress and should not be used in production environments.
