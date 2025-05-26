package badger

import (
	"bytes"
	"testing"
)

func TestBadgerDb_New(t *testing.T) {
	t.Run("should create non-nil db", func(t *testing.T) {
		db, err := New(t.TempDir())

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if db == nil {
			t.Errorf("expected non-nil db, got nil")
		}
	})
}

func TestBadgerDb_Close(t *testing.T) {
	t.Run("should close db", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err = db.Close(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("consecutive calls should fail after close", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err = db.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, err = db.Has([]byte("some_key")); err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestBadgerDb_Has(t *testing.T) {
	t.Run("should not find key if no key in db", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		exists, err := db.Has([]byte("some_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got true")
		}
	})

	t.Run("should not find non-existing key", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		if err = db.Put([]byte("existing_key"), []byte("existing_value")); err != nil {
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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		if err = db.Put([]byte("existing_key"), []byte("existing_value")); err != nil {
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

func TestBadgerDb_Get(t *testing.T) {
	t.Run("should return nil for non-existing key", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		val, err := db.Get([]byte("non_existing_key"))
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if val != nil {
			t.Errorf("expected val to be nil, got %v", val)
		}
	})

	t.Run("should return val for existing key", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		val := []byte("val")
		if err = db.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
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

func TestBadgerDb_Put(t *testing.T) {
	t.Run("should insert key without error", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		if err = db.Put([]byte("key"), []byte("val")); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should get previously stored val", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		val := []byte("val")
		if err = db.Put(key, val); err != nil {
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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		first := []byte("first")
		if err = db.Put(key, first); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		second := []byte("second")
		err = db.Put(key, second)
		if err != nil {
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

func TestBadgerDb_Delete(t *testing.T) {
	t.Run("should delete without error", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		if err = db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err = db.Delete(key); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should delete existing key", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		if err = db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err = db.Delete(key); err != nil {
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

func TestBadgerDb_Batch(t *testing.T) {
	t.Run("should insert key-value pair without error", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		b := db.NewBatch()
		if err := b.Put([]byte("key"), []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := b.Write(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should write changes only after 'Write' is called", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		val := []byte("val")

		b := db.NewBatch()
		if err = b.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, err = db.Get(key); err == nil {
			t.Errorf("expected not found error, got nil")
		}
		if err = b.Write(); err != nil {
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

	t.Run("should delete key", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		key := []byte("key")
		val := []byte("val")

		if err = db.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b := db.NewBatch()
		if err = b.Delete(key); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err = b.Write(); err != nil {
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
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		b := db.NewBatch()
		if err = b.Put([]byte("key"), []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b.Reset()
		if size := b.ValueSize(); size != 0 {
			t.Errorf("expected batch size to be 0 after reset, got %d", size)
		}
	})

	t.Run("should replay batch contents", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		delKey := []byte("del_key")
		if err = db.Put(delKey, []byte("del_val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		b := db.NewBatch()
		if err = b.Delete(delKey); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		key := []byte("key")
		val := []byte("val")
		if err = b.Put(key, val); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err = b.Replay(db); err != nil {
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
