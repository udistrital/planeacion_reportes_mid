package helpers

import (
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"reflect"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/planeacion_reportes_mid/models"
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

var estadoHttp string = "500"
var validDataT = []string{}
var hijos_key []interface{}
var hijos_data [][]map[string]interface{}
var detalles []map[string]interface{}
var detalles_armonizacion map[string]interface{}
var ids [][]string
var id_arr []string
var detallesLlenados bool

func LimpiarDetalles() {
	detalles = []map[string]interface{}{}
	detalles_armonizacion = map[string]interface{}{}
	detallesLlenados = false
}

func LimpiarIds() {
	id_arr = []string{}
}

func Limpiar() {
	validDataT = []string{}
	ids = [][]string{}
	hijos_data = nil
	hijos_key = nil
}

func Add(id string) {
	if !contains(validDataT, id) {
		validDataT = append(validDataT, id)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Convertir(valid []string, index string) ([]map[string]interface{}, map[string]interface{}) {
	var validadores []map[string]interface{}
	var actividad map[string]interface{}
	var dato_armonizacion map[string]interface{}
	armonizacion := make(map[string]interface{})
	forkData := make(map[string]interface{})
	for i, v := range valid {
		var res map[string]interface{}
		var subgrupo_detalle []map[string]interface{}
		var dato_plan map[string]interface{}
		if !detallesLlenados {
			detalles = append(detalles, map[string]interface{}{})
			url := "http://" + beego.AppConfig.String("PlanesService") + "/subgrupo-detalle?query=subgrupo_id:" + v + "&fields=dato_plan,armonizacion_dato"
			if err := request.GetJson(url, &res); err == nil {
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

func MinComMulArmonizacion(armoPED, armoPI []map[string]interface{}, lenIndicadores int) int {
	sizePED := &models.Nodo{Valor: len(armoPED)}
	for _, n2 := range armoPED {
		h1 := &models.Nodo{Valor: len(n2["meta"].([]map[string]interface{}))}
		sizePED.Hijos = append(sizePED.Hijos, h1)
		for _, n3 := range n2["meta"].([]map[string]interface{}) {
			h2 := &models.Nodo{Valor: len(n3["estrategias"].([]map[string]interface{}))}
			h1.Hijos = append(h1.Hijos, h2)
		}
	}

	sizePI := &models.Nodo{Valor: len(armoPI)}
	for _, n2 := range armoPI {
		h1 := &models.Nodo{Valor: len(n2["lineamientos"].([]map[string]interface{}))}
		sizePI.Hijos = append(sizePI.Hijos, h1)
		for _, n3 := range n2["lineamientos"].([]map[string]interface{}) {
			h2 := &models.Nodo{Valor: len(n3["estrategias"].([]map[string]interface{}))}
			h1.Hijos = append(h1.Hijos, h2)
		}
	}

	fitSize1 := false
	fitSize2 := false
	fitSize3 := false
	rowMax := lenIndicadores
	for !(fitSize1 && fitSize2 && fitSize3) {
		fitSize1 = calcMinCol(sizePED, rowMax)
		fitSize2 = calcMinCol(sizePI, rowMax)
		fitSize3 = (rowMax % lenIndicadores) == 0
		rowMax++
	}
	return rowMax - 1
}

func calcMinCol(node *models.Nodo, size int) bool {
	if node.Valor == 0 {
		node.Valor = 1
	}
	if (size % node.Valor) == 0 {
		node.Divisible = true
		div := size / node.Valor
		for _, hijo := range node.Hijos {
			if !calcMinCol(hijo, div) {
				return false
			}
		}
	} else {
		node.Divisible = false
	}
	return node.Divisible
}

func IdentificacionAntigua(dato string) interface{} {
	var result interface{}
	var identi map[string]interface{}
	var data_identi []map[string]interface{}

	if dato == "{}" {
		result = "{}"
	} else {
		json.Unmarshal([]byte(dato), &identi)
		for key := range identi {
			element := identi[key].(map[string]interface{})
			if element["activo"] == true {
				data_identi = append(data_identi, element)
			}
		}
		result = data_identi
	}
	return result
}

func NombreRubroPorCodigo(rubros []map[string]interface{}, codigo string) string {
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

func CodigoRubrosDocentes(rubros []map[string]interface{}, categoria string) string {
	for _, rubro := range rubros {
		if rubro["categoria"] == categoria {
			if codigo, exist := rubro["rubro"].(string); exist {
				return codigo
			}
			break
		}
	}
	return "No definido"
}

func DeformatNumberInt(atributo interface{}) (int, error) {
	strAtributo := strings.TrimLeft(atributo.(string), "$")
	strAtributo = strings.ReplaceAll(strAtributo, ",", "")
	arrAtributo := strings.Split(strAtributo, ".")
	auxAtributo, err := strconv.Atoi(arrAtributo[0])
	if err != nil {
		return 0, err
	}
	return auxAtributo, nil
}

func DeformatNumberFloat(atributo interface{}) (float64, error) {
	strAtributo := strings.TrimLeft(atributo.(string), "$")
	strAtributo = strings.ReplaceAll(strAtributo, ",", "")
	arrAtributo := strings.Split(strAtributo, ".")
	auxAtributo, err := strconv.ParseFloat(arrAtributo[0], 64)
	if err != nil {
		return 0, err
	}
	return auxAtributo, nil
}

func getTotales(detalles []map[string]interface{}, totales *models.TotalDocentVal) {
	for _, aux := range detalles {
		if aux["sueldoBasico"] != nil {
			auxSueldoBasico, err := DeformatNumberInt(aux["sueldoBasico"])
			if err == nil {
				totales.SalarioBasico += auxSueldoBasico * int(aux["cantidad"].(float64))
			}
		}
		if aux["primaServicios"] != nil {
			auxPrimaServicios, err := DeformatNumberInt(aux["primaServicios"])
			if err == nil {
				totales.PrimaServicios += auxPrimaServicios
			}
		}
		if aux["primaNavidad"] != nil {
			auxPrimaNavidad, err := DeformatNumberInt(aux["primaNavidad"])
			if err == nil {
				totales.PrimaNavidad += auxPrimaNavidad
			}
		}
		if aux["primaVacaciones"] != nil {
			auxPrimaVacaciones, err := DeformatNumberInt(aux["primaVacaciones"])
			if err == nil {
				totales.PrimaVacaciones += auxPrimaVacaciones
			}
		}
		if aux["bonificacion"] != nil && aux["bonificacion"] != "N/A" {
			auxBonificacion, err := DeformatNumberInt(aux["bonificacion"])
			if err == nil {
				totales.Bonificacion += auxBonificacion
			}
		}
		if aux["interesesCesantias"] != nil && aux["interesesCesantias"] != "N/A" {
			auxInteresesCesantias, err := DeformatNumberInt(aux["interesesCesantias"])
			if err == nil {
				totales.InteresesCesantias += auxInteresesCesantias
			}
		}
		if aux["cesantiasPublico"] != nil {
			auxCesantiasPublico, err := DeformatNumberInt(aux["cesantiasPublico"])
			if err == nil {
				totales.CesantiasPublicas += auxCesantiasPublico
			}
		}
		if aux["cesantiasPrivado"] != nil {
			auxCesantiasPrivado, err := DeformatNumberInt(aux["cesantiasPrivado"])
			if err == nil {
				totales.CesantiasPrivadas += auxCesantiasPrivado
			}
		}
		if aux["totalSalud"] != nil {
			auxSalud, err := DeformatNumberInt(aux["totalSalud"])
			if err == nil {
				totales.Salud += auxSalud
			}
		}
		if aux["pensionesPublico"] != nil {
			auxPensionesPublicas, err := DeformatNumberInt(aux["pensionesPublico"])
			if err == nil {
				totales.PensionesPublicas += auxPensionesPublicas
			}
		}
		if aux["pensionesPrivado"] != nil {
			auxPensionesPrivadas, err := DeformatNumberInt(aux["pensionesPrivado"])
			if err == nil {
				totales.PensionesPrivadas += auxPensionesPrivadas
			}
		}
		if aux["caja"] != nil {
			auxCaja, err := DeformatNumberInt(aux["caja"])
			if err == nil {
				totales.Caja += auxCaja
			}
		}
		if aux["totalArl"] != nil {
			auxArl, err := DeformatNumberInt(aux["totalArl"])
			if err == nil {
				totales.Arl += auxArl
			}
		}
		if aux["icbf"] != nil {
			auxIcbf, err := DeformatNumberInt(aux["icbf"])
			if err == nil {
				totales.Icbf += auxIcbf
			}
		}
	}
}

func GetTotalDocentes(docentes map[string]interface{}) map[string]interface{} {
	var rhf []map[string]interface{}
	var rhvPre []map[string]interface{}
	var rhvPos []map[string]interface{}
	if docentes["rhf"] != "{}" {
		rhf = docentes["rhf"].([]map[string]interface{})
	}
	if docentes["rhv_pre"] != "{}" {
		rhvPre = docentes["rhv_pre"].([]map[string]interface{})
	}
	if docentes["rhv_pos"] != "{}" {
		rhvPos = docentes["rhv_pos"].([]map[string]interface{})
	}

	totales := models.TotalesDocentes{}
	getTotales(rhf, &totales.Rhf)
	getTotales(rhvPre, &totales.Rhv_pre)
	getTotales(rhvPos, &totales.Rhv_pos)

	totalDocentes := make(map[string]interface{})
	totalDocentes["sueldoBasico"] = totales.Rhf.SalarioBasico + totales.Rhv_pre.SalarioBasico + totales.Rhv_pos.SalarioBasico
	totalDocentes["primaServicios"] = totales.Rhf.PrimaServicios + totales.Rhv_pre.PrimaServicios + totales.Rhv_pos.PrimaServicios
	totalDocentes["primaNavidad"] = totales.Rhf.PrimaNavidad + totales.Rhv_pre.PrimaNavidad + totales.Rhv_pos.PrimaNavidad
	totalDocentes["primaVacaciones"] = totales.Rhf.PrimaVacaciones + totales.Rhv_pre.PrimaVacaciones + totales.Rhv_pos.PrimaVacaciones
	totalDocentes["bonificacion"] = totales.Rhf.Bonificacion + totales.Rhv_pre.Bonificacion + totales.Rhv_pos.Bonificacion
	totalDocentes["interesesCesantias"] = totales.Rhf.InteresesCesantias + totales.Rhv_pre.InteresesCesantias + totales.Rhv_pos.InteresesCesantias
	totalDocentes["cesantiasPublicas"] = totales.Rhf.CesantiasPublicas + totales.Rhv_pre.CesantiasPublicas + totales.Rhv_pos.CesantiasPublicas
	totalDocentes["cesantiasPrivadas"] = totales.Rhf.CesantiasPrivadas + totales.Rhv_pre.CesantiasPrivadas + totales.Rhv_pos.CesantiasPrivadas
	totalDocentes["salud"] = totales.Rhf.Salud + totales.Rhv_pre.Salud + totales.Rhv_pos.Salud
	totalDocentes["pensionesPublicas"] = totales.Rhf.PensionesPublicas + totales.Rhv_pre.PensionesPublicas + totales.Rhv_pos.PensionesPublicas
	totalDocentes["pensionesPrivadas"] = totales.Rhf.PensionesPrivadas + totales.Rhv_pre.PensionesPrivadas + totales.Rhv_pos.PensionesPrivadas
	totalDocentes["arl"] = totales.Rhf.Arl + totales.Rhv_pre.Arl + totales.Rhv_pos.Arl
	totalDocentes["caja"] = totales.Rhf.Caja + totales.Rhv_pre.Caja + totales.Rhv_pos.Caja
	totalDocentes["icbf"] = totales.Rhf.Icbf + totales.Rhv_pre.Icbf + totales.Rhv_pos.Icbf
	totalDocentes["TotalesPorTipo"] = totales
	return totalDocentes
}

func SombrearCeldas(excel *excelize.File, idActividad int, sheetName string, hCell string, vCell string, style int, styleSombreado int) {
	if idActividad%2 == 0 {
		excel.SetCellStyle(sheetName, hCell, vCell, style)
	} else {
		excel.SetCellStyle(sheetName, hCell, vCell, styleSombreado)
	}
}

func ConvertirNumero(value interface{}) interface{} {
	switch value := value.(type) {
	case float64:
		return value
	case string:
		num, _ := strconv.ParseFloat(value, 64)
		return num
	default:
		return "-"
	}
}

func ActualizarRecursosGeneral(valor int, recursosGeneral interface{}) int {
	result := 0
	if recursosGeneral != nil {
		if fmt.Sprint(reflect.TypeOf(recursosGeneral)) == "int" {
			result = recursosGeneral.(int) + valor
		} else {
			auxValor, err := DeformatNumberInt(recursosGeneral)
			if err == nil {
				result = auxValor + valor
			}
		}
	} else {
		result = valor
	}
	return result
}

func ActualizarRubro(valor int, rubro interface{}) int {
	result := 0
	if rubro != nil {
		auxValor, err := DeformatNumberInt(rubro)
		if err == nil {
			result = auxValor + valor
		}
	} else {
		result = valor
	}
	return result
}

func EstiloExcelRotacion(file *excelize.File, horizontal, vertical, fillColor string, rotation int, conFill bool) (int, error) {
	style := &excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: horizontal, Vertical: vertical, WrapText: true, TextRotation: rotation},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "top", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	}
	if conFill {
		style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{fillColor}}
	}
	return file.NewStyle(style)
}

func EstiloExcel(file *excelize.File, horizontal, vertical string, color string, conColor bool) (int, error) {
	style := &excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: horizontal, Vertical: vertical, WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{color}},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "top", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: 1},
		},
	}
	if conColor {
		style.Font = &excelize.Font{Bold: true, Color: ColorBlanco}
	} else {
		style.Font = &excelize.Font{Bold: true}
	}
	return file.NewStyle(style)
}

func EstiloExcelBordes(file *excelize.File, horizontal, vertical string, fillColor string, ult int, conFill bool) (int, error) {
	style := &excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: horizontal, Vertical: vertical, WrapText: true},
		Border: []excelize.Border{
			{Type: "right", Color: ColorNegro, Style: 1},
			{Type: "left", Color: ColorNegro, Style: 1},
			{Type: "bottom", Color: ColorNegro, Style: ult},
		},
	}
	if conFill {
		style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{fillColor}}
	}
	return file.NewStyle(style)
}
