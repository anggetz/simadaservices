package pipelines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libcore/models"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/go-redis/cache/v9"
	"gopkg.in/mgo.v2/bson"
	"gorm.io/gorm"
)

type SyncInventaris struct {
	es         *elasticsearch.Client
	db         *gorm.DB
	redisCache *cache.Cache
}

func NewSyncInventaris() *SyncInventaris {
	return &SyncInventaris{}
}

func (s *SyncInventaris) SetRedisCache(redisCache *cache.Cache) *SyncInventaris {
	s.redisCache = redisCache
	return s
}

func (s *SyncInventaris) SetDB(db *gorm.DB) *SyncInventaris {
	s.db = db
	return s
}

func (s *SyncInventaris) SyncPgToElastic() {
	fmt.Println("performing elastic ", time.Now(), s.es)

	resp, err := esapi.IndicesExistsRequest{
		Index: []string{"inventaris-index"},
	}.Do(context.Background(), s.es)

	if err != nil {
		fmt.Println("error", err.Error())
		return
	}

	if resp.StatusCode == 404 {
		fmt.Println("creating new index")

		s.es.Indices.Create("inventaris-index")

	}

	inventaris := []models.Inventaris{}

	take := 30
	page := 0
	lenData := take
	isFirstQuery := true

	// 2024-01-01 updated_at
	//

	// get latest data
	queryJson := bson.M{
		"sort": bson.M{
			"updated_at": "desc",
		},
		"size": 1,
	}

	query, err := json.Marshal(queryJson)
	if err != nil {
		fmt.Println("error", err.Error())
		return
	}

	data, err := s.es.Search(
		s.es.Search.WithIndex("inventaris-index"),
		s.es.Search.WithBody(strings.NewReader(string(query))),
	)

	elasticResponse := models.Elastic{}

	if b, err := io.ReadAll(data.Body); err == nil && data.StatusCode == 200 {
		err = json.Unmarshal(b, &elasticResponse)
		if err != nil {
			fmt.Println("error", err.Error())
			return
		}
	}

	lastTimePull := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)

	if len(elasticResponse.Hits.Hits) > 0 {
		timeUpdatedAt, err := time.Parse(time.RFC3339, elasticResponse.Hits.Hits[0].Source["updated_at"].(string))

		fmt.Println(elasticResponse.Hits.Hits[0].Source["updated_at"], "elastic response")

		if err != nil {
			fmt.Println("error get updated at", err.Error())
			return
		}
		lastTimePull = timeUpdatedAt
	}

	fmt.Println(lastTimePull, "start date to pull")

	for next := true; next; next = lenData >= take {

		txGorm := s.db.
			Preload("BarangMaster").
			Preload("DetilTanahRel").
			Preload("DetilAsetLainnyaRel").
			Preload("DetilBangunanRel").
			Preload("DetilMesinRel").
			Preload("DetilMesinRel.MerkMaster").
			Preload("DetilJalanRel").
			Preload("DetilKonstruksiRel").
			Preload("AlamatKotaRel").
			Preload("AlamatKecamatanRel").
			Model(new(models.Inventaris)).
			Where("updated_at >= ? ", lastTimePull).
			Order("updated_at ASC").Offset(page * take).Limit(take).Find(&inventaris)

		if txGorm.Error != nil {
			fmt.Println("error", txGorm.Error.Error())
			return
		}

		// mutex := sync.Mutex{}

		if len(inventaris) > 0 {

			for _, inven := range inventaris {

				byInven, err := json.Marshal(inven)
				if err != nil {
					fmt.Println("error", err.Error())
					return
				}

				req := esapi.ExistsRequest{
					Index:      "inventaris-index",
					DocumentID: strconv.Itoa(inven.ID),
				}

				resExist, err := req.Do(context.Background(), s.es)
				if err != nil {
					fmt.Println("error get exist document", err.Error(), resExist.String(), strconv.Itoa(inven.ID))
					return
				}

				var res *esapi.Response

				if resExist.StatusCode == 200 {
					res, err = s.es.Delete("inventaris-index", strconv.Itoa(inven.ID))
				}

				res, err = s.es.Create("inventaris-index", strconv.Itoa(inven.ID), bytes.NewReader(byInven))

				if err != nil {
					fmt.Println("error", err.Error(), res.String())
					return
				}

				if res.StatusCode != 201 {
					fmt.Println("error", res.String())
					return
				}

			}

			// mutex.Lock()
		}

		// mutex.Unlock()
		if err != nil {
			fmt.Println("error", err.Error())
			return
		}

		lenData = len(inventaris)

		if lenData == 0 && isFirstQuery {

			inventarisCheckFirstDate := []models.Inventaris{}
			// get first data from db
			txCheckFirstDate := s.db.
				Model(new(models.Inventaris)).
				Order("updated_at ASC").
				Limit(take).
				Find(&inventarisCheckFirstDate)

			if txCheckFirstDate.Error != nil {
				fmt.Println("error", "getting first date data from inventaris", txCheckFirstDate.Error.Error())
				return
			}

			if len(inventarisCheckFirstDate) > 0 {
				lastTimePull = *inventarisCheckFirstDate[0].UpdatedAt

				lenData = take
			}

		}
		isFirstQuery = false
		page++
	}

	fmt.Println("end performing elastic ", time.Now())

}

func (s *SyncInventaris) CountInventaris() {
	// set count inventaris all

	var countAll int64

	txCountAll := s.db.Table(new(models.Inventaris).TableName()).Count(&countAll)

	if txCountAll.Error != nil {
		fmt.Println("error pipeline", txCountAll.Error.Error())
		return
	}

	err := s.redisCache.Set(&cache.Item{
		Ctx:   context.TODO(),
		Key:   "inventaris-count-all",
		Value: countAll,
		TTL:   time.Hour,
	})

	if err != nil {
		fmt.Println("error set cache", err.Error())
		return
	}

}
