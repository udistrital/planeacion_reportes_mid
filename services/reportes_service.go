package services

import (
	"encoding/json"
	"errors"

	reporteshelper "github.com/udistrital/planeacion_reportes_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func ValidarReporte(data []byte) (interface{}, error) {
	//resultado solicitud

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + err.Error())
	}

	if data, err := reporteshelper.Validar(body); err == nil {
		return data, nil
	} else {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: ")
	}

}

func Desagregado(data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + err.Error())
	}

	if data, err := reporteshelper.ProcesarDesagregado(body); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al decodificar el cuerpo de la solicitud: ")
	}

}

func PlanAccionAnual(nombre string, data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("Error: 400 Not found " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("Error: 400 Not found " + err.Error())
	}

	if data, err := reporteshelper.ProcesarPlanAccionAnual(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al decodificar el cuerpo de la solicitud: ")
	}

}
func PlanAccionAnualGeneral(nombre string, data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("Error: 400 Not found " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("Error: 400 Not found " + err.Error())
	}

	if data, err := reporteshelper.ProcesarPlanAccionAnualGeneral(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al decodificar el cuerpo de la solicitud: ")
	}
}
func Necesidades(nombre string, data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("Error: 400 Not found " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("Error: 400 Not found " + err.Error())
	}

	if data, err := reporteshelper.ProcesarNecesidades(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al decodificar el cuerpo de la solicitud: ")
	}
}
func PlanAccionEvaluacion(nombre string, data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("Error: 400 Not found " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("Error: 400 Not found " + err.Error())
	}

	if data, err := reporteshelper.ProcesarPlanAccionEvaluacion(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al decodificar el cuerpo de la solicitud: ")
	}
}
