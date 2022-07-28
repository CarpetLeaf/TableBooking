package main

// Структура ресторана
type Restaurant struct {
	Id             int    //Id
	Name           string //Имя
	WaitTime       int    //Среднее время ожидания
	Bill           int    //Средний чек
	AvailableSeats int    //Доступно мест
}

func (r *Restaurant) setRestaurant(id int, name string, waitTime int, bill int, avS int) {
	r.Id = id
	r.Name = name
	r.WaitTime = waitTime
	r.Bill = bill
	r.AvailableSeats = avS
}
