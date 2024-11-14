package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	evaluacionService "github.com/udistrital/planeacion_evaluacion_mid/services"
	reporteshelper "github.com/udistrital/planeacion_reportes_mid/helpers"
	seguimientoService "github.com/udistrital/planeacion_seguimiento_mid/services"
	"github.com/udistrital/utils_oas/formatdata"
	"github.com/udistrital/utils_oas/request"
	"github.com/xuri/excelize/v2"
)

const (
	ColorBlanco      string = "FFFFFF"
	ColorNegro       string = "000000"
	ColorRojo        string = "CC0000"
	ColorGrisClaro   string = "F2F2F2"
	ColorGrisOscuro  string = "C2C2C2"
	ColorGrisOscuro2 string = "D9D9D9"
	ColorRosado      string = "FCE4D6"
	CodigoAval       string = "A_SP"
)

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

func NombreRubroByCodigo(rubros []map[string]interface{}, codigo string) string {
	nombre := "No definido"
	if codigo == "No definido" {
		return nombre
	}
	for i := 0; i < len(rubros); i++ {
		if rubros[i]["Codigo"] == codigo {
			nombre = rubros[i]["Nombre"].(string)
			break
		}
	}
	return nombre
}

func codigoRubrosDocentes(rubros []map[string]interface{}, categoria string) string {
	var codigo string
	for i := 0; i < len(rubros); i++ {
		rubro := rubros[i]
		if rubro["categoria"] == categoria {
			if _, exist := rubro["rubro"]; exist {
				codigo = rubro["rubro"].(string)
			} else {
				codigo = ""
			}
			break
		}
	}
	if codigo == "" {
		codigo = "No definido"
	}
	return codigo
}

func ValidarReporte(data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("error al ValidarBody: " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + err.Error())
	}

	if data, err := Validar(body); err == nil {
		return data, nil
	} else {
		return nil, errors.New("error al Validar el cuerpo de la solicitud: " + err.Error())
	}

}

func Validar(body map[string]interface{}) (res map[string]interface{}, outputError error) {
	defer func() {
		if err := recover(); err != nil {
			outputError = errors.New("error Validar status : 500")
		}
	}()

	var res1 map[string]interface{}
	var resFilter []map[string]interface{}
	res = make(map[string]interface{})

	if body["categoria"].(string) == "Evaluacion" {
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/plan?query=activo:true,tipo_plan_id:"+body["tipo_plan_id"].(string)+",dependencia_id:"+body["unidad_id"].(string), &res1); err == nil {
			request.LimpiezaRespuestaRefactor(res1, &resFilter)

			if len(resFilter) == 0 {
				res["mensaje"] = "No existen planes para la unidad seleccionada"
				res["reporte"] = false
			} else {
				noPlan := true
				noVigencia := true
				noEstado := true
				for i := 0; i < len(resFilter); i++ {
					if resFilter[i]["nombre"] == body["nombre"].(string) {
						noPlan = false
						if resFilter[i]["vigencia"] == body["vigencia"].(string) {
							noVigencia = false
							if resFilter[i]["estado_plan_id"] == "6153355601c7a2365b2fb2a1" { //Estado Avalado
								noEstado = false
								res["mensaje"] = ""
								res["reporte"] = true
								break
							}
						}
					}
				}

				if noPlan {
					res["mensaje"] = "La unidad no tiene registros con el plan seleccionado o el tipo-plan no coincide con un Plan de Acción de Funcionamiento"
					res["reporte"] = false
				} else if noVigencia {
					res["mensaje"] = "La unidad no cuenta con registros para la vigencia y el plan selecionados"
					res["reporte"] = false
				} else if noEstado {
					res["mensaje"] = "La unidad no cuenta con plan avalado"
					res["reporte"] = false
				}
			}
		} else {
			res["mensaje"] = "Ocurrio un error"
		}
	} else if body["categoria"].(string) == "Necesidades" {
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/plan?query=activo:true,tipo_plan_id:"+body["tipo_plan_id"].(string)+",vigencia:"+body["vigencia"].(string), &res1); err == nil {
			request.LimpiezaRespuestaRefactor(res1, &resFilter)
			if len(resFilter) == 0 {
				res["mensaje"] = "No existen planes para la vigencia seleccionada"
				res["reporte"] = false
			} else {
				noPlan := true
				noEstado := true
				for i := 0; i < len(resFilter); i++ {
					if resFilter[i]["nombre"] == body["nombre"].(string) {
						noPlan = false
						if resFilter[i]["estado_plan_id"] == body["estado_plan_id"].(string) {
							noEstado = false
							res["mensaje"] = ""
							res["reporte"] = true
							break
						}
					}
				}

				if noPlan {
					res["mensaje"] = "No existen registros con el plan seleccionado o el tipo-plan no coincide con un Plan de Acción de Funcionamiento"
					res["reporte"] = false
				} else if noEstado {
					res["mensaje"] = "No existen registros con el estado y plan seleccionado"
					res["reporte"] = false
				}
			}
		} else {
			res["mensaje"] = "Ocurrio un error"
		}
	} else if body["categoria"].(string) == "Plan_Accion_Unidad" {
		// TODO: hacer validación de tipo de plan de acción para mostrar que los demás planes aún no están soportados por el sistema.
		// "Por favor verificar el tipo de plan de acción. Actualmente NO soportado por el módulo de reportes."
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/plan?query=activo:true,tipo_plan_id:"+body["tipo_plan_id"].(string)+",dependencia_id:"+body["unidad_id"].(string), &res1); err == nil {
			request.LimpiezaRespuestaRefactor(res1, &resFilter)
			if len(resFilter) == 0 {
				res["mensaje"] = "No existen planes para la unidad seleccionada"
				res["reporte"] = false
			} else {
				noPlan := true
				noVigencia := true
				noEstado := true
				for i := 0; i < len(resFilter); i++ {
					if resFilter[i]["nombre"] == body["nombre"].(string) {
						noPlan = false
						if resFilter[i]["vigencia"] == body["vigencia"].(string) {
							noVigencia = false
							if resFilter[i]["estado_plan_id"] == body["estado_plan_id"].(string) {
								noEstado = false
								res["mensaje"] = ""
								res["reporte"] = true
								break
							}
						}
					}
				}

				if noPlan {
					res["mensaje"] = "La unidad no tiene registros con el plan seleccionado o el tipo-plan no coincide con un Plan de Acción de Funcionamiento"
					res["reporte"] = false
				} else if noVigencia {
					res["mensaje"] = "La unidad no cuenta con registros para la vigencia y el plan selecionados"
					res["reporte"] = false
				} else if noEstado {
					res["mensaje"] = "La unidad no cuenta con plan en el estado solicitado"
					res["reporte"] = false
				}
			}
		} else {
			res["mensaje"] = "Ocurrió un error"
		}
	} else if body["categoria"].(string) == "Plan_Accion_General" {
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/plan?query=activo:true,tipo_plan_id:"+body["tipo_plan_id"].(string)+",vigencia:"+body["vigencia"].(string), &res1); err == nil {
			request.LimpiezaRespuestaRefactor(res1, &resFilter)
			if len(resFilter) == 0 {
				res["mensaje"] = "No existen planes para la vigencia seleccionada"
				res["reporte"] = false
			} else {
				noPlan := true
				noEstado := true
				for i := 0; i < len(resFilter); i++ {
					if resFilter[i]["nombre"] == body["nombre"].(string) {
						noPlan = false
						if resFilter[i]["estado_plan_id"] == body["estado_plan_id"].(string) {
							noEstado = false
							res["mensaje"] = ""
							res["reporte"] = true
							break
						}
					}
				}

				if noPlan {
					res["mensaje"] = "No existen registros con el plan seleccionado"
					res["reporte"] = false
				} else if noEstado {
					res["mensaje"] = "No existen registros con el estado y plan seleccionado"
					res["reporte"] = false
				}
			}
		} else {
			res["mensaje"] = "Ocurrio un error"
		}
	} else {
		res["mensaje"] = "Categoria incorrecta"
		res["reporte"] = false
	}

	return res, outputError
}

func getIdEstadoAval() (string, error) {
	var resEstado map[string]interface{}
	var estado []map[string]interface{}
	url := "http://" + beego.AppConfig.String("PlanesService") + "/estado-plan?query=activo:true,codigo_abreviacion:" + CodigoAval
	err := request.GetJson(url, &resEstado)
	if err != nil {
		return "", err
	}
	request.LimpiezaRespuestaRefactor(resEstado, &estado)
	return estado[0]["_id"].(string), nil
}

func Desagregado(data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(data); !v || e != nil {
		return nil, errors.New("error al ValidarBody: " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, errors.New("error al decodificar el cuerpo de la solicitud: " + err.Error())
	}

	if data, err := ProcesarDesagregado(body); err == nil {
		return data, nil
	} else {
		return nil, errors.New("Error al ProcesarDesagregado: " + err.Error())
	}

}

func ProcesarDesagregado(body map[string]interface{}) (dataRes interface{}, outputError error) {
	defer func() {
		if err := recover(); err != nil {
			outputError = errors.New("error ProcesarDesagregado status : 500")
		}
	}()

	var respuesta map[string]interface{}
	var planesFilter []map[string]interface{}
	var respuestaOikos []map[string]interface{}
	var nombreDep map[string]interface{}
	var identificacionres []map[string]interface{}
	var res map[string]interface{}
	var identificacion map[string]interface{}
	var dato map[string]interface{}
	var data_identi []map[string]interface{}
	var nombreUnidadVer string

	// excel
	var consolidadoExcel = excelize.NewFile()

	url := "http://" + beego.AppConfig.String("PlanesService") + "/plan?query=activo:true,tipo_plan_id:" + body["tipo_plan_id"].(string) + ",vigencia:" + body["vigencia"].(string) + ",estado_plan_id:" + body["estado_plan_id"].(string)
	if err := request.GetJson(url, &respuesta); err != nil {
		outputError = errors.New("error al procesar el cuerpo de la peticion" + err.Error())
	}

	request.LimpiezaRespuestaRefactor(respuesta, &planesFilter)
	for _, plan := range planesFilter {
		planId := plan["_id"].(string)
		dependencia := plan["dependencia_id"].(string)

		url2 := "http://" + beego.AppConfig.String("OikosService") + "/dependencia?query=Id:" + dependencia
		if err := request.GetJson(url2, &respuestaOikos); err != nil {
			outputError = errors.New("error al procesar el cuerpo de la peticion" + err.Error())
		}
		nombreDep = respuestaOikos[0]

		url3 := "http://" + beego.AppConfig.String("PlanesService") + "/identificacion?query=activo:true,plan_id:" + planId + ",tipo_identificacion_id:" + "617b6630f6fc97b776279afa"
		if err := request.GetJson(url3, &res); err != nil {
			outputError = errors.New("error al procesar el cuerpo de la peticion" + err.Error())
		}
		request.LimpiezaRespuestaRefactor(res, &identificacionres)
		identificacion = identificacionres[0]

		if identificacion["dato"] != "{}" {
			dato_str := identificacion["dato"].(string)
			json.Unmarshal([]byte(dato_str), &dato)
			for key := range dato {
				element := dato[key].(map[string]interface{})
				if element["activo"] == true {
					delete(element, "actividades")
					delete(element, "activo")
					delete(element, "index")
					element["unidad"] = nombreDep["Nombre"]
					data_identi = append(data_identi, element)
				}
			}
		} else {
			dataRes = ""
		}
	}

	stylehead, _ := reporteshelper.EstiloExcel(consolidadoExcel, "center", "center", ColorRojo, true)
	styletitles, _ := reporteshelper.EstiloExcel(consolidadoExcel, "center", "center", ColorGrisClaro, false)
	stylecontent, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcel, "center", "center", "", 0, false)

	contadorDesagregado := 3
	for h := 0; h < len(data_identi); h++ {
		datosArreglo := data_identi[h]
		nombreUnidadVerIn := datosArreglo["unidad"].(string)
		if h == 0 {
			nombreUnidadVer = nombreUnidadVerIn
		}

		if nombreUnidadVerIn != nombreUnidadVer {
			contadorDesagregado = 3
		}

		nombreUnidadVer = datosArreglo["unidad"].(string)
		nombreHoja := nombreUnidadVer
		sheetName := nombreHoja
		index, _ := consolidadoExcel.NewSheet(sheetName)
		consolidadoExcel.MergeCell(sheetName, "B1", "D1")

		consolidadoExcel.SetRowHeight(sheetName, 1, 20)
		consolidadoExcel.SetRowHeight(sheetName, 2, 20)
		consolidadoExcel.SetRowHeight(sheetName, contadorDesagregado, 50)

		consolidadoExcel.SetColWidth(sheetName, "A", "A", 30)
		consolidadoExcel.SetColWidth(sheetName, "B", "B", 50)
		consolidadoExcel.SetColWidth(sheetName, "C", "C", 30)
		consolidadoExcel.SetColWidth(sheetName, "D", "D", 60)

		consolidadoExcel.SetCellValue(sheetName, "A1", "Dependencia Responsable")
		consolidadoExcel.SetCellValue(sheetName, "B1", nombreUnidadVer)
		consolidadoExcel.SetCellValue(sheetName, "A2", "Código del rubro")
		consolidadoExcel.SetCellValue(sheetName, "B2", "Nombre del rubro")
		consolidadoExcel.SetCellValue(sheetName, "C2", "valor")
		consolidadoExcel.SetCellValue(sheetName, "D2", "Descripción del bien y/o servicio")
		consolidadoExcel.SetCellValue(sheetName, "A"+fmt.Sprint(contadorDesagregado), datosArreglo["codigo"])
		consolidadoExcel.SetCellValue(sheetName, "B"+fmt.Sprint(contadorDesagregado), datosArreglo["Nombre"])
		consolidadoExcel.SetCellValue(sheetName, "C"+fmt.Sprint(contadorDesagregado), datosArreglo["valor"])
		consolidadoExcel.SetCellValue(sheetName, "D"+fmt.Sprint(contadorDesagregado), datosArreglo["descripcion"])

		consolidadoExcel.SetCellStyle(sheetName, "A1", "A1", stylehead)
		consolidadoExcel.SetCellStyle(sheetName, "B1", "D1", stylehead)
		consolidadoExcel.SetCellStyle(sheetName, "A2", "D2", styletitles)
		consolidadoExcel.SetCellStyle(sheetName, "A"+fmt.Sprint(contadorDesagregado), "D"+fmt.Sprint(contadorDesagregado), stylecontent)
		consolidadoExcel.SetActiveSheet(index)

		if nombreUnidadVerIn == nombreUnidadVer {
			contadorDesagregado++
		}
	}

	if len(consolidadoExcel.GetSheetList()) > 1 {
		consolidadoExcel.DeleteSheet("Sheet1")
	}

	dataSend := make(map[string]interface{})

	buf, _ := consolidadoExcel.WriteToBuffer()
	strings.NewReader(buf.String())

	encoded := base64.StdEncoding.EncodeToString([]byte(buf.Bytes()))

	dataSend["generalData"] = data_identi
	dataSend["excelB64"] = encoded

	dataRes = dataSend
	return dataRes, outputError
}

func PlanAccionAnual(nombre string, Data []byte) (interface{}, error) {

	if v, e := request.ValidarBody(Data); !v || e != nil {
		return nil, errors.New("Error: 400 Not found " + e.Error())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(Data, &body); err != nil {
		return nil, errors.New("Error: 400 Not found " + err.Error())
	}

	if Data, err := ProcesarPlanAccionAnual(body, nombre); err == nil {
		return Data, nil
	} else {
		return nil, errors.New("error al PlanAccionAnual: " + err.Error())
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

	if data, err := ProcesarPlanAccionAnualGeneral(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("error al ProcesarPlanAccionAnualGeneral: " + err.Error())
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

	if data, err := ProcesarNecesidades(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("error al ProcesarNecesidades: " + err.Error())
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

	if data, err := ProcesarPlanAccionEvaluacion(body, nombre); err == nil {
		return data, nil
	} else {
		return nil, errors.New("error al ProcesarPlanAccionEvaluacion: " + err.Error())
	}
}

func construirTablas(consolidadoExcelPlanAnual *excelize.File, recursos []map[string]interface{}, contratistas []map[string]interface{}, docentes map[string]interface{}, rubro string, nombreRubro string, esReporteAntiguo bool) (*excelize.File, error) {
	stylecontent, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentMR, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentC, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentMRS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentCS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	styletitles, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 12, Color: "000000"},
		Border:    []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}}})
	stylehead, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true, Color: "ffffff"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"CC0000"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentTotal, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 6},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentTotalCant, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 6},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentTotalM, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 6},
			{Type: "bottom", Color: "000000", Style: 1}}})
	stylecontentRubro, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Font:      &excelize.Font{Bold: true},
		Border: []excelize.Border{{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1}}})

	sheetName := "Identificaciones"

	consolidadoExcelPlanAnual.NewSheet(sheetName)
	consolidadoExcelPlanAnual.InsertCols(sheetName, "A", 1)
	disable := false
	err := consolidadoExcelPlanAnual.SetSheetView(sheetName, -1, &excelize.ViewOptions{ShowGridLines: &disable})
	if err != nil {
		return nil, errors.New("error al construirTablas:" + err.Error())
	}
	consolidadoExcelPlanAnual.MergeCell(sheetName, "B1", "F1")

	consolidadoExcelPlanAnual.SetColWidth(sheetName, "A", "A", 2)
	consolidadoExcelPlanAnual.SetColWidth(sheetName, "B", "D", 26)
	consolidadoExcelPlanAnual.SetColWidth(sheetName, "C", "C", 30)
	consolidadoExcelPlanAnual.SetColWidth(sheetName, "E", "E", 35)
	consolidadoExcelPlanAnual.SetColWidth(sheetName, "F", "G", 20)
	consolidadoExcelPlanAnual.SetColWidth(sheetName, "H", "I", 20)
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "B1", "Identificación de recursos")
	consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B1", "F1", styletitles)
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "B3", "Código del rubro")
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "C3", "Nombre del rubro")
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "D3", "Valor")
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "E3", "Descripción del bien y/o servicio")
	consolidadoExcelPlanAnual.SetCellValue(sheetName, "F3", "Actividades")
	consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B3", "F3", stylehead)
	consolidadoExcelPlanAnual.SetRowHeight(sheetName, 2, 7)

	contador := 4
	for i := 0; i < len(recursos); i++ {
		aux := recursos[i]
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), aux["codigo"])
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), aux["Nombre"])
		auxValor, err := reporteshelper.DeformatNumberInt(aux["valor"])
		if err == nil {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), auxValor)
		}
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), aux["descripcion"])
		auxStrString := aux["actividades"].([]interface{})
		var strActividades string
		for j := 0; j < len(auxStrString); j++ {
			if j != len(auxStrString)-1 {
				strActividades = strActividades + " " + auxStrString[j].(string) + ","
			} else {
				strActividades = strActividades + " " + auxStrString[j].(string)
			}
		}
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), strActividades)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "B"+fmt.Sprint(contador), "F"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "D"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
	}
	contador++
	if esReporteAntiguo {
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "H"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Identificación de contratistas")
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contador+1, 7)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), styletitles)

		contador++
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Descripción de la necesidad")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), "Perfil")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "Cantidad")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), "Valor Total")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), "Valor Total Incremento")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contador), "Actividades")
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), stylehead)

		contador++
		var total float64 = 0
		var valorTotal int = 0
		var valorTotalInc int = 0
		for i := 0; i < len(contratistas); i++ {
			var respuestaParametro map[string]interface{}
			var perfil map[string]interface{}

			aux := contratistas[i]

			total = total + aux["cantidad"].(float64)
			aux1 := fmt.Sprintf("%v", aux["valorTotal"])
			strValorTotal := strings.TrimLeft(aux1, "$")
			strValorTotal = strings.ReplaceAll(strValorTotal, ",", "")
			arrValorTotal := strings.Split(strValorTotal, ".")
			auxValorTotal, err := strconv.Atoi(arrValorTotal[0])

			if err == nil {
				valorTotal = valorTotal + auxValorTotal
			}
			consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), aux["descripcionNecesidad"])
			if err := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+"/parametro/"+fmt.Sprint(aux["perfil"]), &respuestaParametro); err == nil {
				request.LimpiezaRespuestaRefactor(respuestaParametro, &perfil)
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), perfil["Nombre"])
			}
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), aux["cantidad"])
			aux2 := fmt.Sprintf("%v", aux["valorTotal"])
			strValor := strings.TrimLeft(aux2, "$")
			strValor = strings.ReplaceAll(strValor, ",", "")
			arrValor := strings.Split(strValor, ".")
			auxValor, err := strconv.Atoi(arrValor[0])
			if err == nil {
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), auxValor)
			}
			aux3 := fmt.Sprintf("%v", aux["valorTotalInc"])
			strValorInc := strings.TrimLeft(aux3, "$")
			strValorInc = strings.ReplaceAll(strValorInc, ",", "")
			arrValorInc := strings.Split(strValorInc, ".")
			auxValorInc, err := strconv.Atoi(arrValorInc[0])
			if err == nil {
				valorTotalInc = valorTotalInc + auxValorInc
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), auxValorInc)
			}
			auxStrString := aux["actividades"].([]interface{})
			var strActividades string
			for j := 0; j < len(auxStrString); j++ {
				if j != len(auxStrString)-1 {
					strActividades = strActividades + " " + auxStrString[j].(string) + ","
				} else {
					strActividades = strActividades + " " + auxStrString[j].(string)
				}
			}
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contador), strActividades)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "B"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), stylecontent, stylecontentS)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "F"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentC, stylecontentCS)
			contador++
		}
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Total")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), total)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), valorTotal)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), valorTotalInc)

		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentTotal)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentTotalCant)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "F"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentTotalM)

		contador++
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Rubro")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), rubro)
		consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(contador), "G"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), nombreRubro)

		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador), stylecontentRubro)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "D"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentC)

		contador++
		contador++
	} else {
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Identificación de contratistas")
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador), styletitles)

		contador++
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Descripción de la necesidad (objeto contractual)")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), "Requisitos mínimos (formación académica y experiencia)")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "Perfil")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), "Cantidad")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), "Valor Total")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contador), "Valor Total Incremento")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(contador), "Actividades")
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador), stylehead)

		contador++
		var total float64 = 0
		var valorTotal int = 0
		var valorTotalInc int = 0
		for i := 0; i < len(contratistas); i++ {
			var respuestaParametro map[string]interface{}
			var perfil map[string]interface{}

			aux := contratistas[i]

			total = total + aux["cantidad"].(float64)
			aux1 := fmt.Sprintf("%v", aux["valorTotal"])
			strValorTotal := strings.TrimLeft(aux1, "$")
			strValorTotal = strings.ReplaceAll(strValorTotal, ",", "")
			arrValorTotal := strings.Split(strValorTotal, ".")
			auxValorTotal, err := strconv.Atoi(arrValorTotal[0])

			if err == nil {
				valorTotal = valorTotal + auxValorTotal
			}
			consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), aux["descripcionNecesidad"])
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), aux["requisitos"])
			if err := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+"/parametro/"+fmt.Sprint(aux["perfil"]), &respuestaParametro); err == nil {
				request.LimpiezaRespuestaRefactor(respuestaParametro, &perfil)
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), perfil["Nombre"])
			}
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), aux["cantidad"])
			aux2 := fmt.Sprintf("%v", aux["valorTotal"])
			strValor := strings.TrimLeft(aux2, "$")
			strValor = strings.ReplaceAll(strValor, ",", "")
			arrValor := strings.Split(strValor, ".")
			auxValor, err := strconv.Atoi(arrValor[0])
			if err == nil {
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), auxValor)
			}
			aux3 := fmt.Sprintf("%v", aux["valorTotalInc"])
			strValorInc := strings.TrimLeft(aux3, "$")
			strValorInc = strings.ReplaceAll(strValorInc, ",", "")
			arrValorInc := strings.Split(strValorInc, ".")
			auxValorInc, err := strconv.Atoi(arrValorInc[0])
			if err == nil {
				valorTotalInc = valorTotalInc + auxValorInc
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contador), auxValorInc)
			}
			auxStrString := aux["actividades"].([]interface{})
			var strActividades string
			for j := 0; j < len(auxStrString); j++ {
				if j != len(auxStrString)-1 {
					strActividades = strActividades + " " + auxStrString[j].(string) + ","
				} else {
					strActividades = strActividades + " " + auxStrString[j].(string)
				}
			}
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(contador), strActividades)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador), stylecontent, stylecontentS)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "G"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, i, sheetName, "F"+fmt.Sprint(contador), "F"+fmt.Sprint(contador), stylecontentC, stylecontentCS)
			contador++
		}
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Total")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contador), total)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contador), valorTotal)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contador), valorTotalInc)

		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentTotal)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "F"+fmt.Sprint(contador), "F"+fmt.Sprint(contador), stylecontentTotalCant)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "G"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), stylecontentTotalM)

		contador++
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "C"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Rubro")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), rubro)
		consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(contador), "H"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), nombreRubro)

		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontentRubro)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "E"+fmt.Sprint(contador), "H"+fmt.Sprint(contador), stylecontentC)

		contador++
		contador++
	}

	if docentes != nil {
		infoDocentes := reporteshelper.TotalDocentes(docentes)["TotalesPorTipo"].(TotalesDocentes)
		rubros := docentes["rubros"].([]map[string]interface{})
		rubros_pos := docentes["rubros_pos"].([]map[string]interface{})
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Identificación docente")
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contador+1, 7)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), styletitles)

		contador++
		contador++

		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Categoría")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), "Código del rubro")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), "Nombre del rubro")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "Valor")
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylehead)

		contador++

		//Cuerpo Tabla
		content, _ := ioutil.ReadFile("static/json/rubros.json")
		rubrosJson := []map[string]interface{}{}
		_ = json.Unmarshal(content, &rubrosJson)

		code := ""
		nombre := ""

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Salario básico")
		code = codigoRubrosDocentes(rubros, "Salario básico")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.SalarioBasico+infoDocentes.Rhv_pre.SalarioBasico)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Salario básico")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.SalarioBasico <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.SalarioBasico)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Prima de Servicios")
		code = codigoRubrosDocentes(rubros, "Prima de Servicios")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.PrimaServicios+infoDocentes.Rhv_pre.PrimaServicios)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Prima de Servicios")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.PrimaServicios <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.PrimaServicios)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Prima de navidad")
		code = codigoRubrosDocentes(rubros, "Prima de navidad")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.PrimaNavidad+infoDocentes.Rhv_pre.PrimaNavidad)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Prima de navidad")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.PrimaNavidad <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.PrimaNavidad)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Prima de vacaciones")
		code = codigoRubrosDocentes(rubros, "Prima de vacaciones")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.PrimaVacaciones+infoDocentes.Rhv_pre.PrimaVacaciones)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Prima de vacaciones")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.PrimaVacaciones <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.PrimaVacaciones)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Fondo pensiones público")
		code = codigoRubrosDocentes(rubros, "Fondo pensiones público")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.PensionesPublicas+infoDocentes.Rhv_pre.PensionesPublicas)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Fondo pensiones público")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.PensionesPublicas <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.PensionesPublicas)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Fondo pensiones privado")
		code = codigoRubrosDocentes(rubros, "Fondo pensiones privado")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.PensionesPrivadas+infoDocentes.Rhv_pre.PensionesPrivadas+infoDocentes.Rhv_pos.PensionesPrivadas)
		//consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), infoDocentes.Rhf.PensionesPrivadas+infoDocentes.Rhv_pre.PensionesPrivadas)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Fondo pensiones privado")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		//consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), infoDocentes.Rhv_pos.PensionesPrivadas)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte salud")
		code = codigoRubrosDocentes(rubros, "Aporte salud")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.Salud+infoDocentes.Rhv_pre.Salud)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte salud")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.Salud <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.Salud)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte cesantías público")
		code = codigoRubrosDocentes(rubros, "Aporte cesantías público")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.CesantiasPublicas+infoDocentes.Rhv_pre.CesantiasPublicas)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte cesantías público")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.CesantiasPublicas <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.CesantiasPublicas)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte cesantías privado")
		code = codigoRubrosDocentes(rubros, "Aporte cesantías privado")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.CesantiasPrivadas+infoDocentes.Rhv_pre.CesantiasPrivadas)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte cesantías privado")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.CesantiasPrivadas <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.CesantiasPrivadas)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte CCF")
		code = codigoRubrosDocentes(rubros, "Aporte CCF")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.Caja+infoDocentes.Rhv_pre.Caja)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte CCF")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.Caja <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.Caja)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte ARL")
		code = codigoRubrosDocentes(rubros, "Aporte ARL")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.Arl+infoDocentes.Rhv_pre.Arl)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte ARL")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.Arl <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.Arl)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contador), "Aporte ICBF")
		code = codigoRubrosDocentes(rubros, "Aporte ICBF")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhf.Icbf+infoDocentes.Rhv_pre.Icbf)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++
		code = codigoRubrosDocentes(rubros_pos, "Aporte ICBF")
		nombre = NombreRubroByCodigo(rubrosJson, code)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contador), code)
		if code == "No definido" && infoDocentes.Rhv_pos.Icbf <= 0 {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre+" Posgrado")
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), "N/A")
		} else {
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contador), nombre)
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contador), infoDocentes.Rhv_pos.Icbf)
		}
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "B"+fmt.Sprint(contador), "D"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, contador, sheetName, "E"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontentMR, stylecontentMRS)
		contador++

	}

	consolidadoExcelPlanAnual.InsertRows(sheetName, 1, 7)
	consolidadoExcelPlanAnual.MergeCell(sheetName, "C2", "G6")
	return consolidadoExcelPlanAnual, nil
}

