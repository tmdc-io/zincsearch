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

package ider

import (
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/zinclabs/zincsearch/pkg/zutils/base62"
)

const (
	NodeContextKey string = "zincsearch-node"
)

func GetNode(c *gin.Context) *Node {
	n := c.MustGet(NodeContextKey).(*Node)
	return n
}

func InjectNode(n *Node) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(NodeContextKey, n)
		c.Next()
	}
}

var local *Node

type Node struct {
	node *snowflake.Node
}

func LocalNode() *Node {
	if local == nil {
		local, _ = newNode(1)
	}
	return local
}

func newNode(id int) (*Node, error) {
	node, err := snowflake.NewNode(int64(id % 1024))
	return &Node{node: node}, err
}

func (n *Node) Generate() string {
	return base62.Encode(n.node.Generate().Int64())
}
