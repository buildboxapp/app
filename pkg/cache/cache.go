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
	"time"
)

type cache struct {
	DB *reindexer.Reindexer
	cfg config.Config
	logger log.Log
	function function.Function
	active bool `json:"active"`
}

type Cache interface {
	Active() bool
	GenKey(uid, path, query, ignorePath, ignoreQuery string) (key, keyParam string)
	SetStatus(key, status string) (err error)
	Read(key string) (result, status string, fresh bool, err error)
	Write(key string, cacheInterval int, blockUid, pageUid string, value, url string) (err error)
}

// проверяем статус соединения с базой
func (c *cache) Active() bool {
	if c.active {
		return true
	}
	return false
}

// формируем ключ кеша
// ignorePath, ignoreQuery - признаки игнорирования пути или запроса в ключе (указываются в кеше блока)
func (c *cache) GenKey(uid, path, query, ignorePath, ignoreQuery string) (key, keyParam string)  {
	key2 := ""
	key3 := ""

	// формируем сложный ключ-хеш
	key1, _ := json.Marshal(uid)
	key2 = path // переводим в текст параметры пути запроса (/nedra/user)
	key3 = fmt.Sprintf("%v", query) // переводим в текст параметры строки запроса (?sdf=df&df=df)

	// учитываем путь и параметры
	if ignorePath == "" && ignoreQuery == "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + c.function.TplFunc().Hash(string(key2)) + "_" + c.function.TplFunc().Hash(string(key3))
	}

	// учитываем только путь
	if ignorePath != "" && ignoreQuery == "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + c.function.TplFunc().Hash(string(key2)) + "_"
	}

	// учитываем только параметры
	if ignorePath == "" && ignoreQuery != "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + "_" + c.function.TplFunc().Hash(string(key3))
	}

	// учитываем путь и параметры
	if ignorePath != "" && ignoreQuery != "" {
		key = c.function.TplFunc().Hash(string(key1)) + "_" + "_"
	}

	return key, "url:"+key2+"; params:"+key3
}

// меняем статус по ключу
// для решения проблему дублированного обновления кеша первый, кто инициирует обновление кеша меняет статус на updated
// и меняет время в поле Deadtime на время изменения статуса + 2 минуты = максимальное время ожидания обновления кеша
// это сделано для того, чтобы не залипал кеш, у которых воркер который решил его обновить Отвалился, или был передернут сервис
// таким образом, запрос, который получает старый кеш у которого статут updated проверяем время старта обновления и если оно просрочено
// то сам инициирует обновление кеша (меняя время на свое)
func (c *cache) SetStatus(key, status string) (err error) {
	var rows *reindexer.Iterator
	var deadTime = time.Now().UTC().Add(c.cfg.TimeoutCacheGenerate.Value)	// время, когда статус updated перестанет быть валидным

	rows = c.DB.Query(c.cfg.Namespace).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()

	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*model.ValueCache)

		// меняем статус
		elem.Status = status
		if status == "updated" {
			elem.Deadtime = deadTime.String()
		}
		err = c.DB.Upsert(c.cfg.Namespace, elem)
	}

	rows.Close()
	return
}

// key - ключ, который будет указан в кеше
// получаем:
// result, status - результат и статус (текст)
// fresh - признак того, что данные актуальны (свежие)
func (c *cache) Read(key string) (result, status string, flagExpired bool, err error)  {
	var rows *reindexer.Iterator

	rows = c.DB.Query(c.cfg.Namespace).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()

	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*model.ValueCache)
		result = elem.Value

		// функция Timefresh показывает пора ли обновить время (не признак свежести, а наоборот)
		// оставил для совместимости со сторыми версиями
		flagExpired = c.function.TplFunc().TimeExpired(elem.Deadtime)
	}

	rows.Close()
	return
}


// key - ключ, который будет указан в кеше
// cacheInterval - время хранени кеша
// blockUid, pageUid - ид-ы блока и страницы (для формирования возможности выборочного сброса кеша)
// data - то, что кладется в кеш
func (c *cache) Write(key string, cacheInterval int, blockUid, pageUid string, value, url string) (err error) {
	var valueCache = model.ValueCache{}
	var deadTime time.Duration

	// интервал не указан - значит не кешируем (не пишем в кеш)
	if cacheInterval == 0 {
		return fmt.Errorf("%s", "Cache interval is empty")
	}

	valueCache.Uid = key
	valueCache.Value = value
	deadTime = time.Minute * time.Duration(cacheInterval)
	dt := time.Now().UTC().Add(deadTime)

	// дополнитлельные ключи для поиска кешей страницы и блока (отдельно)
	var link []string

	link = append(link, pageUid)
	link = append(link, blockUid)

	valueCache.Link = link
	valueCache.Url = url
	valueCache.Deadtime = dt.String()
	valueCache.Status = ""

	err = c.DB.Upsert(c.cfg.Namespace, valueCache)
	if err != nil {
		c.logger.Error(err, "Error! Created cache from is failed!")
		return fmt.Errorf("%s", "Error! Created cache from is failed!")
	}

	return
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
			fmt.Printf("%s Cache-service is running\n", done)
			logger.Info("Cache-service is running")
			cach.active = true
		}
	}

	return &cach
}