func ProcesarPlanAccionAnual(body map[string]interface{}, nombre string) (map[string]interface{}, error) {
	var respuesta map[string]interface{}
	var planesFilter []map[string]interface{}
	var res map[string]interface{}
	var respuestaUnidad []map[string]interface{}
	var subgrupos []map[string]interface{}
	var plan_id string
	var actividadName string
	var arregloPlanAnual []map[string]interface{}
	var nombreUnidad string
	var resPeriodo map[string]interface{}
	var periodo []map[string]interface{}
	var unidadNombre string
	consolidadoExcelPlanAnual := excelize.NewFile()

	//? Estilos
	styletitle, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "000000"},
		Border: []excelize.Border{
			{Type: "right", Color: "ffffff", Style: 1},
			{Type: "left", Color: "ffffff", Style: 1},
			{Type: "top", Color: "ffffff", Style: 1},
			{Type: "bottom", Color: "ffffff", Style: 1},
		},
	})

	stylehead, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"CC0000"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styletitles, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontent, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontentS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontentC, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontentCL, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 4},
		},
	})
	stylecontentCLD, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontentCS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	stylecontentCLS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 4},
		},
	})
	stylecontentCLDS, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styleLineamiento, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal:   "center",
			Vertical:     "center",
			WrapText:     true,
			TextRotation: 90,
		},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styleLineamientoSombra, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal:   "center",
			Vertical:     "center",
			WrapText:     true,
			TextRotation: 90,
		},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	if body["unidad_id"].(string) != "" {
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/plan?query=activo:true,tipo_plan_id:"+body["tipo_plan_id"].(string)+",vigencia:"+body["vigencia"].(string)+",estado_plan_id:"+body["estado_plan_id"].(string)+",dependencia_id:"+body["unidad_id"].(string)+",nombre:"+nombre+"&sortby=fecha_modificacion&order=asc", &respuesta); err == nil {
			request.LimpiezaRespuestaRefactor(respuesta, &planesFilter)
			if err := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+`/periodo?query=Id:`+body["vigencia"].(string), &resPeriodo); err == nil {
				request.LimpiezaRespuestaRefactor(resPeriodo, &periodo)
			}

			anioPlan := periodo[0]["Year"].(float64)
			var esReporteAntiguo bool = true

			if anioPlan > 2024 { //? Vigencias 2025 en adelante son reportes con nueva estructura
				esReporteAntiguo = false
			}

			if len(planesFilter) <= 0 {
				return nil, errors.New("error: no se encontraron planes con los parametros ingresados")
			}

			for planes := 0; planes < len(planesFilter); planes++ {
				planesFilterData := planesFilter[planes]
				plan_id = planesFilterData["_id"].(string)
				if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo?query=padre:"+plan_id+"&fields=nombre,_id,hijos,activo", &res); err != nil {
					panic(err)
				}
				request.LimpiezaRespuestaRefactor(res, &subgrupos)
				if esReporteAntiguo {
					for i := 0; i < len(subgrupos); i++ {
						if strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "actividad") && strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "general") {
							actividades, _ := reporteshelper.GetActividades(subgrupos[i]["_id"].(string))
							var arregloLineamieto []map[string]interface{}
							var arregloLineamietoPI []map[string]interface{}
							if len(actividades) == 1 {
								for index := range actividades {
									if val, ok := actividades[index]["index"].(float64); ok {
										actividades[index]["index"] = fmt.Sprintf("%v", int(val))
									}
								}
							} else {
								sort.SliceStable(actividades, func(i int, j int) bool {
									if _, ok := actividades[i]["index"].(float64); ok {
										actividades[i]["index"] = fmt.Sprintf("%v", int(actividades[i]["index"].(float64)))
									}
									if _, ok := actividades[j]["index"].(float64); ok {
										actividades[j]["index"] = fmt.Sprintf("%v", int(actividades[j]["index"].(float64)))
									}
									aux, _ := strconv.Atoi((actividades[i]["index"]).(string))
									aux1, _ := strconv.Atoi((actividades[j]["index"]).(string))
									return aux < aux1
								})
							}

							reporteshelper.LimpiarDetalles()
							for j := 0; j < len(actividades); j++ {
								arregloLineamieto = nil
								arregloLineamietoPI = nil
								actividad := actividades[j]
								actividadName = actividad["dato"].(string)
								index := actividad["index"].(string)
								datosArmonizacion := make(map[string]interface{})
								titulosArmonizacion := make(map[string]interface{})

								reporteshelper.Limpia()
								tree := reporteshelper.BuildTreeFa(subgrupos, index)
								treeDatos := tree[0]
								treeDatas := tree[1]
								treeArmo := tree[2]
								armonizacionTercer := treeArmo[0]
								var armonizacionTercerNivel interface{}
								var armonizacionTercerNivelPI interface{}

								if armonizacionTercer["armo"] != nil {
									armonizacionTercerNivel = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPED"]
									armonizacionTercerNivelPI = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPI"]
								}

								for datoGeneral := 0; datoGeneral < len(treeDatos); datoGeneral++ {
									treeDato := treeDatos[datoGeneral]
									treeData := treeDatas[0]
									if treeDato["sub"] == "" {
										nombre := strings.ToLower(treeDato["nombre"].(string))
										if strings.Contains(nombre, "ponderación") || strings.Contains(nombre, "ponderacion") && strings.Contains(nombre, "actividad") {
											datosArmonizacion["Ponderación de la actividad"] = treeData[treeDato["id"].((string))]
										} else if strings.Contains(nombre, "período") || strings.Contains(nombre, "periodo") && strings.Contains(nombre, "ejecucion") || strings.Contains(nombre, "ejecución") {
											datosArmonizacion["Periodo de ejecución"] = treeData[treeDato["id"].(string)]
										} else if strings.Contains(nombre, "actividad") && strings.Contains(nombre, "general") {
											datosArmonizacion["Actividad general"] = treeData[treeDato["id"].(string)]
										} else if strings.Contains(nombre, "tarea") || strings.Contains(nombre, "actividades específicas") {
											datosArmonizacion["Tareas"] = treeData[treeDato["id"].(string)]
										} else {
											datosArmonizacion[treeDato["nombre"].(string)] = treeData[treeDato["id"].(string)]
										}
									}
								}
								var treeIndicador map[string]interface{}
								auxTree := tree[0]
								for i := 0; i < len(auxTree); i++ {
									subgrupo := auxTree[i]
									if strings.Contains(strings.ToLower(subgrupo["nombre"].(string)), "indicador") {
										treeIndicador = auxTree[i]
									}
								}

								subIndicador := treeIndicador["sub"].([]map[string]interface{})
								for ind := 0; ind < len(subIndicador); ind++ {
									subIndicadorRes := subIndicador[ind]
									treeData := treeDatas[0]
									dataIndicador := make(map[string]interface{})
									auxSubIndicador := subIndicadorRes["sub"].([]map[string]interface{})
									for subInd := 0; subInd < len(auxSubIndicador); subInd++ {
										if treeData[auxSubIndicador[subInd]["id"].(string)] == nil {
											treeData[auxSubIndicador[subInd]["id"].(string)] = ""
										}
										dataIndicador[auxSubIndicador[subInd]["nombre"].(string)] = treeData[auxSubIndicador[subInd]["id"].(string)]
									}
									titulosArmonizacion[subIndicadorRes["nombre"].(string)] = dataIndicador
								}

								datosArmonizacion["indicadores"] = titulosArmonizacion
								arregloLineamieto, _ = ArbolArmonizacionV2(armonizacionTercerNivel.(string))
								arregloLineamietoPI, _ = ArbolArmonizacionPIV2(armonizacionTercerNivelPI.(string))
								generalData := make(map[string]interface{})
								if err := request.GetJson("http://"+beego.AppConfig.String("OikosService")+"/dependencia_tipo_dependencia?query=DependenciaId:"+body["unidad_id"].(string), &respuestaUnidad); err == nil {
									aux := respuestaUnidad[0]
									dependenciaNombre := aux["DependenciaId"].(map[string]interface{})
									nombreUnidad = dependenciaNombre["Nombre"].(string)
								} else {
									panic(map[string]interface{}{"funcion": "GetUnidades", "err": "Error ", "status": "400", "log": err})
								}

								generalData["nombreUnidad"] = nombreUnidad
								generalData["nombreActividad"] = actividadName
								generalData["numeroActividad"] = index
								generalData["datosArmonizacion"] = arregloLineamieto
								generalData["datosArmonizacionPI"] = arregloLineamietoPI
								generalData["datosComplementarios"] = datosArmonizacion

								arregloPlanAnual = append(arregloPlanAnual, generalData)
							}
							break
						}
					}
					unidadNombre = arregloPlanAnual[0]["nombreUnidad"].(string)
					sheetName := "Actividades del plan"
					indexPlan, _ := consolidadoExcelPlanAnual.NewSheet(sheetName)

					if planes == 0 {
						consolidadoExcelPlanAnual.DeleteSheet("Sheet1")

						disable := false
						if err := consolidadoExcelPlanAnual.SetSheetView(sheetName, -1, &excelize.ViewOptions{
							ShowGridLines: &disable,
						}); err != nil {
							logs.Error(err)
						}
					}
					consolidadoExcelPlanAnual.MergeCell(sheetName, "B1", "D1")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "E1", "G1")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "H1", "H2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "I1", "I2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "J1", "J2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "K1", "K2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "L1", "L2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "Q1", "Q2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "M1", "P1")
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "B", "B", 18)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "C", "C", 11)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "D", "D", 35)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "E", "E", 16)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "F", "G", 35)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "H", "H", 6)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "I", "J", 12)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "K", "K", 30)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "L", "L", 35)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "M", "N", 52)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "O", "O", 30)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "P", "P", 10)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "Q", "Q", 30)
					consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B1", "Q1", stylehead)
					consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B2", "Q2", styletitles)

					// encabezado excel
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "B1", "Armonización PED")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "B2", "Lineamiento")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "C2", "Meta")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "D2", "Estrategias")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "E1", "Armonización Plan Indicativo")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "E2", "Ejes transformadores")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "F2", "Lineamientos de acción")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "G2", "Estrategias")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "H2", "N°.")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "I2", "Ponderación de la actividad")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "J2", "Periodo de ejecución")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "K2", "Actividad")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "L2", "Actividades específicas")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "M1", "Indicador")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "M2", "Nombre")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "N2", "Fórmula")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "O2", "Criterio del indicador")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "P2", "Meta")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q2", "Producto esperado")

					rowPos := 3

					for excelPlan := 0; excelPlan < len(arregloPlanAnual); excelPlan++ {

						datosExcelPlan := arregloPlanAnual[excelPlan]
						armoPED := datosExcelPlan["datosArmonizacion"].([]map[string]interface{})
						armoPI := datosExcelPlan["datosArmonizacionPI"].([]map[string]interface{})
						datosComplementarios := datosExcelPlan["datosComplementarios"].(map[string]interface{})
						indicadores := datosComplementarios["indicadores"].(map[string]interface{})

						MaxRowsXActivity := reporteshelper.MinComMul_Armonization(armoPED, armoPI, len(indicadores))

						y_lin := rowPos
						h_lin := MaxRowsXActivity / len(armoPED)

						for _, lin := range armoPED {
							consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1))
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1), styleLineamiento, styleLineamientoSombra)
							y_met := y_lin
							h_met := h_lin / len(lin["meta"].([]map[string]interface{}))
							for _, met := range lin["meta"].([]map[string]interface{}) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1))
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(y_met), met["nombreMeta"])
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1), stylecontentC, stylecontentCS)
								y_est := y_met
								h_est := h_met / len(met["estrategias"].([]map[string]interface{}))
								for _, est := range met["estrategias"].([]map[string]interface{}) {
									consolidadoExcelPlanAnual.MergeCell(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1))
									consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(y_est), est["descripcionEstrategia"])
									if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
									} else {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
									}
									y_est += h_est
								}
								y_met += h_met
							}
							y_lin += h_lin
						}

						y_eje := rowPos
						h_eje := MaxRowsXActivity / len(armoPI)
						for _, eje := range armoPI {
							consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1))
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(y_eje), eje["nombreFactor"])
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1), stylecontentC, stylecontentCS)
							y_lin := y_eje
							h_lin := h_eje / len(eje["lineamientos"].([]map[string]interface{}))
							for _, lin := range eje["lineamientos"].([]map[string]interface{}) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1))
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1), stylecontentC, stylecontentCS)
								y_est := y_lin
								h_est := h_lin / len(lin["estrategias"].([]map[string]interface{}))
								for _, est := range lin["estrategias"].([]map[string]interface{}) {
									consolidadoExcelPlanAnual.MergeCell(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1))
									consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(y_est), est["descripcionEstrategia"])
									if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
									} else {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
									}
									y_est += h_est
								}
								y_lin += h_lin
							}
							y_eje += h_eje
						}

						consolidadoExcelPlanAnual.MergeCell(sheetName, "H"+fmt.Sprint(rowPos), "H"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "I"+fmt.Sprint(rowPos), "I"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "J"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "K"+fmt.Sprint(rowPos), "K"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "L"+fmt.Sprint(rowPos), "L"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(rowPos), excelPlan+1)
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(rowPos), datosComplementarios["Ponderación de la actividad"])
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "J"+fmt.Sprint(rowPos), datosComplementarios["Periodo de ejecución"])
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "K"+fmt.Sprint(rowPos), datosComplementarios["Actividad general"])
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "L"+fmt.Sprint(rowPos), datosComplementarios["Tareas"])
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "H"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "K"+fmt.Sprint(rowPos), "L"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontent, stylecontentS)

						y_ind := rowPos
						h_ind := MaxRowsXActivity / len(indicadores)
						idx := int(0)
						var indicadoresVacios []map[string]interface{}
						var indicadoresNoVacios []map[string]interface{}
						var indicadoresOrdenados []map[string]interface{}

						for _, indicador := range indicadores {
							auxIndicador := indicador
							var nombreIndicador interface{}
							var formula interface{}
							var criterio interface{}
							var meta interface{}
							for key, element := range auxIndicador.(map[string]interface{}) {
								if strings.Contains(strings.ToLower(key), "nombre") {
									nombreIndicador = element
								}
								if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
									formula = element
								}
								if strings.Contains(strings.ToLower(key), "criterio") {
									criterio = element
								}
								if strings.Contains(strings.ToLower(key), "meta") {
									meta = element
								}
							}
							if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
								indicadoresVacios = append(indicadoresVacios, auxIndicador.(map[string]interface{}))
							} else {
								indicadoresNoVacios = append(indicadoresNoVacios, auxIndicador.(map[string]interface{}))
							}
						}

						indicadoresOrdenados = append(indicadoresNoVacios, indicadoresVacios...)
						for _, indicador := range indicadoresOrdenados {
							auxIndicador := indicador
							var nombreIndicador interface{}
							var formula interface{}
							var criterio interface{}
							var meta interface{}
							for key, element := range auxIndicador {
								if strings.Contains(strings.ToLower(key), "nombre") {
									nombreIndicador = element
								}
								if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
									formula = element
								}
								if strings.Contains(strings.ToLower(key), "criterio") {
									criterio = element
								}
								if strings.Contains(strings.ToLower(key), "meta") {
									meta = element
								}
							}
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "M"+fmt.Sprint(y_ind), nombreIndicador)
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "N"+fmt.Sprint(y_ind), formula)
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "O"+fmt.Sprint(y_ind), criterio)
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "P"+fmt.Sprint(y_ind), meta)
							if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind-1), "M"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind-1), "N"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind-1), "O"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "P"+fmt.Sprint(y_ind-1), "P"+fmt.Sprint(y_ind+h_ind-1))
							} else {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind), "M"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind), "N"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind), "O"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "P"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1))
							}

							idx++
							if idx < len(indicadores) {
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "M"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1), stylecontentCL, stylecontentCLS)
							} else {
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "M"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1), stylecontentCLD, stylecontentCLDS)
							}
							y_ind += h_ind
						}

						consolidadoExcelPlanAnual.MergeCell(sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q"+fmt.Sprint(rowPos), datosComplementarios["Producto esperado"])
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)

						rowPos += MaxRowsXActivity

						consolidadoExcelPlanAnual.SetActiveSheet(indexPlan)
					}
				} else { //? Es reporte nuevo
					for i := 0; i < len(subgrupos); i++ {
						if strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "actividad") {
							actividades, _ := reporteshelper.GetActividades(subgrupos[i]["_id"].(string))
							var arregloLineamieto []map[string]interface{}
							var arregloLineamietoPI []map[string]interface{}
							if len(actividades) == 1 {
								for index := range actividades {
									if val, ok := actividades[index]["index"].(float64); ok {
										actividades[index]["index"] = fmt.Sprintf("%v", int(val))
									}
								}
							} else {
								sort.SliceStable(actividades, func(i int, j int) bool {
									if _, ok := actividades[i]["index"].(float64); ok {
										actividades[i]["index"] = fmt.Sprintf("%v", int(actividades[i]["index"].(float64)))
									}
									if _, ok := actividades[j]["index"].(float64); ok {
										actividades[j]["index"] = fmt.Sprintf("%v", int(actividades[j]["index"].(float64)))
									}
									aux, _ := strconv.Atoi((actividades[i]["index"]).(string))
									aux1, _ := strconv.Atoi((actividades[j]["index"]).(string))
									return aux < aux1
								})
							}

							reporteshelper.LimpiarDetalles()
							for j := 0; j < len(actividades); j++ {
								arregloLineamieto = nil
								arregloLineamietoPI = nil
								actividad := actividades[j]
								actividadName = actividad["dato"].(string)
								index := actividad["index"].(string)
								datosArmonizacion := make(map[string]interface{})
								titulosArmonizacion := make(map[string]interface{})

								reporteshelper.Limpia()
								tree := reporteshelper.BuildTreeFa(subgrupos, index)
								treeDatos := tree[0]
								treeDatas := tree[1]
								treeArmo := tree[2]
								armonizacionTercer := treeArmo[0]
								var armonizacionTercerNivel interface{}
								var armonizacionTercerNivelPI interface{}

								if armonizacionTercer["armo"] != nil {
									armonizacionTercerNivel = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPED"]
									armonizacionTercerNivelPI = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPI"]
								}

								for datoGeneral := 0; datoGeneral < len(treeDatos); datoGeneral++ {
									treeDato := treeDatos[datoGeneral]
									treeData := treeDatas[0]
									if treeDato["sub"] == "" {
										nombre := strings.ToLower(treeDato["nombre"].(string))
										if strings.Contains(nombre, "ponderación") || strings.Contains(nombre, "ponderacion") && strings.Contains(nombre, "actividad") {
											datosArmonizacion["Ponderación de la actividad"] = treeData[treeDato["id"].((string))]
										} else if strings.Contains(nombre, "período") || strings.Contains(nombre, "periodo") && strings.Contains(nombre, "ejecucion") || strings.Contains(nombre, "ejecución") {
											datosArmonizacion["Periodo de ejecución"] = treeData[treeDato["id"].(string)]
										} else if strings.Contains(nombre, "actividad") {
											datosArmonizacion["Actividad general"] = treeData[treeDato["id"].(string)]
										} else if strings.Contains(nombre, "unidad") || strings.Contains(nombre, "grupo") {
											datosArmonizacion["Responsable"] = treeData[treeDato["id"].(string)]
										} else {
											datosArmonizacion[treeDato["nombre"].(string)] = treeData[treeDato["id"].(string)]
										}
									}
								}
								var treeIndicador map[string]interface{}
								auxTree := tree[0]
								for i := 0; i < len(auxTree); i++ {
									subgrupo := auxTree[i]
									if strings.Contains(strings.ToLower(subgrupo["nombre"].(string)), "indicador") {
										treeIndicador = auxTree[i]
									}
								}

								subIndicador := treeIndicador["sub"].([]map[string]interface{})
								// for ind := 0; ind < len(subIndicador); ind++ {
								// 	subIndicadorRes := subIndicador[ind]
								// 	treeData := treeDatas[0]
								// 	dataIndicador := make(map[string]interface{})
								// 	auxSubIndicador := subIndicadorRes["sub"].([]map[string]interface{})
								// 	for subInd := 0; subInd < len(auxSubIndicador); subInd++ {
								// 		if treeData[auxSubIndicador[subInd]["id"].(string)] == nil {
								// 			treeData[auxSubIndicador[subInd]["id"].(string)] = ""
								// 		}
								// 		dataIndicador[auxSubIndicador[subInd]["nombre"].(string)] = treeData[auxSubIndicador[subInd]["id"].(string)]
								// 	}
								// 	titulosArmonizacion[subIndicadorRes["nombre"].(string)] = dataIndicador
								// }
								for ind := 0; ind < len(subIndicador); ind++ {
									subIndicadorRes := subIndicador[ind]
									treeData := treeDatas[0]
									dataIndicador := make(map[string]interface{})
									auxSubIndicador := subIndicadorRes["sub"].([]map[string]interface{})
									for subInd := 0; subInd < len(auxSubIndicador); subInd++ {
										dataIndicador[auxSubIndicador[subInd]["nombre"].(string)] = treeData[auxSubIndicador[subInd]["id"].(string)]
									}
									titulosArmonizacion[subIndicadorRes["nombre"].(string)] = dataIndicador
								}

								datosArmonizacion["indicadores"] = titulosArmonizacion
								arregloLineamieto, _ = ArbolArmonizacionV2(armonizacionTercerNivel.(string))
								arregloLineamietoPI, _ = ArbolArmonizacionPIV2(armonizacionTercerNivelPI.(string))

								generalData := make(map[string]interface{})
								if err := request.GetJson("http://"+beego.AppConfig.String("OikosService")+"/dependencia_tipo_dependencia?query=DependenciaId:"+body["unidad_id"].(string), &respuestaUnidad); err == nil {
									aux := respuestaUnidad[0]
									dependenciaNombre := aux["DependenciaId"].(map[string]interface{})
									nombreUnidad = dependenciaNombre["Nombre"].(string)
								} else {
									panic(map[string]interface{}{"funcion": "GetUnidades", "err": "Error ", "status": "400", "log": err})
								}

								generalData["nombreUnidad"] = nombreUnidad
								generalData["nombreActividad"] = actividadName
								generalData["numeroActividad"] = index
								generalData["datosArmonizacion"] = arregloLineamieto
								generalData["datosArmonizacionPI"] = arregloLineamietoPI
								generalData["datosComplementarios"] = datosArmonizacion

								arregloPlanAnual = append(arregloPlanAnual, generalData)
							}
							break
						}
					}
					unidadNombre = arregloPlanAnual[0]["nombreUnidad"].(string)
					sheetName := "Actividades del plan"
					indexPlan, _ := consolidadoExcelPlanAnual.NewSheet(sheetName)

					if planes == 0 {
						consolidadoExcelPlanAnual.DeleteSheet("Sheet1")

						disable := false
						if err := consolidadoExcelPlanAnual.SetSheetView(sheetName, -1, &excelize.ViewOptions{
							ShowGridLines: &disable,
						}); err != nil {
							logs.Error(err)
						}
					}
					consolidadoExcelPlanAnual.MergeCell(sheetName, "B1", "D1")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "E1", "G1")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "H1", "H2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "I1", "I2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "J1", "J2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "K1", "K2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "P1", "P2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "Q1", "Q2")
					consolidadoExcelPlanAnual.MergeCell(sheetName, "L1", "O1")
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "B", "B", 18)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "C", "C", 11)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "D", "D", 35)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "E", "E", 16)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "F", "G", 35)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "H", "H", 6)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "I", "J", 12)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "K", "K", 30)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "L", "M", 52)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "N", "N", 30)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "O", "O", 10)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "P", "P", 25)
					consolidadoExcelPlanAnual.SetColWidth(sheetName, "Q", "Q", 30)
					consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B1", "Q1", stylehead)
					consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B2", "Q2", styletitles)

					// encabezado excel
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "B1", "Armonización PED")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "B2", "Lineamiento")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "C2", "Meta")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "D2", "Estrategias")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "E1", "Armonización Plan Indicativo")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "E2", "Ejes transformadores")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "F2", "Lineamientos de acción")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "G2", "Estrategias")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "H2", "N°.")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "I2", "Peso (%)")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "J2", "Periodo de ejecución")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "K2", "Actividad")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "L1", "Indicador")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "L2", "Nombre")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "M2", "Fórmula")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "N2", "Criterio del indicador")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "O2", "Meta")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "P2", "Producto esperado")
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q2", "Unidad o grupo responsable")

					rowPos := 3

					for excelPlan := 0; excelPlan < len(arregloPlanAnual); excelPlan++ {

						datosExcelPlan := arregloPlanAnual[excelPlan]
						armoPED := datosExcelPlan["datosArmonizacion"].([]map[string]interface{})
						armoPI := datosExcelPlan["datosArmonizacionPI"].([]map[string]interface{})
						datosComplementarios := datosExcelPlan["datosComplementarios"].(map[string]interface{})
						indicadores := datosComplementarios["indicadores"].(map[string]interface{})

						MaxRowsXActivity := reporteshelper.MinComMul_Armonization(armoPED, armoPI, len(indicadores))

						y_lin := rowPos
						h_lin := MaxRowsXActivity / len(armoPED)

						for _, lin := range armoPED {
							consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1))
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1), styleLineamiento, styleLineamientoSombra)
							y_met := y_lin
							h_met := h_lin / len(lin["meta"].([]map[string]interface{}))
							for _, met := range lin["meta"].([]map[string]interface{}) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1))
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(y_met), met["nombreMeta"])
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1), stylecontentC, stylecontentCS)
								y_est := y_met
								h_est := h_met / len(met["estrategias"].([]map[string]interface{}))
								for _, est := range met["estrategias"].([]map[string]interface{}) {
									consolidadoExcelPlanAnual.MergeCell(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1))
									consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(y_est), est["descripcionEstrategia"])
									if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
									} else {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
									}
									y_est += h_est
								}
								y_met += h_met
							}
							y_lin += h_lin
						}

						y_eje := rowPos
						h_eje := MaxRowsXActivity / len(armoPI)
						for _, eje := range armoPI {
							consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1))
							consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(y_eje), eje["nombreFactor"])
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1), stylecontentC, stylecontentCS)
							y_lin := y_eje
							h_lin := h_eje / len(eje["lineamientos"].([]map[string]interface{}))
							for _, lin := range eje["lineamientos"].([]map[string]interface{}) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1))
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1), stylecontentC, stylecontentCS)
								y_est := y_lin
								h_est := h_lin / len(lin["estrategias"].([]map[string]interface{}))
								for _, est := range lin["estrategias"].([]map[string]interface{}) {
									consolidadoExcelPlanAnual.MergeCell(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1))
									consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(y_est), est["descripcionEstrategia"])
									if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
									} else {
										reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
									}
									y_est += h_est
								}
								y_lin += h_lin
							}
							y_eje += h_eje
						}

						consolidadoExcelPlanAnual.MergeCell(sheetName, "H"+fmt.Sprint(rowPos), "H"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "I"+fmt.Sprint(rowPos), "I"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "J"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.MergeCell(sheetName, "K"+fmt.Sprint(rowPos), "K"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(rowPos), excelPlan+1)
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(rowPos), datosComplementarios["Ponderación de la actividad"])
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "J"+fmt.Sprint(rowPos), datosComplementarios["Periodo de ejecución"])
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "K"+fmt.Sprint(rowPos), datosExcelPlan["nombreActividad"].(string))
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "H"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "K"+fmt.Sprint(rowPos), "L"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontent, stylecontentS)

						y_ind := rowPos
						h_ind := MaxRowsXActivity / len(indicadores)
						idx := int(0)
						var indicadoresVacios []map[string]interface{}
						var indicadoresNoVacios []map[string]interface{}
						var indicadoresOrdenados []map[string]interface{}

						for _, indicador := range indicadores {
							auxIndicador := indicador
							var nombreIndicador interface{}
							var formula interface{}
							var criterio interface{}
							var meta interface{}
							for key, element := range auxIndicador.(map[string]interface{}) {
								if strings.Contains(strings.ToLower(key), "nombre") {
									nombreIndicador = element
								}
								if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
									formula = element
								}
								if strings.Contains(strings.ToLower(key), "criterio") {
									criterio = element
								}
								if strings.Contains(strings.ToLower(key), "meta") {
									meta = element
								}
							}
							if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
								indicadoresVacios = append(indicadoresVacios, auxIndicador.(map[string]interface{}))
							} else {
								indicadoresNoVacios = append(indicadoresNoVacios, auxIndicador.(map[string]interface{}))
							}
						}

						indicadoresOrdenados = append(indicadoresNoVacios, indicadoresVacios...)
						for _, indicador := range indicadoresOrdenados {
							auxIndicador := indicador
							var nombreIndicador interface{}
							var formula interface{}
							var criterio interface{}
							var meta interface{}
							for key, element := range auxIndicador {
								if strings.Contains(strings.ToLower(key), "nombre") {
									nombreIndicador = element
								}
								if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
									formula = element
								}
								if strings.Contains(strings.ToLower(key), "criterio") {
									criterio = element
								}
								if strings.Contains(strings.ToLower(key), "meta") {
									meta = element
								}
							}
							if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
								consolidadoExcelPlanAnual.MergeCell(sheetName, "L"+fmt.Sprint(y_ind-1), "L"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind-1), "M"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind-1), "N"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind-1), "O"+fmt.Sprint(y_ind+h_ind-1))
							} else {
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "L"+fmt.Sprint(y_ind), nombreIndicador)
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "M"+fmt.Sprint(y_ind), formula)
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "N"+fmt.Sprint(y_ind), criterio)
								consolidadoExcelPlanAnual.SetCellValue(sheetName, "O"+fmt.Sprint(y_ind), meta)
								consolidadoExcelPlanAnual.MergeCell(sheetName, "L"+fmt.Sprint(y_ind), "L"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind), "M"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind), "N"+fmt.Sprint(y_ind+h_ind-1))
								consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind), "O"+fmt.Sprint(y_ind+h_ind-1))
							}
							idx++
							if idx < len(indicadores) {
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "L"+fmt.Sprint(y_ind), "O"+fmt.Sprint(y_ind+h_ind-1), stylecontentCL, stylecontentCLS)
							} else {
								reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "L"+fmt.Sprint(y_ind), "O"+fmt.Sprint(y_ind+h_ind-1), stylecontentCLD, stylecontentCLDS)
							}
							y_ind += h_ind
						}

						consolidadoExcelPlanAnual.MergeCell(sheetName, "P"+fmt.Sprint(rowPos), "P"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "P"+fmt.Sprint(rowPos), datosComplementarios["Producto esperado"])
						consolidadoExcelPlanAnual.MergeCell(sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q"+fmt.Sprint(rowPos), datosComplementarios["Responsable"])
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "P"+fmt.Sprint(rowPos), "P"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)
						reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)

						rowPos += MaxRowsXActivity

						consolidadoExcelPlanAnual.SetActiveSheet(indexPlan)
					}
				}
			}

			consolidadoExcelPlanAnual, _ = TablaIdentificaciones(consolidadoExcelPlanAnual, plan_id, esReporteAntiguo)

			consolidadoExcelPlanAnual.InsertRows("Actividades del plan", 1, 7)
			consolidadoExcelPlanAnual.MergeCell("Actividades del plan", "C2", "P6")
			consolidadoExcelPlanAnual.SetCellStyle("Actividades del plan", "C2", "P6", styletitle)
			consolidadoExcelPlanAnual.SetCellStyle("Identificaciones", "C2", "G6", styletitle)

			if periodo[0] != nil {
				consolidadoExcelPlanAnual.SetCellValue("Actividades del plan", "C2", "Plan de Acción "+periodo[0]["Nombre"].(string)+"\n"+unidadNombre)
				consolidadoExcelPlanAnual.SetCellValue("Identificaciones", "C2", "Proyección de necesidades "+periodo[0]["Nombre"].(string)+"\n"+unidadNombre)
			} else {
				consolidadoExcelPlanAnual.SetCellValue("Actividades del plan", "C2", "Plan de Acción")
				consolidadoExcelPlanAnual.SetCellValue("Identificaciones", "C2", "Proyección de necesidades")
			}

			if err := consolidadoExcelPlanAnual.AddPicture("Actividades del plan", "B1", "static/img/UDEscudo2.png",
				&excelize.GraphicOptions{ScaleX: 0.1, ScaleY: 0.1, Positioning: "oneCell", OffsetX: 10}); err != nil {
				logs.Error(err)
			}
			if err := consolidadoExcelPlanAnual.AddPicture("Identificaciones", "B1", "static/img/UDEscudo2.png",
				&excelize.GraphicOptions{ScaleX: 0.1, ScaleY: 0.1, Positioning: "absolute", OffsetX: 10}); err != nil {
				logs.Error(err)
			}

			consolidadoExcelPlanAnual.SetColWidth("Actividades del plan", "A", "A", 2)
			buf, _ := consolidadoExcelPlanAnual.WriteToBuffer()
			strings.NewReader(buf.String())
			encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

			dataSend := make(map[string]interface{})
			dataSend["generalData"] = arregloPlanAnual
			dataSend["excelB64"] = encoded

			return dataSend, nil
		} else {
			return nil, errors.New("error de longitud")
		}
	} else {
		return nil, errors.New("error en ProcesarPlanAccionAnual: No se agrego el unidad_id en la petición")
	}
}

