/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package metadata

import (
	"errors"
	"strings"

	"github.com/zinclabs/zincsearch/pkg/config"
	"github.com/zinclabs/zincsearch/pkg/metadata/storage"
	"github.com/zinclabs/zincsearch/pkg/metadata/storage/badger"
	"github.com/zinclabs/zincsearch/pkg/metadata/storage/bolt"
	"github.com/zinclabs/zincsearch/pkg/metadata/storage/etcd"
)

var ErrorKeyNotExists = errors.New("key not exists")

var db storage.Storager

//var cfg = config.NewEnvFileGlobalConfig([]string{"../../.env"})

func NewStorager(cfg *config.Config) {
	//func init() {
	if strings.ToLower(cfg.ServerMode) == "cluster" {
		db = etcd.New(cfg.Etcd.Prefix+"/metadata", cfg.Etcd)
	} else {
		switch strings.ToLower(cfg.MetadataStorage) {
		case "badger":
			db = badger.New("_metadata.db", cfg.DataPath)
		default:
			db = bolt.New("_metadata.bolt", cfg.DataPath)
		}
	}
}

func Close() error {
	return db.Close()
}
