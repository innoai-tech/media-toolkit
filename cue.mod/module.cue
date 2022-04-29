module: "github.com/innoai-tech/media-toolkit"

require: {
	"dagger.io":                      "v0.2.18-0.20220608075319-28308bda2857"
	"github.com/innoai-tech/runtime": "v0.0.0-20220613045720-209a645a0e65"
	"k8s.io/api":                     "v0.24.1"
	"k8s.io/apimachinery":            "v0.24.1"
	"universe.dagger.io":             "v0.2.18-0.20220608075319-28308bda2857"
}

replace: {
	"dagger.io":          "github.com/morlay/dagger/pkg/dagger.io@v0.2.18-0.20220608075319-28308bda2857"
	"universe.dagger.io": "github.com/morlay/dagger/pkg/universe.dagger.io@v0.2.18-0.20220608075319-28308bda2857"
}

replace: {
	"k8s.io/api":          "" @import("go")
	"k8s.io/apimachinery": "" @import("go")
}
