package main

import (
	"math/rand"

	"time"

	"fmt"

	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/badgerdb"
	"github.com/FactomProject/factomd/database/leveldb"
)

const RandomSeed int64 = 0

type Report struct {
	Title       string
	Description string

	ExteriorTime time.Duration
	TotalTime    time.Duration
	Operations   int64
}

func NewReport() *Report {
	r := new(Report)

	return r
}

func (r *Report) AddOp(duration time.Duration) {
	r.TotalTime += duration
	r.Operations++
}

func (r *Report) OpTime() time.Duration {
	return r.TotalTime / time.Duration(r.Operations)
}

func (r *Report) String() string {
	return fmt.Sprintf("%s\n%s\nTime: %s\nOpTime: %s\nOps: %d\n",
		r.Title, r.Description, r.ExteriorTime, r.OpTime(), r.Operations)
}

func main() {
	amt := int64(10)
	ks := int64(32)
	vs := int64(2048)

	bad, _ := badgerdb.NewBadgerDB("badgerdb")
	r := RandomKeyWriteSpeed(bad, amt, ks, vs)
	r.Title = "Badger"
	fmt.Println(r)
	Cleanup(bad, "badgerdb")

	lev, _ := leveldb.NewLevelDB("leveldb", true)
	r = RandomKeyWriteSpeed(bad, amt, ks, vs)
	r.Title = "LevelDB"
	fmt.Println(r)
	Cleanup(lev, "leveldb")
}

func Cleanup(m interfaces.IDatabase, path string) {
	//m.Close()
	os.RemoveAll(path)
	os.Remove(path)
}

func RandomKeyWriteSpeed(db interfaces.IDatabase, n, keySize, valueSize int64) *Report {
	report := NewReport()
	report.Description = fmt.Sprintf("WriteOnly - Keysize: %d, ValuSize: %d", keySize, valueSize)
	rand.Seed(RandomSeed)
	start := time.Now()
	for i := int64(0); i < n; i++ {
		os := time.Now()
		key := make([]byte, keySize)
		value := make([]byte, valueSize)
		rand.Read(key)
		rand.Read(value)

		db.Put([]byte{}, key, &TestMarshalByte{value})
		report.AddOp(time.Since(os))
	}
	report.ExteriorTime = time.Since(start)
	return report
}

type TestMarshalByte struct {
	Data []byte
}

func (t *TestMarshalByte) New() interfaces.BinaryMarshallableAndCopyable {
	return new(TestMarshalByte)
}

func (t *TestMarshalByte) MarshalBinary() ([]byte, error) {
	return t.Data, nil
}

func (t *TestMarshalByte) UnmarshalBinaryData(data []byte) ([]byte, error) {
	t.Data = data
	return nil, nil
}

func (t *TestMarshalByte) UnmarshalBinary(data []byte) (err error) {
	_, err = t.UnmarshalBinaryData(data)
	return
}
