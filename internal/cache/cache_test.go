package cache

import (
	"OrderInfo/internal/model"
	"reflect"
	"testing"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache()

	orderUID := "test-uid"
	expectedOrder := model.Order{OrderUID: orderUID}

	cache.Set(orderUID, expectedOrder)

	retrievedOrder, found := cache.Get(orderUID)

	if !found {
		t.Errorf("Get(%s) returned found=false, want true", orderUID)
	}

	if !reflect.DeepEqual(retrievedOrder, expectedOrder) {
		t.Errorf("Get(%s) returned order %+v, want %+v", orderUID, retrievedOrder, expectedOrder)
	}
}

func TestInMemoryCache_GetNotFound(t *testing.T) {
	cache := NewInMemoryCache()

	_, found := cache.Get("non-existent-uid")

	if found {
		t.Errorf("Get(non-existent-uid) returned found=true, want false")
	}
}
