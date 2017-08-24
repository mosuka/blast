//  Copyright (c) 2017 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/coreos/etcd/clientv3"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type PutClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	cluster            string
	indexMapping       string
	indexType          string
	kvstore            string
	kvconfig           string
}

var putClusterCmdOpts = PutClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	cluster:            "",
	indexMapping:       "",
	indexType:          "upside_down",
	kvstore:            "boltdb",
	kvconfig:           "",
}

var putClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "edits the cluster information",
	Long:  `The edit cluster command edits the cluster information.`,
	RunE:  runEPutClusterCmd,
}

func runEPutClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if putClusterCmdOpts.cluster == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if putClusterCmdOpts.indexMapping != "" {
		file, err := os.Open(putClusterCmdOpts.indexMapping)
		if err != nil {
			return err
		}
		defer file.Close()

		indexMapping, err = util.NewIndexMapping(file)
		if err != nil {
			return err
		}
	}

	// Kvconfig
	kvconfig := make(map[string]interface{})
	if putClusterCmdOpts.kvconfig != "" {
		file, err := os.Open(putClusterCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	var err error

	//var bytesIndexMapping []byte
	//if indexMapping != nil {
	//	bytesIndexMapping, err = json.Marshal(indexMapping)
	//	if err != nil {
	//		return err
	//	}
	//}

	//var bytesKvconfig []byte
	//if kvconfig != nil {
	//	bytesKvconfig, err = json.Marshal(kvconfig)
	//	if err != nil {
	//		return err
	//	}
	//}

	cfg := clientv3.Config{
		Endpoints:   putClusterCmdOpts.etcdEndpoints,
		DialTimeout: time.Duration(putClusterCmdOpts.etcdDialTimeout) * time.Millisecond,
		Context:     context.Background(),
	}

	//c, err := clientv3.New(cfg)
	c, err := client.NewCluster(cfg)
	if err != nil {
		return err
	}
	defer c.Close()

	//var kv clientv3.KV
	//if c != nil {
	//	kv = clientv3.NewKV(c)
	//}

	resp := struct {
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	}{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(putClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	if cmd.Flag("index-mapping").Changed {
		//keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/index_mapping", putClusterCmdOpts.cluster)
		//
		//_, err = kv.Put(ctx, keyIndexMapping, string(bytesIndexMapping))
		//if err != nil {
		//	return err
		//}

		err := c.PutIndexMapping(ctx, putClusterCmdOpts.cluster, indexMapping)
		if err != nil {
			return err
		}

		resp.IndexMapping = indexMapping
	}

	if cmd.Flag("index-type").Changed {
		//keyIndexType := fmt.Sprintf("/blast/clusters/%s/index_type", putClusterCmdOpts.cluster)
		//
		//_, err = kv.Put(ctx, keyIndexType, putClusterCmdOpts.indexType)
		//if err != nil {
		//	return err
		//}

		err := c.PutIndexType(ctx, putClusterCmdOpts.cluster, putClusterCmdOpts.indexType)
		if err != nil {
			return err
		}

		resp.IndexType = putClusterCmdOpts.indexType
	}

	if cmd.Flag("kvstore").Changed {
		//keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", putClusterCmdOpts.cluster)
		//
		//_, err = kv.Put(ctx, keyKvstore, putClusterCmdOpts.kvstore)
		//if err != nil {
		//	return err
		//}

		err := c.PutKvstore(ctx, putClusterCmdOpts.cluster, putClusterCmdOpts.kvstore)
		if err != nil {
			return err
		}

		resp.Kvstore = putClusterCmdOpts.kvstore
	}

	if cmd.Flag("kvconfig").Changed {
		//keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", putClusterCmdOpts.cluster)
		//
		//_, err = kv.Put(ctx, keyKvconfig, string(bytesKvconfig))
		//if err != nil {
		//	return err
		//}

		err := c.PutKvconfig(ctx, putClusterCmdOpts.cluster, kvconfig)
		if err != nil {
			return err
		}

		resp.Kvconfig = kvconfig
	}

	// output response
	switch rootCmdOpts.outputFormat {
	case "text":
		fmt.Printf("%v\n", resp)
	case "json":
		output, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)
	default:
		fmt.Printf("%v\n", resp)
	}

	return nil
}

func init() {
	putClusterCmd.Flags().SortFlags = false

	putClusterCmd.Flags().StringSliceVar(&putClusterCmdOpts.etcdEndpoints, "etcd-endpoint", putClusterCmdOpts.etcdEndpoints, "etcd eendpoint")
	putClusterCmd.Flags().IntVar(&putClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", putClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	putClusterCmd.Flags().IntVar(&putClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", putClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.cluster, "cluster", putClusterCmdOpts.cluster, "cluster name")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.indexMapping, "index-mapping", putClusterCmdOpts.indexMapping, "index mapping")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.indexType, "index-type", putClusterCmdOpts.indexType, "index type")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.kvstore, "kvstore", putClusterCmdOpts.kvstore, "kvstore")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.kvconfig, "kvconfig", putClusterCmdOpts.kvconfig, "kvconfig")

	putCmd.AddCommand(putClusterCmd)
}