func GetActividades(subgrupo_id string) ([]map[string]interface{}, error) {
	var res map[string]interface{}
	var subgrupoDetalle map[string]interface{}
	var datoPlan map[string]interface{}
	var actividades []map[string]interface{}

	url := "http://" + beego.AppConfig.String("PlanesService") + "/subgrupo-detalle?query=subgrupo_id:" + subgrupo_id + "&fields=dato_plan"
	if err := request.GetJson(url, &res); err != nil {
		return nil, errors.New("error al procesar la peticion " + err.Error())
	}
	aux := make([]map[string]interface{}, 1)
	request.LimpiezaRespuestaRefactor(res, &aux)
	subgrupoDetalle = aux[0]
	if subgrupoDetalle["dato_plan"] != nil {
		dato_plan_str := subgrupoDetalle["dato_plan"].(string)
		json.Unmarshal([]byte(dato_plan_str), &datoPlan)
		for _, element := range datoPlan {
			if element.(map[string]interface{})["activo"] == true {
				actividades = append(actividades, element.(map[string]interface{}))
			}
		}
	}
	return actividades, nil
}

func ArbolArmonizacionV2(armonizacion string) ([]map[string]interface{}, error) {

	var estrategias []map[string]interface{}
	var metas []map[string]interface{}
	var lineamientos []map[string]interface{}
	var arregloArmo []map[string]interface{}

	if armonizacion != "" {
		armonizacionPED := strings.Split(armonizacion, ",")
		for i := 0; i < len(armonizacionPED); i++ {
			var respuesta map[string]interface{}
			var respuestaSubgrupo map[string]interface{}
			if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo/"+armonizacionPED[i], &respuesta); err == nil {
				request.LimpiezaRespuestaRefactor(respuesta, &respuestaSubgrupo)

				if len(respuestaSubgrupo) > 0 {
					nombre := strings.ToLower(respuestaSubgrupo["nombre"].(string))
					if strings.Contains(nombre, "lineamiento") {
						lineamientos = append(lineamientos, respuestaSubgrupo)
					} else if strings.Contains(nombre, "meta") {
						metas = append(metas, respuestaSubgrupo)
					} else if strings.Contains(nombre, "estrategia") {
						estrategias = append(estrategias, respuestaSubgrupo)
					}
				}
			} else {
				return nil, errors.New("error al procesar la peticion GetUnidades	" + err.Error())
			}
		}

		for i := 0; i < len(lineamientos); i++ {
			arregloArmo = append(arregloArmo, map[string]interface{}{
				"_id":                  lineamientos[i]["_id"],
				"nombreLineamiento":    lineamientos[i]["nombre"],
				"meta":                 []map[string]interface{}{},
				"nombrePlanDesarrollo": "Plan Estrategico de Desarrollo",
				"hijos":                lineamientos[i]["hijos"],
			})
		}

		for i := 0; i < len(metas); i++ {
			foundPadreMeta := false
			for j := 0; j < len(arregloArmo); j++ {
				if arregloArmo[j]["_id"] == metas[i]["padre"] {
					arregloArmo[j]["meta"] = append(arregloArmo[j]["meta"].([]map[string]interface{}), map[string]interface{}{
						"_id":         metas[i]["_id"],
						"nombreMeta":  metas[i]["nombre"],
						"estrategias": []map[string]interface{}{},
					})
					foundPadreMeta = true
					break
				}
			}
			if !foundPadreMeta {
				arregloArmo = append(arregloArmo, map[string]interface{}{
					"_id":               metas[i]["padre"],
					"nombreLineamiento": "No seleccionado",
					"meta": []map[string]interface{}{
						{
							"_id":         metas[i]["_id"],
							"nombreMeta":  metas[i]["nombre"],
							"estrategias": []map[string]interface{}{},
						},
					},
					"nombrePlanDesarrollo": "Plan Estrategico de Desarrollo",
					"hijos":                []interface{}{},
				})
			}
		}

		for i := 0; i < len(estrategias); i++ {
			foundPadreEstrategia := false
			for j := 0; j < len(arregloArmo); j++ {
				for k := 0; k < len(arregloArmo[j]["meta"].([]map[string]interface{})); k++ {
					if arregloArmo[j]["meta"].([]map[string]interface{})[k]["_id"] == estrategias[i]["padre"] {
						arregloArmo[j]["meta"].([]map[string]interface{})[k]["estrategias"] = append(arregloArmo[j]["meta"].([]map[string]interface{})[k]["estrategias"].([]map[string]interface{}), map[string]interface{}{
							"_id":                   estrategias[i]["_id"],
							"nombreEstrategia":      estrategias[i]["nombre"],
							"descripcionEstrategia": estrategias[i]["descripcion"],
						})
						foundPadreEstrategia = true
						break
					}
				}
			}
			if !foundPadreEstrategia {
				for j := 0; j < len(arregloArmo); j++ {
					for k := 0; k < len(arregloArmo[j]["hijos"].([]interface{})); k++ {
						if arregloArmo[j]["hijos"].([]interface{})[k] == estrategias[i]["padre"] {
							arregloArmo[j]["meta"] = append(arregloArmo[j]["meta"].([]map[string]interface{}), map[string]interface{}{
								"_id":        estrategias[i]["padre"],
								"nombreMeta": "No seleccionado",
								"estrategias": []map[string]interface{}{
									{
										"_id":                   estrategias[i]["_id"],
										"nombreEstrategia":      estrategias[i]["nombre"],
										"descripcionEstrategia": estrategias[i]["descripcion"],
									},
								},
							})
							foundPadreEstrategia = true
							break
						}
					}
					if foundPadreEstrategia {
						break
					}
				}
			}
			if !foundPadreEstrategia {
				arregloArmo = append(arregloArmo, map[string]interface{}{
					"_id":               "",
					"nombreLineamiento": "No seleccionado",
					"meta": []map[string]interface{}{
						{
							"_id":        estrategias[i]["padre"],
							"nombreMeta": "No seleccionado",
							"estrategias": []map[string]interface{}{
								{
									"_id":                   estrategias[i]["_id"],
									"nombreEstrategia":      estrategias[i]["nombre"],
									"descripcionEstrategia": estrategias[i]["descripcion"],
								},
							},
						},
					},
					"nombrePlanDesarrollo": "Plan Estrategico de Desarrollo",
					"hijos": []interface{}{
						estrategias[i]["padre"],
					},
				})
			}
		}

		if len(arregloArmo) > 0 {
			for i := 0; i < len(arregloArmo); i++ {
				if len(arregloArmo[i]["meta"].([]map[string]interface{})) == 0 {
					arregloArmo[i]["meta"] = append(arregloArmo[i]["meta"].([]map[string]interface{}), map[string]interface{}{
						"_id":        "",
						"nombreMeta": "No seleccionado",
						"estrategias": []map[string]interface{}{
							{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							},
						},
					})
				} else {
					for j := 0; j < len(arregloArmo[i]["meta"].([]map[string]interface{})); j++ {
						if len(arregloArmo[i]["meta"].([]map[string]interface{})[j]["estrategias"].([]map[string]interface{})) == 0 {
							arregloArmo[i]["meta"].([]map[string]interface{})[j]["estrategias"] = append(arregloArmo[i]["meta"].([]map[string]interface{})[j]["estrategias"].([]map[string]interface{}), map[string]interface{}{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							})
						}
					}
				}
				delete(arregloArmo[i], "hijos")
			}
		} else {
			arregloArmo = append(arregloArmo, map[string]interface{}{
				"_id":               "",
				"nombreLineamiento": "No seleccionado",
				"meta": []map[string]interface{}{
					{
						"_id":        "",
						"nombreMeta": "No seleccionado",
						"estrategias": []map[string]interface{}{
							{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							},
						},
					},
				},
				"nombrePlanDesarrollo": "Plan Estrategico de Desarrollo",
			})
		}
	} else {
		arregloArmo = append(arregloArmo, map[string]interface{}{
			"_id":               "",
			"nombreLineamiento": "No seleccionado",
			"meta": []map[string]interface{}{
				{
					"_id":        "",
					"nombreMeta": "No seleccionado",
					"estrategias": []map[string]interface{}{
						{
							"_id":                   "",
							"nombreEstrategia":      "No seleccionado",
							"descripcionEstrategia": "No seleccionado",
						},
					},
				},
			},
			"nombrePlanDesarrollo": "Plan Estrategico de Desarrollo",
		})
	}

	return arregloArmo, nil
}

func ArbolArmonizacion(armonizacion string) ([]map[string]interface{}, error) {

	var respuesta map[string]interface{}
	var lineamientos []map[string]interface{}
	var metas []map[string]interface{}
	var estrategias []map[string]interface{}
	var arreglo []map[string]interface{}
	armonizacionPED := strings.Split(armonizacion, ",")
	for i := 0; i < len(armonizacionPED); i++ {
		var respuestaSubgrupo map[string]interface{}
		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo/"+armonizacionPED[i], &respuesta); err == nil {
			request.LimpiezaRespuestaRefactor(respuesta, &respuestaSubgrupo)
			if len(respuestaSubgrupo) > 0 {
				nombre := strings.ToLower(respuestaSubgrupo["nombre"].(string))
				if strings.Contains(nombre, "lineamiento") {
					lineamientos = append(lineamientos, respuestaSubgrupo)
				}
				if strings.Contains(nombre, "meta") {
					metas = append(metas, respuestaSubgrupo)
				}
				if strings.Contains(nombre, "estrategia") {
					estrategias = append(estrategias, respuestaSubgrupo)
				}
			}
		} else {
			return nil, errors.New("error al procesar la peticion GetUnidades	" + err.Error())
		}
	}

	for i := 0; i < len(lineamientos); i++ {
		meta := make(map[string]interface{})
		lineamiento := make(map[string]interface{})
		estrategia := make(map[string]interface{})
		var arregloEstrategias []map[string]interface{}
		var arregloMetas []map[string]interface{}

		estrategia["_id"] = ""
		estrategia["nombreEstrategia"] = ""
		estrategia["descripcionEstrategia"] = ""

		arregloEstrategias = append(arregloEstrategias, estrategia)

		meta["_id"] = ""
		meta["nombreMeta"] = ""
		meta["estrategias"] = arregloEstrategias

		arregloMetas = append(arregloMetas, meta)

		lineamiento["_id"] = lineamientos[i]["_id"]
		lineamiento["nombreLineamiento"] = lineamientos[i]["nombre"]
		lineamiento["nombrePlanDesarrollo"] = "Plan Estrategico de Desarrollo"
		lineamiento["meta"] = arregloMetas

		arreglo = append(arreglo, lineamiento)
	}

	for i := 0; i < len(metas); i++ {

		meta := make(map[string]interface{})
		estrategia := make(map[string]interface{})
		var auxMeta = metas[i]
		bandera := false
		var arregloEstrategias []map[string]interface{}
		estrategia["_id"] = ""
		estrategia["nombreEstrategia"] = ""
		estrategia["descripcionEstrategia"] = ""

		arregloEstrategias = append(arregloEstrategias, estrategia)

		meta["_id"] = auxMeta["_id"]
		meta["nombreMeta"] = auxMeta["nombre"]
		meta["estrategias"] = arregloEstrategias

		for j := 0; j < len(arreglo); j++ {
			if arreglo[j]["_id"] == auxMeta["padre"] {

				bandera = true
				aux := arreglo[j]["meta"].([]map[string]interface{})
				if aux[0]["_id"] == "" {
					aux = append(aux[:0], aux[1:]...)
				}
				aux = append(aux, meta)
				arreglo[j]["meta"] = aux
				break
			}
		}
		if !bandera {
			lineamiento := make(map[string]interface{})
			var arregloMetas []map[string]interface{}
			arregloMetas = append(arregloMetas, meta)
			lineamiento["_id"] = ""
			lineamiento["nombreLineamiento"] = ""
			lineamiento["nombrePlanDesarrollo"] = "Plan Estrategico de Desarrollo"
			lineamiento["meta"] = arregloMetas
			arreglo = append(arreglo, lineamiento)
		}
	}

	for i := 0; i < len(estrategias); i++ {
		var auxEstrategia = estrategias[i]
		estrategia := make(map[string]interface{})
		bandera := false

		estrategia["_id"] = auxEstrategia["_id"]
		estrategia["nombreEstrategia"] = auxEstrategia["nombre"]
		estrategia["descripcionEstrategia"] = auxEstrategia["descripcion"]

		for j := 0; j < len(metas); j++ {
			if metas[j]["_id"] == auxEstrategia["padre"] {
				for n := 0; n < len(arreglo); n++ {
					if arreglo[n]["_id"] == metas[j]["padre"] {
						bandera = true
						auxMetas := arreglo[n]["meta"].([]map[string]interface{})

						for k := 0; k < len(auxMetas); k++ {
							if auxMetas[k]["_id"] == auxEstrategia["padre"] {
								aux2 := auxMetas[k]["estrategias"].([]map[string]interface{})
								if aux2[0]["_id"] == "" {
									aux2 = append(aux2[:0], aux2[1:]...)
								}
								aux2 = append(aux2, estrategia)
								auxMetas[k]["estrategias"] = aux2

								arreglo[n]["meta"] = auxMetas
								break
							}
						}
						break
					}
				}
				break
			}
		}
		if !bandera {
			meta := make(map[string]interface{})
			lineamiento := make(map[string]interface{})
			var arregloEstrategias []map[string]interface{}
			var arregloMetas []map[string]interface{}
			arregloEstrategias = append(arregloEstrategias, estrategia)

			meta["_id"] = ""
			meta["nombreMeta"] = ""
			meta["estrategias"] = arregloEstrategias

			arregloMetas = append(arregloMetas, meta)

			lineamiento["_id"] = ""
			lineamiento["nombreLineamiento"] = ""
			lineamiento["nombrePlanDesarrollo"] = "Plan Estrategico de Desarrollo"
			lineamiento["meta"] = arregloMetas
			arreglo = append(arreglo, lineamiento)
		}
	}

	return arreglo, nil
}

func ArbolArmonizacionPIV2(armonizacion string) ([]map[string]interface{}, error) {

	var estrategias []map[string]interface{}
	var lineamientos []map[string]interface{}
	var factores []map[string]interface{}
	var arregloArmo []map[string]interface{}

	if armonizacion != "" {
		armonizacionPI := strings.Split(armonizacion, ",")
		for i := 0; i < len(armonizacionPI); i++ {
			var respuesta map[string]interface{}
			var respuestaSubgrupo map[string]interface{}
			if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo/"+armonizacionPI[i], &respuesta); err == nil {
				request.LimpiezaRespuestaRefactor(respuesta, &respuestaSubgrupo)
				if len(respuestaSubgrupo) > 0 {
					nombre := strings.ToLower(respuestaSubgrupo["nombre"].(string))
					if (strings.Contains(nombre, "eje") || strings.Contains(nombre, "transformador")) || strings.Contains(nombre, "nivel 1") {
						factores = append(factores, respuestaSubgrupo)
					} else if strings.Contains(nombre, "lineamientos") || strings.Contains(nombre, "lineamiento") || strings.Contains(nombre, "nivel 2") {
						lineamientos = append(lineamientos, respuestaSubgrupo)
					} else if strings.Contains(nombre, "estrategia") || strings.Contains(nombre, "proyecto") || strings.Contains(nombre, "nivel 3") {
						estrategias = append(estrategias, respuestaSubgrupo)
					}
				}
			} else {
				return nil, errors.New("error al procesar la peticion ArbolArmonizacionPIV2	" + err.Error())
			}
		}

		for i := 0; i < len(factores); i++ {
			arregloArmo = append(arregloArmo, map[string]interface{}{
				"_id":                  factores[i]["_id"],
				"nombreFactor":         factores[i]["nombre"],
				"lineamientos":         []map[string]interface{}{},
				"nombrePlanDesarrollo": "Plan Indicativo",
				"hijos":                factores[i]["hijos"],
			})
		}

		for i := 0; i < len(lineamientos); i++ {
			foundPadreMeta := false
			for j := 0; j < len(arregloArmo); j++ {
				if arregloArmo[j]["_id"] == lineamientos[i]["padre"] {
					arregloArmo[j]["lineamientos"] = append(arregloArmo[j]["lineamientos"].([]map[string]interface{}), map[string]interface{}{
						"_id":               lineamientos[i]["_id"],
						"nombreLineamiento": lineamientos[i]["nombre"],
						"estrategias":       []map[string]interface{}{},
					})
					foundPadreMeta = true
					break
				}
			}
			if !foundPadreMeta {
				arregloArmo = append(arregloArmo, map[string]interface{}{
					"_id":          lineamientos[i]["padre"],
					"nombreFactor": "No seleccionado",
					"lineamientos": []map[string]interface{}{
						{
							"_id":               lineamientos[i]["_id"],
							"nombreLineamiento": lineamientos[i]["nombre"],
							"estrategias":       []map[string]interface{}{},
						},
					},
					"nombrePlanDesarrollo": "Plan Indicativo",
					"hijos":                []interface{}{},
				})
			}
		}

		for i := 0; i < len(estrategias); i++ {
			foundPadreEstrategia := false
			for j := 0; j < len(arregloArmo); j++ {
				for k := 0; k < len(arregloArmo[j]["lineamientos"].([]map[string]interface{})); k++ {
					if arregloArmo[j]["lineamientos"].([]map[string]interface{})[k]["_id"] == estrategias[i]["padre"] {
						arregloArmo[j]["lineamientos"].([]map[string]interface{})[k]["estrategias"] = append(arregloArmo[j]["lineamientos"].([]map[string]interface{})[k]["estrategias"].([]map[string]interface{}), map[string]interface{}{
							"_id":                   estrategias[i]["_id"],
							"nombreEstrategia":      estrategias[i]["nombre"],
							"descripcionEstrategia": estrategias[i]["descripcion"],
						})
						foundPadreEstrategia = true
						break
					}
				}
			}
			if !foundPadreEstrategia {
				for j := 0; j < len(arregloArmo); j++ {
					for k := 0; k < len(arregloArmo[j]["hijos"].([]interface{})); k++ {
						if arregloArmo[j]["hijos"].([]interface{})[k] == estrategias[i]["padre"] {
							arregloArmo[j]["lineamientos"] = append(arregloArmo[j]["lineamientos"].([]map[string]interface{}), map[string]interface{}{
								"_id":               estrategias[i]["padre"],
								"nombreLineamiento": "No seleccionado",
								"estrategias": []map[string]interface{}{
									{
										"_id":                   estrategias[i]["_id"],
										"nombreEstrategia":      estrategias[i]["nombre"],
										"descripcionEstrategia": estrategias[i]["descripcion"],
									},
								},
							})
							foundPadreEstrategia = true
							break
						}
					}
					if foundPadreEstrategia {
						break
					}
				}
			}
			if !foundPadreEstrategia {
				arregloArmo = append(arregloArmo, map[string]interface{}{
					"_id":          "",
					"nombreFactor": "No seleccionado",
					"lineamientos": []map[string]interface{}{
						{
							"_id":               estrategias[i]["padre"],
							"nombreLineamiento": "No seleccionado",
							"estrategias": []map[string]interface{}{
								{
									"_id":                   estrategias[i]["_id"],
									"nombreEstrategia":      estrategias[i]["nombre"],
									"descripcionEstrategia": estrategias[i]["descripcion"],
								},
							},
						},
					},
					"nombrePlanDesarrollo": "Plan Indicativo",
					"hijos": []interface{}{
						estrategias[i]["padre"],
					},
				})
			}
		}

		if len(arregloArmo) > 0 {
			for i := 0; i < len(arregloArmo); i++ {
				if len(arregloArmo[i]["lineamientos"].([]map[string]interface{})) == 0 {
					arregloArmo[i]["lineamientos"] = append(arregloArmo[i]["lineamientos"].([]map[string]interface{}), map[string]interface{}{
						"_id":               "",
						"nombreLineamiento": "No seleccionado",
						"estrategias": []map[string]interface{}{
							{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							},
						},
					})
				} else {
					for j := 0; j < len(arregloArmo[i]["lineamientos"].([]map[string]interface{})); j++ {
						if len(arregloArmo[i]["lineamientos"].([]map[string]interface{})[j]["estrategias"].([]map[string]interface{})) == 0 {
							arregloArmo[i]["lineamientos"].([]map[string]interface{})[j]["estrategias"] = append(arregloArmo[i]["lineamientos"].([]map[string]interface{})[j]["estrategias"].([]map[string]interface{}), map[string]interface{}{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							})
						}
					}
				}
				delete(arregloArmo[i], "hijos")
			}
		} else {
			arregloArmo = append(arregloArmo, map[string]interface{}{
				"_id":          "",
				"nombreFactor": "No seleccionado",
				"lineamientos": []map[string]interface{}{
					{
						"_id":               "",
						"nombreLineamiento": "No seleccionado",
						"estrategias": []map[string]interface{}{
							{
								"_id":                   "",
								"nombreEstrategia":      "No seleccionado",
								"descripcionEstrategia": "No seleccionado",
							},
						},
					},
				},
				"nombrePlanDesarrollo": "Plan Indicativo",
			})
		}
	} else {
		arregloArmo = append(arregloArmo, map[string]interface{}{
			"_id":          "",
			"nombreFactor": "No seleccionado",
			"lineamientos": []map[string]interface{}{
				{
					"_id":               "",
					"nombreLineamiento": "No seleccionado",
					"estrategias": []map[string]interface{}{
						{
							"_id":                   "",
							"nombreEstrategia":      "No seleccionado",
							"descripcionEstrategia": "No seleccionado",
						},
					},
				},
			},
			"nombrePlanDesarrollo": "Plan Indicativo",
		})
	}

	return arregloArmo, nil
}

