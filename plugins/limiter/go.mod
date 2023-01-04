module github.com/nyan233/littlerpc/plugins/limiter

go 1.19

replace (
	"github.com/nyan233/littlerpc" => "../../"
)

require (
	github.com/juju/ratelimit v1.0.2 // indirect
	github.com/nyan233/littlerpc v0.0.0-00010101000000-000000000000
)