package controllers

import (
	"bytes"
	"net/http"
	"testing"
)

func TestValidarReporte(t *testing.T) {
	body := []byte(`{
		"tipo_plan_id": "61639b8c1634adf976ed4b4c" ,
		"vigencia": "35" ,
		"nombre" : "Ejemplo para Subdetalle V1" ,
		"unidad_id" : "8" ,
		"categoria" : "Evaluación"
	}
	`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/validacion", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestValidarReporte Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestValidarReporte Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestValidarReporte:", err.Error())
		t.Fail()
	}
}
func TestDesagregado(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/desagregado", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestDesagregado Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestDesagregado Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestDesagregado:", err.Error())
		t.Fail()
	}
}

func TestPlanAccionAnual(t *testing.T) {
	body := []byte(`{
		"unidad_id": "8" ,
		"tipo_plan_id": "61639b8c1634adf976ed4b4c" ,
		"estado_plan_id" : "6153355601c7a2365b2fb2a1" ,
		"vigencia" : "35"
	}
	`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual/Seguimiento%20PruebaSure", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestPlanAccionAnual Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestPlanAccionAnual Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestPlanAccionAnual:", err.Error())
		t.Fail()
	}
}
func TestPlanAccionAnualGeneral(t *testing.T) {
	body := []byte(`{
		"tipo_plan_id": "61639b8c1634adf976ed4b4c" ,
		"estado_plan_id" : "6153355601c7a2365b2fb2a1" ,
		"vigencia" : "35"
	}	
	`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual-general/Seguimiento%20PruebaSure", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestPlanAccionAnualGeneral Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestPlanAccionAnualGeneral Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestPlanAccionAnualGeneral:", err.Error())
		t.Fail()
	}
}

func TestNecesidades(t *testing.T) {
	body := []byte(`{
		"tipo_plan_id": "61639b8c1634adf976ed4b4c" ,
		"estado_plan_id" : "614d3aeb01c7a245952fabff" ,
		"vigencia" : "3"
	}	
	`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/necesidades/Plan%20de%20Acci%C3%B3n%20de%20Funcionamiento%202022", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestNecesidades Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestNecesidades Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestNecesidades:", err.Error())
		t.Fail()
	}
}

func TestPlanAccionEvaluacion(t *testing.T) {
	body := []byte(`{
		"unidad_id": "14" ,
		"tipo_plan_id": "61639b8c1634adf976ed4b4c" ,
		"vigencia" : "35"
	}
	
	`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual-evaluacion/Plan%20de%20acci%C3%B3n%202023%20Prod", "application/json", bytes.NewBuffer(body)); err == nil {
		if response.StatusCode != 200 {
			t.Error("Error TestPlanAccionEvaluacion Se esperaba 200 y se obtuvo", response.StatusCode)
			t.Fail()
		} else {
			t.Log("TestPlanAccionEvaluacion Finalizado Correctamente (OK)")
		}
	} else {
		t.Error("Error TestPlanAccionEvaluacion:", err.Error())
		t.Fail()
	}
}
