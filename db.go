package main

import (
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/lib/pq"
)

//Функция, возвращающая рестораны, которые могут вместить требуемое количество посетителей
func getRestaurants(db *sql.DB, h, m, persons int) map[int]Restaurant {
	restaurants := make(map[int]Restaurant)
	rows, err := db.Query(fmt.Sprintf("SELECT DISTINCT rests.\"id\", rests.\"Name\", rests.\"WaitTime\", rests.\"Bill\", tab2.avSeats FROM \"Restaurants\" rests JOIN (SELECT tbs.\"Restaurant\",  tbs.\"Number\" FROM \"Tables\" tbs LEFT JOIN \"Timetable\" tt ON (tbs.\"Restaurant\" = tt.\"Restaurant\" AND "+
		"tbs.\"Number\" = tt.\"tableNum\" and ABS(%v*60+%v - (tt.\"hour\"*60+tt.\"minutes\")) <= 120 ) "+
		"WHERE tt.\"Restaurant\" IS NULL OR tt.\"tableNum\" IS NULL "+
		"ORDER BY 1, 2) avtabs  ON (rests.\"id\" = avtabs.\"Restaurant\") JOIN "+
		"(SELECT avtabs.\"Restaurant\", SUM(avtabs.\"SeatsNum\") avSeats FROM (SELECT tbs.\"Restaurant\",  tbs.\"Number\", tbs.\"SeatsNum\" FROM \"Tables\" tbs LEFT JOIN \"Timetable\" tt ON (tbs.\"Restaurant\" = tt.\"Restaurant\" and tbs.\"Number\" = tt.\"tableNum\" AND ABS(%v*60+%v - (tt.\"hour\"*60+tt.\"minutes\")) <= 120 ) "+
		"WHERE tt.\"Restaurant\" IS NULL OR tt.\"tableNum\" IS NULL "+
		"ORDER BY 1, 2) avtabs GROUP BY avtabs.\"Restaurant\") tab2 ON (rests.\"id\" = tab2.\"Restaurant\") WHERE tab2.avSeats > %v", h, m, h, m, persons))
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			rest := Restaurant{}
			err := rows.Scan(&rest.Id, &rest.Name, &rest.WaitTime, &rest.Bill, &rest.AvailableSeats)
			if err != nil {
				fmt.Println(err)
				continue
			}
			restaurants[rest.Id] = rest
		}
	}
	return restaurants
}

//Вспомогательная функция для прохода отображения по возрастанию ключа
func sortKeys(arr map[int]int) []int {
	var keys []int
	for k := range arr {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

//Вспомогательная функция для прохода отображения по убыванию ключа
func sortKeysDesc(arr map[int]int) []int {
	var keys []int
	for k := range arr {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys
}

//Функция, которая возвращает количество мест и их вместимость, которые нужно забронировать
func getResTables(arr map[int]int, reqSeats int) map[int]int {
	res := make(map[int]int)

	forwardKeys := sortKeys(arr)
	descKeys := sortKeysDesc(arr)
	for reqSeats > 0 {
		for _, k := range descKeys {
			if arr[k] == 0 {
				continue
			}
			sum := 0
			for _, j := range forwardKeys {
				if j == k {
					break
				}
				sum += j * arr[j]
			}
			if reqSeats > sum {
				reqSeats -= k
				arr[k]--
				res[k]++
				break
			}
		}
	}
	return res
}

//Вспомогательная функция, которая возвращает вместимость столов и их количество для заданного ресторана
func buildMapFromID(id, h, m int, db *sql.DB) map[int]int {
	res := make(map[int]int)
	rows, err := db.Query(fmt.Sprintf("SELECT tbs.\"SeatsNum\",  COUNT(tbs.\"Number\") FROM \"Tables\" tbs LEFT JOIN \"Timetable\" tt ON (tbs.\"Restaurant\" = tt.\"Restaurant\" and "+
		"tbs.\"Number\" = tt.\"tableNum\" and ABS(%v*60+%v - (tt.\"hour\"*60+tt.\"minutes\")) <= 120 ) "+
		"WHERE (tt.\"Restaurant\" IS NULL OR tt.\"tableNum\" IS NULL) AND tbs.\"Restaurant\" = %v "+
		"GROUP BY 1", h, m, id))
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			var x, y int
			err := rows.Scan(&x, &y)
			if err != nil {
				fmt.Println(err)
				continue
			}
			res[x] = y
		}
	}
	fmt.Printf("buildFromID %v\n", res)
	return res
}

//Функция бронирования мест
func bookTables(db *sql.DB, h, m, id, visitorId int, tabs map[int]int) {
	var restSeats []int
	//Поиск номеров столов для бронирования
	for k, v := range tabs {
		rows, err := db.Query(fmt.Sprintf("SELECT avtabs.\"Restaurant\", avtabs.\"Number\" FROM (SELECT tbs.\"Restaurant\",  tbs.\"Number\", tbs.\"SeatsNum\" FROM \"Tables\" tbs LEFT JOIN \"Timetable\" tt ON (tbs.\"Restaurant\" = tt.\"Restaurant\" and "+
			"tbs.\"Number\" = tt.\"tableNum\" and ABS(%v*60+%v - (tt.\"hour\"*60+tt.\"minutes\")) <= 120 ) "+
			"WHERE tt.\"Restaurant\" IS NULL OR tt.\"tableNum\" IS NULL "+
			"ORDER BY 1, 2) avtabs WHERE avtabs.\"Restaurant\" = %v AND avtabs.\"SeatsNum\" = %v LIMIT %v", h, m, id, k, v))
		if err != nil {
			fmt.Println(err)
		} else {
			for rows.Next() {
				var x, y int
				err := rows.Scan(&x, &y)
				if err != nil {
					fmt.Println(err)
					continue
				}
				restSeats = append(restSeats, y)
			}
		}
	}
	fmt.Println(restSeats)
	//Добавление записей в список брони
	for _, v := range restSeats {
		_, err := db.Query(fmt.Sprintf("INSERT INTO public.\"Timetable\"(hour, minutes, \"Restaurant\", \"tableNum\", \"Visitor\")	VALUES (%v, %v, %v, %v, %v);",
			h, m, id, v, visitorId))
		if err != nil {
			fmt.Println(err)
		}
	}
}

//Функция, возвращающая id посетителя. Если посетитель уже есть в базе, то возвращается его id
func addVisitor(db *sql.DB, name, phone string) int {
	var exist int
	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM \"Visitors\" WHERE \"Name\" = '%v' AND \"Phone\" = '8%v'",
		name, phone))
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			var x int
			err := rows.Scan(&x)
			if err != nil {
				fmt.Println(err)
				continue
			}
			exist = x
		}
	}
	if exist == 0 {
		_, err = db.Query(fmt.Sprintf("INSERT INTO public.\"Visitors\"(\"Name\", \"Phone\") VALUES ('%v', '8%v');",
			name, phone))
		if err != nil {
			fmt.Println(err)
		}
	}
	rows, err = db.Query(fmt.Sprintf("SELECT \"Id\" FROM \"Visitors\" WHERE \"Name\" = '%v' AND \"Phone\" = '8%v'",
		name, phone))
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			var x int
			err := rows.Scan(&x)
			if err != nil {
				fmt.Println(err)
				continue
			}
			exist = x
		}
	}
	return exist
}
