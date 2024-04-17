package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "Desagregado",
            Router: "/desagregado",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "Necesidades",
            Router: "/necesidades/:nombre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "PlanAccionEvaluacion",
            Router: "/plan-anual-evaluacion/:nombre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "PlanAccionAnualGeneral",
            Router: "/plan-anual-general/:nombre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "PlanAccionAnual",
            Router: "/plan-anual/:nombre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/planeacion_reportes_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "ValidarReporte",
            Router: "/validacion",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
