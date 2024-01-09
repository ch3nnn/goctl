package gogen

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/zeromicro/go-zero/core/collection"

	"gitlab.bolean.com/sa-micro-team/goctl/api/spec"
	"gitlab.bolean.com/sa-micro-team/goctl/config"
	"gitlab.bolean.com/sa-micro-team/goctl/util/format"
	"gitlab.bolean.com/sa-micro-team/goctl/util/pathx"
	"gitlab.bolean.com/sa-micro-team/goctl/vars"
)

const (
	jwtTransKey    = "jwtTransition"
	routesFilename = "routes"
	routesTemplate = `// Code generated by goctl. DO NOT EDIT.
package handler

import (
	"net/http"{{if .hasTimeout}}
	"time"{{end}}

	{{.importPackages}}
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	{{.routesAdditions}}
}
`
	routesAdditionTemplate = `
	server.AddRoutes(
		{{.routes}} {{.jwt}}{{.signature}} {{.prefix}} {{.timeout}} {{.maxBytes}}
	)
`
	timeoutThreshold = time.Millisecond
)

var mapping = map[string]string{
	"delete":  "http.MethodDelete",
	"get":     "http.MethodGet",
	"head":    "http.MethodHead",
	"post":    "http.MethodPost",
	"put":     "http.MethodPut",
	"patch":   "http.MethodPatch",
	"connect": "http.MethodConnect",
	"options": "http.MethodOptions",
	"trace":   "http.MethodTrace",
}

type (
	group struct {
		routes           []route
		jwtEnabled       bool
		signatureEnabled bool
		authName         string
		timeout          string
		middlewares      []string
		prefix           string
		jwtTrans         string
		maxBytes         string
	}
	route struct {
		method  string
		path    string
		handler string
	}
)

func genRoutes(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	var builder strings.Builder
	groups, err := getRoutes(api)
	if err != nil {
		return err
	}

	templateText, err := pathx.LoadTemplate(category, routesAdditionTemplateFile, routesAdditionTemplate)
	if err != nil {
		return err
	}

	var hasTimeout bool
	gt := template.Must(template.New("groupTemplate").Parse(templateText))
	for _, g := range groups {
		var gbuilder strings.Builder
		gbuilder.WriteString("[]rest.Route{")
		for _, r := range g.routes {
			fmt.Fprintf(&gbuilder, `
		{
			Method:  %s,
			Path:    "%s",
			Handler: %s,
		},`,
				r.method, r.path, r.handler)
		}

		var jwt string
		if g.jwtEnabled {
			jwt = fmt.Sprintf("\n rest.WithJwt(serverCtx.Config.%s.AccessSecret),", g.authName)
		}
		if len(g.jwtTrans) > 0 {
			jwt = jwt + fmt.Sprintf("\n rest.WithJwtTransition(serverCtx.Config.%s.PrevSecret,serverCtx.Config.%s.Secret),", g.jwtTrans, g.jwtTrans)
		}
		var signature, prefix string
		if g.signatureEnabled {
			signature = "\n rest.WithSignature(serverCtx.Config.Signature),"
		}
		if len(g.prefix) > 0 {
			prefix = fmt.Sprintf(`
rest.WithPrefix("%s"),`, g.prefix)
		}

		var timeout string
		if len(g.timeout) > 0 {
			duration, err := time.ParseDuration(g.timeout)
			if err != nil {
				return err
			}

			// why we check this, maybe some users set value 1, it's 1ns, not 1s.
			if duration < timeoutThreshold {
				return fmt.Errorf("timeout should not less than 1ms, now %v", duration)
			}

			timeout = fmt.Sprintf("\n rest.WithTimeout(%d * time.Millisecond),", duration.Milliseconds())
			hasTimeout = true
		}

		var maxBytes string
		if len(g.maxBytes) > 0 {
			_, err := strconv.ParseInt(g.maxBytes, 10, 64)
			if err != nil {
				return fmt.Errorf("maxBytes %s parse error,it is an invalid number", g.maxBytes)
			}

			maxBytes = fmt.Sprintf("\n rest.WithMaxBytes(%s),", g.maxBytes)
		}

		var routes string
		if len(g.middlewares) > 0 {
			gbuilder.WriteString("\n}...,")
			params := g.middlewares
			for i := range params {
				params[i] = "serverCtx." + params[i]
			}
			middlewareStr := strings.Join(params, ", ")
			routes = fmt.Sprintf("rest.WithMiddlewares(\n[]rest.Middleware{ %s }, \n %s \n),",
				middlewareStr, strings.TrimSpace(gbuilder.String()))
		} else {
			gbuilder.WriteString("\n},")
			routes = strings.TrimSpace(gbuilder.String())
		}

		if err := gt.Execute(&builder, map[string]string{
			"routes":    routes,
			"jwt":       jwt,
			"signature": signature,
			"prefix":    prefix,
			"timeout":   timeout,
			"maxBytes":  maxBytes,
		}); err != nil {
			return err
		}
	}

	routeFilename, err := format.FileNamingFormat(cfg.NamingFormat, routesFilename)
	if err != nil {
		return err
	}

	routeFilename = routeFilename + ".go"
	filename := path.Join(dir, handlerDir, routeFilename)
	os.Remove(filename)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          handlerDir,
		filename:        routeFilename,
		templateName:    "routesTemplate",
		category:        category,
		templateFile:    routesTemplateFile,
		builtinTemplate: routesTemplate,
		data: map[string]any{
			"hasTimeout":      hasTimeout,
			"importPackages":  genRouteImports(rootPkg, api),
			"routesAdditions": strings.TrimSpace(builder.String()),
		},
	})
}

