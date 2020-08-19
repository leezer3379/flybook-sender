package esc

import (
	"context"
	"github.com/leezer3379/flybook-sender/config"
	"gopkg.in/olivere/elastic.v5"
	"log"
	"os"
	"fmt"
)

var (
	client *elastic.Client
	cfg config.Config
)

func InitEs()  {
	cfg = config.Get()
	addr := cfg.Es.Addr
	index := cfg.Es.Index
	//c = new(elastic.Client)
	errorlog := log.New(os.Stdout, "APP ", log.LstdFlags)
	// Obtain a client. You can also provide your own HTTP client here.
	var err error
	client, err = elastic.NewClient(elastic.SetURL("http://"+addr),elastic.SetErrorLog(errorlog))
	if err != nil {
		// Handle error
		panic(err)
	}
	IsExists(index)

}

func IsExists(index string) {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}


	if !exists {
		// Create a new index.
		mapping := `
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"_default_": {
			"_all": {
				"enabled": true
			}
		},
		"n9ealert": {
			"properties": {
				"Status": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Sname": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Endpoint": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Metric": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Tags": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Value": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Info": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Etime": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Elink": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"Priority": {
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"@timestamp": {"type": "date"}
				
			
			}
		}
	}
}`
		createIndex, err := client.CreateIndex(index).Body(mapping).Do(context.Background())
		if err != nil {
			// Handle error
			fmt.Println("Connet Es Error %s", err)
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			fmt.Println("Creat index %s OK", index)
		}

	}
}

func CloseEs() {
	client.CloseIndex(cfg.Es.Index)
}