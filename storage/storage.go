// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package storage provides the functions that interacts with Spanner to fetch the data
package storage

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/firestore"
	"github.com/cloudspannerecosystem/dynamodb-adapter/config"
	"github.com/cloudspannerecosystem/dynamodb-adapter/models"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/logger"
	"github.com/cloudspannerecosystem/dynamodb-adapter/utils"
)

// Storage object for intracting with storage package
type Storage struct {
	firestoreClient map[string] *firestore.Client
	spannerClient 	map[string]	*spanner.Client
}

// storage - global instance of storage
var storage *Storage

func initFirestoreDriver() *firestore.Client {
	client, err := firestore.NewClient(context.Background(), config.ConfigurationMap.GoogleProjectID)
	if err != nil {
		logger.LogFatal(err)
	}
	return client
}

func initSpannerDriver(instance string) *spanner.Client {
	conf := spanner.ClientConfig{}

	str := "projects/" + config.ConfigurationMap.GoogleProjectID + "/instances/" + instance + "/databases/" + config.ConfigurationMap.SpannerDb
	client, err := spanner.NewClientWithConfig(context.Background(), str, conf)
	if err != nil {
		logger.LogFatal(err)
	}
	return client
}

// InitializeDriver - this will Initialize databases object in global map
func InitializeDriver() {
	storage = new(Storage)
	storage.firestoreClient = make(map[string]*firestore.Client)
	storage.spannerClient = make(map[string]*spanner.Client)

	for _, table := range models.SpannerTableMap {
		if _, ok := storage.spannerClient[table]; !ok {
			storage.spannerClient[table] = initSpannerDriver(table)
		}
	}

	for _, table := range models.FirestoreTableMap {
		if _, ok := storage.firestoreClient[table]; !ok {
			storage.firestoreClient[table] = initFirestoreDriver()
		}
	}
}

// Close - This gracefully returns the session pool objects, when driver gets exit signal
func (s Storage) Close() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	logger.LogDebug("Connection Shutdown start")
	for _, v := range s.spannerClient {
		v.Close()
	}
	logger.LogDebug("Connection shutted down")
}

var once sync.Once

// GetStorageInstance - return storage instance to call db functionalities
func GetStorageInstance() *Storage {
	once.Do(func() {
		if storage == nil {
			InitializeDriver()
		}
	})

	return storage
}

func (s Storage) getSpannerClient(tableName string) *spanner.Client {
	return s.spannerClient[models.SpannerTableMap[utils.ChangeTableNameForSpanner(tableName)]]
}

func (s Storage) getFirestoreClient(tableName string) *firestore.Client {
	return s.firestoreClient[models.FirestoreTableMap[utils.ChangeTableNameForSpanner(tableName)]]
}