func ArbolArmonizacionPI(armonizacion interface{}) ([]map[string]interface{}, error) {

	var respuesta map[string]interface{}
	var lineamientos []map[string]interface{}
	var factores []map[string]interface{}
	var estrategias []map[string]interface{}
	var arreglo []map[string]interface{}
	if armonizacion != "" {

		armonizacionPI := strings.Split(armonizacion.(string), ",")

		for i := 0; i < len(armonizacionPI); i++ {
			var respuestaSubgrupo map[string]interface{}
			if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo/"+armonizacionPI[i], &respuesta); err == nil {

				request.LimpiezaRespuestaRefactor(respuesta, &respuestaSubgrupo)
				if len(respuestaSubgrupo) > 0 {
					nombre := strings.ToLower(respuestaSubgrupo["nombre"].(string))
					if (strings.Contains(nombre, "eje") && strings.Contains(nombre, "transformador")) || strings.Contains(nombre, "nivel 1") {
						factores = append(factores, respuestaSubgrupo)
					}
					if strings.Contains(nombre, "lineamientos") || strings.Contains(nombre, "lineamiento") || strings.Contains(nombre, "nivel 2") {
						lineamientos = append(lineamientos, respuestaSubgrupo)
					}
					if strings.Contains(nombre, "estrategia") || strings.Contains(nombre, "proyecto") || strings.Contains(nombre, "nivel 3") {
						estrategias = append(estrategias, respuestaSubgrupo)
					}
				}
			} else {
				return nil, errors.New("error al procesar la peticion ArbolArmonizacionPI	" + err.Error())
			}
		}

		for i := 0; i < len(factores); i++ {
			factor := make(map[string]interface{})
			lineamiento := make(map[string]interface{})
			estrategia := make(map[string]interface{})
			var arregloEstrategias []map[string]interface{}
			var arregloLineamientos []map[string]interface{}

			estrategia["_id"] = ""
			estrategia["nombreEstrategia"] = ""
			estrategia["descripcionEstrategia"] = ""

			arregloEstrategias = append(arregloEstrategias, estrategia)

			lineamiento["_id"] = ""
			lineamiento["nombreLineamiento"] = ""
			lineamiento["estrategias"] = arregloEstrategias

			arregloLineamientos = append(arregloLineamientos, lineamiento)

			factor["_id"] = factores[i]["_id"]
			factor["nombreFactor"] = factores[i]["nombre"]
			factor["nombrePlanDesarrollo"] = "Plan Indicativo"
			factor["lineamientos"] = arregloLineamientos

			arreglo = append(arreglo, factor)
		}

		for i := 0; i < len(lineamientos); i++ {

			lineamiento := make(map[string]interface{})
			estrategia := make(map[string]interface{})
			var auxLineamiento = lineamientos[i]
			bandera := false
			var arregloEstrategias []map[string]interface{}
			estrategia["_id"] = ""
			estrategia["nombreEstrategia"] = ""
			estrategia["descripcionEstrategia"] = ""

			arregloEstrategias = append(arregloEstrategias, estrategia)

			lineamiento["_id"] = auxLineamiento["_id"]
			lineamiento["nombreLineamiento"] = auxLineamiento["nombre"]
			lineamiento["estrategias"] = arregloEstrategias

			for j := 0; j < len(arreglo); j++ {

				if arreglo[j]["_id"] == auxLineamiento["padre"] {

					bandera = true
					aux := arreglo[j]["lineamientos"].([]map[string]interface{})
					if aux[0]["_id"] == "" {
						aux = append(aux[:0], aux[1:]...)
					}
					aux = append(aux, lineamiento)
					arreglo[j]["lineamientos"] = aux
					break
				}
			}
			if !bandera {
				factor := make(map[string]interface{})
				var arregloLineamientos []map[string]interface{}
				arregloLineamientos = append(arregloLineamientos, lineamiento)
				factor["_id"] = ""
				factor["nombreFactor"] = ""
				factor["nombrePlanDesarrollo"] = "Plan Indicativo"
				factor["lineamientos"] = arregloLineamientos
				arreglo = append(arreglo, factor)
			}
		}

		for i := 0; i < len(estrategias); i++ {
			var auxEstrategia = estrategias[i]
			estrategia := make(map[string]interface{})
			bandera := false

			estrategia["_id"] = auxEstrategia["_id"]
			estrategia["nombreEstrategia"] = auxEstrategia["nombre"]
			estrategia["descripcionEstrategia"] = auxEstrategia["descripcion"]

			for j := 0; j < len(lineamientos); j++ {
				if lineamientos[j]["_id"] == auxEstrategia["padre"] {
					for n := 0; n < len(arreglo); n++ {
						if arreglo[n]["_id"] == lineamientos[j]["padre"] {
							bandera = true
							auxLineamientos := arreglo[n]["lineamientos"].([]map[string]interface{})

							for k := 0; k < len(auxLineamientos); k++ {
								if auxLineamientos[k]["_id"] == auxEstrategia["padre"] {
									aux2 := auxLineamientos[k]["estrategias"].([]map[string]interface{})
									if aux2[0]["_id"] == "" {
										aux2 = append(aux2[:0], aux2[1:]...)
									}
									aux2 = append(aux2, estrategia)
									auxLineamientos[k]["estrategias"] = aux2

									arreglo[n]["lineamientos"] = auxLineamientos
									break
								}
							}
							break
						}
					}
					break
				}
			}
			if !bandera {
				lineamiento := make(map[string]interface{})
				factor := make(map[string]interface{})
				var arregloEstrategias []map[string]interface{}
				var arregloLineamientos []map[string]interface{}
				arregloEstrategias = append(arregloEstrategias, estrategia)

				lineamiento["_id"] = ""
				lineamiento["nombreLineamiento"] = ""
				lineamiento["estrategias"] = arregloEstrategias

				arregloLineamientos = append(arregloLineamientos, lineamiento)

				factor["_id"] = ""
				factor["nombreFactor"] = ""
				factor["nombrePlanDesarrollo"] = "Plan Indicativo"
				factor["lineamientos"] = arregloLineamientos
				arreglo = append(arreglo, factor)
			}
		}
	} else {
		lineamiento := make(map[string]interface{})
		factor := make(map[string]interface{})
		estrategia := make(map[string]interface{})

		estrategia["_id"] = ""
		estrategia["nombreEstrategia"] = ""
		estrategia["descripcionEstrategia"] = ""

		var arregloEstrategias []map[string]interface{}
		var arregloLineamientos []map[string]interface{}
		arregloEstrategias = append(arregloEstrategias, estrategia)

		lineamiento["_id"] = ""
		lineamiento["nombreLineamiento"] = ""
		lineamiento["estrategias"] = arregloEstrategias

		arregloLineamientos = append(arregloLineamientos, lineamiento)

		factor["_id"] = ""
		factor["nombreFactor"] = ""
		factor["nombrePlanDesarrollo"] = "Plan Indicativo"
		factor["lineamientos"] = arregloLineamientos
		arreglo = append(arreglo, factor)
	}

	return arreglo, nil
}

func identificacionNueva(iddetail string) interface{} {
	var result interface{}
	var identi map[string]interface{}
	var data_identi []map[string]interface{}
	identificacionDetalle := map[string]interface{}{}

	url := "http://" + beego.AppConfig.String("PlanesService") + "/identificacion-detalle/" + iddetail
	err := request.GetJson(url, &identificacionDetalle)
	if err == nil && identificacionDetalle["Status"] == "200" && identificacionDetalle["Data"] != nil {
		dato_aux := identificacionDetalle["Data"].(map[string]interface{})["dato"].(string)
		if dato_aux == "{}" {
			result = "{}"
		} else {
			json.Unmarshal([]byte(dato_aux), &identi)
			for key := range identi {
				element := identi[key].(map[string]interface{})
				if element["activo"] == true {
					data_identi = append(data_identi, element)
				}
			}
			result = data_identi
		}
	} else {
		result = "{}"
	}
	return result
}

func TablaIdentificaciones(consolidadoExcelPlanAnual *excelize.File, planId string, esReporteAntiguo bool) (*excelize.File, error) {
	var res map[string]interface{}
	var identificaciones []map[string]interface{}
	var recursos []map[string]interface{}
	var contratistas []map[string]interface{}
	var docentes map[string]interface{}
	var data_identi []map[string]interface{}
	var rubro string
	var nombreRubro string

	url := "http://" + beego.AppConfig.String("PlanesService") + "/identificacion?query=plan_id:" + planId
	if err := request.GetJson(url, &res); err == nil {
		request.LimpiezaRespuestaRefactor(res, &identificaciones)
	}

	for _, identificacion := range identificaciones {
		nombre := strings.ToLower(identificacion["nombre"].(string))
		if strings.Contains(nombre, "recurso") {
			if identificacion["dato"] != nil {
				var dato map[string]interface{}
				dato_str := identificacion["dato"].(string)
				json.Unmarshal([]byte(dato_str), &dato)
				for key := range dato {
					element := dato[key].(map[string]interface{})
					if element["activo"] == true {
						data_identi = append(data_identi, element)
					}
				}
				recursos = data_identi
			}
		} else if strings.Contains(nombre, "contratista") {
			if identificacion["dato"] != nil {
				var dato map[string]interface{}
				var data_identi []map[string]interface{}
				dato_str := identificacion["dato"].(string)
				json.Unmarshal([]byte(dato_str), &dato)
				for key := range dato {
					element := dato[key].(map[string]interface{})
					if element["rubro"] == nil {
						rubro = "Información no suministrada"
					} else {
						rubro = element["rubro"].(string)
					}
					if element["rubroNombre"] == nil {
						nombreRubro = "Información no suministrada"
					} else {
						nombreRubro = element["rubroNombre"].(string)
					}
					if element["activo"] == true {
						data_identi = append(data_identi, element)
					}
				}
				contratistas = data_identi
			}
		} else if strings.Contains(nombre, "docente") {
			if identificacion["dato"] != nil && identificacion["dato"] != "{}" {
				dato := map[string]interface{}{}
				result := make(map[string]interface{})
				dato_str := identificacion["dato"].(string)

				/* Se tiene en cuenta la nueva estructura la info ahora está en identificacion-detalle,
				pero tambien se tiene en cuenta la estructura de indentificaciones viejas (else) */
				if strings.Contains(dato_str, "ids_detalle") {
					json.Unmarshal([]byte(dato_str), &dato)

					iddetail := ""
					iddetail = dato["ids_detalle"].(map[string]interface{})["rhf"].(string)
					result["rhf"] = identificacionNueva(iddetail)

					iddetail = dato["ids_detalle"].(map[string]interface{})["rhv_pre"].(string)
					result["rhv_pre"] = identificacionNueva(iddetail)

					iddetail = dato["ids_detalle"].(map[string]interface{})["rhv_pos"].(string)
					result["rhv_pos"] = identificacionNueva(iddetail)

					iddetail = dato["ids_detalle"].(map[string]interface{})["rubros"].(string)
					result["rubros"] = identificacionNueva(iddetail)

					iddetail = dato["ids_detalle"].(map[string]interface{})["rubros_pos"].(string)
					result["rubros_pos"] = identificacionNueva(iddetail)
				} else {
					json.Unmarshal([]byte(dato_str), &dato)
					result["rhf"] = reporteshelper.IdentificacionAntigua(dato["rhf"].(string))
					result["rhv_pre"] = reporteshelper.IdentificacionAntigua(dato["rhv_pre"].(string))
					result["rhv_pos"] = reporteshelper.IdentificacionAntigua(dato["rhv_pos"].(string))
					if dato["rubros"] != nil {
						result["rubros"] = reporteshelper.IdentificacionAntigua(dato["rubros"].(string))
					}
					if dato["rubros_pos"] != nil {
						result["rubros"] = reporteshelper.IdentificacionAntigua(dato["rubros_pos"].(string))
					}
				}
				docentes = result
			}
		}
	}
	return construirTablas(consolidadoExcelPlanAnual, recursos, contratistas, docentes, rubro, nombreRubro, esReporteAntiguo)
}

func getDataDocentes(docentes map[string]interface{}, dependencia_id string) map[string]interface{} {
	var respuestaDependencia map[string]interface{}
	dataDocentes := map[string]interface{}{
		"tco":            0,
		"mto":            0,
		"hch":            0,
		"hcp":            0,
		"hchPos":         0,
		"hcpPos":         0,
		"valorPre":       0,
		"valorPos":       0,
		"nombreFacultad": "",
	}

	url := "http://" + beego.AppConfig.String("OikosService") + "/dependencia/" + dependencia_id
	if err := request.GetJson(url, &respuestaDependencia); err == nil {
		dataDocentes["nombreFacultad"] = respuestaDependencia["Nombre"]
	}

	for _, tipo := range []string{"rhf", "rhv_pre", "rhv_pos"} {
		if docentes[tipo] != nil && docentes[tipo] != "{}" {
			docentesTipo := docentes[tipo].([]map[string]interface{})
			for i := 0; i < len(docentesTipo); i++ {
				switch docentesTipo[i]["tipo"] {
				case "Tiempo Completo":
					dataDocentes["tco"] = dataDocentes["tco"].(int) + 1
				case "Medio Tiempo":
					dataDocentes["mto"] = dataDocentes["mto"].(int) + 1
				case "H. Catedra Honorarios":
					if tipo == "rhv_pos" {
						dataDocentes["hchPos"] = dataDocentes["hchPos"].(int) + 1
					} else {
						dataDocentes["hch"] = dataDocentes["hch"].(int) + 1
					}
				case "H. Catedra Prestacional":
					if tipo == "rhv_pos" {
						dataDocentes["hcpPos"] = dataDocentes["hcpPos"].(int) + 1
					} else {
						dataDocentes["hcp"] = dataDocentes["hcp"].(int) + 1
					}
				}
				auxTotal, err := reporteshelper.DeformatNumberInt(docentesTipo[i]["total"])
				if err == nil {
					if tipo == "rhv_pos" {
						dataDocentes["valorPos"] = dataDocentes["valorPos"].(int) + auxTotal
					} else {
						dataDocentes["valorPre"] = dataDocentes["valorPre"].(int) + auxTotal
					}
				}
			}
		}
	}
	return dataDocentes
}

