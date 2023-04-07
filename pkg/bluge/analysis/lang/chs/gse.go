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

package chs

import (
	"strings"

	"github.com/blugelabs/bluge/analysis"
	"github.com/go-ego/gse"

	"github.com/zinclabs/zincsearch/pkg/bluge/analysis/lang/chs/analyzer"
	"github.com/zinclabs/zincsearch/pkg/bluge/analysis/lang/chs/token"
	"github.com/zinclabs/zincsearch/pkg/bluge/analysis/lang/chs/tokenizer"
	"github.com/zinclabs/zincsearch/pkg/zutils"
)

func NewGseStandardAnalyzer() *analysis.Analyzer {
	return analyzer.NewStandardAnalyzer(seg)
}

func NewGseSearchAnalyzer() *analysis.Analyzer {
	return analyzer.NewSearchAnalyzer(seg)
}

func NewGseStandardTokenizer() analysis.Tokenizer {
	return tokenizer.NewStandardTokenizer(seg)
}
func NewGseSearchTokenizer() analysis.Tokenizer {
	return tokenizer.NewSearchTokenizer(seg)
}

func NewGseStopTokenFilter() analysis.TokenFilter {
	return token.NewStopTokenFilter(seg, nil)
}

var seg *gse.Segmenter

func NewSegmenter(cfgEnable bool, cfgEmbed, cfgDictPath string) {
	seg = new(gse.Segmenter)
	enable := cfgEnable // true / false
	embed := cfgEmbed   // small / big
	embed = strings.ToUpper(embed)
	loadDict(enable, embed, cfgDictPath)
}

func loadDict(enable bool, embed, dictPath string) {
	if enable {
		if embed == "BIG" {
			_ = seg.LoadDictEmbed("zh_s")
			_ = seg.LoadStopEmbed()
		} else {
			_ = seg.LoadDictStr(_dictCHS)
			_ = seg.LoadStopStr(_dictStop)
		}
	} else {
		_ = seg.LoadDictStr(`zinc`)
		_ = seg.LoadStopStr(_dictStop)
	}
	seg.Load = true
	seg.SkipLog = true

	// load user dict
	dataPath := dictPath
	userDict := dataPath + "/user.txt"
	if ok, _ := zutils.IsExist(userDict); ok {
		_ = seg.LoadDict(userDict)
	}
	stopDict := dataPath + "/stop.txt"
	if ok, _ := zutils.IsExist(stopDict); ok {
		_ = seg.LoadStop(stopDict)
	}
}
