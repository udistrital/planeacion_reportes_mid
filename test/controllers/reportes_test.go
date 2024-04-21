package controllers

import (
	"bytes"
	"net/http"
	"testing"
)

func TestValidarReporte(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/validar_reporte", "application/json", bytes.NewBuffer(body)); err == nil {
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

func TestPlanAccionAnual(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual/Seguimiento%20PruebaSure", "application/json", bytes.NewBuffer(body)); err == nil {
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
func TestPlanAccionAnualGeneral(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual-general/Seguimiento%20PruebaSure", "application/json", bytes.NewBuffer(body)); err == nil {
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

func TestNecesidades(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/necesidades/Plan%20de%20Acci%C3%B3n%20Funcionamiento%202022", "application/json", bytes.NewBuffer(body)); err == nil {
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

func TestPlanAccionEvaluacion(t *testing.T) {
	body := []byte(`{}`)

	if response, err := http.Post("http://localhost:8080/v1/reportes/plan-anual-evaluacion/Plan%20de%20Acci%C3%B3n%202023%20Prod", "application/json", bytes.NewBuffer(body)); err == nil {
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
