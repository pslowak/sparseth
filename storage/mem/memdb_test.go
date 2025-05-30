package mem

import (
	"bytes"
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

func TestMemDb_DeleteRange(t *testing.T) {
	t.Run("should delete without error", func(t *testing.T) {
		db := New()

		key := []byte("key")
		if err := db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err := db.DeleteRange(key, []byte("zzz")); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should delete existing key", func(t *testing.T) {
		db := New()
		key := []byte("key")

		if err := db.Put(key, []byte("val")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err := db.DeleteRange(key, []byte("zzz")); err != nil {
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

	t.Run("should only delete keys within range", func(t *testing.T) {
		db := New()

		// Indicates if key should exist
		// after range deletion
		keys := map[string]bool{
			"alpha":   true,
			"bravo":   false,
			"charlie": false,
			"delta":   true,
		}

		for k := range keys {
			if err := db.Put([]byte(k), []byte("val")); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		if err := db.DeleteRange([]byte("bravo"), []byte("delta")); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		for k, shouldExist := range keys {
			exists, err := db.Has([]byte(k))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if exists != shouldExist {
				t.Errorf("expected key %s to exist: %v, got %v", k, shouldExist, exists)
			}
		}
	})
}
