package main

type Table struct {
	Num      int  //Номер стола
	NumSeats int  //Доступно мест за столом
	Booked   bool //Флаг, обозначающий занят стол или нет
}

func (t *Table) setTable(num int, numSeats int, booked bool) {
	t.Num = num
	t.NumSeats = numSeats
	t.Booked = booked
}
