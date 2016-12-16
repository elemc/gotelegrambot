package db

import (
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Cache is struct for store date caches
type Cache struct {
	Years        []string
	MonthsByYear map[int][]time.Month
	mutex        *sync.Mutex
}

// Caches is a map cache to chat ID
type Caches map[int64]*Cache

// CreateNewCache function create new Cache pointer
func CreateNewCache() *Cache {
	cache := new(Cache)
	cache.mutex = new(sync.Mutex)
	cache.MonthsByYear = make(map[int][]time.Month)
	return cache
}

// AddedDateToCaches added date to caches
func AddedDateToCaches(chatID int64, d time.Time) {
	if _, ok := caches[chatID]; !ok {
		caches[chatID] = CreateNewCache()
	}

	cache := caches[chatID]

	cache.mutex.Lock()
	strYear := strconv.Itoa(d.Year())
	month := d.Month()

	cache.Years = appendIfNotFound(cache.Years, strYear)
	cache.MonthsByYear[d.Year()] = appendIfNotFoundMonth(cache.MonthsByYear[d.Year()], month)
	cache.mutex.Unlock()
}

func updateDateCaches() {
	chats, err := GetChats()
	if err != nil {
		return
	}
	for _, chat := range chats {
		chatID := chat.ID
		listDates, err := getDates(chatID, 0, 0)
		if err != nil {
			return
		}
		for _, t := range listDates {
			AddedDateToCaches(chatID, t)
		}
	}
	log.Printf("Time caches updated.")
	//go updateDateCaches()
}

// GetCache function returns Cache pointer by Chat ID
func getCache(chatID int64) *Cache {
	if cache, ok := caches[chatID]; ok {
		return cache
	}

	cache := CreateNewCache()
	caches[chatID] = cache
	return cache
}

func sortMonths(a []time.Month) (result []time.Month) {
	var temp []int
	for _, value := range a {
		temp = append(temp, int(value))
	}
	sort.Ints(temp)
	for _, value := range temp {
		result = append(result, time.Month(value))
	}
	return
}
