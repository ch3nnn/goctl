package bug

import (
	"github.com/spf13/cobra"

	"gitlab.bolean.com/sa-micro-team/goctl/internal/cobrax"
)

// Cmd describes a bug command.
var Cmd = cobrax.NewCommand("bug", cobrax.WithRunE(cobra.NoArgs), cobrax.WithArgs(cobra.NoArgs))