func ProcesarPlanAccionAnualGeneral(body map[string]interface{}, nombre string) (map[string]interface{}, error) {
	var respuesta map[string]interface{}
	var planesFilter []map[string]interface{}
	var res map[string]interface{}
	var respuestaUnidad []map[string]interface{}
	var respuestaEstado map[string]interface{}
	var respuestaTipoPlan map[string]interface{}
	var estado map[string]interface{}
	var tipoPlan map[string]interface{}
	var subgrupos []map[string]interface{}
	var plan_id string
	var actividadName string
	var arregloPlanAnual []map[string]interface{}
	var arregloInfoReportes []map[string]interface{}
	var nombreUnidad string
	var idUnidad string
	var resPeriodo map[string]interface{}
	var periodo []map[string]interface{}
	contadorGeneral := 4

	consolidadoExcelPlanAnual := excelize.NewFile()

	url := "http://" + beego.AppConfig.String("PlanesService") + "/plan?query=activo:true,tipo_plan_id:" + body["tipo_plan_id"].(string) + ",vigencia:" + body["vigencia"].(string) + ",estado_plan_id:" + body["estado_plan_id"].(string) + ",nombre:" + nombre + "&fields=_id,dependencia_id,estado_plan_id,tipo_plan_id"
	if err := request.GetJson(url, &respuesta); err != nil {
		return nil, errors.New("error al procesar la peticion " + err.Error())
	}
	request.LimpiezaRespuestaRefactor(respuesta, &planesFilter)

	for _, planes := range planesFilter {
		if idUnidad != planes["dependencia_id"].(string) {
			url2 := "http://" + beego.AppConfig.String("OikosService") + "/dependencia_tipo_dependencia?query=DependenciaId:" + planes["dependencia_id"].(string)
			if err := request.GetJson(url2, &respuestaUnidad); err != nil {
				return nil, errors.New("error al procesar la peticion " + err.Error())
			}
			planes["nombreUnidad"] = respuestaUnidad[0]["DependenciaId"].(map[string]interface{})["Nombre"].(string)
		}
	}

	sort.SliceStable(planesFilter, func(i, j int) bool {
		a := (planesFilter)[i]["nombreUnidad"].(string)
		b := (planesFilter)[j]["nombreUnidad"].(string)
		return a < b
	})

	for planes := 0; planes < len(planesFilter); planes++ {
		reporteshelper.Limpiar()
		planesFilterData := planesFilter[planes]
		plan_id = planesFilterData["_id"].(string)
		infoReporte := make(map[string]interface{})
		url3 := "http://" + beego.AppConfig.String("PlanesService") + "/subgrupo?query=padre:" + plan_id + "&fields=nombre,_id,hijos,activo"
		if err := request.GetJson(url3, &res); err != nil {
			return nil, errors.New("error al procesar la peticion status: 500" + err.Error())
		}
		request.LimpiezaRespuestaRefactor(res, &subgrupos)

		for i := 0; i < len(subgrupos); i++ {
			if strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "actividad") && strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "general") {
				actividades, _ := GetActividades(subgrupos[i]["_id"].(string))
				var arregloLineamieto []map[string]interface{}
				var arregloLineamietoPI []map[string]interface{}
				sort.SliceStable(actividades, func(i int, j int) bool {
					if _, ok := actividades[i]["index"].(float64); ok {
						actividades[i]["index"] = fmt.Sprintf("%v", int(actividades[i]["index"].(float64)))
					}
					if _, ok := actividades[j]["index"].(float64); ok {
						actividades[j]["index"] = fmt.Sprintf("%v", int(actividades[j]["index"].(float64)))
					}
					aux, _ := strconv.Atoi((actividades[i]["index"]).(string))
					aux1, _ := strconv.Atoi((actividades[j]["index"]).(string))
					return aux < aux1
				})
				reporteshelper.LimpiarDetalles()
				for j := 0; j < len(actividades); j++ {
					var index string
					arregloLineamieto = nil
					arregloLineamietoPI = nil
					actividad := actividades[j]
					actividadName = actividad["dato"].(string)
					switch v := actividad["index"].(type) {
					case int:
						// Convertir el número a string
						index = strconv.Itoa(v)
					case float64:
						// Convertir el número a string
						index = strconv.FormatFloat(v, 'f', -1, 64)
					case string:
						index = v
					}
					datosArmonizacion := make(map[string]interface{})
					titulosArmonizacion := make(map[string]interface{})

					tree := reporteshelper.BuildTreeFa(subgrupos, index)
					treeDatos := tree[0]
					treeDatas := tree[1]
					treeArmo := tree[2]
					armonizacionTercer := treeArmo[0]
					var armonizacionTercerNivel interface{}
					var armonizacionTercerNivelPI interface{}
					if armonizacionTercer["armo"] != nil {
						armonizacionTercerNivel = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPED"]
						armonizacionTercerNivelPI = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPI"]
					}

					for datoGeneral := 0; datoGeneral < len(treeDatos); datoGeneral++ {
						treeDato := treeDatos[datoGeneral]
						treeData := treeDatas[0]
						if treeDato["sub"] == "" {
							nombreMinuscula := strings.ToLower(treeDato["nombre"].(string))
							if strings.Contains(nombreMinuscula, "ponderación") || strings.Contains(nombreMinuscula, "ponderacion") && strings.Contains(nombreMinuscula, "actividad") {
								datosArmonizacion["Ponderación de la actividad"] = treeData[treeDato["id"].(string)]
							} else if strings.Contains(nombreMinuscula, "período") || strings.Contains(nombreMinuscula, "periodo") && strings.Contains(nombreMinuscula, "ejecucion") || strings.Contains(nombreMinuscula, "ejecución") {
								datosArmonizacion["Periodo de ejecución"] = treeData[treeDato["id"].(string)]
							} else if strings.Contains(nombreMinuscula, "actividad") && strings.Contains(nombreMinuscula, "general") {
								datosArmonizacion["Actividad general"] = treeData[treeDato["id"].(string)]
							} else if strings.Contains(nombreMinuscula, "tarea") || strings.Contains(nombreMinuscula, "actividades específicas") {
								datosArmonizacion["Tareas"] = treeData[treeDato["id"].(string)]
							} else if strings.Contains(nombreMinuscula, "producto") {
								datosArmonizacion["Producto esperado"] = treeData[treeDato["id"].(string)]
							} else {
								datosArmonizacion[treeDato["nombre"].(string)] = treeData[treeDato["id"].(string)]
							}
						}
					}
					var treeIndicador map[string]interface{}
					auxTree := tree[0]
					for i := 0; i < len(auxTree); i++ {
						subgrupo := auxTree[i]
						if strings.Contains(strings.ToLower(subgrupo["nombre"].(string)), "indicador") {
							treeIndicador = auxTree[i]
						}
					}

					subIndicador := treeIndicador["sub"].([]map[string]interface{})
					for ind := 0; ind < len(subIndicador); ind++ {
						subIndicadorRes := subIndicador[ind]
						treeData := treeDatas[0]
						dataIndicador := make(map[string]interface{})
						auxSubIndicador := subIndicadorRes["sub"].([]map[string]interface{})
						for subInd := 0; subInd < len(auxSubIndicador); subInd++ {
							dataIndicador[auxSubIndicador[subInd]["nombre"].(string)] = treeData[auxSubIndicador[subInd]["id"].(string)]
						}
						titulosArmonizacion[subIndicadorRes["nombre"].(string)] = dataIndicador
					}

					datosArmonizacion["indicadores"] = titulosArmonizacion
					//if armonizacionTercerNivel != nil {
					//arregloLineamieto = reporteshelper.ArbolArmonizacion(armonizacionTercerNivel.(string))
					arregloLineamieto, _ = ArbolArmonizacionV2(armonizacionTercerNivel.(string))
					//} else {
					//	arregloLineamieto = []map[string]interface{}{}
					//}
					//if armonizacionTercerNivelPI != nil {
					//arregloLineamietoPI = reporteshelper.ArbolArmonizacionPI(armonizacionTercerNivelPI)
					arregloLineamietoPI, _ = ArbolArmonizacionPIV2(armonizacionTercerNivelPI.(string))
					//} else {
					//	arregloLineamietoPI = []map[string]interface{}{}
					//}

					generalData := make(map[string]interface{})

					nombreUnidad = planesFilterData["nombreUnidad"].(string)
					generalData["nombreUnidad"] = nombreUnidad
					generalData["nombreActividad"] = actividadName
					generalData["numeroActividad"] = index
					generalData["datosArmonizacion"] = arregloLineamieto
					generalData["datosArmonizacionPI"] = arregloLineamietoPI
					generalData["datosComplementarios"] = datosArmonizacion
					arregloPlanAnual = append(arregloPlanAnual, generalData)
				}
				break
			}
		}

		url4 := "http://" + beego.AppConfig.String("PlanesService") + "/estado-plan/" + planesFilter[planes]["estado_plan_id"].(string)
		if err := request.GetJson(url4, &respuestaEstado); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		}
		request.LimpiezaRespuestaRefactor(respuestaEstado, &estado)

		url5 := "http://" + beego.AppConfig.String("PlanesService") + "/tipo-plan/" + planesFilter[planes]["tipo_plan_id"].(string)
		if err := request.GetJson(url5, &respuestaTipoPlan); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		}
		request.LimpiezaRespuestaRefactor(respuestaTipoPlan, &tipoPlan)

		infoReporte["tipo_plan"] = tipoPlan["nombre"]
		infoReporte["vigencia"] = body["vigencia"]
		infoReporte["estado_plan"] = estado["nombre"]
		infoReporte["nombre_unidad"] = nombreUnidad

		arregloInfoReportes = append(arregloInfoReportes, infoReporte)

		rowPos := contadorGeneral + 5

		unidadNombre := arregloPlanAnual[0]["nombreUnidad"]
		sheetName := "REPORTE GENERAL"
		indexPlan, _ := consolidadoExcelPlanAnual.NewSheet(sheetName)

		if planes == 0 {
			consolidadoExcelPlanAnual.DeleteSheet("Sheet1")
			consolidadoExcelPlanAnual.InsertCols("REPORTE GENERAL", "A", 1)
			disable := false
			err := consolidadoExcelPlanAnual.SetSheetView(sheetName, -1, &excelize.ViewOptions{ShowGridLines: &disable})
			if err != nil {
				return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
			}

			url6 := "http://" + beego.AppConfig.String("ParametrosService") + `/periodo?query=Id:` + body["vigencia"].(string)
			if err := request.GetJson(url6, &resPeriodo); err != nil {
				return nil, errors.New("error al procesar la peticion " + err.Error())
			}
			request.LimpiezaRespuestaRefactor(resPeriodo, &periodo)
		}

		stylehead, _ := reporteshelper.EstiloExcel(consolidadoExcelPlanAnual, "center", "center", ColorRojo, true)
		styletitles, _ := reporteshelper.EstiloExcel(consolidadoExcelPlanAnual, "center", "center", ColorGrisClaro, false)
		stylecontent, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "justify", "center", "", 0, false)
		stylecontentS, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "justify", "center", ColorGrisClaro, 0, true)
		stylecontentC, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "center", "center", "", 0, false)
		stylecontentCS, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "center", "center", ColorGrisClaro, 0, true)
		styleLineamiento, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "center", "center", "", 90, false)
		styleLineamientoSombra, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelPlanAnual, "center", "center", ColorGrisClaro, 90, true)
		stylecontentCL, _ := reporteshelper.EstiloExcelBordes(consolidadoExcelPlanAnual, "center", "center", "", 4, false)
		stylecontentCLD, _ := reporteshelper.EstiloExcelBordes(consolidadoExcelPlanAnual, "center", "center", "", 1, false)
		stylecontentCLS, _ := reporteshelper.EstiloExcelBordes(consolidadoExcelPlanAnual, "center", "center", ColorGrisClaro, 4, true)
		stylecontentCLDS, _ := reporteshelper.EstiloExcelBordes(consolidadoExcelPlanAnual, "center", "center", ColorGrisClaro, 1, true)

		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contadorGeneral+1), "P"+fmt.Sprint(contadorGeneral+1))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(contadorGeneral+2), "D"+fmt.Sprint(contadorGeneral+2))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(contadorGeneral+2), "G"+fmt.Sprint(contadorGeneral+2))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "H"+fmt.Sprint(contadorGeneral+2), "H"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "I"+fmt.Sprint(contadorGeneral+2), "I"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "J"+fmt.Sprint(contadorGeneral+2), "J"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "K"+fmt.Sprint(contadorGeneral+2), "K"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "L"+fmt.Sprint(contadorGeneral+2), "L"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "Q"+fmt.Sprint(contadorGeneral+2), "Q"+fmt.Sprint(contadorGeneral+3))
		consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(contadorGeneral+2), "P"+fmt.Sprint(contadorGeneral+2))
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contadorGeneral+1, 20)
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contadorGeneral+2, 20)
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contadorGeneral+3, 20)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "B", "B", 19)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "C", "P", 35)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "C", "C", 13)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "E", "E", 16)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "H", "H", 6)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "I", "J", 12)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "K", "K", 30)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "L", "L", 35)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "M", "N", 52)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "O", "O", 30)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "P", "P", 10)
		consolidadoExcelPlanAnual.SetColWidth(sheetName, "Q", "Q", 30)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contadorGeneral+1), "Q"+fmt.Sprint(contadorGeneral+1), stylehead)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contadorGeneral+2), "Q"+fmt.Sprint(contadorGeneral+2), stylehead)
		consolidadoExcelPlanAnual.SetCellStyle(sheetName, "B"+fmt.Sprint(contadorGeneral+3), "Q"+fmt.Sprint(contadorGeneral+3), styletitles)
		consolidadoExcelPlanAnual.SetRowHeight(sheetName, contadorGeneral+3, 30)

		var tituloExcel string
		if periodo[0] != nil {
			tituloExcel = "Plan de acción " + periodo[0]["Nombre"].(string) + " - " + unidadNombre.(string)
		} else {
			tituloExcel = "Plan de acción - " + unidadNombre.(string)
		}

		// encabezado excel
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contadorGeneral+1), tituloExcel)
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contadorGeneral+2), "Armonización PED")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(contadorGeneral+3), "Lineamiento")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(contadorGeneral+3), "Meta")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(contadorGeneral+3), "Estrategias")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contadorGeneral+2), "Armonización Plan Indicativo")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(contadorGeneral+3), "Ejes transformadores")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(contadorGeneral+3), "Lineamientos de acción")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(contadorGeneral+3), "Estrategias")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(contadorGeneral+3), "N°.")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(contadorGeneral+3), "Ponderación de la actividad")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "J"+fmt.Sprint(contadorGeneral+3), "Periodo de ejecución")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "K"+fmt.Sprint(contadorGeneral+3), "Actividad")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "L"+fmt.Sprint(contadorGeneral+3), "Actividades específicas")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "M"+fmt.Sprint(contadorGeneral+2), "Indicador")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "M"+fmt.Sprint(contadorGeneral+3), "Nombre")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "N"+fmt.Sprint(contadorGeneral+3), "Fórmula")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "O"+fmt.Sprint(contadorGeneral+3), "Criterio del indicador")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "P"+fmt.Sprint(contadorGeneral+3), "Meta")
		consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q"+fmt.Sprint(contadorGeneral+3), "Producto esperado")
		consolidadoExcelPlanAnual.InsertRows(sheetName, 1, 1)

		for excelPlan := 0; excelPlan < len(arregloPlanAnual); excelPlan++ {
			datosExcelPlan := arregloPlanAnual[excelPlan]
			armoPED := datosExcelPlan["datosArmonizacion"].([]map[string]interface{})
			armoPI := datosExcelPlan["datosArmonizacionPI"].([]map[string]interface{})
			datosComplementarios := datosExcelPlan["datosComplementarios"].(map[string]interface{})
			indicadores := datosComplementarios["indicadores"].(map[string]interface{})

			MaxRowsXActivity := reporteshelper.MinComMul_Armonization(armoPED, armoPI, len(indicadores))

			y_lin := rowPos
			h_lin := MaxRowsXActivity / len(armoPED)
			for _, lin := range armoPED {
				consolidadoExcelPlanAnual.MergeCell(sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1))
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "B"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
				reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1), styleLineamiento, styleLineamientoSombra)
				y_met := y_lin
				h_met := h_lin / len(lin["meta"].([]map[string]interface{}))
				for _, met := range lin["meta"].([]map[string]interface{}) {
					consolidadoExcelPlanAnual.MergeCell(sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1))
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "C"+fmt.Sprint(y_met), met["nombreMeta"])
					reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1), stylecontentC, stylecontentCS)
					y_est := y_met
					h_est := h_met / len(met["estrategias"].([]map[string]interface{}))
					for _, est := range met["estrategias"].([]map[string]interface{}) {
						consolidadoExcelPlanAnual.MergeCell(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "D"+fmt.Sprint(y_est), est["descripcionEstrategia"])
						if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
						} else {
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
						}
						y_est += h_est
					}
					y_met += h_met
				}
				y_lin += h_lin
			}

			y_eje := rowPos
			h_eje := MaxRowsXActivity / len(armoPI)
			for _, eje := range armoPI {
				consolidadoExcelPlanAnual.MergeCell(sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1))
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "E"+fmt.Sprint(y_eje), eje["nombreFactor"])
				reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1), stylecontentC, stylecontentCS)
				y_lin := y_eje
				h_lin := h_eje / len(eje["lineamientos"].([]map[string]interface{}))
				for _, lin := range eje["lineamientos"].([]map[string]interface{}) {
					consolidadoExcelPlanAnual.MergeCell(sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1))
					consolidadoExcelPlanAnual.SetCellValue(sheetName, "F"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
					reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1), stylecontentC, stylecontentCS)
					y_est := y_lin
					h_est := h_lin / len(lin["estrategias"].([]map[string]interface{}))
					for _, est := range lin["estrategias"].([]map[string]interface{}) {
						consolidadoExcelPlanAnual.MergeCell(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1))
						consolidadoExcelPlanAnual.SetCellValue(sheetName, "G"+fmt.Sprint(y_est), est["descripcionEstrategia"])
						if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontentC, stylecontentCS)
						} else {
							reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontent, stylecontentS)
						}
						y_est += h_est
					}
					y_lin += h_lin
				}
				y_eje += h_eje
			}

			consolidadoExcelPlanAnual.MergeCell(sheetName, "H"+fmt.Sprint(rowPos), "H"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.MergeCell(sheetName, "I"+fmt.Sprint(rowPos), "I"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.MergeCell(sheetName, "J"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.MergeCell(sheetName, "K"+fmt.Sprint(rowPos), "K"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.MergeCell(sheetName, "L"+fmt.Sprint(rowPos), "L"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "H"+fmt.Sprint(rowPos), datosExcelPlan["numeroActividad"])
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "I"+fmt.Sprint(rowPos), datosComplementarios["Ponderación de la actividad"])
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "J"+fmt.Sprint(rowPos), datosComplementarios["Periodo de ejecución"])
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "K"+fmt.Sprint(rowPos), datosComplementarios["Actividad general"])
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "L"+fmt.Sprint(rowPos), datosComplementarios["Tareas"])
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "H"+fmt.Sprint(rowPos), "J"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "K"+fmt.Sprint(rowPos), "L"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontent, stylecontentS)

			y_ind := rowPos
			h_ind := MaxRowsXActivity / len(indicadores)
			idx := int(0)

			var indicadoresVacios []map[string]interface{}
			var indicadoresNoVacios []map[string]interface{}
			var indicadoresOrdenados []map[string]interface{}

			for _, indicador := range indicadores {
				auxIndicador := indicador
				var nombreIndicador interface{}
				var formula interface{}
				var criterio interface{}
				var meta interface{}
				for key, element := range auxIndicador.(map[string]interface{}) {
					if strings.Contains(strings.ToLower(key), "nombre") {
						nombreIndicador = element
					}
					if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
						formula = element
					}
					if strings.Contains(strings.ToLower(key), "criterio") {
						criterio = element
					}
					if strings.Contains(strings.ToLower(key), "meta") {
						meta = element
					}
				}
				if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
					indicadoresVacios = append(indicadoresVacios, auxIndicador.(map[string]interface{}))
				} else {
					indicadoresNoVacios = append(indicadoresNoVacios, auxIndicador.(map[string]interface{}))
				}
			}

			indicadoresOrdenados = append(indicadoresNoVacios, indicadoresVacios...)
			for _, indicador := range indicadoresOrdenados {
				auxIndicador := indicador
				var nombreIndicador interface{}
				var formula interface{}
				var criterio interface{}
				var meta interface{}
				for key, element := range auxIndicador {
					if strings.Contains(strings.ToLower(key), "nombre") {
						nombreIndicador = element
					}
					if strings.Contains(strings.ToLower(key), "formula") || strings.Contains(strings.ToLower(key), "fórmula") {
						formula = element
					}
					if strings.Contains(strings.ToLower(key), "criterio") {
						criterio = element
					}
					if strings.Contains(strings.ToLower(key), "meta") {
						meta = element
					}
				}
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "M"+fmt.Sprint(y_ind), nombreIndicador)
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "N"+fmt.Sprint(y_ind), formula)
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "O"+fmt.Sprint(y_ind), criterio)
				consolidadoExcelPlanAnual.SetCellValue(sheetName, "P"+fmt.Sprint(y_ind), meta)
				if (nombreIndicador == "" && formula == "" && criterio == "" && meta == "") || (nombreIndicador == nil && formula == nil && criterio == nil && meta == nil) {
					consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind-1), "M"+fmt.Sprint(y_ind))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind-1), "N"+fmt.Sprint(y_ind))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind-1), "O"+fmt.Sprint(y_ind))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "P"+fmt.Sprint(y_ind-1), "P"+fmt.Sprint(y_ind))
				} else {
					consolidadoExcelPlanAnual.MergeCell(sheetName, "M"+fmt.Sprint(y_ind), "M"+fmt.Sprint(y_ind+h_ind-1))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "N"+fmt.Sprint(y_ind), "N"+fmt.Sprint(y_ind+h_ind-1))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "O"+fmt.Sprint(y_ind), "O"+fmt.Sprint(y_ind+h_ind-1))
					consolidadoExcelPlanAnual.MergeCell(sheetName, "P"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1))
				}
				idx++
				if idx < len(indicadores) {
					reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "M"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1), stylecontentCL, stylecontentCLS)
				} else {
					reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "M"+fmt.Sprint(y_ind), "P"+fmt.Sprint(y_ind+h_ind-1), stylecontentCLD, stylecontentCLDS)
				}
				y_ind += h_ind
			}

			consolidadoExcelPlanAnual.MergeCell(sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1))
			consolidadoExcelPlanAnual.SetCellValue(sheetName, "Q"+fmt.Sprint(rowPos), datosComplementarios["Producto esperado"])
			reporteshelper.SombrearCeldas(consolidadoExcelPlanAnual, excelPlan, sheetName, "Q"+fmt.Sprint(rowPos), "Q"+fmt.Sprint(rowPos+MaxRowsXActivity-1), stylecontentC, stylecontentCS)

			rowPos += MaxRowsXActivity

			contadorGeneral = rowPos - 2

			consolidadoExcelPlanAnual.SetActiveSheet(indexPlan)
		}
		arregloPlanAnual = nil
		consolidadoExcelPlanAnual.RemoveRow(sheetName, 1)
	}

	if len(planesFilter) <= 0 {
		return nil, errors.New("error de longitud")
	}

	styletitle, _ := consolidadoExcelPlanAnual.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 18, Color: ColorNegro},
		Border: []excelize.Border{
			{Type: "right", Color: ColorBlanco, Style: 1},
			{Type: "left", Color: ColorBlanco, Style: 1},
			{Type: "top", Color: ColorBlanco, Style: 1},
			{Type: "bottom", Color: ColorBlanco, Style: 1},
		},
	})

	consolidadoExcelPlanAnual.InsertRows("REPORTE GENERAL", 1, 3)
	consolidadoExcelPlanAnual.MergeCell("REPORTE GENERAL", "C2", "P6")
	consolidadoExcelPlanAnual.SetCellStyle("REPORTE GENERAL", "C2", "P6", styletitle)
	if periodo[0] != nil {
		consolidadoExcelPlanAnual.SetCellValue("REPORTE GENERAL", "C2", "Plan de Acción Anual "+periodo[0]["Nombre"].(string)+"\nUniversidad Distrital Franciso José de Caldas")
	} else {
		consolidadoExcelPlanAnual.SetCellValue("REPORTE GENERAL", "C2", "Plan de Acción Anual \nUniversidad Distrital Franciso José de Caldas")
	}

	if err := consolidadoExcelPlanAnual.AddPicture("REPORTE GENERAL", "B1", "static/img/UDEscudo2.png",
		&excelize.GraphicOptions{ScaleX: 0.1, ScaleY: 0.1, Positioning: "oneCell", OffsetX: 10}); err != nil {
		return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
	}

	consolidadoExcelPlanAnual.SetColWidth("REPORTE GENERAL", "A", "A", 2)
	buf, _ := consolidadoExcelPlanAnual.WriteToBuffer()
	strings.NewReader(buf.String())
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	dataSend := make(map[string]interface{})
	dataSend["generalData"] = arregloInfoReportes
	dataSend["excelB64"] = encoded
	return dataSend, nil
}

