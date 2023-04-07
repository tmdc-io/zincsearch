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

package aggregation

import (
	"math"
	"sort"
	"time"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"
	"github.com/rs/zerolog/log"

	"github.com/zinclabs/zincsearch/pkg/zutils"
)

type AutoDateHistogramAggregation struct {
	src             search.FieldSource
	size            int
	minimumInterval string
	format          string
	timeZone        *time.Location

	aggregations map[string]search.Aggregation

	lessFunc             func(a, b *search.Bucket) bool
	desc                 bool
	sortFunc             func(p sort.Interface)
	aggregationTermsSize int
}

// NewAutoDateHistogramAggregation returns a termsAggregation
// field use to set the field use to terms aggregation
func NewAutoDateHistogramAggregation(
	field search.FieldSource,
	buckets int, minimumInterval,
	format string, timeZone *time.Location,
	aggregationTermsSize int,
) *AutoDateHistogramAggregation {
	rv := &AutoDateHistogramAggregation{
		src:             field,
		size:            buckets,
		minimumInterval: minimumInterval,
		format:          format,
		timeZone:        timeZone,
		desc:            false,
		lessFunc: func(a, b *search.Bucket) bool {
			return a.Name() < b.Name()
		},
		aggregations:         make(map[string]search.Aggregation),
		sortFunc:             sort.Sort,
		aggregationTermsSize: aggregationTermsSize,
	}
	rv.aggregations["count"] = aggregations.CountMatches()
	return rv
}

func (t *AutoDateHistogramAggregation) Fields() []string {
	rv := t.src.Fields()
	for _, agg := range t.aggregations {
		rv = append(rv, agg.Fields()...)
	}
	return rv
}

func (t *AutoDateHistogramAggregation) Calculator() search.Calculator {
	return &AutoDateHistogramCalculator{
		src:                  t.src,
		size:                 t.size,
		format:               t.format,
		timeZone:             t.timeZone,
		intervals:            t.getIntervals(),
		minValue:             math.MaxInt64,
		maxValue:             math.MinInt64,
		aggregations:         t.aggregations,
		desc:                 t.desc,
		lessFunc:             t.lessFunc,
		sortFunc:             t.sortFunc,
		bucketsMap:           make(map[int64]*search.Bucket),
		aggregationTermsSize: t.aggregationTermsSize,
	}
}

func (t *AutoDateHistogramAggregation) AddAggregation(name string, aggregation search.Aggregation) {
	t.aggregations[name] = aggregation
}

func (t *AutoDateHistogramAggregation) getIntervals() []time.Duration {
	intervals := make([]time.Duration, 0, 20)
	switch t.minimumInterval {
	case "second":
		intervals = append(intervals, time.Second*1, time.Second*5, time.Second*10, time.Second*30)
		fallthrough
	case "minute":
		intervals = append(intervals, time.Minute*1, time.Minute*5, time.Minute*10, time.Minute*30)
		fallthrough
	case "hour":
		intervals = append(intervals, time.Hour*1, time.Hour*3, time.Hour*12)
		fallthrough
	case "day":
		intervals = append(intervals, time.Hour*24*1, time.Hour*24*7)
		fallthrough
	case "month":
		intervals = append(intervals, time.Hour*24*30*1, time.Hour*24*30*3, time.Hour*24*30*6)
		fallthrough
	case "year":
		intervals = append(intervals, time.Hour*24*30*12*1, time.Hour*24*30*12*5, time.Hour*24*30*12*10, time.Hour*24*30*12*20, time.Hour*24*30*12*50)
	default:
		// noop
	}
	return intervals
}

type AutoDateHistogramCalculator struct {
	src       interface{}
	size      int
	intervals []time.Duration
	format    string
	timeZone  *time.Location

	currentInterval int
	minValue        int64
	maxValue        int64

	aggregations map[string]search.Aggregation

	bucketsList []*search.Bucket
	bucketsMap  map[int64]*search.Bucket
	total       int
	other       int

	desc                 bool
	lessFunc             func(a, b *search.Bucket) bool
	sortFunc             func(p sort.Interface)
	aggregationTermsSize int
}

func (a *AutoDateHistogramCalculator) Consume(d *search.DocumentMatch) {
	a.total++
	src := a.src.(search.DateValuesSource)
	for _, term := range src.Dates(d) {
		key := term.UnixNano()
		if key < a.minValue {
			a.minValue = key
		}
		if key > a.maxValue {
			a.maxValue = key
		}
		termKey, termStr := a.bucketKey(key)
		bucket, ok := a.bucketsMap[termKey]
		if ok {
			bucket.Consume(d)
		} else {
			newBucket := search.NewBucket(termStr, a.aggregations)
			newBucket.Consume(d)
			a.bucketsMap[termKey] = newBucket
		}
	}
}

