package badger

import (
	"bytes"
	"fmt"
	"testing"
)

func TestBadgerDb_Iterator(t *testing.T) {
	t.Run("should be exhausted if empty db", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		it := db.NewIterator(nil, nil)
		defer it.Release()

		if it.Next() {
			t.Errorf("expected iterator to be exhausted, got next item")
		}
	})

	t.Run("should be exhausted if no keys match", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

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