func ProcesarNecesidades(body map[string]interface{}, nombre string) (dataSend map[string]interface{}, outputError error) {
	defer func() {
		if err := recover(); err != nil {
			outputError = errors.New("error al ProcesarNecesidades status error : 500")
		}
	}()

	var respuesta map[string]interface{}
	var respuestaIdentificaciones map[string]interface{}
	var identificaciones []map[string]interface{}
	var planes []map[string]interface{}
	var recursos []map[string]interface{}
	var recursosGeneral []map[string]interface{}
	var rubros []map[string]interface{}
	var rubrosGeneral []map[string]interface{}
	var unidades_total []string
	var unidades_rubros_total []string
	var respuestaEstado map[string]interface{}
	var estado map[string]interface{}
	var respuestaTipo map[string]interface{}
	var tipo map[string]interface{}
	var arregloInfoReportes []map[string]interface{}
	docentesPregrado := make(map[string]interface{})
	docentesPosgrado := make(map[string]interface{})
	var docentesGeneral map[string]interface{}
	var arrDataDocentes []map[string]interface{}

	primaServicios := 0
	primaNavidad := 0
	primaVacaciones := 0
	bonificacion := 0
	interesesCesantias := 0
	cesantiasPublicas := 0
	cesantiasPrivadas := 0
	salud := 0
	pensionesPublicas := 0
	pensionesPrivadas := 0
	arl := 0
	caja := 0
	icbf := 0
	docentesPregrado["tco"] = 0
	docentesPregrado["mto"] = 0
	docentesPregrado["hch"] = 0
	docentesPregrado["hcp"] = 0
	docentesPregrado["valor"] = 0

	docentesPosgrado["hch"] = 0
	docentesPosgrado["hcp"] = 0
	docentesPosgrado["valor"] = 0

	necesidadesExcel := excelize.NewFile()

	stylecontent, _ := reporteshelper.EstiloExcelRotacion(necesidadesExcel, "justify", "center", "", 0, false)
	stylecontentS, _ := reporteshelper.EstiloExcelRotacion(necesidadesExcel, "justify", "center", ColorGrisClaro, 0, true)
	stylecontentM, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "top", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	stylecontentMS, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisClaro}},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "top", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	styletitles, _ := reporteshelper.EstiloExcel(necesidadesExcel, "center", "center", ColorGrisClaro, false)
	stylehead, _ := reporteshelper.EstiloExcel(necesidadesExcel, "center", "center", ColorRojo, true)
	stylecontentCL, _ := reporteshelper.EstiloExcelBordes(necesidadesExcel, "justify", "center", "", 4, false)
	stylecontentCML, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 4},
		},
	})
	stylecontentCMD, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisOscuro}},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	stylecontentCM, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisOscuro}},
		Border: []excelize.Border{
			{Type: "top", Color: ColorBlanco, Style: 1},
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	stylecontentC, _ := necesidadesExcel.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisOscuro}},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "top", Color: ColorBlanco, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	stylecontentCD, _ := necesidadesExcel.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisOscuro}},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	})
	stylecontentCLS, _ := reporteshelper.EstiloExcelBordes(necesidadesExcel, "justify", "center", ColorGrisClaro, 4, true)
	stylecontentCMLS, _ := necesidadesExcel.NewStyle(&excelize.Style{
		NumFmt:    183,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisClaro}},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 4},
		},
	})

	necesidadesExcel.NewSheet("Necesidades")
	necesidadesExcel.DeleteSheet("Sheet1")
	disable := false
	err := necesidadesExcel.SetSheetView("Necesidades", -1, &excelize.ViewOptions{ShowGridLines: &disable})
	if err != nil {
		return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
	}

	necesidadesExcel.MergeCell("Necesidades", "C1", "E1")

	necesidadesExcel.SetColWidth("Necesidades", "A", "A", 4)
	necesidadesExcel.SetColWidth("Necesidades", "B", "B", 26)
	necesidadesExcel.SetColWidth("Necesidades", "C", "C", 15)
	necesidadesExcel.SetColWidth("Necesidades", "C", "E", 15)
	necesidadesExcel.SetColWidth("Necesidades", "F", "F", 20)
	necesidadesExcel.SetColWidth("Necesidades", "G", "G", 35)
	necesidadesExcel.SetColWidth("Necesidades", "H", "I", 12)
	necesidadesExcel.SetColWidth("Necesidades", "J", "J", 35)

	necesidadesExcel.SetCellValue("Necesidades", "B1", "Código del rubro")
	necesidadesExcel.SetCellValue("Necesidades", "C1", "Nombre del rubro")
	necesidadesExcel.SetCellValue("Necesidades", "F1", "Valor")
	necesidadesExcel.SetCellValue("Necesidades", "G1", "Dependencias")
	necesidadesExcel.SetCellStyle("Necesidades", "B1", "G1", stylehead)

	contador := 2
	url := "http://" + beego.AppConfig.String("PlanesService") + "/plan?query=activo:true,tipo_plan_id:" + body["tipo_plan_id"].(string) + ",vigencia:" + body["vigencia"].(string) + ",estado_plan_id:" + body["estado_plan_id"].(string) + ",nombre:" + nombre
	if err := request.GetJson(url, &respuesta); err != nil {
		outputError = errors.New("error al procesar la peticion " + err.Error())
	}
	request.LimpiezaRespuestaRefactor(respuesta, &planes)

	for i := 0; i < len(planes); i++ {
		flag := true
		var docentes map[string]interface{}
		var aux map[string]interface{}
		var dependencia_nombre string
		dependencia := planes[i]["dependencia_id"]
		url2 := "http://" + beego.AppConfig.String("OikosService") + "/dependencia/" + dependencia.(string)
		if err := request.GetJson(url2, &aux); err != nil {
			outputError = errors.New("error al procesar la peticion " + err.Error())
		}
		dependencia_nombre = aux["Nombre"].(string)
		url3 := "http://" + beego.AppConfig.String("PlanesService") + "/identificacion?query=plan_id:" + planes[i]["_id"].(string) + ",activo:true"
		if err := request.GetJson(url3, &respuestaIdentificaciones); err != nil {
			outputError = errors.New("error al procesar la peticion " + err.Error())
		}
		request.LimpiezaRespuestaRefactor(respuestaIdentificaciones, &identificaciones)
		for i := 0; i < len(identificaciones); i++ {
			identificacion := identificaciones[i]
			nombre := strings.ToLower(identificacion["nombre"].(string))
			if strings.Contains(nombre, "recurso") {
				if identificacion["dato"] != nil {
					var dato map[string]interface{}
					var data_identi []map[string]interface{}
					dato_str := identificacion["dato"].(string)
					json.Unmarshal([]byte(dato_str), &dato)
					for key := range dato {
						element := dato[key].(map[string]interface{})
						if element["activo"] == true {
							data_identi = append(data_identi, element)
						}
					}
					recursos = data_identi
				}
			}
			if strings.Contains(nombre, "contratista") && flag {
				if identificacion["dato"] != nil && identificacion["dato"].(string) != "{}" {
					var dato map[string]interface{}
					var dato_contratistas []map[string]interface{}
					dato_str := identificacion["dato"].(string)
					json.Unmarshal([]byte(dato_str), &dato)
					element := dato["0"].(map[string]interface{})
					if element["activo"] == true {
						dato_contratistas = append(dato_contratistas, element)
						flag = false
					}
					rubros = dato_contratistas
				}
			} else if strings.Contains(nombre, "docente") {
				if identificacion["dato"] != nil && identificacion["dato"] != "{}" {
					dato := map[string]interface{}{}
					result := make(map[string]interface{})
					dato_str := identificacion["dato"].(string)

					/* Se tiene en cuenta la nueva estructura la info ahora está en identificacion-detalle,
					pero tambien se tiene en cuenta la estructura de indentificaciones viejas (else) */
					if strings.Contains(dato_str, "ids_detalle") {
						json.Unmarshal([]byte(dato_str), &dato)

						iddetail := ""
						iddetail = dato["ids_detalle"].(map[string]interface{})["rhf"].(string)
						result["rhf"] = identificacionNueva(iddetail)

						iddetail = dato["ids_detalle"].(map[string]interface{})["rhv_pre"].(string)
						result["rhv_pre"] = identificacionNueva(iddetail)

						iddetail = dato["ids_detalle"].(map[string]interface{})["rhv_pos"].(string)
						result["rhv_pos"] = identificacionNueva(iddetail)

						iddetail = dato["ids_detalle"].(map[string]interface{})["rubros"].(string)
						result["rubros"] = identificacionNueva(iddetail)
					} else {
						json.Unmarshal([]byte(dato_str), &dato)
						result["rhf"] = reporteshelper.IdentificacionAntigua(dato["rhf"].(string))
						result["rhv_pre"] = reporteshelper.IdentificacionAntigua(dato["rhv_pre"].(string))
						result["rhv_pos"] = reporteshelper.IdentificacionAntigua(dato["rhv_pos"].(string))
						if dato["rubros"] != nil {
							result["rubros"] = reporteshelper.IdentificacionAntigua(dato["rubros"].(string))
						}
					}
					docentes = result
				}
			}
		}

		for i := 0; i < len(recursos); i++ {
			var aux bool
			var aux1 []string
			if len(recursosGeneral) == 0 {
				recursosGeneral = append(recursosGeneral, recursos[i])

				var valorU []float64
				var auxValor float64
				if fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "int" {
					auxValor = recursos[i]["valor"].(float64)
				} else {
					aux2, err := reporteshelper.DeformatNumberFloat(recursos[i]["valor"])
					if err == nil {
						auxValor = aux2
					}
				}
				index := len(recursosGeneral) - 1
				valorU = append(valorU, auxValor)
				recursosGeneral[index]["valorU"] = valorU
				aux1 = append(aux1, dependencia_nombre)
				recursosGeneral[index]["unidades"] = aux1
				unidades_total = append(unidades_total, dependencia_nombre)
			} else {
				for j := 0; j < len(recursosGeneral); j++ {
					if recursosGeneral[j]["codigo"] == recursos[i]["codigo"] {
						if recursosGeneral[j]["unidades"] != nil {
							recursosGeneral[j]["unidades"] = append(recursosGeneral[j]["unidades"].([]string), dependencia_nombre)
						}
						flag1 := false
						for k := 0; k < len(unidades_total); k++ {
							if unidades_total[k] == dependencia_nombre {
								flag1 = true
							}
						}
						if !flag1 {
							unidades_total = append(unidades_total, dependencia_nombre)
						}

						if recursosGeneral[j]["valor"] != nil {
							var auxValor float64
							var auxValor2 float64
							if fmt.Sprint(reflect.TypeOf(recursosGeneral[j]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursosGeneral[j]["valor"])) == "float64" {
								auxValor = recursosGeneral[j]["valor"].(float64)
							} else {
								aux1, err := reporteshelper.DeformatNumberFloat(recursosGeneral[j]["valor"])
								if err == nil {
									auxValor = aux1
								}
							}

							if fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "float64" {
								auxValor2 = recursos[i]["valor"].(float64)
							} else {
								aux2, err := reporteshelper.DeformatNumberFloat(recursos[i]["valor"])
								if err == nil {
									auxValor2 = aux2
								}
							}

							recursosGeneral[j]["valor"] = auxValor + auxValor2
							if recursosGeneral[j]["valorU"] == nil {
								var valorU []float64
								valorU = append(valorU, auxValor2)
								recursosGeneral[j]["valorU"] = valorU
							} else {
								valorU := recursosGeneral[j]["valorU"].([]float64)
								recursosGeneral[j]["valorU"] = append(valorU, auxValor2)
							}
						} else {
							var valorU []float64
							var auxValor float64
							if fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "float64" {
								auxValor := recursos[i]["valor"].(float64)
								valorU = append(valorU, auxValor)
							} else {
								auxValor, err := reporteshelper.DeformatNumberFloat(recursos[i]["valor"])
								if err == nil {
									valorU = append(valorU, auxValor)
								}
							}
							recursosGeneral[j]["valor"] = auxValor
							recursosGeneral[j]["valorU"] = valorU
						}
						aux = true
						break
					} else {
						aux = false
					}
				}
				if !aux {
					flag := false
					var valorU []float64
					var auxValor float64
					if fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "flaot64" {
						auxValor = recursos[i]["valor"].(float64)
					} else {
						aux2, err := reporteshelper.DeformatNumberFloat(recursos[i]["valor"])
						if err == nil {
							auxValor = aux2
						}
					}
					recursosGeneral = append(recursosGeneral, recursos[i])
					index := len(recursosGeneral) - 1
					valorU = append(valorU, auxValor)
					recursosGeneral[index]["valorU"] = valorU
					aux1 = append(aux1, dependencia_nombre)
					recursosGeneral[index]["unidades"] = aux1
					for k := 0; k < len(unidades_total); k++ {
						if unidades_total[k] == dependencia_nombre {
							flag = true
						}
					}
					if !flag {
						unidades_total = append(unidades_total, dependencia_nombre)
					}
				}
			}
		}

		for i := 0; i < len(rubros); i++ {
			var aux bool
			var aux1 []string
			if len(rubrosGeneral) == 0 {
				var auxValor2 float64
				var valorU []float64
				if _, ok := rubros[i]["totalInc"].(float64); ok {
					rubros[i]["totalInc"] = fmt.Sprintf("%f", rubros[i]["totalInc"])
				}
				if rubros[i]["totalInc"] != nil {
					auxValor2, _ = strconv.ParseFloat(rubros[i]["totalInc"].(string), 64)
				} else {
					auxValor2 = 0.0
				}

				rubrosGeneral = append(rubrosGeneral, rubros[i])
				index := len(rubrosGeneral) - 1
				valorU = append(valorU, auxValor2)
				rubrosGeneral[index]["valorU"] = valorU
				aux1 = append(aux1, dependencia_nombre)
				rubrosGeneral[index]["unidades"] = aux1
				unidades_rubros_total = append(unidades_rubros_total, dependencia_nombre)
			} else {
				for j := 0; j < len(rubrosGeneral); j++ {
					if rubrosGeneral[j]["rubro"] == rubros[i]["rubro"] {
						flag := false
						for k := 0; k < len(rubrosGeneral[j]["unidades"].([]string)); k++ {
							aux2 := rubrosGeneral[j]["unidades"].([]string)
							if aux2[k] == dependencia_nombre {
								flag = true
							}
						}
						if !flag {
							rubrosGeneral[j]["unidades"] = append(rubrosGeneral[j]["unidades"].([]string), dependencia_nombre)
						}
						flag2 := false
						for k := 0; k < len(unidades_total); k++ {
							if unidades_total[k] == dependencia_nombre {
								flag2 = true
							}
						}
						if !flag2 {
							unidades_total = append(unidades_total, dependencia_nombre)
						}
						flag1 := false
						for k := 0; k < len(unidades_rubros_total); k++ {
							if unidades_rubros_total[k] == dependencia_nombre {
								flag1 = true
							}
						}
						if !flag1 {
							unidades_rubros_total = append(unidades_rubros_total, dependencia_nombre)
						}
						if rubrosGeneral[j]["totalInc"] != nil {
							var auxValor float64
							var auxValor2 float64
							if _, ok := rubrosGeneral[j]["totalInc"].(float64); ok {
								rubrosGeneral[j]["totalInc"] = fmt.Sprintf("%f", rubrosGeneral[j]["totalInc"])
							}
							auxValor, _ = strconv.ParseFloat(rubrosGeneral[j]["totalInc"].(string), 64)
							auxValor2, _ = strconv.ParseFloat(rubros[i]["totalInc"].(string), 64)

							if rubrosGeneral[j]["valorU"] == nil {
								var valorU []float64
								valorU = append(valorU, auxValor2)
								rubrosGeneral[j]["valorU"] = valorU
							} else {
								rubrosGeneral[j]["valorU"] = append(rubrosGeneral[j]["valorU"].([]float64), auxValor2)
							}

							rubrosGeneral[j]["totalInc"] = auxValor + auxValor2
						} else {
							rubrosGeneral[j]["valorU"] = append(rubrosGeneral[j]["valorU"].([]float64), 0.0)
							rubrosGeneral[j]["totalInc"] = rubros[i]["totalInc"]
						}
						aux = true
						break
					} else {
						aux = false
					}
				}
				if !aux {
					flag := false
					var valorU []float64
					var auxValor float64
					// ? puede haber recursos[] sin datos
					if len(recursos) > 0 {
						if fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursos[i]["valor"])) == "float64" {
							auxValor = recursos[i]["valor"].(float64)
						} else {
							aux2, err := reporteshelper.DeformatNumberFloat(recursos[i]["valor"])
							if err == nil {
								auxValor = aux2
							}
						}
						rubrosGeneral = append(rubrosGeneral, recursos[i])

						index := len(rubrosGeneral) - 1
						valorU = append(valorU, auxValor)
						rubrosGeneral[index]["valorU"] = valorU
						aux1 = append(aux1, dependencia_nombre)
						rubrosGeneral[index]["unidades"] = aux1
						for k := 0; k < len(unidades_total); k++ {
							if unidades_total[k] == dependencia_nombre {
								flag = true
							}
						}
					}
					if !flag {
						unidades_total = append(unidades_total, dependencia_nombre)
					}
				}
				if !aux {
					flag := false
					var valorU []float64
					var auxValor2 float64
					if rubros[i]["totalInc"] != nil {
						if _, ok := rubros[i]["totalInc"].(float64); ok {
							rubros[i]["totalInc"] = fmt.Sprintf("%f", rubros[i]["totalInc"])
						}
						auxValor2, _ = strconv.ParseFloat(rubros[i]["totalInc"].(string), 64)
						valorU = append(valorU, auxValor2)
					} else {
						valorU = append(valorU, 0.0)
					}

					rubrosGeneral = append(rubrosGeneral, rubros[i])
					index := len(rubrosGeneral) - 1
					rubrosGeneral[index]["valorU"] = valorU
					aux1 = append(aux1, dependencia_nombre)
					rubrosGeneral[index]["unidades"] = aux1
					for k := 0; k < len(unidades_rubros_total); k++ {
						if unidades_rubros_total[k] == dependencia_nombre {
							flag = true
						}
					}
					if !flag {
						unidades_rubros_total = append(unidades_rubros_total, dependencia_nombre)
					}
				}
			}
		}

		if len(docentes) > 0 {
			docentesGeneral = reporteshelper.GetTotalDocentes(docentes)
			primaServicios = primaServicios + docentesGeneral["primaServicios"].(int)
			primaNavidad = primaNavidad + docentesGeneral["primaNavidad"].(int)
			primaVacaciones = primaVacaciones + docentesGeneral["primaVacaciones"].(int)
			bonificacion = bonificacion + docentesGeneral["bonificacion"].(int)
			interesesCesantias = interesesCesantias + docentesGeneral["interesesCesantias"].(int)
			cesantiasPublicas = cesantiasPublicas + docentesGeneral["cesantiasPublicas"].(int)
			cesantiasPrivadas = cesantiasPrivadas + docentesGeneral["cesantiasPrivadas"].(int)
			salud = salud + docentesGeneral["salud"].(int)
			pensionesPublicas = pensionesPublicas + docentesGeneral["pensionesPublicas"].(int)
			pensionesPrivadas = pensionesPrivadas + docentesGeneral["pensionesPrivadas"].(int)
			arl = arl + docentesGeneral["arl"].(int)
			caja = caja + docentesGeneral["caja"].(int)
			icbf = icbf + docentesGeneral["icbf"].(int)
			arrDataDocentes = append(arrDataDocentes, getDataDocentes(docentes, planes[i]["dependencia_id"].(string)))
		}

		if docentes["rubros"] != nil {
			var aux bool
			var respuestaRubro map[string]interface{}
			rubros := docentes["rubros"].([]map[string]interface{})
			for i := 0; i < len(rubros); i++ {
				if rubros[i]["rubro"] != "" {
					for j := 0; j < len(recursosGeneral); j++ {
						if recursosGeneral[j]["codigo"] == rubros[i]["rubro"] {
							aux = true
							categoria := strings.ToLower(rubros[i]["categoria"].(string))
							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "servicio") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaServicios, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "navidad") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaNavidad, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "vacaciones") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaVacaciones, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "bonificacion") || strings.Contains(categoria, "bonificación") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(bonificacion, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "interes") && strings.Contains(categoria, "cesantía") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(interesesCesantias, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "público") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(cesantiasPublicas, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "privado") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(cesantiasPrivadas, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "salud") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(salud, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "pension") && strings.Contains(categoria, "público") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(pensionesPublicas, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "pension") && strings.Contains(categoria, "privado") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(pensionesPrivadas, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "arl") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(arl, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "ccf") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(caja, recursosGeneral[j]["valor"])
							}
							if strings.Contains(categoria, "icbf") {
								recursosGeneral[j]["valor"] = reporteshelper.ActualizarRecursosGeneral(icbf, recursosGeneral[j]["valor"])
							}
							break
						} else {
							aux = false
						}
					}
					if !aux && rubros[i]["rubro"] != nil {
						rubro := make(map[string]interface{})
						url4 := "http://" + beego.AppConfig.String("PlanCuentasService") + "/arbol_rubro/" + rubros[i]["rubro"].(string)
						if err := request.GetJson(url4, &respuestaRubro); err != nil {
							outputError = errors.New("error al procesar la peticion " + err.Error())
						}
						if respuestaRubro["Body"] == nil {
							continue
						}
						aux := respuestaRubro["Body"].(map[string]interface{})
						rubro["codigo"] = aux["Codigo"]
						rubro["nombre"] = aux["Nombre"]
						rubro["categoria"] = rubros[i]["categoria"]

						if rubro["categoria"] != nil {
							categoria := strings.ToLower(rubro["categoria"].(string))

							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "servicio") {
								rubro["valor"] = reporteshelper.ActualizarRubro(primaServicios, rubro["valor"])
							}
							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "navidad") {
								rubro["valor"] = reporteshelper.ActualizarRubro(primaNavidad, rubro["valor"])
							}
							if strings.Contains(categoria, "prima") && strings.Contains(categoria, "vacaciones") {
								rubro["valor"] = reporteshelper.ActualizarRubro(primaVacaciones, rubro["valor"])
							}
							if strings.Contains(categoria, "bonificacion") || strings.Contains(categoria, "bonificación") {
								rubro["valor"] = reporteshelper.ActualizarRubro(bonificacion, rubro["valor"])
							}
							if strings.Contains(categoria, "interes") && strings.Contains(categoria, "cesantía") {
								rubro["valor"] = reporteshelper.ActualizarRubro(interesesCesantias, rubro["valor"])
							}
							if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "público") {
								rubro["valor"] = reporteshelper.ActualizarRubro(cesantiasPublicas, rubro["valor"])
							}
							if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "privado") {
								rubro["valor"] = reporteshelper.ActualizarRubro(cesantiasPrivadas, rubro["valor"])
							}
							if strings.Contains(categoria, "salud") {
								rubro["valor"] = reporteshelper.ActualizarRubro(salud, rubro["valor"])
							}
							if strings.Contains(categoria, "pension") && strings.Contains(categoria, "público") {
								rubro["valor"] = reporteshelper.ActualizarRubro(pensionesPublicas, rubro["valor"])
							}
							if strings.Contains(categoria, "pension") && strings.Contains(categoria, "privado") {
								rubro["valor"] = reporteshelper.ActualizarRubro(pensionesPrivadas, rubro["valor"])
							}
							if strings.Contains(categoria, "arl") {
								rubro["valor"] = reporteshelper.ActualizarRubro(arl, rubro["valor"])
							}
							if strings.Contains(categoria, "ccf") {
								rubro["valor"] = reporteshelper.ActualizarRubro(caja, rubro["valor"])
							}
							if strings.Contains(categoria, "icbf") {
								rubro["valor"] = reporteshelper.ActualizarRubro(icbf, rubro["valor"])
							}
						}
						recursosGeneral = append(recursosGeneral, rubro)
					}
				}
			}
		}
	}

	for i := 0; i < len(recursosGeneral); i++ {
		if recursosGeneral[i]["categoria"] != nil {
			categoria := strings.ToLower(recursosGeneral[i]["categoria"].(string))
			if strings.Contains(categoria, "prima") && strings.Contains(categoria, "servicio") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaServicios, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "prima") && strings.Contains(categoria, "navidad") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaNavidad, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "prima") && strings.Contains(categoria, "vacaciones") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(primaVacaciones, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "bonificacion") || strings.Contains(categoria, "bonificación") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(bonificacion, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "interes") && strings.Contains(categoria, "cesantía") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(interesesCesantias, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "público") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(cesantiasPublicas, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "cesantía") && strings.Contains(categoria, "privado") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(cesantiasPrivadas, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "salud") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(salud, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "pension") && strings.Contains(categoria, "público") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(pensionesPublicas, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "pension") && strings.Contains(categoria, "privado") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(pensionesPrivadas, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "arl") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(arl, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "ccf") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(caja, recursosGeneral[i]["valor"])
			}
			if strings.Contains(categoria, "icbf") {
				recursosGeneral[i]["valor"] = reporteshelper.ActualizarRecursosGeneral(icbf, recursosGeneral[i]["valor"])
			}
		}
	}

	if len(arrDataDocentes) < 7 {
		ingenieria := true
		ciencias := true
		ambiente := true
		tecnologica := true
		ASAB := true
		ILUD := true
		matematicas := true
		for _, data := range arrDataDocentes {
			facultad := strings.ToLower(data["nombreFacultad"].(string))
			if strings.Contains(facultad, "ingenieria") {
				ingenieria = false
			}
			if strings.Contains(facultad, "ciencias") {
				ciencias = false
			}
			if strings.Contains(facultad, "medio ambiente") {
				ambiente = false
			}
			if strings.Contains(facultad, "tecnologica") {
				tecnologica = false
			}
			if strings.Contains(facultad, "asab") {
				ASAB = false
			}
			if strings.Contains(facultad, "ilud") {
				ILUD = false
			}
			if strings.Contains(facultad, "matematicas") {
				matematicas = false
			}
		}
		if ingenieria {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD DE INGENIERIA", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if ciencias {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD DE CIENCIAS Y EDUCACION", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if ambiente {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD DE MEDIO AMBIENTE", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if tecnologica {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD TECNOLOGICA", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if ASAB {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD DE ARTES - ASAB", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if ILUD {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "INSTITUTO DE LENGUAS - ILUD", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
		if matematicas {
			vacio := map[string]interface{}{"hch": 0, "hchPos": 0, "hcp": 0, "hcpPos": 0, "mto": 0, "nombreFacultad": "FACULTAD DE CIENCIAS MATEMATICAS Y NATURALES", "tco": 0, "valorPos": 0, "valorPre": 0}
			arrDataDocentes = append(arrDataDocentes, vacio)
		}
	}

	//Completado de tablas
	idActividad := 0
	for i := 0; i < len(recursosGeneral); i++ {
		necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), recursosGeneral[i]["codigo"])
		if recursosGeneral[i]["Nombre"] != nil {
			necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "E"+fmt.Sprint(contador))
			necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), recursosGeneral[i]["Nombre"])

			if recursosGeneral[i]["unidades"] != nil {
				unidades := recursosGeneral[i]["unidades"].([]string)
				valores := recursosGeneral[i]["valorU"].([]float64)
				necesidadesExcel.MergeCell("Necesidades", "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+len(unidades)))
				necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "C"+fmt.Sprint(contador+len(unidades)))
				reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador+len(unidades)), stylecontent, stylecontentS)
				for j := 0; j < len(unidades); j++ {
					necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador+j), unidades[j])
					necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador+j), valores[j])
					reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "F"+fmt.Sprint(contador+j), "F"+fmt.Sprint(contador+j), stylecontentCML, stylecontentCMLS)
					reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "G"+fmt.Sprint(contador+j), "G"+fmt.Sprint(contador+j), stylecontentCL, stylecontentCLS)
				}
				contador = contador + len(unidades)
			}
		} else {
			necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "E"+fmt.Sprint(contador))
			necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), recursosGeneral[i]["nombre"])
			if fmt.Sprint(reflect.TypeOf(recursosGeneral[i]["valor"])) == "int" {
				necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), recursosGeneral[i]["valor"])
				necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
			} else {
				auxValor, err := reporteshelper.DeformatNumberInt(recursosGeneral[i]["valor"])
				if err == nil {
					necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), auxValor)
					necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
				}
			}
		}
		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "F"+fmt.Sprint(contador), "F"+fmt.Sprint(contador), stylecontentCMD, stylecontentCMD)
		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "G"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentCD, stylecontentCD)

		if fmt.Sprint(reflect.TypeOf(recursosGeneral[i]["valor"])) == "int" || fmt.Sprint(reflect.TypeOf(recursosGeneral[i]["valor"])) == "float64" {
			necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), recursosGeneral[i]["valor"])
			necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
		} else {
			auxValor, err := reporteshelper.DeformatNumberInt(recursosGeneral[i]["valor"])
			if err == nil {
				necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), auxValor)
				necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
			}
		}
		idActividad = i
		contador++
	}

	for i := 0; i < len(rubrosGeneral); i++ {
		if rubrosGeneral[i]["rubro"] != nil {
			idActividad++
			necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), rubrosGeneral[i]["rubro"])
			if rubrosGeneral[i]["rubroNombre"] != nil {
				necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "E"+fmt.Sprint(contador))
				necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), rubrosGeneral[i]["rubroNombre"])
				if rubrosGeneral[i]["unidades"] != nil {

					unidades := rubrosGeneral[i]["unidades"].([]string)
					valores := rubrosGeneral[i]["valorU"].([]float64)

					// TODO: Revisar este machete, hay menos valores que unidades, se repite la primera unidad, el problema se presenta más arriba.
					if unidades[0] == unidades[1] && (len(unidades)-1) == len(valores) {
						unidades = unidades[1:]
					}
					// --- end of machete

					necesidadesExcel.MergeCell("Necesidades", "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+len(unidades)))
					necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "C"+fmt.Sprint(contador+len(unidades)))
					reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador+len(unidades)), stylecontent, stylecontentS)
					for j := 0; j < len(unidades); j++ {
						necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador+j), unidades[j])
						if j < len(valores) {
							necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador+j), valores[j])
						} else {
							necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador+j), 0.0)
						}
						reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "F"+fmt.Sprint(contador+j), "F"+fmt.Sprint(contador+j), stylecontentCML, stylecontentCMLS)
						reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "G"+fmt.Sprint(contador+j), "G"+fmt.Sprint(contador+j), stylecontentCL, stylecontentCLS)
					}

					if len(rubrosGeneral[i]["unidades"].([]string)) == 1 {
						necesidadesExcel.MergeCell("Necesidades", "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+len(unidades)+1))
						necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "C"+fmt.Sprint(contador+len(unidades)+1))
						reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador+len(unidades)+1), stylecontent, stylecontentS)
						contador = contador + len(unidades) + 1
					} else {
						contador = contador + len(unidades)
					}
				}
			} else {
				necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), rubrosGeneral[i]["rubroNombre"])
				necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "E"+fmt.Sprint(contador))
			}

			if fmt.Sprint(reflect.TypeOf(rubrosGeneral[i]["totalInc"])) == "float64" {
				necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), rubrosGeneral[i]["totalInc"])
				necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
			} else {
				if rubrosGeneral[i]["totalInc"] != nil {
					strValor := strings.TrimLeft(rubrosGeneral[i]["totalInc"].(string), "$")
					strValor = strings.ReplaceAll(strValor, ",", "")
					auxValor, err := strconv.ParseFloat(strValor, 64)
					if err == nil {
						necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), auxValor)
						necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Total")
					}
				}
			}

			reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "B"+fmt.Sprint(contador), "E"+fmt.Sprint(contador), stylecontent, stylecontentS)
			reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "F"+fmt.Sprint(contador), "F"+fmt.Sprint(contador), stylecontentCMD, stylecontentCMD)
			reporteshelper.SombrearCeldas(necesidadesExcel, idActividad, "Necesidades", "G"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentCD, stylecontentCD)

			contador++
		}
	}

	url5 := "http://" + beego.AppConfig.String("PlanesService") + "/estado-plan/" + planes[0]["estado_plan_id"].(string)
	if err := request.GetJson(url5, &respuestaEstado); err != nil {
		outputError = errors.New("error al procesar la peticion " + err.Error())
	}
	request.LimpiezaRespuestaRefactor(respuestaEstado, &estado)

	url6 := "http://" + beego.AppConfig.String("PlanesService") + "/tipo-plan/" + planes[0]["tipo_plan_id"].(string)
	if err := request.GetJson(url6, &respuestaTipo); err != nil {
		outputError = errors.New("error al procesar la peticion " + err.Error())
	}
	request.LimpiezaRespuestaRefactor(respuestaTipo, &tipo)

	contador++
	necesidadesExcel.MergeCell("Necesidades", "B"+fmt.Sprint(contador), "J"+fmt.Sprint(contador))

	necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), "Docentes por tipo de vinculación:")
	necesidadesExcel.SetCellStyle("Necesidades", "B"+fmt.Sprint(contador), "J"+fmt.Sprint(contador), styletitles)

	contador++
	necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), "Facultad")
	necesidadesExcel.MergeCell("Necesidades", "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1))
	necesidadesExcel.MergeCell("Necesidades", "C"+fmt.Sprint(contador), "G"+fmt.Sprint(contador))

	necesidadesExcel.MergeCell("Necesidades", "H"+fmt.Sprint(contador), "J"+fmt.Sprint(contador))

	necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), "Pregrado")
	necesidadesExcel.SetCellValue("Necesidades", "H"+fmt.Sprint(contador), "Posgrado")

	necesidadesExcel.SetCellStyle("Necesidades", "B"+fmt.Sprint(contador), "J"+fmt.Sprint(contador), stylehead)
	necesidadesExcel.SetCellStyle("Necesidades", "B"+fmt.Sprint(contador), "B"+fmt.Sprint(contador+1), stylehead)

	contador++
	necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), "TCO")
	necesidadesExcel.SetCellValue("Necesidades", "D"+fmt.Sprint(contador), "MTO")
	necesidadesExcel.SetCellValue("Necesidades", "E"+fmt.Sprint(contador), "HCH")
	necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), "HCP")
	necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), "Valor")
	necesidadesExcel.SetCellValue("Necesidades", "H"+fmt.Sprint(contador), "HCH")
	necesidadesExcel.SetCellValue("Necesidades", "I"+fmt.Sprint(contador), "HCP")
	necesidadesExcel.SetCellValue("Necesidades", "J"+fmt.Sprint(contador), "Valor")
	necesidadesExcel.SetCellStyle("Necesidades", "C"+fmt.Sprint(contador), "J"+fmt.Sprint(contador), styletitles)

	contador++

	tco := 0
	mto := 0
	hch := 0
	hcp := 0
	valorPre := 0.0
	hchPos := 0
	hcpPos := 0
	valorPos := 0.0

	for i := 0; i < len(arrDataDocentes); i++ {
		necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), arrDataDocentes[i]["nombreFacultad"])
		necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), arrDataDocentes[i]["tco"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["tco"])) == "int" {
			tco += arrDataDocentes[i]["tco"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["tco"].(string))
			tco += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "D"+fmt.Sprint(contador), arrDataDocentes[i]["mto"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["mto"])) == "int" {
			mto += arrDataDocentes[i]["mto"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["mto"].(string))
			mto += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "E"+fmt.Sprint(contador), arrDataDocentes[i]["hch"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["hch"])) == "int" {
			hch += arrDataDocentes[i]["hch"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["hch"].(string))
			hch += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), arrDataDocentes[i]["hcp"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["hcp"])) == "int" {
			hcp += arrDataDocentes[i]["hcp"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["hcp"].(string))
			hcp += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), arrDataDocentes[i]["valorPre"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["valorPre"])) == "int" || fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["valorPre"])) == "float64" {
			valorPre += float64(arrDataDocentes[i]["valorPre"].(int))
		} else {
			aux2, _ := strconv.ParseFloat(arrDataDocentes[i]["valorPre"].(string), 64)
			valorPre += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "H"+fmt.Sprint(contador), arrDataDocentes[i]["hchPos"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["hchPos"])) == "int" {
			hchPos += arrDataDocentes[i]["hchPos"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["hchPos"].(string))
			hchPos += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "I"+fmt.Sprint(contador), arrDataDocentes[i]["hcpPos"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["hcpPos"])) == "int" {
			hcpPos += arrDataDocentes[i]["hcpPos"].(int)
		} else {
			aux2, _ := strconv.Atoi(arrDataDocentes[i]["hcpPos"].(string))
			hcpPos += aux2
		}

		necesidadesExcel.SetCellValue("Necesidades", "J"+fmt.Sprint(contador), arrDataDocentes[i]["valorPos"])
		if fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["valorPos"])) == "int" || fmt.Sprint(reflect.TypeOf(arrDataDocentes[i]["valorPos"])) == "float64" {
			valorPos += float64(arrDataDocentes[i]["valorPos"].(int))
		} else {
			aux2, _ := strconv.ParseFloat(arrDataDocentes[i]["valorPos"].(string), 64)
			valorPos += aux2
		}

		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador), stylecontent, stylecontentS)
		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "G"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentM, stylecontentMS)
		reporteshelper.SombrearCeldas(necesidadesExcel, i, "Necesidades", "J"+fmt.Sprint(contador), "J"+fmt.Sprint(contador), stylecontentM, stylecontentMS)
		contador++
	}

	necesidadesExcel.SetCellValue("Necesidades", "B"+fmt.Sprint(contador), "Total")
	necesidadesExcel.SetCellValue("Necesidades", "C"+fmt.Sprint(contador), tco)
	necesidadesExcel.SetCellValue("Necesidades", "D"+fmt.Sprint(contador), mto)
	necesidadesExcel.SetCellValue("Necesidades", "E"+fmt.Sprint(contador), hch)
	necesidadesExcel.SetCellValue("Necesidades", "F"+fmt.Sprint(contador), hcp)
	necesidadesExcel.SetCellValue("Necesidades", "G"+fmt.Sprint(contador), valorPre)
	necesidadesExcel.SetCellValue("Necesidades", "H"+fmt.Sprint(contador), hchPos)
	necesidadesExcel.SetCellValue("Necesidades", "I"+fmt.Sprint(contador), hcpPos)
	necesidadesExcel.SetCellValue("Necesidades", "J"+fmt.Sprint(contador), valorPos)

	reporteshelper.SombrearCeldas(necesidadesExcel, 0, "Necesidades", "B"+fmt.Sprint(contador), "I"+fmt.Sprint(contador), stylecontentC, stylecontentC)
	reporteshelper.SombrearCeldas(necesidadesExcel, 0, "Necesidades", "G"+fmt.Sprint(contador), "G"+fmt.Sprint(contador), stylecontentCM, stylecontentCM)
	reporteshelper.SombrearCeldas(necesidadesExcel, 0, "Necesidades", "J"+fmt.Sprint(contador), "J"+fmt.Sprint(contador), stylecontentCM, stylecontentCM)

	styletitle, _ := necesidadesExcel.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 18, Color: ColorNegro},
		Border: []excelize.Border{
			{Type: "right", Color: ColorBlanco, Style: 1},
			{Type: "left", Color: ColorBlanco, Style: 1},
			{Type: "top", Color: ColorBlanco, Style: 1},
			{Type: "bottom", Color: ColorBlanco, Style: 1},
		},
	})
	necesidadesExcel.InsertRows("Necesidades", 1, 7)
	necesidadesExcel.MergeCell("Necesidades", "C2", "G6")
	necesidadesExcel.SetCellStyle("Necesidades", "C2", "G6", styletitle)
	var resPeriodo map[string]interface{}
	var periodo []map[string]interface{}
	url7 := "http://" + beego.AppConfig.String("ParametrosService") + `/periodo?query=Id:` + body["vigencia"].(string)
	if err := request.GetJson(url7, &resPeriodo); err != nil {
		outputError = errors.New("error al procesar la peticion " + err.Error())
	}
	request.LimpiezaRespuestaRefactor(resPeriodo, &periodo)

	if periodo[0] != nil {
		necesidadesExcel.SetCellValue("Necesidades", "C2", "Consolidado proyeccción de necesidades "+periodo[0]["Nombre"].(string))
	} else {
		necesidadesExcel.SetCellValue("Necesidades", "C2", "Consolidado proyeccción de necesidades")
	}

	if err := necesidadesExcel.AddPicture("Necesidades", "B1", "static/img/UDEscudo2.png",
		&excelize.GraphicOptions{ScaleX: 0.1, ScaleY: 0.1, Positioning: "oneCell", OffsetX: 25}); err != nil {
		return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
	}

	contador = 1

	necesidadesExcel.NewSheet("Total Unidades")
	necesidadesExcel.MergeCell("Total Unidades", "A", "B")
	necesidadesExcel.MergeCell("Total Unidades", "A1", "A2")
	necesidadesExcel.SetColWidth("Total Unidades", "A", "B", 30)

	necesidadesExcel.MergeCell("Total Unidades", "A"+fmt.Sprint(contador), "B"+fmt.Sprint(contador))
	necesidadesExcel.MergeCell("Total Unidades", "A"+fmt.Sprint(contador), "A"+fmt.Sprint(contador+1))
	necesidadesExcel.SetCellValue("Total Unidades", "A"+fmt.Sprint(contador), "Total de unidades generadas:")
	necesidadesExcel.SetCellStyle("Total Unidades", "A"+fmt.Sprint(contador), "B"+fmt.Sprint(contador), styletitles)
	necesidadesExcel.SetCellStyle("Total Unidades", "A"+fmt.Sprint(contador+1), "B"+fmt.Sprint(contador+1), styletitles)
	contador++
	contador++
	necesidadesExcel.SetCellValue("Total Unidades", "A"+fmt.Sprint(contador), "Total de Unidades Generadas")
	necesidadesExcel.SetCellValue("Total Unidades", "B"+fmt.Sprint(contador), "Unidades Generadas")
	necesidadesExcel.SetCellStyle("Total Unidades", "A"+fmt.Sprint(contador), "B"+fmt.Sprint(contador), stylehead)
	contador++
	unid_total := ""
	for j := 0; j < len(unidades_total); j++ {
		infoReporte := make(map[string]interface{})
		infoReporte["vigencia"] = body["vigencia"].(string)
		infoReporte["estado_plan"] = estado["nombre"]
		infoReporte["tipo_plan"] = tipo["nombre"]
		infoReporte["nombre_unidad"] = unidades_total[j]
		unid_total = unid_total + unidades_total[j] + ", "
		arregloInfoReportes = append(arregloInfoReportes, infoReporte)
	}
	unid_total = strings.TrimRight(unid_total, ", ")
	necesidadesExcel.SetCellValue("Total Unidades", "A"+fmt.Sprint(contador), len(unidades_total))
	necesidadesExcel.SetCellValue("Total Unidades", "B"+fmt.Sprint(contador), unid_total)
	necesidadesExcel.SetCellStyle("Total Unidades", "A"+fmt.Sprint(contador), "B"+fmt.Sprint(contador), stylecontent)

	buf, _ := necesidadesExcel.WriteToBuffer()
	strings.NewReader(buf.String())
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataSend = make(map[string]interface{})
	dataSend["generalData"] = arregloInfoReportes
	dataSend["excelB64"] = encoded
	return dataSend, outputError
}

