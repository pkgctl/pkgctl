package logs

import "flag"

type UpdateSubCmd struct {
	fs         *flag.FlagSet
	filterTool string
}
