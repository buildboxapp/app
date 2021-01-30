package cache

import "C"
import (
	"fmt"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
	"github.com/labstack/gommon/color"

	"encoding/json"
	"github.com/restream/reindexer"
	"strconv"
	"time"
)

type cache struct {
	DB *reindexer.Reindexer
	cfg config.Config
	logger log.Log
	function function.Function
}

type Cache interface {
	SetCahceKey(p model.Data, path, query string) (key, keyParam string)
	СacheGet(key string, block model.Data, page model.Data, values map[string]interface{}, url string) (string, bool)
	CacheSet(key string, block model.Data, page model.Data, value, url string) bool
	CacheUpdate(key string, block model.Data, page model.Data, values map[string]interface{}, url string)
	RefreshTime(options model.Data) int
}

// формируем ключ кеша
func (c *cache) SetCahceKey(p model.Data, path, query string) (key, keyParam string)  {
	key2 := ""
	key3 := ""

	// формируем сложный ключ-хеш
	key1, _ := json.Marshal(p.Uid)
	key2 = path // переводим в текст параметры пути запроса (/nedra/user)
	key3 = fmt.Sprintf("%v", query) // переводим в текст параметры строки запроса (?sdf=df&df=df)

	cache_nokey2, _ := p.Attr("cache_nokey2", "value")
	cache_nokey3, _ := p.Attr("cache_nokey3", "value")

	// учитываем путь и параметры
	if cache_nokey2 == "" && cache_nokey3 == "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + c.function.TplFunc().Hash(string(key2)) + "_" + c.function.TplFunc().Hash(string(key3))
	}

	// учитываем только путь
	if cache_nokey2 != "" && cache_nokey3 == "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + c.function.TplFunc().Hash(string(key2)) + "_"
	}

	// учитываем только параметры
	if cache_nokey2 == "" && cache_nokey3 != "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + "_" + c.function.TplFunc().Hash(string(key3))
	}

	// учитываем путь и параметры
	if cache_nokey2 != "" && cache_nokey3 != "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + "_"
	}

	return key, "url:"+key2+"; params:"+key3
}

// key - ключ, который будет указан в кеше
// option - объект блока (запроса и тд) то, где хранится время кеширования
func (c *cache) СacheGet(key string, block model.Data, page model.Data, values map[string]interface{}, url string) (string, bool)  {
	var res string
	var rows *reindexer.Iterator

	rows = c.DB.Query(c.cfg.Namespace).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()


	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*model.ValueCache)
		res = elem.Value

		flagFresh := c.function.TplFunc().Timefresh(elem.Deadtime)

		if flagFresh == "true" {

			// блокируем запись, чтобы другие процессы не стали ее обновлять также
			if elem.Status != "updating" {

				if 	f := c.RefreshTime(block); f == 0 {
					return "", false
				}

				// меняем статус
				elem.Status = "updating"
				c.DB.Upsert(c.cfg.Namespace, elem)

				// запускаем обновение кеша фоном
				go c.CacheUpdate(key, block, page, values, url)
			}
		}

		//fmt.Println("Отдали из кеша")

		return res, true
	}

	//fmt.Println("Нет в кеша")

	return "", false
}


// key - ключ, который будет указан в кеше
// option - объект блока (запроса и тд) то, где хранится время кеширования
// data - то, что кладется в кеш
func (c *cache) CacheSet(key string, block model.Data, page model.Data, value, url string) bool {
	var valueCache = model.ValueCache{}
	var deadTime time.Duration

	// если интервал не задан, то не кешируем
	f := c.RefreshTime(block)

	//log.Warning("block: ", block)
	if f == 0 {
		return false
	}

	valueCache.Uid = key
	valueCache.Value = value

	deadTime = time.Minute * time.Duration(f)
	dt := time.Now().UTC().Add(deadTime)

	// дополнитлельные ключи для поиска кешей страницы и блока (отдельно)
	var link []string

	link = append(link, page.Uid)
	link = append(link, block.Uid)

	valueCache.Link = link
	valueCache.Url = url
	valueCache.Deadtime = dt.String()
	valueCache.Status = ""

	err := c.DB.Upsert(c.cfg.Namespace, valueCache)
	if err != nil {
		c.logger.Error(err, "Error! Created cache from is failed! ")
		return false
	}

	//fmt.Println("Пишем в кеш")


	return true
}

func (c *cache) CacheUpdate(key string, blk model.Data, page model.Data, values map[string]interface{}, url string) {
	//var md = block.New()
	//// получаем контент модуля
	//value := md.Generate(blk, page, values, false)

	// обновляем кеш
	//c.CacheSet(key, blk, page, string(value.Result), url)

	return
}

func (c *cache) RefreshTime(options model.Data) int {

	refresh, _ := options.Attr("cache", "value")
	if refresh == "" {
		return 0
	}

	f, err := strconv.Atoi(refresh)
	if err != nil {
		return 0
	}

	return f
}

func New(cfg config.Config, logger log.Log, function function.Function) Cache {
	done := color.Green("[OK]")
	fail := color.Red("[Fail]")
	var cach = cache{
		cfg: cfg,
		logger: logger,
		function: function,
	}

	// включено кеширование
	if cfg.CachePointsrc != "" {
		cach.DB = reindexer.NewReindex(cfg.CachePointsrc)
		err := cach.DB.OpenNamespace(cfg.Namespace, reindexer.DefaultNamespaceOptions(), model.ValueCache{})
		if err != nil {
			fmt.Printf("%s Error connecting to database. Plaese check this parameter in the configuration. %s\n", fail, cfg.CachePointsrc)
			fmt.Printf("%s\n", err)
			logger.Error(err, "Error connecting to database. Plaese check this parameter in the configuration: ", cfg.CachePointsrc)
			return &cach
		} else {
			fmt.Printf("%s Cache-service is running", done)
			logger.Info("Cache-service is running")
			cfg.BaseCache = "on"
		}
	}

	return &cach
}