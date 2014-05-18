MoREST
======
Sometimes adding mongodb driver code to our application can be overkill. Maybe we are dealing with an embedded device, maybe we are coding an app (html5, native), maybe platform we are using doesn't have yet a mongodb driver, or we are making a simple script, or maybe we are just lazy. In all these cases (and even in others) you can try MoREST, the simplistic, universal mongodb driver.

It sits in front your mongodb server (or replica set!) and exposes, via a RESTful-like interface, a **subset** of mongodb commands. Being based on the amazing `mgo <http://labix.org/mgo>`_, you can configure it to act in three consistency modes in case you are using replication:
- **Strong** consistency uses a unique connection with the master so that all reads and writes are as up-to-date as possible and consistent with each other.
- **Monotonic** consistency will start reading from a slave if possible, so that the load is better distributed, and once the first write happens the connection is switched to the master. This offers consistent reads and writes, but may not show the most up-to-date data on reads which precede the first write.
- **Eventual** consistency offers the best resource usage, distributing reads across multiple slaves and writes across multiple connections to the master, but consistency isn't guaranteed.

MoREST *mimics* mongodb syntax so you dont have to learn some other rules. 

Note on RESTful
---------------
Before you complain about the RESTful term used above, MoREST doesn't really apply the RESTful paradigm, it just follows some of its patterns. Maybe the results is not elegant and many will frown, sorry.

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

        $ curl -g -X GET "localhost:9002/my-db.my-coll.count()"

Find documents with a given pattern::

        $ curl -g -X GET "localhost:9002/my-db.my-coll.find({\"name\":\"pippo\"}).limit(5)"

Delete a single document::

        $ curl -g -X GET "localhost:9002/my-db.my-coll.remove({\"name\":\"pippo\"})"

Find documents, sort them and limit results::

        $ curl -g -X GET "localhost:9002/my-db.my-coll.find({\"number\":\42\"}).sort({\"name\":-1}).limit(5)"
Note
~~~~
- Do not use any whitespace in query passed with url.
- You have to escape double quotes.

Important notice
================
This code is not even alfa quality. It is a work in progress and should not be used in production environments.