func (a *AutoDateHistogramCalculator) Merge(other search.Calculator) {
	if other, ok := other.(*AutoDateHistogramCalculator); ok {
		// first sum to the totals and others
		a.total += other.total
		// now, walk all of the other buckets
		// if we have a local match, merge otherwise append
		for i := range other.bucketsList {
			var foundLocal bool
			for j := range a.bucketsList {
				if other.bucketsList[i].Name() == a.bucketsList[j].Name() {
					a.bucketsList[j].Merge(other.bucketsList[i])
					foundLocal = true
				}
			}
			if !foundLocal {
				// a.bucketsList = append(a.bucketsList, other.bucketsList[i])
				keyTime, err := time.ParseInLocation(a.format, other.bucketsList[i].Name(), a.timeZone)
				if err != nil {
					log.Error().Msgf("auto_date_histogram: Failed to parse time: %v", err)
					continue
				}
				key := keyTime.UnixNano()
				a.bucketsMap[key] = other.bucketsList[i]
				if a.minValue > key {
					a.minValue = key
				}
				if a.maxValue < key {
					a.maxValue = key
				}
				if a.currentInterval < other.currentInterval {
					a.currentInterval = other.currentInterval
				}
			}
		}
		// now re-invoke finish, this should trim to correct size again
		// and recalculate other
		a.afterMerge()
		a.Finish()
	}
}

func (a *AutoDateHistogramCalculator) afterMerge() {
	for key, bucket := range a.bucketsMap {
		delete(a.bucketsMap, key)
		termKey, termStr := a.bucketKey(key)
		newBucket, ok := a.bucketsMap[termKey]
		if ok {
			newBucket.Merge(bucket)
		} else {
			newBucket = search.NewBucket(termStr, a.aggregations)
			newBucket.Merge(bucket)
			a.bucketsMap[termKey] = newBucket
		}
	}
}

func (a *AutoDateHistogramCalculator) Finish() {
	// Calc AggregationTermsSize
	for a.minValue+int64(a.intervals[a.currentInterval])*int64(a.aggregationTermsSize) < a.maxValue {
		a.currentInterval++
	}

	for {
		// Replenish bucket
		if len(a.bucketsMap) <= a.size {
			for value := a.minValue; value < a.maxValue; value += int64(a.intervals[a.currentInterval]) {
				termKey, termStr := a.bucketKey(value)
				if a.bucketsMap == nil {
					a.bucketsMap = make(map[int64]*search.Bucket)
				}
				if _, ok := a.bucketsMap[termKey]; !ok {
					a.bucketsMap[termKey] = search.NewBucket(termStr, a.aggregations)
				}
			}
		}
		// check bucket size
		if !(len(a.bucketsMap) > a.size && a.currentInterval < len(a.intervals)-1) {
			break
		}
		a.currentInterval++
		for key, bucket := range a.bucketsMap {
			delete(a.bucketsMap, key)
			termKey, termStr := a.bucketKey(key)
			newBucket, ok := a.bucketsMap[termKey]
			if ok {
				newBucket.Merge(bucket)
			} else {
				newBucket = search.NewBucket(termStr, a.aggregations)
				newBucket.Merge(bucket)
				a.bucketsMap[termKey] = newBucket
			}
		}
	}

	a.bucketsList = a.bucketsList[:0]
	for _, bucket := range a.bucketsMap {
		a.bucketsList = append(a.bucketsList, bucket)
	}
	// a.bucketsMap = nil

	// sort the buckets
	if a.desc {
		a.sortFunc(sort.Reverse(a))
	} else {
		a.sortFunc(a)
	}

	var notOther int
	for _, bucket := range a.bucketsList {
		notOther += int(bucket.Aggregations()["count"].(search.MetricCalculator).Value())
	}
	a.other = a.total - notOther
}

func (a *AutoDateHistogramCalculator) Buckets() []*search.Bucket {
	return a.bucketsList
}

func (a *AutoDateHistogramCalculator) Other() int {
	return a.other
}

func (a *AutoDateHistogramCalculator) Interval() string {
	return zutils.FormatDuration(a.intervals[a.currentInterval])
}

func (a *AutoDateHistogramCalculator) Len() int {
	return len(a.bucketsList)
}

func (a *AutoDateHistogramCalculator) Less(i, j int) bool {
	return a.lessFunc(a.bucketsList[i], a.bucketsList[j])
}

func (a *AutoDateHistogramCalculator) Swap(i, j int) {
	a.bucketsList[i], a.bucketsList[j] = a.bucketsList[j], a.bucketsList[i]
}

func (a *AutoDateHistogramCalculator) bucketKey(value int64) (int64, string) {
	var nsec int64
	if a.intervals[a.currentInterval] >= time.Hour*24*30*12 {
		t := time.Unix(0, value)
		t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
		nsec = t.UnixNano()
	} else if a.intervals[a.currentInterval] >= time.Hour*24*30 {
		t := time.Unix(0, value)
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		nsec = t.UnixNano()
	} else {
		nsec = (value / int64(a.intervals[a.currentInterval])) * int64(a.intervals[a.currentInterval])
	}
	return nsec, time.Unix(0, nsec).Format(a.format)
}
