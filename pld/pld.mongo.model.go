package pld

type PLD struct {
	Ano        int64   `bson:"ano"`
	Mes        int64   `bson:"mes"`
	Dia        int64   `bson:"dia"`
	Hora       int64   `bson:"hora"`
	Submercado int64   `bson:"submercado"`
	Valor      float64 `bson:"valor"`
}
