# cznicb Bleve KV store

Although the cznicb KV store is pure-Go, it is in the extensions package since it doesn't not fully satisfy the Bleve contract, which requires reader isolation.

If, for example, you always load the entire dataset, and then **ONLY** query it, you can safely use this store.