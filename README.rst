Quirks
======
 - sort function args...

Examples of usage
=================

curl
----
::

        curl -g -X GET "localhost:9002/my-db.my-coll.find({\"name\":\"pippo\"}).sort("name").limit(5)"
Notes:
- do not use spaces
