package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
    require.NotEmpty(t, id)
	// get
	new, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, new.Client)
	assert.Equal(t, parcel.Status, new.Status)
	assert.Equal(t, parcel.Address, new.Address)
	assert.Equal(t, parcel.CreatedAt, new.CreatedAt)
	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
    require.NotEmpty(t, id)
	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)
	// check
	new, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, new.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
    require.NotEmpty(t, id)
	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)
	// check
	new, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, new.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
    	require.NotEmpty(t, id)
		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))
	// check
	for _, parcel := range storedParcels {
		orig, ok := parcelMap[parcel.Number]
		require.True(t, ok, "parcel with number %d not found in parcelMap", parcel.Number)
		require.Equal(t, orig.Client, parcel.Client)
        require.Equal(t, orig.Status, parcel.Status)
        require.Equal(t, orig.Address, parcel.Address)
        require.Equal(t, orig.CreatedAt, parcel.CreatedAt)
	}
}
