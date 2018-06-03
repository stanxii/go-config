package etcd

import (
	"encoding/json"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func makeEvMap(data map[string]interface{}, kv []*clientv3.Event, stripPrefix string) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	for _, v := range kv {
		switch v.Type {
		case mvccpb.DELETE:
			data = update(data, v.Kv, "delete", stripPrefix)
		default:
			data = update(data, v.Kv, "insert", stripPrefix)
		}
	}

	return data
}

func makeMap(kv []*mvccpb.KeyValue, stripPrefix string) map[string]interface{} {
	data := make(map[string]interface{})

	for _, v := range kv {
		data = update(data, v, "put", stripPrefix)
	}

	return data
}

func update(data map[string]interface{}, v *mvccpb.KeyValue, action, stripPrefix string) map[string]interface{} {
	// remove prefix if non empty, and ensure leading / is removed as well
	vkey := strings.TrimPrefix(strings.TrimPrefix(string(v.Key), stripPrefix), "/")
	// split on prefix
	keys := strings.Split(vkey, "/")

	var vals interface{}
	json.Unmarshal(v.Value, &vals)

	// set data for first iteration
	kvals := data

	// iterate the keys and make maps
	for i, k := range keys {
		kval, ok := kvals[k].(map[string]interface{})
		if !ok {
			// create next map
			kval = make(map[string]interface{})
			// set it
			kvals[k] = kval
		}

		// last key: write vals
		if l := len(keys) - 1; i == l {
			switch action {
			case "delete":
				delete(kvals, k)
			default:
				kvals[k] = vals
			}
			break
		}

		// set kvals for next iterator
		kvals = kval
	}

	return data
}