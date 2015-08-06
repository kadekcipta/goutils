package bucket

import (
	"fmt"
	"testing"
)

func TestCRUDValue(t *testing.T) {
	b := NewLocalBucket("my.db")
	if err := b.Open(); err != nil {
		t.Error(err)
	}
	defer b.Close()

	n := 10

	for i := 0; i < n; i++ {
		err := b.Put(fmt.Sprintf("Key#%d", i+1), []byte(fmt.Sprintf("Value#%d", i+1)))
		if err != nil {
			t.Error(err)
		}
	}

	k, v, err := b.First()
	if err != nil {
		t.Error(err)
	}
	t.Log(k, string(v))

	keys, err := b.BucketKeys()
	if err != nil {
		t.Error(err)
	}

	if len(keys) != n {
		t.Error("Number of keys not matches")
	}

	for _, k := range keys {
		if err := b.Remove(k); err != nil {
			t.Error(err)
		}
	}

	keys, err = b.BucketKeys()
	if err != nil {
		t.Error(err)
	}

	if len(keys) != 0 {
		t.Error("Should be empty")
	}
}

func TestBuckets(t *testing.T) {
	b := NewLocalBucket("my.db")
	if err := b.Open(); err != nil {
		t.Error(err)
	}
	defer b.Close()

	if err := b.CreateBucket("B1"); err != nil {
		t.Error(err)
	}

	if err := b.CreateBucket("B2"); err != nil {
		t.Error(err)
	}

	names, err := b.BucketNames()
	if err != nil {
		t.Error(err)
	}

	if err := b.RemoveAll(); err != nil {
		t.Error(err)
	}

	names, err = b.BucketNames()
	if err != nil {
		t.Error(err)
	}

	if len(names) != 0 {
		t.Error("Should no buckets")
	}
}
