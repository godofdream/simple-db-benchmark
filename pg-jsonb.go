package main

import "fmt"
import _ "github.com/lib/pq"
import "github.com/jmoiron/sqlx"
import "time"
import "log"

func main() {
	
	db := sqlx.MustConnect(`postgres`, `user=test2 dbname=test2 sslmode=disable`)
	var err error
	fmt.Println(`test2: postgresql jsonb`)
	_, err = db.Exec(`CREATE TABLE test2 (id BIGSERIAL PRIMARY KEY, k TEXT UNIQUE, v JSONB)`)
	if err != nil {
		_, err = db.Exec(`TRUNCATE TABLE test2`)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	
	const max = 9999
	const jump = 40
	t := time.Now()
	for x := 1; x <= max; x++ {
		_, err = db.Exec(fmt.Sprintf(`INSERT INTO test2(k,v)VALUES('%05d','{"v":"%05d"}')`, x, x))
		if err != nil {
			log.Fatal(err)
			return
		}
		if x % 200 == 0 {
			fmt.Print(`.`)
		}
	}
	dur := time.Now().Sub(t)
	fmt.Printf("INSERT: %v (%.2f ms/op)\n", dur, float64(dur.Nanoseconds()) / 1000000 / max)
	
	t = time.Now()
	for x := 1; x <= max; x++ {
		_, err = db.Exec(fmt.Sprintf(`UPDATE test2 SET v = v || '{"v":"%06d"}' WHERE k = '%05d'`, x, x))
		if err != nil {
			log.Fatal(err)
			return
		}
		if x % 200 == 0 {
			fmt.Print(`.`)
		}
	}
	dur = time.Now().Sub(t)
	fmt.Printf("UPDATE: %v (%.2f ms/op)\n", dur, float64(dur.Nanoseconds()) / 1000000 / max)
	
	t = time.Now()
	ops := int64(0)
	for y := 2; y < jump; y++ {
		for x := max - 1; x > 0; x -= y {
			ops++
			rows, err := db.Queryx(fmt.Sprintf(`SELECT id, k, v->>'v' FROM test2 WHERE k >= '%05d' ORDER BY k ASC LIMIT %d`, x, y * y))
			if err != nil {
				log.Fatal(err)
				return
			}
			for rows.Next() {
				m := map[string]interface{}{}
				rows.MapScan(m)
			}
			rows.Close()
			if ops % 500 == 0 {
				fmt.Print(`.`)
			}
		}
		for x := 1; x < max; x += y {
			ops++
			rows, err := db.Queryx(fmt.Sprintf(`SELECT id, k, v->>'v' FROM test2 WHERE k <= '%05d' ORDER BY k DESC LIMIT %d`, x, y * y))
			if err != nil {
				log.Fatal(err)
				return
			}
			for rows.Next() {
				m := map[string]interface{}{}
				rows.MapScan(m)
			}
			rows.Close()
			if ops % 500 == 0 {
				fmt.Print(`.`)
			}
		}
	}
	dur = time.Now().Sub(t)
	fmt.Printf("SELECT: %v (%.2f ms/op)\n", dur, float64(dur.Nanoseconds()) / 1000000 / float64(ops))
}