func ProcesarPlanAccionEvaluacion(body map[string]interface{}, nombre string) (map[string]interface{}, error) {
	var respuesta map[string]interface{}
	var versionesPlan []map[string]interface{}
	var arregloPlanAnual []map[string]interface{}
	var periodo []map[string]interface{}
	var unidadNombre string
	var evaluacion []map[string]interface{}
	var respuestaOikos []map[string]interface{}
	var resPeriodo map[string]interface{}
	var res map[string]interface{}
	var subgrupos []map[string]interface{}
	var seguimientos []map[string]interface{}
	var ultimoPlanAvalado map[string]interface{}

	var reformulaciones []map[string]interface{}
	var respuestaReformulacion []map[string]interface{}

	excelArmonizacion := make([]map[string]interface{}, 0)

	consolidadoExcelEvaluacion := excelize.NewFile()

	if body["unidad_id"].(string) != "" {
		idEstadoAval, errId := getIdEstadoAval()
		if errId != nil {
			return nil, errors.New("error al procesar la peticion " + errId.Error())
		}

		if err := request.GetJson("http://"+beego.AppConfig.String("FormulacionService")+"/formulacion/plan/versiones/"+body["unidad_id"].(string)+"/"+body["vigencia"].(string)+"/"+nombre, &respuesta); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		}
		request.LimpiezaRespuestaRefactor(respuesta, &versionesPlan)
		for posVersion := len(versionesPlan) - 1; posVersion > 0; posVersion-- {
			if versionesPlan[posVersion]["estado_plan_id"] == idEstadoAval {
				ultimoPlanAvalado = versionesPlan[posVersion]
			}
		}
		if ultimoPlanAvalado["reformulacion"].(bool) {
			for _, version := range versionesPlan {
				if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/reformulacion?query=plan_id:"+version["_id"].(string), &respuesta); err == nil {
					request.LimpiezaRespuestaRefactor(respuesta, &respuestaReformulacion)
					reformulaciones = append(reformulaciones, respuestaReformulacion...)
				}
			}
		}
		fmt.Println("\n\nReformulacionesss")
		formatdata.JsonPrint(reformulaciones)

		if auxSeguimientos, err := seguimientoService.ObtenerSeguimientos(ultimoPlanAvalado["_id"].(string)); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		} else {
			seguimientos = auxSeguimientos
		}

		trimestres := reporteshelper.GetPeriodosPlan(body["vigencia"].(string), versionesPlan[0]["_id"].(string))

		if len(versionesPlan) <= 0 {
			return nil, errors.New("error de longitud")
		}

		if err := request.GetJson("http://"+beego.AppConfig.String("OikosService")+"/dependencia?query=Id:"+body["unidad_id"].(string), &respuestaOikos); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		}
		unidadNombre = respuestaOikos[0]["Nombre"].(string)
		arregloPlanAnual = append(arregloPlanAnual, map[string]interface{}{"nombreUnidad": unidadNombre})

		if err := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+`/periodo?query=Id:`+body["vigencia"].(string), &resPeriodo); err != nil {
			return nil, errors.New("error al procesar la peticion " + err.Error())
		}
		request.LimpiezaRespuestaRefactor(resPeriodo, &periodo)

		trimestreDelAnio := "-1"
		trimestresConContenido := make([]map[string]interface{}, 0)

		for indexTrimestre := 0; indexTrimestre <= len(trimestres)-1; indexTrimestre++ {
			auxEvaluacion := evaluacionService.GetEvaluacionInterno(seguimientos[indexTrimestre]["plan_id"].(string), trimestres, indexTrimestre)
			if auxEvaluacion != nil && fmt.Sprintf("%v", auxEvaluacion) != "[]" {
				evaluacion = auxEvaluacion
				trimestreDelAnio = trimestres[indexTrimestre]["codigo_trimestre"].(string)
				trimestresConContenido = append(trimestresConContenido, trimestres[indexTrimestre])
			}
		}
		trimestreVacio := map[string]interface{}{"actividad": 0.0, "acumulado": 0.0, "denominador": 0.0, "meta": 0.0, "numerador": 0.0, "periodo": 0.0, "numeradorAcumulado": 0.0, "denominadorAcumulado": 0.0, "brecha": 0.0, "cualitativo": map[string]interface{}{"reporte": "", "dificultades": ""}}
		for _, actividad := range evaluacion {
			for auxTrim := len(trimestresConContenido) - 1; auxTrim >= 0; auxTrim-- {
				seguimiento, err := seguimientoService.ConsultarSeguimiento(seguimientos[auxTrim]["plan_id"].(string), actividad["numero"].(string), trimestresConContenido[auxTrim]["_id"].(string))
				if err == nil {
					switch auxTrim {
					case 0:
						actividad["trimestre1"].(map[string]interface{})["cualitativo"] = seguimiento["cualitativo"]
					case 1:
						actividad["trimestre2"].(map[string]interface{})["cualitativo"] = seguimiento["cualitativo"]
					case 2:
						actividad["trimestre3"].(map[string]interface{})["cualitativo"] = seguimiento["cualitativo"]
					case 3:
						actividad["trimestre4"].(map[string]interface{})["cualitativo"] = seguimiento["cualitativo"]
					}
				} else {
					logs.Error("Error --> ", err)
					return nil, errors.New(err.Error())
				}
			}
		}

		switch trimestreDelAnio {
		case "4":
			for _, actividad := range evaluacion {
				if len(actividad["trimestre4"].(map[string]interface{})) == 0 {
					actividad["trimestre4"] = trimestreVacio
				}
				if len(actividad["trimestre3"].(map[string]interface{})) == 0 {
					actividad["trimestre3"] = trimestreVacio
				}
				if len(actividad["trimestre2"].(map[string]interface{})) == 0 {
					actividad["trimestre2"] = trimestreVacio
				}
				if len(actividad["trimestre1"].(map[string]interface{})) == 0 {
					actividad["trimestre1"] = trimestreVacio
				}
			}
		case "3":
			for _, actividad := range evaluacion {
				actividad["trimestre4"] = trimestreVacio
				if len(actividad["trimestre3"].(map[string]interface{})) == 0 {
					actividad["trimestre3"] = trimestreVacio
				}
				if len(actividad["trimestre2"].(map[string]interface{})) == 0 {
					actividad["trimestre2"] = trimestreVacio
				}
				if len(actividad["trimestre1"].(map[string]interface{})) == 0 {
					actividad["trimestre1"] = trimestreVacio
				}
			}
		case "2":
			for _, actividad := range evaluacion {
				actividad["trimestre4"] = trimestreVacio
				actividad["trimestre3"] = trimestreVacio
				if len(actividad["trimestre2"].(map[string]interface{})) == 0 {
					actividad["trimestre2"] = trimestreVacio
				}
				if len(actividad["trimestre1"].(map[string]interface{})) == 0 {
					actividad["trimestre1"] = trimestreVacio
				}
			}
		case "1":
			for _, actividad := range evaluacion {
				actividad["trimestre4"] = trimestreVacio
				actividad["trimestre3"] = trimestreVacio
				actividad["trimestre2"] = trimestreVacio
				if len(actividad["trimestre1"].(map[string]interface{})) == 0 {
					actividad["trimestre1"] = trimestreVacio
				}
			}
		case "-1":
			return nil, errors.New("error de indexación")
		}

		if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo?query=padre:"+ultimoPlanAvalado["_id"].(string)+"&fields=nombre,_id,hijos,activo", &res); err == nil {
			request.LimpiezaRespuestaRefactor(res, &subgrupos)

			for i := 0; i < len(subgrupos); i++ {
				if strings.Contains(strings.ToLower(subgrupos[i]["nombre"].(string)), "actividad") {
					actividades, _ := reporteshelper.GetActividades(subgrupos[i]["_id"].(string))
					var arregloLineamietoPED []map[string]interface{}
					var arregloLineamietoPI []map[string]interface{}
					if len(actividades) == 1 {
						for index := range actividades {
							if val, ok := actividades[index]["index"].(float64); ok {
								actividades[index]["index"] = fmt.Sprintf("%v", int(val))
							}
						}
					} else {
						sort.SliceStable(actividades, func(i int, j int) bool {
							if _, ok := actividades[i]["index"].(float64); ok {
								actividades[i]["index"] = fmt.Sprintf("%v", int(actividades[i]["index"].(float64)))
							}
							if _, ok := actividades[j]["index"].(float64); ok {
								actividades[j]["index"] = fmt.Sprintf("%v", int(actividades[j]["index"].(float64)))
							}
							aux, _ := strconv.Atoi((actividades[i]["index"]).(string))
							aux1, _ := strconv.Atoi((actividades[j]["index"]).(string))
							return aux < aux1
						})
					}

					reporteshelper.LimpiarDetalles()
					for j := 0; j < len(actividades); j++ {
						arregloLineamietoPED = nil
						arregloLineamietoPI = nil
						actividad := actividades[j]
						actividadName := actividad["dato"].(string)
						index := actividad["index"].(string)
						datosArmonizacion := make(map[string]interface{})
						titulosArmonizacion := make(map[string]interface{})

						reporteshelper.Limpia()
						tree := reporteshelper.BuildTreeFa(subgrupos, index)
						treeDatos := tree[0]
						treeDatas := tree[1]
						treeArmo := tree[2]
						armonizacionTercer := treeArmo[0]
						var armonizacionTercerNivel interface{}
						var armonizacionTercerNivelPI interface{}

						if armonizacionTercer["armo"] != nil {
							armonizacionTercerNivel = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPED"]
							armonizacionTercerNivelPI = armonizacionTercer["armo"].(map[string]interface{})["armonizacionPI"]
						}

						for datoGeneral := 0; datoGeneral < len(treeDatos); datoGeneral++ {
							treeDato := treeDatos[datoGeneral]
							treeData := treeDatas[0]
							if treeDato["sub"] == "" {
								nombre := strings.ToLower(treeDato["nombre"].(string))
								if strings.Contains(nombre, "ponderación") || strings.Contains(nombre, "ponderacion") && strings.Contains(nombre, "actividad") {
									datosArmonizacion["Ponderación de la actividad"] = treeData[treeDato["id"].((string))]
								} else if strings.Contains(nombre, "período") || strings.Contains(nombre, "periodo") && strings.Contains(nombre, "ejecucion") || strings.Contains(nombre, "ejecución") {
									datosArmonizacion["Periodo de ejecución"] = treeData[treeDato["id"].(string)]
								} else if strings.Contains(nombre, "actividad") && strings.Contains(nombre, "general") {
									datosArmonizacion["Actividad general"] = treeData[treeDato["id"].(string)]
								} else if strings.Contains(nombre, "tarea") || strings.Contains(nombre, "actividades específicas") {
									datosArmonizacion["Tareas"] = treeData[treeDato["id"].(string)]
								} else {
									datosArmonizacion[treeDato["nombre"].(string)] = treeData[treeDato["id"].(string)]
								}
							}
						}
						var treeIndicador map[string]interface{}
						auxTree := tree[0]
						for i := 0; i < len(auxTree); i++ {
							subgrupo := auxTree[i]
							if strings.Contains(strings.ToLower(subgrupo["nombre"].(string)), "indicador") {
								treeIndicador = auxTree[i]
							}
						}

						subIndicador := treeIndicador["sub"].([]map[string]interface{})
						for ind := 0; ind < len(subIndicador); ind++ {
							subIndicadorRes := subIndicador[ind]
							treeData := treeDatas[0]
							dataIndicador := make(map[string]interface{})
							auxSubIndicador := subIndicadorRes["sub"].([]map[string]interface{})
							for subInd := 0; subInd < len(auxSubIndicador); subInd++ {
								if treeData[auxSubIndicador[subInd]["id"].(string)] == nil {
									treeData[auxSubIndicador[subInd]["id"].(string)] = ""
								}
								dataIndicador[auxSubIndicador[subInd]["nombre"].(string)] = treeData[auxSubIndicador[subInd]["id"].(string)]
							}
							titulosArmonizacion[subIndicadorRes["nombre"].(string)] = dataIndicador
						}

						datosArmonizacion["indicadores"] = titulosArmonizacion
						arregloLineamietoPED, _ = ArbolArmonizacionV2(armonizacionTercerNivel.(string))
						arregloLineamietoPI, _ = ArbolArmonizacionPIV2(armonizacionTercerNivelPI.(string))
						generalData := make(map[string]interface{})
						generalData["nombreActividad"] = actividadName
						generalData["numeroActividad"] = index
						generalData["datosArmonizacionPED"] = arregloLineamietoPED
						generalData["datosArmonizacionPI"] = arregloLineamietoPI
						generalData["datosComplementarios"] = datosArmonizacion
						excelArmonizacion = append(excelArmonizacion, generalData)
					}
					break
				}
			}
		} else {
			panic(err)
		}

		sheetName := "Evaluación"
		consolidadoExcelEvaluacion.NewSheet(sheetName)
		consolidadoExcelEvaluacion.DeleteSheet("Sheet1")

		disable := false
		err := consolidadoExcelEvaluacion.SetSheetView(sheetName, -1, &excelize.ViewOptions{ShowGridLines: &disable})
		if err != nil {
			return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
		}

		styleUnidad, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Vertical: "center"},
			Font:      &excelize.Font{Bold: true, Color: ColorNegro, Family: "Bahnschrift SemiBold SemiConden", Size: 20},
			Border: []excelize.Border{
				{Type: "bottom", Color: ColorNegro, Style: 2},
			},
		})
		styleTituloSB, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Font:      &excelize.Font{Bold: true, Color: ColorBlanco, Family: "Calibri", Size: 11},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorRojo}},
		})
		styleSombreadoSB, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Font:      &excelize.Font{Color: ColorNegro, Family: "Calibri", Size: 11},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorGrisClaro}},
		})
		styleNegrilla, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Font:      &excelize.Font{Color: ColorNegro, Family: "Calibri", Size: 12, Bold: true},
		})
		styleTitulo, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Font:      &excelize.Font{Bold: true, Color: ColorBlanco, Family: "Calibri", Size: 11},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorRojo}},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		styleContenido, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelEvaluacion, "justify", "center", "", 0, false)
		styleContenidoC, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelEvaluacion, "center", "center", "", 0, false)
		styleContenidoCI, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Font:      &excelize.Font{Color: ColorBlanco},
		})
		styleContenidoCIP, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    10,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Font:      &excelize.Font{Color: ColorBlanco},
		})
		styleContenidoCE, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    1,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		styleContenidoCD, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    4,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		styleSubTitles, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Font:      &excelize.Font{Bold: true},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}},
			Border: []excelize.Border{
				{Type: "right", Color: "000000", Style: 1},
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})
		styleContenidoCS, _ := reporteshelper.EstiloExcelRotacion(consolidadoExcelEvaluacion, "center", "center", ColorRosado, 0, true)
		styleContenidoCP, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    10,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		styleContenidoCPSR, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    10,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Font:      &excelize.Font{Bold: true, Color: ColorBlanco, Family: "Calibri", Size: 11},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorRojo}},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		styleContenidoCPS, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			NumFmt:    10,
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{ColorRosado}},
			Border: []excelize.Border{
				{Type: "right", Color: ColorNegro, Style: 1},
				{Type: "left", Color: ColorNegro, Style: 1},
				{Type: "top", Color: ColorNegro, Style: 1},
				{Type: "bottom", Color: ColorNegro, Style: 1},
			},
		})
		stylecontent, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "justify", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "right", Color: "000000", Style: 1},
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})
		stylecontentC, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "right", Color: "000000", Style: 1},
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})
		styleLineamiento, _ := consolidadoExcelEvaluacion.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal:   "center",
				Vertical:     "center",
				WrapText:     true,
				TextRotation: 90,
			},
			Border: []excelize.Border{
				{Type: "right", Color: "000000", Style: 1},
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})

		// Size
		consolidadoExcelEvaluacion.SetRowHeight(sheetName, 1, 12)
		consolidadoExcelEvaluacion.SetRowHeight(sheetName, 2, 27)
		consolidadoExcelEvaluacion.SetRowHeight(sheetName, 19, 31)
		consolidadoExcelEvaluacion.SetRowHeight(sheetName, 22, 27)

		consolidadoExcelEvaluacion.SetColWidth(sheetName, "A", "A", 3)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "B", "B", 19)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "C", "C", 13)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "D", "G", 35)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "E", "E", 16)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "H", "H", 4)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "I", "I", 8)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "J", "J", 13)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "K", "K", 42)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "L", "L", 16)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "M", "M", 21)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "N", "BG", 14)
		consolidadoExcelEvaluacion.SetColWidth(sheetName, "BJ", "BN", 3)
		// Merge
		consolidadoExcelEvaluacion.MergeCell(sheetName, "B4", "D4")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "B19", "E19")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "B21", "D21")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "E21", "G21")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "H21", "H22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "I21", "I22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "J21", "J22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "K21", "K22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "L21", "L22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "M21", "M22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "N21", "N22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "O21", "O22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "BL19", "BM19")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "BL21", "BL22")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "BM21", "BM22")

		consolidadoExcelEvaluacion.MergeCell(sheetName, "P21", "Z21")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "AA21", "AK21")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "AL21", "AV21")
		consolidadoExcelEvaluacion.MergeCell(sheetName, "AW21", "BG21")
		// Style
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B2", "T2", styleUnidad)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B4", "D4", styleTituloSB)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "E4", "E4", styleSombreadoSB)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B19", "B19", styleNegrilla)
		//  Estilos títulos
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B21", "BG21", styleTitulo)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "J22", "BG22", styleContenidoC)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B22", "G22", styleSubTitles)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "P22", "BG22", styleSubTitles)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Z22", "Z22", styleContenidoCS)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AK22", "AK22", styleContenidoCS)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AV22", "AV22", styleContenidoCS)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BG22", "BG22", styleContenidoCS)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "H22", "O22", styleTitulo)

		if periodo[0] != nil {
			consolidadoExcelEvaluacion.SetCellValue(sheetName, "B2", "Evaluación Plan de Acción "+periodo[0]["Nombre"].(string)+" - "+unidadNombre)
			consolidadoExcelEvaluacion.SetCellValue(sheetName, "B19", "Cumplimiento General Plan de Acción "+periodo[0]["Nombre"].(string)+" - "+unidadNombre)
		} else {
			consolidadoExcelEvaluacion.SetCellValue(sheetName, "B2", "Evaluación Plan de Acción - "+unidadNombre)
			consolidadoExcelEvaluacion.SetCellValue(sheetName, "B19", "Cumplimiento General Plan de Acción - "+unidadNombre)
		}

		ddRango1 := excelize.NewDataValidation(true)
		ddRango1.Sqref = "E4:E4"
		ddRango1.SetDropList([]string{"Trimestre I", "Trimestre II", "Trimestre III", "Trimestre IV"})

		if err := consolidadoExcelEvaluacion.AddDataValidation(sheetName, ddRango1); err != nil {
			return nil, errors.New("error del servicio ProcesarPlanAccionAnual" + err.Error())
		}

		// Titles
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "B4", "Seleccione el periodo:")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "B21", "Armonización PED")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "E21", "Armonización Plan Indicativo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "H21", "No.")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "I21", "Pond.")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "J21", "Periodo de ejecución")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "K21", "Actividad General")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "L21", "Indicador asociado")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "M21", "Fórmula del Indicador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "N21", "Meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "O21", "Tipo de Unidad")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "P21", "Trimestre I")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AA21", "Trimestre II")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AL21", "Trimestre III")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AW21", "Trimestre IV")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "B22", "Lineamiento")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "C22", "Meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "D22", "Estrategias")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "E22", "Ejes transformadores")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "F22", "Lineamientos de acción")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "G22", "Estrategias")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "P22", "Reporte de avance")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "Q22", "Dificultades")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "R22", "Numerador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "S22", "Denominador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "T22", "Indicador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "U22", "Numerador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "V22", "Denominador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "W22", "Indicador Acumulado")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "X22", "Cumplimiento por meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "Y22", "Brecha")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "Z22", "Cumplimiento por actividad")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AA22", "Reporte de avance")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AB22", "Dificultades")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AC22", "Numerador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AD22", "Denominador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AE22", "Indicador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AF22", "Numerador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AG22", "Denominador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AH22", "Indicador Acumulado")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AI22", "Cumplimiento por meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AJ22", "Brecha")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AK22", "Cumplimiento por actividad")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AL22", "Reporte de avance")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AM22", "Dificultades")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AN22", "Numerador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AO22", "Denominador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AP22", "Indicador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AQ22", "Numerador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AR22", "Denominador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AS22", "Indicador Acumulado")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AT22", "Cumplimiento por meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AU22", "Brecha")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AV22", "Cumplimiento por actividad")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AW22", "Reporte de avance")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AX22", "Dificultades")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AY22", "Numerador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AZ22", "Denominador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BA22", "Indicador del Periodo")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BB22", "Numerador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BC22", "Denominador Acumulador")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BD22", "Indicador Acumulado")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BE22", "Cumplimiento por meta")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BF22", "Brecha")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BG22", "Cumplimiento por actividad")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BL19", "Gráfica")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BL21", "No.")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BM21", "Cumplimiento")

		if ultimoPlanAvalado["reformulacion"].(bool) {
			consolidadoExcelEvaluacion.MergeCell(sheetName, "R4", "S5")
			consolidadoExcelEvaluacion.SetCellStyle(sheetName, "R4", "R4", styleTituloSB)
			consolidadoExcelEvaluacion.SetCellValue(sheetName, "R4", "Reformulaciones realizadas")
			for posRef := 0; posRef < len(reformulaciones); posRef++ {
				consolidadoExcelEvaluacion.MergeCell(sheetName, "R"+fmt.Sprint(6+posRef), "S"+fmt.Sprint(6+posRef))
				consolidadoExcelEvaluacion.SetCellStyle(sheetName, "R"+fmt.Sprint(6+posRef), "R"+fmt.Sprint(6+posRef), styleSombreadoSB)

				textoTrimestre := "Trimestre "
				switch string(reformulaciones[posRef]["periodo"].(string)[1]) {
				case "1":
					textoTrimestre += "I"
				case "2":
					textoTrimestre += "II"
				case "3":
					textoTrimestre += "III"
				case "4":
					textoTrimestre += "IV"
				}
				consolidadoExcelEvaluacion.SetCellValue(sheetName, "R"+fmt.Sprint(6+posRef), textoTrimestre)
			}
		}

		indice := 23
		indiceGraficos := 23
		// Agregar armonización
		contadorArmonizacion := 21
		for posArmonizacion := 0; posArmonizacion < len(excelArmonizacion); posArmonizacion++ {
			datosExcelArmonizacion := excelArmonizacion[posArmonizacion]
			armoPED := datosExcelArmonizacion["datosArmonizacionPED"].([]map[string]interface{})
			armoPI := datosExcelArmonizacion["datosArmonizacionPI"].([]map[string]interface{})
			datosComplementarios := datosExcelArmonizacion["datosComplementarios"].(map[string]interface{})
			indicadores := datosComplementarios["indicadores"].(map[string]interface{})
			numeroActividad := datosExcelArmonizacion["numeroActividad"]

			MaxRowsXActivity := reporteshelper.MinComMul_Armonization(armoPED, armoPI, len(indicadores))

			y_lin := indice
			h_lin := MaxRowsXActivity / len(armoPED)
			consolidadoExcelEvaluacion.SetRowHeight(sheetName, y_lin, 27*5)
			for _, lin := range armoPED {
				consolidadoExcelEvaluacion.MergeCell(sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1))
				consolidadoExcelEvaluacion.SetCellValue(sheetName, "B"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
				consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B"+fmt.Sprint(y_lin), "B"+fmt.Sprint(y_lin+h_lin-1), styleLineamiento)
				y_met := y_lin
				h_met := h_lin / len(lin["meta"].([]map[string]interface{}))
				for _, met := range lin["meta"].([]map[string]interface{}) {
					consolidadoExcelEvaluacion.MergeCell(sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "C"+fmt.Sprint(y_met), met["nombreMeta"])
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "C"+fmt.Sprint(y_met), "C"+fmt.Sprint(y_met+h_met-1), stylecontentC)
					y_est := y_met
					h_est := h_met / len(met["estrategias"].([]map[string]interface{}))
					for _, est := range met["estrategias"].([]map[string]interface{}) {
						consolidadoExcelEvaluacion.MergeCell(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1))
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "D"+fmt.Sprint(y_est), est["descripcionEstrategia"])
						if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontentC)
						} else {
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "D"+fmt.Sprint(y_est), "D"+fmt.Sprint(y_est+h_est-1), stylecontent)
						}
						y_est += h_est
					}
					y_met += h_met
				}
				y_lin += h_lin
			}
			y_eje := indice
			h_eje := MaxRowsXActivity / len(armoPI)
			for _, eje := range armoPI {
				consolidadoExcelEvaluacion.MergeCell(sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1))
				consolidadoExcelEvaluacion.SetCellValue(sheetName, "E"+fmt.Sprint(y_eje), eje["nombreFactor"])
				consolidadoExcelEvaluacion.SetCellStyle(sheetName, "E"+fmt.Sprint(y_eje), "E"+fmt.Sprint(y_eje+h_eje-1), stylecontentC)
				y_lin := y_eje
				h_lin := h_eje / len(eje["lineamientos"].([]map[string]interface{}))
				for _, lin := range eje["lineamientos"].([]map[string]interface{}) {
					consolidadoExcelEvaluacion.MergeCell(sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "F"+fmt.Sprint(y_lin), lin["nombreLineamiento"])
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "F"+fmt.Sprint(y_lin), "F"+fmt.Sprint(y_lin+h_lin-1), stylecontentC)
					y_est := y_lin
					h_est := h_lin / len(lin["estrategias"].([]map[string]interface{}))
					for _, est := range lin["estrategias"].([]map[string]interface{}) {
						consolidadoExcelEvaluacion.MergeCell(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1))
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "G"+fmt.Sprint(y_est), est["descripcionEstrategia"])
						if (est["nombreEstrategia"].(string) == "No seleccionado") || strings.Contains(strings.ToLower(est["nombreEstrategia"].(string)), "no aplica") {
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontentC)
						} else {
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "G"+fmt.Sprint(y_est), "G"+fmt.Sprint(y_est+h_est-1), stylecontent)
						}
						y_est += h_est
					}
					y_lin += h_lin
				}
				y_eje += h_eje
			}

			for i, actividad := range evaluacion {
				if numeroActividad == actividad["numero"] {
					// Union de celdas
					consolidadoExcelEvaluacion.MergeCell(sheetName, "B"+fmt.Sprint(indice), "B"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "C"+fmt.Sprint(indice), "C"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "D"+fmt.Sprint(indice), "D"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "E"+fmt.Sprint(indice), "E"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "F"+fmt.Sprint(indice), "F"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "G"+fmt.Sprint(indice), "G"+fmt.Sprint(indice+MaxRowsXActivity-1))

					consolidadoExcelEvaluacion.MergeCell(sheetName, "H"+fmt.Sprint(indice), "H"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "I"+fmt.Sprint(indice), "I"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "J"+fmt.Sprint(indice), "J"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "K"+fmt.Sprint(indice), "K"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "L"+fmt.Sprint(indice), "L"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "M"+fmt.Sprint(indice), "M"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "N"+fmt.Sprint(indice), "N"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "O"+fmt.Sprint(indice), "O"+fmt.Sprint(indice+MaxRowsXActivity-1))

					// Datos
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "H"+fmt.Sprint(indice), posArmonizacion+1)
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "I"+fmt.Sprint(indice), actividad["ponderado"].(float64)/100)
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "J"+fmt.Sprint(indice), actividad["periodo"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "K"+fmt.Sprint(indice), actividad["actividad"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "L"+fmt.Sprint(indice), actividad["indicador"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "M"+fmt.Sprint(indice), actividad["formula"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "N"+fmt.Sprint(indice), actividad["meta"].(float64))
					if actividad["unidad"] == "Porcentaje" {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "N"+fmt.Sprint(indice), actividad["meta"].(float64)/100)
					}
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "O"+fmt.Sprint(indice), actividad["unidad"])

					//TODO: AQUI MERGE DE CELDAS

					consolidadoExcelEvaluacion.MergeCell(sheetName, "P"+fmt.Sprint(indice), "P"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "Q"+fmt.Sprint(indice), "Q"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "R"+fmt.Sprint(indice), "R"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "S"+fmt.Sprint(indice), "S"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "T"+fmt.Sprint(indice), "T"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "U"+fmt.Sprint(indice), "U"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "V"+fmt.Sprint(indice), "V"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "W"+fmt.Sprint(indice), "W"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "X"+fmt.Sprint(indice), "X"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "Y"+fmt.Sprint(indice), "Y"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "Z"+fmt.Sprint(indice), "Z"+fmt.Sprint(indice+MaxRowsXActivity-1))

					consolidadoExcelEvaluacion.MergeCell(sheetName, "AA"+fmt.Sprint(indice), "AA"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AB"+fmt.Sprint(indice), "AB"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AC"+fmt.Sprint(indice), "AC"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AD"+fmt.Sprint(indice), "AD"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AE"+fmt.Sprint(indice), "AE"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AF"+fmt.Sprint(indice), "AF"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AG"+fmt.Sprint(indice), "AG"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AH"+fmt.Sprint(indice), "AH"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AI"+fmt.Sprint(indice), "AI"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AJ"+fmt.Sprint(indice), "AJ"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AK"+fmt.Sprint(indice), "AK"+fmt.Sprint(indice+MaxRowsXActivity-1))

					consolidadoExcelEvaluacion.MergeCell(sheetName, "AL"+fmt.Sprint(indice), "AL"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AM"+fmt.Sprint(indice), "AM"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AN"+fmt.Sprint(indice), "AN"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AO"+fmt.Sprint(indice), "AO"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AP"+fmt.Sprint(indice), "AP"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AQ"+fmt.Sprint(indice), "AQ"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AR"+fmt.Sprint(indice), "AR"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AS"+fmt.Sprint(indice), "AS"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AT"+fmt.Sprint(indice), "AT"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AU"+fmt.Sprint(indice), "AU"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AV"+fmt.Sprint(indice), "AV"+fmt.Sprint(indice+MaxRowsXActivity-1))

					consolidadoExcelEvaluacion.MergeCell(sheetName, "AW"+fmt.Sprint(indice), "AW"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AX"+fmt.Sprint(indice), "AX"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AY"+fmt.Sprint(indice), "AY"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "AZ"+fmt.Sprint(indice), "AZ"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BA"+fmt.Sprint(indice), "BA"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BB"+fmt.Sprint(indice), "BB"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BC"+fmt.Sprint(indice), "BC"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BD"+fmt.Sprint(indice), "BD"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BE"+fmt.Sprint(indice), "BE"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BF"+fmt.Sprint(indice), "BF"+fmt.Sprint(indice+MaxRowsXActivity-1))
					consolidadoExcelEvaluacion.MergeCell(sheetName, "BG"+fmt.Sprint(indice), "BG"+fmt.Sprint(indice+MaxRowsXActivity-1))

					// Trimestres
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "P"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["cualitativo"].(map[string]interface{})["reporte"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "Q"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["cualitativo"].(map[string]interface{})["dificultades"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "R"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre1"].(map[string]interface{})["numerador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "S"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre1"].(map[string]interface{})["denominador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "T"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["periodo"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "U"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["numeradorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "V"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["denominadorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "W"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["acumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "X"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["meta"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "Y"+fmt.Sprint(indice), actividad["trimestre1"].(map[string]interface{})["brecha"])

					if valor, existe := actividad["trimestre1"].(map[string]interface{})["actividad"].(float64); existe {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "Z"+fmt.Sprint(indice), valor)
					} else {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "Z"+fmt.Sprint(indice), 0)
					}

					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AA"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["cualitativo"].(map[string]interface{})["reporte"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AB"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["cualitativo"].(map[string]interface{})["dificultades"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AC"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre2"].(map[string]interface{})["numerador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AD"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre2"].(map[string]interface{})["denominador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AE"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["periodo"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AF"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["numeradorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AG"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["denominadorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AH"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["acumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AI"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["meta"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AJ"+fmt.Sprint(indice), actividad["trimestre2"].(map[string]interface{})["brecha"])

					if valor, existe := actividad["trimestre2"].(map[string]interface{})["actividad"].(float64); existe {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "AK"+fmt.Sprint(indice), valor)
					} else {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "AK"+fmt.Sprint(indice), 0)
					}

					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AL"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["cualitativo"].(map[string]interface{})["reporte"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AM"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["cualitativo"].(map[string]interface{})["dificultades"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AN"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre3"].(map[string]interface{})["numerador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AO"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre3"].(map[string]interface{})["denominador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AP"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["periodo"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AQ"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["numeradorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AR"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["denominadorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AS"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["acumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AT"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["meta"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AU"+fmt.Sprint(indice), actividad["trimestre3"].(map[string]interface{})["brecha"])
					if valor, existe := actividad["trimestre3"].(map[string]interface{})["actividad"].(float64); existe {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "AV"+fmt.Sprint(indice), valor)
					} else {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "AV"+fmt.Sprint(indice), 0)
					}

					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AW"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["cualitativo"].(map[string]interface{})["reporte"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AX"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["cualitativo"].(map[string]interface{})["dificultades"].(string))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AY"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre4"].(map[string]interface{})["numerador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "AZ"+fmt.Sprint(indice), reporteshelper.Convert2Num(actividad["trimestre4"].(map[string]interface{})["denominador"]))
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BA"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["periodo"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BB"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["numeradorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BC"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["denominadorAcumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BD"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["acumulado"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BE"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["meta"])
					consolidadoExcelEvaluacion.SetCellValue(sheetName, "BF"+fmt.Sprint(indice), actividad["trimestre4"].(map[string]interface{})["brecha"])
					if valor, existe := actividad["trimestre4"].(map[string]interface{})["actividad"].(float64); existe {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "BG"+fmt.Sprint(indice), valor)
					} else {
						consolidadoExcelEvaluacion.SetCellValue(sheetName, "BG"+fmt.Sprint(indice), 0)
					}

					// Estilos
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "H"+fmt.Sprint(indice), "AY"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoC)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "I"+fmt.Sprint(indice), "I"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "K"+fmt.Sprint(indice), "K"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenido)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "P"+fmt.Sprint(indice), "BG"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "R"+fmt.Sprint(indice), "S"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "U"+fmt.Sprint(indice), "V"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AC"+fmt.Sprint(indice), "AD"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AF"+fmt.Sprint(indice), "AG"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AN"+fmt.Sprint(indice), "AO"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AQ"+fmt.Sprint(indice), "AR"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AY"+fmt.Sprint(indice), "AZ"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BB"+fmt.Sprint(indice), "BC"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)

					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Z"+fmt.Sprint(indice), "Z"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCPS)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AK"+fmt.Sprint(indice), "AK"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCPS)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AV"+fmt.Sprint(indice), "AV"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCPS)
					consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BG"+fmt.Sprint(indice), "BG"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCPS)

					if actividad["unidad"] == "Porcentaje" {
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "N"+fmt.Sprint(indice), "N"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "T"+fmt.Sprint(indice), "T"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "W"+fmt.Sprint(indice), "W"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Y"+fmt.Sprint(indice), "Y"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AE"+fmt.Sprint(indice), "AE"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AH"+fmt.Sprint(indice), "AH"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AJ"+fmt.Sprint(indice), "AJ"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AP"+fmt.Sprint(indice), "AP"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AS"+fmt.Sprint(indice), "AS"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AU"+fmt.Sprint(indice), "AU"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BA"+fmt.Sprint(indice), "BA"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BD"+fmt.Sprint(indice), "BD"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BF"+fmt.Sprint(indice), "BF"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCP)
					} else if actividad["unidad"] == "Tasa" {
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "N"+fmt.Sprint(indice), "N"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "T"+fmt.Sprint(indice), "T"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "W"+fmt.Sprint(indice), "W"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Y"+fmt.Sprint(indice), "Y"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AE"+fmt.Sprint(indice), "AE"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AH"+fmt.Sprint(indice), "AH"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AJ"+fmt.Sprint(indice), "AJ"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AP"+fmt.Sprint(indice), "AP"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AS"+fmt.Sprint(indice), "AS"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AU"+fmt.Sprint(indice), "AU"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BA"+fmt.Sprint(indice), "BA"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BD"+fmt.Sprint(indice), "BD"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BF"+fmt.Sprint(indice), "BF"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCD)
					} else if actividad["unidad"] == "Unidad" {
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "N"+fmt.Sprint(indice), "N"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "T"+fmt.Sprint(indice), "T"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "W"+fmt.Sprint(indice), "W"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Y"+fmt.Sprint(indice), "Y"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AE"+fmt.Sprint(indice), "AE"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AH"+fmt.Sprint(indice), "AH"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AJ"+fmt.Sprint(indice), "AJ"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AP"+fmt.Sprint(indice), "AP"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AS"+fmt.Sprint(indice), "AS"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AU"+fmt.Sprint(indice), "AU"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BA"+fmt.Sprint(indice), "BA"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BD"+fmt.Sprint(indice), "BD"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
						consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BF"+fmt.Sprint(indice), "BF"+fmt.Sprint(indice+MaxRowsXActivity-1), styleContenidoCE)
					}

					// Unión de celdas por indicador
					if i > 0 {
						if actividad["numero"] == evaluacion[i-1]["numero"] {
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "H"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "I"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "J"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "K"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "Z"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "AK"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "AV"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "BG"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "BL"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.SetCellValue(sheetName, "BM"+fmt.Sprint(indice), nil)
							consolidadoExcelEvaluacion.MergeCell(sheetName, "P"+fmt.Sprint(indice-1), "P"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "Q"+fmt.Sprint(indice-1), "Q"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AA"+fmt.Sprint(indice-1), "AA"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AB"+fmt.Sprint(indice-1), "AB"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AL"+fmt.Sprint(indice-1), "AL"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AM"+fmt.Sprint(indice-1), "AM"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AW"+fmt.Sprint(indice-1), "AW"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AX"+fmt.Sprint(indice-1), "AX"+fmt.Sprint(indice))

							consolidadoExcelEvaluacion.MergeCell(sheetName, "B"+fmt.Sprint(indice-1), "B"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "C"+fmt.Sprint(indice-1), "C"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "D"+fmt.Sprint(indice-1), "D"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "E"+fmt.Sprint(indice-1), "E"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "F"+fmt.Sprint(indice-1), "F"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "G"+fmt.Sprint(indice-1), "G"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "B"+fmt.Sprint(indice-1), "B"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "C"+fmt.Sprint(indice-1), "C"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "D"+fmt.Sprint(indice-1), "D"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "E"+fmt.Sprint(indice-1), "E"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "F"+fmt.Sprint(indice-1), "F"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)
							consolidadoExcelEvaluacion.SetCellStyle(sheetName, "G"+fmt.Sprint(indice-1), "G"+fmt.Sprint(indice+MaxRowsXActivity-1), styleLineamiento)

							consolidadoExcelEvaluacion.MergeCell(sheetName, "H"+fmt.Sprint(indice-1), "H"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "I"+fmt.Sprint(indice-1), "I"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "J"+fmt.Sprint(indice-1), "J"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "K"+fmt.Sprint(indice-1), "K"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "Z"+fmt.Sprint(indice-1), "Z"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AK"+fmt.Sprint(indice-1), "AK"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "AV"+fmt.Sprint(indice-1), "AV"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.MergeCell(sheetName, "BG"+fmt.Sprint(indice-1), "BG"+fmt.Sprint(indice))
						} else {
							// Gaficos
							consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BL"+fmt.Sprint(indiceGraficos), "=H"+fmt.Sprint(indice))
							consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BM"+fmt.Sprint(indiceGraficos), "=IF(E4=\"Trimestre I\",Z"+fmt.Sprint(indice)+",IF(E4=\"Trimestre II\",AK"+fmt.Sprint(indice)+",IF(E4=\"Trimestre III\",AV"+fmt.Sprint(indice)+",IF(E4=\"Trimestre IV\",BG"+fmt.Sprint(indice)+"))))")
							indiceGraficos++
						}
					} else if i == 0 {
						// Gaficos
						consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BL"+fmt.Sprint(indiceGraficos), "=H"+fmt.Sprint(indice))
						consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BM"+fmt.Sprint(indiceGraficos), "=IF(E4=\"Trimestre I\",Z"+fmt.Sprint(indice)+",IF(E4=\"Trimestre II\",AK"+fmt.Sprint(indice)+",IF(E4=\"Trimestre III\",AV"+fmt.Sprint(indice)+",IF(E4=\"Trimestre IV\",BG"+fmt.Sprint(indice)+"))))")
						indiceGraficos++
					}
					indice += MaxRowsXActivity
				}
			}
			contadorArmonizacion += MaxRowsXActivity
		}

		consolidadoExcelEvaluacion.MergeCell(sheetName, "H"+fmt.Sprint(indice), "O"+fmt.Sprint(indice))
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "H"+fmt.Sprint(indice), "O"+fmt.Sprint(indice), styleTituloSB)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "P"+fmt.Sprint(indice), "BG"+fmt.Sprint(indice), styleContenidoC)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "Z"+fmt.Sprint(indice), "Z"+fmt.Sprint(indice), styleContenidoCPSR)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AK"+fmt.Sprint(indice), "AK"+fmt.Sprint(indice), styleContenidoCPSR)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "AV"+fmt.Sprint(indice), "AV"+fmt.Sprint(indice), styleContenidoCPSR)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BG"+fmt.Sprint(indice), "BG"+fmt.Sprint(indice), styleContenidoCPSR)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BL19", "BL"+fmt.Sprint(indice+1), styleContenidoCI)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BM19", "BN"+fmt.Sprint(indice+1), styleContenidoCIP)
		consolidadoExcelEvaluacion.SetCellStyle(sheetName, "BM21", "BM22", styleContenidoCI)
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "H"+fmt.Sprint(indice), "Avance General del Plan de Acción")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "E4", "Trimestre I")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "P"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "Q"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "R"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "S"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "T"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "U"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "V"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "W"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "X"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "Y"+fmt.Sprint(indice), "-")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AA"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AB"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AC"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AD"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AE"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AF"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AG"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AH"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AI"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AJ"+fmt.Sprint(indice), "-")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AL"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AM"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AN"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AO"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AP"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AQ"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AR"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AS"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AT"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AU"+fmt.Sprint(indice), "-")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AW"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AX"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AY"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "AZ"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BA"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BB"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BC"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BD"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BE"+fmt.Sprint(indice), "-")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BF"+fmt.Sprint(indice), "-")

		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BL"+fmt.Sprint(indice), "General")

		filaAnt := fmt.Sprint(indice - 1)
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "Z"+fmt.Sprint(indice), "=SUMPRODUCT(I23:I"+filaAnt+",Z23:Z"+filaAnt+")")
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "AK"+fmt.Sprint(indice), "=SUMPRODUCT(I23:I"+filaAnt+",AK23:AK"+filaAnt+")")
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "AV"+fmt.Sprint(indice), "=SUMPRODUCT(I23:I"+filaAnt+",AV23:AV"+filaAnt+")")
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BG"+fmt.Sprint(indice), "=SUMPRODUCT(I23:I"+filaAnt+",BG23:BG"+filaAnt+")")
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BM"+fmt.Sprint(indice), "=IF(E4=\"Trimestre I\",Z"+fmt.Sprint(indice)+",IF(E4=\"Trimestre II\",AK"+fmt.Sprint(indice)+",IF(E4=\"Trimestre III\",AV"+fmt.Sprint(indice)+",IF(E4=\"Trimestre IV\",BG"+fmt.Sprint(indice)+"))))")
		consolidadoExcelEvaluacion.SetCellFormula(sheetName, "BN"+fmt.Sprint(indice), "=100%-BM"+fmt.Sprint(indice))
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BM"+fmt.Sprint(indice+1), "Avance")
		consolidadoExcelEvaluacion.SetCellValue(sheetName, "BN"+fmt.Sprint(indice+1), "Restante")

		consolidadoExcelEvaluacion.AddChart(sheetName, "B5", &excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "",
					Categories: sheetName + "!$BM$" + fmt.Sprint(indice+1) + ":$BN$" + fmt.Sprint(indice+1),
					Values:     sheetName + "!$BM$" + fmt.Sprint(indice) + ":$BN$" + fmt.Sprint(indice),
				},
			},
			Format: excelize.GraphicOptions{
				ScaleX:          1.0,
				ScaleY:          1.0,
				OffsetX:         15,
				OffsetY:         10,
				LockAspectRatio: false,
				Locked:          &disable,
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName:     false,
				ShowLeaderLines: false,
				ShowPercent:     true,
				ShowSerName:     false,
				ShowVal:         false,
			},
			ShowBlanksAs: "zero",
			Dimension: excelize.ChartDimension{
				Height: 265,
				Width:  454,
			},
			XAxis: excelize.ChartAxis{
				None: true,
			},
			YAxis: excelize.ChartAxis{
				None: true,
			},
		})

		consolidadoExcelEvaluacion.AddChart(sheetName, "F4", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{
				{
					Name:       "",
					Categories: sheetName + "!$BL$23:$BL$" + fmt.Sprint(indiceGraficos-1),
					Values:     sheetName + "!$BM$23:$BM$" + fmt.Sprint(indiceGraficos-1),
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX:         15,
				LockAspectRatio: false,
				Locked:          &disable,
			},
			Dimension: excelize.ChartDimension{
				Height: 344,
				Width:  1605,
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName:     false,
				ShowLeaderLines: false,
				ShowPercent:     false,
				ShowSerName:     false,
				ShowVal:         true,
			},
			YAxis: excelize.ChartAxis{
				MajorGridLines: true,
				Font:           excelize.Font{Family: "Calibri", Size: 9, Color: ColorNegro},
			},
			XAxis: excelize.ChartAxis{
				Font: excelize.Font{Family: "Calibri", Size: 9, Color: ColorNegro},
			},
			VaryColors:   &disable,
			ShowBlanksAs: "span",
		})

		buf, _ := consolidadoExcelEvaluacion.WriteToBuffer()
		strings.NewReader(buf.String())
		encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

		dataSend := make(map[string]interface{})
		dataSend["generalData"] = arregloPlanAnual
		dataSend["excelB64"] = encoded
		return dataSend, nil
	} else {
		return nil, errors.New("error ProcesarPlanAccionEvaluacion: No se agrego la propiedad unidad_id")
	}
}

