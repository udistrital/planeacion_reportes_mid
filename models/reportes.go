package models

type Nodo struct {
	Valor     int
	Divisible bool
	Hijos     []*Nodo
}

type TotalDocentVal struct {
	SalarioBasico      int
	PrimaServicios     int
	PrimaNavidad       int
	PrimaVacaciones    int
	Bonificacion       int
	PensionesPublicas  int
	PensionesPrivadas  int
	Salud              int
	InteresesCesantias int
	CesantiasPublicas  int
	CesantiasPrivadas  int
	Caja               int
	Arl                int
	Icbf               int
}

type TotalesDocentes struct {
	Rhf     TotalDocentVal
	Rhv_pre TotalDocentVal
	Rhv_pos TotalDocentVal
}
