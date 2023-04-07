package search

import (
	"github.com/zinclabs/zincsearch/pkg/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/zinclabs/zincsearch/pkg/core"
	"github.com/zinclabs/zincsearch/pkg/errors"
	"github.com/zinclabs/zincsearch/pkg/meta"
	"github.com/zinclabs/zincsearch/pkg/zutils"
)

// DeleteByQuery searches the index and deletes all matches
//
// @Id DeleteByQuery
// @Summary Searches the index and deletes all matched documents
// @security BasicAuth
// @Tags    Search
// @Accept  json
// @Produce json
// @Param   index  path  string  true  "Index"
// @Param   query  body  meta.ZincQueryForSDK true  "Query"
// @Success 200 {object} meta.HTTPResponseDeleteByQuery
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/{index}/_delete_by_query [post]
func DeleteByQuery(c *gin.Context) {
	start := time.Now()
	query := &meta.ZincQuery{Size: 10}
	if err := zutils.GinBindJSON(c, query); err != nil {
		log.Printf("handlers.search.searchDSL: %s", err.Error())
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	cfg := config.GetConfig(c)

	indexName := c.Param("target")
	resp, err := searchIndex([]string{indexName}, query, cfg)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	failures := []string{}
	for _, hit := range resp.Hits.Hits {
		index, _ := core.GetIndex(hit.Index)
		err := index.DeleteDocument(hit.ID, cfg.Shard.GoroutineNum)
		if err != nil {
			failures = append(failures, hit.ID)
		}
	}

	totalDeletes := resp.Hits.Total.Value - len(failures)
	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponseDeleteByQuery{
		Took:             time.Since(start).Milliseconds(),
		TimedOut:         false,
		Total:            totalDeletes,
		Deleted:          totalDeletes,
		Batches:          0,
		VersionConflicts: 0,
		Noops:            0,
		Failures:         failures,
		Retries: meta.HttpRetriesResponse{
			Bulk:   0,
			Search: 0,
		},
		ThrottledMillis:      0,
		RequestsPerSecond:    -1,
		ThrottledUntilMillis: 0,
	})
}
