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

func TestBadgerDb_PutBatch(t *testing.T) {
	t.Run("should insert keys without error", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		pairs := map[string][]byte{
			"first_key":  []byte("first_val"),
			"second_key": []byte("second_val"),
		}
		if err = db.PutBatch(pairs); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should get previously stored vals", func(t *testing.T) {
		db, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()

		firstVal := []byte("first_val")
		secondVal := []byte("second_val")
		pairs := map[string][]byte{
			"first_key":  firstVal,
			"second_key": secondVal,
		}
		if err = db.PutBatch(pairs); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		firstRes, err := db.Get([]byte("first_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(firstRes, firstVal) {
			t.Errorf("expected val to be %v, got %v", firstVal, firstRes)
		}

		secondRes, err := db.Get([]byte("second_key"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(secondRes, secondVal) {
			t.Errorf("expected val to be %v, got %v", secondVal, secondRes)
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
