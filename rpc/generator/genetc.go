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

//go:embed etc.tpl
var etcTemplate string

// GenEtc generates the yaml configuration file of the rpc service,
// including host, port monitoring configuration items and etcd configuration
func (g *Generator) GenEtc(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetEtc()
	etcFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, fmt.Sprintf("%v.yaml", etcFilename))

	text, err := pathx.LoadTemplate(category, etcTemplateFileFile, etcTemplate)
	if err != nil {
		return err
	}

	serviceName := strings.ToLower(stringx.From(ctx.GetServiceName().Source()).ToCamel())
	if i := strings.Index(serviceName, "service"); i > 0 {
		serviceName = strings.TrimSuffix(serviceName[:i], "-")
	}

	return util.With("etc").Parse(text).SaveTo(map[string]any{
		"serviceName": serviceName,
	}, fileName, false)
}
