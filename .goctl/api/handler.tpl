package {{.PkgName}}

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

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
// @Success 200 {{.ResponseParamType}} types.{{.ResponseDataType}} "{"id":1}"
// @Router {{.PathName}} [{{.MethodName}}]
func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        {{if .HasRequest}}var req types.{{.RequestType}}
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        {{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
        {{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            {{if .HasResp}}httpx.OkJsonCtx(r.Context(), w, resp){{else}}httpx.Ok(w){{end}}
        }
	}
}