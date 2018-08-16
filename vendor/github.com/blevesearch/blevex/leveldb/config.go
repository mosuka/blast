package leveldb

import "github.com/jmhodges/levigo"

func applyConfig(o *levigo.Options, config map[string]interface{}) (
	*levigo.Options, error) {

	cim, ok := config["create_if_missing"].(bool)
	if ok {
		o.SetCreateIfMissing(cim)
	}

	eie, ok := config["error_if_exists"].(bool)
	if ok {
		o.SetErrorIfExists(eie)
	}

	wbs, ok := config["write_buffer_size"].(float64)
	if ok {
		o.SetWriteBufferSize(int(wbs))
	}

	bs, ok := config["block_size"].(float64)
	if ok {
		o.SetBlockSize(int(bs))
	}

	bri, ok := config["block_restart_interval"].(float64)
	if ok {
		o.SetBlockRestartInterval(int(bri))
	}

	lcc, ok := config["lru_cache_capacity"].(float64)
	if ok {
		lruCache := levigo.NewLRUCache(int(lcc))
		o.SetCache(lruCache)
	}

	bfbpk, ok := config["bloom_filter_bits_per_key"].(float64)
	if ok {
		bf := levigo.NewBloomFilter(int(bfbpk))
		o.SetFilterPolicy(bf)
	}

	mof, ok := config["max_open_files"].(float64)
	if ok {
		o.SetMaxOpenFiles(int(mof))
	}

	return o, nil
}
