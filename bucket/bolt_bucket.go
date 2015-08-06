package bucket

import "github.com/boltdb/bolt"

const (
	DefaultBucket = "default"
)

type LocalBucket struct {
	dbName string
	db     *bolt.DB
}

func (b *LocalBucket) Open() error {
	db, err := bolt.Open(b.dbName, 0600, nil)
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

func (b *LocalBucket) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *LocalBucket) getBucketName(v ...string) string {
	if len(v) > 0 {
		return v[0]
	}
	return DefaultBucket
}

func (b *LocalBucket) First(bucketName ...string) (string, []byte, error) {
	var value []byte
	var key string

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.getBucketName(bucketName...)))
		if bucket != nil {
			c := bucket.Cursor()
			if c != nil {
				k, v := c.First()
				if k != nil {
					key = string(k)
					value = v
				}
			}
			return nil
		}
		return bolt.ErrBucketNotFound
	})

	return key, value, err
}

func (b *LocalBucket) Put(key string, data []byte, bucketName ...string) error {
	bn := b.getBucketName(bucketName...)
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bn))
		if err != nil {
			return err
		}

		bucket := tx.Bucket([]byte(bn))
		if bucket != nil {
			return bucket.Put([]byte(key), []byte(data))
		}

		return bolt.ErrBucketNotFound
	})
}

func (b *LocalBucket) Get(key string, bucketName ...string) ([]byte, error) {
	var v []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.getBucketName(bucketName...)))
		if bucket != nil {
			v = bucket.Get([]byte(key))
			return nil
		}
		return bolt.ErrBucketNotFound
	})

	return v, err
}

func (b *LocalBucket) BucketKeys(bucketName ...string) ([]string, error) {
	keys := []string{}
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.getBucketName(bucketName...)))
		if bucket != nil {
			return bucket.ForEach(func(k []byte, v []byte) error {
				keys = append(keys, string(k))
				return nil
			})
		}
		return bolt.ErrBucketNotFound
	})

	return keys, err
}

func (b *LocalBucket) BucketNames() ([]string, error) {
	buckets := []string{}
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})

	return buckets, err
}

func (b *LocalBucket) CreateBucket(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(name))
		return err
	})
}

func (b *LocalBucket) RemoveBucket(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
}

func (b *LocalBucket) Remove(key string, bucketName ...string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.getBucketName(bucketName...)))
		if bucket != nil {
			return bucket.Delete([]byte(key))
		}
		return bolt.ErrBucketNotFound
	})
}

func (b *LocalBucket) RemoveAll() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}

func NewLocalBucket(dbName string) *LocalBucket {
	return &LocalBucket{dbName: dbName}
}
