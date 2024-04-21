package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/planeacion_reportes_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
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
// @router /validacion [post]
func (c *ReportesController) ValidarReporte() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	resultado, err := services.ValidarReporte(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
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
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	resultado, err := services.Desagregado(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
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
// @router /plan-anual/:nombre [post]
func (c *ReportesController) PlanAccionAnual() {
	defer errorhandler.HandlePanic(&c.Controller)

	nombre := c.Ctx.Input.Param(":nombre")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PlanAccionAnual(nombre, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
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
// @router /plan-anual-general/:nombre [post]
func (c *ReportesController) PlanAccionAnualGeneral() {
	defer errorhandler.HandlePanic(&c.Controller)

	nombre := c.Ctx.Input.Param(":nombre")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PlanAccionAnualGeneral(nombre, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
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
	defer errorhandler.HandlePanic(&c.Controller)

	nombre := c.Ctx.Input.Param(":nombre")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.Necesidades(nombre, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
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
// @router /plan-anual-evaluacion/:nombre [post]
func (c *ReportesController) PlanAccionEvaluacion() {
	defer errorhandler.HandlePanic(&c.Controller)

	nombre := c.Ctx.Input.Param(":nombre")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PlanAccionEvaluacion(nombre, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}
