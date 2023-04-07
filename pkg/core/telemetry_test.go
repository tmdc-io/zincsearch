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

package core

import (
	"github.com/zinclabs/zincsearch/pkg/config"
	"github.com/zinclabs/zincsearch/pkg/ider"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTelemetry(t *testing.T) {
	indexName := "TestTelemetry.index_1"
	cfg := config.NewGlobalConfig()
	node, _ := ider.NewNode(1)
	t.Run("prepare", func(t *testing.T) {
		index, err := NewIndex(indexName, "disk", 1, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, index)

		err = StoreIndex(index)
		assert.NoError(t, err)
	})

	tel := NewTelemetry(true, node)
	t.Run("telemetry", func(t *testing.T) {
		id := tel.createInstanceID()
		assert.NotEmpty(t, id)
		tel.Instance()
		tel.Event("server_start", nil)
		tel.Cron()

		tel.GetIndexSize(indexName)
		tel.HeartBeat()
	})

	t.Run("cleanup", func(t *testing.T) {
		err := DeleteIndex(indexName, cfg.DataPath)
		assert.NoError(t, err)
	})
}
