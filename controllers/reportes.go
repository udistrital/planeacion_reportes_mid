package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	reporteshelper "github.com/udistrital/planeacion_reportes_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

// ReportesController operations for Reportes
type ReportesController struct {
	beego.Controller
}

// URLMapping ...
func (c *ReportesController) URLMapping() {
	c.Mapping("ValidarReporte", c.ValidarReporte)
	c.Mapping("Desagregado", c.Desagregado)
	c.Mapping("PlanAccionAnual", c.PlanAccionAnual)
	c.Mapping("PlanAccionAnualGeneral", c.PlanAccionAnualGeneral)
	c.Mapping("Necesidades", c.Necesidades)
	c.Mapping("PlanAccionEvaluacion", c.PlanAccionEvaluacion)
}

// ReportesController ...
// @Title ValidarReporte
// @Description post ValidarReporte
// @Param	body		body 	{}	true		"body for Plan content"
// @Success 201 {object} models.Reportes
// @router /validar_reporte [post]
func (c *ReportesController) ValidarReporte() {
	defer request.ErrorController(c.Controller, "ReportesController")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "ValidarReporte", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "ValidarReporte", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.Validar(body); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "ValidarReporte", "err": err, "status": "400"})
	}
	c.ServeJSON()
}

// ReportesController ...
// @Title Desagregado
// @Description post Desagregado
// @Param	body		body 	{}	true		"body for Plan content"
// @Success 201 {object} models.Reportes
// @Failure 403 :plan_id is empty
// @router /desagregado [post]
func (c *ReportesController) Desagregado() {
	defer request.ErrorController(c.Controller, "ReportesController")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "Desagregado", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "Desagregado", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.ProcesarDesagregado(body); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "Desagregado", "err": err, "status": "400"})
	}
	c.ServeJSON()
}

// ReportesController ...
// @Title PlanAccionAnual
// @Description post PlanAccionAnual
// @Param	body		body 	{}	true		"body for Plan content"
// @Param	nombre		path 	string	true		"The key for staticblock"
// @Success 201 {object} models.Reportes
// @Failure 403 :plan_id is empty
// @router /plan_anual/:nombre [post]
func (c *ReportesController) PlanAccionAnual() {
	defer request.ErrorController(c.Controller, "ReportesController")

	nombre := c.Ctx.Input.Param(":nombre")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionAnual", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionAnual", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.ProcesarPlanAccionAnual(body, nombre); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "PlanAccionAnual", "err": err, "status": "400"})
	}
	c.ServeJSON()
}

// ReportesController ...
// @Title PlanAccionAnualGeneral
// @Description post PlanAccionAnualGeneral
// @Param	body		body 	{}	true		"body for Plan content"
// @Param	nombre		path 	string	true		"The key for staticblock"
// @Success 201 {object} models.Reportes
// @Failure 403 :plan_id is empty
// @router /plan_anual_general/:nombre [post]
func (c *ReportesController) PlanAccionAnualGeneral() {
	defer request.ErrorController(c.Controller, "ReportesController")

	nombre := c.Ctx.Input.Param(":nombre")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionAnualGeneral", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionAnualGeneral", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.ProcesarPlanAccionAnualGeneral(body, nombre); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "PlanAccionAnualGeneral", "err": err, "status": "400"})
	}
	c.ServeJSON()
}

// ReportesController ...
// @Title Necesidades
// @Description post Necesidades
// @Param	body		body 	{}	true		"body for Plan content"
// @Param	nombre		path 	string	true		"The key for staticblock"
// @Success 201 {object} models.Reportes
// @Failure 403 :plan_id is empty
// @router /necesidades/:nombre [post]
func (c *ReportesController) Necesidades() {
	defer request.ErrorController(c.Controller, "ReportesController")

	nombre := c.Ctx.Input.Param(":nombre")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "Necesidades", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "Necesidades", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.ProcesarNecesidades(body, nombre); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "Necesidades", "err": err, "status": "400"})
	}
	c.ServeJSON()
}

// ReportesController ...
// @Title PlanAccionEvaluacion
// @Description post PlanAccionEvaluacion
// @Param	body		body 	{}	true		"body for Plan content"
// @Param	nombre		path 	string	true		"The key for staticblock"
// @Success 201 {object} models.Reportes
// @Failure 403 :nombre is empty
// @router /plan_anual_evaluacion/:nombre [post]
func (c *ReportesController) PlanAccionEvaluacion() {
	defer request.ErrorController(c.Controller, "ReportesController")

	nombre := c.Ctx.Input.Param(":nombre")

	if v, e := request.ValidarBody(c.Ctx.Input.RequestBody); !v || e != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionEvaluacion", "err": request.ErrorBody, "status": "400"})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		panic(map[string]interface{}{"funcion": "PlanAccionEvaluacion", "err": err.Error(), "status": "400"})
	}

	if data, err := reporteshelper.ProcesarPlanAccionEvaluacion(body, nombre); err == nil {
		c.Data["json"] = map[string]interface{}{"Success": true, "Status": "201", "Message": "Successful", "Data": data}
	} else {
		panic(map[string]interface{}{"funcion": "PlanAccionEvaluacion", "err": err, "status": "400"})
	}
	c.ServeJSON()
}
