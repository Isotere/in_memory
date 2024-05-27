package in_memory

import (
	"fmt"
	"sync"
	"time"
)

var (
	once sync.Once

	// Это что-то типа внешнего хранилища у нас, инициализируем только один раз, при запуске программы, в памяти
	storageInstance *storage
)

type InMemoryConnection struct {
	timeoutMs       time.Duration
	storageInstance *storage
}

func NewConnection(timeoutMs uint) *InMemoryConnection {
	// Инициализируем "базу данных"
	// Для того, чтобы убедиться, что она будет всего один раз создана и не использовать при этом init
	// используем для этого пакет sync
	once.Do(func() {
		fmt.Println("Init empty database, run once at programm start")

		storageInstance = newStorage()
	})

	return &InMemoryConnection{
		timeoutMs:       time.Millisecond * time.Duration(timeoutMs),
		storageInstance: storageInstance,
	}
}

func (c *InMemoryConnection) GenerateQueryID() uint64 {
	return c.storageInstance.newQueryID()
}

func (c *InMemoryConnection) Collections() *CollectionsList {
	return c.storageInstance.collections
}
