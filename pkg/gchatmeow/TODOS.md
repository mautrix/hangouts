* Look into deduplicating pblite with gmessages
* Remove pointless methods that are already in the stdlib
  * Delete data/methods package
* Delete debug package
* data/endpoints, data/query and cookies can probably just be in the top-level library package
* types package also seems too small, and has things it shouldn't have like http methods
* Remove pointless wrapping of events
  * No reason to wrap once in library and again in connector, just do it once in connector
  * event package can probably be deleted entirely