func genRouteImports(parentPkg string, api *spec.ApiSpec) string {
	importSet := collection.NewSet()
	importSet.AddStr(fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, contextDir)))
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			folder := route.GetAnnotation(groupProperty)
			if len(folder) == 0 {
				folder = group.GetAnnotation(groupProperty)
				if len(folder) == 0 {
					continue
				}
			}
			importSet.AddStr(fmt.Sprintf("%s \"%s\"", toPrefix(folder),
				pathx.JoinPackages(parentPkg, handlerDir, folder)))
		}
	}
	imports := importSet.KeysStr()
	sort.Strings(imports)
	projectSection := strings.Join(imports, "\n\t")
	depSection := fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL)
	return fmt.Sprintf("%s\n\n\t%s", depSection, projectSection)
}

func getRoutes(api *spec.ApiSpec) ([]group, error) {
	var routes []group

	for _, g := range api.Service.Groups {
		var groupedRoutes group
		for _, r := range g.Routes {
			handler := getHandlerName(r)
			handler = handler + "(serverCtx)"
			folder := r.GetAnnotation(groupProperty)
			if len(folder) > 0 {
				handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
			} else {
				folder = g.GetAnnotation(groupProperty)
				if len(folder) > 0 {
					handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
				}
			}
			groupedRoutes.routes = append(groupedRoutes.routes, route{
				method:  mapping[r.Method],
				path:    r.Path,
				handler: handler,
			})
		}

		groupedRoutes.timeout = g.GetAnnotation("timeout")
		groupedRoutes.maxBytes = g.GetAnnotation("maxBytes")

		jwt := g.GetAnnotation("jwt")
		if len(jwt) > 0 {
			groupedRoutes.authName = jwt
			groupedRoutes.jwtEnabled = true
		}
		jwtTrans := g.GetAnnotation(jwtTransKey)
		if len(jwtTrans) > 0 {
			groupedRoutes.jwtTrans = jwtTrans
		}

		signature := g.GetAnnotation("signature")
		if signature == "true" {
			groupedRoutes.signatureEnabled = true
		}
		middleware := g.GetAnnotation("middleware")
		if len(middleware) > 0 {
			groupedRoutes.middlewares = append(groupedRoutes.middlewares,
				strings.Split(middleware, ",")...)
		}
		prefix := g.GetAnnotation(spec.RoutePrefixKey)
		prefix = strings.ReplaceAll(prefix, `"`, "")
		prefix = strings.TrimSpace(prefix)
		if len(prefix) > 0 {
			prefix = path.Join("/", prefix)
			groupedRoutes.prefix = prefix
		}
		routes = append(routes, groupedRoutes)
	}

	return routes, nil
}

func toPrefix(folder string) string {
	return strings.ReplaceAll(folder, "/", "")
}
