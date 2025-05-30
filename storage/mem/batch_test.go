package mem

import (
	"bytes"
	"testing"
)

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
