package mem

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMemDb_New(t *testing.T) {
	t.Run("should create non-nil db", func(t *testing.T) {
		db := New()

		if db == nil {
			t.Errorf("expected non-nil db, got nil")
		}
	})
}

func TestMemDb_Close(t *testing.T) {
	t.Run("should close db", func(t *testing.T) {
		db := New()

		if err := db.Close(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("consecutive calls should fail after close", func(t *testing.T) {
		db := New()

		if err := db.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, err := db.Has([]byte("some_key")); err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestMemDb_Has(t *testing.T) {
	t.Run("should not find key if no key in db", func(t *testing.T) {
		db := New()

		exists, err := db.Has([]byte("some_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got true")
		}
	})

	t.Run("should not find non-existing key", func(t *testing.T) {
		db := New()

		if err := db.Put([]byte("existing_key"), []byte("existing_value")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		exists, err := db.Has([]byte("non_existing_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got true")
		}
	})

	t.Run("should find existing key", func(t *testing.T) {
		db := New()

		if err := db.Put([]byte("existing_key"), []byte("existing_value")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		exists, err := db.Has([]byte("existing_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !exists {
			t.Errorf("expected key to exist, got false")
		}
	})
}

func TestMemDb_Get(t *testing.T) {
	t.Run("should return nil for non-existing key", func(t *testing.T) {
		db := New()

		val, err := db.Get([]byte("non_existing_key"))
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if val != nil {
			t.Errorf("expected val to be nil, got %v", val)
		}
	})

	t.Run("should return val for existing key", func(t *testing.T) {
		db := New()

		key := []byte("key")
		val := []byte("val")
		err := db.Put(key, val)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		res, err := db.Get(key)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !bytes.Equal(res, val) {
			t.Errorf("expected val to be %v, got %v", val, res)
		}
	})
}

func TestMemDb_Put(t *testing.T) {
	t.Run("should insert key without error", func(t *testing.T) {
		db := New()

		err := db.Put([]byte("key"), []byte("val"))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should get previously stored val", func(t *testing.T) {
		db := New()

		key := []byte("key")
		val := []byte("val")
		err := db.Put(key, val)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		res, err := db.Get(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(res, val) {
			t.Errorf("expected val to be %v, got %v", val, res)
		}
	})

	t.Run("should override val", func(t *testing.T) {
		db := New()

		key := []byte("key")
		first := []byte("first")
		if err := db.Put(key, first); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		second := []byte("second")
		if err := db.Put(key, second); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		res, err := db.Get(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(res, second) {
			t.Errorf("expected val to be %v, got %v", second, res)
		}
	})
}

func TestMemDb_Delete(t *testing.T) {
	t.Run("should delete without error", func(t *testing.T) {
		db := New()

		key := []byte("key")
		if err := db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err := db.Delete(key); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should delete existing key", func(t *testing.T) {
		db := New()
		key := []byte("key")

		if err := db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err := db.Delete(key); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		exists, err := db.Has(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got true")
		}
	})
}

func TestMemDb_Batch(t *testing.T) {
	t.Run("should insert key-value pair without error", func(t *testing.T) {
		db := New()

		b := db.NewBatch()
		if err := b.Put([]byte("key"), []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := b.Write(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should write changes only after 'Write' is called", func(t *testing.T) {
		db := New()

		key := []byte("key")
		val := []byte("val")

		b := db.NewBatch()
		if err := b.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, err := db.Get(key); err == nil {
			t.Errorf("expected not found error, got nil")
		}
		if err := b.Write(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		res, err := db.Get(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(res, val) {
			t.Errorf("expected val to be %v, got %v", val, res)
		}
	})

	t.Run("should fail to write batch if db is closed", func(t *testing.T) {
		db := New()
		if err := db.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b := db.NewBatch()
		if err := b.Put([]byte("key"), []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := b.Write(); err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should delete key", func(t *testing.T) {
		db := New()

		key := []byte("key")
		val := []byte("val")

		if err := db.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b := db.NewBatch()
		if err := b.Delete(key); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := b.Write(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		exists, err := db.Has(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got true")
		}
	})

	t.Run("should clear batch", func(t *testing.T) {
		db := New()

		b := db.NewBatch()
		if err := b.Put([]byte("key"), []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b.Reset()
		if size := b.ValueSize(); size != 0 {
			t.Errorf("expected batch size to be 0 after reset, got %d", size)
		}
	})

	t.Run("should replay batch contents", func(t *testing.T) {
		db := New()

		delKey := []byte("del_key")
		if err := db.Put(delKey, []byte("del_val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b := db.NewBatch()
		if err := b.Delete(delKey); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		key := []byte("key")
		val := []byte("val")
		if err := b.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := b.Replay(db); err != nil {
			t.Fatalf("expected no error during replay, got %v", err)
		}
		delExists, err := db.Has(delKey)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if delExists {
			t.Errorf("expected key to not exist after replay, got true")
		}
		res, err := db.Get(key)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !bytes.Equal(res, val) {
			t.Errorf("expected val to be %v, got %v", val, res)
		}
	})
}

func TestMemDb_Iterator(t *testing.T) {
	t.Run("should be exhausted if empty db", func(t *testing.T) {
		db := New()

		it := db.NewIterator(nil, nil)
		defer it.Release()

		if it.Next() {
			t.Errorf("expected iterator to be exhausted, got next item")
		}
	})

	t.Run("should be exhausted if no keys match", func(t *testing.T) {
		db := New()

		if err := db.Put([]byte("first"), []byte("first_val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := db.Put([]byte("second"), []byte("second_val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		it := db.NewIterator([]byte("non_existing"), []byte("non_existing"))
		defer it.Release()

		if it.Next() {
			t.Errorf("expected iterator to be exhausted, got next item")
		}
	})

	t.Run("should iterate without errors", func(t *testing.T) {
		db := New()

		items := 10
		for i := 0; i < items; i++ {
			key := []byte(fmt.Sprintf("key-%d", i))
			val := []byte(fmt.Sprintf("val-%d", i))
			if err := db.Put(key, val); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		it := db.NewIterator(nil, nil)
		defer it.Release()

		for it.Next() {
			if it.Error() != nil {
				t.Errorf("expected no error during iteration, got %v", it.Error())
			}
		}
	})

	t.Run("should iterate over all keys if nil range", func(t *testing.T) {
		db := New()

		items := 10
		for i := 0; i < items; i++ {
			key := []byte(fmt.Sprintf("key-%d", i))
			val := []byte(fmt.Sprintf("val-%d", i))
			if err := db.Put(key, val); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		it := db.NewIterator(nil, nil)
		defer it.Release()

		count := 0
		for it.Next() {
			count++
		}

		if count != items {
			t.Errorf("expected %d items, got %d", items, count)
		}
	})

	t.Run("should iterate in binary-alphabetical order", func(t *testing.T) {
		db := New()

		items := map[string][]byte{
			"alpha":   []byte("alpha_val"),
			"bravo":   []byte("bravo_val"),
			"charlie": []byte("charlie_val"),
			"delta":   []byte("delta_val"),
		}

		for key, val := range items {
			if err := db.Put([]byte(key), val); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		it := db.NewIterator(nil, nil)
		defer it.Release()

		expected := []string{"alpha", "bravo", "charlie", "delta"}
		for i := 0; it.Next(); i++ {
			key := string(it.Key())
			if key != expected[i] {
				t.Errorf("expected key %v, got %v", expected[i], key)
			}

			val := it.Value()
			if !bytes.Equal(val, items[key]) {
				t.Errorf("expected value for %v to be %v, got %v", key, items[key], val)
			}
		}
	})

	t.Run("should skip keys before start", func(t *testing.T) {
		db := New()

		items := map[string][]byte{
			"alpha":   []byte("alpha_val"),
			"bravo":   []byte("bravo_val"),
			"charlie": []byte("charlie_val"),
			"delta":   []byte("delta_val"),
		}

		for key, val := range items {
			if err := db.Put([]byte(key), val); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		it := db.NewIterator(nil, []byte("charlie"))
		defer it.Release()

		expected := []string{"charlie", "delta"}
		for i := 0; it.Next(); i++ {
			key := string(it.Key())
			if key != expected[i] {
				t.Errorf("expected key %v, got %v", expected[i], key)
			}

			val := it.Value()
			if !bytes.Equal(val, items[key]) {
				t.Errorf("expected value for %v to be %v, got %v", key, items[key], val)
			}
		}
	})
}
