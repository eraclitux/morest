Intro
=====
Sometimes addind mongo driver code to our application can be overkill. Maybe we are dealing with an embedded device, maybe we are daeling with an app (html5, native), maybe platform we are using doesn have yet a mongodb driver, or we are making a simplescript, or maybe we are just lazy. In all these cases you can try MoREST, the ingenuos mongodb proxy.

It sits in front you mongodb server (or replica set) and exposes, via a RESTful interface, a **subset** of mongodb commands. Being based on the amazingo mgo, you can configure it to act in three consistency modes:
1
2
3

Examples of usage
=================

curl
----
Get numbers of all documents in a collection::

        curl -g -X GET "localhost:9002/my-db.my-coll.count()"

Find documents with a given pattern::

        curl -g -X GET "localhost:9002/my-db.my-coll.find({\"name\":\"pippo\"}).sort("name").limit(5)"

Find documents, sort them and limit results::

        curl -g -X GET "localhost:9002/my-db.my-coll.find({\"name\":\"pippo\"}).sort("name").limit(5)"
Notes:
- do not use spaces
