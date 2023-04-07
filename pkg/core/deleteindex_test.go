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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteIndex(t *testing.T) {
	var indexName = "TestDeleteIndex.index_1"
	type args struct {
		name string
	}
	cfg := config.NewGlobalConfig()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exist",
			args: args{
				name: indexName,
			},
			wantErr: false,
		},
		{
			name: "not exist",
			args: args{
				name: "my-index-not-exist",
			},
			wantErr: true,
		},
	}

	t.Run("prepare", func(t *testing.T) {
		index, err := NewIndex(indexName, "disk", 2, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, index)
		err = StoreIndex(index)
		assert.NoError(t, err)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteIndex(tt.args.name, cfg.DataPath); (err != nil) != tt.wantErr {
				t.Errorf("DeleteIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
