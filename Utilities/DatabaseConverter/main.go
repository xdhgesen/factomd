package DatabaseConverter

import (
	"flag"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/badgerdb"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
)

var (
	DBTypes = []string{"level", "bolt", "badger", "map"}
)

func DBTypeString() string {
	return fmt.Sprintf("%v", DBTypes)
}

func main() {
	var (
		db1loc  = flag.String("db1loc", "", "Location of the first database")
		db1type = flag.String("db1type", "", "Type of the first database: "+DBTypeString())

		db2loc  = flag.String("db2loc", "", "Location of the second database")
		db2type = flag.String("db2type", "", "Type of the second database: "+DBTypeString())

		convertmod = flag.String("m", "", "Mode can be 'entry', ''")
	)
	flag.Parse()

	if *convertmod == "" {
		fmt.Println(usuage())
		fmt.Println("Mode is not defined")
		return
	}

	db1, err := OpenDB(*db1loc, *db1type)
	if err != nil {
		fmt.Println(usuage())
		fmt.Println(err.Error())
		return
	}

	db2, err := OpenDB(*db2loc, *db2type)
	if err != nil {
		fmt.Println(usuage())
		fmt.Println(err.Error())
		return
	}

	a := databaseOverlay.NewOverlay(db1)
	b := databaseOverlay.NewOverlay(db2)

	if *convertmod == "entry" {
		err := ConvertEntries(a, b)
		if err != nil {
			fmt.Println(usuage())
			fmt.Println(err.Error())
			return
		}
	}

}

func ConvertEntries(a, b interfaces.DBOverlay) error {

	return nil
}

func usuage() string {
	return fmt.Sprintf("EntrySteal -db1loc=$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db -db1type=level -db2loc=$HOME/.factom/m2/main-database/badger/MAIN/factoid_badger.db -db2type=badger")
}

func OpenDB(loc, t string) (interfaces.IDatabase, error) {
	switch t {
	case "level":
		return leveldb.NewLevelDB(loc, true)
	case "bolt":
		return boltdb.NewBoltDB(nil, loc), nil
	case "badger":
		return badgerdb.NewBadgerDB(loc)
	case "map":
		return new(mapdb.MapDB), nil

	}
	return nil, fmt.Errorf("Expect %s as db types", DBTypeString())
}
