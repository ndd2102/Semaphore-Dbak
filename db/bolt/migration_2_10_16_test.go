package bolt

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"testing"
)

func TestMigration_2_10_16_Apply(t *testing.T) {
	store := CreateTestStore()

	err := store.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("0000000001"), []byte("{}"))
		if err != nil {
			return err
		}

		r, err := tx.CreateBucketIfNotExists([]byte("project__template_0000000001"))
		if err != nil {
			return err
		}

		err = r.Put([]byte("0000000001"),
			[]byte("{\"id\":\"1\",\"project_id\":\"1\"}"))

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	err = migration_2_10_16{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}

	var repo map[string]interface{}
	err = store.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__template_0000000001"))
		str := string(b.Get([]byte("0000000001")))
		return json.Unmarshal([]byte(str), &repo)
	})
	if err != nil {
		t.Fatal(err)
	}

	if repo["app"] == nil {
		t.Fatal("app must be set")
	}

	if repo["app"].(string) != "ansible" {
		t.Fatal("invalid app: " + repo["app"].(string))
	}

	if repo["alias"] != nil {
		t.Fatal("alias must be deleted")
	}
}

func TestMigration_2_10_16_Apply2(t *testing.T) {
	store := CreateTestStore()

	err := store.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("0000000001"), []byte("{}"))

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	err = migration_2_10_16{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}
}