func convert(valid []string, index string) ([]map[string]interface{}, map[string]interface{}) {
	var validadores []map[string]interface{}
	var actividad map[string]interface{}
	var dato_armonizacion map[string]interface{}
	var detallesLlenados bool
	var detalles []map[string]interface{}
	var detalles_armonizacion map[string]interface{}

	armonizacion := make(map[string]interface{})
	forkData := make(map[string]interface{})
	for i, v := range valid {
		var res map[string]interface{}
		var subgrupo_detalle []map[string]interface{}
		var dato_plan map[string]interface{}

		if !detallesLlenados {
			detalles = append(detalles, map[string]interface{}{})
			if err := request.GetJson("http://"+beego.AppConfig.String("PlanesService")+"/subgrupo-detalle?query=subgrupo_id:"+v+"&fields=dato_plan,armonizacion_dato", &res); err == nil {
				request.LimpiezaRespuestaRefactor(res, &subgrupo_detalle)

				if len(subgrupo_detalle) > 0 {
					if subgrupo_detalle[0]["armonizacion_dato"] != nil {
						dato_armonizacion_str := subgrupo_detalle[0]["armonizacion_dato"].(string)
						json.Unmarshal([]byte(dato_armonizacion_str), &dato_armonizacion)
						detalles_armonizacion = dato_armonizacion
						armonizacion["armo"] = dato_armonizacion[index]
					}
					if subgrupo_detalle[0]["dato_plan"] != nil {
						dato_plan_str := subgrupo_detalle[0]["dato_plan"].(string)
						json.Unmarshal([]byte(dato_plan_str), &dato_plan)

						if dato_plan[index] != nil {
							actividad = dato_plan[index].(map[string]interface{})
							detalles[i] = dato_plan
							if v != "" {
								forkData[v] = actividad["dato"]
							}
						} else {
							detalles = append(detalles, map[string]interface{}{})
						}
					}
				}
			}
		} else {
			if detalles[i][index] != nil {
				forkData[v] = detalles[i][index].(map[string]interface{})["dato"]
			}
			if detalles_armonizacion[index] != nil {
				armonizacion["armo"] = detalles_armonizacion[index]
			}
		}
	}
	if !detallesLlenados {
		detallesLlenados = true
	}

	if detalles_armonizacion[index] == nil {
		armonizacion["armo"] = map[string]interface{}{
			"armonizacionPED": "",
			"armonizacionPI":  "",
		}
	}

	validadores = append(validadores, forkData)
	return validadores, armonizacion

}
