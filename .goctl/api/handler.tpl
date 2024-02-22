package {{.PkgName}}

import (
	"net/http"

	xhttp "gitlab.bolean.com/sa-micro-team/sa-micro-pkg/http/x"

	{{.ImportPackages}}
)

// {{.HandlerName}} {{.Summary}}
// @Tags {{.Tag}}
// @Summary {{.Summary}}
{{- if .HasSecurity}}
// @Security ApiKeyAuth{{end}}
// @Accept {{.Accept}}
// @Produce application/json
{{- range .SwagParams}}
// @Param {{.ParamName}} {{.ParamType}} {{.DataType}} {{.IsMandatory}} "{{.Comment}}" {{.Attribute}}{{- end}}
{{- if .HasRequestBody}}
// @Param data body types.{{.RequestType}} true "{{.Summary}}"{{end}}
// @Success 200 {object} types.Response{data={{.ResponseDataType}}} "A successful response."
// @Router {{.PathName}} [{{.MethodName}}]
func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := xhttp.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			{{if .HasResp}}xhttp.JsonBaseResponseCtx(r.Context(), w, resp){{else}}xhttp.JsonBaseResponseCtx(r.Context(), w, nil){{end}}
		}
	}
}
