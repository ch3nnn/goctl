package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	conf "gitlab.bolean.com/sa-micro-team/goctl/config"
	"gitlab.bolean.com/sa-micro-team/goctl/rpc/parser"
	"gitlab.bolean.com/sa-micro-team/goctl/util"
	"gitlab.bolean.com/sa-micro-team/goctl/util/format"
	"gitlab.bolean.com/sa-micro-team/goctl/util/pathx"
	"gitlab.bolean.com/sa-micro-team/goctl/util/stringx"
)

//go:embed svc.tpl
var svcTemplate string

// GenSvc generates the servicecontext.go file, which is the resource dependency of a service,
// such as rpc dependency, model dependency, etc.
func (g *Generator) GenSvc(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetSvc()
	svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "service_context")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, svcFilename+".go")
	text, err := pathx.LoadTemplate(category, svcTemplateFile, svcTemplate)
	if err != nil {
		return err
	}

	serviceName := strings.ToLower(stringx.From(ctx.GetServiceName().Source()).ToCamel())
	if i := strings.Index(serviceName, "service"); i > 0 {
		serviceName = strings.TrimSuffix(serviceName[:i], "-")
	}

	return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]any{
		"imports":     fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
		"serviceName": serviceName,
	}, fileName, false)
}
